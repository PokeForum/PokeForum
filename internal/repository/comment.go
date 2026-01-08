package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
)

// ICommentRepository Comment repository interface | 评论仓储接口
type ICommentRepository interface {
	// Create Create comment | 创建评论
	Create(ctx context.Context, userID, postID int, content, commenterIP, deviceInfo string, parentID, replyToUserID *int) (*ent.Comment, error)
	// GetByID Get comment by ID | 根据ID获取评论
	GetByID(ctx context.Context, id int) (*ent.Comment, error)
	// GetByIDAndUserID Get comment by ID and user ID | 根据ID和用户ID获取评论
	GetByIDAndUserID(ctx context.Context, id, userID int) (*ent.Comment, error)
	// Update Update comment | 更新评论
	Update(ctx context.Context, id int, updateFunc func(*ent.CommentUpdateOne) *ent.CommentUpdateOne) (*ent.Comment, error)
	// UpdateContent Update comment content | 更新评论内容
	UpdateContent(ctx context.Context, id int, content string) (*ent.Comment, error)
	// List Get comment list by post ID | 根据帖子ID获取评论列表
	List(ctx context.Context, postID int, page, pageSize int) ([]*ent.Comment, int, error)
	// CountByUserID Count comments by user ID | 统计用户评论数
	CountByUserID(ctx context.Context, userID int) (int, error)
	// GetByUserID Get comments by user ID | 根据用户ID获取评论列表
	GetByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Comment, int, error)
	// GetByIDs Batch get comments by IDs | 批量根据ID获取评论
	GetByIDs(ctx context.Context, ids []int) ([]*ent.Comment, error)
	// ExistsByID Check if comment exists by ID | 检查评论是否存在
	ExistsByID(ctx context.Context, id int) (bool, error)
	// Delete Delete comment by ID | 根据ID删除评论
	Delete(ctx context.Context, id int) error
	// Count Get total comment count | 获取评论总数
	Count(ctx context.Context) (int, error)
	// CountWithCondition Count comments with condition | 条件统计评论数
	CountWithCondition(ctx context.Context, conditionFunc func(*ent.CommentQuery) *ent.CommentQuery) (int, error)
	// ListWithCondition List comments with condition | 条件查询评论列表
	ListWithCondition(ctx context.Context, conditionFunc func(*ent.CommentQuery) *ent.CommentQuery, limit int) ([]*ent.Comment, error)
	// GetAll Get all comments | 获取所有评论
	GetAll(ctx context.Context) ([]*ent.Comment, error)
	// CountByPostID Count comments by post ID | 统计帖子评论数
	CountByPostID(ctx context.Context, postID int) (int, error)
}

// CommentRepository Comment repository implementation | 评论仓储实现
type CommentRepository struct {
	db *ent.Client
}

// NewCommentRepository Create comment repository instance | 创建评论仓储实例
func NewCommentRepository(db *ent.Client) ICommentRepository {
	return &CommentRepository{db: db}
}

// Create Create comment | 创建评论
func (r *CommentRepository) Create(ctx context.Context, userID, postID int, content, commenterIP, deviceInfo string, parentID, replyToUserID *int) (*ent.Comment, error) {
	creator := r.db.Comment.Create().
		SetUserID(userID).
		SetPostID(postID).
		SetContent(content).
		SetCommenterIP(commenterIP).
		SetDeviceInfo(deviceInfo)

	if parentID != nil {
		creator = creator.SetNillableParentID(parentID)
	}
	if replyToUserID != nil {
		creator = creator.SetNillableReplyToUserID(replyToUserID)
	}

	c, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}
	return c, nil
}

