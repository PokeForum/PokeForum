package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get trace ID from upstream if present, otherwise generate a new one | 获取上游trace_id，否则生成新的trace_id
		traceID := ctx.GetHeader(tracing.TraceIDHeader)
		if traceID == "" {
			traceID = tracing.GenerateTraceID()
		}

		// Store trace_id/user_id in context for subsequent use | 将trace_id/user_id存储在context中，供后续使用
		reqCtx := tracing.WithTraceID(ctx.Request.Context(), traceID)
		reqCtx = tracing.ContextWithUserID(ctx, reqCtx)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		// Set trace ID in response header to return to client | 将trace_id设置在response header中，返回给客户端
		ctx.Header(tracing.TraceIDHeader, traceID)

		// Start time | 开始时间
		startTime := time.Now()

		// Process request | 处理请求
		ctx.Next()

		// End time | 结束时间
		endTime := time.Now()
		userID := tracing.GetUserID(ctx.Request.Context())

		// Include trace ID in logs for request tracking | 将trace_id/user_id包含在日志中，用于请求跟踪
		configs.Log.Info("Request",
			zap.String("trace_id", traceID),
			zap.Int("user_id", userID),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("method", ctx.Request.Method),
			zap.String("url", ctx.Request.URL.String()),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("request_time", TimeFormat(startTime)),
			zap.String("response_time", TimeFormat(endTime)),
			zap.String("cost_time", endTime.Sub(startTime).String()),
		)
	}
}

func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
