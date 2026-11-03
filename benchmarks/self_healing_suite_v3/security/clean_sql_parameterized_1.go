package security

import "database/sql"

func QueryUserSafeClean(db *sql.DB, id string) (*sql.Rows, error) {
	return db.Query("SELECT * FROM items WHERE id = ?", id)
}
