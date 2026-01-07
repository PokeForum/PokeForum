// Package cmd provides command-line interface and server startup functionality | Package cmd 提供命令行接口和服务器启动功能
//
// @title PokeForum API
// @version 1.0
// @description PokeForum is a forum application based on Gin framework | PokeForum 是一个基于 Gin 框架的论坛应用程序
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:9876
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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

	jsoniter "github.com/json-iterator/go"
	"github.com/samber/do"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/initializer"
	"github.com/PokeForum/PokeForum/internal/pkg/asynq"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/logging"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/PokeForum/PokeForum/internal/utils"
)

var ServerCMD = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server`,
	// Start the server | 启动服务器
	RunE: func(cmd *cobra.Command, args []string) error {
		RunServer()
		return nil
	},
}

// RunServer starts the server | 启动服务器
func RunServer() {
	// Create injector | 创建注入器
	injector := do.New()

	// Initialize configuration file using the config file path from command line flags | 初始化配置文件，使用命令行标志中的配置文件路径
	configs.VP = initializer.Viper(configs.ConfigPath)

	// Initialize logger | 初始化日志
	configs.Log = logging.Zap()

	// Initialize JSON processor | 初始化 JSON 处理器
	configs.Json = jsoniter.ConfigCompatibleWithStandardLibrary

	// Create data directory | 创建数据目录
	if err := utils.CreatNestedFolder(utils.DataFolder); err != nil {
		configs.Log.Error(err.Error())
		return
	}

	// Create theme directory | 创建主题目录
	if err := utils.CreatNestedFolder(utils.DataFolder + "/theme"); err != nil {
		configs.Log.Error(err.Error())
		return
	}

	// Initialize database | 初始化数据库
	configs.DB = initializer.DB()
	if configs.DB == nil {
		configs.Log.Error("DB initializer failed")
		return
	}

	// Migrate database | 迁移数据库
	initializer.AutoMigrate(configs.DB)
	configs.Log.Info("DB initializer succeeded")

	// Initialize cache | 初始化缓存
	configs.Cache = initializer.Cache()
	if configs.Cache == nil {
		configs.Log.Error("Cache initializer failed")
		return
	}
	configs.Log.Info("Cache initializer succeeded")

	// Initialize cache service | 初始化缓存服务
	cacheService := cache.NewRedisCacheService(configs.Cache, configs.Log)

	// Initialize asynq task manager | 初始化asynq任务管理器
	taskManager := asynq.NewTaskManagerFromRedis(configs.Cache, 10, configs.Log)

	// Register sign-in async task handler | 注册签到异步任务处理器
	signinAsyncTask := service.NewSigninAsyncTask(configs.DB, taskManager, configs.Log)
	signinAsyncTask.RegisterHandler()

	// Register stats sync task handler and scheduled task (sync every 5 minutes) | 注册统计数据同步任务处理器和定时任务(每5分钟同步一次)
	syncTask := service.NewStatsSyncTask(configs.DB, cacheService, taskManager, configs.Log)
	syncTask.RegisterHandler()
	if err := syncTask.RegisterSchedule(5 * time.Minute); err != nil {
		configs.Log.Error("Failed to register stats sync scheduled task | 注册统计同步定时任务失败", zap.Error(err))
	}

	// Start asynq task server | 启动asynq任务服务器
	if err := taskManager.Start(); err != nil {
		configs.Log.Error("Failed to start asynq task server | 启动asynq任务服务器失败", zap.Error(err))
		return
	}

	// Execute stats sync immediately on startup | 启动时立即执行一次统计同步
	syncTask.SyncNow(context.Background())

	// Inject SigninAsyncTask into injector for SigninService to use | 将SigninAsyncTask注入到injector供SigninService使用
	do.ProvideValue(injector, signinAsyncTask)
	do.ProvideValue(injector, taskManager)

	// Register routes | 注册路由
	router := initializer.Routers(injector)

	// Start service | 启动服务
	sDSN := fmt.Sprintf("%s:%s", configs.Host, configs.Port)
	fmt.Printf("Server Run: %s \n", sDSN)
	srv := &http.Server{
		Addr:              sDSN,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start | 启动
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			configs.Log.Error(err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds | 等待终端信号来优雅关闭服务器，为关闭服务器设置10秒超时
	quit := make(chan os.Signal, 1) // Create a channel that receives signals | 创建一个接受信号的通道

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // This will not block here | 此处不会阻塞
	<-quit                                               // Block here until receiving the above two signals | 阻塞此处，当接受到上述两种信号时，才继续往下执行
	configs.Log.Info("Service ready to shut down")

	// Stop asynq task server | 停止asynq任务服务器
	taskManager.Stop()

	// Create context with 10 second timeout | 创建10秒超时的Context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Shutdown injector and release all services | 关闭注入器，释放所有服务
	if err := injector.Shutdown(); err != nil {
		configs.Log.Fatal(err.Error())
	}
	// Gracefully shutdown server within 10 seconds (finish processing unfinished requests before shutdown), timeout after 10 seconds | 10秒内优雅关闭服务（将未处理完成的请求处理完再关闭服务），超过10秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		configs.Log.Fatal("Service timed out has been shut down: ", zap.Error(err))
	}

	configs.Log.Info("Service has been shut down")
}
