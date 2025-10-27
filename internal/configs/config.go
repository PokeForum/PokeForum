package configs

import (
	"github.com/PokeForum/PokeForum/internal/configs/autoload"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	Host       string
	Port       string
	ConfigPath string
	Debug      bool
)

type Configuration struct {
	DB    autoload.DB    `mapstructure:"db" json:"db" yaml:"db"`
	Cache autoload.Cache `mapstructure:"cache" json:"cache" yaml:"cache"`
}

// 全局方法
var (
	Config Configuration
	Log    *zap.Logger
	Json   jsoniter.API
	VP     *viper.Viper
)
