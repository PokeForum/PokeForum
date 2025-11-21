package initializer

import (
	"context"
	"strconv"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.uber.org/zap"
)

// Cache 初始化Redis客户端
func Cache() *redis.Client {
	m := configs.Config.Cache

	client := redis.NewClient(&redis.Options{
		Addr:     m.Host + ":" + strconv.Itoa(m.Port),
		Password: m.Password,
		DB:       m.DB,
		PoolSize: 10,
		// 禁用 maint_notifications [企业版功能。开源版需关闭]
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		configs.Log.Panic("Failed to create Redis connection: ", zap.Error(err))
	}

	return client
}
