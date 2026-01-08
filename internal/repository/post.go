package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/post"
)

// IPostRepository Post repository interface | 帖子仓储接口
type IPostRepository interface {
	// Create Create post | 创建帖子
	Create(ctx context.Context, userID, categoryID int, title, content string, readPermission string, status post.Status) (*ent.Post, error)
	// GetByID Get post by ID | 根据ID获取帖子
	GetByID(ctx context.Context, id int) (*ent.Post, error)
	// GetByIDWithStatus Get post by ID with status filter | 根据ID和状态获取帖子
	GetByIDWithStatus(ctx context.Context, id int, status post.Status) (*ent.Post, error)
	// Update Update post | 更新帖子
	Update(ctx context.Context, id int, updateFunc func(*ent.PostUpdateOne) *ent.PostUpdateOne) (*ent.Post, error)
	// UpdateStatus Update post status | 更新帖子状态
	UpdateStatus(ctx context.Context, id int, status post.Status) error
	// List Get post list with pagination | 获取帖子列表（分页）
	List(ctx context.Context, opts ListPostOptions) ([]*ent.Post, int, error)
	// CountByUserID Count posts by user ID | 统计用户发帖数
	CountByUserID(ctx context.Context, userID int) (int, error)
	// CountByUserIDWithStatus Count posts by user ID and status | 统计用户指定状态的发帖数
	CountByUserIDWithStatus(ctx context.Context, userID int, status post.Status) (int, error)
	// CountByCategoryID Count posts by category ID | 统计版块帖子数
	CountByCategoryID(ctx context.Context, categoryID int) (int, error)
	// GetByUserID Get posts by user ID | 根据用户ID获取帖子列表
	GetByUserID(ctx context.Context, userID int, opts ListPostOptions) ([]*ent.Post, int, error)
	// GetByIDs Batch get posts by IDs | 批量根据ID获取帖子
	GetByIDs(ctx context.Context, ids []int) ([]*ent.Post, error)
	// GetByIDsWithFields Batch get posts by IDs with specific fields | 批量根据ID获取帖子（指定字段）
	GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.Post, error)
	// ExistsByID Check if post exists by ID | 检查帖子是否存在
	ExistsByID(ctx context.Context, id int) (bool, error)
	// Count Get total post count | 获取帖子总数
	Count(ctx context.Context) (int, error)
	// CountWithCondition Count posts with condition | 条件统计帖子数
	CountWithCondition(ctx context.Context, conditionFunc func(*ent.PostQuery) *ent.PostQuery) (int, error)
	// ListWithCondition List posts with condition | 条件查询帖子列表
	ListWithCondition(ctx context.Context, conditionFunc func(*ent.PostQuery) *ent.PostQuery, limit int) ([]*ent.Post, error)
	// GetAll Get all posts | 获取所有帖子
	GetAll(ctx context.Context) ([]*ent.Post, error)
}

// ListPostOptions Post list query options | 帖子列表查询选项
type ListPostOptions struct {
	CategoryID int         // Category ID filter | 版块ID筛选
	Status     post.Status // Status filter | 状态筛选
	SortBy     string      // Sort field: latest, hot, essence | 排序字段
	Page       int         // Page number | 页码
	PageSize   int         // Page size | 每页数量
}

// PostRepository Post repository implementation | 帖子仓储实现
type PostRepository struct {
	db *ent.Client
}

// NewPostRepository Create post repository instance | 创建帖子仓储实例
func NewPostRepository(db *ent.Client) IPostRepository {
	return &PostRepository{db: db}
}

// Create Create post | 创建帖子
func (r *PostRepository) Create(ctx context.Context, userID, categoryID int, title, content string, readPermission string, status post.Status) (*ent.Post, error) {
	p, err := r.db.Post.Create().
		SetUserID(userID).
		SetCategoryID(categoryID).
		SetTitle(title).
		SetContent(content).
		SetReadPermission(readPermission).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建帖子失败: %w", err)
	}
	return p, nil
}

// GetByID Get post by ID | 根据ID获取帖子
func (r *PostRepository) GetByID(ctx context.Context, id int) (*ent.Post, error) {
	p, err := r.db.Post.Query().
		Where(post.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("帖子不存在")
		}
		return nil, fmt.Errorf("查询帖子失败: %w", err)
	}
	return p, nil
}

// GetByIDWithStatus Get post by ID with status filter | 根据ID和状态获取帖子
func (r *PostRepository) GetByIDWithStatus(ctx context.Context, id int, status post.Status) (*ent.Post, error) {
	p, err := r.db.Post.Query().
		Where(post.IDEQ(id), post.StatusEQ(status)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("帖子不存在或已删除")
		}
		return nil, fmt.Errorf("查询帖子失败: %w", err)
	}
	return p, nil
}

// Update Update post | 更新帖子
func (r *PostRepository) Update(ctx context.Context, id int, updateFunc func(*ent.PostUpdateOne) *ent.PostUpdateOne) (*ent.Post, error) {
	updater := r.db.Post.UpdateOneID(id)
	updater = updateFunc(updater)
	p, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}
	return p, nil
}

