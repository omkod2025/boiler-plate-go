// Package amqp คือ client RabbitMQ (amqp091-go) ที่ role worker ใช้:
//
//   - Conn ต่อใหม่เองเมื่อหลุด (backoff + jitter สูงสุด 30 วินาที)
//   - Publisher รอ publisher confirm ทุกข้อความ — ได้ nil เมื่อ broker เก็บข้อความแล้วจริงเท่านั้น
//   - Consumer ack ด้วยมือหลัง handler ทำงานเสร็จ (commit DB แล้ว), จำกัด prefetch
//
// queue/exchange ประกาศไว้ฝั่ง broker (definitions) ไม่ประกาศจาก service
package amqp

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/omkod2025/boiler-plate-go/pkg/logger"
)

// ErrClosed ได้เมื่อเรียกหลัง Close
var ErrClosed = errors.New("amqp: connection closed")

// Conn คือ connection ที่ต่อใหม่เอง
type Conn struct {
	url  string
	name string

	mu      sync.Mutex
	conn    *amqp091.Connection
	ready   chan struct{}
	closing bool
}

// Dial ต่อครั้งแรกแล้ว watch เพื่อต่อใหม่เองเมื่อหลุด — url แบบ amqps:// ใช้ TLS
func Dial(url, name string) (*Conn, error) {
	c := &Conn{url: url, name: name, ready: make(chan struct{})}
	conn, err := c.dial()
	if err != nil {
		return nil, err
	}
	c.set(conn)
	go c.watch()
	return c, nil
}

func (c *Conn) dial() (*amqp091.Connection, error) {
	props := amqp091.NewConnectionProperties()
	props.SetClientConnectionName(c.name)
	conn, err := amqp091.DialConfig(c.url, amqp091.Config{Properties: props, Heartbeat: 10 * time.Second, Locale: "en_US"})
	if err != nil {
		return nil, errors.New("amqp: dial failed") // error เดิมอาจมี URL ที่มีรหัสผ่าน
	}
	return conn, nil
}

func (c *Conn) set(conn *amqp091.Connection) {
	c.mu.Lock()
	c.conn = conn
	close(c.ready)
	c.mu.Unlock()
}

func (c *Conn) watch() {
	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		<-conn.NotifyClose(make(chan *amqp091.Error, 1))
		c.mu.Lock()
		if c.closing {
			c.mu.Unlock()
			return
		}
		c.ready = make(chan struct{})
		c.mu.Unlock()
		logger.Warn("amqp: connection lost; reconnecting")
		backoff := 250 * time.Millisecond
		for {
			c.mu.Lock()
			closing := c.closing
			c.mu.Unlock()
			if closing {
				return
			}
			if conn, err := c.dial(); err == nil {
				logger.Info("amqp: reconnected")
				c.set(conn)
				break
			}
			time.Sleep(backoff + time.Duration(rand.Int64N(int64(backoff/2)+1))) // #nosec G404 -- jitter
			backoff = min(backoff*2, 30*time.Second)
		}
	}
}

