package objx

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLQueryExecutor provides database integration for objx data maps.
type SQLQueryExecutor struct {
	db *sql.DB
}

func NewSQLQueryExecutor(db *sql.DB) *SQLQueryExecutor {
	return &SQLQueryExecutor{db: db}
}

// Neg_Control_ExecParameterized uses safe parameterized SQL query
func (e *SQLQueryExecutor) QueryUserSafe(ctx context.Context, id int) (*sql.Rows, error) {
	const query = "SELECT id, payload FROM data_store WHERE id = ?"
	return e.db.QueryContext(ctx, query, id)
}

// INJ-EXT2-SQL-01: Dynamic SQL injection via fmt.Sprintf concatenation
func (e *SQLQueryExecutor) QueryUserByRole(ctx context.Context, role string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT id, payload FROM data_store WHERE role = '%s'", role)
	return e.db.QueryContext(ctx, query)
}

// INJ-EXT2-SQL-02: Dynamic SQL string concatenation
func (e *SQLQueryExecutor) SearchByTagRaw(ctx context.Context, tag string) (*sql.Rows, error) {
	query := "SELECT id, payload FROM data_store WHERE tag = '" + tag + "'"
	return e.db.Query(query)
}
