package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/blacklist"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IBlacklistService Blacklist service interface | 黑名单服务接口
type IBlacklistService interface {
	// GetUserBlacklist Get user blacklist | 获取用户黑名单列表
	GetUserBlacklist(ctx context.Context, userID int, page int, pageSize int) (*schema.UserBlacklistListResponse, error)
	// AddToBlacklist Add user to blacklist | 添加用户到黑名单
	AddToBlacklist(ctx context.Context, userID int, blockedUserID int) (*schema.UserBlacklistAddResponse, error)
	// RemoveFromBlacklist Remove user from blacklist | 从黑名单移除用户
	RemoveFromBlacklist(ctx context.Context, userID int, blockedUserID int) error
	// IsUserBlocked Check if user is blocked | 检查用户是否被拉黑
	IsUserBlocked(ctx context.Context, userID int, targetUserID int) (bool, error)
}

// BlacklistService Blacklist service implementation | 黑名单服务实现
type BlacklistService struct {
	db     *ent.Client
	logger *zap.Logger
}

// NewBlacklistService Create blacklist service instance | 创建黑名单服务实例
func NewBlacklistService(db *ent.Client, logger *zap.Logger) IBlacklistService {
	return &BlacklistService{
		db:     db,
		logger: logger,
	}
}

// GetUserBlacklist Get user blacklist | 获取用户黑名单列表
func (s *BlacklistService) GetUserBlacklist(ctx context.Context, userID int, page int, pageSize int) (*schema.UserBlacklistListResponse, error) {
	// Log record | 记录日志
	s.logger.Info("获取用户黑名单列表",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// Calculate offset | 计算偏移量
	offset := (page - 1) * pageSize

	// Query total blacklist count | 查询黑名单总数
	total, err := s.db.Blacklist.Query().
		Where(blacklist.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		s.logger.Error("查询黑名单总数失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("查询黑名单总数失败: %w", err)
	}

	// Query blacklist | 查询黑名单列表
	blacklistItems, err := s.db.Blacklist.Query().
		Where(blacklist.UserIDEQ(userID)).
		Order(ent.Desc(blacklist.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("查询黑名单列表失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("查询黑名单列表失败: %w", err)
	}

	// Collect all blocked user IDs | 收集所有被拉黑用户的ID
	blockedUserIDs := make([]int, len(blacklistItems))
	for i, item := range blacklistItems {
		blockedUserIDs[i] = item.BlockedUserID
	}

	// Batch query blocked user information | 批量查询被拉黑用户信息
	blockedUsers, err := s.db.User.Query().
		Where(user.IDIn(blockedUserIDs...)).
		Select(user.FieldID, user.FieldUsername, user.FieldAvatar).
		All(ctx)
	if err != nil {
		s.logger.Error("查询被拉黑用户信息失败",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("查询被拉黑用户信息失败: %w", err)
	}

	// Create mapping from user ID to user information | 创建用户ID到用户信息的映射
	userMap := make(map[int]*ent.User)
	for _, u := range blockedUsers {
		userMap[u.ID] = u
	}

	// Build response data | 构建响应数据
	list := make([]schema.UserBlacklistItem, 0, len(blacklistItems))
	for _, item := range blacklistItems {
		blockedUser := userMap[item.BlockedUserID]
		if blockedUser == nil {
			continue // User does not exist, skip | 用户不存在，跳过
		}

		listItem := schema.UserBlacklistItem{
			ID:              item.ID,
			BlockedUserID:   item.BlockedUserID,
			BlockedUsername: blockedUser.Username,
			BlockedAvatar:   blockedUser.Avatar,
			CreatedAt:       item.CreatedAt.Format(time_tools.DateTimeFormat),
		}
		list = append(list, listItem)
	}

	// Calculate total pages | 计算总页数
	totalPages := (int64(total) + int64(pageSize) - 1) / int64(pageSize)

	result := &schema.UserBlacklistListResponse{
		List:       list,
		Total:      int64(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	s.logger.Info("获取用户黑名单列表成功",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int64("total", int64(total)),
		zap.Int("page", page),
	)

	return result, nil
}

// AddToBlacklist Add user to blacklist | 添加用户到黑名单
func (s *BlacklistService) AddToBlacklist(ctx context.Context, userID int, blockedUserID int) (*schema.UserBlacklistAddResponse, error) {
	// Log record | 记录日志
	s.logger.Info("添加用户到黑名单",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
	)

	// Check if blocked user exists | 检查被拉黑用户是否存在
	exists, err := s.db.User.Query().
		Where(user.IDEQ(blockedUserID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查被拉黑用户是否存在失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("blocked_user_id", blockedUserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if !exists {
		return nil, errors.New("被拉黑用户不存在")
	}

	// Check if already in blacklist | 检查是否已经在黑名单中
	isBlocked, err := s.IsUserBlocked(ctx, userID, blockedUserID)
	if err != nil {
		return nil, fmt.Errorf("检查黑名单状态失败: %w", err)
	}
	if isBlocked {
		return nil, errors.New("用户已在黑名单中")
	}

	// Create blacklist record | 创建黑名单记录
	blacklistItem, err := s.db.Blacklist.Create().
		SetUserID(userID).
		SetBlockedUserID(blockedUserID).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建黑名单记录失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("user_id", userID),
			zap.Int("blocked_user_id", blockedUserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("创建黑名单记录失败: %w", err)
	}

	result := &schema.UserBlacklistAddResponse{
		ID:            blacklistItem.ID,
		UserID:        blacklistItem.UserID,
		BlockedUserID: blacklistItem.BlockedUserID,
		CreatedAt:     blacklistItem.CreatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("添加用户到黑名单成功",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
		zap.Int("blacklist_id", blacklistItem.ID),
	)

	return result, nil
}

// RemoveFromBlacklist Remove user from blacklist | 从黑名单移除用户
func (s *BlacklistService) RemoveFromBlacklist(ctx context.Context, userID int, blockedUserID int) error {
	// Log record | 记录日志
	s.logger.Info("从黑名单移除用户",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
	)

	// Find and delete blacklist record | 查找并删除黑名单记录
	affected, err := s.db.Blacklist.Delete().
		Where(
			blacklist.UserIDEQ(userID),
			blacklist.BlockedUserIDEQ(blockedUserID),
		).
		Exec(ctx)
	if err != nil {
		s.logger.Error("删除黑名单记录失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("user_id", userID),
			zap.Int("blocked_user_id", blockedUserID),
			zap.Error(err),
		)
		return fmt.Errorf("删除黑名单记录失败: %w", err)
	}

	if affected == 0 {
		return errors.New("黑名单记录不存在")
	}

	s.logger.Info("从黑名单移除用户成功",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
	)

	return nil
}

// IsUserBlocked Check if user is blocked | 检查用户是否被拉黑
func (s *BlacklistService) IsUserBlocked(ctx context.Context, userID int, targetUserID int) (bool, error) {
	// Log record | 记录日志
	s.logger.Debug("检查用户是否被拉黑",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("target_user_id", targetUserID),
	)

	// Query blacklist record | 查询黑名单记录
	exists, err := s.db.Blacklist.Query().
		Where(
			blacklist.UserIDEQ(userID),
			blacklist.BlockedUserIDEQ(targetUserID),
		).
		Exist(ctx)
	if err != nil {
		s.logger.Error("查询黑名单记录失败",
			tracing.WithTraceIDField(ctx),
			zap.Int("user_id", userID),
			zap.Int("target_user_id", targetUserID),
			zap.Error(err),
		)
		return false, fmt.Errorf("查询黑名单记录失败: %w", err)
	}

	return exists, nil
}
