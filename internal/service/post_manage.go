package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IPostManageService Post management service interface | 帖子管理服务接口
type IPostManageService interface {
	// GetPostList Get post list | 获取帖子列表
	GetPostList(ctx context.Context, req schema.PostListRequest) (*schema.PostListResponse, error)
	// CreatePost Create a post | 创建帖子
	CreatePost(ctx context.Context, req schema.PostCreateRequest) (*ent.Post, error)
	// UpdatePost Update post information | 更新帖子信息
	UpdatePost(ctx context.Context, req schema.PostUpdateRequest) (*ent.Post, error)
	// UpdatePostStatus Update post status | 更新帖子状态
	UpdatePostStatus(ctx context.Context, req schema.PostStatusUpdateRequest) error
	// GetPostDetail Get post detail | 获取帖子详情
	GetPostDetail(ctx context.Context, id int) (*schema.PostDetailResponse, error)
	// SetPostEssence Set post as essence | 设置帖子精华
	SetPostEssence(ctx context.Context, req schema.PostEssenceUpdateRequest) error
	// SetPostPin Set post as pinned | 设置帖子置顶
	SetPostPin(ctx context.Context, req schema.PostPinUpdateRequest) error
	// MovePost Move post to another category | 移动帖子到其他版块
	MovePost(ctx context.Context, req schema.PostMoveRequest) error
	// DeletePost Delete post (soft delete, set status to Ban) | 删除帖子（软删除，状态设为Ban）
	DeletePost(ctx context.Context, id int) error
}

// PostManageService Post management service implementation | 帖子管理服务实现
type PostManageService struct {
	db           *ent.Client
	postRepo     repository.IPostRepository
	userRepo     repository.IUserRepository
	categoryRepo repository.ICategoryRepository
	cache        cache.ICacheService
	logger       *zap.Logger
}

// NewPostManageService Create a post management service instance | 创建帖子管理服务实例
func NewPostManageService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) IPostManageService {
	return &PostManageService{
		db:           db,
		postRepo:     repos.Post,
		userRepo:     repos.User,
		categoryRepo: repos.Category,
		cache:        cacheService,
		logger:       logger,
	}
}

