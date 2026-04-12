package app

import (
	"database/sql"
	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/client"
	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/repo"
	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/transport/http"
	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/usecase"
	"github.com/fipaan/ap2-uni/op-assign/common"
	"fmt"
)

type App struct {
	Handler *http.Handler
}

func NewApp(db *sql.DB) *App {
	orderRepo     := repo.NewOrderRepository(db)
	paymentClient := client.NewPaymentClient(fmt.Sprintf("http://localhost:%v", common.PaymentPORT))
	orderUC       := usecase.NewOrderUsecase(orderRepo, paymentClient)
	h             := http.NewHandler(orderUC)

	return &App{
		Handler: h,
	}
}
