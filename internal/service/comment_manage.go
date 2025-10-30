package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// ICommentManageService 评论管理服务接口
type ICommentManageService interface {
	// GetCommentList 获取评论列表
	GetCommentList(ctx context.Context, req schema.CommentListRequest) (*schema.CommentListResponse, error)
	// CreateComment 创建评论
	CreateComment(ctx context.Context, req schema.CommentCreateRequest) (*ent.Comment, error)
	// UpdateComment 更新评论信息
	UpdateComment(ctx context.Context, req schema.CommentUpdateRequest) (*ent.Comment, error)
	// GetCommentDetail 获取评论详情
	GetCommentDetail(ctx context.Context, id int) (*schema.CommentDetailResponse, error)
	// SetCommentSelected 设置评论精选
	SetCommentSelected(ctx context.Context, req schema.CommentSelectedUpdateRequest) error
	// SetCommentPin 设置评论置顶
	SetCommentPin(ctx context.Context, req schema.CommentPinUpdateRequest) error
	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, id int) error
}

// CommentManageService 评论管理服务实现
type CommentManageService struct {
	db     *ent.Client
	cache  *redis.Pool
	logger *zap.Logger
}

// NewCommentManageService 创建评论管理服务实例
func NewCommentManageService(db *ent.Client, cache *redis.Pool, logger *zap.Logger) ICommentManageService {
	return &CommentManageService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// GetCommentList 获取评论列表
func (s *CommentManageService) GetCommentList(ctx context.Context, req schema.CommentListRequest) (*schema.CommentListResponse, error) {
	s.logger.Info("获取评论列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Comment.Query().
		WithPost().
		WithAuthor().
		WithParent().
		WithReplyToUser()

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where(comment.ContentContains(req.Keyword))
	}

	// 帖子筛选
	if req.PostID > 0 {
		query = query.Where(comment.PostIDEQ(req.PostID))
	}

	// 用户筛选
	if req.UserID > 0 {
		query = query.Where(comment.UserIDEQ(req.UserID))
	}

	// 父评论筛选
	if req.ParentID != nil {
		query = query.Where(comment.ParentIDEQ(*req.ParentID))
	}

	// 精选评论筛选
	if req.IsSelected != nil {
		query = query.Where(comment.IsSelectedEQ(*req.IsSelected))
	}

	// 置顶评论筛选
	if req.IsPinned != nil {
		query = query.Where(comment.IsPinnedEQ(*req.IsPinned))
	}

	// 回复目标用户筛选
	if req.ReplyToID > 0 {
		query = query.Where(comment.ReplyToUserIDEQ(req.ReplyToID))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// 分页查询，置顶评论在前，然后按创建时间倒序
	comments, err := query.
		Order(ent.Desc(comment.FieldIsPinned), ent.Desc(comment.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取评论列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论列表失败: %w", err)
	}

	// 转换为响应格式
	list := make([]schema.CommentListItem, len(comments))
	for i, c := range comments {
		// 获取关联信息
		postTitle := ""
		if c.Edges.Post != nil {
			postTitle = c.Edges.Post.Title
		}

		username := ""
		if c.Edges.Author != nil {
			username = c.Edges.Author.Username
		}

		var replyToUsername string
		if c.Edges.ReplyToUser != nil {
			replyToUsername = c.Edges.ReplyToUser.Username
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
			CreatedAt:       c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.CommentListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateComment 创建评论
func (s *CommentManageService) CreateComment(ctx context.Context, req schema.CommentCreateRequest) (*ent.Comment, error) {
	s.logger.Info("创建评论", zap.String("content", req.Content), zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
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

	// 检查帖子是否存在
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

	// 检查父评论是否存在（如果提供了）
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

	// 检查回复目标用户是否存在（如果提供了）
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

	// 创建评论
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

// UpdateComment 更新评论信息
func (s *CommentManageService) UpdateComment(ctx context.Context, req schema.CommentUpdateRequest) (*ent.Comment, error) {
	s.logger.Info("更新评论信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查评论是否存在
	_, err := s.db.Comment.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// 更新评论内容
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

// GetCommentDetail 获取评论详情
func (s *CommentManageService) GetCommentDetail(ctx context.Context, id int) (*schema.CommentDetailResponse, error) {
	s.logger.Info("获取评论详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 获取评论信息
	c, err := s.db.Comment.Query().
		WithPost().
		WithAuthor().
		WithParent().
		WithReplyToUser().
		Where(comment.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取关联信息
	postTitle := ""
	if c.Edges.Post != nil {
		postTitle = c.Edges.Post.Title
	}

	username := ""
	if c.Edges.Author != nil {
		username = c.Edges.Author.Username
	}

	var replyToUsername string
	if c.Edges.ReplyToUser != nil {
		replyToUsername = c.Edges.ReplyToUser.Username
	}

	// 转换为响应格式
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
		CreatedAt:       c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return result, nil
}

// SetCommentSelected 设置评论精选
func (s *CommentManageService) SetCommentSelected(ctx context.Context, req schema.CommentSelectedUpdateRequest) error {
	s.logger.Info("设置评论精选", zap.Int("id", req.ID), zap.Bool("is_selected", req.IsSelected), tracing.WithTraceIDField(ctx))

	// 检查评论是否存在
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

	// 设置精选状态
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

// SetCommentPin 设置评论置顶
func (s *CommentManageService) SetCommentPin(ctx context.Context, req schema.CommentPinUpdateRequest) error {
	s.logger.Info("设置评论置顶", zap.Int("id", req.ID), zap.Bool("is_pinned", req.IsPinned), tracing.WithTraceIDField(ctx))

	// 检查评论是否存在
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

	// 设置置顶状态
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

// DeleteComment 删除评论
func (s *CommentManageService) DeleteComment(ctx context.Context, id int) error {
	s.logger.Info("删除评论", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 检查评论是否存在
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

	// 删除评论（物理删除，因为评论没有状态字段）
	err = s.db.Comment.DeleteOneID(id).Exec(ctx)
	if err != nil {
		s.logger.Error("删除评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除评论失败: %w", err)
	}

	s.logger.Info("评论删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
