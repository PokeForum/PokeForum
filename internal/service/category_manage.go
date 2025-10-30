package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// ICategoryManageService 版块管理服务接口
type ICategoryManageService interface {
	// GetCategoryList 获取版块列表
	GetCategoryList(ctx context.Context, req schema.CategoryListRequest) (*schema.CategoryListResponse, error)
	// CreateCategory 创建版块
	CreateCategory(ctx context.Context, req schema.CategoryCreateRequest) (*ent.Category, error)
	// UpdateCategory 更新版块信息
	UpdateCategory(ctx context.Context, req schema.CategoryUpdateRequest) (*ent.Category, error)
	// UpdateCategoryStatus 更新版块状态
	UpdateCategoryStatus(ctx context.Context, req schema.CategoryStatusUpdateRequest) error
	// GetCategoryDetail 获取版块详情
	GetCategoryDetail(ctx context.Context, id int) (*schema.CategoryDetailResponse, error)
	// SetCategoryModerators 设置版块版主
	SetCategoryModerators(ctx context.Context, req schema.CategoryModeratorRequest) error
	// DeleteCategory 隐藏版块（软删除，状态设为Hidden）
	DeleteCategory(ctx context.Context, id int) error
}

// CategoryManageService 版块管理服务实现
type CategoryManageService struct {
	db     *ent.Client
	cache  *redis.Pool
	logger *zap.Logger
}

