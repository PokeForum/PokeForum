package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/usersigninlogs"
	"github.com/PokeForum/PokeForum/ent/usersigninstatus"
	pkgasynq "github.com/PokeForum/PokeForum/internal/pkg/asynq"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// SigninAsyncTask 签到异步任务处理器
// 使用asynq实现可靠的消息队列，确保数据持久化和故障恢复
type SigninAsyncTask struct {
	db          *ent.Client
	logger      *zap.Logger
	taskManager *pkgasynq.TaskManager
}

// SigninTaskPayload 签到任务载荷
type SigninTaskPayload struct {
	UserID         int64     `json:"user_id"`
	SignDate       time.Time `json:"sign_date"`
	ContinuousDays int       `json:"continuous_days"`
	TotalDays      int       `json:"total_days"`
	TraceID        string    `json:"trace_id"` // 用于链路追踪
}

// NewSigninAsyncTask 创建签到异步任务处理器
func NewSigninAsyncTask(db *ent.Client, taskManager *pkgasynq.TaskManager, logger *zap.Logger) *SigninAsyncTask {
	return &SigninAsyncTask{
		db:          db,
		logger:      logger,
		taskManager: taskManager,
	}
}

// RegisterHandler 注册任务处理器到TaskManager
func (s *SigninAsyncTask) RegisterHandler() {
	s.taskManager.RegisterHandlerFunc(pkgasynq.TypeSigninPersist, s.HandleSigninTask)
	s.logger.Info("签到异步任务处理器已注册")
}

// NewSigninTask 创建签到任务
func NewSigninTask(payload *SigninTaskPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化签到任务失败: %w", err)
	}
	return asynq.NewTask(pkgasynq.TypeSigninPersist, data, asynq.MaxRetry(3), asynq.Queue(pkgasynq.QueueDefault)), nil
}

// SubmitTask 提交签到任务
func (s *SigninAsyncTask) SubmitTask(ctx context.Context, payload *SigninTaskPayload) error {
	task, err := NewSigninTask(payload)
	if err != nil {
		s.logger.Error("创建签到任务失败",
			zap.Int64("user_id", payload.UserID),
			zap.String("trace_id", payload.TraceID),
			zap.Error(err))
		return err
	}

	info, err := s.taskManager.EnqueueContext(ctx, task)
	if err != nil {
		s.logger.Error("提交签到任务失败",
			zap.Int64("user_id", payload.UserID),
			zap.String("trace_id", payload.TraceID),
			zap.Error(err))
		return fmt.Errorf("提交任务失败: %w", err)
	}

	s.logger.Debug("提交签到任务成功",
		zap.Int64("user_id", payload.UserID),
		zap.String("trace_id", payload.TraceID),
		zap.String("task_id", info.ID))

	return nil
}

// HandleSigninTask 处理签到任务（asynq Handler）
func (s *SigninAsyncTask) HandleSigninTask(ctx context.Context, t *asynq.Task) error {
	var payload SigninTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		s.logger.Error("反序列化签到任务失败", zap.Error(err))
		return fmt.Errorf("反序列化失败: %v: %w", err, asynq.SkipRetry)
	}

	// 创建带链路ID的context
	if payload.TraceID != "" {
		ctx = tracing.WithTraceID(ctx, payload.TraceID)
	}

	startTime := time.Now()
	s.logger.Info("开始处理签到任务",
		zap.Int64("user_id", payload.UserID),
		zap.String("trace_id", payload.TraceID),
		tracing.WithTraceIDField(ctx))

	// 处理签到状态表
	if err := s.updateSigninStatus(ctx, &payload); err != nil {
		s.logger.Error("更新签到状态失败",
			zap.Int64("user_id", payload.UserID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return err
	}

	// 处理签到日志表
	if err := s.insertSigninLog(ctx, &payload); err != nil {
		s.logger.Error("插入签到日志失败",
			zap.Int64("user_id", payload.UserID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return err
	}

	duration := time.Since(startTime)
	s.logger.Info("签到任务处理完成",
		zap.Int64("user_id", payload.UserID),
		zap.Duration("duration", duration),
		tracing.WithTraceIDField(ctx))

	return nil
}

// updateSigninStatus 更新签到状态表
func (s *SigninAsyncTask) updateSigninStatus(ctx context.Context, payload *SigninTaskPayload) error {
	existing, err := s.db.UserSigninStatus.Query().
		Where(usersigninstatus.UserID(payload.UserID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 记录不存在，创建新记录
			err = s.db.UserSigninStatus.Create().
				SetUserID(payload.UserID).
				SetLastSigninDate(payload.SignDate).
				SetContinuousDays(payload.ContinuousDays).
				SetTotalDays(payload.TotalDays).
				Exec(ctx)
		} else {
			return err
		}
	} else {
		// 记录存在，更新记录
		_, err = s.db.UserSigninStatus.UpdateOne(existing).
			SetLastSigninDate(payload.SignDate).
			SetContinuousDays(payload.ContinuousDays).
			SetTotalDays(payload.TotalDays).
			Save(ctx)
	}

	if err != nil {
		return err
	}

	s.logger.Debug("签到状态更新成功",
		zap.Int64("user_id", payload.UserID),
		zap.Time("sign_date", payload.SignDate),
		zap.Int("continuous_days", payload.ContinuousDays),
		zap.Int("total_days", payload.TotalDays),
		tracing.WithTraceIDField(ctx))

	return nil
}

// insertSigninLog 插入签到日志
func (s *SigninAsyncTask) insertSigninLog(ctx context.Context, payload *SigninTaskPayload) error {
	// 检查今日是否已签到（防止重复插入）
	exists, err := s.db.UserSigninLogs.Query().
		Where(
			usersigninlogs.UserID(payload.UserID),
			usersigninlogs.SignDate(payload.SignDate),
		).
		Exist(ctx)

	if err != nil {
		return err
	}

	if exists {
		s.logger.Debug("签到日志已存在，跳过插入",
			zap.Int64("user_id", payload.UserID),
			zap.Time("sign_date", payload.SignDate),
			tracing.WithTraceIDField(ctx))
		return nil
	}

	// 插入新的签到日志
	err = s.db.UserSigninLogs.Create().
		SetUserID(payload.UserID).
		SetSignDate(payload.SignDate).
		Exec(ctx)

	if err != nil {
		return err
	}

	s.logger.Debug("签到日志插入成功",
		zap.Int64("user_id", payload.UserID),
		zap.Time("sign_date", payload.SignDate),
		tracing.WithTraceIDField(ctx))

	return nil
}
