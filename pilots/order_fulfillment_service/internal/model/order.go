package model

// OrderItem represents a line item in an order.
type OrderItem struct {
	SKU      string
	Quantity int
	Price    float64
}

// OrderDetails encapsulates customer order state.
type OrderDetails struct {
	ID         string
	CustomerID string
	Items      []OrderItem
	Total      float64
	Status     string
}
