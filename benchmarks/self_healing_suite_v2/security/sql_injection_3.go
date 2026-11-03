package security

import "database/sql"

func QueryUnsafeSQL3(db *sql.DB, input string) (*sql.Rows, error) {
	q := "SELECT id, name FROM items WHERE name = '" + input + "'"
	return db.Query(q)
}
