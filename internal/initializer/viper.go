package initializer

import (
	"fmt"
	"log/slog"

	"github.com/PokeForum/PokeForum/internal/configs"
	_const "github.com/PokeForum/PokeForum/internal/const"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Viper 初始化配置
func Viper(configPath string) *viper.Viper {
	v := viper.New()
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		configPath = _const.ConfigPath
	}
	// 指定配置文件路径（支持绝对路径或相对路径）
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error configs file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("configs file changed: ", e.Name)
		if err = v.Unmarshal(&configs.Config); err != nil {
			slog.Error(err.Error())
		}
	})
	if err = v.Unmarshal(&configs.Config); err != nil {
		slog.Error(err.Error())
	}

	return v
}
