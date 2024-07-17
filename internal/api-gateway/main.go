package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/erazr/ecommerce-microservices/internal/api-gateway/docs"
	"github.com/erazr/ecommerce-microservices/internal/common/router"
	"github.com/erazr/ecommerce-microservices/internal/common/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	httpSwagger "github.com/swaggo/http-swagger"
)

//	@title			E-commerce Microservices API Gateway
//	@version		1.0
//	@description	This is the API Gateway for the E-commerce Microservices project.
//	@host			localhost:8080
//	@BasePath		/
//	@schemes		http
//	@consumes		json

func main() {
	r := router.New()

	r.Use(cors.AllowAll().Handler)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
	))

	mountServices(r)

	server, err := server.New(server.WithHTTPServer(r, os.Getenv("GATEWAY_PORT")))
	if err != nil {
		panic(err)
	}

	fmt.Println("API Gateway started on port", os.Getenv("GATEWAY_PORT"), "swagger on http://localhost:8080/swagger/index.html")

	if err := server.Start(); err != nil {
		panic(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		panic(err)
	}

	fmt.Println("Gateway stopped")

}

func mountServices(r *chi.Mux) {
	services := []struct {
		Host string
		Port string
		Path string
	}{
		{
			Host: os.Getenv("USER_HOST"),
			Port: os.Getenv("USER_PORT"),
			Path: os.Getenv("USER_PATH"),
		},
		{
			Host: os.Getenv("PRODUCT_HOST"),
			Port: os.Getenv("PRODUCT_PORT"),
			Path: os.Getenv("PRODUCT_PATH"),
		},
		{
			Host: os.Getenv("ORDER_HOST"),
			Port: os.Getenv("ORDER_PORT"),
			Path: os.Getenv("ORDER_PATH"),
		},
		{
			Host: os.Getenv("PAYMENT_HOST"),
			Port: os.Getenv("PAYMENT_PORT"),
			Path: os.Getenv("PAYMENT_PATH"),
		},
	}

	for _, service := range services {
		url, err := url.Parse("http://" + service.Host + ":" + service.Port)
		if err != nil {
			panic(err)
		}

		fmt.Println("Mounting service ", service.Path, "on", url)

		r.Mount(service.Path, server.ProxyRequestHandler(url))
	}
}
