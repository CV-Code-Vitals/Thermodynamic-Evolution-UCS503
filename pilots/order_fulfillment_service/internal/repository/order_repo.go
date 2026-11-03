package repository

import (
	"database/sql"
	"github.com/enterprise/order-fulfillment/internal/model"
)

// OrderRepository provides persistence for order records.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository initializes a new database repository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// FindByID retrieves an order record safely using parameterized binding.
func (r *OrderRepository) FindByID(id string) (*model.OrderDetails, error) {
	row := r.db.QueryRow("SELECT id, customer_id, total, status FROM orders WHERE id = ?", id)
	var o model.OrderDetails
	err := row.Scan(&o.ID, &o.CustomerID, &o.Total, &o.Status)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
