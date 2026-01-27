package service

import (
	"context"
	"encoding/json"
	"fmt"

	hibikenAsynq "github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/pkg/asynq"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
)

type LoginLogAsyncTask struct {
	loginLogRepo repository.IUserLoginLogRepository
	logger       *zap.Logger
	taskManager  *asynq.TaskManager
}

type LoginLogPayload struct {
	UserID  int    `json:"user_id"`
	IP      string `json:"ip"`
	TraceID string `json:"trace_id"`
}

func NewLoginLogAsyncTask(
	loginLogRepo repository.IUserLoginLogRepository,
	logger *zap.Logger,
	taskManager *asynq.TaskManager,
) *LoginLogAsyncTask {
	return &LoginLogAsyncTask{
		loginLogRepo: loginLogRepo,
		logger:       logger,
		taskManager:  taskManager,
	}
}

func (t *LoginLogAsyncTask) RegisterHandler() {
	t.taskManager.RegisterHandlerFunc(asynq.TypeLoginLog, t.HandleLoginLogTask)
	t.logger.Info("Login log async task handler registered | 登录日志异步任务处理器已注册")
}

func NewLoginLogTask(payload *LoginLogPayload) (*hibikenAsynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize login log task | 序列化登录日志任务失败: %w", err)
	}
	return hibikenAsynq.NewTask(asynq.TypeLoginLog, data, hibikenAsynq.MaxRetry(3), hibikenAsynq.Queue(asynq.QueueDefault)), nil
}

func (t *LoginLogAsyncTask) SubmitTask(ctx context.Context, payload *LoginLogPayload) error {
	task, err := NewLoginLogTask(payload)
	if err != nil {
		t.logger.Error("Failed to create login log task | 创建登录日志任务失败",
			zap.Int("user_id", payload.UserID),
			zap.String("trace_id", payload.TraceID),
			zap.Error(err))
		return err
	}

	info, err := t.taskManager.EnqueueContext(ctx, task)
	if err != nil {
		t.logger.Error("Failed to submit login log task | 提交登录日志任务失败",
			zap.Int("user_id", payload.UserID),
			zap.String("trace_id", payload.TraceID),
			zap.Error(err))
		return fmt.Errorf("failed to submit task | 提交任务失败: %w", err)
	}

	t.logger.Debug("Login log task submitted successfully | 提交登录日志任务成功",
		zap.Int("user_id", payload.UserID),
		zap.String("trace_id", payload.TraceID),
		zap.String("task_id", info.ID))

	return nil
}

func (t *LoginLogAsyncTask) HandleLoginLogTask(ctx context.Context, task *hibikenAsynq.Task) error {
	var payload LoginLogPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.logger.Error("Failed to deserialize login log task | 反序列化登录日志任务失败", zap.Error(err))
		return fmt.Errorf("deserialization failed | 反序列化失败: %v: %w", err, hibikenAsynq.SkipRetry)
	}

	if payload.TraceID != "" {
		ctx = tracing.WithTraceID(ctx, payload.TraceID)
	}

	t.logger.Info("Start processing login log task | 开始处理登录日志任务",
		zap.Int("user_id", payload.UserID),
		zap.String("trace_id", payload.TraceID),
		tracing.WithTraceIDField(ctx))

	_, err := t.loginLogRepo.Create(ctx, payload.UserID, payload.IP, "", true)
	if err != nil {
		t.logger.Error("Failed to create login log | 创建登录日志失败",
			zap.Int("user_id", payload.UserID),
			zap.String("ip_address", payload.IP),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return err
	}

	t.logger.Info("Login log task processing completed | 登录日志任务处理完成",
		zap.Int("user_id", payload.UserID),
		tracing.WithTraceIDField(ctx))

	return nil
}
