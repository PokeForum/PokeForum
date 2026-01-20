package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/invitationcode"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
)

// IInvitationCodeService Invitation code service interface | 邀请码服务接口
type IInvitationCodeService interface {
	// IsInvitationCodeEnabled Check if invitation code is enabled | 检查是否启用邀请码
	IsInvitationCodeEnabled(ctx context.Context) (bool, error)
	// ValidateCode Validate invitation code | 验证邀请码
	ValidateCode(ctx context.Context, code string) (*ent.InvitationCode, error)
	// UseCode Use invitation code (mark as used) | 使用邀请码（标记为已使用）
	UseCode(ctx context.Context, code string, userID int, ip, ua string) error
	// GenerateCode Generate invitation code for user | 为用户生成邀请码
	GenerateCode(ctx context.Context, userID int, username string) (*ent.InvitationCode, error)
	// GetUserCodes Get user's invitation codes | 获取用户的邀请码列表
	GetUserCodes(ctx context.Context, userID int, page, pageSize int) ([]*ent.InvitationCode, int, error)
}

// InvitationCodeService Invitation code service implementation | 邀请码服务实现
type InvitationCodeService struct {
	db                 *ent.Client
	invitationCodeRepo repository.IInvitationCodeRepository
	userRepo           repository.IUserRepository
	settingsService    ISettingsService
	logger             *zap.Logger
}

// NewInvitationCodeService Create invitation code service instance | 创建邀请码服务实例
func NewInvitationCodeService(
	db *ent.Client,
	invitationCodeRepo repository.IInvitationCodeRepository,
	userRepo repository.IUserRepository,
	settingsService ISettingsService,
	logger *zap.Logger,
) IInvitationCodeService {
	return &InvitationCodeService{
		db:                 db,
		invitationCodeRepo: invitationCodeRepo,
		userRepo:           userRepo,
		settingsService:    settingsService,
		logger:             logger,
	}
}

