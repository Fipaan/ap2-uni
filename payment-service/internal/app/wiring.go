package app

import (
	"database/sql"
	"os"

	"github.com/Fipaan/ap2-uni/payment-service/internal/infrastructure/mq"
	"github.com/Fipaan/ap2-uni/payment-service/internal/repo"
	"github.com/Fipaan/ap2-uni/payment-service/internal/usecase"
	grpctransport "github.com/Fipaan/ap2-uni/payment-service/internal/transport/grpc"
)

type App struct {
	GRPCServer *grpctransport.Server
	Publisher  *mq.Publisher
}

func NewApp(db *sql.DB) (*App, error) {
	pub, err := mq.NewPublisher(os.Getenv("RABBITMQ_URL"), "payments.events")
	if err != nil {
		return nil, err
	}

	paymentRepo := repo.NewPaymentRepository(db)
	paymentUC   := usecase.NewPaymentUsecase(paymentRepo, pub)
	server      := grpctransport.NewServer(paymentUC)

	return &App{
		GRPCServer: server,
		Publisher:  pub,
	}, nil
}

func (app *App) Close() {
	app.Publisher.Close()
}
