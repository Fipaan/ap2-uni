package app

import (
	"database/sql"

	"github.com/Fipaan/ap2-uni/op-assign/config"
	"github.com/Fipaan/ap2-uni/op-assign/order-service/internal/repo"
	"github.com/Fipaan/ap2-uni/op-assign/order-service/internal/usecase"
	grpcclient "github.com/Fipaan/ap2-uni/op-assign/order-service/internal/client/grpc"
	grpctransport "github.com/Fipaan/ap2-uni/op-assign/order-service/internal/transport/grpc"
	httptransport "github.com/Fipaan/ap2-uni/op-assign/order-service/internal/transport/http"
)

type App struct {
	Handler    *httptransport.Handler
	GRPCServer *grpctransport.Server
	Hub        *grpctransport.Hub
	payment    grpcclient.PaymentClient
}

func NewApp(db *sql.DB) (*App, error) {
	orderRepo := repo.NewOrderRepository(db)

	paymentClient, err := grpcclient.NewPaymentClient(config.PaymentGRPCAddr())
	if err != nil {
		return nil, err
	}

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient)
	h := httptransport.NewHandler(orderUC)

	hub := grpctransport.NewHub()
	grpcServer := grpctransport.NewServer(hub)

	return &App{
		Handler:    h,
		GRPCServer: grpcServer,
		Hub:        hub,
		payment:    paymentClient,
	}, nil
}
