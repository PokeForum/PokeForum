package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/blacklist"
)

// IBlacklistRepository Blacklist repository interface | 黑名单仓储接口
type IBlacklistRepository interface {
	// Create Create blacklist record | 创建黑名单记录
	Create(ctx context.Context, userID, blockedUserID int) (*ent.Blacklist, error)
	// GetByUserID Get blacklist by user ID | 根据用户ID获取黑名单列表
	GetByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Blacklist, int, error)
	// Delete Delete blacklist record | 删除黑名单记录
	Delete(ctx context.Context, userID, blockedUserID int) (int, error)
	// Exists Check if user is blocked | 检查用户是否被拉黑
	Exists(ctx context.Context, userID, targetUserID int) (bool, error)
}

// BlacklistRepository Blacklist repository implementation | 黑名单仓储实现
type BlacklistRepository struct {
	db *ent.Client
}

// NewBlacklistRepository Create blacklist repository instance | 创建黑名单仓储实例
func NewBlacklistRepository(db *ent.Client) IBlacklistRepository {
	return &BlacklistRepository{db: db}
}

// Create Create blacklist record | 创建黑名单记录
func (r *BlacklistRepository) Create(ctx context.Context, userID, blockedUserID int) (*ent.Blacklist, error) {
	item, err := r.db.Blacklist.Create().
		SetUserID(userID).
		SetBlockedUserID(blockedUserID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建黑名单记录失败: %w", err)
	}
	return item, nil
}

// GetByUserID Get blacklist by user ID | 根据用户ID获取黑名单列表
func (r *BlacklistRepository) GetByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Blacklist, int, error) {
	query := r.db.Blacklist.Query().
		Where(blacklist.UserIDEQ(userID)).
		Order(ent.Desc(blacklist.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("查询黑名单总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	items, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("查询黑名单列表失败: %w", err)
	}

	return items, total, nil
}

// Delete Delete blacklist record | 删除黑名单记录
func (r *BlacklistRepository) Delete(ctx context.Context, userID, blockedUserID int) (int, error) {
	affected, err := r.db.Blacklist.Delete().
		Where(
			blacklist.UserIDEQ(userID),
			blacklist.BlockedUserIDEQ(blockedUserID),
		).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("删除黑名单记录失败: %w", err)
	}
	return affected, nil
}

// Exists Check if user is blocked | 检查用户是否被拉黑
func (r *BlacklistRepository) Exists(ctx context.Context, userID, targetUserID int) (bool, error) {
	exists, err := r.db.Blacklist.Query().
		Where(
			blacklist.UserIDEQ(userID),
			blacklist.BlockedUserIDEQ(targetUserID),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("查询黑名单记录失败: %w", err)
	}
	return exists, nil
}
