package grpc

import (
	"sync"

	orderV1 "github.com/fipaan/ap2-uni-op-gen/op-assign/order-service/proto/v1"
)

type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan *orderV1.OrderStatusUpdate]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string]map[chan *orderV1.OrderStatusUpdate]struct{}),
	}
}

func (h *Hub) Subscribe(orderID string) (chan *orderV1.OrderStatusUpdate, func()) {
	ch := make(chan *orderV1.OrderStatusUpdate, 16)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subs[orderID]; !ok {
		h.subs[orderID] = make(map[chan *orderV1.OrderStatusUpdate]struct{})
	}
	h.subs[orderID][ch] = struct{}{}

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if subs, ok := h.subs[orderID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(h.subs, orderID)
			}
		}
		close(ch)
	}

	return ch, cancel
}

func (h *Hub) Publish(update *orderV1.OrderStatusUpdate) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subs[update.GetOrderId()] {
		select {
		case ch <- update:
		default:
			// drop if subscriber is slow
		}
	}
}
