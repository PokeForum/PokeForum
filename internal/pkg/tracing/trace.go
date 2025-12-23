package tracing

import (
	"context"
	"strconv"

	"github.com/click33/sa-token-go/stputil"
	"github.com/google/uuid"
)

// TraceID 是链路追踪的唯一标识
type TraceID string

// contextKey 用于在context中存储TraceID
type contextKey string

const (
	// traceIDKey 是存储TraceID的context key
	traceIDKey contextKey = "trace_id"
	// userIDKey 是存储用户ID的context key
	userIDKey contextKey = "user_id"
	// TraceIDHeader 是返回给客户端的Header名称
	TraceIDHeader = "X-Trace-ID"
)

// GenerateTraceID 生成一个新的链路ID
// 使用UUID v4生成唯一的链路ID
func GenerateTraceID() string {
	return uuid.New().String()
}

// WithTraceID 将TraceID存储到context中
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID 从context中获取TraceID
// 如果context中不存在TraceID，返回空字符串
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

// WithUserID 将用户ID存储到context中
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID 从context中获取用户ID
// 如果context中不存在用户ID，返回0
func GetUserID(ctx context.Context) int {
	userID, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}

// ContextWithUserID 从gin.Context获取用户ID并设置到context.Context中
// 如果用户未登录或获取失败，返回原始context
func ContextWithUserID(ginCtx interface{}, ctx context.Context) context.Context {
	// 尝试从gin.Context获取Authorization token
	var token string

	// 使用类型断言来检查是否是gin.Context
	if c, ok := ginCtx.(interface{ GetHeader(string) string }); ok {
		token = c.GetHeader("Authorization")
	}

	if token == "" {
		return ctx
	}

	// 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return ctx
	}

	// String转Int
	userID, err := strconv.Atoi(loginID)
	if err != nil {
		return ctx
	}

	return WithUserID(ctx, userID)
}
