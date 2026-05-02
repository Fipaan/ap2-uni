package http

import (
	"errors"
	"log"

	"github.com/Fipaan/ap2-uni/order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc *usecase.OrderUsecase
}

func NewHandler(uc *usecase.OrderUsecase) *Handler {
	return &Handler{uc: uc}
}

func (handler *Handler) Close() {
	handler.uc.Close()
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID 	  string `json:"customer_id"`
		CustomerEmail string `json:"customer_email"`
		ItemName      string `json:"item_name"`
		Amount        int64  `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	idempotencyKey := c.GetHeader("Idempotency-Key")

	code := 201

	order, err := h.uc.Create(c, req.CustomerID, req.CustomerEmail, req.ItemName, req.Amount, idempotencyKey)
	if errors.Is(err, usecase.ErrNoErrIdempotency) {
		code = 200
		err = nil
	} else if err != nil {
		var errorMsg string
		switch {
		case errors.Is(err, usecase.ErrInvalidCustomerID): fallthrough
		case errors.Is(err, usecase.ErrInvalidItem):       fallthrough
		case errors.Is(err, usecase.ErrInvalidAmount):
			code     = 400
			errorMsg = err.Error()
		case errors.Is(err, usecase.ErrPaymentNotAvailable):
			code     = 503
			errorMsg = err.Error()
		default:
			code     = 500
			errorMsg = "internal error"
			log.Printf("InternalError(%v): %v", code, err.Error())
		}
		c.JSON(code, gin.H{"error": errorMsg})
		return
	}

	c.JSON(code, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.Get(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, order)
}

func (h *Handler) GetPaymentsList(c *gin.Context) {
	status := c.Query("status")

	payments, err := h.uc.ListPayments(c, status)
	if err != nil {
		code := 500
		var errorMsg string
		switch {
		case errors.Is(err, usecase.ErrPaymentNotAvailable):
			code     = 503
			errorMsg = err.Error()
		default:
			code     = 500
			errorMsg = "internal error"
			log.Printf("InternalError(%v): %v", code, err.Error())
		}
		c.JSON(code, errorMsg)
		return
	}
	
	c.JSON(200, payments)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	err := h.uc.Cancel(c, id)
	if err != nil {
		var code int
		var errorMsg string
		switch {
		case errors.Is(err, usecase.ErrCancelNonExist):
			code = 404
			errorMsg = err.Error()
		case errors.Is(err, usecase.ErrNotPending):
			code = 409
			errorMsg = err.Error()
		default:
			code = 500
			errorMsg = "internal error"
		}
		c.JSON(code, gin.H{"error": errorMsg})
		return
	}

	c.JSON(200, gin.H{"status": "cancelled"})
}
