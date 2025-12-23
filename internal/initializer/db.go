package initializer

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/migrate"
	"github.com/PokeForum/PokeForum/internal/configs"

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

// PgDB 初始化原生 PostgreSQL 连接（用于性能监控等需要原生 SQL 的场景）
func PgDB() *sql.DB {
	m := configs.Config.DB
	if m.Name == "" {
		return nil
	}

	// 构建PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		m.Host, m.Port, m.UserName, m.Password, m.Name, m.Config)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		configs.Log.Fatal("failed opening raw connection to postgres", zap.Error(err))
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		configs.Log.Fatal("failed to ping postgres", zap.Error(err))
	}

	return db
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
