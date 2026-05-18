package middleware

import (
	"context"
	"sync"
	"time"
)

// slidingWindow Sliding window rate limiter | 滑动窗口限流器
type slidingWindow struct {
	mu         sync.Mutex    // Mutex for thread safety | 互斥锁，保证线程安全
	requests   []int64       // Request timestamps | 请求时间戳列表
	windowSize time.Duration // Window size | 窗口大小
}

// newSlidingWindow Create a new sliding window | 创建新的滑动窗口
func newSlidingWindow(windowSize time.Duration) *slidingWindow {
	return &slidingWindow{
		requests:   make([]int64, 0), // Initialize request list | 初始化请求列表
		windowSize: windowSize,
	}
}

// checkAndAdd Check limit and add current request atomically | 原子检查限流并添加当前请求
func (sw *slidingWindow) checkAndAdd(now int64, maxCount int) (bool, int) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	windowStart := now - int64(sw.windowSize) // Calculate window start time | 计算窗口开始时间

	validRequests := make([]int64, 0, len(sw.requests))
	for _, req := range sw.requests {
		if req > windowStart {
			validRequests = append(validRequests, req) // Keep valid requests | 保留有效请求
		}
	}
	sw.requests = validRequests // Update request list | 更新请求列表

	currentCount := len(sw.requests)
	if currentCount >= maxCount {
		return false, currentCount
	}
	sw.requests = append(sw.requests, now) // Add current request | 添加当前请求
	return true, currentCount
}

// memoryRateLimiter Memory-based rate limiter | 基于内存的限流器
type memoryRateLimiter struct {
	mu      sync.RWMutex              // Read-write mutex | 读写互斥锁
	windows map[string]*slidingWindow // Sliding windows for each key | 每个键的滑动窗口
}

// newMemoryRateLimiter Create a new memory rate limiter | 创建新的内存限流器
func newMemoryRateLimiter() *memoryRateLimiter {
	return &memoryRateLimiter{
		windows: make(map[string]*slidingWindow), // Initialize windows map | 初始化窗口映射
	}
}

// check Check if request is allowed | 检查请求是否允许
func (mrl *memoryRateLimiter) check(key string, config RateLimitConfig) (bool, int, int64, error) {
	nowTime := time.Now()
	now := nowTime.UnixNano()
	windowSize := time.Duration(config.WindowSize) * time.Second
	resetTime := nowTime.Add(windowSize).Unix()

	mrl.mu.RLock()
	window, exists := mrl.windows[key] // Get existing window | 获取已存在的窗口
	mrl.mu.RUnlock()

	if !exists {
		mrl.mu.Lock()
		window, exists = mrl.windows[key] // Double check after lock | 加锁后再次检查
		if !exists {
			window = newSlidingWindow(windowSize) // Create new window | 创建新窗口
			mrl.windows[key] = window
		}
		mrl.mu.Unlock()
	}

	allowed, currentCount := window.checkAndAdd(now, config.MaxRequests)
	remaining := config.MaxRequests - currentCount - 1
	if remaining < 0 {
		remaining = 0 // Ensure remaining is not negative | 确保剩余数不为负
	}

	if !allowed {
		return false, 0, resetTime, nil // Rate limit exceeded | 超过限流
	}

	return true, remaining, resetTime, nil
}

var globalMemoryRateLimiter = newMemoryRateLimiter() // Global memory rate limiter | 全局内存限流器

// checkMemoryRateLimit Check rate limit using memory limiter | 使用内存限流器检查限流
func checkMemoryRateLimit(_ context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	return globalMemoryRateLimiter.check(key, config)
}
