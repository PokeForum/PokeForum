package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// ICategoryManageService Category management service interface | 版块管理服务接口
type ICategoryManageService interface {
	// GetCategoryList Get category list | 获取版块列表
	GetCategoryList(ctx context.Context, req schema.CategoryListRequest) (*schema.CategoryListResponse, error)
	// CreateCategory Create category | 创建版块
	CreateCategory(ctx context.Context, req schema.CategoryCreateRequest) (*ent.Category, error)
	// UpdateCategory Update category information | 更新版块信息
	UpdateCategory(ctx context.Context, req schema.CategoryUpdateRequest) (*ent.Category, error)
	// UpdateCategoryStatus Update category status | 更新版块状态
	UpdateCategoryStatus(ctx context.Context, req schema.CategoryStatusUpdateRequest) error
	// GetCategoryDetail Get category details | 获取版块详情
	GetCategoryDetail(ctx context.Context, id int) (*schema.CategoryDetailResponse, error)
	// SetCategoryModerators Set category moderators | 设置版块版主
	SetCategoryModerators(ctx context.Context, req schema.CategoryModeratorRequest) error
	// DeleteCategory Hide category (soft delete, set status to Hidden) | 隐藏版块（软删除，状态设为Hidden）
	DeleteCategory(ctx context.Context, id int) error
}

// CategoryManageService Category management service implementation | 版块管理服务实现
type CategoryManageService struct {
	db            *ent.Client
	categoryRepo  repository.ICategoryRepository
	moderatorRepo repository.ICategoryModeratorRepository
	userRepo      repository.IUserRepository
	cache         cache.ICacheService
	logger        *zap.Logger
}

// NewCategoryManageService Create category management service instance | 创建版块管理服务实例
func NewCategoryManageService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) ICategoryManageService {
	return &CategoryManageService{
		db:            db,
		categoryRepo:  repos.Category,
		moderatorRepo: repos.CategoryModerator,
		userRepo:      repos.User,
		cache:         cacheService,
		logger:        logger,
	}
}

