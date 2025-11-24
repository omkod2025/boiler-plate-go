package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// DatabaseHelper provides helper functions for common database operations
type DatabaseHelper struct {
	db *PGX
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(db *PGX) *DatabaseHelper {
	return &DatabaseHelper{
		db: db,
	}
}

// Insert executes an INSERT query and returns the number of rows affected
func (h *DatabaseHelper) Insert(ctx context.Context, query string, args ...interface{}) (int64, error) {
	_, err := h.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	// For now, return 1 as default affected rows
	// In a real implementation, you might want to parse the result
	return 1, nil
}

// Update executes an UPDATE query and returns the number of rows affected
func (h *DatabaseHelper) Update(ctx context.Context, query string, args ...interface{}) (int64, error) {
	_, err := h.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w", err)
	}

	// For now, return 1 as default affected rows
	// In a real implementation, you might want to parse the result
	return 1, nil
}

// Delete executes a DELETE query and returns the number of rows affected
func (h *DatabaseHelper) Delete(ctx context.Context, query string, args ...interface{}) (int64, error) {
	_, err := h.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	// For now, return 1 as default affected rows
	// In a real implementation, you might want to parse the result
	return 1, nil
}

// Select executes a SELECT query and returns rows
func (h *DatabaseHelper) Select(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return rows, nil
}

// SelectOne executes a SELECT query and returns a single row
func (h *DatabaseHelper) SelectOne(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return h.db.QueryRowWithContext(ctx, query, args...)
}

// Count executes a COUNT query and returns the count
func (h *DatabaseHelper) Count(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := h.db.QueryRowWithContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}
	return count, nil
}

// Exists checks if a record exists
func (h *DatabaseHelper) Exists(ctx context.Context, query string, args ...interface{}) (bool, error) {
	var exists bool
	err := h.db.QueryRowWithContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists check failed: %w", err)
	}
	return exists, nil
}

// Transaction executes a function within a transaction
func (h *DatabaseHelper) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithTimeout executes a function within a transaction with timeout
func (h *DatabaseHelper) TransactionWithTimeout(ctx context.Context, timeout time.Duration, fn func(pgx.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return h.Transaction(ctx, fn)
}

// BatchInsert executes multiple INSERT queries in a batch
func (h *DatabaseHelper) BatchInsert(ctx context.Context, queries []string, args [][]interface{}) error {
	if len(queries) != len(args) {
		return fmt.Errorf("queries and args must have the same length")
	}

	batch := &pgx.Batch{}
	for i, query := range queries {
		batch.Queue(query, args[i]...)
	}

	br := h.db.ExecBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch insert failed at query %d: %w", i, err)
		}
	}

	return nil
}

// BatchUpdate executes multiple UPDATE queries in a batch
func (h *DatabaseHelper) BatchUpdate(ctx context.Context, queries []string, args [][]interface{}) error {
	if len(queries) != len(args) {
		return fmt.Errorf("queries and args must have the same length")
	}

	batch := &pgx.Batch{}
	for i, query := range queries {
		batch.Queue(query, args[i]...)
	}

	br := h.db.ExecBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch update failed at query %d: %w", i, err)
		}
	}

	return nil
}

// BatchDelete executes multiple DELETE queries in a batch
func (h *DatabaseHelper) BatchDelete(ctx context.Context, queries []string, args [][]interface{}) error {
	if len(queries) != len(args) {
		return fmt.Errorf("queries and args must have the same length")
	}

	batch := &pgx.Batch{}
	for i, query := range queries {
		batch.Queue(query, args[i]...)
	}

	br := h.db.ExecBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch delete failed at query %d: %w", i, err)
		}
	}

	return nil
}

// HealthCheck performs a health check on the database
func (h *DatabaseHelper) HealthCheck(ctx context.Context) error {
	return h.db.HealthCheck()
}

// GetPoolStats returns pool statistics
func (h *DatabaseHelper) GetPoolStats() interface{} {
	return h.db.Stat()
}

// Close closes the database connection
func (h *DatabaseHelper) Close() {
	h.db.Close()
}

// SimpleExec executes a query without returning rows (simplified version)
func (h *DatabaseHelper) SimpleExec(ctx context.Context, query string, args ...interface{}) error {
	_, err := h.db.ExecWithContext(ctx, query, args...)
	return err
}

// SimpleQuery executes a query and returns rows (simplified version)
func (h *DatabaseHelper) SimpleQuery(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return h.db.Query(ctx, query, args...)
}

// SimpleQueryRow executes a query and returns a single row (simplified version)
func (h *DatabaseHelper) SimpleQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return h.db.QueryRowWithContext(ctx, query, args...)
}
