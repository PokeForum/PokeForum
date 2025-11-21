package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/commentaction"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"go.uber.org/zap"
)

// ICommentStatsService 评论统计服务接口
type ICommentStatsService interface {
	// PerformAction 执行评论操作(点赞/点踩)
	// userID: 用户ID
	// commentID: 评论ID
	// actionType: 操作类型
	// 返回: 更新后的统计数据和错误
	PerformAction(ctx context.Context, userID, commentID int, actionType stats.ActionType) (*stats.Stats, error)

	// CancelAction 取消评论操作
	// userID: 用户ID
	// commentID: 评论ID
	// actionType: 操作类型
	// 返回: 更新后的统计数据和错误
	CancelAction(ctx context.Context, userID, commentID int, actionType stats.ActionType) (*stats.Stats, error)

	// GetStats 获取评论统计数据
	// commentID: 评论ID
	// 返回: 统计数据和错误
	GetStats(ctx context.Context, commentID int) (*stats.Stats, error)

	// GetUserActionStatus 获取用户对评论的操作状态
	// userID: 用户ID
	// commentID: 评论ID
	// 返回: 用户操作状态和错误
	GetUserActionStatus(ctx context.Context, userID, commentID int) (*stats.UserActionStatus, error)

	// GetStatsMap 批量获取评论统计数据
	// commentIDs: 评论ID列表
	// 返回: 评论ID到统计数据的映射和错误
	GetStatsMap(ctx context.Context, commentIDs []int) (map[int]*stats.Stats, error)

	// SyncStatsToDatabase 同步统计数据到数据库
	// 从Redis的dirty集合获取需要同步的评论ID,批量聚合CommentAction表统计真实数据,更新Comment表
	// 返回: 同步数量和错误
	SyncStatsToDatabase(ctx context.Context) (int, error)
}

// CommentStatsService 评论统计服务实现
type CommentStatsService struct {
	db          *ent.Client
	cache       cache.ICacheService
	statsHelper *stats.Helper
	logger      *zap.Logger
}

// NewCommentStatsService 创建评论统计服务实例
func NewCommentStatsService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ICommentStatsService {
	return &CommentStatsService{
		db:          db,
		cache:       cacheService,
		statsHelper: stats.NewStatsHelper(cacheService, logger),
		logger:      logger,
	}
}

