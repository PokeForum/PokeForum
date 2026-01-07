package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
)

var CorsConfig = cors.Config{
	AllowAllOrigins:  true,
	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},  // Allowed HTTP methods list | 允许的HTTP方法列表
	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"}, // Allowed HTTP headers list | 允许的HTTP头部列表
	AllowCredentials: true,                                                                  // Whether to allow browser to send cookies | 是否允许浏览器发送Cookie
	MaxAge:           30 * time.Minute,                                                      // Preflight request (OPTIONS) cache time (seconds) | 预检请求（OPTIONS）的缓存时间（秒）
}
