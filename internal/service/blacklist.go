package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/blacklist"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IBlacklistService 黑名单服务接口
type IBlacklistService interface {
	// GetUserBlacklist 获取用户黑名单列表
	GetUserBlacklist(ctx context.Context, userID int, page int, pageSize int) (*schema.UserBlacklistListResponse, error)
	// AddToBlacklist 添加用户到黑名单
	AddToBlacklist(ctx context.Context, userID int, blockedUserID int) (*schema.UserBlacklistAddResponse, error)
	// RemoveFromBlacklist 从黑名单移除用户
	RemoveFromBlacklist(ctx context.Context, userID int, blockedUserID int) error
	// IsUserBlocked 检查用户是否被拉黑
	IsUserBlocked(ctx context.Context, userID int, targetUserID int) (bool, error)
}

// BlacklistService 黑名单服务实现
type BlacklistService struct {
	db     *ent.Client
	logger *zap.Logger
}

// NewBlacklistService 创建黑名单服务实例
func NewBlacklistService(db *ent.Client, logger *zap.Logger) IBlacklistService {
	return &BlacklistService{
		db:     db,
		logger: logger,
	}
}

// GetUserBlacklist 获取用户黑名单列表
func (s *BlacklistService) GetUserBlacklist(ctx context.Context, userID int, page int, pageSize int) (*schema.UserBlacklistListResponse, error) {
	// 记录日志
	s.logger.Info("获取用户黑名单列表",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询黑名单总数
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

	// 查询黑名单列表
	blacklistItems, err := s.db.Blacklist.Query().
		Where(blacklist.UserIDEQ(userID)).
		WithBlockedUser(func(query *ent.UserQuery) {
			query.Select(user.FieldID, user.FieldUsername, user.FieldAvatar)
		}).
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

	// 构建响应数据
	list := make([]schema.UserBlacklistItem, 0, len(blacklistItems))
	for _, item := range blacklistItems {
		blockedUser := item.Edges.BlockedUser

		listItem := schema.UserBlacklistItem{
			ID:              item.ID,
			BlockedUserID:   item.BlockedUserID,
			BlockedUsername: blockedUser.Username,
			BlockedAvatar:   blockedUser.Avatar,
			CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		}
		list = append(list, listItem)
	}

	// 计算总页数
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

// AddToBlacklist 添加用户到黑名单
func (s *BlacklistService) AddToBlacklist(ctx context.Context, userID int, blockedUserID int) (*schema.UserBlacklistAddResponse, error) {
	// 记录日志
	s.logger.Info("添加用户到黑名单",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
	)

	// 检查被拉黑用户是否存在
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

	// 检查是否已经在黑名单中
	isBlocked, err := s.IsUserBlocked(ctx, userID, blockedUserID)
	if err != nil {
		return nil, fmt.Errorf("检查黑名单状态失败: %w", err)
	}
	if isBlocked {
		return nil, errors.New("用户已在黑名单中")
	}

	// 创建黑名单记录
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
		CreatedAt:     blacklistItem.CreatedAt.Format(time.RFC3339),
	}

	s.logger.Info("添加用户到黑名单成功",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
		zap.Int("blacklist_id", blacklistItem.ID),
	)

	return result, nil
}

// RemoveFromBlacklist 从黑名单移除用户
func (s *BlacklistService) RemoveFromBlacklist(ctx context.Context, userID int, blockedUserID int) error {
	// 记录日志
	s.logger.Info("从黑名单移除用户",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("blocked_user_id", blockedUserID),
	)

	// 查找并删除黑名单记录
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

// IsUserBlocked 检查用户是否被拉黑
func (s *BlacklistService) IsUserBlocked(ctx context.Context, userID int, targetUserID int) (bool, error) {
	// 记录日志
	s.logger.Debug("检查用户是否被拉黑",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.Int("target_user_id", targetUserID),
	)

	// 查询黑名单记录
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
