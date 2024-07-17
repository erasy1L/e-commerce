package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/router"
	"github.com/erazr/ecommerce-microservices/internal/common/server"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

func main() {
	db, err := store.New(os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}

	userRepository := newUserRepository(db.Client)

	userHandler := newUserHandler(userRepository)

	r := router.New()

	r.Mount("/users", userHandler.Routes())

	server, err := server.New(server.WithHTTPServer(r, os.Getenv("USER_PORT")))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service started on port %s\n", os.Getenv("USER_PORT"))

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

	fmt.Println("Service stopped")
}