// GetCategoryList Get category list | 获取版块列表
func (s *CategoryManageService) GetCategoryList(ctx context.Context, req schema.CategoryListRequest) (*schema.CategoryListResponse, error) {
	s.logger.Info("获取版块列表", tracing.WithTraceIDField(ctx))

	// Build query condition function | 构建查询条件函数
	conditionFunc := func(q *ent.CategoryQuery) *ent.CategoryQuery {
		// Keyword search | 关键词搜索
		if req.Keyword != "" {
			q = q.Where(
				category.Or(
					category.NameContains(req.Keyword),
					category.DescriptionContains(req.Keyword),
				),
			)
		}

		// Status filter | 状态筛选
		if req.Status != "" {
			q = q.Where(category.StatusEQ(category.Status(req.Status)))
		}
		return q
	}

	// Get total count | 获取总数
	total, err := s.categoryRepo.CountWithCondition(ctx, conditionFunc)
	if err != nil {
		s.logger.Error("获取版块总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块总数失败: %w", err)
	}

	// Paginated query | 分页查询
	categories, err := s.categoryRepo.ListWithCondition(ctx, func(q *ent.CategoryQuery) *ent.CategoryQuery {
		q = conditionFunc(q)
		return q.Order(ent.Desc(category.FieldCreatedAt)).
			Offset((req.Page - 1) * req.PageSize)
	}, req.PageSize)
	if err != nil {
		s.logger.Error("获取版块列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块列表失败: %w", err)
	}

	// Convert to response format | 转换为响应格式
	list := make([]schema.CategoryListItem, len(categories))
	for i, cat := range categories {
		list[i] = schema.CategoryListItem{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.Icon,
			Weight:      cat.Weight,
			Status:      cat.Status.String(),
			CreatedAt:   cat.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:   cat.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.CategoryListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateCategory Create category | 创建版块
func (s *CategoryManageService) CreateCategory(ctx context.Context, req schema.CategoryCreateRequest) (*ent.Category, error) {
	s.logger.Info("创建版块", zap.String("name", req.Name), tracing.WithTraceIDField(ctx))

	// Check if slug already exists | 检查slug是否已存在
	exists, err := s.categoryRepo.ExistsBySlug(ctx, req.Slug)
	if err != nil {
		s.logger.Error("检查版块标识失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查版块标识失败: %w", err)
	}
	if exists {
		return nil, errors.New("版块标识已存在")
	}

	// Create category | 创建版块
	categories, err := s.categoryRepo.Create(ctx, req.Name, req.Slug, req.Description, req.Icon, req.Weight, category.Status(req.Status))
	if err != nil {
		s.logger.Error("创建版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建版块失败: %w", err)
	}

	s.logger.Info("版块创建成功", zap.Int("id", categories.ID), tracing.WithTraceIDField(ctx))
	return categories, nil
}

// UpdateCategory Update category information | 更新版块信息
func (s *CategoryManageService) UpdateCategory(ctx context.Context, req schema.CategoryUpdateRequest) (*ent.Category, error) {
	s.logger.Info("更新版块信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// Check if category exists | 检查版块是否存在
	existingCategory, err := s.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// If updating slug, check for conflicts with other categories | 如果要更新slug，检查是否与其他版块冲突
	if req.Slug != "" && req.Slug != existingCategory.Slug {
		exists, err := s.categoryRepo.ExistsBySlugExcludeID(ctx, req.Slug, req.ID)
		if err != nil {
			s.logger.Error("检查版块标识失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查版块标识失败: %w", err)
		}
		if exists {
			return nil, errors.New("版块标识已存在")
		}
	}

	// Update category | 更新版块
	updatedCategory, err := s.categoryRepo.Update(ctx, req.ID, func(u *ent.CategoryUpdateOne) *ent.CategoryUpdateOne {
		if req.Name != "" {
			u = u.SetName(req.Name)
		}
		if req.Slug != "" {
			u = u.SetSlug(req.Slug)
		}
		if req.Description != "" {
			u = u.SetDescription(req.Description)
		}
		if req.Icon != "" {
			u = u.SetIcon(req.Icon)
		}
		if req.Weight != 0 {
			u = u.SetWeight(req.Weight)
		}
		if req.Status != "" {
			u = u.SetStatus(category.Status(req.Status))
		}
		return u
	})
	if err != nil {
		s.logger.Error("更新版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新版块失败: %w", err)
	}

	s.logger.Info("版块更新成功", zap.Int("id", updatedCategory.ID), tracing.WithTraceIDField(ctx))
	return updatedCategory, nil
}

// UpdateCategoryStatus Update category status | 更新版块状态
func (s *CategoryManageService) UpdateCategoryStatus(ctx context.Context, req schema.CategoryStatusUpdateRequest) error {
	s.logger.Info("更新版块状态", zap.Int("id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// Check if category exists | 检查版块是否存在
	exists, err := s.categoryRepo.ExistsByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查版块失败: %w", err)
	}
	if !exists {
		return errors.New("版块不存在")
	}

	// Update status | 更新状态
	_, err = s.categoryRepo.Update(ctx, req.ID, func(u *ent.CategoryUpdateOne) *ent.CategoryUpdateOne {
		return u.SetStatus(category.Status(req.Status))
	})
	if err != nil {
		s.logger.Error("更新版块状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新版块状态失败: %w", err)
	}

	s.logger.Info("版块状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetCategoryDetail Get category details | 获取版块详情
func (s *CategoryManageService) GetCategoryDetail(ctx context.Context, id int) (*schema.CategoryDetailResponse, error) {
	s.logger.Info("获取版块详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Get category information | 获取版块信息
	categories, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          categories.ID,
		Name:        categories.Name,
		Slug:        categories.Slug,
		Description: categories.Description,
		Icon:        categories.Icon,
		Weight:      categories.Weight,
		Status:      categories.Status.String(),
		CreatedAt:   categories.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:   categories.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	return result, nil
}

// SetCategoryModerators Set category moderators | 设置版块版主
func (s *CategoryManageService) SetCategoryModerators(ctx context.Context, req schema.CategoryModeratorRequest) error {
	s.logger.Info("设置版块版主",
		zap.Int("category_id", req.CategoryID),
		zap.Int("user_id", req.UserID),
		tracing.WithTraceIDField(ctx))

	// Check if category exists | 检查版块是否存在
	_, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// Check if user exists and has moderator role | 检查用户是否存在且是版主身份
	u, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查用户失败: %w", err)
	}
	if u.Role != user.RoleModerator {
		return errors.New("用户不是版主身份")
	}

	// Use transaction to ensure data consistency | 使用事务确保数据一致性
	tx, err := s.db.Tx(ctx)
	if err != nil {
		s.logger.Error("开启事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("开启事务失败: %w", err)
	}

	// Check if user is already a moderator of this category | 检查该用户是否已经是该版块的版主
	existsModerator, err := s.moderatorRepo.Exists(ctx, req.CategoryID, req.UserID)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		s.logger.Error("检查版主关联失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查版主关联失败: %w", err)
	}

	// If not a moderator, add moderator association | 如果不是版主，则添加版主关联
	if !existsModerator {
		err = s.moderatorRepo.Create(ctx, req.CategoryID, req.UserID)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return err
			}
			s.logger.Error("添加版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("添加版主关联记录失败: %w", err)
		}
	}

	// Commit transaction | 提交事务
	if err = tx.Commit(); err != nil {
		s.logger.Error("提交事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Info("版块版主设置成功",
		zap.Int("category_id", req.CategoryID),
		zap.Int("user_id", req.UserID),
		tracing.WithTraceIDField(ctx))
	return nil
}

// RemoveCategoryModerator Remove category moderator | 移除版块版主
func (s *CategoryManageService) RemoveCategoryModerator(ctx context.Context, req schema.CategoryModeratorRequest) error {
	s.logger.Info("移除版块版主",
		zap.Int("category_id", req.CategoryID),
		zap.Int("user_id", req.UserID),
		tracing.WithTraceIDField(ctx))

	// Check if category exists | 检查版块是否存在
	_, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// Check if user exists | 检查用户是否存在
	userExists, err := s.userRepo.ExistsByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查用户失败: %w", err)
	}
	if !userExists {
		return errors.New("用户不存在")
	}

	// Delete moderator association record | 删除版主关联记录
	affected, err := s.moderatorRepo.Delete(ctx, req.CategoryID, req.UserID)
	if err != nil {
		s.logger.Error("删除版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除版主关联记录失败: %w", err)
	}

	if affected == 0 {
		return errors.New("该用户不是该版块的版主")
	}

	s.logger.Info("版块版主移除成功",
		zap.Int("category_id", req.CategoryID),
		zap.Int("user_id", req.UserID),
		tracing.WithTraceIDField(ctx))
	return nil
}

// DeleteCategory Delete category (soft delete, set status to Hidden) | 删除版块（软删除，状态设为Hidden）
func (s *CategoryManageService) DeleteCategory(ctx context.Context, id int) error {
	s.logger.Info("删除版块", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Check if category exists | 检查版块是否存在
	exists, err := s.categoryRepo.ExistsByID(ctx, id)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查版块失败: %w", err)
	}
	if !exists {
		return errors.New("版块不存在")
	}

	// Soft delete: set status to Hidden | 软删除：将状态设为Hidden
	_, err = s.categoryRepo.Update(ctx, id, func(u *ent.CategoryUpdateOne) *ent.CategoryUpdateOne {
		return u.SetStatus(category.StatusHidden)
	})
	if err != nil {
		s.logger.Error("删除版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除版块失败: %w", err)
	}

	s.logger.Info("版块删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
