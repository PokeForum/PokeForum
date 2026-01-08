package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/postaction"
)

// IPostActionRepository PostAction repository interface | 帖子操作仓储接口
type IPostActionRepository interface {
	// Create Create post action | 创建帖子操作记录
	Create(ctx context.Context, userID, postID int, actionType postaction.ActionType) (*ent.PostAction, error)
	// GetByUserAndPost Get post action by user and post | 根据用户和帖子获取操作记录
	GetByUserAndPost(ctx context.Context, userID, postID int, actionType postaction.ActionType) (*ent.PostAction, error)
	// GetUserActions Get user actions for a post | 获取用户对帖子的所有操作
	GetUserActions(ctx context.Context, userID, postID int) ([]*ent.PostAction, error)
	// GetUserActionsForPosts Get user actions for multiple posts | 获取用户对多个帖子的操作
	GetUserActionsForPosts(ctx context.Context, userID int, postIDs []int) ([]*ent.PostAction, error)
	// Delete Delete post action | 删除帖子操作记录
	Delete(ctx context.Context, userID, postID int, actionType postaction.ActionType) error
	// GetFavoritesByUserID Get user's favorite posts | 获取用户收藏的帖子列表
	GetFavoritesByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.PostAction, int, error)
	// CountFavoritesByUserID Count user's favorites | 统计用户收藏数
	CountFavoritesByUserID(ctx context.Context, userID int) (int, error)
}

// PostActionRepository PostAction repository implementation | 帖子操作仓储实现
type PostActionRepository struct {
	db *ent.Client
}

// NewPostActionRepository Create post action repository instance | 创建帖子操作仓储实例
func NewPostActionRepository(db *ent.Client) IPostActionRepository {
	return &PostActionRepository{db: db}
}

// Create Create post action | 创建帖子操作记录
func (r *PostActionRepository) Create(ctx context.Context, userID, postID int, actionType postaction.ActionType) (*ent.PostAction, error) {
	action, err := r.db.PostAction.Create().
		SetUserID(userID).
		SetPostID(postID).
		SetActionType(actionType).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建帖子操作记录失败: %w", err)
	}
	return action, nil
}

// GetByUserAndPost Get post action by user and post | 根据用户和帖子获取操作记录
func (r *PostActionRepository) GetByUserAndPost(ctx context.Context, userID, postID int, actionType postaction.ActionType) (*ent.PostAction, error) {
	action, err := r.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
			postaction.ActionTypeEQ(actionType),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询帖子操作记录失败: %w", err)
	}
	return action, nil
}

// GetUserActions Get user actions for a post | 获取用户对帖子的所有操作
func (r *PostActionRepository) GetUserActions(ctx context.Context, userID, postID int) ([]*ent.PostAction, error) {
	actions, err := r.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询用户帖子操作失败: %w", err)
	}
	return actions, nil
}

// GetUserActionsForPosts Get user actions for multiple posts | 获取用户对多个帖子的操作
func (r *PostActionRepository) GetUserActionsForPosts(ctx context.Context, userID int, postIDs []int) ([]*ent.PostAction, error) {
	if len(postIDs) == 0 {
		return []*ent.PostAction{}, nil
	}
	actions, err := r.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDIn(postIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询用户帖子操作失败: %w", err)
	}
	return actions, nil
}

// Delete Delete post action | 删除帖子操作记录
func (r *PostActionRepository) Delete(ctx context.Context, userID, postID int, actionType postaction.ActionType) error {
	_, err := r.db.PostAction.Delete().
		Where(
			postaction.UserIDEQ(userID),
			postaction.PostIDEQ(postID),
			postaction.ActionTypeEQ(actionType),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除帖子操作记录失败: %w", err)
	}
	return nil
}

// GetFavoritesByUserID Get user's favorite posts | 获取用户收藏的帖子列表
func (r *PostActionRepository) GetFavoritesByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.PostAction, int, error) {
	query := r.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		).
		Order(ent.Desc(postaction.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户收藏总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	favorites, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户收藏列表失败: %w", err)
	}

	return favorites, total, nil
}

// CountFavoritesByUserID Count user's favorites | 统计用户收藏数
func (r *PostActionRepository) CountFavoritesByUserID(ctx context.Context, userID int) (int, error) {
	count, err := r.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户收藏数失败: %w", err)
	}
	return count, nil
}
