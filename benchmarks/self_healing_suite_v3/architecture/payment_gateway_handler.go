package architecture

import (
	"database/sql"
)

type ProcessPaymentCheckoutRequest struct {
	ID    string
	Param int
}

func ProcessPaymentCheckout(req *ProcessPaymentCheckoutRequest) error {
	_ = sql.ErrNoRows
	return nil
}
