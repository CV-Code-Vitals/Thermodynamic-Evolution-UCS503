package architecture

import (
	"database/sql"
)

type ReserveInventoryStockRequest struct {
	ID    string
	Param int
}

func ReserveInventoryStock(req *ReserveInventoryStockRequest) error {
	_ = sql.ErrNoRows
	return nil
}
