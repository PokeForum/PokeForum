package cmd

import (
	"fmt"
	"os"

	"github.com/PokeForum/PokeForum/internal/configs"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/spf13/cobra"
)

// GetEnv 获取环境变量
func GetEnv(k, v string) string {
	value := os.Getenv(k)
	if value == "" {
		return v
	}
	return value
}

var (
	host   = GetEnv("host", "localhost")
	port   = GetEnv("port", "9876")
	config = GetEnv("config", _const.ConfigPath)
)

var RootCMD = &cobra.Command{
	Use:   "PokeForum",
	Short: "站在巨人肩膀的论坛程序",
	Long:  "Go+Gin+PgSQL+Redis+MeiliSearch构建的高性能论坛程序",
	// 如果没有指定子命令，默认运行 server
	RunE: func(cmd *cobra.Command, args []string) error {
		return ServerCMD.RunE(cmd, args)
	},
}

// Execute 启动
func Execute() {
	if err := RootCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 设置命令行参数，提供环境变量作为默认值
	RootCMD.PersistentFlags().StringVar(&configs.Host, "host", host, "主机地址")
	RootCMD.PersistentFlags().StringVar(&configs.Port, "port", port, "端口号")
	RootCMD.PersistentFlags().StringVarP(&configs.ConfigPath, "config", "c", config, "配置文件路径")
	RootCMD.PersistentFlags().BoolVar(&configs.Debug, "debug", false, "是否开启调试模式")
	RootCMD.PersistentFlags().BoolVar(&configs.Prometheus, "prometheus", false, "是否开启Prometheus监控")

	// 将 ServerCMD 添加为 RootCMD 的子命令
	RootCMD.AddCommand(ServerCMD)
}