// GetByID Get comment by ID | 根据ID获取评论
func (r *CommentRepository) GetByID(ctx context.Context, id int) (*ent.Comment, error) {
	c, err := r.db.Comment.Query().
		Where(comment.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("评论不存在")
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return c, nil
}

// GetByIDAndUserID Get comment by ID and user ID | 根据ID和用户ID获取评论
func (r *CommentRepository) GetByIDAndUserID(ctx context.Context, id, userID int) (*ent.Comment, error) {
	c, err := r.db.Comment.Query().
		Where(comment.IDEQ(id), comment.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("评论不存在或无权限")
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return c, nil
}

// Update Update comment | 更新评论
func (r *CommentRepository) Update(ctx context.Context, id int, updateFunc func(*ent.CommentUpdateOne) *ent.CommentUpdateOne) (*ent.Comment, error) {
	updater := r.db.Comment.UpdateOneID(id)
	updater = updateFunc(updater)
	c, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}
	return c, nil
}

// UpdateContent Update comment content | 更新评论内容
func (r *CommentRepository) UpdateContent(ctx context.Context, id int, content string) (*ent.Comment, error) {
	c, err := r.db.Comment.UpdateOneID(id).
		SetContent(content).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新评论内容失败: %w", err)
	}
	return c, nil
}

// List Get comment list by post ID | 根据帖子ID获取评论列表
func (r *CommentRepository) List(ctx context.Context, postID int, page, pageSize int) ([]*ent.Comment, int, error) {
	query := r.db.Comment.Query().
		Where(comment.PostIDEQ(postID)).
		Order(ent.Asc(comment.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	comments, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取评论列表失败: %w", err)
	}

	return comments, total, nil
}

// CountByUserID Count comments by user ID | 统计用户评论数
func (r *CommentRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	count, err := r.db.Comment.Query().
		Where(comment.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户评论数失败: %w", err)
	}
	return count, nil
}

// GetByUserID Get comments by user ID | 根据用户ID获取评论列表
func (r *CommentRepository) GetByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Comment, int, error) {
	query := r.db.Comment.Query().
		Where(comment.UserIDEQ(userID)).
		Order(ent.Desc(comment.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户评论总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	comments, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户评论列表失败: %w", err)
	}

	return comments, total, nil
}

// GetByIDs Batch get comments by IDs | 批量根据ID获取评论
func (r *CommentRepository) GetByIDs(ctx context.Context, ids []int) ([]*ent.Comment, error) {
	if len(ids) == 0 {
		return []*ent.Comment{}, nil
	}
	comments, err := r.db.Comment.Query().
		Where(comment.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询评论失败: %w", err)
	}
	return comments, nil
}

// ExistsByID Check if comment exists by ID | 检查评论是否存在
func (r *CommentRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	exists, err := r.db.Comment.Query().
		Where(comment.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查评论是否存在失败: %w", err)
	}
	return exists, nil
}

// Delete Delete comment by ID | 根据ID删除评论
func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	err := r.db.Comment.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}
	return nil
}

// Count Get total comment count | 获取评论总数
func (r *CommentRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.Comment.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询评论总数失败: %w", err)
	}
	return count, nil
}

// CountWithCondition Count comments with condition | 条件统计评论数
func (r *CommentRepository) CountWithCondition(ctx context.Context, conditionFunc func(*ent.CommentQuery) *ent.CommentQuery) (int, error) {
	query := r.db.Comment.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("条件统计评论数失败: %w", err)
	}
	return count, nil
}

// ListWithCondition List comments with condition | 条件查询评论列表
func (r *CommentRepository) ListWithCondition(ctx context.Context, conditionFunc func(*ent.CommentQuery) *ent.CommentQuery, limit int) ([]*ent.Comment, error) {
	query := r.db.Comment.Query()
	query = conditionFunc(query)
	if limit > 0 {
		query = query.Limit(limit)
	}
	comments, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("条件查询评论列表失败: %w", err)
	}
	return comments, nil
}

// GetAll Get all comments | 获取所有评论
func (r *CommentRepository) GetAll(ctx context.Context) ([]*ent.Comment, error) {
	comments, err := r.db.Comment.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取所有评论失败: %w", err)
	}
	return comments, nil
}

// CountByPostID Count comments by post ID | 统计帖子评论数
func (r *CommentRepository) CountByPostID(ctx context.Context, postID int) (int, error) {
	count, err := r.db.Comment.Query().
		Where(comment.PostIDEQ(postID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计帖子评论数失败: %w", err)
	}
	return count, nil
}
