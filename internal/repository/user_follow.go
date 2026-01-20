package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/userfollow"
)

// IUserFollowRepository User follow repository interface | 用户关注仓库接口
type IUserFollowRepository interface {
	// Create Create follow relationship | 创建关注关系
	Create(ctx context.Context, followerID, followingID int) (*ent.UserFollow, error)
	// Delete Delete follow relationship | 删除关注关系
	Delete(ctx context.Context, followerID, followingID int) (int, error)
	// Exists Check if follow relationship exists | 检查关注关系是否存在
	Exists(ctx context.Context, followerID, followingID int) (bool, error)
	// GetFollowers Get user's followers list | 获取用户粉丝列表
	GetFollowers(ctx context.Context, userID int, page, pageSize int) ([]*ent.UserFollow, int, error)
	// GetFollowing Get user's following list | 获取用户关注列表
	GetFollowing(ctx context.Context, userID int, page, pageSize int) ([]*ent.UserFollow, int, error)
	// CountFollowers Count user's followers | 统计用户粉丝数
	CountFollowers(ctx context.Context, userID int) (int, error)
	// CountFollowing Count user's following | 统计用户关注数
	CountFollowing(ctx context.Context, userID int) (int, error)
	// IsMutualFollow Check if two users follow each other | 检查是否互相关注
	IsMutualFollow(ctx context.Context, userIDA, userIDB int) (bool, error)
	// GetFollowStatus Get follow status between two users | 获取两个用户之间的关注状态
	GetFollowStatus(ctx context.Context, followerID, followingID int) (*ent.UserFollow, error)
}

// UserFollowRepository User follow repository implementation | 用户关注仓库实现
type UserFollowRepository struct {
	db *ent.Client
}

// NewUserFollowRepository Create user follow repository instance | 创建用户关注仓库实例
func NewUserFollowRepository(db *ent.Client) IUserFollowRepository {
	return &UserFollowRepository{db: db}
}

// Create Create follow relationship | 创建关注关系
func (r *UserFollowRepository) Create(ctx context.Context, followerID, followingID int) (*ent.UserFollow, error) {
	follow, err := r.db.UserFollow.Create().
		SetFollowerID(followerID).
		SetFollowingID(followingID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建关注关系失败: %w", err)
	}
	return follow, nil
}

// Delete Delete follow relationship | 删除关注关系
func (r *UserFollowRepository) Delete(ctx context.Context, followerID, followingID int) (int, error) {
	affected, err := r.db.UserFollow.Delete().
		Where(
			userfollow.FollowerIDEQ(followerID),
			userfollow.FollowingIDEQ(followingID),
		).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("删除关注关系失败: %w", err)
	}
	return affected, nil
}

// Exists Check if follow relationship exists | 检查关注关系是否存在
func (r *UserFollowRepository) Exists(ctx context.Context, followerID, followingID int) (bool, error) {
	exists, err := r.db.UserFollow.Query().
		Where(
			userfollow.FollowerIDEQ(followerID),
			userfollow.FollowingIDEQ(followingID),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("检查关注关系失败: %w", err)
	}
	return exists, nil
}

// GetFollowers Get user's followers list | 获取用户粉丝列表
func (r *UserFollowRepository) GetFollowers(ctx context.Context, userID int, page, pageSize int) ([]*ent.UserFollow, int, error) {
	query := r.db.UserFollow.Query().
		Where(userfollow.FollowingIDEQ(userID)).
		Order(ent.Desc(userfollow.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取粉丝总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	followers, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取粉丝列表失败: %w", err)
	}

	return followers, total, nil
}

// GetFollowing Get user's following list | 获取用户关注列表
func (r *UserFollowRepository) GetFollowing(ctx context.Context, userID int, page, pageSize int) ([]*ent.UserFollow, int, error) {
	query := r.db.UserFollow.Query().
		Where(userfollow.FollowerIDEQ(userID)).
		Order(ent.Desc(userfollow.FieldCreatedAt))

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取关注总数失败: %w", err)
	}

	// Apply pagination | 应用分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Execute query | 执行查询
	following, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取关注列表失败: %w", err)
	}

	return following, total, nil
}

// CountFollowers Count user's followers | 统计用户粉丝数
func (r *UserFollowRepository) CountFollowers(ctx context.Context, userID int) (int, error) {
	count, err := r.db.UserFollow.Query().
		Where(userfollow.FollowingIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计粉丝数失败: %w", err)
	}
	return count, nil
}

// CountFollowing Count user's following | 统计用户关注数
func (r *UserFollowRepository) CountFollowing(ctx context.Context, userID int) (int, error) {
	count, err := r.db.UserFollow.Query().
		Where(userfollow.FollowerIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计关注数失败: %w", err)
	}
	return count, nil
}

// IsMutualFollow Check if two users follow each other | 检查是否互相关注
func (r *UserFollowRepository) IsMutualFollow(ctx context.Context, userIDA, userIDB int) (bool, error) {
	// 查询 A 关注 B
	aFollowsB, err := r.Exists(ctx, userIDA, userIDB)
	if err != nil {
		return false, err
	}

	// 查询 B 关注 A
	bFollowsA, err := r.Exists(ctx, userIDB, userIDA)
	if err != nil {
		return false, err
	}

	return aFollowsB && bFollowsA, nil
}

// GetFollowStatus Get follow status between two users | 获取两个用户之间的关注状态
func (r *UserFollowRepository) GetFollowStatus(ctx context.Context, followerID, followingID int) (*ent.UserFollow, error) {
	follow, err := r.db.UserFollow.Query().
		Where(
			userfollow.FollowerIDEQ(followerID),
			userfollow.FollowingIDEQ(followingID),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取关注状态失败: %w", err)
	}
	return follow, nil
}
