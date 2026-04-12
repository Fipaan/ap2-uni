package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"fmt"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/app"
	"github.com/fipaan/ap2-uni/op-assign/common"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/payment_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	application := app.NewApp(db)

	r := gin.Default()

	r.POST("/payments", application.Handler.CreatePayment)
	r.GET("/payments/:order_id", application.Handler.GetPayment)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%v", common.PaymentPORT),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	log.Printf("Payment Service running on %v\n", common.PaymentPORT)
	log.Fatal(s.ListenAndServe())
}
