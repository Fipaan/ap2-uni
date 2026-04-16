package grpc

import (
	"context"
	"time"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/domain"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/usecase"
	paymentV1 "github.com/Fipaan/ap2-uni-op-gen/op-assign/payment-service/proto/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	paymentV1.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUsecase
}

func NewServer(uc *usecase.PaymentUsecase) *Server {
	return &Server{uc: uc}
}

func (s *Server) ProcessPayment(ctx context.Context, req *paymentV1.PaymentRequest) (*paymentV1.PaymentResponse, error) {
	p, err := s.uc.Process(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		return nil, mapUsecaseErr(err)
	}

	return toResponse(p), nil
}

func toResponse(p *domain.Payment) *paymentV1.PaymentResponse {
	return &paymentV1.PaymentResponse{
		PaymentId:     p.ID,
		TransactionId: p.TransactionID,
		Status:        p.Status,
		ProcessedAt:   timestamppb.New(time.Now().UTC()),
	}
}

func (s *Server) ListPayments(ctx context.Context, req *paymentV1.ListPaymentsRequest) (*paymentV1.ListPaymentsResponse, error) {
	ps, err := s.uc.List(ctx, req.GetStatus())
	if err != nil {
		return nil, mapUsecaseErr(err)
	}

	return toListPaymentsResponse(ps), nil
}

func toPaymentFull(p *domain.Payment) *paymentV1.PaymentFull {
	return &paymentV1.PaymentResponse{
		PaymentId:     p.ID,
		OrderId:       p.OrderID,
		TransactionId: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
		ProcessedAt:   timestamppb.New(time.Now().UTC()),
	}
}

func toListPaymentsResponse(ps *[]domain.Payment) *paymentV1.ListPaymentResponse {
	payments := make([]*paymentV1.PaymentResponse, 0, len(*ps))

	for i := range *ps {
		payments = append(payments, toResponse(&(*ps)[i]))
	}

	return &paymentV1.ListPaymentsResponse{
		Payments:    payments,
	}
}

func mapUsecaseErr(err error) error {
	switch err {
	case usecase.ErrInvalidOrderID, usecase.ErrInvalidAmount:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
