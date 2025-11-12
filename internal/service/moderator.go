package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/categorymoderator"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IModeratorService 版主服务接口
type IModeratorService interface {
	// GetModeratorCategories 获取版主管理的版块列表
	GetModeratorCategories(ctx context.Context, userID int) (*schema.ModeratorCategoriesResponse, error)
	// BanPost 封禁帖子
	BanPost(ctx context.Context, userID int, req schema.PostBanRequest) error
	// EditPost 编辑帖子
	EditPost(ctx context.Context, userID int, req schema.PostEditRequest) (*schema.ModeratorPostResponse, error)
	// MovePost 移动帖子（仅限有权限的版块）
	MovePost(ctx context.Context, userID int, req schema.PostMoveRequest) error
	// SetPostEssence 设置帖子精华
	SetPostEssence(ctx context.Context, userID int, req schema.PostEssenceRequest) error
	// LockPost 锁定帖子
	LockPost(ctx context.Context, userID int, req schema.PostLockRequest) error
	// PinPost 置顶帖子
	PinPost(ctx context.Context, userID int, req schema.PostPinRequest) error
	// EditCategory 编辑版块
	EditCategory(ctx context.Context, userID int, req schema.CategoryEditRequest) error
	// CreateCategoryAnnouncement 创建版块公告
	CreateCategoryAnnouncement(ctx context.Context, userID int, req schema.CategoryAnnouncementRequest) (*schema.CategoryAnnouncementResponse, error)
	// GetCategoryAnnouncements 获取版块公告列表
	GetCategoryAnnouncements(ctx context.Context, userID int, categoryID int) ([]schema.CategoryAnnouncementResponse, error)
}

