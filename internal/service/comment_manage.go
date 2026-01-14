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
	"github.com/PokeForum/PokeForum/internal/repository"
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
	db          *ent.Client
	commentRepo repository.ICommentRepository
	userRepo    repository.IUserRepository
	postRepo    repository.IPostRepository
	cache       cache.ICacheService
	logger      *zap.Logger
}

// NewCommentManageService Create a comment management service instance | 创建评论管理服务实例
func NewCommentManageService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) ICommentManageService {
	return &CommentManageService{
		db:          db,
		commentRepo: repos.Comment,
		userRepo:    repos.User,
		postRepo:    repos.Post,
		cache:       cacheService,
		logger:      logger,
	}
}

// GetCommentList Get comment list | 获取评论列表
func (s *CommentManageService) GetCommentList(ctx context.Context, req schema.CommentListRequest) (*schema.CommentListResponse, error) {
	s.logger.Info("获取评论列表", tracing.WithTraceIDField(ctx))

	// Build query condition function | 构建查询条件函数
	conditionFunc := func(q *ent.CommentQuery) *ent.CommentQuery {
		// Keyword search | 关键词搜索
		if req.Keyword != "" {
			q = q.Where(comment.ContentContains(req.Keyword))
		}

		// Post filter | 帖子筛选
		if req.PostID > 0 {
			q = q.Where(comment.PostIDEQ(req.PostID))
		}

		// User filter | 用户筛选
		if req.UserID > 0 {
			q = q.Where(comment.UserIDEQ(req.UserID))
		}
		return q
	}

	// Get total count | 获取总数
	total, err := s.commentRepo.CountWithCondition(ctx, conditionFunc)
	if err != nil {
		s.logger.Error("获取评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// Paginated query | 分页查询
	comments, err := s.commentRepo.ListWithCondition(ctx, func(q *ent.CommentQuery) *ent.CommentQuery {
		q = conditionFunc(q)
		return q.Order(ent.Desc(comment.FieldCreatedAt)).
			Offset((req.Page - 1) * req.PageSize)
	}, req.PageSize)
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
	users, _ := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar})
	userMap := make(map[int]string)
	avatarMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
		avatarMap[u.ID] = u.Avatar
	}

	// Batch query post information | 批量查询帖子信息
	postIDList := make([]int, 0, len(postIDs))
	for id := range postIDs {
		postIDList = append(postIDList, id)
	}
	posts, _ := s.postRepo.GetByIDsWithFields(ctx, postIDList, []string{post.FieldID, post.FieldTitle})
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
		avatar := avatarMap[c.UserID]
		var replyToUsername string
		var replyToAvatar string
		if c.ReplyToUserID != 0 {
			replyToUsername = userMap[c.ReplyToUserID]
			replyToAvatar = avatarMap[c.ReplyToUserID]
		}

		list[i] = schema.CommentListItem{
			ID:              c.ID,
			PostID:          c.PostID,
			PostTitle:       postTitle,
			UserID:          c.UserID,
			Username:        username,
			Avatar:          avatar,
			ParentID:        &c.ParentID,
			ReplyToUserID:   &c.ReplyToUserID,
			ReplyToUsername: replyToUsername,
			ReplyToAvatar:   replyToAvatar,
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
	userExists, err := s.userRepo.ExistsByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if !userExists {
		return nil, errors.New("用户不存在")
	}

	// Check if post exists | 检查帖子是否存在
	postExists, err := s.postRepo.ExistsByID(ctx, req.PostID)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查帖子失败: %w", err)
	}
	if !postExists {
		return nil, errors.New("帖子不存在")
	}

	// Check if parent comment exists (if provided) | 检查父评论是否存在（如果提供了）
	if req.ParentID != nil {
		parentExists, err := s.commentRepo.ExistsByID(ctx, *req.ParentID)
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
		replyUserExists, err := s.userRepo.ExistsByID(ctx, *req.ReplyToUserID)
		if err != nil {
			s.logger.Error("检查回复目标用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查回复目标用户失败: %w", err)
		}
		if !replyUserExists {
			return nil, errors.New("回复目标用户不存在")
		}
	}

	// Create comment | 创建评论
	cmt, err := s.commentRepo.Create(ctx, req.UserID, req.PostID, req.Content, req.CommenterIP, req.DeviceInfo, req.ParentID, req.ReplyToUserID)
	if err != nil {
		s.logger.Error("创建评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	s.logger.Info("评论创建成功", zap.Int("id", cmt.ID), tracing.WithTraceIDField(ctx))
	return cmt, nil
}

// UpdateComment Update comment information | 更新评论信息
func (s *CommentManageService) UpdateComment(ctx context.Context, req schema.CommentUpdateRequest) (*ent.Comment, error) {
	s.logger.Info("更新评论信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// Update comment content | 更新评论内容
	updatedComment, err := s.commentRepo.UpdateContent(ctx, req.ID, req.Content)
	if err != nil {
		s.logger.Error("更新评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	s.logger.Info("评论更新成功", zap.Int("id", updatedComment.ID), tracing.WithTraceIDField(ctx))
	return updatedComment, nil
}

// GetCommentDetail Get comment details | 获取评论详情
func (s *CommentManageService) GetCommentDetail(ctx context.Context, id int) (*schema.CommentDetailResponse, error) {
	s.logger.Info("获取评论详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Get comment information | 获取评论信息
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Query post information | 查询帖子信息
	postTitle := ""
	posts, err := s.postRepo.GetByIDsWithFields(ctx, []int{c.PostID}, []string{post.FieldID, post.FieldTitle})
	if err == nil && len(posts) > 0 {
		postTitle = posts[0].Title
	}

	// Query author information | 查询作者信息
	username := ""
	users, err := s.userRepo.GetByIDsWithFields(ctx, []int{c.UserID}, []string{user.FieldID, user.FieldUsername})
	if err == nil && len(users) > 0 {
		username = users[0].Username
	}

	// Query reply target user information | 查询回复目标用户信息
	var replyToUsername string
	if c.ReplyToUserID != 0 {
		replyUsers, err := s.userRepo.GetByIDsWithFields(ctx, []int{c.ReplyToUserID}, []string{user.FieldID, user.FieldUsername})
		if err == nil && len(replyUsers) > 0 {
			replyToUsername = replyUsers[0].Username
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

	// Set selected status | 设置精选状态
	_, err := s.commentRepo.Update(ctx, req.ID, func(u *ent.CommentUpdateOne) *ent.CommentUpdateOne {
		return u.SetIsSelected(req.IsSelected)
	})
	if err != nil {
		s.logger.Error("设置评论精选失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("评论精选设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetCommentPin Set comment pin status | 设置评论置顶
func (s *CommentManageService) SetCommentPin(ctx context.Context, req schema.CommentPinUpdateRequest) error {
	s.logger.Info("设置评论置顶", zap.Int("id", req.ID), zap.Bool("is_pinned", req.IsPinned), tracing.WithTraceIDField(ctx))

	// Set pin status | 设置置顶状态
	_, err := s.commentRepo.Update(ctx, req.ID, func(u *ent.CommentUpdateOne) *ent.CommentUpdateOne {
		return u.SetIsPinned(req.IsPinned)
	})
	if err != nil {
		s.logger.Error("设置评论置顶失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("评论置顶设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// DeleteComment Delete a comment | 删除评论
func (s *CommentManageService) DeleteComment(ctx context.Context, id int) error {
	s.logger.Info("删除评论", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Delete comment (physical deletion, because comments don't have a status field) | 删除评论（物理删除，因为评论没有状态字段）
	err := s.commentRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("删除评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("评论删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
