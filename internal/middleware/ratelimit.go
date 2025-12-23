package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 时间窗口大小（秒）
	WindowSize int
	// 窗口内最大请求数
	MaxRequests int
	// 限流键前缀
	KeyPrefix string
}

// DefaultRateLimitConfig 默认配置：每秒100个请求
var DefaultRateLimitConfig = RateLimitConfig{
	WindowSize:  1,
	MaxRequests: 100,
	KeyPrefix:   "ratelimit:global",
}

// APIRateLimitConfig API接口限流配置：每分钟60个请求
var APIRateLimitConfig = RateLimitConfig{
	WindowSize:  60,
	MaxRequests: 60,
	KeyPrefix:   "ratelimit:api",
}

// AuthRateLimitConfig 认证接口限流配置：每分钟10次（防止暴力破解）
var AuthRateLimitConfig = RateLimitConfig{
	WindowSize:  60,
	MaxRequests: 10,
	KeyPrefix:   "ratelimit:auth",
}

// RateLimit 基于Redis的滑动窗口速率限制中间件
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP作为限流键
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)

		// 检查是否超过速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			// Redis错误时记录日志，但不阻断请求（降级处理）
			configs.Log.Warn("速率限制检查失败，降级放行",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("client_ip", clientIP),
				zap.Error(err),
			)
			c.Next()
			return
		}

		// 设置速率限制相关响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		if !allowed {
			configs.Log.Warn("请求被速率限制拦截",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.Int("window_size", config.WindowSize),
				zap.Int("max_requests", config.MaxRequests),
			)

			// 设置 Retry-After 响应头
			c.Header("Retry-After", fmt.Sprintf("%d", config.WindowSize))
			response.ResError(c, response.CodeTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查速率限制（使用Redis滑动窗口算法）
// 返回: 是否允许请求、剩余请求数、重置时间戳、错误
func checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-time.Duration(config.WindowSize) * time.Second)
	resetTime := now.Add(time.Duration(config.WindowSize) * time.Second).Unix()

	// 使用Redis Pipeline执行原子操作
	pipe := configs.Cache.Pipeline()

	// 移除窗口外的旧记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 获取当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	// 添加当前请求记录（使用纳秒时间戳作为score和member保证唯一性）
	member := fmt.Sprintf("%d", now.UnixNano())
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: member,
	})

	// 设置键的过期时间（窗口大小的2倍，确保数据自动清理）
	pipe.Expire(ctx, key, time.Duration(config.WindowSize*2)*time.Second)

	// 执行Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, resetTime, err
	}

	// 获取当前请求数（在添加新请求之前的数量）
	currentCount := int(countCmd.Val())
	remaining := config.MaxRequests - currentCount - 1
	if remaining < 0 {
		remaining = 0
	}

	// 判断是否超过限制
	if currentCount >= config.MaxRequests {
		return false, 0, resetTime, nil
	}

	return true, remaining, resetTime, nil
}

// RateLimitByKey 自定义键的速率限制（用于更细粒度的控制）
// keyFunc: 自定义键生成函数，参数为gin.Context，返回限流键
func RateLimitByKey(config RateLimitConfig, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用自定义函数生成限流键
		customKey := keyFunc(c)
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, customKey)

		// 检查是否超过速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			configs.Log.Warn("速率限制检查失败，降级放行",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("key", key),
				zap.Error(err),
			)
			c.Next()
			return
		}

		// 设置速率限制相关响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		if !allowed {
			configs.Log.Warn("请求被速率限制拦截",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
			)

			c.Header("Retry-After", fmt.Sprintf("%d", config.WindowSize))
			response.ResError(c, response.CodeTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
