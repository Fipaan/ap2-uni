package grpc

import (
	"context"
	"time"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/domain"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/usecase"
	paymentV1 "github.com/fipaan/ap2-uni/op-assign/gen/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	paymentV!.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUsecase
}

func NewServer(uc *usecase.PaymentUsecase) *Server {
	return &Server{uc: uc}
}

func (s *Server) ProcessPayment(ctx context.Context, req *paymentV!.PaymentRequest) (*paymentV!.PaymentResponse, error) {
	p, err := s.uc.Process(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		return nil, mapUsecaseErr(err)
	}

	return toResponse(p), nil
}

func toResponse(p *domain.Payment) *paymentV!.PaymentResponse {
	return &paymentV!.PaymentResponse{
		PaymentId:     p.ID,
		TransactionId: p.TransactionID,
		Status:        p.Status,
		ProcessedAt:   timestamppb.New(time.Now().UTC()),
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
