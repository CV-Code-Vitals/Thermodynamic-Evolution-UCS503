package security
import "database/sql"
func QuerySafeSQL(db *sql.DB, id int) *sql.Row {
	return db.QueryRow("SELECT id, name FROM items WHERE id = $1", id)
}
