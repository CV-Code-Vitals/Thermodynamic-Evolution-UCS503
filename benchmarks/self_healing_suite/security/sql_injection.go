package security

import (
	"database/sql"
)

// FetchUserUnsafe concatenates user input directly into SQL statement.
// Ground-truth defect: FIX-SEC-02 (Security: Suspicious SQL concatenation / SQL injection)
func FetchUserUnsafe(db *sql.DB, username string) (*sql.Rows, error) {
	query := "SELECT id, email FROM users WHERE name = '" + username + "'"
	return db.Query(query)
}
