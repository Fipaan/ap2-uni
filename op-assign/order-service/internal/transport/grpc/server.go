package grpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	orderV1 "github.com/fipaan/ap2-uni-op-gen/op-assign/order-service/proto/v1"

	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	orderV1.UnimplementedOrderServiceServer
	hub *Hub
}

func NewServer(hub *Hub) *Server {
	return &Server{hub: hub}
}

type orderStatusNotification struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Server) StartStatusListener(ctx context.Context, dbDSN string) error {
	listener := pq.NewListener(dbDSN, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("order listener event=%v err=%v", ev, err)
		}
	})
	if err := listener.Listen("order_updates"); err != nil {
		return err
	}

	go func() {
		defer listener.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case n := <-listener.Notify:
				if n == nil {
					continue
				}
				var payload orderStatusNotification
				if err := json.Unmarshal([]byte(n.Extra), &payload); err != nil {
					log.Printf("invalid order notification payload: %v", err)
					continue
				}

				updatedAt := time.Now().UTC()
				if payload.UpdatedAt != "" {
					if parsed, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt); err == nil {
						updatedAt = parsed.UTC()
					}
				}

				s.hub.Publish(&orderV1.OrderStatusUpdate{
					OrderId:   payload.OrderID,
					Status:    payload.Status,
					UpdatedAt: timestamppb.New(updatedAt),
				})
			}
		}
	}()

	return nil
}

func (s *Server) SubscribeToOrderUpdates(req *orderV1.OrderRequest, stream orderV1.OrderService_SubscribeToOrderUpdatesServer) error {
	if req.GetOrderId() == "" {
		return statusErrorInvalidArgument("order_id is empty")
	}

	ch, cancel := s.hub.Subscribe(req.GetOrderId())
	defer cancel()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

func statusErrorInvalidArgument(msg string) error {
	return &invalidArgumentError{msg: msg}
}

type invalidArgumentError struct{ msg string }

func (e *invalidArgumentError) Error() string { return e.msg }

var _ = sql.ErrNoRows
