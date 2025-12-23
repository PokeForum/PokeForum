package sa_token

import (
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/redis"

	"github.com/PokeForum/PokeForum/internal/configs"
)

// NewSaToken 创建 SaToken
func NewSaToken() *saGin.Manager {
	// 创建SaToken配置
	saCfg := &saGin.Config{
		TokenName:              "Authorization",
		Timeout:                2592000, // 30天（秒）
		ActiveTimeout:          259200,  // 3天（秒）活跃阈值
		IsConcurrent:           true,    // 允许多设备登录
		IsShare:                false,   // 不共享Token
		MaxLoginCount:          5,       // 限制最多5台设备登录
		IsReadHeader:           true,
		TokenStyle:             saGin.TokenStyleHash, // SHA256哈希风格
		DataRefreshPeriod:      604800,               // 自动续签7天（单位：秒）
		TokenSessionCheckLogin: true,                 // 在登录时是否检查Token
		AutoRenew:              true,                 // 自动续期
		IsLog:                  configs.Debug,        // 输出操作日志
		IsPrintBanner:          false,                // 关闭控制台Banner输出
		KeyPrefix:              "forum:",             // 储存键前缀
		CookieConfig: &saGin.CookieConfig{
			Domain:   "",
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
			SameSite: "Lax",
			MaxAge:   0,
		},
	}

	// 创建存储
	storage := redis.NewStorageFromClient(configs.Cache)
	return saGin.NewManager(storage, saCfg)
}
