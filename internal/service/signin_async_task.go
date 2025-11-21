package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/usersigninlogs"
	"github.com/PokeForum/PokeForum/ent/usersigninstatus"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"go.uber.org/zap"
)

// SigninAsyncTask 签到异步任务处理器
// 使用Redis Streams实现可靠的消息队列，确保数据持久化和故障恢复
type SigninAsyncTask struct {
	db         *ent.Client
	cache      cache.ICacheService
	logger     *zap.Logger
	streamName string // Stream名称
	groupName  string // 消费者组名称
	stopChan   chan struct{}
	wg         sync.WaitGroup
	isStarted  bool
	mu         sync.RWMutex
}

// SigninTask 签到任务
type SigninTask struct {
	UserID         int64
	SignDate       time.Time
	ContinuousDays int
	TotalDays      int
	TraceID        string // 用于链路追踪
}

// NewSigninAsyncTask 创建签到异步任务处理器
func NewSigninAsyncTask(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) *SigninAsyncTask {
	return &SigninAsyncTask{
		db:         db,
		cache:      cacheService,
		logger:     logger,
		streamName: "signin:task:stream", // Stream名称
		groupName:  "signin:workers",     // 消费者组名称
		stopChan:   make(chan struct{}),
	}
}

// Start 启动异步任务处理器
func (s *SigninAsyncTask) Start(workerCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	if s.isStarted {
		s.logger.Warn("签到异步任务处理器已经启动")
		return
	}

	s.logger.Debug("启动签到异步任务处理器", zap.Int("worker_count", workerCount))

	// 初始化Stream和消费者组
	err := s.initializeStream(ctx)
	if err != nil {
		s.logger.Error("初始化Redis Stream失败", zap.Error(err))
		return
	}

	// 启动多个worker处理任务
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	s.isStarted = true
	s.logger.Debug("签到异步任务处理器已启动")
}

// initializeStream 初始化Stream和消费者组
func (s *SigninAsyncTask) initializeStream(ctx context.Context) error {
	// 创建消费者组，如果已存在会返回错误但不影响使用
	err := s.cache.XGroupCreate(ctx, s.streamName, s.groupName, "0")
	if err != nil {
		// 如果消费者组已存在，忽略错误
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("创建消费者组失败: %w", err)
		}
		s.logger.Debug("消费者组已存在", zap.String("group", s.groupName))
	} else {
		s.logger.Debug("创建消费者组成功", zap.String("group", s.groupName))
	}

	// 检查Stream长度
	length, err := s.cache.XLen(ctx, s.streamName)
	if err != nil {
		return fmt.Errorf("获取Stream长度失败: %w", err)
	}

	s.logger.Debug("Stream初始化完成",
		zap.String("stream", s.streamName),
		zap.String("group", s.groupName),
		zap.Int64("pending_messages", length))

	return nil
}

// Stop 停止异步任务处理器
func (s *SigninAsyncTask) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isStarted {
		return
	}

	s.logger.Debug("正在停止签到异步任务处理器")

	close(s.stopChan)
	s.wg.Wait()

	// Redis Streams会自动处理未确认的消息，重启后会重新投递
	s.logger.Debug("所有worker已停止，未处理的消息将保留在Stream中")

	s.isStarted = false
	s.logger.Debug("签到异步任务处理器已停止")
}

