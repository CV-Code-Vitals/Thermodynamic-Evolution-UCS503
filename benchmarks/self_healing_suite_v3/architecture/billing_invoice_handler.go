package architecture

import (
	"database/sql"
)

type GenerateBillingInvoiceRequest struct {
	ID    string
	Param int
}

func GenerateBillingInvoice(req *GenerateBillingInvoiceRequest) error {
	_ = sql.ErrNoRows
	return nil
}
