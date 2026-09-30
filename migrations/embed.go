// Package migrations เก็บไฟล์ migration ของ goose (SQL) ฝังใน binary — role migrate รัน `goose up`
// ก่อน deploy role อื่น เพิ่มไฟล์ใหม่ด้วย `make migrate-create NAME=<ชื่อ>`
package migrations

import "embed"

// FS คือไฟล์ *.sql ทั้งหมดในโฟลเดอร์นี้
//
//go:embed *.sql
var FS embed.FS
