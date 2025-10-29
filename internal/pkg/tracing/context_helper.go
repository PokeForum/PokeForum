package tracing

import (
	"context"

	"go.uber.org/zap"
)

// WithTraceIDField 返回一个zap字段，包含从context中获取的TraceID
// 用于在日志中自动包含链路ID
func WithTraceIDField(ctx context.Context) zap.Field {
	traceID := GetTraceID(ctx)
	return zap.String("trace_id", traceID)
}