// NewCategoryManageService 创建版块管理服务实例
func NewCategoryManageService(db *ent.Client, cache *redis.Pool, logger *zap.Logger) ICategoryManageService {
	return &CategoryManageService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// GetCategoryList 获取版块列表
func (s *CategoryManageService) GetCategoryList(ctx context.Context, req schema.CategoryListRequest) (*schema.CategoryListResponse, error) {
	s.logger.Info("获取版块列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Category.Query()

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where(
			category.Or(
				category.NameContains(req.Keyword),
				category.DescriptionContains(req.Keyword),
			),
		)
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where(category.StatusEQ(category.Status(req.Status)))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取版块总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块总数失败: %w", err)
	}

	// 分页查询
	categories, err := query.
		Order(ent.Asc(category.FieldWeight), ent.Desc(category.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取版块列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块列表失败: %w", err)
	}

	// 转换为响应格式
	list := make([]schema.CategoryListItem, len(categories))
	for i, cat := range categories {
		// TODO: 添加帖子统计功能，暂时设为0
		postCount := 0

		list[i] = schema.CategoryListItem{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.Icon,
			Weight:      cat.Weight,
			Status:      cat.Status.String(),
			PostCount:   postCount,
			CreatedAt:   cat.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   cat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.CategoryListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateCategory 创建版块
func (s *CategoryManageService) CreateCategory(ctx context.Context, req schema.CategoryCreateRequest) (*ent.Category, error) {
	s.logger.Info("创建版块", zap.String("name", req.Name), tracing.WithTraceIDField(ctx))

	// 检查slug是否已存在
	exists, err := s.db.Category.Query().
		Where(category.SlugEQ(req.Slug)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查版块标识失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查版块标识失败: %w", err)
	}
	if exists {
		return nil, errors.New("版块标识已存在")
	}

	// 创建版块
	category, err := s.db.Category.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetDescription(req.Description).
		SetIcon(req.Icon).
		SetWeight(req.Weight).
		SetStatus(category.Status(req.Status)).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建版块失败: %w", err)
	}

	s.logger.Info("版块创建成功", zap.Int("id", category.ID), tracing.WithTraceIDField(ctx))
	return category, nil
}

// UpdateCategory 更新版块信息
func (s *CategoryManageService) UpdateCategory(ctx context.Context, req schema.CategoryUpdateRequest) (*ent.Category, error) {
	s.logger.Info("更新版块信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	category, err := s.db.Category.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("版块不存在")
		}
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块失败: %w", err)
	}

	// 如果要更新slug，检查是否与其他版块冲突
	if req.Slug != "" && req.Slug != category.Slug {
		exists, err := s.db.Category.Query().
			Where(
				category.And(
					category.SlugEQ(req.Slug),
					category.IDNEQ(req.ID),
				),
			).
			Exist(ctx)
		if err != nil {
			s.logger.Error("检查版块标识失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查版块标识失败: %w", err)
		}
		if exists {
			return nil, errors.New("版块标识已存在")
		}
	}

	// 构建更新操作
	update := s.db.Category.UpdateOneID(req.ID)

	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.Slug != "" {
		update = update.SetSlug(req.Slug)
	}
	if req.Description != "" {
		update = update.SetDescription(req.Description)
	}
	if req.Icon != "" {
		update = update.SetIcon(req.Icon)
	}
	if req.Weight != 0 {
		update = update.SetWeight(req.Weight)
	}
	if req.Status != "" {
		update = update.SetStatus(category.Status(req.Status))
	}

	// 执行更新
	updatedCategory, err := update.Save(ctx)
	if err != nil {
		s.logger.Error("更新版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新版块失败: %w", err)
	}

	s.logger.Info("版块更新成功", zap.Int("id", updatedCategory.ID), tracing.WithTraceIDField(ctx))
	return updatedCategory, nil
}

// UpdateCategoryStatus 更新版块状态
func (s *CategoryManageService) UpdateCategoryStatus(ctx context.Context, req schema.CategoryStatusUpdateRequest) error {
	s.logger.Info("更新版块状态", zap.Int("id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	exists, err := s.db.Category.Query().
		Where(category.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查版块失败: %w", err)
	}
	if !exists {
		return errors.New("版块不存在")
	}

	// 更新状态
	_, err = s.db.Category.UpdateOneID(req.ID).
		SetStatus(category.Status(req.Status)).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新版块状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新版块状态失败: %w", err)
	}

	s.logger.Info("版块状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetCategoryDetail 获取版块详情
func (s *CategoryManageService) GetCategoryDetail(ctx context.Context, id int) (*schema.CategoryDetailResponse, error) {
	s.logger.Info("获取版块详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 获取版块信息
	category, err := s.db.Category.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("版块不存在")
		}
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块失败: %w", err)
	}

	// TODO: 添加帖子统计功能，暂时设为0
	postCount := 0

	// 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Weight:      category.Weight,
		Status:      category.Status.String(),
		PostCount:   postCount,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return result, nil
}

// SetCategoryModerators 设置版块版主
func (s *CategoryManageService) SetCategoryModerators(ctx context.Context, req schema.CategoryModeratorRequest) error {
	s.logger.Info("设置版块版主", zap.Int("category_id", req.CategoryID), zap.Ints("user_ids", req.UserIDs), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	category, err := s.db.Category.Get(ctx, req.CategoryID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("版块不存在")
		}
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取版块失败: %w", err)
	}

	// 检查所有用户是否存在且是版主身份
	users, err := s.db.User.Query().
		Where(
			user.And(
				user.IDIn(req.UserIDs...),
				user.RoleEQ(user.RoleModerator),
			),
		).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户失败: %w", err)
	}

	if len(users) != len(req.UserIDs) {
		return errors.New("部分用户不存在或不是版主身份")
	}

	// 清除现有的版主关联
	_, err = s.db.Category.UpdateOne(category).
		ClearModerators().
		Save(ctx)
	if err != nil {
		s.logger.Error("清除版块版主失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("清除版块版主失败: %w", err)
	}

	// 添加新的版主关联
	if len(req.UserIDs) > 0 {
		_, err = s.db.Category.UpdateOne(category).
			AddModeratorIDs(req.UserIDs...).
			Save(ctx)
		if err != nil {
			s.logger.Error("设置版块版主失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("设置版块版主失败: %w", err)
		}
	}

	s.logger.Info("版块版主设置成功", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
	return nil
}

// DeleteCategory 删除版块（软删除，状态设为Hidden）
func (s *CategoryManageService) DeleteCategory(ctx context.Context, id int) error {
	s.logger.Info("删除版块", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	exists, err := s.db.Category.Query().
		Where(category.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查版块失败: %w", err)
	}
	if !exists {
		return errors.New("版块不存在")
	}

	// 软删除：将状态设为Hidden
	_, err = s.db.Category.UpdateOneID(id).
		SetStatus(category.StatusHidden).
		Save(ctx)
	if err != nil {
		s.logger.Error("删除版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除版块失败: %w", err)
	}

	s.logger.Info("版块删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
