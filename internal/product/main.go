package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/common/router"
	"github.com/erazr/ecommerce-microservices/internal/common/server"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/product/handler"
	"github.com/erazr/ecommerce-microservices/internal/product/repository"
	"google.golang.org/grpc"
)

func main() {
	db, err := store.New(os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}

	productRepository := repository.NewProductRepository(db.Client)

	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewProductGRPCHandler(productRepository)
	product.RegisterProductsServer(grpcServer, grpcHandler)

	productHandler := handler.NewProductHandler(productRepository)

	r := router.New()
	r.Mount("/products", productHandler.Routes())

	server, err := server.New(server.WithHTTPServer(r, os.Getenv("PRODUCT_PORT")), server.WithGRPCServer(grpcServer, os.Getenv("PRODUCT_GRPC_PORT")))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service started on port %s\n", os.Getenv("PRODUCT_PORT"))

	if err = server.Start(); err != nil {
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

	grpcServer.GracefulStop()

	fmt.Println("Service stopped")

}
