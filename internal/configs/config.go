package configs

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/configs/autoload"
)

var (
	Host       string
	Port       string
	ConfigPath string
	Debug      bool
	Prometheus bool // Prometheus监控开关
)

type Configuration struct {
	DB    autoload.DB    `mapstructure:"db" json:"db" yaml:"db"`
	Cache autoload.Cache `mapstructure:"cache" json:"cache" yaml:"cache"`
}

// 全局方法
var (
	Config Configuration
	Log    *zap.Logger
	DB     *ent.Client
	Cache  *redis.Client
	Json   jsoniter.API
	VP     *viper.Viper
	PgDB   interface{} // 原生 PostgreSQL 连接 (*sql.DB)
)
