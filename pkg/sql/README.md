# PostgreSQL Package Documentation

Package สำหรับจัดการการเชื่อมต่อและ operations กับ PostgreSQL database ใช้ pgx driver

## Features

- ✅ **Connection Pooling** - จัดการ connection pool อัตโนมัติ
- ✅ **Context Support** - รองรับ context สำหรับ timeout และ cancellation
- ✅ **Transaction Support** - รองรับ transactions
- ✅ **Batch Operations** - รองรับ batch queries
- ✅ **Health Check** - ตรวจสอบสถานะการเชื่อมต่อ
- ✅ **Statistics** - ดูสถิติ connection pool
- ✅ **Helper Functions** - ฟังก์ชันช่วยเหลือสำหรับ operations ต่างๆ

## Installation

```bash
go get github.com/jackc/pgx/v5
```

## Usage

### 1. Basic Setup

```go
package main

import (
    "context"
    "time"
    "fmg-auth-api/pkg/sql"
)

func main() {
    ctx := context.Background()
    
    // สร้าง config
    config := sql.PGXConfig{
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
    db, err := sql.NewPGX(ctx, config)
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // ทดสอบการเชื่อมต่อ
    if err := db.Ping(); err != nil {
        panic(err)
    }
}
```

### 2. Basic Operations

#### Exec (INSERT, UPDATE, DELETE)

```go
// INSERT
query := "INSERT INTO users (name, email) VALUES ($1, $2)"
commandTag, err := db.Exec(ctx, query, "John Doe", "john@example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rows affected: %d\n", commandTag.RowsAffected())

// UPDATE
query := "UPDATE users SET name = $1 WHERE id = $2"
commandTag, err := db.Exec(ctx, query, "Jane Doe", 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rows affected: %d\n", commandTag.RowsAffected())

// DELETE
query := "DELETE FROM users WHERE id = $1"
commandTag, err := db.Exec(ctx, query, 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rows affected: %d\n", commandTag.RowsAffected())
```

#### Query (SELECT)

```go
// Query multiple rows
query := "SELECT id, name, email FROM users WHERE active = $1"
rows, err := db.Query(ctx, query, true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name, email string
    err := rows.Scan(&id, &name, &email)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("ID: %d, Name: %s, Email: %s\n", id, name, email)
}

// Query single row
query := "SELECT id, name, email FROM users WHERE id = $1"
row := db.QueryRow(ctx, query, 1)

var id int
var name, email string
err := row.Scan(&id, &name, &email)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("ID: %d, Name: %s, Email: %s\n", id, name, email)
```

### 3. Transactions

```go
// Simple transaction
err := db.Transaction(ctx, func(tx pgx.Tx) error {
    // INSERT
    _, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
    if err != nil {
        return err
    }
    
    // UPDATE
    _, err = tx.Exec(ctx, "UPDATE users SET active = $1 WHERE name = $2", true, "John")
    if err != nil {
        return err
    }
    
    return nil
})

if err != nil {
    log.Fatal(err)
}
```

#### Transaction with Timeout

```go
// Transaction with 30 second timeout
err := db.TransactionWithTimeout(ctx, 30*time.Second, func(tx pgx.Tx) error {
    // Your transaction logic here
    return nil
})
```

### 4. Batch Operations

```go
// Batch INSERT
batch := &pgx.Batch{}
batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "Jane", "jane@example.com")
batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "Bob", "bob@example.com")

br := db.ExecBatch(ctx, batch)
defer br.Close()

for i := 0; i < batch.Len(); i++ {
    _, err := br.Exec()
    if err != nil {
        log.Fatal(err)
    }
}
```

### 5. Using DatabaseHelper

```go
// สร้าง helper
helper := sql.NewDatabaseHelper(db)

// Insert
rowsAffected, err := helper.Insert(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
if err != nil {
    log.Fatal(err)
}

// Update
rowsAffected, err = helper.Update(ctx, "UPDATE users SET name = $1 WHERE id = $2", "Jane", 1)
if err != nil {
    log.Fatal(err)
}

// Delete
rowsAffected, err = helper.Delete(ctx, "DELETE FROM users WHERE id = $1", 1)
if err != nil {
    log.Fatal(err)
}

// Select
rows, err := helper.Select(ctx, "SELECT id, name, email FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

// SelectOne
row := helper.SelectOne(ctx, "SELECT id, name, email FROM users WHERE id = $1", 1)
var id int
var name, email string
err = row.Scan(&id, &name, &email)

// Count
count, err := helper.Count(ctx, "SELECT COUNT(*) FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}

// Exists
exists, err := helper.Exists(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", "john@example.com")
if err != nil {
    log.Fatal(err)
}
```

### 6. Batch Operations with Helper

```go
// Batch Insert
queries := []string{
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "INSERT INTO users (name, email) VALUES ($1, $2)",
}

args := [][]interface{}{
    {"John", "john@example.com"},
    {"Jane", "jane@example.com"},
    {"Bob", "bob@example.com"},
}

err := helper.BatchInsert(ctx, queries, args)
if err != nil {
    log.Fatal(err)
}

// Batch Update
queries := []string{
    "UPDATE users SET active = $1 WHERE id = $2",
    "UPDATE users SET active = $1 WHERE id = $2",
}

args := [][]interface{}{
    {true, 1},
    {false, 2},
}

err = helper.BatchUpdate(ctx, queries, args)
if err != nil {
    log.Fatal(err)
}

// Batch Delete
queries := []string{
    "DELETE FROM users WHERE id = $1",
    "DELETE FROM users WHERE id = $1",
}

args := [][]interface{}{
    {1},
    {2},
}

err = helper.BatchDelete(ctx, queries, args)
if err != nil {
    log.Fatal(err)
}
```

