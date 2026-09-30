package amqp

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/omkod2025/boiler-plate-go/pkg/testhelper"
)

func TestPublishConsumeAckAndReject(t *testing.T) {
	url := testhelper.RabbitMQ(t)
	conn, err := Dial(url, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// queue + dead-letter ของ test (ใน production ประกาศจาก definitions ของ broker)
	ch, err := conn.Channel(ctx)
	if err != nil {
		t.Fatal(err)
	}
	q := "test.publish-consume"
	if _, err := ch.QueueDeclare(q+".dlq", false, true, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := ch.QueueDeclare(q, false, true, false, false, amqp091.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": q + ".dlq"}); err != nil {
		t.Fatal(err)
	}

	pub := &Publisher{Conn: conn}
	for _, body := range []string{`{"ok":true}`, `bad`} {
		if err := pub.Publish(ctx, "", q, body, []byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	var mu sync.Mutex
	got := map[string]int{}
	c := &Consumer{Conn: conn, Queue: q, Prefetch: 2, Handler: func(_ context.Context, d amqp091.Delivery) error {
		mu.Lock()
		got[string(d.Body)]++
		mu.Unlock()
		if string(d.Body) == "bad" {
			return Permanent(errors.New("invalid"))
		}
		return nil
	}}
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx) }()

	deadline := time.Now().Add(15 * time.Second)
	for {
		dlq, err := ch.QueueDeclarePassive(q+".dlq", false, true, false, false, nil)
		mu.Lock()
		n := got[`{"ok":true}`]
		mu.Unlock()
		if err == nil && dlq.Messages == 1 && n == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("got %v, dlq %+v %v", got, dlq, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}
