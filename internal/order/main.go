package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderpb "github.com/erazr/ecommerce-microservices/internal/common/pb/order"
	"github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/common/router"
	"github.com/erazr/ecommerce-microservices/internal/common/server"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/order/domain/order"
	"github.com/erazr/ecommerce-microservices/internal/order/handler"
	"github.com/erazr/ecommerce-microservices/internal/order/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	db, err := store.New(os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient("product:"+os.Getenv("PRODUCT_GRPC_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	orderRepository := repository.NewOrderRepository(db.Client)
	idempotencyCache := store.NewInMemoryIdempotencyCache[order.Order]()
	orderHandler := handler.NewOrderHandler(orderRepository, handler.WithIdempotencyCache(idempotencyCache), handler.WithProductGRPCService(product.NewProductsClient(conn)))

	r := router.New()
	r.Mount("/orders", orderHandler.Routes())

	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewOrderGRPCHandler(orderRepository)
	orderpb.RegisterOrdersServer(grpcServer, grpcHandler)

	server, err := server.New(server.WithHTTPServer(r, os.Getenv("ORDER_PORT")), server.WithGRPCServer(grpcServer, os.Getenv("ORDER_GRPC_PORT")))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service started on port %s\n", os.Getenv("ORDER_PORT"))

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
