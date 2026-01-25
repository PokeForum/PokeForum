package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Generate trace ID | 生成链路ID
		traceID := tracing.GenerateTraceID()

		// Store trace ID in context for subsequent use | 将链路ID存储到context中，方便后续使用
		ctx.Request = ctx.Request.WithContext(tracing.WithTraceID(ctx.Request.Context(), traceID))

		// Set trace ID in response header to return to client | 在响应header中设置链路ID，返回给客户端
		ctx.Header(tracing.TraceIDHeader, traceID)

		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = bodyLogWriter

		// Start time | 开始时间
		startTime := time.Now()

		b, err := ctx.Copy().GetRawData()
		if err != nil {
			b = []byte{}
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(b))

		// Process request | 处理请求
		ctx.Next()

		// End time | 结束时间
		endTime := time.Now()

		// Include trace ID in logs for request tracking | 在日志中包含链路ID，方便追踪请求
		configs.Log.Info("Request",
			zap.String("trace_id", traceID),
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
