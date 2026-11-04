package service

import (
	"testing"
	"github.com/enterprise/order-fulfillment/internal/model"
)

func TestCalculateOrderTotal(t *testing.T) {
	s := NewOrderService(nil)
	items := []model.OrderItem{
		{SKU: "A1", Quantity: 2, Price: 15.0},
		{SKU: "B2", Quantity: 1, Price: 20.0},
	}
	total := s.CalculateOrderTotal(items)
	if total != 50.0 {
		t.Errorf("Expected 50.0, got %f", total)
	}
	_, _ = s.FulfillOrder("")
}
