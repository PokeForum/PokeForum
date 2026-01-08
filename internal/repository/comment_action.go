package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/commentaction"
)

// ICommentActionRepository CommentAction repository interface | 评论操作仓储接口
type ICommentActionRepository interface {
	// Create Create comment action | 创建评论操作记录
	Create(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) (*ent.CommentAction, error)
	// GetByUserAndComment Get comment action by user and comment | 根据用户和评论获取操作记录
	GetByUserAndComment(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) (*ent.CommentAction, error)
	// GetUserActionsForComments Get user actions for multiple comments | 获取用户对多个评论的操作
	GetUserActionsForComments(ctx context.Context, userID int, commentIDs []int) ([]*ent.CommentAction, error)
	// Delete Delete comment action | 删除评论操作记录
	Delete(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) error
}

// CommentActionRepository CommentAction repository implementation | 评论操作仓储实现
type CommentActionRepository struct {
	db *ent.Client
}

// NewCommentActionRepository Create comment action repository instance | 创建评论操作仓储实例
func NewCommentActionRepository(db *ent.Client) ICommentActionRepository {
	return &CommentActionRepository{db: db}
}

// Create Create comment action | 创建评论操作记录
func (r *CommentActionRepository) Create(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) (*ent.CommentAction, error) {
	action, err := r.db.CommentAction.Create().
		SetUserID(userID).
		SetCommentID(commentID).
		SetActionType(actionType).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建评论操作记录失败: %w", err)
	}
	return action, nil
}

// GetByUserAndComment Get comment action by user and comment | 根据用户和评论获取操作记录
func (r *CommentActionRepository) GetByUserAndComment(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) (*ent.CommentAction, error) {
	action, err := r.db.CommentAction.Query().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
			commentaction.ActionTypeEQ(actionType),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询评论操作记录失败: %w", err)
	}
	return action, nil
}

// GetUserActionsForComments Get user actions for multiple comments | 获取用户对多个评论的操作
func (r *CommentActionRepository) GetUserActionsForComments(ctx context.Context, userID int, commentIDs []int) ([]*ent.CommentAction, error) {
	if len(commentIDs) == 0 {
		return []*ent.CommentAction{}, nil
	}
	actions, err := r.db.CommentAction.Query().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDIn(commentIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量查询用户评论操作失败: %w", err)
	}
	return actions, nil
}

// Delete Delete comment action | 删除评论操作记录
func (r *CommentActionRepository) Delete(ctx context.Context, userID, commentID int, actionType commentaction.ActionType) error {
	_, err := r.db.CommentAction.Delete().
		Where(
			commentaction.UserIDEQ(userID),
			commentaction.CommentIDEQ(commentID),
			commentaction.ActionTypeEQ(actionType),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除评论操作记录失败: %w", err)
	}
	return nil
}
