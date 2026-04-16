package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Fipaan/ap2-uni/op-assign/config"
	"github.com/Fipaan/ap2-uni/op-assign/payment-service/internal/app"
	paymentV1 "github.com/Fipaan/ap2-uni-op-gen/op-assign/payment-service/proto/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

	lis, err := net.Listen("tcp", config.PaymentGRPCAddr())
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
