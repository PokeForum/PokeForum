package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"go.uber.org/zap"
)

// IPostStatsService 帖子统计服务接口
type IPostStatsService interface {
	// PerformAction 执行帖子操作(点赞/点踩/收藏)
	// userID: 用户ID
	// postID: 帖子ID
	// actionType: 操作类型
	// 返回: 更新后的统计数据和错误
	PerformAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error)

	// CancelAction 取消帖子操作
	// userID: 用户ID
	// postID: 帖子ID
	// actionType: 操作类型
	// 返回: 更新后的统计数据和错误
	CancelAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error)

	// GetStats 获取帖子统计数据
	// postID: 帖子ID
	// 返回: 统计数据和错误
	GetStats(ctx context.Context, postID int) (*stats.Stats, error)

	// GetUserActionStatus 获取用户对帖子的操作状态
	// userID: 用户ID
	// postID: 帖子ID
	// 返回: 用户操作状态和错误
	GetUserActionStatus(ctx context.Context, userID, postID int) (*stats.UserActionStatus, error)

	// IncrViewCount 增加帖子浏览数
	// postID: 帖子ID
	// 返回: 错误
	IncrViewCount(ctx context.Context, postID int) error

	// SyncStatsToDatabase 同步统计数据到数据库
	// 从Redis的dirty集合获取需要同步的帖子ID,批量聚合PostAction表统计真实数据,更新Post表
	// 返回: 同步数量和错误
	SyncStatsToDatabase(ctx context.Context) (int, error)
}

// PostStatsService 帖子统计服务实现
type PostStatsService struct {
	db          *ent.Client
	cache       cache.ICacheService
	statsHelper *stats.StatsHelper
	logger      *zap.Logger
}

// NewPostStatsService 创建帖子统计服务实例
func NewPostStatsService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IPostStatsService {
	return &PostStatsService{
		db:          db,
		cache:       cacheService,
		statsHelper: stats.NewStatsHelper(cacheService, logger),
		logger:      logger,
	}
}

// PerformAction 执行帖子操作(点赞/点踩/收藏)
func (s *PostStatsService) PerformAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("执行帖子操作",
		zap.Int("user_id", userID),
		zap.Int("post_id", postID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	exists, err := s.db.Post.Query().Where(post.IDEQ(postID)).Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查帖子是否存在失败: %w", err)
	}
	if !exists {
		return nil, errors.New("帖子不存在")
	}

	// 开启数据库事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		s.logger.Error("开启事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	// 检查是否已经存在该操作
	existingAction, err := tx.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
			postaction.ActionTypeEQ(postaction.ActionType(actionType)),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		_ = tx.Rollback()
		s.logger.Error("查询操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询操作记录失败: %w", err)
	}

	// 如果已存在,直接返回当前统计数据
	if existingAction != nil {
		_ = tx.Rollback()
		return s.GetStats(ctx, postID)
	}

	// 处理点赞和点踩互斥逻辑
	if actionType == stats.ActionTypeLike || actionType == stats.ActionTypeDislike {
		// 删除相反的操作
		oppositeType := stats.ActionTypeDislike
		if actionType == stats.ActionTypeDislike {
			oppositeType = stats.ActionTypeLike
		}

		deletedCount, err := tx.PostAction.Delete().
			Where(
				postaction.UserIDEQ(userID),
				postaction.PostIDEQ(postID),
				postaction.ActionTypeEQ(postaction.ActionType(oppositeType)),
			).
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			s.logger.Error("删除相反操作失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("删除相反操作失败: %w", err)
		}

		// 如果删除了相反操作,需要更新Redis计数
		if deletedCount > 0 {
			oppositeField := s.getStatsField(oppositeType)
			statsKey := stats.GetPostStatsKey(postID)
			_ = s.statsHelper.IncrStats(ctx, statsKey, oppositeField, -1)

			// 移除用户操作缓存
			userActionKey := stats.GetPostUserActionKey(userID, postID)
			_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, oppositeType)
		}
	}

	// 创建新的操作记录
	_, err = tx.PostAction.Create().
		SetUserID(userID).
		SetPostID(postID).
		SetActionType(postaction.ActionType(actionType)).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		s.logger.Error("创建操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建操作记录失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		s.logger.Error("提交事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新Redis统计数据(异步,失败不影响主流程)
	statsKey := stats.GetPostStatsKey(postID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, 1)

	// 更新用户操作缓存
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	_ = s.statsHelper.SetUserAction(ctx, userActionKey, actionType)

	// 标记帖子为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID)

	s.logger.Info("执行帖子操作成功", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, postID)
}

