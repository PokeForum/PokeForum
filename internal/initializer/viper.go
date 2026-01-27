package initializer

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/PokeForum/PokeForum/internal/configs"
	_const "github.com/PokeForum/PokeForum/internal/consts"
)

// Viper initializes configuration | 初始化配置
func Viper(configPath string) *viper.Viper {
	v := viper.New()
	// If config file path is not specified, use the default path | 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		configPath = _const.ConfigPath
	}
	// Specify config file path (supports absolute or relative path) | 指定配置文件路径（支持绝对路径或相对路径）
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error configs file: %w", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("configs file changed: ", e.Name)
		if err = v.Unmarshal(&configs.Config); err != nil {
			slog.Error(err.Error())
		}
		// Apply timezone when config changes | 配置变更时应用时区
		applyTimezone()
	})
	if err = v.Unmarshal(&configs.Config); err != nil {
		slog.Error(err.Error())
	}

	// 应用配置文件中的 APP 配置到全局变量（命令行参数优先级更高）
	// Apply APP config from file to global variables (CLI args have higher priority)
	applyAppConfig()

	// Apply timezone configuration | 应用时区配置
	applyTimezone()

	return v
}

// applyAppConfig 将配置文件中的 APP 配置应用到全局变量
// 命令行参数优先级高于配置文件：只有当命令行参数为默认值时，才使用配置文件的值
func applyAppConfig() {
	app := configs.Config.APP

	// Host: 命令行默认值为 "localhost"，如果未修改则使用配置文件
	if configs.Host == "localhost" && app.Host != "" {
		configs.Host = app.Host
	}

	// Port: 命令行默认值为 "9876"，如果未修改则使用配置文件
	if configs.Port == "9876" && app.Port != "" {
		configs.Port = app.Port
	}

	// Debug: 命令行默认值为 false，如果未修改则使用配置文件
	if !configs.Debug && app.Debug {
		configs.Debug = app.Debug
	}

	// Prometheus: 命令行默认值为 false，如果未修改则使用配置文件
	if !configs.Prometheus && app.Prometheus {
		configs.Prometheus = app.Prometheus
	}

	// CookieDomain: 命令行默认值为 ""，如果未修改则使用配置文件
	if configs.CookieDomain == "" && app.CookieDomain != "" {
		configs.CookieDomain = app.CookieDomain
	}

	// Timezone: 命令行默认值为 ""，如果未修改则使用配置文件
	if configs.Timezone == "" && app.Timezone != "" {
		configs.Timezone = app.Timezone
	}
}

// applyTimezone applies the configured timezone to the system | 应用时区配置到系统
func applyTimezone() {
	timezone := configs.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai" // Default timezone | 默认时区
		configs.Timezone = timezone
	}

	// Load timezone location | 加载时区位置
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Printf("Failed to load timezone %s: %v. Using UTC as fallback.\n", timezone, err)
		timezone = "UTC"
		loc = time.UTC
		configs.Timezone = timezone
	}

	// Set system local time to the configured timezone | 将系统本地时间设置为配置的时区
	time.Local = loc
	fmt.Printf("System timezone set to: %s (UTC offset: %s)\n", timezone, loc.String())
}