// Channel เปิด channel ใหม่ รอจนต่อได้หรือ ctx หมด
func (c *Conn) Channel(ctx context.Context) (*amqp091.Channel, error) {
	for {
		c.mu.Lock()
		if c.closing {
			c.mu.Unlock()
			return nil, ErrClosed
		}
		ready, conn := c.ready, c.conn
		c.mu.Unlock()
		select {
		case <-ready:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		ch, err := conn.Channel()
		if err == nil {
			return ch, nil
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Ping ใช้เป็น readiness check
func (c *Conn) Ping(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.conn.IsClosed() {
		return errors.New("amqp: not connected")
	}
	return nil
}

// Close หยุดต่อใหม่และปิด connection
func (c *Conn) Close() error {
	c.mu.Lock()
	c.closing = true
	conn := c.conn
	c.mu.Unlock()
	if err := conn.Close(); err != nil && !errors.Is(err, amqp091.ErrClosed) {
		return err
	}
	return nil
}

// Publisher ส่งข้อความ persistent และรอ confirm (ปลอดภัยต่อการเรียกพร้อมกัน)
type Publisher struct {
	Conn *Conn

	mu sync.Mutex
	ch *amqp091.Channel
}

// ErrNacked คือ broker ปฏิเสธข้อความ (เช่น queue เต็ม) — ถือว่ายังส่งไม่สำเร็จ
var ErrNacked = errors.New("amqp: broker nacked the message")

// Publish ส่งแล้วรอ confirm ไม่เกิน 5 วินาที
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey, messageID string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch == nil || p.ch.IsClosed() {
		ch, err := p.Conn.Channel(ctx)
		if err != nil {
			return err
		}
		if err := ch.Confirm(false); err != nil {
			_ = ch.Close()
			return fmt.Errorf("amqp: confirm mode: %w", err)
		}
		p.ch = ch
	}
	conf, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, false, false, amqp091.Publishing{
		ContentType: "application/json", DeliveryMode: amqp091.Persistent, MessageId: messageID, Timestamp: time.Now().UTC(), Body: body,
	})
	if err != nil {
		_ = p.ch.Close()
		p.ch = nil
		return fmt.Errorf("amqp: publish: %w", err)
	}
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ok, err := conf.WaitContext(wctx)
	if err != nil {
		_ = p.ch.Close()
		p.ch = nil
		return fmt.Errorf("amqp: confirm: %w", err)
	}
	if !ok {
		return ErrNacked
	}
	return nil
}

// Handler ประมวลผลข้อความหนึ่งข้อความ — คืน nil หลัง commit DB แล้วเท่านั้น
type Handler func(ctx context.Context, d amqp091.Delivery) error

// Permanent ห่อ error ที่ลองใหม่ก็ไม่หาย ข้อความจะถูก reject (ไป dead-letter ของ queue) แทนการ requeue
func Permanent(err error) error { return permanentError{err} }

type permanentError struct{ error }

func (p permanentError) Unwrap() error { return p.error }

// Consumer รับจาก queue เดียว ack ด้วยมือ
type Consumer struct {
	Conn     *Conn
	Queue    string
	Prefetch int // ค่าเริ่มต้น 10
	Handler  Handler

	wg sync.WaitGroup
}

// Run รับข้อความจนกว่า ctx จะถูก cancel แล้วรอ handler ที่ทำงานอยู่ให้จบ; subscribe ใหม่เมื่อ connection หลุด
func (c *Consumer) Run(ctx context.Context) error {
	defer c.wg.Wait()
	for ctx.Err() == nil {
		err := c.consume(ctx)
		if ctx.Err() != nil {
			return nil
		}
		logger.Warn("amqp: consumer of " + c.Queue + " stopped; resubscribing: " + err.Error())
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
		}
	}
	return nil
}

func (c *Consumer) consume(ctx context.Context) error {
	ch, err := c.Conn.Channel(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()
	prefetch := c.Prefetch
	if prefetch <= 0 {
		prefetch = 10
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}
	msgs, err := ch.ConsumeWithContext(ctx, c.Queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	sem := make(chan struct{}, prefetch)
	for d := range msgs {
		sem <- struct{}{}
		c.wg.Add(1)
		go func() {
			defer func() { <-sem; c.wg.Done() }()
			// handler ทำงานจนจบแม้ระหว่างปิดโปรแกรม เพื่อไม่ทิ้งงานครึ่งทาง
			err := c.Handler(context.WithoutCancel(ctx), d)
			var perm permanentError
			switch {
			case err == nil:
				_ = d.Ack(false)
			case errors.As(err, &perm):
				logger.Error("amqp: permanent failure; rejecting message " + d.MessageId + ": " + err.Error())
				_ = d.Nack(false, false)
			default:
				logger.Warn("amqp: transient failure; requeue message " + d.MessageId + ": " + err.Error())
				_ = d.Nack(false, true)
			}
		}()
	}
	return errors.New("delivery channel closed")
}
