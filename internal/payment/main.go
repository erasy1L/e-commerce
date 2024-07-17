package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/pb/order"
	"github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/common/router"
	"github.com/erazr/ecommerce-microservices/internal/common/server"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/payment/domain/payment"
	"github.com/erazr/ecommerce-microservices/internal/payment/epay"
	"github.com/erazr/ecommerce-microservices/internal/payment/handler"
	"github.com/erazr/ecommerce-microservices/internal/payment/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	db, err := store.New(os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}

	paymentRepository := repository.NewPaymentRepository(db.Client)

	orderConn, err := grpc.NewClient("order:"+os.Getenv("ORDER_GRPC_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer orderConn.Close()

	productConn, err := grpc.NewClient("product:"+os.Getenv("PRODUCT_GRPC_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer productConn.Close()

	ePayService := epay.NewService()

	idempotencyCache := store.NewInMemoryIdempotencyCache[payment.Payment]()

	paymentHandler := handler.NewPaymentHandler(paymentRepository, *ePayService,
		handler.WithIdempotencyCache(idempotencyCache),
		handler.WithOrderGRPCService(order.NewOrdersClient(orderConn)),
		handler.WithProductGRPCService(product.NewProductsClient(productConn)),
	)

	r := router.New()

	r.Mount("/payments", paymentHandler.Routes())

	server, err := server.New(server.WithHTTPServer(r, os.Getenv("PAYMENT_PORT")))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service started on port %s\n", os.Getenv("PAYMENT_PORT"))

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
