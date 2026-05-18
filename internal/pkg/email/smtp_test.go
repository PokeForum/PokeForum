package email

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestSMTPPool(buffer int) *SMTPPool {
	return &SMTPPool{
		config: SMTPConfig{
			Name:    "PokeForum",
			Address: "noreply@example.com",
		},
		ch:     make(chan *message, buffer),
		chOpen: true,
		ready:  make(chan struct{}),
		logger: zap.NewNop(),
	}
}

func TestSMTPPoolCloseIsIdempotent(t *testing.T) {
	pool := newTestSMTPPool(1)

	pool.Close()
	pool.Close()

	if err := pool.Send(context.Background(), "user@example.com", "title", "body"); err == nil {
		t.Fatal("关闭后的 SMTP 队列应拒绝发送")
	}
}

func TestSMTPPoolConcurrentSendAndCloseDoesNotPanic(t *testing.T) {
	pool := newTestSMTPPool(100)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = pool.Send(ctx, "user@example.com", "title", "body")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		pool.Close()
	}()

	wg.Wait()
	pool.Close()
}
