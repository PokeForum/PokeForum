package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
)

// ICategoryRepository Category repository interface | 版块仓储接口
type ICategoryRepository interface {
	// GetByID Get category by ID | 根据ID获取版块
	GetByID(ctx context.Context, id int) (*ent.Category, error)
	// GetByIDs Batch get categories by IDs | 批量根据ID获取版块
	GetByIDs(ctx context.Context, ids []int) ([]*ent.Category, error)
	// GetByIDsWithFields Batch get categories by IDs with specified fields | 批量根据ID获取版块（指定字段）
	GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.Category, error)
	// GetVisibleCategories Get visible categories for users | 获取用户可见的版块列表
	GetVisibleCategories(ctx context.Context) ([]*ent.Category, error)
	// ExistsByID Check if category exists | 检查版块是否存在
	ExistsByID(ctx context.Context, id int) (bool, error)
	// ExistsBySlug Check if category exists by slug | 检查版块标识是否存在
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	// ExistsBySlugExcludeID Check if category exists by slug excluding specific ID | 检查版块标识是否存在（排除指定ID）
	ExistsBySlugExcludeID(ctx context.Context, slug string, excludeID int) (bool, error)
	// Create Create category | 创建版块
	Create(ctx context.Context, name, slug, description, icon string, weight int, status category.Status) (*ent.Category, error)
	// Update Update category | 更新版块
	Update(ctx context.Context, id int, updateFn func(*ent.CategoryUpdateOne) *ent.CategoryUpdateOne) (*ent.Category, error)
	// Count Get total category count | 获取版块总数
	Count(ctx context.Context) (int, error)
	// CountWithCondition Count categories with condition | 条件统计版块数
	CountWithCondition(ctx context.Context, conditionFunc func(*ent.CategoryQuery) *ent.CategoryQuery) (int, error)
	// ListWithCondition List categories with condition | 条件查询版块列表
	ListWithCondition(ctx context.Context, conditionFunc func(*ent.CategoryQuery) *ent.CategoryQuery, limit int) ([]*ent.Category, error)
}

// CategoryRepository Category repository implementation | 版块仓储实现
type CategoryRepository struct {
	db *ent.Client
}

// NewCategoryRepository Create category repository instance | 创建版块仓储实例
func NewCategoryRepository(db *ent.Client) ICategoryRepository {
	return &CategoryRepository{db: db}
}

// GetByID Get category by ID | 根据ID获取版块
func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*ent.Category, error) {
	cat, err := r.db.Category.Query().
		Where(category.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("版块不存在")
		}
		return nil, fmt.Errorf("查询版块失败: %w", err)
	}
	return cat, nil
}

// GetByIDs Batch get categories by IDs | 批量根据ID获取版块
func (r *CategoryRepository) GetByIDs(ctx context.Context, ids []int) ([]*ent.Category, error) {
	if len(ids) == 0 {
		return []*ent.Category{}, nil
	}
	categories, err := r.db.Category.Query().
		Where(category.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询版块失败: %w", err)
	}
	return categories, nil
}

// GetByIDsWithFields Batch get categories by IDs with specified fields | 批量根据ID获取版块（指定字段）
func (r *CategoryRepository) GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.Category, error) {
	if len(ids) == 0 {
		return []*ent.Category{}, nil
	}
	query := r.db.Category.Query().Where(category.IDIn(ids...))
	if len(fields) > 0 {
		categories, err := query.Select(fields...).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("批量查询版块失败: %w", err)
		}
		return categories, nil
	}
	categories, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询版块失败: %w", err)
	}
	return categories, nil
}

// GetVisibleCategories Get visible categories for users | 获取用户可见的版块列表
func (r *CategoryRepository) GetVisibleCategories(ctx context.Context) ([]*ent.Category, error) {
	categories, err := r.db.Category.Query().Where(
		category.Or(
			category.StatusEQ(category.StatusNormal),
			category.StatusEQ(category.StatusLoginRequired),
			category.StatusEQ(category.StatusLocked),
		),
	).
		Order(ent.Asc(category.FieldWeight), ent.Desc(category.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取版块列表失败: %w", err)
	}
	return categories, nil
}

// ExistsByID Check if category exists | 检查版块是否存在
func (r *CategoryRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	exists, err := r.db.Category.Query().
		Where(category.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查版块是否存在失败: %w", err)
	}
	return exists, nil
}

// ExistsBySlug Check if category exists by slug | 检查版块标识是否存在
func (r *CategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.db.Category.Query().
		Where(category.SlugEQ(slug)).
		Exist(ctx)
}

// ExistsBySlugExcludeID Check if category exists by slug excluding specific ID | 检查版块标识是否存在（排除指定ID）
func (r *CategoryRepository) ExistsBySlugExcludeID(ctx context.Context, slug string, excludeID int) (bool, error) {
	return r.db.Category.Query().
		Where(
			category.And(
				category.SlugEQ(slug),
				category.IDNEQ(excludeID),
			),
		).
		Exist(ctx)
}

// Create Create category | 创建版块
func (r *CategoryRepository) Create(ctx context.Context, name, slug, description, icon string, weight int, status category.Status) (*ent.Category, error) {
	return r.db.Category.Create().
		SetName(name).
		SetSlug(slug).
		SetDescription(description).
		SetIcon(icon).
		SetWeight(weight).
		SetStatus(status).
		Save(ctx)
}

// Update Update category | 更新版块
func (r *CategoryRepository) Update(ctx context.Context, id int, updateFn func(*ent.CategoryUpdateOne) *ent.CategoryUpdateOne) (*ent.Category, error) {
	update := r.db.Category.UpdateOneID(id)
	return updateFn(update).Save(ctx)
}

// Count Get total category count | 获取版块总数
func (r *CategoryRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.Category.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询版块总数失败: %w", err)
	}
	return count, nil
}

// CountWithCondition Count categories with condition | 条件统计版块数
func (r *CategoryRepository) CountWithCondition(ctx context.Context, conditionFunc func(*ent.CategoryQuery) *ent.CategoryQuery) (int, error) {
	query := r.db.Category.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("条件统计版块数失败: %w", err)
	}
	return count, nil
}

// ListWithCondition List categories with condition | 条件查询版块列表
func (r *CategoryRepository) ListWithCondition(ctx context.Context, conditionFunc func(*ent.CategoryQuery) *ent.CategoryQuery, limit int) ([]*ent.Category, error) {
	query := r.db.Category.Query()
	query = conditionFunc(query)
	if limit > 0 {
		query = query.Limit(limit)
	}
	categories, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("条件查询版块列表失败: %w", err)
	}
	return categories, nil
}
