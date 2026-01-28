package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
)

// Timeout 请求超时控制中间件
// timeout: 超时时间，默认30秒
func Timeout(timeout time.Duration) gin.HandlerFunc {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return func(c *gin.Context) {
		// 创建带超时的 context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求的 context
		c.Request = c.Request.WithContext(ctx)

		// 创建完成通道
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			c.Abort()
			response.ResErrorWithHTTPStatus(c, http.StatusGatewayTimeout, response.CodeServerBusy, "Request timeout | 请求超时")
			return
		}
	}
}
