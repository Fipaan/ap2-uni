package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"fmt"

	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/app"
	"github.com/fipaan/ap2-uni/op-assign/common"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/order_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(db)

	r := gin.Default()

	r.POST("/orders", app.Handler.CreateOrder)
	r.GET("/orders/:id", app.Handler.GetOrder)
	r.PATCH("/orders/:id/cancel", app.Handler.CancelOrder)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%v", common.OrderPORT),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	log.Printf("Order Service running on :%v\n", common.OrderPORT)
	log.Fatal(s.ListenAndServe())
}
