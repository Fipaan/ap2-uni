package app

import (
	"database/sql"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/repo"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/transport/http"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/usecase"
)

type App struct {
	Handler *http.Handler
}

func NewApp(db *sql.DB) *App {
	paymentRepo    := repo.NewPaymentRepository(db)
	paymentUC      := usecase.NewPaymentUsecase(paymentRepo)
	paymentHandler := http.NewHandler(paymentUC)

	return &App{
		Handler: paymentHandler,
	}
}
