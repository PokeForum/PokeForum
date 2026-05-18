package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IUserFollowService User follow service interface | 用户关注服务接口
type IUserFollowService interface {
	// FollowUser Follow a user | 关注用户
	FollowUser(ctx context.Context, followerID int, req schema.UserFollowRequest) (*schema.UserFollowResponse, error)
	// UnfollowUser Unfollow a user | 取消关注用户
	UnfollowUser(ctx context.Context, followerID int, req schema.UserUnfollowRequest) (*schema.UserUnfollowResponse, error)
	// GetFollowers Get user's followers list | 获取用户粉丝列表
	GetFollowers(ctx context.Context, currentUserID int, req schema.UserFollowersRequest) (*schema.UserFollowersResponse, error)
	// GetFollowing Get user's following list | 获取用户关注列表
	GetFollowing(ctx context.Context, currentUserID int, req schema.UserFollowingRequest) (*schema.UserFollowingResponse, error)
	// GetFollowStatus Get follow status between two users | 获取两个用户之间的关注状态
	GetFollowStatus(ctx context.Context, followerID, followingID int) (*schema.UserFollowStatusResponse, error)
	// GetFollowCounts Get user's follow counts | 获取用户关注数统计
	GetFollowCounts(ctx context.Context, userID int) (followersCount, followingCount int, err error)
}

// UserFollowService User follow service implementation | 用户关注服务实现
type UserFollowService struct {
	followRepo    repository.IUserFollowRepository
	userRepo      repository.IUserRepository
	blacklistRepo repository.IBlacklistRepository
	cache         cache.ICacheService
	logger        *zap.Logger
}

// NewUserFollowService Create user follow service instance | 创建用户关注服务实例
func NewUserFollowService(
	followRepo repository.IUserFollowRepository,
	userRepo repository.IUserRepository,
	blacklistRepo repository.IBlacklistRepository,
	cacheService cache.ICacheService,
	logger *zap.Logger,
) IUserFollowService {
	return &UserFollowService{
		followRepo:    followRepo,
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		cache:         cacheService,
		logger:        logger,
	}
}