// GetPostList Get post list | 获取帖子列表
func (s *PostManageService) GetPostList(ctx context.Context, req schema.PostListRequest) (*schema.PostListResponse, error) {
	s.logger.Info("获取帖子列表", tracing.WithTraceIDField(ctx))

	// Build query condition function | 构建查询条件函数
	conditionFunc := func(q *ent.PostQuery) *ent.PostQuery {
		// Keyword search | 关键词搜索
		if req.Keyword != "" {
			q = q.Where(
				post.Or(
					post.TitleContains(req.Keyword),
					post.ContentContains(req.Keyword),
				),
			)
		}

		// Status filter | 状态筛选
		if req.Status != "" {
			q = q.Where(post.StatusEQ(post.Status(req.Status)))
		}

		// Category filter | 版块筛选
		if req.CategoryID > 0 {
			q = q.Where(post.CategoryIDEQ(req.CategoryID))
		}

		// User filter | 用户筛选
		if req.UserID > 0 {
			q = q.Where(post.UserIDEQ(req.UserID))
		}

		// Essence post filter | 精华帖筛选
		if req.IsEssence != nil {
			q = q.Where(post.IsEssenceEQ(*req.IsEssence))
		}

		// Pinned filter | 置顶筛选
		if req.IsPinned != nil {
			q = q.Where(post.IsPinnedEQ(*req.IsPinned))
		}
		return q
	}

	// Get total count | 获取总数
	total, err := s.postRepo.CountWithCondition(ctx, conditionFunc)
	if err != nil {
		s.logger.Error("获取帖子总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子总数失败: %w", err)
	}

	// Paginated query, pinned posts first, then sorted by creation time descending | 分页查询，置顶帖在前，然后按创建时间倒序
	posts, err := s.postRepo.ListWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		q = conditionFunc(q)
		return q.Order(ent.Desc(post.FieldIsPinned), ent.Desc(post.FieldCreatedAt)).
			Offset((req.Page - 1) * req.PageSize)
	}, req.PageSize)
	if err != nil {
		s.logger.Error("获取帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子列表失败: %w", err)
	}

	// Collect user IDs and category IDs | 收集用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
	}

	// Batch query user information | 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	userMap := make(map[int]string)
	avatarMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
		avatarMap[u.ID] = u.Avatar
	}

	// Batch query category information | 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.logger.Warn("批量查询版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// Convert to response format | 转换为响应格式
	list := make([]schema.PostListItem, len(posts))
	for i, p := range posts {
		// Truncate content to first 100 characters | 截取内容前100字符
		content := p.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		// Get related information | 获取关联信息
		username := userMap[p.UserID]
		avatar := avatarMap[p.UserID]
		categoryName := categoryMap[p.CategoryID]

		list[i] = schema.PostListItem{
			ID:            p.ID,
			UserID:        p.UserID,
			Username:      username,
			Avatar:        avatar,
			CategoryID:    p.CategoryID,
			CategoryName:  categoryName,
			Title:         p.Title,
			Content:       content,
			ViewCount:     p.ViewCount,
			LikeCount:     p.LikeCount,
			DislikeCount:  p.DislikeCount,
			FavoriteCount: p.FavoriteCount,
			IsEssence:     p.IsEssence,
			IsPinned:      p.IsPinned,
			Status:        p.Status.String(),
			PublishIP:     p.PublishIP,
			CreatedAt:     p.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:     p.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.PostListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreatePost Create a post | 创建帖子
func (s *PostManageService) CreatePost(ctx context.Context, req schema.PostCreateRequest) (*ent.Post, error) {
	s.logger.Info("创建帖子", zap.String("title", req.Title), zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// Check if user exists | 检查用户是否存在
	userExists, err := s.userRepo.ExistsByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if !userExists {
		return nil, errors.New("用户不存在")
	}

	// Check if category exists | 检查版块是否存在
	categoryExists, err := s.categoryRepo.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查版块失败: %w", err)
	}
	if !categoryExists {
		return nil, errors.New("版块不存在")
	}

	// Create post | 创建帖子
	p, err := s.postRepo.Create(ctx, req.UserID, req.CategoryID, req.Title, req.Content, req.ReadPermission, post.Status(req.Status))
	if err != nil {
		s.logger.Error("创建帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Update PublishIP if needed | 如果需要更新发布IP
	if req.PublishIP != "" {
		p, err = s.postRepo.Update(ctx, p.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
			return u.SetPublishIP(req.PublishIP)
		})
		if err != nil {
			s.logger.Warn("更新发布IP失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		}
	}

	s.logger.Info("帖子创建成功", zap.Int("id", p.ID), tracing.WithTraceIDField(ctx))
	return p, nil
}

// UpdatePost Update post information | 更新帖子信息
func (s *PostManageService) UpdatePost(ctx context.Context, req schema.PostUpdateRequest) (*ent.Post, error) {
	s.logger.Info("更新帖子信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// Update post | 更新帖子
	updatedPost, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		if req.Title != "" {
			u = u.SetTitle(req.Title)
		}
		if req.Content != "" {
			u = u.SetContent(req.Content)
		}
		if req.ReadPermission != "" {
			u = u.SetReadPermission(req.ReadPermission)
		}
		if req.Status != "" {
			u = u.SetStatus(post.Status(req.Status))
		}
		return u
	})
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	s.logger.Info("帖子更新成功", zap.Int("id", updatedPost.ID), tracing.WithTraceIDField(ctx))
	return updatedPost, nil
}

// UpdatePostStatus Update post status | 更新帖子状态
func (s *PostManageService) UpdatePostStatus(ctx context.Context, req schema.PostStatusUpdateRequest) error {
	s.logger.Info("更新帖子状态", zap.Int("id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// Update status | 更新状态
	err := s.postRepo.UpdateStatus(ctx, req.ID, post.Status(req.Status))
	if err != nil {
		s.logger.Error("更新帖子状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetPostDetail Get post detail | 获取帖子详情
func (s *PostManageService) GetPostDetail(ctx context.Context, id int) (*schema.PostDetailResponse, error) {
	s.logger.Info("获取帖子详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Get post information | 获取帖子信息
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Query author information | 查询作者信息
	username := ""
	users, err := s.userRepo.GetByIDsWithFields(ctx, []int{p.UserID}, []string{user.FieldID, user.FieldUsername})
	if err == nil && len(users) > 0 {
		username = users[0].Username
	}

	// Query category information | 查询版块信息
	categoryName := ""
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, []int{p.CategoryID}, []string{category.FieldID, category.FieldName})
	if err == nil && len(categories) > 0 {
		categoryName = categories[0].Name
	}

	// Convert to response format | 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             p.ID,
		UserID:         p.UserID,
		Username:       username,
		CategoryID:     p.CategoryID,
		CategoryName:   categoryName,
		Title:          p.Title,
		Content:        p.Content,
		ReadPermission: p.ReadPermission,
		ViewCount:      p.ViewCount,
		LikeCount:      p.LikeCount,
		DislikeCount:   p.DislikeCount,
		FavoriteCount:  p.FavoriteCount,
		IsEssence:      p.IsEssence,
		IsPinned:       p.IsPinned,
		Status:         p.Status.String(),
		PublishIP:      p.PublishIP,
		CreatedAt:      p.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      p.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	return result, nil
}

// SetPostEssence Set post as essence | 设置帖子精华
func (s *PostManageService) SetPostEssence(ctx context.Context, req schema.PostEssenceUpdateRequest) error {
	s.logger.Info("设置帖子精华", zap.Int("id", req.ID), zap.Bool("is_essence", req.IsEssence), tracing.WithTraceIDField(ctx))

	// Set essence status | 设置精华状态
	_, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetIsEssence(req.IsEssence)
	})
	if err != nil {
		s.logger.Error("设置帖子精华失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子精华设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetPostPin Set post as pinned | 设置帖子置顶
func (s *PostManageService) SetPostPin(ctx context.Context, req schema.PostPinUpdateRequest) error {
	s.logger.Info("设置帖子置顶", zap.Int("id", req.ID), zap.Bool("is_pinned", req.IsPinned), tracing.WithTraceIDField(ctx))

	// Set pinned status | 设置置顶状态
	_, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetIsPinned(req.IsPinned)
	})
	if err != nil {
		s.logger.Error("设置帖子置顶失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子置顶设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// MovePost Move post to another category | 移动帖子到其他版块
func (s *PostManageService) MovePost(ctx context.Context, req schema.PostMoveRequest) error {
	s.logger.Info("移动帖子", zap.Int("id", req.ID), zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postExists, err := s.postRepo.ExistsByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !postExists {
		return errors.New("帖子不存在")
	}

	// Check if target category exists | 检查目标版块是否存在
	categoryExists, err := s.categoryRepo.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("检查目标版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查目标版块失败: %w", err)
	}
	if !categoryExists {
		return errors.New("目标版块不存在")
	}

	// Move post | 移动帖子
	_, err = s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetCategoryID(req.CategoryID)
	})
	if err != nil {
		s.logger.Error("移动帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子移动成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// DeletePost Delete post (soft delete, set status to Ban) | 删除帖子（软删除，状态设为Ban）
func (s *PostManageService) DeletePost(ctx context.Context, id int) error {
	s.logger.Info("删除帖子", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Soft delete: set status to Ban | 软删除：将状态设为Ban
	err := s.postRepo.UpdateStatus(ctx, id, post.StatusBan)
	if err != nil {
		s.logger.Error("删除帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
