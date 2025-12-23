package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// ICategoryService 用户侧版块服务接口
type ICategoryService interface {
	// GetUserCategories 获取用户可见的版块列表
	GetUserCategories(ctx context.Context) (*schema.UserCategoryResponse, error)
}

// CategoryService 用户侧版块服务实现
type CategoryService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewCategoryService 创建用户侧版块服务实例
func NewCategoryService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ICategoryService {
	return &CategoryService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetUserCategories 获取用户可见的版块列表
// 用户侧可以看到状态为 Normal、LoginRequired、Locked 的版块
// 其中 Locked 状态的版块只能查看，不能发帖和评论
// 隐藏(Hidden)的版块对用户不可见
func (s *CategoryService) GetUserCategories(ctx context.Context) (*schema.UserCategoryResponse, error) {
	s.logger.Info("获取用户版块列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件，只查询用户可见的版块状态
	categories, err := s.db.Category.Query().Where(
		category.Or(
			category.StatusEQ(category.StatusNormal),
			category.StatusEQ(category.StatusLoginRequired),
			category.StatusEQ(category.StatusLocked),
		),
	).
		// 按权重升序、创建时间降序排列
		Order(ent.Asc(category.FieldWeight), ent.Desc(category.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户版块列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块列表失败: %w", err)
	}

	// 转换为响应格式
	list := make([]schema.UserCategoryListItem, len(categories))
	for i, cat := range categories {
		list[i] = schema.UserCategoryListItem{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.Icon,
			Weight:      cat.Weight,
			CreatedAt:   cat.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.UserCategoryResponse{
		List: list,
	}, nil
}
