package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	paymentV1 "github.com/Fipaan/ap2-uni-op-gen/payment-service/proto/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var ErrPaymentNotAvailable = errors.New("payment service is not available")

type PaymentClient interface {
	Pay(ctx context.Context, orderID string, amount int64) (string, error)
	ListPayments(ctx context.Context, status string) ([]*paymentV1.PaymentFull, error)
}

type paymentClient struct {
	conn   *grpc.ClientConn
	client  paymentV1.PaymentServiceClient
}

func NewPaymentClient(url string) (PaymentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

   conn, err := grpc.DialContext(
           ctx,
           url,
           grpc.WithTransportCredentials(insecure.NewCredentials()),
   )

	if err != nil {
		return nil, err
	}

	return &paymentClient{
		conn:   conn,
		client: paymentV1.NewPaymentServiceClient(conn),
	}, nil
}

func (c *paymentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *paymentClient) Pay(ctx context.Context, orderID string, amount int64) (string, error) {
	resp, err := c.client.ProcessPayment(ctx, &paymentV1.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})

    if err != nil {
		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded) {
			return "", ErrPaymentNotAvailable
		}
		return "", fmt.Errorf("payment rpc failed: %w", err)
    }

	return resp.GetStatus(), nil
}

func (c *paymentClient) ListPayments(ctx context.Context, Status string) ([]*paymentV1.PaymentFull, error) {
	resp, err := c.client.ListPayments(ctx, &paymentV1.ListPaymentsRequest{
		Status: Status,
	})

    if err != nil {
		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded) {
			return nil, ErrPaymentNotAvailable
		}
		return nil, fmt.Errorf("payment rpc failed: %w", err)
    }

	return resp.GetPayments(), nil
}
