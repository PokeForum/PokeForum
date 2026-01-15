package repository

import (
	"context"
	"errors"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/oauthprovider"
)

// IOAuthProviderRepository OAuth provider repository interface | OAuth提供商仓储接口
type IOAuthProviderRepository interface {
	// List Get OAuth provider list with filters | 获取OAuth提供商列表（带筛选）
	List(ctx context.Context, provider string, enabled *bool) ([]*ent.OAuthProvider, error)
	// GetByID Get OAuth provider by ID | 根据ID获取OAuth提供商
	GetByID(ctx context.Context, id int) (*ent.OAuthProvider, error)
	// ExistsByProvider Check if provider exists by provider type | 检查提供商是否存在
	ExistsByProvider(ctx context.Context, provider string) (bool, error)
	// ExistsByID Check if provider exists by ID | 检查提供商是否存在
	ExistsByID(ctx context.Context, id int) (bool, error)
	// Create Create OAuth provider | 创建OAuth提供商
	Create(ctx context.Context, provider, clientID, clientSecret, authURL, tokenURL, userInfoURL string, scopes []string, extraConfig map[string]interface{}, enabled bool, sortOrder int) (*ent.OAuthProvider, error)
	// Update Update OAuth provider | 更新OAuth提供商
	Update(ctx context.Context, id int, updateFn func(*ent.OAuthProviderUpdateOne) *ent.OAuthProviderUpdateOne) (*ent.OAuthProvider, error)
	// UpdateStatus Update OAuth provider status | 更新OAuth提供商状态
	UpdateStatus(ctx context.Context, id int, enabled bool) error
	// Delete Delete OAuth provider | 删除OAuth提供商
	Delete(ctx context.Context, id int) error
}

// OAuthProviderRepository OAuth provider repository implementation | OAuth提供商仓储实现
type OAuthProviderRepository struct {
	db *ent.Client
}

// NewOAuthProviderRepository Create OAuth provider repository instance | 创建OAuth提供商仓储实例
func NewOAuthProviderRepository(db *ent.Client) IOAuthProviderRepository {
	return &OAuthProviderRepository{db: db}
}

// List Get OAuth provider list with filters | 获取OAuth提供商列表（带筛选）
func (r *OAuthProviderRepository) List(ctx context.Context, provider string, enabled *bool) ([]*ent.OAuthProvider, error) {
	query := r.db.OAuthProvider.Query()

	if provider != "" {
		query = query.Where(oauthprovider.ProviderEQ(oauthprovider.Provider(provider)))
	}

	if enabled != nil {
		query = query.Where(oauthprovider.EnabledEQ(*enabled))
	}

	return query.
		Order(ent.Asc(oauthprovider.FieldSortOrder), ent.Desc(oauthprovider.FieldCreatedAt)).
		All(ctx)
}

// GetByID Get OAuth provider by ID | 根据ID获取OAuth提供商
func (r *OAuthProviderRepository) GetByID(ctx context.Context, id int) (*ent.OAuthProvider, error) {
	provider, err := r.db.OAuthProvider.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("OAuth提供商不存在")
		}
		return nil, err
	}
	return provider, nil
}

// ExistsByProvider Check if provider exists by provider type | 检查提供商是否存在
func (r *OAuthProviderRepository) ExistsByProvider(ctx context.Context, provider string) (bool, error) {
	return r.db.OAuthProvider.Query().
		Where(oauthprovider.ProviderEQ(oauthprovider.Provider(provider))).
		Exist(ctx)
}

// ExistsByID Check if provider exists by ID | 检查提供商是否存在
func (r *OAuthProviderRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	return r.db.OAuthProvider.Query().
		Where(oauthprovider.IDEQ(id)).
		Exist(ctx)
}

// Create Create OAuth provider | 创建OAuth提供商
func (r *OAuthProviderRepository) Create(ctx context.Context, provider, clientID, clientSecret, authURL, tokenURL, userInfoURL string, scopes []string, extraConfig map[string]interface{}, enabled bool, sortOrder int) (*ent.OAuthProvider, error) {
	return r.db.OAuthProvider.Create().
		SetProvider(oauthprovider.Provider(provider)).
		SetClientID(clientID).
		SetClientSecret(clientSecret).
		SetAuthURL(authURL).
		SetTokenURL(tokenURL).
		SetUserInfoURL(userInfoURL).
		SetScopes(scopes).
		SetExtraConfig(extraConfig).
		SetEnabled(enabled).
		SetSortOrder(sortOrder).
		Save(ctx)
}

// Update Update OAuth provider | 更新OAuth提供商
func (r *OAuthProviderRepository) Update(ctx context.Context, id int, updateFn func(*ent.OAuthProviderUpdateOne) *ent.OAuthProviderUpdateOne) (*ent.OAuthProvider, error) {
	update := r.db.OAuthProvider.UpdateOneID(id)
	return updateFn(update).Save(ctx)
}

// UpdateStatus Update OAuth provider status | 更新OAuth提供商状态
func (r *OAuthProviderRepository) UpdateStatus(ctx context.Context, id int, enabled bool) error {
	return r.db.OAuthProvider.UpdateOneID(id).
		SetEnabled(enabled).
		Exec(ctx)
}

// Delete Delete OAuth provider | 删除OAuth提供商
func (r *OAuthProviderRepository) Delete(ctx context.Context, id int) error {
	return r.db.OAuthProvider.DeleteOneID(id).Exec(ctx)
}
