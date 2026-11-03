package security

import "database/sql"

func FilterAuditRecordsUnsafe(db *sql.DB, inputID string) (*sql.Rows, error) {
	query := "SELECT * FROM audit_logs WHERE id = '" + inputID + "'"
	return db.Query(query)
}
