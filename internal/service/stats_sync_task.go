package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	pkgasynq "github.com/PokeForum/PokeForum/internal/pkg/asynq"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
)

// StatsSyncTask Statistics data synchronization task | 统计数据同步任务
type StatsSyncTask struct {
	db                  *ent.Client
	cache               cache.ICacheService
	logger              *zap.Logger
	postStatsService    IPostStatsService
	commentStatsService ICommentStatsService
	taskManager         *pkgasynq.TaskManager
}

// StatsSyncPayload Statistics synchronization task payload | 统计同步任务载荷
type StatsSyncPayload struct {
	TriggerTime int64 `json:"trigger_time"`
}

// NewStatsSyncTask Create statistics data synchronization task instance | 创建统计数据同步任务实例
func NewStatsSyncTask(db *ent.Client, cacheService cache.ICacheService, taskManager *pkgasynq.TaskManager, logger *zap.Logger) *StatsSyncTask {
	return &StatsSyncTask{
		db:                  db,
		cache:               cacheService,
		logger:              logger,
		postStatsService:    NewPostStatsService(db, cacheService, logger),
		commentStatsService: NewCommentStatsService(db, cacheService, logger),
		taskManager:         taskManager,
	}
}

// RegisterHandler Register task handler | 注册任务处理器
func (t *StatsSyncTask) RegisterHandler() {
	t.taskManager.RegisterHandlerFunc(pkgasynq.TypeStatsSync, t.HandleStatsSyncTask)
	t.logger.Info("Statistics data synchronization task handler registered | 统计数据同步任务处理器已注册")
}

// RegisterSchedule Register scheduled task | 注册定时任务
// interval: Synchronization interval time | 同步间隔时间
func (t *StatsSyncTask) RegisterSchedule(interval time.Duration) error {
	// Create task | 创建任务
	payload := &StatsSyncPayload{TriggerTime: time.Now().Unix()}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化任务载荷失败: %w", err)
	}

	task := asynq.NewTask(pkgasynq.TypeStatsSync, data)

	// Convert to cron expression, e.g. @every 5m | 转换为cron表达式，如 @every 5m
	cronSpec := fmt.Sprintf("@every %s", interval.String())

	entryID, err := t.taskManager.RegisterSchedule(cronSpec, task, asynq.Queue(pkgasynq.QueueLow))
	if err != nil {
		return fmt.Errorf("注册定时任务失败: %w", err)
	}

	t.logger.Info("Statistics data synchronization scheduled task registered | 统计数据同步定时任务已注册",
		zap.String("entry_id", entryID),
		zap.Duration("interval", interval))

	return nil
}

// HandleStatsSyncTask Handle statistics synchronization task | 处理统计同步任务
func (t *StatsSyncTask) HandleStatsSyncTask(ctx context.Context, task *asynq.Task) error {
	var payload StatsSyncPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.logger.Error("Failed to deserialize statistics synchronization task | 反序列化统计同步任务失败", zap.Error(err))
		return fmt.Errorf("反序列化失败: %v: %w", err, asynq.SkipRetry)
	}

	t.logger.Debug("Start executing statistics data synchronization | 开始执行统计数据同步")
	startTime := time.Now()

	// Synchronize post statistics data | 同步帖子统计数据
	postCount, err := t.postStatsService.SyncStatsToDatabase(ctx)
	if err != nil {
		t.logger.Error("Failed to synchronize post statistics data | 同步帖子统计数据失败", zap.Error(err))
	} else {
		t.logger.Debug("Post statistics data synchronization completed | 同步帖子统计数据完成", zap.Int("count", postCount))
	}

	// Synchronize comment statistics data | 同步评论统计数据
	commentCount, err := t.commentStatsService.SyncStatsToDatabase(ctx)
	if err != nil {
		t.logger.Error("Failed to synchronize comment statistics data | 同步评论统计数据失败", zap.Error(err))
	} else {
		t.logger.Debug("Comment statistics data synchronization completed | 同步评论统计数据完成", zap.Int("count", commentCount))
	}

	duration := time.Since(startTime)
	t.logger.Debug("Statistics data synchronization completed | 统计数据同步完成",
		zap.Int("post_count", postCount),
		zap.Int("comment_count", commentCount),
		zap.Duration("duration", duration))

	return nil
}

// SyncNow Execute synchronization immediately (for startup) | 立即执行一次同步（用于启动时）
func (t *StatsSyncTask) SyncNow(ctx context.Context) {
	t.logger.Debug("Execute statistics data synchronization immediately | 立即执行统计数据同步")

	payload := &StatsSyncPayload{TriggerTime: time.Now().Unix()}
	data, _ := json.Marshal(payload) //nolint:errcheck // 序列化简单结构不会失败
	task := asynq.NewTask(pkgasynq.TypeStatsSync, data)

	_, err := t.taskManager.Enqueue(task, asynq.Queue(pkgasynq.QueueLow))
	if err != nil {
		t.logger.Error("Failed to submit immediate synchronization task | 提交立即同步任务失败", zap.Error(err))
	}
}