// IsInvitationCodeEnabled Check if invitation code is enabled | 检查是否启用邀请码
func (s *InvitationCodeService) IsInvitationCodeEnabled(ctx context.Context) (bool, error) {
	isEnable, err := s.settingsService.GetSettingByKey(ctx, _const.InvitationCodeIsEnable, _const.SettingBoolFalse.String())
	if err != nil {
		s.logger.Error("Failed to get invitation code setting | 获取邀请码设置失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return false, err
	}
	return isEnable == _const.SettingBoolTrue.String(), nil
}

// ValidateCode Validate invitation code | 验证邀请码
func (s *InvitationCodeService) ValidateCode(ctx context.Context, code string) (*ent.InvitationCode, error) {
	if code == "" {
		return nil, errors.New("邀请码不能为空")
	}

	invCode, err := s.invitationCodeRepo.GetByCode(ctx, code)
	if err != nil {
		s.logger.Error("Failed to query invitation code | 查询邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("邀请码查询失败")
	}
	if invCode == nil {
		return nil, errors.New("邀请码不存在")
	}

	// Check status | 检查状态
	if invCode.Status != invitationcode.StatusUnused {
		switch invCode.Status {
		case invitationcode.StatusUsed:
			return nil, errors.New("邀请码已被使用")
		case invitationcode.StatusExpired:
			return nil, errors.New("邀请码已过期")
		case invitationcode.StatusDisabled:
			return nil, errors.New("邀请码已被禁用")
		default:
			return nil, errors.New("邀请码状态无效")
		}
	}

	// Check expiration | 检查过期时间
	if invCode.ExpiresAt != nil && invCode.ExpiresAt.Before(time.Now()) {
		// Update status to expired | 更新状态为已过期
		_, _ = s.invitationCodeRepo.Update(ctx, invCode.ID, func(u *ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne {
			return u.SetStatus(invitationcode.StatusExpired)
		})
		return nil, errors.New("邀请码已过期")
	}

	return invCode, nil
}

// UseCode Use invitation code (mark as used) | 使用邀请码（标记为已使用）
func (s *InvitationCodeService) UseCode(ctx context.Context, code string, userID int, ip, ua string) error {
	invCode, err := s.ValidateCode(ctx, code)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = s.invitationCodeRepo.Update(ctx, invCode.ID, func(u *ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne {
		return u.
			SetStatus(invitationcode.StatusUsed).
			SetUsedByID(userID).
			SetUsedAt(now).
			SetUsedIP(ip).
			SetUsedUserAgent(ua)
	})
	if err != nil {
		s.logger.Error("Failed to update invitation code | 更新邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("code", code))
		return errors.New("邀请码使用失败")
	}

	return nil
}

// GenerateCode Generate invitation code for user | 为用户生成邀请码
func (s *InvitationCodeService) GenerateCode(ctx context.Context, userID int, username string) (*ent.InvitationCode, error) {
	// Check if invitation code is enabled | 检查是否启用邀请码
	enabled, err := s.IsInvitationCodeEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, errors.New("邀请码功能未启用")
	}

	// Get settings | 获取设置
	settings, err := s.settingsService.GetInvitationCodeSettings(ctx)
	if err != nil {
		s.logger.Error("Failed to get invitation code settings | 获取邀请码设置失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("获取邀请码设置失败")
	}

	// Check user's code count | 检查用户已生成的邀请码数量
	count, err := s.invitationCodeRepo.CountByCreatorID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count user invitation codes | 统计用户邀请码数量失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("查询邀请码数量失败")
	}
	if count >= settings.MaxGenerationCount {
		return nil, errors.New("已达到最大生成数量限制")
	}

	// Deduct cost if needed | 如果需要，扣除费用
	mode := settings.Mode
	cost := settings.Cost
	if mode == "points" || mode == "currency" {
		if cost > 0 {
			if err := s.deductCost(ctx, userID, username, mode, cost); err != nil {
				return nil, err
			}
		}
	}

	// Generate code | 生成邀请码
	codeStr, err := generateInvitationCode()
	if err != nil {
		s.logger.Error("Failed to generate invitation code | 生成邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("生成邀请码失败")
	}

	// Determine generation mode enum | 确定生成模式枚举
	var genMode invitationcode.GenerationMode
	switch mode {
	case "points":
		genMode = invitationcode.GenerationModePoints
	case "currency":
		genMode = invitationcode.GenerationModeCurrency
	default:
		genMode = invitationcode.GenerationModeDirect
	}

	// Create invitation code | 创建邀请码
	invCode, err := s.invitationCodeRepo.Create(ctx, func(c *ent.InvitationCodeCreate) *ent.InvitationCodeCreate {
		return c.
			SetCode(codeStr).
			SetCreatorID(userID).
			SetStatus(invitationcode.StatusUnused).
			SetGenerationMode(genMode).
			SetCostAmount(cost)
	})
	if err != nil {
		s.logger.Error("Failed to create invitation code | 创建邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("创建邀请码失败")
	}

	s.logger.Info("Invitation code generated successfully | 邀请码生成成功",
		tracing.WithTraceIDField(ctx),
		zap.Int("user_id", userID),
		zap.String("code", codeStr))

	return invCode, nil
}

// GetUserCodes Get user's invitation codes | 获取用户的邀请码列表
func (s *InvitationCodeService) GetUserCodes(ctx context.Context, userID int, page, pageSize int) ([]*ent.InvitationCode, int, error) {
	offset := (page - 1) * pageSize

	codes, err := s.invitationCodeRepo.Query(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
		return q.
			Where(invitationcode.CreatorID(userID)).
			Order(ent.Desc(invitationcode.FieldCreatedAt)).
			Offset(offset).
			Limit(pageSize)
	})
	if err != nil {
		s.logger.Error("Failed to query user invitation codes | 查询用户邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, 0, errors.New("查询邀请码列表失败")
	}

	total, err := s.invitationCodeRepo.Count(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
		return q.Where(invitationcode.CreatorID(userID))
	})
	if err != nil {
		s.logger.Error("Failed to count user invitation codes | 统计用户邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, 0, errors.New("统计邀请码数量失败")
	}

	return codes, total, nil
}

// deductCost Deduct cost from user | 从用户扣除费用
func (s *InvitationCodeService) deductCost(ctx context.Context, userID int, username, mode string, cost int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("用户不存在")
	}

	switch mode {
	case "points":
		currentBalance := user.Points
		if currentBalance < cost {
			return errors.New("积分不足")
		}
		newBalance := currentBalance - cost
		if err = s.userRepo.UpdatePoints(ctx, userID, newBalance); err != nil {
			return errors.New("扣除积分失败")
		}
	case "currency":
		currentBalance := user.Currency
		if currentBalance < cost {
			return errors.New("货币不足")
		}
		newBalance := currentBalance - cost
		if err = s.userRepo.UpdateCurrency(ctx, userID, newBalance); err != nil {
			return errors.New("扣除货币失败")
		}
	default:
		return nil
	}

	return nil
}

// generateInvitationCode Generate a random invitation code | 生成随机邀请码
func generateInvitationCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
