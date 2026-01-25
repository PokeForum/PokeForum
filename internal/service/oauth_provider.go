package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IOAuthProviderService OAuth provider management service interface | OAuth提供商管理服务接口
type IOAuthProviderService interface {
	// GetProviderList Get OAuth provider list | 获取OAuth提供商列表
	GetProviderList(ctx context.Context, req schema.OAuthProviderListRequest) (*schema.OAuthProviderListResponse, error)
	// CreateProvider Create OAuth provider | 创建OAuth提供商
	CreateProvider(ctx context.Context, req schema.OAuthProviderCreateRequest) (*ent.OAuthProvider, error)
	// UpdateProvider Update OAuth provider information | 更新OAuth提供商信息
	UpdateProvider(ctx context.Context, req schema.OAuthProviderUpdateRequest) (*ent.OAuthProvider, error)
	// UpdateProviderStatus Update OAuth provider status | 更新OAuth提供商状态
	UpdateProviderStatus(ctx context.Context, req schema.OAuthProviderStatusUpdateRequest) error
	// GetProviderDetail Get OAuth provider details | 获取OAuth提供商详情
	GetProviderDetail(ctx context.Context, id int) (*schema.OAuthProviderDetailResponse, error)
	// DeleteProvider Delete OAuth provider | 删除OAuth提供商
	DeleteProvider(ctx context.Context, id int) error
}

// OAuthProviderService OAuth provider management service implementation | OAuth提供商管理服务实现
type OAuthProviderService struct {
	db        *ent.Client
	oauthRepo repository.IOAuthProviderRepository
	cache     cache.ICacheService
	logger    *zap.Logger
}

// NewOAuthProviderService Create OAuth provider management service instance | 创建OAuth提供商管理服务实例
func NewOAuthProviderService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) IOAuthProviderService {
	return &OAuthProviderService{
		db:        db,
		oauthRepo: repos.OAuthProvider,
		cache:     cacheService,
		logger:    logger,
	}
}

