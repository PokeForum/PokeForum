package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
)

// IUserRepository User repository interface | 用户仓储接口
type IUserRepository interface {
	// Create Create user | 创建用户
	Create(ctx context.Context, username, email, password string) (*ent.User, error)
	// CreateWithBuilder Create user with builder function | 使用构建器函数创建用户
	CreateWithBuilder(ctx context.Context, builderFunc func(*ent.UserCreate) *ent.UserCreate) (*ent.User, error)
	// GetByID Get user by ID | 根据ID获取用户
	GetByID(ctx context.Context, id int) (*ent.User, error)
	// GetByIDWithFields Get user by ID with specified fields | 根据ID获取用户（指定字段）
	GetByIDWithFields(ctx context.Context, id int, fields []string) (*ent.User, error)
	// GetByEmail Get user by email | 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
	// GetByUsername Get user by username | 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*ent.User, error)
	// GetByIDs Batch get users by IDs | 批量根据ID获取用户
	GetByIDs(ctx context.Context, ids []int) ([]*ent.User, error)
	// GetByIDsWithFields Batch get users by IDs with specified fields | 批量根据ID获取用户（指定字段）
	GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.User, error)
	// Update Update user | 更新用户
	Update(ctx context.Context, id int, updateFunc func(*ent.UserUpdateOne) *ent.UserUpdateOne) (*ent.User, error)
	// UpdateFields Update user fields without returning entity | 更新用户字段（不返回实体）
	UpdateFields(ctx context.Context, id int, updateFunc func(*ent.UserUpdateOne) *ent.UserUpdateOne) error
	// UpdatePassword Update user password | 更新用户密码
	UpdatePassword(ctx context.Context, id int, password string) error
	// UpdateAvatar Update user avatar | 更新用户头像
	UpdateAvatar(ctx context.Context, id int, avatarURL string) error
	// UpdateUsername Update username | 更新用户名
	UpdateUsername(ctx context.Context, id int, username string) error
	// UpdateEmailVerified Update email verification status | 更新邮箱验证状态
	UpdateEmailVerified(ctx context.Context, id int, verified bool) error
	// ExistsByEmail Check if email exists | 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	// ExistsByUsername Check if username exists | 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	// ExistsByID Check if user exists by ID | 检查用户是否存在
	ExistsByID(ctx context.Context, id int) (bool, error)
	// Count Get total user count | 获取用户总数
	Count(ctx context.Context) (int, error)
	// CountWithCondition Count users with condition | 条件统计用户数
	CountWithCondition(ctx context.Context, conditionFunc func(*ent.UserQuery) *ent.UserQuery) (int, error)
	// ListWithCondition List users with condition | 条件查询用户列表
	ListWithCondition(ctx context.Context, conditionFunc func(*ent.UserQuery) *ent.UserQuery, limit int) ([]*ent.User, error)
}

// UserRepository User repository implementation | 用户仓储实现
type UserRepository struct {
	db *ent.Client
}

// NewUserRepository Create user repository instance | 创建用户仓储实例
func NewUserRepository(db *ent.Client) IUserRepository {
	return &UserRepository{db: db}
}

// Create Create user | 创建用户
func (r *UserRepository) Create(ctx context.Context, username, email, password string) (*ent.User, error) {
	u, err := r.db.User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return u, nil
}

// CreateWithBuilder Create user with builder function | 使用构建器函数创建用户
func (r *UserRepository) CreateWithBuilder(ctx context.Context, builderFunc func(*ent.UserCreate) *ent.UserCreate) (*ent.User, error) {
	creator := r.db.User.Create()
	creator = builderFunc(creator)
	u, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return u, nil
}

