package sa_token

import (
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/redis"

	"github.com/PokeForum/PokeForum/internal/configs"
)

// NewSaToken Create SaToken | 创建 SaToken
func NewSaToken() *saGin.Manager {
	// Create SaToken configuration | 创建SaToken配置
	saCfg := &saGin.Config{
		TokenName:              "Authorization",
		Timeout:                2592000, // 30 days (seconds) | 30天（秒）
		ActiveTimeout:          259200,  // 3 days (seconds) activity threshold | 3天（秒）活跃阈值
		IsConcurrent:           true,    // Allow multi-device login | 允许多设备登录
		IsShare:                false,   // Do not share Token | 不共享Token
		MaxLoginCount:          5,       // Limit to a maximum of 5 devices | 限制最多5台设备登录
		IsReadHeader:           true,
		TokenStyle:             saGin.TokenStyleHash, // SHA256 hash style | SHA256哈希风格
		DataRefreshPeriod:      604800,               // Auto-renewal 7 days (in seconds) | 自动续签7天（单位：秒）
		TokenSessionCheckLogin: true,                 // Check Token during login | 在登录时是否检查Token
		AutoRenew:              true,                 // Automatic renewal | 自动续期
		IsLog:                  configs.Debug,        // Output operation logs | 输出操作日志
		IsPrintBanner:          false,                // Disable console banner output | 关闭控制台Banner输出
		KeyPrefix:              "forum:",             // Storage key prefix | 储存键前缀
		CookieConfig: &saGin.CookieConfig{
			Domain:   "",
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
			SameSite: "Lax",
			MaxAge:   0,
		},
	}

	// Create storage | 创建存储
	storage := redis.NewStorageFromClient(configs.Cache)
	return saGin.NewManager(storage, saCfg)
}
