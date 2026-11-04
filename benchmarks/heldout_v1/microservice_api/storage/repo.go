package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Neg_Control_SafeFindUser uses safe parameterized query
func (r *UserRepository) FindUserSafe(ctx context.Context, id int) (*sql.Rows, error) {
	const query = "SELECT id, name FROM users WHERE id = ?"
	return r.db.QueryContext(ctx, query, id)
}

// DEF-HOLDOUT-02: SQL injection via fmt.Sprintf
func (r *UserRepository) SearchUsersDynamic(ctx context.Context, role string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT id, name FROM users WHERE role = '%s'", role)
	return r.db.QueryContext(ctx, query)
}

