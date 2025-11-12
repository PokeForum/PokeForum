package initializer

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/migrate"
	"github.com/PokeForum/PokeForum/internal/configs"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

// DB 初始化数据库
func DB() *ent.Client {
	m := configs.Config.DB
	if m.Name == "" {
		return nil
	}

	// 构建PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		m.Host, m.Port, m.UserName, m.Password, m.Name, m.Config)

	// 调试模式下打印DSN
	if configs.Debug {
		configs.Log.Debug("DSN: ", zap.String("dsn", dsn))
	}

	// 建立数据库连接
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		configs.Log.Fatal("failed opening connection to postgres", zap.Error(err))
	}

	return client
}

// AutoMigrate 自动迁移
func AutoMigrate(client *ent.Client) {
	if err := client.Schema.Create(
		context.Background(),
		migrate.WithForeignKeys(false), // 禁用所有外键
	); err != nil {
		configs.Log.Fatal("failed creating schema resources: ", zap.Error(err))
	}
}
