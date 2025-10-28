package sa_token

import (
	"github.com/PokeForum/PokeForum/internal/configs"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/redis"
	"go.uber.org/zap"
)

// NewSaToken 创建 SaToken
func NewSaToken() *saGin.Manager {
	// 从 configs 中获取 redis 配置
	rdbCfg := &redis.Config{
		Host:     configs.Config.Cache.Host,
		Port:     configs.Config.Cache.Port,
		Password: configs.Config.Cache.Password,
		Database: configs.Config.Cache.DB,
		PoolSize: 10,
	}

	// 创建SaToken配置
	saCfg := &saGin.Config{
		TokenName:              "forum",
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
	storage, err := redis.NewStorageFromConfig(rdbCfg)
	if err != nil {
		configs.Log.Panic("Failed to create Redis storage: ", zap.Error(err))
		return nil
	}

	return saGin.NewManager(storage, saCfg)
}
