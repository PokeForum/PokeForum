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
	"github.com/PokeForum/PokeForum/internal/repository"
)

// RankingSyncTask 排行榜缓存刷新任务
type RankingSyncTask struct {
	db             *ent.Client
	cache          cache.ICacheService
	logger         *zap.Logger
	rankingService *RankingService
	taskManager    *pkgasynq.TaskManager
}

// RankingSyncPayload 排行榜刷新任务载荷
type RankingSyncPayload struct {
	TriggerTime int64 `json:"trigger_time"`
}

// NewRankingSyncTask 创建排行榜缓存刷新任务实例
func NewRankingSyncTask(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, taskManager *pkgasynq.TaskManager, logger *zap.Logger) *RankingSyncTask {
	return &RankingSyncTask{
		db:          db,
		cache:       cacheService,
		logger:      logger,
		taskManager: taskManager,
		rankingService: &RankingService{
			db:           db,
			userRepo:     repos.User,
			categoryRepo: repos.Category,
			cache:        cacheService,
			log:          logger,
		},
	}
}

// RegisterHandler 注册任务处理器
func (t *RankingSyncTask) RegisterHandler() {
	t.taskManager.RegisterHandlerFunc(pkgasynq.TypeRankingRefresh, t.HandleRankingSyncTask)
	t.logger.Info("排行榜缓存刷新任务处理器已注册")
}

// RegisterSchedule 注册定时任务（每15分钟执行一次）
func (t *RankingSyncTask) RegisterSchedule() error {
	payload := &RankingSyncPayload{TriggerTime: time.Now().Unix()}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化任务载荷失败: %w", err)
	}

	task := asynq.NewTask(pkgasynq.TypeRankingRefresh, data)

	// 每15分钟执行一次
	cronSpec := "@every 15m"

	entryID, err := t.taskManager.RegisterSchedule(cronSpec, task, asynq.Queue(pkgasynq.QueueLow))
	if err != nil {
		return fmt.Errorf("注册定时任务失败: %w", err)
	}

	t.logger.Info("排行榜缓存刷新定时任务已注册",
		zap.String("entry_id", entryID),
		zap.String("interval", "15m"))

	return nil
}

// HandleRankingSyncTask 处理排行榜缓存刷新任务
func (t *RankingSyncTask) HandleRankingSyncTask(ctx context.Context, task *asynq.Task) error {
	var payload RankingSyncPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.logger.Error("反序列化排行榜刷新任务失败", zap.Error(err))
		return fmt.Errorf("反序列化失败: %v: %w", err, asynq.SkipRetry)
	}

	startTime := time.Now()

	// 刷新所有排行榜缓存
	if err := t.rankingService.RefreshAllRankings(ctx); err != nil {
		t.logger.Error("刷新排行榜缓存失败", zap.Error(err))
		return err
	}

	duration := time.Since(startTime)
	t.logger.Debug("排行榜缓存刷新完成", zap.Duration("duration", duration))

	return nil
}

// SyncNow 立即执行一次刷新（用于启动时）
func (t *RankingSyncTask) SyncNow(ctx context.Context) {
	t.logger.Debug("立即执行排行榜缓存刷新")

	payload := &RankingSyncPayload{TriggerTime: time.Now().Unix()}
	data, _ := json.Marshal(payload) //nolint:errcheck // Marshal simple struct never fails
	task := asynq.NewTask(pkgasynq.TypeRankingRefresh, data)

	_, err := t.taskManager.Enqueue(task, asynq.Queue(pkgasynq.QueueLow))
	if err != nil {
		t.logger.Error("提交立即刷新任务失败", zap.Error(err))
	}
}
