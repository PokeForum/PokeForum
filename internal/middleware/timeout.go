package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
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
		c.Next()
	}
}
