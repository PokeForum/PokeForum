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
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// ICommentManageService Comment management service interface | 评论管理服务接口
type ICommentManageService interface {
	// GetCommentList Get comment list | 获取评论列表
	GetCommentList(ctx context.Context, req schema.CommentListRequest) (*schema.CommentListResponse, error)
	// CreateComment Create a comment | 创建评论
	CreateComment(ctx context.Context, req schema.CommentCreateRequest) (*ent.Comment, error)
	// UpdateComment Update comment information | 更新评论信息
	UpdateComment(ctx context.Context, req schema.CommentUpdateRequest) (*ent.Comment, error)
	// GetCommentDetail Get comment details | 获取评论详情
	GetCommentDetail(ctx context.Context, id int) (*schema.CommentDetailResponse, error)
	// SetCommentSelected Set comment as selected | 设置评论精选
	SetCommentSelected(ctx context.Context, req schema.CommentSelectedUpdateRequest) error
	// SetCommentPin Set comment pin status | 设置评论置顶
	SetCommentPin(ctx context.Context, req schema.CommentPinUpdateRequest) error
	// DeleteComment Delete a comment | 删除评论
	DeleteComment(ctx context.Context, id int) error
}

// CommentManageService Comment management service implementation | 评论管理服务实现
type CommentManageService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewCommentManageService Create a comment management service instance | 创建评论管理服务实例
func NewCommentManageService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ICommentManageService {
	return &CommentManageService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetCommentList Get comment list | 获取评论列表
func (s *CommentManageService) GetCommentList(ctx context.Context, req schema.CommentListRequest) (*schema.CommentListResponse, error) {
	s.logger.Info("获取评论列表", tracing.WithTraceIDField(ctx))

	// Build query conditions | 构建查询条件
	query := s.db.Comment.Query()

	// Keyword search | 关键词搜索
	if req.Keyword != "" {
		query = query.Where(comment.ContentContains(req.Keyword))
	}

	// Filter by post | 帖子筛选
	if req.PostID > 0 {
		query = query.Where(comment.PostIDEQ(req.PostID))
	}

	// Filter by user | 用户筛选
	if req.UserID > 0 {
		query = query.Where(comment.UserIDEQ(req.UserID))
	}

	// Filter by parent comment | 父评论筛选
	if req.ParentID != nil {
		query = query.Where(comment.ParentIDEQ(*req.ParentID))
	}

	// Filter by selected comments | 精选评论筛选
	if req.IsSelected != nil {
		query = query.Where(comment.IsSelectedEQ(*req.IsSelected))
	}

	// Filter by pinned comments | 置顶评论筛选
	if req.IsPinned != nil {
		query = query.Where(comment.IsPinnedEQ(*req.IsPinned))
	}

	// Filter by reply target user | 回复目标用户筛选
	if req.ReplyToID > 0 {
		query = query.Where(comment.ReplyToUserIDEQ(req.ReplyToID))
	}

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// Paginated query, pinned comments first, then sorted by creation time descending | 分页查询，置顶评论在前，然后按创建时间倒序
	comments, err := query.
		Order(ent.Desc(comment.FieldIsPinned), ent.Desc(comment.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取评论列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论列表失败: %w", err)
	}

	// Collect IDs that need to be queried | 收集需要查询的ID
	userIDs := make(map[int]bool)
	postIDs := make(map[int]bool)
	for _, c := range comments {
		userIDs[c.UserID] = true
		postIDs[c.PostID] = true
		if c.ReplyToUserID != 0 {
			userIDs[c.ReplyToUserID] = true
		}
	}

	// Batch query user information | 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, _ := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername).
		All(ctx)
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// Batch query post information | 批量查询帖子信息
	postIDList := make([]int, 0, len(postIDs))
	for id := range postIDs {
		postIDList = append(postIDList, id)
	}
	posts, _ := s.db.Post.Query().
		Where(post.IDIn(postIDList...)).
		Select(post.FieldID, post.FieldTitle).
		All(ctx)
	postMap := make(map[int]string)
	for _, p := range posts {
		postMap[p.ID] = p.Title
	}

	// Convert to response format | 转换为响应格式
	list := make([]schema.CommentListItem, len(comments))
	for i, c := range comments {
		// Get associated information | 获取关联信息
		postTitle := postMap[c.PostID]
		username := userMap[c.UserID]
		var replyToUsername string
		if c.ReplyToUserID != 0 {
			replyToUsername = userMap[c.ReplyToUserID]
		}

		list[i] = schema.CommentListItem{
			ID:              c.ID,
			PostID:          c.PostID,
			PostTitle:       postTitle,
			UserID:          c.UserID,
			Username:        username,
			ParentID:        &c.ParentID,
			ReplyToUserID:   &c.ReplyToUserID,
			ReplyToUsername: replyToUsername,
			Content:         c.Content,
			LikeCount:       c.LikeCount,
			DislikeCount:    c.DislikeCount,
			IsSelected:      c.IsSelected,
			IsPinned:        c.IsPinned,
			CommenterIP:     c.CommenterIP,
			DeviceInfo:      c.DeviceInfo,
			CreatedAt:       c.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:       c.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.CommentListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateComment Create a comment | 创建评论
func (s *CommentManageService) CreateComment(ctx context.Context, req schema.CommentCreateRequest) (*ent.Comment, error) {
	s.logger.Info("创建评论", zap.String("content", req.Content), zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// Check if user exists | 检查用户是否存在
	userExists, err := s.db.User.Query().
		Where(user.IDEQ(req.UserID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if !userExists {
		return nil, errors.New("用户不存在")
	}

	// Check if post exists | 检查帖子是否存在
	postExists, err := s.db.Post.Query().
		Where(post.IDEQ(req.PostID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查帖子失败: %w", err)
	}
	if !postExists {
		return nil, errors.New("帖子不存在")
	}

	// Check if parent comment exists (if provided) | 检查父评论是否存在（如果提供了）
	if req.ParentID != nil {
		parentExists, err := s.db.Comment.Query().
			Where(comment.IDEQ(*req.ParentID)).
			Exist(ctx)
		if err != nil {
			s.logger.Error("检查父评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查父评论失败: %w", err)
		}
		if !parentExists {
			return nil, errors.New("父评论不存在")
		}
	}

	// Check if reply target user exists (if provided) | 检查回复目标用户是否存在（如果提供了）
	if req.ReplyToUserID != nil {
		replyUserExists, err := s.db.User.Query().
			Where(user.IDEQ(*req.ReplyToUserID)).
			Exist(ctx)
		if err != nil {
			s.logger.Error("检查回复目标用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查回复目标用户失败: %w", err)
		}
		if !replyUserExists {
			return nil, errors.New("回复目标用户不存在")
		}
	}

	// Create comment | 创建评论
	cmt, err := s.db.Comment.Create().
		SetPostID(req.PostID).
		SetUserID(req.UserID).
		SetNillableParentID(req.ParentID).
		SetNillableReplyToUserID(req.ReplyToUserID).
		SetContent(req.Content).
		SetCommenterIP(req.CommenterIP).
		SetDeviceInfo(req.DeviceInfo).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	s.logger.Info("评论创建成功", zap.Int("id", cmt.ID), tracing.WithTraceIDField(ctx))
	return cmt, nil
}

// UpdateComment Update comment information | 更新评论信息
func (s *CommentManageService) UpdateComment(ctx context.Context, req schema.CommentUpdateRequest) (*ent.Comment, error) {
	s.logger.Info("更新评论信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// Check if comment exists | 检查评论是否存在
	_, err := s.db.Comment.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// Update comment content | 更新评论内容
	updatedComment, err := s.db.Comment.UpdateOneID(req.ID).
		SetContent(req.Content).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}

	s.logger.Info("评论更新成功", zap.Int("id", updatedComment.ID), tracing.WithTraceIDField(ctx))
	return updatedComment, nil
}

// GetCommentDetail Get comment details | 获取评论详情
func (s *CommentManageService) GetCommentDetail(ctx context.Context, id int) (*schema.CommentDetailResponse, error) {
	s.logger.Info("获取评论详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Get comment information | 获取评论信息
	c, err := s.db.Comment.Query().
		Where(comment.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// Query post information | 查询帖子信息
	postTitle := ""
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(c.PostID)).
		Select(post.FieldTitle).
		Only(ctx)
	if err == nil {
		postTitle = postData.Title
	}

	// Query author information | 查询作者信息
	username := ""
	author, err := s.db.User.Query().
		Where(user.IDEQ(c.UserID)).
		Select(user.FieldUsername).
		Only(ctx)
	if err == nil {
		username = author.Username
	}

	// Query reply target user information | 查询回复目标用户信息
	var replyToUsername string
	if c.ReplyToUserID != 0 {
		replyToUser, err := s.db.User.Query().
			Where(user.IDEQ(c.ReplyToUserID)).
			Select(user.FieldUsername).
			Only(ctx)
		if err == nil {
			replyToUsername = replyToUser.Username
		}
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CommentDetailResponse{
		ID:              c.ID,
		PostID:          c.PostID,
		PostTitle:       postTitle,
		UserID:          c.UserID,
		Username:        username,
		ParentID:        &c.ParentID,
		ReplyToUserID:   &c.ReplyToUserID,
		ReplyToUsername: replyToUsername,
		Content:         c.Content,
		LikeCount:       c.LikeCount,
		DislikeCount:    c.DislikeCount,
		IsSelected:      c.IsSelected,
		IsPinned:        c.IsPinned,
		CommenterIP:     c.CommenterIP,
		DeviceInfo:      c.DeviceInfo,
		CreatedAt:       c.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:       c.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	return result, nil
}

// SetCommentSelected Set comment as selected | 设置评论精选
func (s *CommentManageService) SetCommentSelected(ctx context.Context, req schema.CommentSelectedUpdateRequest) error {
	s.logger.Info("设置评论精选", zap.Int("id", req.ID), zap.Bool("is_selected", req.IsSelected), tracing.WithTraceIDField(ctx))

	// Check if comment exists | 检查评论是否存在
	exists, err := s.db.Comment.Query().
		Where(comment.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查评论失败: %w", err)
	}
	if !exists {
		return errors.New("评论不存在")
	}

	// Set selected status | 设置精选状态
	_, err = s.db.Comment.UpdateOneID(req.ID).
		SetIsSelected(req.IsSelected).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置评论精选失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("设置评论精选失败: %w", err)
	}

	s.logger.Info("评论精选设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetCommentPin Set comment pin status | 设置评论置顶
func (s *CommentManageService) SetCommentPin(ctx context.Context, req schema.CommentPinUpdateRequest) error {
	s.logger.Info("设置评论置顶", zap.Int("id", req.ID), zap.Bool("is_pinned", req.IsPinned), tracing.WithTraceIDField(ctx))

	// Check if comment exists | 检查评论是否存在
	exists, err := s.db.Comment.Query().
		Where(comment.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查评论失败: %w", err)
	}
	if !exists {
		return errors.New("评论不存在")
	}

	// Set pin status | 设置置顶状态
	_, err = s.db.Comment.UpdateOneID(req.ID).
		SetIsPinned(req.IsPinned).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置评论置顶失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("设置评论置顶失败: %w", err)
	}

	s.logger.Info("评论置顶设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// DeleteComment Delete a comment | 删除评论
func (s *CommentManageService) DeleteComment(ctx context.Context, id int) error {
	s.logger.Info("删除评论", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Check if comment exists | 检查评论是否存在
	exists, err := s.db.Comment.Query().
		Where(comment.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查评论失败: %w", err)
	}
	if !exists {
		return errors.New("评论不存在")
	}

	// Delete comment (physical deletion, because comments don't have a status field) | 删除评论（物理删除，因为评论没有状态字段）
	err = s.db.Comment.DeleteOneID(id).Exec(ctx)
	if err != nil {
		s.logger.Error("删除评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除评论失败: %w", err)
	}

	s.logger.Info("评论删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
