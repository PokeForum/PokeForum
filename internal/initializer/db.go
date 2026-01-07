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

// DB initializes database | 初始化数据库
func DB() *ent.Client {
	m := configs.Config.DB
	if m.Name == "" {
		return nil
	}

	// Build PostgreSQL DSN | 构建PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		m.Host, m.Port, m.UserName, m.Password, m.Name, m.Config)

	// Print DSN in debug mode | 调试模式下打印DSN
	if configs.Debug {
		configs.Log.Debug("DSN: ", zap.String("dsn", dsn))
	}

	// Establish database connection | 建立数据库连接
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		configs.Log.Fatal("failed opening connection to postgres", zap.Error(err))
	}

	return client
}

// PgDB initializes native PostgreSQL connection (for scenarios requiring native SQL such as performance monitoring) | 初始化原生 PostgreSQL 连接（用于性能监控等需要原生 SQL 的场景）
func PgDB() *sql.DB {
	m := configs.Config.DB
	if m.Name == "" {
		return nil
	}

	// Build PostgreSQL DSN | 构建PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		m.Host, m.Port, m.UserName, m.Password, m.Name, m.Config)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		configs.Log.Fatal("failed opening raw connection to postgres", zap.Error(err))
	}

	// Test connection | 测试连接
	if err := db.PingContext(context.Background()); err != nil {
		configs.Log.Fatal("failed to ping postgres", zap.Error(err))
	}

	return db
}

// AutoMigrate performs automatic migration | 自动迁移
func AutoMigrate(client *ent.Client) {
	if err := client.Schema.Create(
		context.Background(),
		migrate.WithForeignKeys(false), // Disable all foreign keys | 禁用所有外键
	); err != nil {
		configs.Log.Fatal("failed creating schema resources: ", zap.Error(err))
	}
}
