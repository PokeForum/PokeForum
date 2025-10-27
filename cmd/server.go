package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/initializer"
	"github.com/PokeForum/PokeForum/internal/pkg/logging"
	"github.com/PokeForum/PokeForum/internal/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ServerCMD = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server`,
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

// RunServer 启动服务器
func RunServer() {
	// 初始化配置文件，使用命令行标志中的配置文件路径
	configs.VP = initializer.Viper(configs.ConfigPath)

	// 初始化日志
	configs.Log = logging.Zap()

	// 初始化 JSON 处理器
	configs.Json = jsoniter.ConfigCompatibleWithStandardLibrary

	// 创建数据目录
	if err := utils.CreatNestedFolder(utils.DataFolder); err != nil {
		configs.Log.Error(err.Error())
		return
	}

	// 创建主题目录
	if err := utils.CreatNestedFolder(utils.DataFolder + "/theme"); err != nil {
		configs.Log.Error(err.Error())
		return
	}

	// 初始化数据库
	configs.DB = initializer.DB()
	if configs.DB == nil {
		configs.Log.Error("DB initializer failed")
		return
	} else {
		// 迁移数据库
		initializer.AutoMigrate(configs.DB)
		configs.Log.Info("DB initializer succeeded")
	}

	// 初始化缓存
	configs.Cache = initializer.Cache()
	if configs.Cache == nil {
		configs.Log.Error("Cache initializer failed")
		return
	} else {
		configs.Log.Info("Cache initializer succeeded")

	}

	// 注册路由
	router := initializer.Routers()

	// 启动服务
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", configs.Host, configs.Port),
		Handler: router,
	}

	// 启动
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			configs.Log.Error(err.Error())
		}
	}()

	// 等待终端信号来优雅关闭服务器，为关闭服务器设置10秒超时
	quit := make(chan os.Signal, 1) // 创建一个接受信号的通道

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞此处，当接受到上述两种信号时，才继续往下执行
	configs.Log.Info("Service ready to shut down")

	// 创建10秒超时的Context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 10秒内优雅关闭服务（将未处理完成的请求处理完再关闭服务），超过10秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		configs.Log.Fatal("Service timed out has been shut down: ", zap.Error(err))
	}

	configs.Log.Info("Service has been shut down")
}
