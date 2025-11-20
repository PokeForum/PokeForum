package service

import (
	"context"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"go.uber.org/zap"
)

// StatsSyncTask 统计数据同步任务
type StatsSyncTask struct {
	db                  *ent.Client
	cache               cache.ICacheService
	logger              *zap.Logger
	postStatsService    IPostStatsService
	commentStatsService ICommentStatsService
	ticker              *time.Ticker
	stopChan            chan struct{}
}

// NewStatsSyncTask 创建统计数据同步任务实例
func NewStatsSyncTask(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) *StatsSyncTask {
	return &StatsSyncTask{
		db:                  db,
		cache:               cacheService,
		logger:              logger,
		postStatsService:    NewPostStatsService(db, cacheService, logger),
		commentStatsService: NewCommentStatsService(db, cacheService, logger),
		stopChan:            make(chan struct{}),
	}
}

// Start 启动同步任务
// interval: 同步间隔时间
func (t *StatsSyncTask) Start(interval time.Duration) {
	t.logger.Debug("启动统计数据同步任务", zap.Duration("interval", interval))

	t.ticker = time.NewTicker(interval)

	go func() {
		// 启动时立即执行一次同步
		t.syncAll()

		// 定时执行同步
		for {
			select {
			case <-t.ticker.C:
				t.syncAll()
			case <-t.stopChan:
				t.logger.Debug("统计数据同步任务已停止")
				return
			}
		}
	}()

	t.logger.Debug("统计数据同步任务已启动")
}

// Stop 停止同步任务
func (t *StatsSyncTask) Stop() {
	t.logger.Debug("正在停止统计数据同步任务")

	if t.ticker != nil {
		t.ticker.Stop()
	}

	close(t.stopChan)

	// 最后执行一次同步,确保数据完整性
	t.syncAll()

	t.logger.Debug("统计数据同步任务已停止")
}

// syncAll 同步所有统计数据
func (t *StatsSyncTask) syncAll() {
	ctx := context.Background()

	t.logger.Debug("开始执行统计数据同步")
	startTime := time.Now()

	// 同步帖子统计数据
	postCount, err := t.postStatsService.SyncStatsToDatabase(ctx)
	if err != nil {
		t.logger.Error("同步帖子统计数据失败", zap.Error(err))
	} else {
		t.logger.Debug("同步帖子统计数据完成", zap.Int("count", postCount))
	}

	// 同步评论统计数据
	commentCount, err := t.commentStatsService.SyncStatsToDatabase(ctx)
	if err != nil {
		t.logger.Error("同步评论统计数据失败", zap.Error(err))
	} else {
		t.logger.Debug("同步评论统计数据完成", zap.Int("count", commentCount))
	}

	duration := time.Since(startTime)
	t.logger.Debug("统计数据同步完成",
		zap.Int("post_count", postCount),
		zap.Int("comment_count", commentCount),
		zap.Duration("duration", duration))
}
