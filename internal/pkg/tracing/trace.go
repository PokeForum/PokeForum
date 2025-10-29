package tracing

import (
	"context"

	"github.com/google/uuid"
)

// TraceID 是链路追踪的唯一标识
type TraceID string

// contextKey 用于在context中存储TraceID
type contextKey string

const (
	// traceIDKey 是存储TraceID的context key
	traceIDKey contextKey = "trace_id"
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
