package middleware

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
)

// RateLimitConfig Rate limit configuration | 速率限制配置
type RateLimitConfig struct {
	// Time window size (seconds) | 时间窗口大小（秒）
	WindowSize int
	// Maximum requests within window | 窗口内最大请求数
	MaxRequests int
	// Rate limit key prefix | 限流键前缀
	KeyPrefix string
}

// DefaultRateLimitConfig Default configuration: 100 requests per second | 默认配置：每秒100个请求
var DefaultRateLimitConfig = RateLimitConfig{
	WindowSize:  _const.DefaultTimeWindowSeconds,
	MaxRequests: _const.DefaultMaxRequests,
	KeyPrefix:   "ratelimit:global",
}

// APIRateLimitConfig API interface rate limit configuration: 60 requests per minute | API接口限流配置：每分钟60个请求
var APIRateLimitConfig = RateLimitConfig{
	WindowSize:  _const.APITimeWindowSeconds,
	MaxRequests: _const.APIMaxRequests,
	KeyPrefix:   "ratelimit:api",
}

// AuthRateLimitConfig Authentication interface rate limit configuration: 10 times per minute (prevent brute force) | 认证接口限流配置：每分钟10次（防止暴力破解）
var AuthRateLimitConfig = RateLimitConfig{
	WindowSize:  _const.AuthTimeWindowSeconds,
	MaxRequests: _const.AuthMaxRequests,
	KeyPrefix:   "ratelimit:auth",
}

var rateLimitMemberSeq uint64

// RateLimit Redis-based sliding window rate limit middleware | 基于Redis的滑动窗口速率限制中间件
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP as rate limit key | 获取客户端IP作为限流键
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)

		// Check if rate limit is exceeded | 检查是否超过速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			// Fallback to memory-based rate limit when Redis fails | Redis失败时降级到内存限流
			configs.Log.Warn("Redis速率限制检查失败，降级到内存限流",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("client_ip", clientIP),
				zap.Error(err),
			)
			allowed, remaining, resetTime, err = checkMemoryRateLimit(c.Request.Context(), key, config)
			if err != nil {
				configs.Log.Error("内存速率限制检查失败，放行请求",
					zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
					zap.String("client_ip", clientIP),
					zap.Error(err),
				)
				c.Next()
				return
			}
		}

		// Set rate limit related response headers | 设置速率限制相关响应头
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

			// Set Retry-After response header | 设置 Retry-After 响应头
			c.Header("Retry-After", fmt.Sprintf("%d", config.WindowSize))
			response.ResError(c, response.CodeTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit Check rate limit (using Redis sliding window algorithm) | 检查速率限制（使用Redis滑动窗口算法）
// Returns: whether request is allowed, remaining requests, reset timestamp, error | 返回: 是否允许请求、剩余请求数、重置时间戳、错误
func checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-time.Duration(config.WindowSize) * time.Second)
	resetTime := now.Add(time.Duration(config.WindowSize) * time.Second).Unix()
	member := fmt.Sprintf("%d:%d", now.UnixNano(), atomic.AddUint64(&rateLimitMemberSeq, 1))

	luaScript := `
redis.call("ZREMRANGEBYSCORE", KEYS[1], "0", ARGV[1])
local count = redis.call("ZCARD", KEYS[1])
local max_requests = tonumber(ARGV[3])
local reset_time = tonumber(ARGV[5])
if count >= max_requests then
	redis.call("EXPIRE", KEYS[1], ARGV[4])
	return {0, 0, reset_time}
end
redis.call("ZADD", KEYS[1], ARGV[2], ARGV[6])
redis.call("EXPIRE", KEYS[1], ARGV[4])
return {1, max_requests - count - 1, reset_time}
`

	result, err := configs.Cache.Eval(ctx, luaScript, []string{key},
		windowStart.UnixNano(),
		now.UnixNano(),
		config.MaxRequests,
		config.WindowSize*2,
		resetTime,
		member,
	).Result()
	if err != nil {
		return false, 0, resetTime, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return false, 0, resetTime, fmt.Errorf("invalid rate limit script result: %v", result)
	}

	allowed := toInt64(values[0]) == 1
	remaining := int(toInt64(values[1]))
	resetTime = toInt64(values[2])
	if remaining < 0 {
		remaining = 0
	}
	return allowed, remaining, resetTime, nil
}

func toInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		result, _ := strconv.ParseInt(v, 10, 64) //nolint:errcheck // 解析失败按0处理
		return result
	default:
		return 0
	}
}

// RateLimitByKey Custom key rate limit (for more fine-grained control) | 自定义键的速率限制（用于更细粒度的控制）
// keyFunc: Custom key generation function, parameter is gin.Context, returns rate limit key | keyFunc: 自定义键生成函数，参数为gin.Context，返回限流键
func RateLimitByKey(config RateLimitConfig, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use custom function to generate rate limit key | 使用自定义函数生成限流键
		customKey := keyFunc(c)
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, customKey)

		// Check if rate limit is exceeded | 检查是否超过速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			// Fallback to memory-based rate limit when Redis fails | Redis失败时降级到内存限流
			configs.Log.Warn("Redis速率限制检查失败，降级到内存限流",
				zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
				zap.String("key", key),
				zap.Error(err),
			)
			allowed, remaining, resetTime, err = checkMemoryRateLimit(c.Request.Context(), key, config)
			if err != nil {
				configs.Log.Error("内存速率限制检查失败，放行请求",
					zap.String("trace_id", tracing.GetTraceID(c.Request.Context())),
					zap.String("key", key),
					zap.Error(err),
				)
				c.Next()
				return
			}
		}

		// Set rate limit related response headers | 设置速率限制相关响应头
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
