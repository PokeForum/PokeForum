package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
)

// IPostStatsService Post statistics service interface | 帖子统计服务接口
type IPostStatsService interface {
	// PerformAction Perform post action (like/dislike/favorite) | 执行帖子操作(点赞/点踩/收藏)
	// userID: User ID | 用户ID
	// postID: Post ID | 帖子ID
	// actionType: Action type | 操作类型
	// Return: Updated statistics and error | 返回: 更新后的统计数据和错误
	PerformAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error)

	// CancelAction Cancel post action | 取消帖子操作
	// userID: User ID | 用户ID
	// postID: Post ID | 帖子ID
	// actionType: Action type | 操作类型
	// Return: Updated statistics and error | 返回: 更新后的统计数据和错误
	CancelAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error)

	// GetStats Get post statistics | 获取帖子统计数据
	// postID: Post ID | 帖子ID
	// Return: Statistics and error | 返回: 统计数据和错误
	GetStats(ctx context.Context, postID int) (*stats.Stats, error)

	// GetUserActionStatus Get user's action status on post | 获取用户对帖子的操作状态
	// userID: User ID | 用户ID
	// postID: Post ID | 帖子ID
	// Return: User action status and error | 返回: 用户操作状态和错误
	GetUserActionStatus(ctx context.Context, userID, postID int) (*stats.UserActionStatus, error)

	// GetStatsMap Batch get post statistics | 批量获取帖子统计数据
	// postIDs: List of post IDs | 帖子ID列表
	// Return: Mapping of post ID to statistics and error | 返回: 帖子ID到统计数据的映射和错误
	GetStatsMap(ctx context.Context, postIDs []int) (map[int]*stats.Stats, error)

	// IncrViewCount Increment post view count | 增加帖子浏览数
	// postID: Post ID | 帖子ID
	// Return: Error | 返回: 错误
	IncrViewCount(ctx context.Context, postID int) error

	// SyncStatsToDatabase Sync statistics to database | 同步统计数据到数据库
	// Get post IDs that need syncing from Redis dirty set, aggregate real data from PostAction table, update Post table | 从Redis的dirty集合获取需要同步的帖子ID,批量聚合PostAction表统计真实数据,更新Post表
	// Return: Sync count and error | 返回: 同步数量和错误
	SyncStatsToDatabase(ctx context.Context) (int, error)
}

// PostStatsService Post statistics service implementation | 帖子统计服务实现
type PostStatsService struct {
	db          *ent.Client
	cache       cache.ICacheService
	statsHelper *stats.Helper
	logger      *zap.Logger
}

// NewPostStatsService Create a post statistics service instance | 创建帖子统计服务实例
func NewPostStatsService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IPostStatsService {
	return &PostStatsService{
		db:          db,
		cache:       cacheService,
		statsHelper: stats.NewStatsHelper(cacheService, logger),
		logger:      logger,
	}
}

