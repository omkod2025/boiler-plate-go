package sql

import (
	"context"
	"encoding/json"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbTimeout = 10 * time.Second

// SelectRows: query ข้อมูลแบบหลายแถว
func (p *PGX) QueryRows(query string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(p.ctx, dbTimeout)
	defer cancel()
	return p.pool.Query(ctx, query, args...)
}

// SelectRow: query ข้อมูลแบบแถวเดียว
func (p *PGX) QueryRow(query string, args ...any) pgx.Row {
	ctx, cancel := context.WithTimeout(p.ctx, dbTimeout)
	defer cancel()
	return p.pool.QueryRow(ctx, query, args...)
}

// CallProcedure: เรียก stored procedure (หรือ function ที่ return rows)
func (p *PGX) CallProcedure(procName string, result any, args ...any) error {
	outJSON := ""
	sql := fmt.Sprintf("CALL %s(%s,null)", procName, placeholders(len(args)))
	logger.Info("CallProcedure", sql)
	row := p.QueryRow(sql, args...)
	err := row.Scan(&outJSON)
	if err != nil {
		logger.Error("error", err)
		return err
	}
	err = json.Unmarshal([]byte(outJSON), result)
	if err != nil {
		logger.Error("error", err)
		return err
	}
	return nil
}

// Function: เรียก stored procedure (หรือ function ที่ return rows)
func (p *PGX) CallFunction(funcName string, args ...any) (pgx.Rows, error) {
	placeholders := ""
	for i := range args {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?" + strconv.Itoa(i+1)
	}
	query := "SELECT " + funcName
	if len(args) > 0 {
		query += "(" + placeholders + ")"
	}
	return p.QueryRows(query, args...)
}

// BeginTx: เริ่ม transaction ใหม่
func (p *PGX) BeginTx() (pgx.Tx, error) {
	ctx, cancel := context.WithTimeout(p.ctx, dbTimeout)
	defer cancel()
	return p.pool.Begin(ctx)
}

// Exec: execute query ที่ไม่ return rows (INSERT, UPDATE, DELETE)
func (p *PGX) Exec(sql string, arguments ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(p.ctx, dbTimeout)
	defer cancel()
	return p.pool.Exec(ctx, sql, arguments...)
}

// ExecWithContext: execute query พร้อม custom context
func (p *PGX) ExecWithContext(ctx context.Context, sql string, arguments ...interface{}) (interface{}, error) {
	return p.pool.Exec(ctx, sql, arguments...)
}

// Query: query ข้อมูลแบบหลายแถว พร้อม context
func (p *PGX) Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, arguments...)
}

// QueryWithContext: query ข้อมูลแบบหลายแถว พร้อม custom context
func (p *PGX) QueryWithContext(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, arguments...)
}

// QueryRowWithContext: query ข้อมูลแบบแถวเดียว พร้อม custom context
func (p *PGX) QueryRowWithContext(ctx context.Context, sql string, arguments ...interface{}) pgx.Row {
	return p.pool.QueryRow(ctx, sql, arguments...)
}

// Begin: เริ่ม transaction ใหม่ พร้อม context
func (p *PGX) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// BeginTxWithOptions: เริ่ม transaction ใหม่ พร้อม options
func (p *PGX) BeginTxWithOptions(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.pool.BeginTx(ctx, txOptions)
}

// ExecBatch: execute multiple queries ใน batch
func (p *PGX) ExecBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	return p.pool.SendBatch(ctx, batch)
}

// CopyFrom: copy ข้อมูลจาก reader ไปยัง table
func (p *PGX) CopyFrom(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return p.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// AcquireConn: acquire connection จาก pool
func (p *PGX) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	return p.pool.Acquire(ctx)
}

// Stat: ดูสถิติ connection pool
func (p *PGX) Stat() *pgxpool.Stat {
	return p.pool.Stat()
}

// Config: ดู configuration ของ pool
func (p *PGX) Config() *pgxpool.Config {
	return p.pool.Config()
}

// GetPool: ดู underlying pool
func (p *PGX) GetPool() *pgxpool.Pool {
	return p.pool
}

func placeholders(n int) string {
	s := ""
	for i := 1; i <= n; i++ {
		if i > 1 {
			s += ", "
		}
		s += fmt.Sprintf("$%d", i)
	}
	return s
}

