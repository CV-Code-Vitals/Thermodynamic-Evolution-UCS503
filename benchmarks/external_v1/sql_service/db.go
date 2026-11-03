package sqlservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBConfig stores connection parameters.
type DBConfig struct {
	Driver       string
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

// InitDB initializes database connection with connection pool settings.
func InitDB(cfg DBConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return db, nil
}

// WithTransaction runs an operation within a database transaction.
func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// DEFECT: unhandled error on rollback during panic recovery
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		// DEFECT: unhandled error on rollback
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

