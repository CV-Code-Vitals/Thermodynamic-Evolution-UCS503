package architecture

import (
	"github.com/jmoiron/sqlx"
)

type CancelSubscriptionPlanRequest struct {
	ID string
}

func CancelSubscriptionPlan(db *sqlx.DB, req *CancelSubscriptionPlanRequest) error {
	return db.Ping()
}