// PerformAction 执行评论操作(点赞/点踩)
func (s *CommentStatsService) PerformAction(ctx context.Context, userID, commentID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("执行评论操作",
		zap.Int("user_id", userID),
		zap.Int("comment_id", commentID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// 检查评论是否存在
	exists, err := s.db.Comment.Query().Where(comment.IDEQ(commentID)).Exist(ctx)
	if err != nil {
		s.logger.Error("检查评论是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查评论是否存在失败: %w", err)
	}
	if !exists {
		return nil, errors.New("评论不存在")
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
	existingAction, err := tx.CommentAction.Query().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
			commentaction.ActionTypeEQ(commentaction.ActionType(actionType)),
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
		return s.GetStats(ctx, commentID)
	}

	// 处理点赞和点踩互斥逻辑
	// 删除相反的操作
	oppositeType := stats.ActionTypeDislike
	if actionType == stats.ActionTypeDislike {
		oppositeType = stats.ActionTypeLike
	}

	deletedCount, err := tx.CommentAction.Delete().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
			commentaction.ActionTypeEQ(commentaction.ActionType(oppositeType)),
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
		statsKey := stats.GetCommentStatsKey(commentID)
		_ = s.statsHelper.IncrStats(ctx, statsKey, oppositeField, -1)

		// 移除用户操作缓存
		userActionKey := stats.GetCommentUserActionKey(userID, commentID)
		_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, oppositeType)
	}

	// 创建新的操作记录
	_, err = tx.CommentAction.Create().
		SetUserID(userID).
		SetCommentID(commentID).
		SetActionType(commentaction.ActionType(actionType)).
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
	statsKey := stats.GetCommentStatsKey(commentID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, 1)

	// 更新用户操作缓存
	userActionKey := stats.GetCommentUserActionKey(userID, commentID)
	_ = s.statsHelper.SetUserAction(ctx, userActionKey, actionType)

	// 标记评论为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.CommentDirtySetKey, commentID)

	s.logger.Info("执行评论操作成功", zap.Int("comment_id", commentID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, commentID)
}

// CancelAction 取消评论操作
func (s *CommentStatsService) CancelAction(ctx context.Context, userID, commentID int, actionType stats.ActionType) (*stats.Stats, error) {
	s.logger.Info("取消评论操作",
		zap.Int("user_id", userID),
		zap.Int("comment_id", commentID),
		zap.String("action_type", string(actionType)),
		tracing.WithTraceIDField(ctx))

	// 删除操作记录
	deletedCount, err := s.db.CommentAction.Delete().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
			commentaction.ActionTypeEQ(commentaction.ActionType(actionType)),
		).
		Exec(ctx)
	if err != nil {
		s.logger.Error("删除操作记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("删除操作记录失败: %w", err)
	}

	// 如果没有删除任何记录,说明操作不存在
	if deletedCount == 0 {
		return s.GetStats(ctx, commentID)
	}

	// 更新Redis统计数据(异步,失败不影响主流程)
	statsKey := stats.GetCommentStatsKey(commentID)
	field := s.getStatsField(actionType)
	_ = s.statsHelper.IncrStats(ctx, statsKey, field, -1)

	// 移除用户操作缓存
	userActionKey := stats.GetCommentUserActionKey(userID, commentID)
	_ = s.statsHelper.RemoveUserAction(ctx, userActionKey, actionType)

	// 标记评论为脏数据
	_ = s.statsHelper.MarkDirty(ctx, stats.CommentDirtySetKey, commentID)

	s.logger.Info("取消评论操作成功", zap.Int("comment_id", commentID), tracing.WithTraceIDField(ctx))
	return s.GetStats(ctx, commentID)
}

// GetStats 获取评论统计数据
func (s *CommentStatsService) GetStats(ctx context.Context, commentID int) (*stats.Stats, error) {
	// 优先从Redis读取
	statsKey := stats.GetCommentStatsKey(commentID)
	fields := []string{"like_count", "dislike_count"}

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
				ID:           commentID,
				LikeCount:    statData["like_count"],
				DislikeCount: statData["dislike_count"],
			}, nil
		}
	}

	// Redis未命中,从数据库读取
	commentData, err := s.db.Comment.Query().Where(comment.IDEQ(commentID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("从数据库获取评论统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("从数据库获取评论统计失败: %w", err)
	}

	result := &stats.Stats{
		ID:           commentData.ID,
		LikeCount:    commentData.LikeCount,
		DislikeCount: commentData.DislikeCount,
	}

	// 回填Redis缓存
	_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{
		"like_count":    result.LikeCount,
		"dislike_count": result.DislikeCount,
	})

	return result, nil
}

// GetUserActionStatus 获取用户对评论的操作状态
func (s *CommentStatsService) GetUserActionStatus(ctx context.Context, userID, commentID int) (*stats.UserActionStatus, error) {
	// 优先从Redis读取
	userActionKey := stats.GetCommentUserActionKey(userID, commentID)
	actions, err := s.statsHelper.GetUserActions(ctx, userActionKey)
	if err == nil && len(actions) > 0 {
		return &stats.UserActionStatus{
			HasLiked:    actions[string(stats.ActionTypeLike)],
			HasDisliked: actions[string(stats.ActionTypeDislike)],
		}, nil
	}

	// Redis未命中,从数据库读取
	userActions, err := s.db.CommentAction.Query().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
		).
		All(ctx)
	if err != nil {
		s.logger.Error("从数据库获取用户操作状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("从数据库获取用户操作状态失败: %w", err)
	}

	result := &stats.UserActionStatus{
		HasLiked:    false,
		HasDisliked: false,
	}

	for _, action := range userActions {
		switch action.ActionType {
		case commentaction.ActionTypeLike:
			result.HasLiked = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeLike)
		case commentaction.ActionTypeDislike:
			result.HasDisliked = true
			_ = s.statsHelper.SetUserAction(ctx, userActionKey, stats.ActionTypeDislike)
		}
	}

	return result, nil
}

