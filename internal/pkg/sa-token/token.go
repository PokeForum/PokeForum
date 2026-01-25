package sa_token

import (
	"fmt"
	"net/http"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/redis"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/configs"
)

// 全局 Manager 引用，用于获取配置
var globalManager *saGin.Manager

// NewSaToken Create SaToken | 创建 SaToken
func NewSaToken() *saGin.Manager {
	// Create SaToken configuration | 创建SaToken配置
	saCfg := &saGin.Config{
		TokenName:              "pf_token",           // Cookie 名称 | Cookie name
		Timeout:                2592000,              // 30 days (seconds) | 30天（秒）
		ActiveTimeout:          259200,               // 3 days (seconds) activity threshold | 3天（秒）活跃阈值
		IsConcurrent:           true,                 // Allow multi-device login | 允许多设备登录
		IsShare:                false,                // Do not share Token | 不共享Token
		MaxLoginCount:          5,                    // Limit to a maximum of 5 devices | 限制最多5台设备登录
		IsReadHeader:           false,                // 关闭 Header 读取 | Disable reading from Header
		IsReadCookie:           true,                 // 开启 Cookie 读取 | Enable reading from Cookie
		TokenStyle:             saGin.TokenStyleHash, // SHA256 hash style | SHA256哈希风格
		DataRefreshPeriod:      604800,               // Auto-renewal 7 days (in seconds) | 自动续签7天（单位：秒）
		TokenSessionCheckLogin: true,                 // Check Token during login | 在登录时是否检查Token
		AutoRenew:              true,                 // Automatic renewal | 自动续期
		IsLog:                  configs.Debug,        // Output operation logs | 输出操作日志
		IsPrintBanner:          false,                // Disable console banner output | 关闭控制台Banner输出
		KeyPrefix:              "forum:",             // Storage key prefix | 储存键前缀
		CookieConfig: &saGin.CookieConfig{
			Domain:   configs.CookieDomain, // 从配置读取，如 ".example.com" | Read from config, e.g. ".example.com"
			Path:     "/",
			Secure:   !configs.Debug, // 生产环境开启 HTTPS Only | Enable HTTPS Only in production
			HttpOnly: true,           // 防止 XSS | Prevent XSS
			SameSite: "Lax",          // 防止 CSRF | Prevent CSRF
			MaxAge:   2592000,        // 30天，与 Token 过期时间一致 | 30 days, same as Token timeout
		},
	}

	// Create storage | 创建存储
	storage := redis.NewStorageFromClient(configs.Cache)
	globalManager = saGin.NewManager(storage, saCfg)
	return globalManager
}

// WriteTokenToCookie 将 Token 写入 Cookie
func WriteTokenToCookie(c *gin.Context, token string) {
	if globalManager == nil {
		return
	}
	cfg := globalManager.GetConfig()
	maxAge := int(cfg.Timeout)
	if maxAge < 0 {
		maxAge = 0
	}

	// 设置 SameSite
	switch cfg.CookieConfig.SameSite {
	case "Strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "Lax":
		c.SetSameSite(http.SameSiteLaxMode)
	case "None":
		c.SetSameSite(http.SameSiteNoneMode)
	default:
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie(
		cfg.TokenName,
		token,
		maxAge,
		cfg.CookieConfig.Path,
		cfg.CookieConfig.Domain,
		cfg.CookieConfig.Secure,
		cfg.CookieConfig.HttpOnly,
	)
}

// DeleteTokenCookie 删除 Token Cookie
func DeleteTokenCookie(c *gin.Context) {
	if globalManager == nil {
		return
	}
	cfg := globalManager.GetConfig()

	// 设置 SameSite
	switch cfg.CookieConfig.SameSite {
	case "Strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "Lax":
		c.SetSameSite(http.SameSiteLaxMode)
	case "None":
		c.SetSameSite(http.SameSiteNoneMode)
	default:
		c.SetSameSite(http.SameSiteLaxMode)
	}

	// MaxAge 设为 -1 删除 Cookie
	c.SetCookie(
		cfg.TokenName,
		"",
		-1,
		cfg.CookieConfig.Path,
		cfg.CookieConfig.Domain,
		cfg.CookieConfig.Secure,
		cfg.CookieConfig.HttpOnly,
	)
}

// GetTokenFromCookie 从 Cookie 获取 Token
func GetTokenFromCookie(c *gin.Context) string {
	if globalManager == nil {
		return ""
	}
	cfg := globalManager.GetConfig()
	token, err := c.Cookie(cfg.TokenName)
	if err != nil {
		return ""
	}
	return token
}

// GetUserIDFromCookie 从 Cookie 获取用户 ID
func GetUserIDFromCookie(c *gin.Context) (int, error) {
	token := GetTokenFromCookie(c)
	if token == "" {
		return 0, fmt.Errorf("token not found | 未找到Token")
	}

	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	userID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// GetUserIDFromCookieOrZero 从 Cookie 获取用户 ID，未登录返回 0（游客模式）
func GetUserIDFromCookieOrZero(c *gin.Context) int {
	userID, err := GetUserIDFromCookie(c)
	if err != nil {
		return 0
	}
	return userID
}

// IsLoggedIn 检查用户是否已登录
func IsLoggedIn(c *gin.Context) bool {
	token := GetTokenFromCookie(c)
	return token != ""
}