// SubmitTask 提交签到任务
func (s *SigninAsyncTask) SubmitTask(ctx context.Context, task *SigninTask) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isStarted {
		return errors.New("异步任务未启动")
	}

	// 将任务序列化为JSON
	taskJSON, err := json.Marshal(task)
	if err != nil {
		s.logger.Error("序列化签到任务失败",
			zap.Int64("user_id", task.UserID),
			zap.String("trace_id", task.TraceID),
			zap.Error(err))
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	// 构建Stream消息字段
	values := map[string]interface{}{
		"user_id":         task.UserID,
		"sign_date":       task.SignDate.Unix(),
		"continuous_days": task.ContinuousDays,
		"total_days":      task.TotalDays,
		"trace_id":        task.TraceID,
		"retry_count":     0, // 初始重试次数为0
		"created_at":      time.Now().Unix(),
		"task_data":       string(taskJSON), // 完整的任务数据
	}

	// 发送消息到Stream
	messageID, err := s.cache.XAdd(ctx, s.streamName, values)
	if err != nil {
		s.logger.Error("提交签到任务到Stream失败",
			zap.Int64("user_id", task.UserID),
			zap.String("trace_id", task.TraceID),
			zap.Error(err))

		// 检查是否是Redis连接问题
		if strings.Contains(err.Error(), "connection") ||
			strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "network") {
			return errors.New("服务繁忙，请稍后再试")
		}

		return fmt.Errorf("提交任务失败: %w", err)
	}

	s.logger.Debug("提交签到任务成功",
		zap.Int64("user_id", task.UserID),
		zap.String("trace_id", task.TraceID),
		zap.String("message_id", messageID))

	return nil
}

// worker 工作协程
func (s *SigninAsyncTask) worker(workerID int) {
	defer s.wg.Done()
	ctx := context.Background()

	s.logger.Debug("签到异步任务worker启动", zap.Int("worker_id", workerID))
	consumerName := fmt.Sprintf("worker-%d", workerID)

	for {
		select {
		case <-s.stopChan:
			s.logger.Debug("签到异步任务worker停止", zap.Int("worker_id", workerID))
			return
		default:
			// 从消费者组读取消息，设置1秒超时避免关闭时阻塞
			streams := map[string]string{s.streamName: ">"} // ">"表示读取新消息
			messages, err := s.cache.XReadGroup(ctx, s.groupName, consumerName, streams, 1, 1*time.Second)

			if err != nil {
				s.logger.Error("从Stream读取消息失败",
					zap.Int("worker_id", workerID),
					zap.Error(err))
				time.Sleep(1 * time.Second) // 等待1秒后重试
				continue
			}

			// 如果没有消息，继续循环（不需要额外休眠，因为XReadGroup已经阻塞了1秒）
			if len(messages) == 0 {
				continue
			}

			// 处理消息
			for _, message := range messages {
				s.processStreamMessage(ctx, message, workerID)
			}
		}
	}
}

// processStreamMessage 处理Stream消息
func (s *SigninAsyncTask) processStreamMessage(ctx context.Context, message map[string]interface{}, workerID int) {
	messageID, ok := message["message_id"].(string)
	if !ok {
		s.logger.Error("消息ID缺失", zap.Int("worker_id", workerID))
		return
	}

	// 解析任务数据
	taskData, ok := message["task_data"].(string)
	if !ok {
		s.logger.Error("任务数据缺失",
			zap.String("message_id", messageID),
			zap.Int("worker_id", workerID))
		// 确认消息避免重复处理
		s.ackMessage(ctx, messageID)
		return
	}

	// 反序列化任务
	var task SigninTask
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		s.logger.Error("反序列化任务失败",
			zap.String("message_id", messageID),
			zap.Int("worker_id", workerID),
			zap.Error(err))
		// 确认消息避免重复处理
		s.ackMessage(ctx, messageID)
		return
	}

	// 获取重试次数
	retryCount := int64(0)
	if rc, exists := message["retry_count"]; exists {
		if rcInt, ok := rc.(int64); ok {
			retryCount = rcInt
		}
	}

	s.logger.Info("开始处理Stream签到任务",
		zap.String("message_id", messageID),
		zap.Int64("user_id", task.UserID),
		zap.Int("worker_id", workerID),
		zap.Int64("retry_count", retryCount),
		zap.String("trace_id", task.TraceID))

	// 处理任务
	success := s.processTaskWithRetry(&task, workerID, messageID, retryCount)

	if success {
		// 任务成功，确认消息
		s.ackMessage(ctx, messageID)
	} else {
		// 任务失败，检查重试次数
		if retryCount >= 3 { // 最大重试3次
			s.moveToDeadLetterQueue(ctx, messageID, message, &task)
			s.ackMessage(ctx, messageID)
		} else {
			// 重新投递消息增加重试次数
			s.requeueMessage(ctx, messageID, message, &task, retryCount+1)
			s.ackMessage(ctx, messageID)
		}
	}
}

