package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/fipaan/ap2-uni/op-assign/config"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/app"
	paymentV1 "github.com/fipaan/ap2-uni-gen/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", config.PaymentDB_DSN())
	if err != nil {
		log.Fatal(err)
	}

	App, err := app.NewApp(db)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingUnaryInterceptor),
	)
	paymentV1.RegisterPaymentServiceServer(server, App.GRPCServer)

	log.Printf("Payment gRPC running on %v\n", config.PaymentGRPCAddr())
	log.Fatal(server.Serve(lis))
}

func loggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("%v(%v): %v: %v\n", info.FullMethod, err, time.Since(start),
		metadata.ValueFromIncomingContext(ctx, ""))
	return resp, err
}

var _ = fmt.Sprintf
