package http

import (
	"errors"
	"log"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc *usecase.PaymentUsecase
}

func NewHandler(uc *usecase.PaymentUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	p, err := h.uc.Process(c, req.OrderID, req.Amount)
	if err != nil {
		var code int
		var errorMsg string
		switch {
		case errors.Is(err, usecase.ErrInvalidOrderID): fallthrough
		case errors.Is(err, usecase.ErrInvalidAmount):
			code     = 400
			errorMsg = err.Error()
		default:
			code     = 500
			errorMsg = "internal error"
			log.Printf("InternalError(%v): %v", code, err.Error())
		}
		c.JSON(code, gin.H{"error": errorMsg})
		return
	}

	c.JSON(200, gin.H{
		"status": p.Status,
		"transaction_id": p.TransactionID,
	})
}

func (h *Handler) GetPayment(c *gin.Context) {
	id := c.Param("order_id")

	p, err := h.uc.Get(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, p)
}
