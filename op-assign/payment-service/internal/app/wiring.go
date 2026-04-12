package app

import (
	"database/sql"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/repo"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/usecase"
	grpctransport "github.com/fipaan/ap2-uni/op-assign/payment-service/internal/transport/grpc"
)

type App struct {
	GRPCServer *grpctransport.Server
}

func NewApp(db *sql.DB) (*App, error) {
	paymentRepo := repo.NewPaymentRepository(db)
	paymentUC   := usecase.NewPaymentUsecase(paymentRepo)
	server      := grpctransport.NewServer(paymentUC)

	return &App{
		GRPCServer: server,
	}, nil
}