// CancelAction 取消帖子操作
func (s *PostStatsService) CancelAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("取消帖子操作",
		zap.Int("user_id", userID),
		zap.Int("post_id", postID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// 删除操作记录
	deletedCount, err := s.db.PostAction.Delete().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
			postaction.ActionTypeEQ(postaction.ActionType(actionType)),
		).
		Exec(ctx)
	if err != nil {
		s.logger.Error("删除操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("删除操作记录失败: %w", err)
	}

	// 如果没有删除任何记录,说明操作不存在
	if deletedCount == 0 {
		return s.GetStats(ctx, postID)
	}

	// 更新Redis统计数据(异步,失败不影响主流程)
	statsKey := stats.GetPostStatsKey(postID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, -1)

	// 移除用户操作缓存
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, actionType)

	// 标记帖子为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID)

	s.logger.Info("取消帖子操作成功", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, postID)
}

// GetStats 获取帖子统计数据
func (s *PostStatsService) GetStats(ctx context.Context, postID int) (*stats.Stats, error) {
	// 优先从Redis读取
	statsKey := stats.GetPostStatsKey(postID)
	fields := []string{"like_count", "dislike_count", "favorite_count", "view_count"}
	
	statData, err := s.statsHelper.GetStats(ctx, statsKey, fields)
	if err == nil && len(statData) > 0 {
		// 检查是否有有效数据
		hasData := false
		for _, v := range statData {
			if v > 0 {
				hasData = true
				break
			}
		}
		if hasData {
			return &stats.Stats{
				ID:            postID,
				LikeCount:     statData["like_count"],
				DislikeCount:  statData["dislike_count"],
				FavoriteCount: statData["favorite_count"],
				ViewCount:     statData["view_count"],
			}, nil
		}
	}

	// Redis未命中,从数据库读取
	postData, err := s.db.Post.Query().Where(post.IDEQ(postID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("从数据库获取帖子统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("从数据库获取帖子统计失败: %w", err)
	}

	result := &stats.Stats{
		ID:            postData.ID,
		LikeCount:     postData.LikeCount,
		DislikeCount:  postData.DislikeCount,
		FavoriteCount: postData.FavoriteCount,
		ViewCount:     postData.ViewCount,
	}

	// 回填Redis缓存
	_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{
		"like_count":     result.LikeCount,
		"dislike_count":  result.DislikeCount,
		"favorite_count": result.FavoriteCount,
		"view_count":     result.ViewCount,
	})

	return result, nil
}

// GetUserActionStatus 获取用户对帖子的操作状态
func (s *PostStatsService) GetUserActionStatus(ctx context.Context, userID, postID int) (*stats.UserActionStatus, error) {
	// 优先从Redis读取
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	actions, err := s.statsHelper.GetUserActions(ctx, userActionKey)
	if err == nil && len(actions) > 0 {
		return &stats.UserActionStatus{
			HasLiked:     actions[string(stats.ActionTypeLike)],
			HasDisliked:  actions[string(stats.ActionTypeDislike)],
			HasFavorited: actions[string(stats.ActionTypeFavorite)],
		}, nil
	}

	// Redis未命中,从数据库读取
	userActions, err := s.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
		).
		All(ctx)
	if err != nil {
		s.logger.Error("从数据库获取用户操作状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("从数据库获取用户操作状态失败: %w", err)
	}

	result := &stats.UserActionStatus{
		HasLiked:     false,
		HasDisliked:  false,
		HasFavorited: false,
	}

	for _, action := range userActions {
		switch action.ActionType {
		case postaction.ActionTypeLike:
			result.HasLiked = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeLike)
		case postaction.ActionTypeDislike:
			result.HasDisliked = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeDislike)
		case postaction.ActionTypeFavorite:
			result.HasFavorited = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeFavorite)
		}
	}

	return result, nil
}

// IncrViewCount 增加帖子浏览数
func (s *PostStatsService) IncrViewCount(ctx context.Context, postID int) error {
	// 直接在Redis中增加浏览数
	statsKey := stats.GetPostStatsKey(postID)
	_ = s.statsHelper.IncrStats(ctx, statsKey, "view_count", 1)

	// 标记帖子为脏数据(异步同步到数据库)
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID)

	return nil
}