// GetStatsMap 批量获取评论统计数据
func (s *CommentStatsService) GetStatsMap(ctx context.Context, commentIDs []int) (map[int]*stats.Stats, error) {
	result := make(map[int]*stats.Stats)
	for _, id := range commentIDs {
		statsData, err := s.GetStats(ctx, id)
		if err != nil {
			s.logger.Warn("获取评论统计失败", zap.Int("comment_id", id), zap.Error(err))
			continue
		}
		result[id] = statsData
	}
	return result, nil
}

// SyncStatsToDatabase 同步统计数据到数据库
func (s *CommentStatsService) SyncStatsToDatabase(ctx context.Context) (int, error) {
	s.logger.Debug("开始同步评论统计数据到数据库", tracing.WithTraceIDField(ctx))

	// 获取所有脏数据ID
	dirtyIDs, err := s.statsHelper.GetDirtyIDs(ctx, stats.CommentDirtySetKey)
	if err != nil {
		s.logger.Error("获取脏数据ID失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return 0, err
	}

	if len(dirtyIDs) == 0 {
		s.logger.Debug("没有需要同步的评论数据", tracing.WithTraceIDField(ctx))
		return 0, nil
	}

	s.logger.Info("需要同步的评论数量", zap.Int("count", len(dirtyIDs)), tracing.WithTraceIDField(ctx))

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

	s.logger.Info("同步评论统计数据完成",
		zap.Int("total", len(dirtyIDs)),
		zap.Int("success", syncCount),
		tracing.WithTraceIDField(ctx))

	return syncCount, nil
}

// syncBatch 批量同步评论统计数据
func (s *CommentStatsService) syncBatch(ctx context.Context, commentIDs []int) (int, error) {
	syncCount := 0

	for _, commentID := range commentIDs {
		// 从CommentAction表聚合真实数据
		var likeCount, dislikeCount int

		// 统计点赞数
		likeCount, err := s.db.CommentAction.Query().
			Where(
				commentaction.CommentIDEQ(commentID),
				commentaction.ActionTypeEQ(commentaction.ActionTypeLike),
			).
			Count(ctx)
		if err != nil {
			s.logger.Error("统计点赞数失败", zap.Int("comment_id", commentID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 统计点踩数
		dislikeCount, err = s.db.CommentAction.Query().
			Where(
				commentaction.CommentIDEQ(commentID),
				commentaction.ActionTypeEQ(commentaction.ActionTypeDislike),
			).
			Count(ctx)
		if err != nil {
			s.logger.Error("统计点踩数失败", zap.Int("comment_id", commentID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 更新Comment表
		err = s.db.Comment.UpdateOneID(commentID).
			SetLikeCount(likeCount).
			SetDislikeCount(dislikeCount).
			Exec(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// 评论已被删除,清理脏标记和缓存
				statsKey := stats.GetCommentStatsKey(commentID)
				_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.CommentDirtySetKey, []int{commentID})
				_ = s.statsHelper.DeleteStatsCache(ctx, statsKey)
				s.logger.Warn("评论不存在,已清理缓存", zap.Int("comment_id", commentID), tracing.WithTraceIDField(ctx))
				continue
			}
			s.logger.Error("更新评论统计失败", zap.Int("comment_id", commentID), zap.Error(err), tracing.WithTraceIDField(ctx))
			continue
		}

		// 更新Redis缓存
		statsKey := stats.GetCommentStatsKey(commentID)
		_ = s.statsHelper.SetStats(ctx, statsKey, map[string]int{
			"like_count":    likeCount,
			"dislike_count": dislikeCount,
		})

		// 清除脏标记
		_ = s.statsHelper.RemoveDirtyIDs(ctx, stats.CommentDirtySetKey, []int{commentID})

		syncCount++
	}

	return syncCount, nil
}

// getStatsField 根据操作类型获取对应的统计字段名
func (s *CommentStatsService) getStatsField(actionType stats.ActionType) string {
	switch actionType {
	case stats.ActionTypeLike:
		return "like_count"
	case stats.ActionTypeDislike:
		return "dislike_count"
	default:
		return ""
	}
}