// GetByID Get user by ID | 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := r.db.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// GetByIDWithFields Get user by ID with specified fields | 根据ID获取用户（指定字段）
func (r *UserRepository) GetByIDWithFields(ctx context.Context, id int, fields []string) (*ent.User, error) {
	query := r.db.User.Query().Where(user.IDEQ(id))
	var u *ent.User
	var err error
	if len(fields) > 0 {
		u, err = query.Select(fields...).Only(ctx)
	} else {
		u, err = query.Only(ctx)
	}
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// GetByEmail Get user by email | 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	u, err := r.db.User.Query().
		Where(user.EmailEQ(email)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// GetByUsername Get user by username | 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	u, err := r.db.User.Query().
		Where(user.UsernameEQ(username)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// GetByIDs Batch get users by IDs | 批量根据ID获取用户
func (r *UserRepository) GetByIDs(ctx context.Context, ids []int) ([]*ent.User, error) {
	if len(ids) == 0 {
		return []*ent.User{}, nil
	}
	users, err := r.db.User.Query().
		Where(user.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询用户失败: %w", err)
	}
	return users, nil
}

// GetByIDsWithFields Batch get users by IDs with specified fields | 批量根据ID获取用户（指定字段）
func (r *UserRepository) GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.User, error) {
	if len(ids) == 0 {
		return []*ent.User{}, nil
	}
	query := r.db.User.Query().Where(user.IDIn(ids...))
	if len(fields) > 0 {
		users, err := query.Select(fields...).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("批量查询用户失败: %w", err)
		}
		return users, nil
	}
	users, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询用户失败: %w", err)
	}
	return users, nil
}

// Update Update user | 更新用户
func (r *UserRepository) Update(ctx context.Context, id int, updateFunc func(*ent.UserUpdateOne) *ent.UserUpdateOne) (*ent.User, error) {
	updater := r.db.User.UpdateOneID(id)
	updater = updateFunc(updater)
	u, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}
	return u, nil
}

// UpdateFields Update user fields without returning entity | 更新用户字段（不返回实体）
func (r *UserRepository) UpdateFields(ctx context.Context, id int, updateFunc func(*ent.UserUpdateOne) *ent.UserUpdateOne) error {
	updater := r.db.User.UpdateOneID(id)
	updater = updateFunc(updater)
	err := updater.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新用户字段失败: %w", err)
	}
	return nil
}

// UpdatePassword Update user password | 更新用户密码
func (r *UserRepository) UpdatePassword(ctx context.Context, id int, password string) error {
	_, err := r.db.User.UpdateOneID(id).
		SetPassword(password).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	return nil
}

// UpdateAvatar Update user avatar | 更新用户头像
func (r *UserRepository) UpdateAvatar(ctx context.Context, id int, avatarURL string) error {
	_, err := r.db.User.UpdateOneID(id).
		SetAvatar(avatarURL).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新头像失败: %w", err)
	}
	return nil
}

// UpdateUsername Update username | 更新用户名
func (r *UserRepository) UpdateUsername(ctx context.Context, id int, username string) error {
	_, err := r.db.User.UpdateOneID(id).
		SetUsername(username).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新用户名失败: %w", err)
	}
	return nil
}

// UpdateEmailVerified Update email verification status | 更新邮箱验证状态
func (r *UserRepository) UpdateEmailVerified(ctx context.Context, id int, verified bool) error {
	_, err := r.db.User.UpdateOneID(id).
		SetEmailVerified(verified).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新邮箱验证状态失败: %w", err)
	}
	return nil
}

// ExistsByEmail Check if email exists | 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.db.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查邮箱是否存在失败: %w", err)
	}
	return exists, nil
}

// ExistsByUsername Check if username exists | 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.db.User.Query().
		Where(user.UsernameEQ(username)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查用户名是否存在失败: %w", err)
	}
	return exists, nil
}

// ExistsByID Check if user exists by ID | 检查用户是否存在
func (r *UserRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	exists, err := r.db.User.Query().
		Where(user.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	return exists, nil
}

// Count Get total user count | 获取用户总数
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.User.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询用户总数失败: %w", err)
	}
	return count, nil
}

// CountWithCondition Count users with condition | 条件统计用户数
func (r *UserRepository) CountWithCondition(ctx context.Context, conditionFunc func(*ent.UserQuery) *ent.UserQuery) (int, error) {
	query := r.db.User.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("条件统计用户数失败: %w", err)
	}
	return count, nil
}

// ListWithCondition List users with condition | 条件查询用户列表
func (r *UserRepository) ListWithCondition(ctx context.Context, conditionFunc func(*ent.UserQuery) *ent.UserQuery, limit int) ([]*ent.User, error) {
	query := r.db.User.Query()
	query = conditionFunc(query)
	if limit > 0 {
		query = query.Limit(limit)
	}
	users, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("条件查询用户列表失败: %w", err)
	}
	return users, nil
}
