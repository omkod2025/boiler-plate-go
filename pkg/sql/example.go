package sql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// ExampleUsage demonstrates how to use the PostgreSQL package
func ExampleUsage() {
	// สร้าง config
	config := PGXConfig{
		DB_HOST:                "localhost",
		DB_PORT:                "5432",
		DB_USER:                "postgres",
		DB_PASSWORD:            "password",
		DB_NAME:                "mydb",
		DB_MAX_CONNS:           10,
		DB_MIN_CONNS:           2,
		DB_MAX_CONN_LIFETIME:   5 * time.Minute,
		DB_MAX_CONN_IDLE_TIME:  1 * time.Minute,
		DB_HEALTH_CHECK_PERIOD: 30 * time.Second,
	}

	// สร้าง connection
	ctx := context.Background()
	db, err := NewPGX(ctx, config)
	if err != nil {
		log.Fatal("Failed to create database connection:", err)
	}
	defer db.Close()

	// ทดสอบการเชื่อมต่อ
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// ตัวอย่างการใช้งาน Exec
	exampleExec(db)

	// ตัวอย่างการใช้งาน Query
	exampleQuery(db)

	// ตัวอย่างการใช้งาน Transaction
	exampleTransaction(db)

	// ตัวอย่างการใช้งาน Helper
	exampleHelper(db)
}

// exampleExec demonstrates Exec operations
func exampleExec(db *PGX) {
	fmt.Println("=== Exec Examples ===")

	// INSERT
	query := "INSERT INTO users (name, email) VALUES ($1, $2)"
	_, err := db.Exec(query, "John Doe", "john@example.com")
	if err != nil {
		log.Printf("Insert failed: %v", err)
		return
	}
	fmt.Println("✓ Insert successful")

	// UPDATE
	query = "UPDATE users SET name = $1 WHERE email = $2"
	_, err = db.Exec(query, "Jane Doe", "john@example.com")
	if err != nil {
		log.Printf("Update failed: %v", err)
		return
	}
	fmt.Println("✓ Update successful")

	// DELETE
	query = "DELETE FROM users WHERE email = $1"
	_, err = db.Exec(query, "john@example.com")
	if err != nil {
		log.Printf("Delete failed: %v", err)
		return
	}
	fmt.Println("✓ Delete successful")
}

// exampleQuery demonstrates Query operations
func exampleQuery(db *PGX) {
	fmt.Println("\n=== Query Examples ===")

	// Query multiple rows
	query := "SELECT id, name, email FROM users WHERE active = $1"
	rows, err := db.Query(context.Background(), query, true)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Users:")
	for rows.Next() {
		var id int
		var name, email string
		err := rows.Scan(&id, &name, &email)
		if err != nil {
			log.Printf("Scan failed: %v", err)
			continue
		}
		fmt.Printf("  ID: %d, Name: %s, Email: %s\n", id, name, email)
	}

	// Query single row
	query = "SELECT id, name, email FROM users WHERE id = $1"
	row := db.QueryRowWithContext(context.Background(), query, 1)

	var id int
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		log.Printf("QueryRow failed: %v", err)
		return
	}
	fmt.Printf("Single user: ID: %d, Name: %s, Email: %s\n", id, name, email)
}

