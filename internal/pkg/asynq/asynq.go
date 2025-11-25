package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TaskManager asynq任务管理器
type TaskManager struct {
	client    *asynq.Client
	server    *asynq.Server
	scheduler *asynq.Scheduler
	mux       *asynq.ServeMux
	logger    *zap.Logger
	redisOpt  asynq.RedisClientOpt
}

// Config 任务管理器配置
type Config struct {
	RedisAddr     string // Redis地址
	RedisPassword string // Redis密码
	RedisDB       int    // Redis数据库
	Concurrency   int    // 并发worker数量
}

// NewTaskManager 创建任务管理器
func NewTaskManager(cfg *Config, logger *zap.Logger) *TaskManager {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	// 创建客户端
	client := asynq.NewClient(redisOpt)

	// 创建服务端配置
	serverCfg := asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues: map[string]int{
			"critical": 6, // 高优先级队列
			"default":  3, // 默认队列
			"low":      1, // 低优先级队列
		},
		// 错误处理
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error("任务处理失败",
				zap.String("task_type", task.Type()),
				zap.Error(err))
		}),
		// 日志
		Logger: &asynqLogger{logger: logger},
	}

	server := asynq.NewServer(redisOpt, serverCfg)
	mux := asynq.NewServeMux()

	// 创建调度器
	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Logger: &asynqLogger{logger: logger},
	})

	return &TaskManager{
		client:    client,
		server:    server,
		scheduler: scheduler,
		mux:       mux,
		logger:    logger,
		redisOpt:  redisOpt,
	}
}

// NewTaskManagerFromRedis 从现有Redis客户端创建任务管理器
func NewTaskManagerFromRedis(rdb *redis.Client, concurrency int, logger *zap.Logger) *TaskManager {
	opts := rdb.Options()
	cfg := &Config{
		RedisAddr:     opts.Addr,
		RedisPassword: opts.Password,
		RedisDB:       opts.DB,
		Concurrency:   concurrency,
	}
	return NewTaskManager(cfg, logger)
}

// Client 获取客户端
func (tm *TaskManager) Client() *asynq.Client {
	return tm.client
}

// Mux 获取路由
func (tm *TaskManager) Mux() *asynq.ServeMux {
	return tm.mux
}

// Scheduler 获取调度器
func (tm *TaskManager) Scheduler() *asynq.Scheduler {
	return tm.scheduler
}

// RegisterHandler 注册任务处理器
func (tm *TaskManager) RegisterHandler(taskType string, handler asynq.Handler) {
	tm.mux.Handle(taskType, handler)
}

// RegisterHandlerFunc 注册任务处理函数
func (tm *TaskManager) RegisterHandlerFunc(taskType string, handler func(context.Context, *asynq.Task) error) {
	tm.mux.HandleFunc(taskType, handler)
}

// RegisterSchedule 注册定时任务
// cronSpec: cron表达式，如 "@every 5m" 或 "0 */5 * * * *"
func (tm *TaskManager) RegisterSchedule(cronSpec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	return tm.scheduler.Register(cronSpec, task, opts...)
}

// Enqueue 提交任务到队列
func (tm *TaskManager) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return tm.client.Enqueue(task, opts...)
}

// EnqueueContext 提交任务到队列（带context）
func (tm *TaskManager) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return tm.client.EnqueueContext(ctx, task, opts...)
}

// Start 启动任务服务器和调度器
func (tm *TaskManager) Start() error {
	tm.logger.Info("启动asynq任务服务器")

	// 启动调度器
	if err := tm.scheduler.Start(); err != nil {
		return err
	}
	tm.logger.Info("asynq调度器已启动")

	// 启动服务器（非阻塞）
	go func() {
		if err := tm.server.Run(tm.mux); err != nil {
			tm.logger.Error("asynq服务器运行错误", zap.Error(err))
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	tm.logger.Info("asynq任务服务器已启动")

	return nil
}

// Stop 停止任务服务器和调度器
func (tm *TaskManager) Stop() {
	tm.logger.Info("正在停止asynq任务服务器")

	// 停止调度器
	tm.scheduler.Shutdown()
	tm.logger.Info("asynq调度器已停止")

	// 停止服务器
	tm.server.Shutdown()
	tm.logger.Info("asynq服务器已停止")

	// 关闭客户端
	if err := tm.client.Close(); err != nil {
		tm.logger.Error("关闭asynq客户端失败", zap.Error(err))
	}

	tm.logger.Info("asynq任务服务器已停止")
}

// asynqLogger 适配zap日志到asynq
type asynqLogger struct {
	logger *zap.Logger
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}
