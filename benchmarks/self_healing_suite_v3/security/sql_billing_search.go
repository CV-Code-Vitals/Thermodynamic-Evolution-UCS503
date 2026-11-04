package security

import "database/sql"

func SearchBillingPlanUnsafe(db *sql.DB, inputID string) (*sql.Rows, error) {
	query := "SELECT * FROM billing_plans WHERE id = '" + inputID + "'"
	return db.Query(query)
}