// SyncStatsToDatabase 同步统计数据到数据库
func (s *PostStatsService) SyncStatsToDatabase(ctx context.Context) (int, error) {
	s.logger.Info("开始同步帖子统计数据到数据库", tracing.WithTraceIDField(ctx))

	// 获取所有脏数据ID
	dirtyIDs, err := s.statsHelper.GetDirtyIDs(ctx, stats.PostDirtySetKey)
	if err != nil {
		s.logger.Error("获取脏数据ID失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return 0, err
	}

	if len(dirtyIDs) == 0 {
		s.logger.Info("没有需要同步的帖子数据", tracing.WithTraceIDField(ctx))
		return 0, nil
	}

	s.logger.Info("需要同步的帖子数量", zap.Int("count", len(dirtyIDs)), tracing.WithTraceIDField(ctx))

	syncCount := 0
	// 批量处理,每次处理100个
	batchSize := 100
	for i := 0; i < len(dirtyIDs); i += batchSize {
		end := i + batchSize
		if end > len(dirtyIDs) {
			end = len(dirtyIDs)
		}
		batch := dirtyIDs[i:end]

		// 处理这一批数据
		count, err := s.syncBatch(ctx, batch)
		if err != nil {
			s.logger.Error("批量同步失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}
		syncCount += count
	}

	s.logger.Info("同步帖子统计数据完成",
		zap.Int("total", len(dirtyIDs)),
		zap.Int("success", syncCount),
		tracing.WithTraceIDField(ctx))

	return syncCount, nil
}

// syncBatch 批量同步帖子统计数据
func (s *PostStatsService) syncBatch(ctx context.Context, postIDs []int) (int, error) {
	syncCount := 0

	for _, postID := range postIDs {
		// 从PostAction表聚合真实数据
		var likeCount, dislikeCount, favoriteCount int

		// 统计点赞数
		likeCount, err := s.db.PostAction.Query().
			Where(
				postaction.PostIDEQ(postID),
				postaction.ActionTypeEQ(postaction.ActionTypeLike),
			).
			Count(ctx)
		if err != nil {
			s.logger.Error("统计点赞数失败", zap.Int("post_id", postID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 统计点踩数
		dislikeCount, err = s.db.PostAction.Query().
			Where(
				postaction.PostIDEQ(postID),
				postaction.ActionTypeEQ(postaction.ActionTypeDislike),
			).
			Count(ctx)
		if err != nil {
			s.logger.Error("统计点踩数失败", zap.Int("post_id", postID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 统计收藏数
		favoriteCount, err = s.db.PostAction.Query().
			Where(
				postaction.PostIDEQ(postID),
				postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
			).
			Count(ctx)
		if err != nil {
			s.logger.Error("统计收藏数失败", zap.Int("post_id", postID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 从Redis获取浏览数(浏览数只在Redis中维护)
		statsKey := stats.GetPostStatsKey(postID)
		viewCountStr, _ := s.cache.HGet(statsKey, "view_count")
		viewCount := 0
		if viewCountStr != "" {
			fmt.Sscanf(viewCountStr, "%d", &viewCount)
		}

		// 更新Post表
		err = s.db.Post.UpdateOneID(postID).
			SetLikeCount(likeCount).
			SetDislikeCount(dislikeCount).
			SetFavoriteCount(favoriteCount).
			SetViewCount(viewCount).
			Exec(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// 帖子已被删除,清理脏标记和缓存
				_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.PostDirtySetKey, []int{postID})
				_ = s.statsHelper.DeleteStatsCache(ctx, statsKey)
				s.logger.Warn("帖子不存在,已清理缓存", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
				continue
			}
			s.logger.Error("更新帖子统计失败", zap.Int("post_id", postID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 更新Redis缓存
		_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{
			"like_count":     likeCount,
			"dislike_count":  dislikeCount,
			"favorite_count": favoriteCount,
			"view_count":     viewCount,
		})

		// 清除脏标记
		_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.PostDirtySetKey, []int{postID})

		syncCount++
	}

	return syncCount, nil
}

// getStatsField 根据操作类型获取对应的统计字段名
func (s *PostStatsService) getStatsField(actionType stats.ActionType) string {
	switch actionType {
	case stats.ActionTypeLike:
		return "like_count"
	case stats.ActionTypeDislike:
		return "dislike_count"
	case stats.ActionTypeFavorite:
		return "favorite_count"
	default:
		return ""
	}
}