// exampleTransaction demonstrates Transaction operations
func exampleTransaction(db *PGX) {
	fmt.Println("\n=== Transaction Examples ===")

	ctx := context.Background()
	// ใช้ helper สำหรับ transaction
	helper := NewDatabaseHelper(db)
	err := helper.Transaction(ctx, func(tx pgx.Tx) error {
		// INSERT
		_, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
		if err != nil {
			return fmt.Errorf("insert failed: %w", err)
		}

		// UPDATE
		_, err = tx.Exec(ctx, "UPDATE users SET active = $1 WHERE name = $2", true, "John")
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}
	fmt.Println("✓ Transaction successful")
}

// exampleHelper demonstrates Helper operations
func exampleHelper(db *PGX) {
	fmt.Println("\n=== Helper Examples ===")

	helper := NewDatabaseHelper(db)
	ctx := context.Background()

	// Insert
	rowsAffected, err := helper.Insert(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
	if err != nil {
		log.Printf("Helper Insert failed: %v", err)
		return
	}
	fmt.Printf("✓ Insert successful, rows affected: %d\n", rowsAffected)

	// Update
	rowsAffected, err = helper.Update(ctx, "UPDATE users SET name = $1 WHERE email = $2", "Jane", "john@example.com")
	if err != nil {
		log.Printf("Helper Update failed: %v", err)
		return
	}
	fmt.Printf("✓ Update successful, rows affected: %d\n", rowsAffected)

	// Select
	rows, err := helper.Select(ctx, "SELECT id, name, email FROM users WHERE active = $1", true)
	if err != nil {
		log.Printf("Helper Select failed: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Users from helper:")
	for rows.Next() {
		var id int
		var name, email string
		err := rows.Scan(&id, &name, &email)
		if err != nil {
			log.Printf("Scan failed: %v", err)
			continue
		}
		fmt.Printf("  ID: %d, Name: %s, Email: %s\n", id, name, email)
	}

	// Count
	count, err := helper.Count(ctx, "SELECT COUNT(*) FROM users WHERE active = $1", true)
	if err != nil {
		log.Printf("Helper Count failed: %v", err)
		return
	}
	fmt.Printf("✓ Count successful: %d active users\n", count)

	// Exists
	exists, err := helper.Exists(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", "john@example.com")
	if err != nil {
		log.Printf("Helper Exists failed: %v", err)
		return
	}
	fmt.Printf("✓ Exists check: user exists = %t\n", exists)

	// Transaction with helper
	err = helper.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "Bob", "bob@example.com")
		if err != nil {
			return fmt.Errorf("insert failed: %w", err)
		}

		_, err = tx.Exec(ctx, "UPDATE counters SET user_count = user_count + 1")
		if err != nil {
			return fmt.Errorf("update counter failed: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Helper Transaction failed: %v", err)
		return
	}
	fmt.Println("✓ Helper Transaction successful")
}

// ExampleBatchOperations demonstrates batch operations
func ExampleBatchOperations(db *PGX) {
	fmt.Println("\n=== Batch Operations Examples ===")

	helper := NewDatabaseHelper(db)
	ctx := context.Background()

	// Batch Insert
	queries := []string{
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"INSERT INTO users (name, email) VALUES ($1, $2)",
	}

	args := [][]interface{}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	err := helper.BatchInsert(ctx, queries, args)
	if err != nil {
		log.Printf("Batch Insert failed: %v", err)
		return
	}
	fmt.Println("✓ Batch Insert successful")

	// Batch Update
	queries = []string{
		"UPDATE users SET active = $1 WHERE name = $2",
		"UPDATE users SET active = $1 WHERE name = $2",
	}

	args = [][]interface{}{
		{true, "Alice"},
		{false, "Bob"},
	}

	err = helper.BatchUpdate(ctx, queries, args)
	if err != nil {
		log.Printf("Batch Update failed: %v", err)
		return
	}
	fmt.Println("✓ Batch Update successful")
}

// ExampleHealthCheck demonstrates health check operations
func ExampleHealthCheck(db *PGX) {
	fmt.Println("\n=== Health Check Examples ===")

	// Health check
	err := db.HealthCheck()
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	fmt.Println("✓ Health check successful")

	// Get pool statistics
	stats := db.Stat()
	fmt.Printf("Pool statistics:\n")
	fmt.Printf("  Total connections: %d\n", stats.TotalConns())
	fmt.Printf("  Idle connections: %d\n", stats.IdleConns())
	fmt.Printf("  Acquired connections: %d\n", stats.AcquiredConns())

	// Get pool configuration
	config := db.Config()
	fmt.Printf("Pool configuration:\n")
	fmt.Printf("  Max connections: %d\n", config.MaxConns)
	fmt.Printf("  Min connections: %d\n", config.MinConns)
}