// PerformAction Perform post action (like/dislike/favorite) | 执行帖子操作(点赞/点踩/收藏)
func (s *PostStatsService) PerformAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("执行帖子操作",
		zap.Int("user_id", userID),
		zap.Int("post_id", postID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	exists, err := s.db.Post.Query().Where(post.IDEQ(postID)).Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查帖子是否存在失败: %w", err)
	}
	if !exists {
		return nil, errors.New("帖子不存在")
	}

	// Start database transaction | 开启数据库事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		s.logger.Error("开启事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback() //nolint:errcheck // No need to handle rollback failure during panic recovery | panic恢复时回滚失败无需处理
			panic(v)
		}
	}()

	// Check if action already exists | 检查是否已经存在该操作
	existingAction, err := tx.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
			postaction.ActionTypeEQ(postaction.ActionType(actionType)),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		_ = tx.Rollback() //nolint:errcheck // No need to handle rollback failure during error handling | 错误处理时回滚失败无需处理
		s.logger.Error("查询操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询操作记录失败: %w", err)
	}

	// If already exists, return current statistics | 如果已存在,直接返回当前统计数据
	if existingAction != nil {
		_ = tx.Rollback() //nolint:errcheck // No need to handle rollback failure when returning early | 提前返回时回滚失败无需处理
		return s.GetStats(ctx, postID)
	}

	// Handle mutually exclusive logic for like and dislike | 处理点赞和点踩互斥逻辑
	if actionType == stats.ActionTypeLike || actionType == stats.ActionTypeDislike {
		// Delete opposite action | 删除相反的操作
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
			_ = tx.Rollback() //nolint:errcheck // No need to handle rollback failure during error handling | 错误处理时回滚失败无需处理
			s.logger.Error("删除相反操作失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("删除相反操作失败: %w", err)
		}

		// If opposite action was deleted, need to update Redis count | 如果删除了相反操作,需要更新Redis计数
		if deletedCount > 0 {
			oppositeField := s.getStatsField(oppositeType)
			statsKey := stats.GetPostStatsKey(postID)
			_ = s.statsHelper.IncrStats(ctx, statsKey, oppositeField, -1) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

			// Remove user action cache | 移除用户操作缓存
			userActionKey := stats.GetPostUserActionKey(userID, postID)
			_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, oppositeType) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
		}
	}

	// Create new action record | 创建新的操作记录
	_, err = tx.PostAction.Create().
		SetUserID(userID).
		SetPostID(postID).
		SetActionType(postaction.ActionType(actionType)).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback() //nolint:errcheck // No need to handle rollback failure during error handling | 错误处理时回滚失败无需处理
		s.logger.Error("创建操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建操作记录失败: %w", err)
	}

	// Commit transaction | 提交事务
	if err = tx.Commit(); err != nil {
		s.logger.Error("提交事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// Update Redis statistics (async, failure doesn't affect main flow) | 更新Redis统计数据(异步,失败不影响主流程)
	statsKey := stats.GetPostStatsKey(postID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, 1) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	// Update user action cache | 更新用户操作缓存
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	_ = s.statsHelper.SetUserAction(ctx, userActionKey, actionType) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	// Mark post as dirty data | 标记帖子为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	s.logger.Info("执行帖子操作成功", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, postID)
}

// CancelAction Cancel post action | 取消帖子操作
func (s *PostStatsService) CancelAction(ctx context.Context, userID, postID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("取消帖子操作",
		zap.Int("user_id", userID),
		zap.Int("post_id", postID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// Delete action record | 删除操作记录
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

	// If no records were deleted, action doesn't exist | 如果没有删除任何记录,说明操作不存在
	if deletedCount == 0 {
		return s.GetStats(ctx, postID)
	}

	// Update Redis statistics (async, failure doesn't affect main flow) | 更新Redis统计数据(异步,失败不影响主流程)
	statsKey := stats.GetPostStatsKey(postID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, -1) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	// Remove user action cache | 移除用户操作缓存
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, actionType) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	// Mark post as dirty data | 标记帖子为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	s.logger.Info("取消帖子操作成功", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, postID)
}

// GetStats Get post statistics | 获取帖子统计数据
func (s *PostStatsService) GetStats(ctx context.Context, postID int) (*stats.Stats, error) {
	// Read from Redis first | 优先从Redis读取
	statsKey := stats.GetPostStatsKey(postID)
	fields := []string{"like_count", "dislike_count", "favorite_count", "view_count"}

	statData, err := s.statsHelper.GetStats(ctx, statsKey, fields)
	if err == nil && len(statData) > 0 {
		// Check if there is valid data | 检查是否有有效数据
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

	// Redis cache miss, read from database | Redis未命中,从数据库读取
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

	// Backfill Redis cache | 回填Redis缓存
	_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{ //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
		"like_count":     result.LikeCount,
		"dislike_count":  result.DislikeCount,
		"favorite_count": result.FavoriteCount,
		"view_count":     result.ViewCount,
	})

	return result, nil
}

// GetUserActionStatus Get user's action status on post | 获取用户对帖子的操作状态
func (s *PostStatsService) GetUserActionStatus(ctx context.Context, userID, postID int) (*stats.UserActionStatus, error) {
	// Read from Redis first | 优先从Redis读取
	userActionKey := stats.GetPostUserActionKey(userID, postID)
	actions, err := s.statsHelper.GetUserActions(ctx, userActionKey)
	if err == nil && len(actions) > 0 {
		return &stats.UserActionStatus{
			HasLiked:     actions[string(stats.ActionTypeLike)],
			HasDisliked:  actions[string(stats.ActionTypeDislike)],
			HasFavorited: actions[string(stats.ActionTypeFavorite)],
		}, nil
	}

	// Redis cache miss, read from database | Redis未命中,从数据库读取
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
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeLike) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
		case postaction.ActionTypeDislike:
			result.HasDisliked = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeDislike) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
		case postaction.ActionTypeFavorite:
			result.HasFavorited = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeFavorite) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
		}
	}

	return result, nil
}

