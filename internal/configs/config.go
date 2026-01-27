package configs

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/configs/autoload"
)

var (
	Host         string
	Port         string
	ConfigPath   string
	Debug        bool
	Prometheus   bool   // Monitoring switch | 监控开关
	CookieDomain string // Cookie 域名，如 ".example.com" | Cookie domain, e.g. ".example.com"
	Timezone     string // System timezone | 系统时区
)

type Configuration struct {
	APP   autoload.APP   `mapstructure:"app" json:"app" yaml:"app"`
	DB    autoload.DB    `mapstructure:"db" json:"db" yaml:"db"`
	Cache autoload.Cache `mapstructure:"cache" json:"cache" yaml:"cache"`
}

// 全局方法
var (
	Config Configuration
	Log    *zap.Logger
	DB     *ent.Client
	Cache  *redis.Client
	VP     *viper.Viper
)