### 7. Health Check and Statistics

```go
// Health check
err := db.HealthCheck()
if err != nil {
    log.Fatal("Database health check failed:", err)
}

// Get pool statistics
stats := db.Stat()
fmt.Printf("Total connections: %d\n", stats.TotalConns())
fmt.Printf("Idle connections: %d\n", stats.IdleConns())
fmt.Printf("Acquired connections: %d\n", stats.AcquiredConns())
fmt.Printf("Constructed connections: %d\n", stats.ConstructedConns())

// Get pool configuration
config := db.Config()
fmt.Printf("Max connections: %d\n", config.MaxConns)
fmt.Printf("Min connections: %d\n", config.MinConns)
```

### 8. Connection Pool Management

```go
// Acquire a connection from the pool
conn, err := db.AcquireConn(ctx)
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Use the connection
var name string
err = conn.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)
if err != nil {
    log.Fatal(err)
}
```

### 9. Copy Operations

```go
// Copy data from CSV to table
copyCount, err := db.CopyFrom(
    ctx,
    []string{"users"}, // table name
    []string{"name", "email"}, // column names
    pgx.CopyFromRows([][]interface{}{
        {"John", "john@example.com"},
        {"Jane", "jane@example.com"},
        {"Bob", "bob@example.com"},
    }),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Copied %d rows\n", copyCount)
```

## Configuration Options

### PGXConfig

```go
type PGXConfig struct {
    DB_HOST                string        // Database host
    DB_PORT                string        // Database port
    DB_USER                string        // Database user
    DB_PASSWORD            string        // Database password
    DB_NAME                string        // Database name
    DB_MAX_CONNS           int           // Maximum connections in pool
    DB_MIN_CONNS           int           // Minimum connections in pool
    DB_MAX_CONN_LIFETIME   time.Duration // Maximum connection lifetime
    DB_MAX_CONN_IDLE_TIME  time.Duration // Maximum idle time
    DB_HEALTH_CHECK_PERIOD time.Duration // Health check period
}
```

### Recommended Settings

```go
config := sql.PGXConfig{
    DB_HOST:                "localhost",
    DB_PORT:                "5432",
    DB_USER:                "postgres",
    DB_PASSWORD:            "password",
    DB_NAME:                "mydb",
    DB_MAX_CONNS:           10,                    // 10 connections max
    DB_MIN_CONNS:           2,                     // 2 connections min
    DB_MAX_CONN_LIFETIME:   5 * time.Minute,      // 5 minutes
    DB_MAX_CONN_IDLE_TIME:  1 * time.Minute,      // 1 minute
    DB_HEALTH_CHECK_PERIOD: 30 * time.Second,     // 30 seconds
}
```

## Error Handling

```go
// Check for specific errors
if err != nil {
    if pgErr, ok := err.(*pgconn.PgError); ok {
        switch pgErr.Code {
        case "23505": // unique_violation
            fmt.Println("Duplicate entry")
        case "23503": // foreign_key_violation
            fmt.Println("Foreign key constraint violation")
        case "23502": // not_null_violation
            fmt.Println("Null value not allowed")
        default:
            fmt.Printf("Database error: %s\n", pgErr.Message)
        }
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Best Practices

### 1. Always use context

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

_, err := db.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
```

### 2. Use parameterized queries

```go
// Good - prevents SQL injection
_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", name, email)

// Bad - vulnerable to SQL injection
_, err := db.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email) VALUES ('%s', '%s')", name, email))
```

### 3. Close rows and transactions

```go
rows, err := db.Query(ctx, "SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close() // Always close rows

// For transactions
tx, err := db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx) // Rollback if not committed
```

### 4. Use transactions for multiple operations

```go
err := db.Transaction(ctx, func(tx pgx.Tx) error {
    // Multiple operations in one transaction
    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err
    }
    
    _, err = tx.Exec(ctx, "UPDATE counters SET user_count = user_count + 1")
    if err != nil {
        return err
    }
    
    return nil
})
```

### 5. Monitor connection pool

```go
// Check pool statistics periodically
stats := db.Stat()
if stats.AcquiredConns() > stats.TotalConns()*8/10 {
    log.Warn("Connection pool is getting full")
}
```

## Performance Tips

1. **Use connection pooling** - Let the pool manage connections
2. **Use batch operations** - For multiple similar operations
3. **Use transactions** - For related operations
4. **Use appropriate indexes** - For better query performance
5. **Monitor pool statistics** - To optimize pool size
6. **Use context timeouts** - To prevent hanging queries
7. **Use prepared statements** - For repeated queries

## Troubleshooting

### Common Issues

1. **Connection timeout**
   - Check network connectivity
   - Increase connection timeout
   - Check firewall settings

2. **Pool exhaustion**
   - Increase max connections
   - Check for connection leaks
   - Use connection timeouts

3. **Query timeout**
   - Use context with timeout
   - Optimize slow queries
   - Add appropriate indexes

4. **Memory leaks**
   - Always close rows and transactions
   - Use defer statements
   - Monitor memory usage

## Examples

ดูตัวอย่างการใช้งานเพิ่มเติมใน `examples/` directory 