// UpdateStatus Update post status | 更新帖子状态
func (r *PostRepository) UpdateStatus(ctx context.Context, id int, status post.Status) error {
	_, err := r.db.Post.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新帖子状态失败: %w", err)
	}
	return nil
}

// List Get post list with pagination | 获取帖子列表（分页）
func (r *PostRepository) List(ctx context.Context, opts ListPostOptions) ([]*ent.Post, int, error) {
	query := r.db.Post.Query()

	// Apply category filter | 应用版块筛选
	if opts.CategoryID > 0 {
		query = query.Where(post.CategoryID(opts.CategoryID))
	}

	// Apply status filter | 应用状态筛选
	if opts.Status != "" {
		query = query.Where(post.StatusEQ(opts.Status))
	}

	// Apply sorting | 应用排序
	switch opts.SortBy {
	case "hot":
		query = query.Order(ent.Desc(post.FieldViewCount), ent.Desc(post.FieldLikeCount))
	case "essence":
		query = query.Where(post.IsEssence(true)).Order(ent.Desc(post.FieldCreatedAt))
	default: // latest
		query = query.Order(ent.Desc(post.FieldCreatedAt))
	}

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取帖子总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	// Execute query | 执行查询
	posts, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取帖子列表失败: %w", err)
	}

	return posts, total, nil
}

// CountByUserID Count posts by user ID | 统计用户发帖数
func (r *PostRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	count, err := r.db.Post.Query().
		Where(post.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户发帖数失败: %w", err)
	}
	return count, nil
}

// CountByUserIDWithStatus Count posts by user ID and status | 统计用户指定状态的发帖数
func (r *PostRepository) CountByUserIDWithStatus(ctx context.Context, userID int, status post.Status) (int, error) {
	count, err := r.db.Post.Query().
		Where(post.UserIDEQ(userID), post.StatusEQ(status)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户发帖数失败: %w", err)
	}
	return count, nil
}

// CountByCategoryID Count posts by category ID | 统计版块帖子数
func (r *PostRepository) CountByCategoryID(ctx context.Context, categoryID int) (int, error) {
	count, err := r.db.Post.Query().
		Where(post.CategoryIDEQ(categoryID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计版块帖子数失败: %w", err)
	}
	return count, nil
}

// GetByUserID Get posts by user ID | 根据用户ID获取帖子列表
func (r *PostRepository) GetByUserID(ctx context.Context, userID int, opts ListPostOptions) ([]*ent.Post, int, error) {
	query := r.db.Post.Query().Where(post.UserIDEQ(userID))

	// Apply status filter | 应用状态筛选
	if opts.Status != "" {
		query = query.Where(post.StatusEQ(opts.Status))
	}

	// Apply sorting | 应用排序
	query = query.Order(ent.Desc(post.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户帖子总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	// Execute query | 执行查询
	posts, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户帖子列表失败: %w", err)
	}

	return posts, total, nil
}

// GetByIDs Batch get posts by IDs | 批量根据ID获取帖子
func (r *PostRepository) GetByIDs(ctx context.Context, ids []int) ([]*ent.Post, error) {
	if len(ids) == 0 {
		return []*ent.Post{}, nil
	}
	posts, err := r.db.Post.Query().
		Where(post.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询帖子失败: %w", err)
	}
	return posts, nil
}

// GetByIDsWithFields Batch get posts by IDs with specific fields | 批量根据ID获取帖子（指定字段）
func (r *PostRepository) GetByIDsWithFields(ctx context.Context, ids []int, fields []string) ([]*ent.Post, error) {
	if len(ids) == 0 {
		return []*ent.Post{}, nil
	}
	posts, err := r.db.Post.Query().
		Where(post.IDIn(ids...)).
		Select(fields...).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询帖子失败: %w", err)
	}
	return posts, nil
}

// ExistsByID Check if post exists by ID | 检查帖子是否存在
func (r *PostRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	exists, err := r.db.Post.Query().
		Where(post.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查帖子是否存在失败: %w", err)
	}
	return exists, nil
}

// Count Get total post count | 获取帖子总数
func (r *PostRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.Post.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询帖子总数失败: %w", err)
	}
	return count, nil
}

// CountWithCondition Count posts with condition | 条件统计帖子数
func (r *PostRepository) CountWithCondition(ctx context.Context, conditionFunc func(*ent.PostQuery) *ent.PostQuery) (int, error) {
	query := r.db.Post.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("条件统计帖子数失败: %w", err)
	}
	return count, nil
}

// ListWithCondition List posts with condition | 条件查询帖子列表
func (r *PostRepository) ListWithCondition(ctx context.Context, conditionFunc func(*ent.PostQuery) *ent.PostQuery, limit int) ([]*ent.Post, error) {
	query := r.db.Post.Query()
	query = conditionFunc(query)
	if limit > 0 {
		query = query.Limit(limit)
	}
	posts, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("条件查询帖子列表失败: %w", err)
	}
	return posts, nil
}

// GetAll Get all posts | 获取所有帖子
func (r *PostRepository) GetAll(ctx context.Context) ([]*ent.Post, error) {
	posts, err := r.db.Post.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取所有帖子失败: %w", err)
	}
	return posts, nil
}
