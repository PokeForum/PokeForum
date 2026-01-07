package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders Security response headers middleware | 安全响应头中间件
// Add common security response headers to prevent common web attacks | 添加常见的安全响应头，防止常见的Web攻击
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: Prevent browser MIME type sniffing | X-Content-Type-Options: 防止浏览器MIME类型嗅探
		// Block browser from parsing response content as a type different from Content-Type | 阻止浏览器将响应内容解析为与Content-Type不同的类型
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevent clickjacking attacks | X-Frame-Options: 防止点击劫持攻击
		// DENY: Prohibit any page from embedding site content | DENY: 禁止任何页面嵌入本站内容
		// SAMEORIGIN: Only allow same-origin pages to embed | SAMEORIGIN: 只允许同源页面嵌入
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: Enable browser built-in XSS filter | X-XSS-Protection: 启用浏览器内置的XSS过滤器
		// 1; mode=block: Block page loading when XSS attack is detected | 1; mode=block: 检测到XSS攻击时阻止页面加载
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Control Referer header sending policy | Referrer-Policy: 控制Referer头的发送策略
		// strict-origin-when-cross-origin: Send full URL for same-origin, only origin for cross-origin | strict-origin-when-cross-origin: 同源发送完整URL，跨域只发送origin
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// X-Permitted-Cross-Domain-Policies: Restrict Adobe products' cross-domain policy | X-Permitted-Cross-Domain-Policies: 限制Adobe产品的跨域策略
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// X-Download-Options: Prevent IE browser from directly opening downloaded files | X-Download-Options: 防止IE浏览器直接打开下载的文件
		c.Header("X-Download-Options", "noopen")

		// Cache-Control: API responses do not cache sensitive data by default | Cache-Control: API响应默认不缓存敏感数据
		// Note: Static resource routes may need separate cache policy settings | 注意：静态资源路由可能需要单独设置缓存策略
		if c.Request.URL.Path != "/ping" && c.Request.URL.Path != "/health" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		// Permissions-Policy: Restrict browser feature usage | Permissions-Policy: 限制浏览器功能的使用
		// Disable sensitive features like camera, microphone, geolocation, etc. | 禁用摄像头、麦克风、地理位置等敏感功能
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		c.Next()
	}
}