// ModeratorService 版主服务实现
type ModeratorService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewModeratorService 创建版主服务实例
func NewModeratorService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IModeratorService {
	return &ModeratorService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetModeratorCategories 获取版主管理的版块列表
func (s *ModeratorService) GetModeratorCategories(ctx context.Context, userID int) (*schema.ModeratorCategoriesResponse, error) {
	s.logger.Info("获取版主管理的版块列表", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 通过中间表查询版主管理的版块
	moderatorRecords, err := s.db.CategoryModerator.Query().
		Where(categorymoderator.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		s.logger.Error("查询版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询版主关联记录失败: %w", err)
	}

	// 收集版块ID
	categoryIDs := make([]int, len(moderatorRecords))
	for i, record := range moderatorRecords {
		categoryIDs[i] = record.CategoryID
	}

	// 批量查询版块信息
	categories, err := s.db.Category.Query().
		Where(category.IDIn(categoryIDs...)).
		All(ctx)
	if err != nil {
		s.logger.Error("获取版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块信息失败: %w", err)
	}

	// 转换为响应格式
	result := make([]schema.ModeratorCategory, len(categories))
	for i, cat := range categories {
		// 获取该版块的帖子数量
		postCount, err := s.db.Post.Query().
			Where(post.CategoryIDEQ(cat.ID)).
			Count(ctx)
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
			CreatedAt:   cat.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.ModeratorCategoriesResponse{
		Categories: result,
	}, nil
}

// BanPost 封禁帖子
func (s *ModeratorService) BanPost(ctx context.Context, userID int, req schema.PostBanRequest) error {
	s.logger.Info("封禁帖子", zap.Int("post_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// 封禁帖子
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetStatus(post.StatusBan).
		Save(ctx)
	if err != nil {
		s.logger.Error("封禁帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("封禁帖子失败: %w", err)
	}

	s.logger.Info("帖子封禁成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// EditPost 编辑帖子
func (s *ModeratorService) EditPost(ctx context.Context, userID int, req schema.PostEditRequest) (*schema.ModeratorPostResponse, error) {
	s.logger.Info("编辑帖子", zap.Int("post_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// 更新帖子
	updatedPost, err := s.db.Post.UpdateOneID(req.ID).
		SetTitle(req.Title).
		SetContent(req.Content).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}

	// 获取评论数（简化处理）
	commentCount := 0

	// 查询作者信息
	username := ""
	author, err := s.db.User.Query().
		Where(user.IDEQ(postData.UserID)).
		Select(user.FieldUsername).
		Only(ctx)
	if err == nil {
		username = author.Username
	}

	// 查询版块信息
	categoryName := ""
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(postData.CategoryID)).
		Select(category.FieldName).
		Only(ctx)
	if err == nil {
		categoryName = categoryData.Name
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
		CreatedAt:    updatedPost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    updatedPost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("帖子编辑成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// MovePost 移动帖子（仅限有权限的版块）
func (s *ModeratorService) MovePost(ctx context.Context, userID int, req schema.PostMoveRequest) error {
	s.logger.Info("移动帖子", zap.Int("post_id", req.ID), zap.Int("target_category_id", req.CategoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有原版块的管理权限
	hasSourcePermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasSourcePermission {
		return errors.New("您没有原版块的管理权限")
	}

	// 检查版主是否有目标版块的管理权限
	hasTargetPermission, err := s.checkModeratorPermission(ctx, userID, req.CategoryID)
	if err != nil {
		return err
	}
	if !hasTargetPermission {
		return errors.New("您没有目标版块的管理权限")
	}

	// 移动帖子
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetCategoryID(req.CategoryID).
		Save(ctx)
	if err != nil {
		s.logger.Error("移动帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("移动帖子失败: %w", err)
	}

	s.logger.Info("帖子移动成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetPostEssence 设置帖子精华
func (s *ModeratorService) SetPostEssence(ctx context.Context, userID int, req schema.PostEssenceRequest) error {
	s.logger.Info("设置帖子精华", zap.Int("post_id", req.ID), zap.Bool("is_essence", req.IsEssence), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// 设置精华状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetIsEssence(req.IsEssence).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置帖子精华失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("设置帖子精华失败: %w", err)
	}

	s.logger.Info("帖子精华设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// LockPost 锁定帖子
func (s *ModeratorService) LockPost(ctx context.Context, userID int, req schema.PostLockRequest) error {
	s.logger.Info("锁定帖子", zap.Int("post_id", req.ID), zap.Bool("is_lock", req.IsLock), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// 设置锁定状态
	targetStatus := post.StatusNormal
	if req.IsLock {
		targetStatus = post.StatusLocked
	}

	_, err = s.db.Post.UpdateOneID(req.ID).
		SetStatus(targetStatus).
		Save(ctx)
	if err != nil {
		s.logger.Error("锁定帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("锁定帖子失败: %w", err)
	}

	s.logger.Info("帖子锁定设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// PinPost 置顶帖子
func (s *ModeratorService) PinPost(ctx context.Context, userID int, req schema.PostPinRequest) error {
	s.logger.Info("置顶帖子", zap.Int("post_id", req.ID), zap.Bool("is_pin", req.IsPin), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, postData.CategoryID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// 设置置顶状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetIsPinned(req.IsPin).
		Save(ctx)
	if err != nil {
		s.logger.Error("置顶帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("置顶帖子失败: %w", err)
	}

	s.logger.Info("帖子置顶设置成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// EditCategory 编辑版块
func (s *ModeratorService) EditCategory(ctx context.Context, userID int, req schema.CategoryEditRequest) error {
	s.logger.Info("编辑版块", zap.Int("category_id", req.ID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, req.ID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("您没有该版块的管理权限")
	}

	// 更新版块信息
	_, err = s.db.Category.UpdateOneID(req.ID).
		SetName(req.Name).
		SetNillableDescription(&req.Description).
		SetNillableIcon(&req.Icon).
		Save(ctx)
	if err != nil {
		s.logger.Error("编辑版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("编辑版块失败: %w", err)
	}

	s.logger.Info("版块编辑成功", zap.Int("category_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// CreateCategoryAnnouncement 创建版块公告
func (s *ModeratorService) CreateCategoryAnnouncement(ctx context.Context, userID int, req schema.CategoryAnnouncementRequest) (*schema.CategoryAnnouncementResponse, error) {
	s.logger.Info("创建版块公告", zap.Int("category_id", req.CategoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// 获取版主用户名
	moderatorData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取版主信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版主信息失败: %w", err)
	}

	// 创建版块公告
	updatedCategory, err := s.db.Category.UpdateOneID(req.CategoryID).
		SetAnnouncement(req.Content).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建版块公告失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建版块公告失败: %w", err)
	}

	// 构建响应数据
	result := &schema.CategoryAnnouncementResponse{
		ID:         updatedCategory.ID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Content:    req.Content,
		IsPinned:   req.IsPinned,
		Username:   moderatorData.Username,
		CreatedAt:  updatedCategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  updatedCategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("版块公告创建成功", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetCategoryAnnouncements 获取版块公告列表
func (s *ModeratorService) GetCategoryAnnouncements(ctx context.Context, userID int, categoryID int) ([]schema.CategoryAnnouncementResponse, error) {
	s.logger.Info("获取版块公告列表", zap.Int("category_id", categoryID), zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 检查版主是否有该版块的管理权限
	hasPermission, err := s.checkModeratorPermission(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("您没有该版块的管理权限")
	}

	// 获取版块信息
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(categoryID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("版块不存在")
		}
		s.logger.Error("获取版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块信息失败: %w", err)
	}

	// 如果版块没有公告，返回空列表
	if categoryData.Announcement == "" {
		return []schema.CategoryAnnouncementResponse{}, nil
	}

	// 获取版主用户名
	moderatorData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取版主信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版主信息失败: %w", err)
	}

	// 构建响应数据
	result := []schema.CategoryAnnouncementResponse{
		{
			ID:         categoryData.ID,
			CategoryID: categoryID,
			Title:      "版块公告",
			Content:    categoryData.Announcement,
			IsPinned:   false,
			Username:   moderatorData.Username,
			CreatedAt:  categoryData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  categoryData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	s.logger.Info("获取版块公告列表成功", zap.Int("category_id", categoryID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// checkModeratorPermission 检查版主是否有指定版块的管理权限（辅助函数）
func (s *ModeratorService) checkModeratorPermission(ctx context.Context, userID, categoryID int) (bool, error) {
	// 通过中间表查询版主权限
	return s.db.CategoryModerator.Query().
		Where(
			categorymoderator.And(
				categorymoderator.CategoryIDEQ(categoryID),
				categorymoderator.UserIDEQ(userID),
			),
		).
		Exist(ctx)
}
