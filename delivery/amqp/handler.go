// Package amqp คือ delivery ของ role worker: แปลงข้อความจาก RabbitMQ เป็น input ของ use case
// เช่นเดียวกับ HTTP handler — business logic อยู่ใน domain/<feature> เสมอ
package amqp

import (
	"context"
	"encoding/json"

	amqp091 "github.com/rabbitmq/amqp091-go"

	pkgamqp "github.com/omkod2025/boiler-plate-go/pkg/amqp"
	"github.com/omkod2025/boiler-plate-go/pkg/logger"
)

// Message คือรูปแบบข้อความตัวอย่าง — ใช้ type ที่ generate จาก contract ของ event จริงแทน
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NewHandler คืน handler ตัวอย่าง: decode ข้อความ, JSON ผิดรูป = permanent (ส่งไป dead-letter)
// เพิ่ม use case ที่ต้องใช้เป็น parameter แล้วเรียกตาม Type
func NewHandler() pkgamqp.Handler {
	return func(ctx context.Context, d amqp091.Delivery) error {
		var m Message
		if err := json.Unmarshal(d.Body, &m); err != nil {
			return pkgamqp.Permanent(err)
		}
		logger.Info("worker: received " + m.Type + " (" + d.MessageId + ")")
		return nil
	}
}
