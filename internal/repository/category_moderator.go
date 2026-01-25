package repository

import (
	"context"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/categorymoderator"
)

// ICategoryModeratorRepository Category moderator repository interface | 版块版主仓储接口
type ICategoryModeratorRepository interface {
	// Exists Check if moderator exists | 检查版主是否存在
	Exists(ctx context.Context, categoryID, userID int) (bool, error)
	// GetByUserID Get category moderator records by user ID | 根据用户ID获取版主记录
	GetByUserID(ctx context.Context, userID int) ([]*ent.CategoryModerator, error)
	// Create Create category moderator | 创建版块版主
	Create(ctx context.Context, categoryID, userID int) error
	// Delete Delete category moderator | 删除版块版主
	Delete(ctx context.Context, categoryID, userID int) (int, error)
	// DeleteByUserID Delete all category moderators by user ID | 根据用户ID删除所有版主记录
	DeleteByUserID(ctx context.Context, userID int) error
}

// CategoryModeratorRepository Category moderator repository implementation | 版块版主仓储实现
type CategoryModeratorRepository struct {
	db *ent.Client
}

// NewCategoryModeratorRepository Create category moderator repository instance | 创建版块版主仓储实例
func NewCategoryModeratorRepository(db *ent.Client) ICategoryModeratorRepository {
	return &CategoryModeratorRepository{db: db}
}

// Exists Check if moderator exists | 检查版主是否存在
func (r *CategoryModeratorRepository) Exists(ctx context.Context, categoryID, userID int) (bool, error) {
	return r.db.CategoryModerator.Query().
		Where(
			categorymoderator.And(
				categorymoderator.CategoryIDEQ(categoryID),
				categorymoderator.UserIDEQ(userID),
			),
		).
		Exist(ctx)
}

// GetByUserID Get category moderator records by user ID | 根据用户ID获取版主记录
func (r *CategoryModeratorRepository) GetByUserID(ctx context.Context, userID int) ([]*ent.CategoryModerator, error) {
	return r.db.CategoryModerator.Query().
		Where(categorymoderator.UserIDEQ(userID)).
		All(ctx)
}

// Create Create category moderator | 创建版块版主
func (r *CategoryModeratorRepository) Create(ctx context.Context, categoryID, userID int) error {
	_, err := r.db.CategoryModerator.Create().
		SetCategoryID(categoryID).
		SetUserID(userID).
		Save(ctx)
	return err
}

// Delete Delete category moderator | 删除版块版主
func (r *CategoryModeratorRepository) Delete(ctx context.Context, categoryID, userID int) (int, error) {
	return r.db.CategoryModerator.Delete().
		Where(
			categorymoderator.And(
				categorymoderator.CategoryIDEQ(categoryID),
				categorymoderator.UserIDEQ(userID),
			),
		).
		Exec(ctx)
}

// DeleteByUserID Delete all category moderators by user ID | 根据用户ID删除所有版主记录
func (r *CategoryModeratorRepository) DeleteByUserID(ctx context.Context, userID int) error {
	_, err := r.db.CategoryModerator.Delete().
		Where(categorymoderator.UserIDEQ(userID)).
		Exec(ctx)
	return err
}