// FollowUser Follow a user | 关注用户
func (s *UserFollowService) FollowUser(ctx context.Context, followerID int, req schema.UserFollowRequest) (*schema.UserFollowResponse, error) {
	s.logger.Info("关注用户",

		tracing.WithTraceIDField(ctx), zap.Int("follower_id", followerID),
		zap.Int("following_id", req.FollowingID),
	)

	// Business validation | 业务验证
	// Cannot follow yourself | 不能关注自己
	if followerID == req.FollowingID {
		return nil, errors.New("不能关注自己 | Cannot follow yourself")
	}

	// Check if target user exists | 检查被关注用户是否存在
	targetUser, err := s.userRepo.GetByID(ctx, req.FollowingID)
	if err != nil {
		s.logger.Error("获取被关注用户信息失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("被关注用户不存在 | Target user does not exist")
	}

	// Check if already following | 检查是否已经关注
	exists, err := s.followRepo.Exists(ctx, followerID, req.FollowingID)
	if err != nil {
		s.logger.Error("检查关注关系失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("检查关注关系失败 | Failed to check follow relationship")
	}
	if exists {
		return nil, errors.New("已经关注该用户 | Already following this user")
	}

	// Check if in blacklist | 检查是否在黑名单中
	isBlocked, err := s.blacklistRepo.Exists(ctx, req.FollowingID, followerID)
	if err != nil {
		s.logger.Warn("检查黑名单失败", tracing.WithTraceIDField(ctx), zap.Error(err))
	}
	if isBlocked {
		return nil, errors.New("对方已将您加入黑名单 | You have been blocked by the target user")
	}

	// Create follow relationship | 创建关注关系
	_, err = s.followRepo.Create(ctx, followerID, req.FollowingID)
	if err != nil {
		s.logger.Error("创建关注关系失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("关注失败 | Failed to follow user")
	}

	// 增加被关注者的粉丝数计数器
	s.incrFollowersCount(ctx, req.FollowingID)
	// 增加关注者的关注数计数器
	s.incrFollowingCount(ctx, followerID)

	s.logger.Info("关注成功",

		tracing.WithTraceIDField(ctx), zap.Int("follower_id", followerID),
		zap.Int("following_id", req.FollowingID),
		zap.String("target_username", targetUser.Username),
	)

	return &schema.UserFollowResponse{
		Success:     true,
		Message:     "关注成功 | Followed successfully",
		FollowingID: req.FollowingID,
	}, nil
}

// UnfollowUser Unfollow a user | 取消关注用户
func (s *UserFollowService) UnfollowUser(ctx context.Context, followerID int, req schema.UserUnfollowRequest) (*schema.UserUnfollowResponse, error) {
	s.logger.Info("取消关注",

		tracing.WithTraceIDField(ctx), zap.Int("follower_id", followerID),
		zap.Int("following_id", req.FollowingID),
	)

	// Check if following exists | 检查关注关系是否存在
	exists, err := s.followRepo.Exists(ctx, followerID, req.FollowingID)
	if err != nil {
		s.logger.Error("检查关注关系失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("检查关注关系失败 | Failed to check follow relationship")
	}
	if !exists {
		return nil, errors.New("未关注该用户 | Not following this user")
	}

	// Delete follow relationship | 删除关注关系
	_, err = s.followRepo.Delete(ctx, followerID, req.FollowingID)
	if err != nil {
		s.logger.Error("删除关注关系失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("取消关注失败 | Failed to unfollow user")
	}

	// 减少被关注者的粉丝数计数器
	s.decrFollowersCount(ctx, req.FollowingID)
	// 减少关注者的关注数计数器
	s.decrFollowingCount(ctx, followerID)

	s.logger.Info("取消关注成功",

		tracing.WithTraceIDField(ctx), zap.Int("follower_id", followerID),
		zap.Int("following_id", req.FollowingID),
	)

	return &schema.UserUnfollowResponse{
		Success:     true,
		Message:     "取消关注成功 | Unfollowed successfully",
		FollowingID: req.FollowingID,
	}, nil
}

// GetFollowers Get user's followers list | 获取用户粉丝列表
func (s *UserFollowService) GetFollowers(ctx context.Context, currentUserID int, req schema.UserFollowersRequest) (*schema.UserFollowersResponse, error) {
	s.logger.Info("获取用户粉丝列表",

		tracing.WithTraceIDField(ctx), zap.Int("user_id", req.UserID),
		zap.Int("current_user_id", currentUserID),
		zap.Int("page", req.Page),
	)

	// Query follower relationships | 查询粉丝关系
	follows, total, err := s.followRepo.GetFollowers(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		s.logger.Error("获取粉丝列表失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}

	// Batch query user information | 批量查询用户信息
	followerIDs := make([]int, len(follows))
	for i, f := range follows {
		followerIDs[i] = f.FollowerID
	}

	users, err := s.userRepo.GetByIDsWithFields(ctx, followerIDs, []string{
		user.FieldID,
		user.FieldUsername,
		user.FieldAvatar,
		user.FieldSignature,
	})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", tracing.WithTraceIDField(ctx), zap.Error(err))
	}

	userMap := make(map[int]*ent.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Check follow status between current user and followers | 检查当前用户与这些粉丝的关注关系
	isFollowingMap := make(map[int]bool)
	if currentUserID > 0 {
		for _, followerID := range followerIDs {
			isFollowing, err := s.followRepo.Exists(ctx, currentUserID, followerID)
			if err == nil {
				isFollowingMap[followerID] = isFollowing
			}
		}
	}

	// Build response data | 构建响应数据
	list := make([]schema.UserFollowItem, 0, len(follows))
	for _, f := range follows {
		userData := userMap[f.FollowerID]
		if userData == nil {
			continue
		}

		item := schema.UserFollowItem{
			UserID:      userData.ID,
			Username:    userData.Username,
			Avatar:      userData.Avatar,
			Signature:   userData.Signature,
			FollowedAt:  f.CreatedAt.Format(time_tools.DateTimeFormat),
			IsFollowing: isFollowingMap[f.FollowerID],
		}

		// Check if mutual follow | 检查是否互相关注
		if currentUserID > 0 {
			isMutual, err := s.followRepo.IsMutualFollow(ctx, currentUserID, f.FollowerID)
			if err == nil {
				item.IsMutual = isMutual
			}
		}

		list = append(list, item)
	}

	return &schema.UserFollowersResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetFollowing Get user's following list | 获取用户关注列表
func (s *UserFollowService) GetFollowing(ctx context.Context, currentUserID int, req schema.UserFollowingRequest) (*schema.UserFollowingResponse, error) {
	s.logger.Info("获取用户关注列表",

		tracing.WithTraceIDField(ctx), zap.Int("user_id", req.UserID),
		zap.Int("current_user_id", currentUserID),
		zap.Int("page", req.Page),
	)

	// Query following relationships | 查询关注关系
	follows, total, err := s.followRepo.GetFollowing(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		s.logger.Error("获取关注列表失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}

	// Batch query user information | 批量查询用户信息
	followingIDs := make([]int, len(follows))
	for i, f := range follows {
		followingIDs[i] = f.FollowingID
	}

	users, err := s.userRepo.GetByIDsWithFields(ctx, followingIDs, []string{
		user.FieldID,
		user.FieldUsername,
		user.FieldAvatar,
		user.FieldSignature,
	})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", tracing.WithTraceIDField(ctx), zap.Error(err))
	}

	userMap := make(map[int]*ent.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Check follow status | 检查关注状态
	isFollowingMap := make(map[int]bool)
	for _, followingID := range followingIDs {
		// If viewing own list, all are following | 如果查看自己的列表，全部都是已关注
		if currentUserID == req.UserID {
			isFollowingMap[followingID] = true
		} else if currentUserID > 0 {
			isFollowing, err := s.followRepo.Exists(ctx, currentUserID, followingID)
			if err == nil {
				isFollowingMap[followingID] = isFollowing
			}
		}
	}

	// Build response data | 构建响应数据
	list := make([]schema.UserFollowItem, 0, len(follows))
	for _, f := range follows {
		userData := userMap[f.FollowingID]
		if userData == nil {
			continue
		}

		item := schema.UserFollowItem{
			UserID:      userData.ID,
			Username:    userData.Username,
			Avatar:      userData.Avatar,
			Signature:   userData.Signature,
			FollowedAt:  f.CreatedAt.Format(time_tools.DateTimeFormat),
			IsFollowing: isFollowingMap[f.FollowingID],
		}

		// Check if mutual follow | 检查是否互相关注
		if currentUserID > 0 {
			isMutual, err := s.followRepo.IsMutualFollow(ctx, currentUserID, f.FollowingID)
			if err == nil {
				item.IsMutual = isMutual
			}
		}

		list = append(list, item)
	}

	return &schema.UserFollowingResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetFollowStatus Get follow status between two users | 获取两个用户之间的关注状态
func (s *UserFollowService) GetFollowStatus(ctx context.Context, followerID, followingID int) (*schema.UserFollowStatusResponse, error) {
	// Check if follower follows following | 检查 follower 是否关注 following
	isFollowing, err := s.followRepo.Exists(ctx, followerID, followingID)
	if err != nil {
		s.logger.Error("检查关注状态失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("检查关注状态失败 | Failed to check follow status: %w", err)
	}

	// Check if following follows follower | 检查 following 是否关注 follower
	isFollower, err := s.followRepo.Exists(ctx, followingID, followerID)
	if err != nil {
		s.logger.Error("检查粉丝状态失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("检查粉丝状态失败 | Failed to check follower status: %w", err)
	}

	// Check if mutual follow | 检查是否互相关注
	isMutual := isFollowing && isFollower

	return &schema.UserFollowStatusResponse{
		IsFollowing: isFollowing,
		IsFollower:  isFollower,
		IsMutual:    isMutual,
	}, nil
}

// GetFollowCounts Get user's follow counts | 获取用户关注数统计
// 优先从 Redis 计数器读取，不存在时从 DB 初始化
func (s *UserFollowService) GetFollowCounts(ctx context.Context, userID int) (followersCount, followingCount int, err error) {
	followersKey := fmt.Sprintf("user:followers:count:%d", userID)
	followingKey := fmt.Sprintf("user:following:count:%d", userID)

	// 尝试从缓存获取粉丝数
	followersCount, err = s.getOrInitCounter(ctx, followersKey, func() (int, error) {
		return s.followRepo.CountFollowers(ctx, userID)
	})
	if err != nil {
		s.logger.Error("获取粉丝数失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return 0, 0, fmt.Errorf("获取粉丝数失败 | Failed to get followers count: %w", err)
	}

	// 尝试从缓存获取关注数
	followingCount, err = s.getOrInitCounter(ctx, followingKey, func() (int, error) {
		return s.followRepo.CountFollowing(ctx, userID)
	})
	if err != nil {
		s.logger.Error("获取关注数失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return 0, 0, fmt.Errorf("获取关注数失败 | Failed to get following count: %w", err)
	}

	return followersCount, followingCount, nil
}

// getOrInitCounter Get counter from cache, init from DB if not exists | 从缓存获取计数器，不存在则从 DB 初始化
func (s *UserFollowService) getOrInitCounter(ctx context.Context, key string, dbCountFunc func() (int, error)) (int, error) {
	// 检查缓存是否存在
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		s.logger.Warn("检查计数器缓存失败，回退到数据库查询", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
		return dbCountFunc()
	}

	if exists {
		// 缓存存在，直接获取
		val, err := s.cache.Get(ctx, key)
		if err != nil {
			s.logger.Warn("获取计数器缓存失败，回退到数据库查询", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
			return dbCountFunc()
		}
		var count int
		if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
			s.logger.Warn("解析计数器缓存失败，回退到数据库查询", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
			return dbCountFunc()
		}
		return count, nil
	}

	// 缓存不存在，从 DB 查询并初始化
	count, err := dbCountFunc()
	if err != nil {
		return 0, err
	}

	// 写入缓存（30天有效期，通过 INCR/DECR 维护）
	if setErr := s.cache.SetEx(ctx, key, count, 2592000); setErr != nil {
		s.logger.Warn("初始化计数器缓存失败", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(setErr))
	}

	return count, nil
}

// incrFollowersCount Increment user followers count | 增加用户粉丝数计数器
func (s *UserFollowService) incrFollowersCount(ctx context.Context, userID int) {
	cacheKey := fmt.Sprintf("user:followers:count:%d", userID)
	if _, err := s.cache.Incr(ctx, cacheKey); err != nil {
		s.logger.Warn("增加粉丝数计数器失败", tracing.WithTraceIDField(ctx), zap.Int("user_id", userID), zap.Error(err))
	}
}

// decrFollowersCount Decrement user followers count | 减少用户粉丝数计数器
func (s *UserFollowService) decrFollowersCount(ctx context.Context, userID int) {
	cacheKey := fmt.Sprintf("user:followers:count:%d", userID)
	if _, err := s.cache.Decr(ctx, cacheKey); err != nil {
		s.logger.Warn("减少粉丝数计数器失败", tracing.WithTraceIDField(ctx), zap.Int("user_id", userID), zap.Error(err))
	}
}

// incrFollowingCount Increment user following count | 增加用户关注数计数器
func (s *UserFollowService) incrFollowingCount(ctx context.Context, userID int) {
	cacheKey := fmt.Sprintf("user:following:count:%d", userID)
	if _, err := s.cache.Incr(ctx, cacheKey); err != nil {
		s.logger.Warn("增加关注数计数器失败", tracing.WithTraceIDField(ctx), zap.Int("user_id", userID), zap.Error(err))
	}
}

// decrFollowingCount Decrement user following count | 减少用户关注数计数器
func (s *UserFollowService) decrFollowingCount(ctx context.Context, userID int) {
	cacheKey := fmt.Sprintf("user:following:count:%d", userID)
	if _, err := s.cache.Decr(ctx, cacheKey); err != nil {
		s.logger.Warn("减少关注数计数器失败", tracing.WithTraceIDField(ctx), zap.Int("user_id", userID), zap.Error(err))
	}
}