// GetStatsMap Batch get post statistics | 批量获取帖子统计数据
func (s *PostStatsService) GetStatsMap(ctx context.Context, postIDs []int) (map[int]*stats.Stats, error) {
	result := make(map[int]*stats.Stats)
	for _, id := range postIDs {
		statsData, err := s.GetStats(ctx, id)
		if err != nil {
			s.logger.Warn("获取帖子统计失败", zap.Int("post_id", id), zap.Error(err))
			continue
		}
		result[id] = statsData
	}
	return result, nil
}

// IncrViewCount Increment post view count | 增加帖子浏览数
func (s *PostStatsService) IncrViewCount(ctx context.Context, postID int) error {
	// Increment view count directly in Redis | 直接在Redis中增加浏览数
	statsKey := stats.GetPostStatsKey(postID)
	_ = s.statsHelper.IncrStats(ctx, statsKey, "view_count", 1) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	// Mark post as dirty data (async sync to database) | 标记帖子为脏数据(异步同步到数据库)
	_ = s.statsHelper.MarkDirty(ctx, stats.PostDirtySetKey, postID) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

	return nil
}

// SyncStatsToDatabase Sync statistics to database | 同步统计数据到数据库
func (s *PostStatsService) SyncStatsToDatabase(ctx context.Context) (int, error) {
	// Get all dirty data IDs | 获取所有脏数据ID
	dirtyIDs, err := s.statsHelper.GetDirtyIDs(ctx, stats.PostDirtySetKey)
	if err != nil {
		s.logger.Error("获取脏数据ID失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return 0, err
	}

	if len(dirtyIDs) == 0 {
		return 0, nil
	}

	s.logger.Info("需要同步的帖子数量", zap.Int("count", len(dirtyIDs)), tracing.WithTraceIDField(ctx))

	syncCount := 0
	// Batch processing, 100 items at a time | 批量处理,每次处理100个
	batchSize := 100
	for i := 0; i < len(dirtyIDs); i += batchSize {
		end := i + batchSize
		if end > len(dirtyIDs) {
			end = len(dirtyIDs)
		}
		batch := dirtyIDs[i:end]

		// Process this batch | 处理这一批数据
		count := s.syncBatch(ctx, batch)
		syncCount += count
	}

	s.logger.Info("同步帖子统计数据完成",
		zap.Int("total", len(dirtyIDs)),
		zap.Int("success", syncCount),
		tracing.WithTraceIDField(ctx))

	return syncCount, nil
}

// syncBatch Batch sync post statistics | 批量同步帖子统计数据
func (s *PostStatsService) syncBatch(ctx context.Context, postIDs []int) int {
	syncCount := 0

	for _, postID := range postIDs {
		// Aggregate real data from PostAction table | 从PostAction表聚合真实数据
		var likeCount, dislikeCount, favoriteCount int

		// Count likes | 统计点赞数
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

		// Count dislikes | 统计点踩数
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

		// Count favorites | 统计收藏数
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

		// Get view count from Redis (view count is only maintained in Redis) | 从Redis获取浏览数(浏览数只在Redis中维护)
		statsKey := stats.GetPostStatsKey(postID)
		viewCountStr, _ := s.cache.HGet(ctx, statsKey, "view_count") //nolint:errcheck // Use default value on Redis operation failure | Redis操作失败使用默认值
		viewCount := 0
		if viewCountStr != "" {
			_, _ = fmt.Sscanf(viewCountStr, "%d", &viewCount) //nolint:errcheck // Use default value on parse failure | 解析失败使用默认值
		}

		// Update Post table | 更新Post表
		err = s.db.Post.UpdateOneID(postID).
			SetLikeCount(likeCount).
			SetDislikeCount(dislikeCount).
			SetFavoriteCount(favoriteCount).
			SetViewCount(viewCount).
			Exec(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// Post has been deleted, clean dirty mark and cache | 帖子已被删除,清理脏标记和缓存
				_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.PostDirtySetKey, []int{postID}) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
				_ = s.statsHelper.DeleteStatsCache(ctx, statsKey)                           //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
				s.logger.Warn("帖子不存在,已清理缓存", zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
				continue
			}
			s.logger.Error("更新帖子统计失败", zap.Int("post_id", postID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// Update Redis cache | 更新Redis缓存
		_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{ //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程
			"like_count":     likeCount,
			"dislike_count":  dislikeCount,
			"favorite_count": favoriteCount,
			"view_count":     viewCount,
		})

		// Clear dirty mark | 清除脏标记
		_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.PostDirtySetKey, []int{postID}) //nolint:errcheck // Redis operation failure doesn't affect main flow | Redis操作失败不影响主流程

		syncCount++
	}

	return syncCount
}

// getStatsField Get corresponding statistics field name based on action type | 根据操作类型获取对应的统计字段名
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
