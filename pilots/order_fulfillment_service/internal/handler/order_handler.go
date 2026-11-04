package handler

import (
	"github.com/enterprise/order-fulfillment/internal/service"
)

// OrderHandler processes incoming HTTP API fulfillment requests.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates an HTTP handler delegating to domain service.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// HandleGetOrder retrieves order status.
func (h *OrderHandler) HandleGetOrder(orderID string) (string, error) {
	if orderID == "" {
		return "invalid", nil
	}
	order, err := h.svc.FulfillOrder(orderID)
	if err != nil {
		return "error", err
	}
	return order.Status, nil
}
