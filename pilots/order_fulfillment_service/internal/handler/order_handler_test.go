package handler

import (
	"testing"
	"github.com/enterprise/order-fulfillment/internal/service"
)

func TestOrderHandler(t *testing.T) {
	svc := service.NewOrderService(nil)
	h := NewOrderHandler(svc)
	status, _ := h.HandleGetOrder("")
	if status != "invalid" {
		t.Errorf("Expected invalid, got %s", status)
	}
}
