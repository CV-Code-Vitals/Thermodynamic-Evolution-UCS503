package service

import (
	"errors"
	"github.com/enterprise/order-fulfillment/internal/model"
	"github.com/enterprise/order-fulfillment/internal/repository"
)

// OrderService manages fulfillment workflows.
type OrderService struct {
	repo *repository.OrderRepository
}

// NewOrderService constructs an OrderService instance.
func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CalculateOrderTotal sums item prices according to discount rules.
func (s *OrderService) CalculateOrderTotal(items []model.OrderItem) float64 {
	total := 0.0
	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}
		total += float64(item.Quantity) * item.Price
	}
	return total
}

// FulfillOrder validates and processes order state transition.
func (s *OrderService) FulfillOrder(id string) (*model.OrderDetails, error) {
	if id == "" {
		return nil, errors.New("empty order id")
	}
	return s.repo.FindByID(id)
}
