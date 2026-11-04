package sqlservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// UserRepository handles user persistence.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a user using parameterized queries (clean implementation).
func (r *UserRepository) CreateUser(ctx context.Context, u *User) error {
	query := "INSERT INTO users (username, email, role, created_at) VALUES (?, ?, ?, ?)"
	now := time.Now()
	res, err := r.db.ExecContext(ctx, query, u.Username, u.Email, u.Role, now)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve inserted id: %w", err)
	}
	u.ID = id
	u.CreatedAt = now
	return nil
}

// GetUserByID fetches a user by primary key using parameterized query (clean implementation).
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := "SELECT id, username, email, role, created_at FROM users WHERE id = ?"
	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &u, nil
}

// FindUserByNameRaw retrieves a user using string formatting.
// DEFECT: SQL Injection vulnerability via unescaped fmt.Sprintf formatting into SQL statement.
func (r *UserRepository) FindUserByNameRaw(ctx context.Context, name string) (*User, error) {
	query := fmt.Sprintf("SELECT id, username, email, role, created_at FROM users WHERE username = '%s'", name)
	var u User
	err := r.db.QueryRowContext(ctx, query).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return &u, nil
}

// ListRecentUsers lists users created recently.
// DEFECT: sql.Rows is never closed, leaking open database cursor/statement handle.
func (r *UserRepository) ListRecentUsers(ctx context.Context, limit int) ([]User, error) {
	query := "SELECT id, username, email, role, created_at FROM users ORDER BY id DESC LIMIT ?"
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	// Missing: defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return users, nil
}

