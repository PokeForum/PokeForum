package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// ICommentService 评论服务接口
type ICommentService interface {
	// CreateComment 创建评论
	CreateComment(ctx context.Context, userID int, clientIP, deviceInfo string, req schema.UserCommentCreateRequest) (*schema.UserCommentCreateResponse, error)
	// UpdateComment 更新评论
	UpdateComment(ctx context.Context, userID int, req schema.UserCommentUpdateRequest) (*schema.UserCommentUpdateResponse, error)
	// LikeComment 点赞评论
	LikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error)
	// DislikeComment 点踩评论
	DislikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error)
	// GetCommentList 获取评论列表
	GetCommentList(ctx context.Context, req schema.UserCommentListRequest) (*schema.UserCommentListResponse, error)
	// GetCommentDetail 获取评论详情
	GetCommentDetail(ctx context.Context, commentID int) (*schema.UserCommentDetailResponse, error)
}

// CommentService 评论服务实现
type CommentService struct {
	db                  *ent.Client
	cache               cache.ICacheService
	logger              *zap.Logger
	commentStatsService ICommentStatsService
}

// NewCommentService 创建评论服务实例
func NewCommentService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ICommentService {
	return &CommentService{
		db:                  db,
		cache:               cacheService,
		logger:              logger,
		commentStatsService: NewCommentStatsService(db, cacheService, logger),
	}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(ctx context.Context, userID int, clientIP, deviceInfo string, req schema.UserCommentCreateRequest) (*schema.UserCommentCreateResponse, error) {
	s.logger.Info("创建评论", zap.Int("user_id", userID), zap.Int("post_id", req.PostID), zap.String("client_ip", clientIP), zap.String("device_info", deviceInfo), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在且状态正常
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.PostID), post.StatusEQ(post.StatusNormal)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在或已删除")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查是否被楼主拉黑
	blacklistService := NewBlacklistService(s.db, s.logger)
	isBlockedByAuthor, err := blacklistService.IsUserBlocked(ctx, postData.UserID, userID)
	if err != nil {
		s.logger.Error("检查是否被楼主拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
	}
	if isBlockedByAuthor {
		return nil, errors.New("您已被楼主拉黑，无法评论该帖子")
	}

	// 如果是回复评论，检查父评论是否存在
	var replyToUsername string
	if req.ParentID != nil {
		parentComment, err := s.db.Comment.Query().
			Where(comment.IDEQ(*req.ParentID), comment.PostIDEQ(req.PostID)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errors.New("父评论不存在或不属于该帖子")
			}
			s.logger.Error("获取父评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取父评论失败: %w", err)
		}
		// 设置回复目标用户ID
		req.ReplyToUserID = &parentComment.UserID
	}

	// 如果指定了回复目标用户，获取用户名并检查拉黑状态
	if req.ReplyToUserID != nil {
		// 检查是否被回复目标用户拉黑
		isBlockedByTarget, err := blacklistService.IsUserBlocked(ctx, *req.ReplyToUserID, userID)
		if err != nil {
			s.logger.Error("检查是否被回复目标用户拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
		}
		if isBlockedByTarget {
			return nil, errors.New("您已被该用户拉黑，无法回复其评论")
		}

		replyToUser, err := s.db.User.Query().
			Where(user.IDEQ(*req.ReplyToUserID)).
			Only(ctx)
		if err != nil {
			s.logger.Error("获取回复目标用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取回复目标用户失败: %w", err)
		}
		replyToUsername = replyToUser.Username
	}

	// 创建评论
	newComment, err := s.db.Comment.Create().
		SetUserID(userID).
		SetPostID(req.PostID).
		SetContent(req.Content).
		SetNillableParentID(req.ParentID).
		SetNillableReplyToUserID(req.ReplyToUserID).
		SetCommenterIP(clientIP).  // 从请求中获取真实IP
		SetDeviceInfo(deviceInfo). // 从请求中获取设备信息
		Save(ctx)
	if err != nil {
		s.logger.Error("创建评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserCommentCreateResponse{
		ID:              newComment.ID,
		PostID:          newComment.PostID,
		ReplyToUsername: replyToUsername,
		Content:         newComment.Content,
		LikeCount:       newComment.LikeCount,
		DislikeCount:    newComment.DislikeCount,
		CreatedAt:       newComment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// 只有当ParentID不为0时才设置
	if newComment.ParentID != 0 {
		result.ParentID = &newComment.ParentID
	}

	// 只有当ReplyToUserID不为0时才设置
	if newComment.ReplyToUserID != 0 {
		result.ReplyToUserID = &newComment.ReplyToUserID
	}

	s.logger.Info("创建评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdateComment 更新评论
func (s *CommentService) UpdateComment(ctx context.Context, userID int, req schema.UserCommentUpdateRequest) (*schema.UserCommentUpdateResponse, error) {
	s.logger.Info("更新评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查评论是否存在且属于当前用户
	commentData, err := s.db.Comment.Query().
		Where(comment.IDEQ(req.ID), comment.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在或无权限修改")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取评论所属帖子信息
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(commentData.PostID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取评论所属帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子信息失败: %w", err)
	}

	// 检查是否被楼主拉黑
	blacklistService := NewBlacklistService(s.db, s.logger)
	isBlockedByAuthor, err := blacklistService.IsUserBlocked(ctx, postData.UserID, userID)
	if err != nil {
		s.logger.Error("检查是否被楼主拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
	}
	if isBlockedByAuthor {
		return nil, errors.New("您已被楼主拉黑，无法编辑该评论")
	}

	// 更新评论
	updatedComment, err := s.db.Comment.UpdateOne(commentData).
		SetContent(req.Content).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserCommentUpdateResponse{
		ID:        updatedComment.ID,
		Content:   updatedComment.Content,
		UpdatedAt: updatedComment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("更新评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error) {
	s.logger.Info("点赞评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点赞操作
	stats, err := s.commentStatsService.PerformAction(ctx, userID, req.ID, "Like")
	if err != nil {
		s.logger.Error("点赞评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
	result := &schema.UserCommentActionResponse{
		ID:           stats.ID,
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
	}

	s.logger.Info("点赞评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// DislikeComment 点踩评论
func (s *CommentService) DislikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error) {
	s.logger.Info("点踩评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点踩操作
	stats, err := s.commentStatsService.PerformAction(ctx, userID, req.ID, "Dislike")
	if err != nil {
		s.logger.Error("点踩评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
	result := &schema.UserCommentActionResponse{
		ID:           stats.ID,
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
	}

	s.logger.Info("点踩评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetCommentList 获取评论列表
func (s *CommentService) GetCommentList(ctx context.Context, req schema.UserCommentListRequest) (*schema.UserCommentListResponse, error) {
	s.logger.Info("获取评论列表", zap.Int("post_id", req.PostID), zap.Int("page", req.Page), tracing.WithTraceIDField(ctx))

	// 设置默认排序
	if req.SortBy == "" {
		req.SortBy = "created_at"
		req.SortDesc = true
	}

	// 构建查询条件
	query := s.db.Comment.Query().
		Where(comment.PostIDEQ(req.PostID))

	// 根据排序字段和方向排序
	switch req.SortBy {
	case "created_at":
		if req.SortDesc {
			query = query.Order(ent.Desc(comment.FieldCreatedAt))
		} else {
			query = query.Order(ent.Asc(comment.FieldCreatedAt))
		}
	case "like_count":
		if req.SortDesc {
			query = query.Order(ent.Desc(comment.FieldLikeCount))
		} else {
			query = query.Order(ent.Asc(comment.FieldLikeCount))
		}
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// 分页查询
	comments, err := query.
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取评论列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论列表失败: %w", err)
	}

	// 收集所有需要查询的用户ID
	userIDs := make(map[int]bool)
	for _, c := range comments {
		userIDs[c.UserID] = true
		if c.ReplyToUserID != 0 {
			userIDs[c.ReplyToUserID] = true
		}
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername).
		All(ctx)
	if err != nil {
		s.logger.Error("查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 创建用户ID到用户名的映射
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 构建响应数据
	list := make([]schema.UserCommentListItem, len(comments))
	for i, commentData := range comments {
		username := "未知用户"
		if name, ok := userMap[commentData.UserID]; ok {
			username = name
		}

		replyToUsername := ""
		if commentData.ReplyToUserID != 0 {
			if name, ok := userMap[commentData.ReplyToUserID]; ok {
				replyToUsername = name
			}
		}

		list[i] = schema.UserCommentListItem{
			ID:           commentData.ID,
			UserID:       commentData.UserID,
			Username:     username,
			Content:      commentData.Content,
			LikeCount:    commentData.LikeCount,
			DislikeCount: commentData.DislikeCount,
			IsSelected:   commentData.IsSelected,
			IsPinned:     commentData.IsPinned,
			CreatedAt:    commentData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    commentData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// 只有当ParentID不为0时才设置
		if commentData.ParentID != 0 {
			list[i].ParentID = &commentData.ParentID
		}

		// 只有当ReplyToUserID不为0时才设置
		if commentData.ReplyToUserID != 0 {
			list[i].ReplyToUserID = &commentData.ReplyToUserID
			list[i].ReplyToUsername = replyToUsername
		}
	}

	result := &schema.UserCommentListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	s.logger.Info("获取评论列表成功", zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetCommentDetail 获取评论详情
func (s *CommentService) GetCommentDetail(ctx context.Context, commentID int) (*schema.UserCommentDetailResponse, error) {
	s.logger.Info("获取评论详情", zap.Int("comment_id", commentID), tracing.WithTraceIDField(ctx))

	// 查询评论详情
	commentData, err := s.db.Comment.Query().
		Where(comment.IDEQ(commentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("获取评论详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论详情失败: %w", err)
	}

	// 查询作者信息
	username := "未知用户"
	author, err := s.db.User.Query().
		Where(user.IDEQ(commentData.UserID)).
		Select(user.FieldUsername).
		Only(ctx)
	if err == nil {
		username = author.Username
	}

	// 查询回复目标用户名
	replyToUsername := ""
	if commentData.ReplyToUserID != 0 {
		replyToUser, err := s.db.User.Query().
			Where(user.IDEQ(commentData.ReplyToUserID)).
			Select(user.FieldUsername).
			Only(ctx)
		if err == nil {
			replyToUsername = replyToUser.Username
		}
	}

	// 构建响应数据
	result := &schema.UserCommentDetailResponse{
		ID:           commentData.ID,
		PostID:       commentData.PostID,
		UserID:       commentData.UserID,
		Username:     username,
		Content:      commentData.Content,
		LikeCount:    commentData.LikeCount,
		DislikeCount: commentData.DislikeCount,
		IsSelected:   commentData.IsSelected,
		IsPinned:     commentData.IsPinned,
		CreatedAt:    commentData.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    commentData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// 只有当ParentID不为0时才设置
	if commentData.ParentID != 0 {
		result.ParentID = &commentData.ParentID
	}

	// 只有当ReplyToUserID不为0时才设置
	if commentData.ReplyToUserID != 0 {
		result.ReplyToUserID = &commentData.ReplyToUserID
		result.ReplyToUsername = replyToUsername
	}

	s.logger.Info("获取评论详情成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}
