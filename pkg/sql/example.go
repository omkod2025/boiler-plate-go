package sql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// ExampleUsage แสดงการใช้ package นี้ตามกติกาการเข้าถึงข้อมูล:
//
//   - อ่าน (GET/list/count/exists) → เรียก function ด้วย SELECT ... FROM schema.<fn>($1)
//   - เขียน (INSERT/UPDATE/DELETE, upsert, soft delete) → เรียก procedure ด้วย CALL schema.<proc>(...)
//     ค่าที่ต้องได้กลับ (id, version) ประกาศเป็น OUT parameter แล้วอ่านจากแถวผลของ CALL
//   - runtime ไม่อ่าน/เขียนตารางตรง; SQL ของตารางอยู่ใน migrations/ เท่านั้น
//   - procedure ไม่ COMMIT เอง ให้ caller รวมหลายคำสั่งใน transaction เดียว
//
// routine ที่ใช้ในตัวอย่างอยู่ใน migrations/00002_routines.sql
func ExampleUsage() {
	ctx := context.Background()
	db, err := NewPGX(ctx, PGXConfig{
		DB_HOST: "localhost", DB_PORT: "5432", DB_USER: "postgres", DB_PASSWORD: "password", DB_NAME: "mydb",
		DB_MAX_CONNS: 10, DB_MIN_CONNS: 2, DB_MAX_CONN_LIFETIME: 5 * time.Minute,
		DB_MAX_CONN_IDLE_TIME: time.Minute, DB_HEALTH_CHECK_PERIOD: 30 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to create database connection:", err)
	}
	defer db.Close()

	// เขียน: CALL procedure แล้วอ่าน OUT parameters (ส่ง NULL ในตำแหน่ง OUT)
	var userID int
	var createdAt, updatedAt time.Time
	err = db.QueryRowWithContext(ctx, `CALL public.oktp_user_insert($1, $2, $3, $4, NULL, NULL, NULL)`,
		"john@example.com", "John", "<bcrypt hash>", "user").Scan(&userID, &createdAt, &updatedAt)
	if err != nil {
		log.Printf("insert failed: %v", err)
		return
	}

	// อ่าน: SELECT จาก function
	var email, name string
	err = db.QueryRowWithContext(ctx, `SELECT email, full_name FROM public.oktf_user_get($1)`, userID).Scan(&email, &name)
	if err != nil {
		log.Printf("read failed: %v", err)
		return
	}
	fmt.Println(email, name)

	// หลายคำสั่งใน transaction เดียว: procedure ทุกตัว commit พร้อมกันหรือ rollback พร้อมกัน
	helper := NewDatabaseHelper(db)
	err = helper.Transaction(ctx, func(tx pgx.Tx) error {
		var categoryID int
		if err := tx.QueryRow(ctx, `CALL public.oktp_category_insert($1, $2, $3, $4, $5, NULL, NULL, NULL)`,
			"Food", userID, "#f00", "food", "expense").Scan(&categoryID, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("insert category: %w", err)
		}
		var deleted bool
		if err := tx.QueryRow(ctx, `CALL public.oktp_category_delete($1, $2, NULL)`, categoryID, userID).Scan(&deleted); err != nil {
			return fmt.Errorf("delete category: %w", err)
		}
		return nil
	})
	if err != nil {
		log.Printf("transaction failed: %v", err)
	}
}

// ExampleHealthCheck แสดง health check และสถิติของ pool (ไม่แตะตารางของ domain)
func ExampleHealthCheck(db *PGX) {
	if err := db.HealthCheck(); err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	stats := db.Stat()
	fmt.Printf("connections: total=%d idle=%d acquired=%d\n", stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())
}
