package security

import "database/sql"

func FetchOrderByCodeUnsafe(db *sql.DB, inputID string) (*sql.Rows, error) {
	query := "SELECT * FROM orders WHERE id = '" + inputID + "'"
	return db.Query(query)
}
