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

func mapUsecaseErr(err error) error {
	switch err {
	case usecase.ErrInvalidOrderID, usecase.ErrInvalidAmount:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