// processTaskWithRetry 处理任务并返回是否成功
func (s *SigninAsyncTask) processTaskWithRetry(task *SigninTask, workerID int, messageID string, retryCount int64) bool {
	// 创建带链路ID的context
	ctx := context.Background()
	if task.TraceID != "" {
		ctx = tracing.WithTraceID(ctx, task.TraceID)
	}

	startTime := time.Now()

	// 处理签到状态表
	err := s.updateSigninStatus(ctx, task)
	if err != nil {
		s.logger.Error("更新签到状态失败",
			zap.String("message_id", messageID),
			zap.Int64("user_id", task.UserID),
			zap.Int64("retry_count", retryCount),
			zap.Int("worker_id", workerID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return false
	}

	// 处理签到日志表
	err = s.insertSigninLog(ctx, task)
	if err != nil {
		s.logger.Error("插入签到日志失败",
			zap.String("message_id", messageID),
			zap.Int64("user_id", task.UserID),
			zap.Int64("retry_count", retryCount),
			zap.Int("worker_id", workerID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return false
	}

	duration := time.Since(startTime)
	s.logger.Info("签到任务处理完成",
		zap.String("message_id", messageID),
		zap.Int64("user_id", task.UserID),
		zap.Duration("duration", duration),
		zap.Int64("retry_count", retryCount),
		zap.Int("worker_id", workerID),
		tracing.WithTraceIDField(ctx))

	return true
}

// ackMessage 确认消息已处理
func (s *SigninAsyncTask) ackMessage(ctx context.Context, messageID string) {
	_, err := s.cache.XAck(ctx, s.streamName, s.groupName, messageID)
	if err != nil {
		s.logger.Error("确认消息失败",
			zap.String("message_id", messageID),
			zap.Error(err))
	}
}

// moveToDeadLetterQueue 移动消息到死信队列
func (s *SigninAsyncTask) moveToDeadLetterQueue(ctx context.Context, messageID string, message map[string]interface{}, task *SigninTask) {
	// 构建死信队列消息
	dlqValues := map[string]interface{}{
		"original_message_id": messageID,
		"user_id":             task.UserID,
		"trace_id":            task.TraceID,
		"failed_at":           time.Now().Unix(),
		"failure_reason":      "max_retries_exceeded",
		"original_message":    message,
	}

	// 发送到死信队列Stream
	dlqStream := "signin:task:dlq"
	_, err := s.cache.XAdd(ctx, dlqStream, dlqValues)
	if err != nil {
		s.logger.Error("移动消息到死信队列失败",
			zap.String("message_id", messageID),
			zap.Int64("user_id", task.UserID),
			zap.Error(err))
	} else {
		s.logger.Warn("消息已移动到死信队列",
			zap.String("message_id", messageID),
			zap.Int64("user_id", task.UserID),
			zap.String("trace_id", task.TraceID))
	}
}

// requeueMessage 重新投递消息增加重试次数
func (s *SigninAsyncTask) requeueMessage(ctx context.Context, messageID string, message map[string]interface{}, task *SigninTask, newRetryCount int64) {
	// 重新序列化任务数据
	taskJSON, err := json.Marshal(task)
	if err != nil {
		s.logger.Error("重新序列化任务失败",
			zap.String("message_id", messageID),
			zap.Error(err))
		return
	}

	// 构建新的消息字段
	newValues := map[string]interface{}{
		"user_id":         task.UserID,
		"sign_date":       task.SignDate.Unix(),
		"continuous_days": task.ContinuousDays,
		"total_days":      task.TotalDays,
		"trace_id":        task.TraceID,
		"retry_count":     newRetryCount,
		"created_at":      message["created_at"],
		"task_data":       string(taskJSON),
	}

	// 重新投递消息
	newMessageID, err := s.cache.XAdd(ctx, s.streamName, newValues)
	if err != nil {
		s.logger.Error("重新投递消息失败",
			zap.String("original_message_id", messageID),
			zap.Int64("user_id", task.UserID),
			zap.Int64("new_retry_count", newRetryCount),
			zap.Error(err))
	} else {
		s.logger.Info("消息已重新投递",
			zap.String("original_message_id", messageID),
			zap.String("new_message_id", newMessageID),
			zap.Int64("user_id", task.UserID),
			zap.Int64("new_retry_count", newRetryCount))
	}
}

// updateSigninStatus 更新签到状态表
func (s *SigninAsyncTask) updateSigninStatus(ctx context.Context, task *SigninTask) error {
	// 先查询是否存在记录
	existing, err := s.db.UserSigninStatus.Query().
		Where(usersigninstatus.UserID(task.UserID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 记录不存在，创建新记录
			err = s.db.UserSigninStatus.Create().
				SetUserID(task.UserID).
				SetLastSigninDate(task.SignDate).
				SetContinuousDays(task.ContinuousDays).
				SetTotalDays(task.TotalDays).
				Exec(ctx)
		} else {
			return err
		}
	} else {
		// 记录存在，更新记录
		_, err = s.db.UserSigninStatus.UpdateOne(existing).
			SetLastSigninDate(task.SignDate).
			SetContinuousDays(task.ContinuousDays).
			SetTotalDays(task.TotalDays).
			Save(ctx)
	}

	if err != nil {
		return err
	}

	s.logger.Debug("签到状态更新成功",
		zap.Int64("user_id", task.UserID),
		zap.Time("sign_date", task.SignDate),
		zap.Int("continuous_days", task.ContinuousDays),
		zap.Int("total_days", task.TotalDays),
		tracing.WithTraceIDField(ctx))

	return nil
}

func (s *SigninAsyncTask) insertSigninLog(ctx context.Context, task *SigninTask) error {
	// 检查今日是否已签到（防止重复插入）
	exists, err := s.db.UserSigninLogs.Query().
		Where(
			usersigninlogs.UserID(task.UserID),
			usersigninlogs.SignDate(task.SignDate),
		).
		Exist(ctx)

	if err != nil {
		return err
	}

	if exists {
		s.logger.Debug("签到日志已存在，跳过插入",
			zap.Int64("user_id", task.UserID),
			zap.Time("sign_date", task.SignDate),
			tracing.WithTraceIDField(ctx))
		return nil
	}

	// 插入新的签到日志
	err = s.db.UserSigninLogs.Create().
		SetUserID(task.UserID).
		SetSignDate(task.SignDate).
		Exec(ctx)

	if err != nil {
		return err
	}

	s.logger.Debug("签到日志插入成功",
		zap.Int64("user_id", task.UserID),
		zap.Time("sign_date", task.SignDate),
		tracing.WithTraceIDField(ctx))

	return nil
}

// GetQueueSize 获取队列大小（Stream中的消息数量）
func (s *SigninAsyncTask) GetQueueSize(ctx context.Context) int {
	length, err := s.cache.XLen(ctx, s.streamName)
	if err != nil {
		s.logger.Error("获取Stream长度失败", zap.Error(err))
		return 0
	}
	return int(length)
}

// GetPendingCount 获取待处理消息数量
func (s *SigninAsyncTask) GetPendingCount(ctx context.Context) (int64, error) {
	pending, err := s.cache.XPending(ctx, s.streamName, s.groupName)
	if err != nil {
		return 0, fmt.Errorf("获取待处理消息失败: %w", err)
	}

	if total, ok := pending["total"].(int64); ok {
		return total, nil
	}

	return 0, nil
}

// IsStarted 检查是否已启动
func (s *SigninAsyncTask) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isStarted
}
