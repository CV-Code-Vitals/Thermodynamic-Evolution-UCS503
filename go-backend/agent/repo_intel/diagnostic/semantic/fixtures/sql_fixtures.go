package fixtures

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// POSITIVE FIXTURES: Vulnerable SQL patterns

// Pos_SQL_Sprintf demonstrates dynamic SQL construction via fmt.Sprintf.
func Pos_SQL_Sprintf(db *sql.DB, username string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT id, name FROM users WHERE username = '%s'", username)
	return db.Query(query)
}

// Pos_SQL_Concat demonstrates dynamic SQL construction via binary string concatenation.
func Pos_SQL_Concat(db *sql.DB, userInput string) *sql.Row {
	query := "SELECT id, category FROM items WHERE category = '" + userInput + "'"
	return db.QueryRow(query)
}

// Pos_SQL_PropagatedLocal demonstrates untrusted input propagated through local assignment into Exec.
func Pos_SQL_PropagatedLocal(db *sql.DB, filter string) (sql.Result, error) {
	formatted := fmt.Sprintf("DELETE FROM logs WHERE level = '%s'", filter)
	statement := formatted
	return db.Exec(statement)
}

// Pos_SQL_ParamAsQuery demonstrates passing a raw query parameter directly to an execution API.
func Pos_SQL_ParamAsQuery(db *sql.DB, rawQuery string) (*sql.Rows, error) {
	return db.Query(rawQuery)
}

// Pos_SQL_ContextSink demonstrates dynamic SQL construction reaching a context-aware sink.
func Pos_SQL_ContextSink(ctx context.Context, db *sql.DB, name string) (*sql.Rows, error) {
	q := fmt.Sprintf("SELECT id, email FROM accounts WHERE name = '%s'", name)
	return db.QueryContext(ctx, q)
}

// NEGATIVE FIXTURES: Safe controls and non-vulnerable patterns

// Neg_SQL_StaticQuery demonstrates constant static SQL execution.
func Neg_SQL_StaticQuery(db *sql.DB) (*sql.Rows, error) {
	return db.Query("SELECT id, name, email FROM users WHERE active = 1")
}

// Neg_SQL_ParameterizedQuery demonstrates safe parameterized SQL with bind placeholders.
func Neg_SQL_ParameterizedQuery(db *sql.DB, username string) (*sql.Rows, error) {
	return db.Query("SELECT id, name FROM users WHERE username = ?", username)
}

// Neg_SQL_ParameterizedContext demonstrates safe parameterized SQL with dollar placeholders ($1).
func Neg_SQL_ParameterizedContext(ctx context.Context, db *sql.DB, name string) (*sql.Rows, error) {
	return db.QueryContext(ctx, "SELECT id FROM accounts WHERE name = $1", name)
}

// Neg_SQL_PreparedStatement demonstrates safe prepared statement binding.
func Neg_SQL_PreparedStatement(db *sql.DB, username string) (*sql.Stmt, error) {
	stmt, err := db.Prepare("SELECT id, name FROM users WHERE username = ?")
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

// Neg_SQL_SprintfNonSQL demonstrates fmt.Sprintf used for non-SQL logging/messaging.
func Neg_SQL_SprintfNonSQL(username string) {
	msg := fmt.Sprintf("Authentication attempt observed for user: %s", username)
	log.Println(msg)
}

// Neg_SQL_DynamicUnused demonstrates dynamic formatting that does not reach any SQL sink.
func Neg_SQL_DynamicUnused(username string) string {
	unusedQuery := fmt.Sprintf("SELECT * FROM audit_logs WHERE user = '%s'", username)
	return unusedQuery
}
