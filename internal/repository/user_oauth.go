package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/useroauth"
)

// IUserOAuthRepository UserOAuth repository interface | UserOAuth仓储接口
type IUserOAuthRepository interface {
	// GetByProviderUserID Get binding by provider user ID | 根据提供商用户ID查询绑定
	GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*ent.UserOAuth, error)
	// GetByUserID Get all bindings by user ID | 获取用户所有绑定
	GetByUserID(ctx context.Context, userID int) ([]*ent.UserOAuth, error)
	// GetByUserIDAndProvider Get binding by user ID and provider | 获取用户特定提供商绑定
	GetByUserIDAndProvider(ctx context.Context, userID int, provider string) (*ent.UserOAuth, error)
	// Create Create binding | 创建绑定
	Create(ctx context.Context, userID int, provider, providerUserID, providerUsername, providerEmail, providerAvatar string, extraData map[string]interface{}) (*ent.UserOAuth, error)
	// Update Update binding | 更新绑定
	Update(ctx context.Context, id int, providerUsername, providerEmail, providerAvatar string, extraData map[string]interface{}) (*ent.UserOAuth, error)
	// Delete Delete binding | 删除绑定
	Delete(ctx context.Context, id int) error
	// CountByUserID Count bindings by user ID | 统计用户绑定数量
	CountByUserID(ctx context.Context, userID int) (int, error)
	// ExistsByProviderUserID Check if provider user ID exists | 检查提供商用户ID是否已存在
	ExistsByProviderUserID(ctx context.Context, provider, providerUserID string) (bool, error)
	// ExistsByUserIDAndProvider Check if user has bound this provider | 检查用户是否已绑定该提供商
	ExistsByUserIDAndProvider(ctx context.Context, userID int, provider string) (bool, error)
}

// UserOAuthRepository UserOAuth repository implementation | UserOAuth仓储实现
type UserOAuthRepository struct {
	db *ent.Client
}

// NewUserOAuthRepository Create UserOAuth repository instance | 创建UserOAuth仓储实例
func NewUserOAuthRepository(db *ent.Client) IUserOAuthRepository {
	return &UserOAuthRepository{db: db}
}

// GetByProviderUserID Get binding by provider user ID | 根据提供商用户ID查询绑定
func (r *UserOAuthRepository) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*ent.UserOAuth, error) {
	binding, err := r.db.UserOAuth.Query().
		Where(
			useroauth.ProviderEQ(useroauth.Provider(provider)),
			useroauth.ProviderUserIDEQ(providerUserID),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询OAuth绑定失败: %w", err)
	}
	return binding, nil
}

// GetByUserID Get all bindings by user ID | 获取用户所有绑定
func (r *UserOAuthRepository) GetByUserID(ctx context.Context, userID int) ([]*ent.UserOAuth, error) {
	bindings, err := r.db.UserOAuth.Query().
		Where(useroauth.UserIDEQ(userID)).
		Order(ent.Asc(useroauth.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询用户OAuth绑定列表失败: %w", err)
	}
	return bindings, nil
}

// GetByUserIDAndProvider Get binding by user ID and provider | 获取用户特定提供商绑定
func (r *UserOAuthRepository) GetByUserIDAndProvider(ctx context.Context, userID int, provider string) (*ent.UserOAuth, error) {
	binding, err := r.db.UserOAuth.Query().
		Where(
			useroauth.UserIDEQ(userID),
			useroauth.ProviderEQ(useroauth.Provider(provider)),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户OAuth绑定失败: %w", err)
	}
	return binding, nil
}

// Create Create binding | 创建绑定
func (r *UserOAuthRepository) Create(ctx context.Context, userID int, provider, providerUserID, providerUsername, providerEmail, providerAvatar string, extraData map[string]interface{}) (*ent.UserOAuth, error) {
	creator := r.db.UserOAuth.Create().
		SetUserID(userID).
		SetProvider(useroauth.Provider(provider)).
		SetProviderUserID(providerUserID)

	if providerUsername != "" {
		creator.SetProviderUsername(providerUsername)
	}
	if providerEmail != "" {
		creator.SetProviderEmail(providerEmail)
	}
	if providerAvatar != "" {
		creator.SetProviderAvatar(providerAvatar)
	}
	if extraData != nil {
		creator.SetExtraData(extraData)
	}

	binding, err := creator.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("该OAuth账号已被其他用户绑定")
		}
		return nil, fmt.Errorf("创建OAuth绑定失败: %w", err)
	}
	return binding, nil
}

// Update Update binding | 更新绑定
func (r *UserOAuthRepository) Update(ctx context.Context, id int, providerUsername, providerEmail, providerAvatar string, extraData map[string]interface{}) (*ent.UserOAuth, error) {
	updater := r.db.UserOAuth.UpdateOneID(id)

	if providerUsername != "" {
		updater.SetProviderUsername(providerUsername)
	}
	if providerEmail != "" {
		updater.SetProviderEmail(providerEmail)
	}
	if providerAvatar != "" {
		updater.SetProviderAvatar(providerAvatar)
	}
	if extraData != nil {
		updater.SetExtraData(extraData)
	}

	binding, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新OAuth绑定失败: %w", err)
	}
	return binding, nil
}

// Delete Delete binding | 删除绑定
func (r *UserOAuthRepository) Delete(ctx context.Context, id int) error {
	err := r.db.UserOAuth.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除OAuth绑定失败: %w", err)
	}
	return nil
}

// CountByUserID Count bindings by user ID | 统计用户绑定数量
func (r *UserOAuthRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	count, err := r.db.UserOAuth.Query().
		Where(useroauth.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户OAuth绑定数量失败: %w", err)
	}
	return count, nil
}

// ExistsByProviderUserID Check if provider user ID exists | 检查提供商用户ID是否已存在
func (r *UserOAuthRepository) ExistsByProviderUserID(ctx context.Context, provider, providerUserID string) (bool, error) {
	exists, err := r.db.UserOAuth.Query().
		Where(
			useroauth.ProviderEQ(useroauth.Provider(provider)),
			useroauth.ProviderUserIDEQ(providerUserID),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查OAuth绑定是否存在失败: %w", err)
	}
	return exists, nil
}

// ExistsByUserIDAndProvider Check if user has bound this provider | 检查用户是否已绑定该提供商
func (r *UserOAuthRepository) ExistsByUserIDAndProvider(ctx context.Context, userID int, provider string) (bool, error) {
	exists, err := r.db.UserOAuth.Query().
		Where(
			useroauth.UserIDEQ(userID),
			useroauth.ProviderEQ(useroauth.Provider(provider)),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查用户OAuth绑定是否存在失败: %w", err)
	}
	return exists, nil
}
