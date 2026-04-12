package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fipaan/ap2-uni/op-assign/config"
	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/app"
	orderV1 "github.com/fipaan/ap2-uni-op-gen/op-assign/order-service/proto/v1"

	"google.golang.org/grpc"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

	db, err := sql.Open("postgres", config.OrderDB_DSN())
	if err != nil {
		log.Fatal(err)
	}

	App, err := app.NewApp(db)
	if err != nil {
		log.Fatal(err)
	}

	if err := App.GRPCServer.StartStatusListener(ctx, config.OrderDB_DSN()); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.POST("/orders", App.Handler.CreateOrder)
	r.GET("/orders/:id", App.Handler.GetOrder)
	r.PATCH("/orders/:id/cancel", App.Handler.CancelOrder)

	grpcLis, err := net.Listen("tcp", config.OrderGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	orderV1.RegisterOrderServiceServer(grpcServer, App.GRPCServer)

	go func() {
        <-ctx.Done()
        grpcServer.GracefulStop()
    }()

	go func() {
		log.Printf("Order gRPC running on %v", config.OrderGRPCAddr())
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatal(err)
		}
	}()

	s := &http.Server{
		Addr:           fmt.Sprintf(":%v", config.OrderPORT()),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Shutdown(shutdownCtx)
	}()

	log.Printf("Order Service running on :%v\n", config.OrderPORT())
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
