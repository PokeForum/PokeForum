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

// IModeratorService Moderator service interface | 版主服务接口
type IModeratorService interface {
	// GetModeratorCategories Get list of categories managed by moderator | 获取版主管理的版块列表
	GetModeratorCategories(ctx context.Context, userID int) (*schema.ModeratorCategoriesResponse, error)
	// BanPost Ban post | 封禁帖子
	BanPost(ctx context.Context, userID int, req schema.PostBanRequest) error
	// EditPost Edit post | 编辑帖子
	EditPost(ctx context.Context, userID int, req schema.PostEditRequest) (*schema.ModeratorPostResponse, error)
	// MovePost Move post (only for categories with permission) | 移动帖子（仅限有权限的版块）
	MovePost(ctx context.Context, userID int, req schema.PostMoveRequest) error
	// SetPostEssence Set post as essence | 设置帖子精华
	SetPostEssence(ctx context.Context, userID int, req schema.PostEssenceRequest) error
	// LockPost Lock post | 锁定帖子
	LockPost(ctx context.Context, userID int, req schema.PostLockRequest) error
	// PinPost Pin post | 置顶帖子
	PinPost(ctx context.Context, userID int, req schema.PostPinRequest) error
	// EditCategory Edit category | 编辑版块
	EditCategory(ctx context.Context, userID int, req schema.CategoryEditRequest) error
	// CreateCategoryAnnouncement Create category announcement | 创建版块公告
	CreateCategoryAnnouncement(ctx context.Context, userID int, req schema.CategoryAnnouncementRequest) (*schema.CategoryAnnouncementResponse, error)
	// GetCategoryAnnouncements Get category announcements list | 获取版块公告列表
	GetCategoryAnnouncements(ctx context.Context, userID int, categoryID int) ([]schema.CategoryAnnouncementResponse, error)
}

// ModeratorService Moderator service implementation | 版主服务实现
type ModeratorService struct {
	db                    *ent.Client
	postRepo              repository.IPostRepository
	userRepo              repository.IUserRepository
	categoryRepo          repository.ICategoryRepository
	categoryModeratorRepo repository.ICategoryModeratorRepository
	cache                 cache.ICacheService
	logger                *zap.Logger
}

// NewModeratorService Create moderator service instance | 创建版主服务实例
func NewModeratorService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) IModeratorService {
	return &ModeratorService{
		db:                    db,
		postRepo:              repos.Post,
		userRepo:              repos.User,
		categoryRepo:          repos.Category,
		categoryModeratorRepo: repos.CategoryModerator,
		cache:                 cacheService,
		logger:                logger,
	}
}