// GetProviderList Get OAuth provider list | 获取OAuth提供商列表
func (s *OAuthProviderService) GetProviderList(ctx context.Context, req schema.OAuthProviderListRequest) (*schema.OAuthProviderListResponse, error) {
	s.logger.Info("Get OAuth provider list | 获取OAuth提供商列表", tracing.WithTraceIDField(ctx))

	// Query OAuth providers with filters | 查询OAuth提供商（带筛选）
	providers, err := s.oauthRepo.List(ctx, req.Provider, req.Enabled)
	if err != nil {
		s.logger.Error("Failed to get OAuth provider list | 获取OAuth提供商列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取OAuth提供商列表失败: %w", err)
	}

	// Convert to response format | 转换为响应格式
	list := make([]schema.OAuthProviderListItem, len(providers))
	for i, provider := range providers {
		list[i] = schema.OAuthProviderListItem{
			ID:          provider.ID,
			Provider:    provider.Provider.String(),
			ClientID:    provider.ClientID,
			AuthURL:     provider.AuthURL,
			TokenURL:    provider.TokenURL,
			UserInfoURL: provider.UserInfoURL,
			Scopes:      provider.Scopes,
			Enabled:     provider.Enabled,
			SortOrder:   provider.SortOrder,
			CreatedAt:   provider.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:   provider.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.OAuthProviderListResponse{
		List: list,
	}, nil
}

// CreateProvider Create OAuth provider | 创建OAuth提供商
func (s *OAuthProviderService) CreateProvider(ctx context.Context, req schema.OAuthProviderCreateRequest) (*ent.OAuthProvider, error) {
	s.logger.Info("Create OAuth provider | 创建OAuth提供商", zap.String("provider", req.Provider), tracing.WithTraceIDField(ctx))

	// Check if provider already exists | 检查提供商是否已存在
	exists, err := s.oauthRepo.ExistsByProvider(ctx, req.Provider)
	if err != nil {
		s.logger.Error("Failed to check OAuth provider | 检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if exists {
		return nil, errors.New("该OAuth提供商已存在")
	}

	// Create OAuth provider | 创建OAuth提供商
	provider, err := s.oauthRepo.Create(ctx, req.Provider, req.ClientID, req.ClientSecret, req.AuthURL, req.TokenURL, req.UserInfoURL, req.Scopes, req.ExtraConfig, req.Enabled, req.SortOrder)
	if err != nil {
		s.logger.Error("Failed to create OAuth provider | 创建OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth provider created successfully | OAuth提供商创建成功", zap.Int("id", provider.ID), tracing.WithTraceIDField(ctx))
	return provider, nil
}

// UpdateProvider Update OAuth provider information | 更新OAuth提供商信息
func (s *OAuthProviderService) UpdateProvider(ctx context.Context, req schema.OAuthProviderUpdateRequest) (*ent.OAuthProvider, error) {
	s.logger.Info("Update OAuth provider information | 更新OAuth提供商信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// Check if provider exists | 检查提供商是否存在
	_, err := s.oauthRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("Failed to get OAuth provider | 获取OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Update OAuth provider | 更新OAuth提供商
	provider, err := s.oauthRepo.Update(ctx, req.ID, func(u *ent.OAuthProviderUpdateOne) *ent.OAuthProviderUpdateOne {
		u = u.SetClientID(req.ClientID).
			SetAuthURL(req.AuthURL).
			SetTokenURL(req.TokenURL).
			SetUserInfoURL(req.UserInfoURL).
			SetScopes(req.Scopes).
			SetExtraConfig(req.ExtraConfig)
		if req.ClientSecret != "" {
			u = u.SetClientSecret(req.ClientSecret)
		}
		if req.Enabled != nil {
			u = u.SetEnabled(*req.Enabled)
		}
		if req.SortOrder != nil {
			u = u.SetSortOrder(*req.SortOrder)
		}
		return u
	})
	if err != nil {
		s.logger.Error("Failed to update OAuth provider | 更新OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth provider updated successfully | OAuth提供商更新成功", zap.Int("id", provider.ID), tracing.WithTraceIDField(ctx))
	return provider, nil
}

// UpdateProviderStatus Update OAuth provider status | 更新OAuth提供商状态
func (s *OAuthProviderService) UpdateProviderStatus(ctx context.Context, req schema.OAuthProviderStatusUpdateRequest) error {
	s.logger.Info("Update OAuth provider status | 更新OAuth提供商状态",
		zap.Int("id", req.ID),
		zap.Bool("enabled", req.Enabled),
		tracing.WithTraceIDField(ctx))

	// Check if provider exists | 检查提供商是否存在
	exists, err := s.oauthRepo.ExistsByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("Failed to check OAuth provider | 检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if !exists {
		return errors.New("OAuth提供商不存在")
	}

	// Update status | 更新状态
	err = s.oauthRepo.UpdateStatus(ctx, req.ID, req.Enabled)
	if err != nil {
		s.logger.Error("Failed to update OAuth provider status | 更新OAuth提供商状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新OAuth提供商状态失败: %w", err)
	}

	s.logger.Info("OAuth provider status updated successfully | OAuth提供商状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetProviderDetail Get OAuth provider details | 获取OAuth提供商详情
func (s *OAuthProviderService) GetProviderDetail(ctx context.Context, id int) (*schema.OAuthProviderDetailResponse, error) {
	s.logger.Info("Get OAuth provider details | 获取OAuth提供商详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	provider, err := s.oauthRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get OAuth provider details | 获取OAuth提供商详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Mask ClientSecret | 脱敏处理ClientSecret
	maskedSecret := "***"
	if provider.ClientSecret != "" {
		maskedSecret = provider.ClientSecret[:3] + "***"
	}

	return &schema.OAuthProviderDetailResponse{
		ID:           provider.ID,
		Provider:     provider.Provider.String(),
		ClientID:     provider.ClientID,
		ClientSecret: maskedSecret,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		UserInfoURL:  provider.UserInfoURL,
		Scopes:       provider.Scopes,
		ExtraConfig:  provider.ExtraConfig,
		Enabled:      provider.Enabled,
		SortOrder:    provider.SortOrder,
		CreatedAt:    provider.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:    provider.UpdatedAt.Format(time_tools.DateTimeFormat),
	}, nil
}

// DeleteProvider Delete OAuth provider | 删除OAuth提供商
func (s *OAuthProviderService) DeleteProvider(ctx context.Context, id int) error {
	s.logger.Info("Delete OAuth provider | 删除OAuth提供商", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// Check if provider exists | 检查提供商是否存在
	exists, err := s.oauthRepo.ExistsByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to check OAuth provider | 检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if !exists {
		return errors.New("OAuth提供商不存在")
	}

	// Delete provider | 删除提供商
	err = s.oauthRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete OAuth provider | 删除OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth provider deleted successfully | OAuth提供商删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
