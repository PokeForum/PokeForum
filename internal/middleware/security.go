package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 安全响应头中间件
// 添加常见的安全响应头，防止常见的Web攻击
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: 防止浏览器MIME类型嗅探
		// 阻止浏览器将响应内容解析为与Content-Type不同的类型
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: 防止点击劫持攻击
		// DENY: 禁止任何页面嵌入本站内容
		// SAMEORIGIN: 只允许同源页面嵌入
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: 启用浏览器内置的XSS过滤器
		// 1; mode=block: 检测到XSS攻击时阻止页面加载
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: 控制Referer头的发送策略
		// strict-origin-when-cross-origin: 同源发送完整URL，跨域只发送origin
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// X-Permitted-Cross-Domain-Policies: 限制Adobe产品的跨域策略
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// X-Download-Options: 防止IE浏览器直接打开下载的文件
		c.Header("X-Download-Options", "noopen")

		// Cache-Control: API响应默认不缓存敏感数据
		// 注意：静态资源路由可能需要单独设置缓存策略
		if c.Request.URL.Path != "/ping" && c.Request.URL.Path != "/health" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		// Permissions-Policy: 限制浏览器功能的使用
		// 禁用摄像头、麦克风、地理位置等敏感功能
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		c.Next()
	}
}
