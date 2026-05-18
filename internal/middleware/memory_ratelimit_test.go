package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryRateLimitUsesConfigMaxRequests(t *testing.T) {
	limiter := newMemoryRateLimiter()
	config := RateLimitConfig{
		WindowSize:  60,
		MaxRequests: 2,
		KeyPrefix:   "test",
	}

	for i := 0; i < config.MaxRequests; i++ {
		allowed, _, _, err := limiter.check("test:key", config)
		if err != nil {
			t.Fatalf("检查内存限流失败: %v", err)
		}
		if !allowed {
			t.Fatalf("第 %d 次请求应被允许", i+1)
		}
	}

	allowed, remaining, resetTime, err := limiter.check("test:key", config)
	if err != nil {
		t.Fatalf("检查内存限流失败: %v", err)
	}
	if allowed {
		t.Fatal("超过配置上限后请求应被拒绝")
	}
	if remaining != 0 {
		t.Fatalf("剩余次数应为 0，实际为 %d", remaining)
	}
	if resetTime <= time.Now().Unix() {
		t.Fatalf("重置时间应为 Unix 秒级未来时间，实际为 %d", resetTime)
	}
}

func TestMemoryRateLimitConcurrentRequests(t *testing.T) {
	limiter := newMemoryRateLimiter()
	config := RateLimitConfig{
		WindowSize:  60,
		MaxRequests: 10,
		KeyPrefix:   "test",
	}

	var allowedCount int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _, _, err := limiter.check("test:concurrent", config)
			if err != nil {
				t.Errorf("检查内存限流失败: %v", err)
				return
			}
			if allowed {
				atomic.AddInt64(&allowedCount, 1)
			}
		}()
	}
	wg.Wait()

	if allowedCount != int64(config.MaxRequests) {
		t.Fatalf("并发允许次数应等于配置上限 %d，实际为 %d", config.MaxRequests, allowedCount)
	}
}

func TestCheckMemoryRateLimitUsesGlobalLimiterConfig(t *testing.T) {
	oldLimiter := globalMemoryRateLimiter
	globalMemoryRateLimiter = newMemoryRateLimiter()
	defer func() {
		globalMemoryRateLimiter = oldLimiter
	}()

	config := RateLimitConfig{
		WindowSize:  60,
		MaxRequests: 1,
		KeyPrefix:   "test",
	}

	allowed, _, _, err := checkMemoryRateLimit(context.Background(), "test:global", config)
	if err != nil {
		t.Fatalf("检查内存限流失败: %v", err)
	}
	if !allowed {
		t.Fatal("首次请求应被允许")
	}

	allowed, _, _, err = checkMemoryRateLimit(context.Background(), "test:global", config)
	if err != nil {
		t.Fatalf("检查内存限流失败: %v", err)
	}
	if allowed {
		t.Fatal("第二次请求应按配置上限被拒绝")
	}
}
