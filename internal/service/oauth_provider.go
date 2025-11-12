package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/oauthprovider"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IOAuthProviderService OAuth提供商管理服务接口
type IOAuthProviderService interface {
	// GetProviderList 获取OAuth提供商列表
	GetProviderList(ctx context.Context, req schema.OAuthProviderListRequest) (*schema.OAuthProviderListResponse, error)
	// CreateProvider 创建OAuth提供商
	CreateProvider(ctx context.Context, req schema.OAuthProviderCreateRequest) (*ent.OAuthProvider, error)
	// UpdateProvider 更新OAuth提供商信息
	UpdateProvider(ctx context.Context, req schema.OAuthProviderUpdateRequest) (*ent.OAuthProvider, error)
	// UpdateProviderStatus 更新OAuth提供商状态
	UpdateProviderStatus(ctx context.Context, req schema.OAuthProviderStatusUpdateRequest) error
	// GetProviderDetail 获取OAuth提供商详情
	GetProviderDetail(ctx context.Context, id int) (*schema.OAuthProviderDetailResponse, error)
	// DeleteProvider 删除OAuth提供商
	DeleteProvider(ctx context.Context, id int) error
}

// OAuthProviderService OAuth提供商管理服务实现
type OAuthProviderService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewOAuthProviderService 创建OAuth提供商管理服务实例
func NewOAuthProviderService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IOAuthProviderService {
	return &OAuthProviderService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetProviderList 获取OAuth提供商列表
func (s *OAuthProviderService) GetProviderList(ctx context.Context, req schema.OAuthProviderListRequest) (*schema.OAuthProviderListResponse, error) {
	s.logger.Info("获取OAuth提供商列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.OAuthProvider.Query()

	// 提供商类型筛选
	if req.Provider != "" {
		query = query.Where(oauthprovider.ProviderEQ(oauthprovider.Provider(req.Provider)))
	}

	// 启用状态筛选
	if req.Enabled != nil {
		query = query.Where(oauthprovider.EnabledEQ(*req.Enabled))
	}

	// 查询所有符合条件的OAuth提供商，按排序顺序和创建时间排序
	providers, err := query.
		Order(ent.Asc(oauthprovider.FieldSortOrder), ent.Desc(oauthprovider.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		s.logger.Error("获取OAuth提供商列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取OAuth提供商列表失败: %w", err)
	}

	// 转换为响应格式
	list := make([]schema.OAuthProviderListItem, len(providers))
	for i, provider := range providers {
		list[i] = schema.OAuthProviderListItem{
			ID:          provider.ID,
			Provider:    provider.Provider.String(),
			ClientID:    provider.ClientID,
			AuthURL:     provider.AuthURL,
			TokenURL:    provider.TokenURL,
			UserInfoURL: provider.UserInfoURL,
			RedirectURL: provider.RedirectURL,
			Scopes:      provider.Scopes,
			Enabled:     provider.Enabled,
			SortOrder:   provider.SortOrder,
			CreatedAt:   provider.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   provider.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.OAuthProviderListResponse{
		List: list,
	}, nil
}

// CreateProvider 创建OAuth提供商
func (s *OAuthProviderService) CreateProvider(ctx context.Context, req schema.OAuthProviderCreateRequest) (*ent.OAuthProvider, error) {
	s.logger.Info("创建OAuth提供商", zap.String("provider", req.Provider), tracing.WithTraceIDField(ctx))

	// 检查提供商是否已存在
	exists, err := s.db.OAuthProvider.Query().
		Where(oauthprovider.ProviderEQ(oauthprovider.Provider(req.Provider))).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if exists {
		return nil, errors.New("该OAuth提供商已存在")
	}

	// 创建OAuth提供商
	provider, err := s.db.OAuthProvider.Create().
		SetProvider(oauthprovider.Provider(req.Provider)).
		SetClientID(req.ClientID).
		SetClientSecret(req.ClientSecret).
		SetAuthURL(req.AuthURL).
		SetTokenURL(req.TokenURL).
		SetUserInfoURL(req.UserInfoURL).
		SetRedirectURL(req.RedirectURL).
		SetScopes(req.Scopes).
		SetExtraConfig(req.ExtraConfig).
		SetEnabled(req.Enabled).
		SetSortOrder(req.SortOrder).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth提供商创建成功", zap.Int("id", provider.ID), tracing.WithTraceIDField(ctx))
	return provider, nil
}

// UpdateProvider 更新OAuth提供商信息
func (s *OAuthProviderService) UpdateProvider(ctx context.Context, req schema.OAuthProviderUpdateRequest) (*ent.OAuthProvider, error) {
	s.logger.Info("更新OAuth提供商信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查提供商是否存在
	_, err := s.db.OAuthProvider.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("OAuth提供商不存在")
		}
		s.logger.Error("获取OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取OAuth提供商失败: %w", err)
	}

	// 构建更新操作
	update := s.db.OAuthProvider.UpdateOneID(req.ID)

	if req.ClientID != "" {
		update = update.SetClientID(req.ClientID)
	}
	// 只有当ClientSecret不为空时才更新（避免意外清空）
	if req.ClientSecret != "" {
		update = update.SetClientSecret(req.ClientSecret)
	}
	if req.AuthURL != "" {
		update = update.SetAuthURL(req.AuthURL)
	}
	if req.TokenURL != "" {
		update = update.SetTokenURL(req.TokenURL)
	}
	if req.UserInfoURL != "" {
		update = update.SetUserInfoURL(req.UserInfoURL)
	}
	if req.RedirectURL != "" {
		update = update.SetRedirectURL(req.RedirectURL)
	}
	if req.Scopes != nil {
		update = update.SetScopes(req.Scopes)
	}
	if req.ExtraConfig != nil {
		update = update.SetExtraConfig(req.ExtraConfig)
	}
	if req.Enabled != nil {
		update = update.SetEnabled(*req.Enabled)
	}
	if req.SortOrder != nil {
		update = update.SetSortOrder(*req.SortOrder)
	}

	// 执行更新
	provider, err := update.Save(ctx)
	if err != nil {
		s.logger.Error("更新OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth提供商更新成功", zap.Int("id", provider.ID), tracing.WithTraceIDField(ctx))
	return provider, nil
}

// UpdateProviderStatus 更新OAuth提供商状态
func (s *OAuthProviderService) UpdateProviderStatus(ctx context.Context, req schema.OAuthProviderStatusUpdateRequest) error {
	s.logger.Info("更新OAuth提供商状态",
		zap.Int("id", req.ID),
		zap.Bool("enabled", req.Enabled),
		tracing.WithTraceIDField(ctx))

	// 检查提供商是否存在
	exists, err := s.db.OAuthProvider.Query().
		Where(oauthprovider.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if !exists {
		return errors.New("OAuth提供商不存在")
	}

	// 更新状态
	err = s.db.OAuthProvider.UpdateOneID(req.ID).
		SetEnabled(req.Enabled).
		Exec(ctx)
	if err != nil {
		s.logger.Error("更新OAuth提供商状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新OAuth提供商状态失败: %w", err)
	}

	s.logger.Info("OAuth提供商状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetProviderDetail 获取OAuth提供商详情
func (s *OAuthProviderService) GetProviderDetail(ctx context.Context, id int) (*schema.OAuthProviderDetailResponse, error) {
	s.logger.Info("获取OAuth提供商详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	provider, err := s.db.OAuthProvider.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("OAuth提供商不存在")
		}
		s.logger.Error("获取OAuth提供商详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取OAuth提供商详情失败: %w", err)
	}

	// 脱敏处理ClientSecret
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
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		ExtraConfig:  provider.ExtraConfig,
		Enabled:      provider.Enabled,
		SortOrder:    provider.SortOrder,
		CreatedAt:    provider.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    provider.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// DeleteProvider 删除OAuth提供商
func (s *OAuthProviderService) DeleteProvider(ctx context.Context, id int) error {
	s.logger.Info("删除OAuth提供商", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 检查提供商是否存在
	exists, err := s.db.OAuthProvider.Query().
		Where(oauthprovider.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查OAuth提供商失败: %w", err)
	}
	if !exists {
		return errors.New("OAuth提供商不存在")
	}

	// 删除提供商
	err = s.db.OAuthProvider.DeleteOneID(id).Exec(ctx)
	if err != nil {
		s.logger.Error("删除OAuth提供商失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除OAuth提供商失败: %w", err)
	}

	s.logger.Info("OAuth提供商删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