// GetModeratorCategories Get list of categories managed by moderator | 获取版主管理的版块列表
func (s *ModeratorService) GetModeratorCategories(ctx context.Context, userID int) (*schema.ModeratorCategoriesResponse, error) {
	s.logger.Info("获取版主管理的版块列表", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Query categories managed by moderator through junction table | 通过中间表查询版主管理的版块
	moderatorRecords, err := s.categoryModeratorRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询版主关联记录失败: %w", err)
	}

	// Collect category IDs | 收集版块ID
	categoryIDs := make([]int, len(moderatorRecords))
	for i, record := range moderatorRecords {
		categoryIDs[i] = record.CategoryID
	}

	// Batch query category information | 批量查询版块信息
	categories, err := s.categoryRepo.GetByIDs(ctx, categoryIDs)
	if err != nil {
		s.logger.Error("获取版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块信息失败: %w", err)
	}

	// Convert to response format | 转换为响应格式
	result := make([]schema.ModeratorCategory, len(categories))
	for i, cat := range categories {
		// Get post count for this category | 获取该版块的帖子数量
		postCount, err := s.postRepo.CountByCategoryID(ctx, cat.ID)
		if err != nil {
			postCount = 0
		}

		result[i] = schema.ModeratorCategory{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.Icon,
			Status:      string(cat.Status),
			PostCount:   postCount,
			CreatedAt:   cat.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.ModeratorCategoriesResponse{
		Categories: result,
	}, nil
}

// BanPost Ban post | 封禁帖子
func (s *ModeratorService) BanPost(ctx context.Context, userID int, req schema.PostBanRequest) error {
	s.logger.Info("封禁帖子", zap.Int("post_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// Ban post | 封禁帖子
	err = s.postRepo.UpdateStatus(ctx, req.ID, post.StatusBan)
	if err != nil {
		s.logger.Error("封禁帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子封禁成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// EditPost Edit post | 编辑帖子
func (s *ModeratorService) EditPost(ctx context.Context, userID int, req schema.PostEditRequest) (*schema.ModeratorPostResponse, error) {
	s.logger.Info("编辑帖子", zap.Int("post_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Check if moderator has permission to manage this category | 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// Update post | 更新帖子
	updatedPost, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetTitle(req.Title).SetContent(req.Content)
	})
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Get comment count (simplified handling) | 获取评论数（简化处理）
	commentCount := 0

	// Query author information | 查询作者信息
	username := ""
	users, err := s.userRepo.GetByIDsWithFields(ctx, []int{postData.UserID}, []string{user.FieldID, user.FieldUsername})
	if err == nil && len(users) > 0 {
		username = users[0].Username
	}

	// Query category information | 查询版块信息
	categoryName := ""
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, []int{postData.CategoryID}, []string{category.FieldID, category.FieldName})
	if err == nil && len(categories) > 0 {
		categoryName = categories[0].Name
	}

	result := &schema.ModeratorPostResponse{
		ID:           updatedPost.ID,
		Title:        updatedPost.Title,
		Content:      updatedPost.Content,
		Username:     username,
		CategoryID:   updatedPost.CategoryID,
		CategoryName: categoryName,
		Status:       string(updatedPost.Status),
		IsEssence:    updatedPost.IsEssence,
		IsPinned:     updatedPost.IsPinned,
		ViewCount:    updatedPost.ViewCount,
		LikeCount:    updatedPost.LikeCount,
		CommentCount: commentCount,
		CreatedAt:    updatedPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:    updatedPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("帖子编辑成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// MovePost Move post (only for categories with permission) | 移动帖子（仅限有权限的版块）
func (s *ModeratorService) MovePost(ctx context.Context, userID int, req schema.PostMoveRequest) error {
	s.logger.Info("移动帖子", zap.Int("post_id", req.ID), zap.Int("target_category_id", req.CategoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// Check if moderator has permission to manage source category | 检查版主是否有原版块的管理权限
	hasSourcePermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasSourcePermission {
		return errors.New("您没有原版块的管理权限")
	}

	// Check if moderator has permission to manage target category | 检查版主是否有目标版块的管理权限
	hasTargetPermission, err := s.checkModeratorPermission(ctx, userID, req.CategoryID)
	if err != nil {
		return err
	}
	if !hasTargetPermission {
		return errors.New("您没有目标版块的管理权限")
	}

	// Move post | 移动帖子
	_, err = s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetCategoryID(req.CategoryID)
	})
	if err != nil {
		s.logger.Error("移动帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子移动成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetPostEssence Set post as essence | 设置帖子精华
func (s *ModeratorService) SetPostEssence(ctx context.Context, userID int, req schema.PostEssenceRequest) error {
	s.logger.Info("设置帖子精华", zap.Int("post_id", req.ID), zap.Bool("is_essence", req.IsEssence), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// Set essence status | 设置精华状态
	_, err = s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetIsEssence(req.IsEssence)
	})
	if err != nil {
		s.logger.Error("设置帖子精华失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子精华设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// LockPost Lock post | 锁定帖子
func (s *ModeratorService) LockPost(ctx context.Context, userID int, req schema.PostLockRequest) error {
	s.logger.Info("锁定帖子", zap.Int("post_id", req.ID), zap.Bool("is_lock", req.IsLock), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// Set lock status | 设置锁定状态
	targetStatus := post.StatusNormal
	if req.IsLock {
		targetStatus = post.StatusLocked
	}

	err = s.postRepo.UpdateStatus(ctx, req.ID, targetStatus)
	if err != nil {
		s.logger.Error("锁定帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子锁定设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// PinPost Pin post | 置顶帖子
func (s *ModeratorService) PinPost(ctx context.Context, userID int, req schema.PostPinRequest) error {
	s.logger.Info("置顶帖子", zap.Int("post_id", req.ID), zap.Bool("is_pin", req.IsPin), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if post exists | 检查帖子是否存在
	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// Set pin status | 设置置顶状态
	_, err = s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetIsPinned(req.IsPin)
	})
	if err != nil {
		s.logger.Error("置顶帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("帖子置顶设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// EditCategory Edit category | 编辑版块
func (s *ModeratorService) EditCategory(ctx context.Context, userID int, req schema.CategoryEditRequest) error {
	s.logger.Info("编辑版块", zap.Int("category_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if moderator has permission to manage this category | 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, req.ID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// Update category information | 更新版块信息
	_, err = s.categoryRepo.Update(ctx, req.ID, func(u *ent.CategoryUpdateOne) *ent.CategoryUpdateOne {
		return u.SetName(req.Name).SetNillableDescription(&req.Description).SetNillableIcon(&req.Icon)
	})
	if err != nil {
		s.logger.Error("编辑版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("版块编辑成功", zap.Int("category_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// CreateCategoryAnnouncement Create category announcement | 创建版块公告
func (s *ModeratorService) CreateCategoryAnnouncement(ctx context.Context, userID int, req schema.CategoryAnnouncementRequest) (*schema.CategoryAnnouncementResponse, error) {
	s.logger.Info("创建版块公告", zap.Int("category_id", req.CategoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if moderator has permission to manage this category | 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// Get moderator username | 获取版主用户名
	moderatorData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取版主信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Create category announcement | 创建版块公告
	updatedCategory, err := s.categoryRepo.Update(ctx, req.CategoryID, func(u *ent.CategoryUpdateOne) *ent.CategoryUpdateOne {
		return u.SetAnnouncement(req.Content)
	})
	if err != nil {
		s.logger.Error("创建版块公告失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.CategoryAnnouncementResponse{
		ID:         updatedCategory.ID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Content:    req.Content,
		IsPinned:   req.IsPinned,
		Username:   moderatorData.Username,
		CreatedAt:  updatedCategory.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:  updatedCategory.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("版块公告创建成功", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetCategoryAnnouncements Get category announcements list | 获取版块公告列表
func (s *ModeratorService) GetCategoryAnnouncements(ctx context.Context, userID int, categoryID int) ([]schema.CategoryAnnouncementResponse, error) {
	s.logger.Info("获取版块公告列表", zap.Int("category_id", categoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// Check if moderator has permission to manage this category | 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// Get category information | 获取版块信息
	categoryData, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("获取版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// If category has no announcement, return empty list | 如果版块没有公告，返回空列表
	if categoryData.Announcement == "" {
		return []schema.CategoryAnnouncementResponse{}, nil
	}

	// Get moderator username | 获取版主用户名
	moderatorData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取版主信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := []schema.CategoryAnnouncementResponse{
		{
			ID:         categoryData.ID,
			CategoryID: categoryID,
			Title:      "版块公告",
			Content:    categoryData.Announcement,
			IsPinned:   false,
			Username:   moderatorData.Username,
			CreatedAt:  categoryData.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:  categoryData.UpdatedAt.Format(time_tools.DateTimeFormat),
		},
	}

	s.logger.Info("获取版块公告列表成功", zap.Int("category_id", categoryID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// checkModeratorPermission Check if moderator has permission to manage specified category (helper function) | 检查版主是否有指定版块的管理权限（辅助函数）
func (s *ModeratorService) checkModeratorPermission(ctx context.Context, userID, categoryID int) (bool, error) {
	// Query moderator permissions through junction table | 通过中间表查询版主权限
	return s.categoryModeratorRepo.Exists(ctx, categoryID, userID)
}
