package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// ICategoryService User-side category service interface | 用户侧版块服务接口
type ICategoryService interface {
	// GetUserCategories Get list of categories visible to users | 获取用户可见的版块列表
	// isLoggedIn: 用户是否已登录，影响返回的版块列表
	GetUserCategories(ctx context.Context, isLoggedIn bool) (*schema.UserCategoryResponse, error)
}

// CategoryService User-side category service implementation | 用户侧版块服务实现
type CategoryService struct {
	categoryRepo repository.ICategoryRepository
	cache        cache.ICacheService
	logger       *zap.Logger
}

// NewCategoryService Create user-side category service instance | 创建用户侧版块服务实例
func NewCategoryService(categoryRepo repository.ICategoryRepository, cacheService cache.ICacheService, logger *zap.Logger) ICategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		cache:        cacheService,
		logger:       logger,
	}
}

// GetUserCategories Get list of categories visible to users | 获取用户可见的版块列表
// Normal: 所有人可见
// LoginRequired: 仅登录用户可见
// Locked: 所有人可见，但不允许发帖
// Hidden: 不在列表返回，但可通过URL直接访问
func (s *CategoryService) GetUserCategories(ctx context.Context, isLoggedIn bool) (*schema.UserCategoryResponse, error) {
	s.logger.Info("获取用户版块列表", zap.Bool("is_logged_in", isLoggedIn), tracing.WithTraceIDField(ctx))

	// 根据登录状态选择缓存key
	cacheKey := _const.UserCategoryListCacheKey
	if isLoggedIn {
		cacheKey = _const.UserCategoryListLoggedInCacheKey
	}

	// 先尝试从缓存获取
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		s.logger.Info("从缓存获取用户版块列表", tracing.WithTraceIDField(ctx))
		var result schema.UserCategoryResponse
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			return &result, nil
		}
	}

	// 缓存未命中，从数据库查询
	categories, err := s.categoryRepo.GetVisibleCategories(ctx, isLoggedIn)
	if err != nil {
		s.logger.Error("获取用户版块列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块列表失败: %w", err)
	}

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

	result := &schema.UserCategoryResponse{
		List: list,
	}

	// 写入缓存，30天过期
	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.logger.Warn("序列化用户版块列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	} else {
		if err := s.cache.SetEx(ctx, cacheKey, resultJSON, 30*24*60*60); err != nil {
			s.logger.Warn("写入用户版块列表缓存失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		}
	}

	return result, nil
}
