package security

import "database/sql"

func QueryUserByIDUnsafe(db *sql.DB, inputID string) (*sql.Rows, error) {
	query := "SELECT * FROM users WHERE id = '" + inputID + "'"
	return db.Query(query)
}
