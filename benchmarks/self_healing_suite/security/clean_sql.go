package security

import (
	"database/sql"
)

// FetchUserSafe uses proper parameterized query binding.
// Ground-truth negative control: CLEAN-SEC-01 (Not a defect)
func FetchUserSafe(db *sql.DB, username string) (*sql.Row, error) {
	query := "SELECT id, email FROM users WHERE name = $1"
	return db.QueryRow(query, username), nil
}

