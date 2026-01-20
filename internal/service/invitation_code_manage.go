package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/invitationcode"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
)

// IInvitationCodeManageService Invitation code management service interface | 邀请码管理服务接口
type IInvitationCodeManageService interface {
	// GetInvitationCodeList Get invitation code list with pagination and filters | 获取邀请码列表（分页和筛选）
	GetInvitationCodeList(ctx context.Context, page, pageSize int, keyword, status, mode string) ([]*ent.InvitationCode, int64, error)
	// CreateInvitationCode Create invitation code manually (admin) | 手动创建邀请码（管理员）
	CreateInvitationCode(ctx context.Context, code string, creatorID int, mode string, costAmount int, expiresAt *time.Time, remark string) (*ent.InvitationCode, error)
	// UpdateInvitationCode Update invitation code information | 更新邀请码信息
	UpdateInvitationCode(ctx context.Context, id int, expiresAt *time.Time, remark string) (*ent.InvitationCode, error)
	// UpdateInvitationCodeStatus Update invitation code status | 更新邀请码状态
	UpdateInvitationCodeStatus(ctx context.Context, id int, status string) (*ent.InvitationCode, error)
	// DeleteInvitationCode Delete invitation code | 删除邀请码
	DeleteInvitationCode(ctx context.Context, id int) error
	// GetInvitationCodeStats Get invitation code statistics | 获取邀请码统计信息
	GetInvitationCodeStats(ctx context.Context) (map[string]int64, error)
}

// InvitationCodeManageService Invitation code management service implementation | 邀请码管理服务实现
type InvitationCodeManageService struct {
	invitationCodeRepo repository.IInvitationCodeRepository
	userRepo           repository.IUserRepository
	logger             *zap.Logger
}

// NewInvitationCodeManageService Create invitation code management service instance | 创建邀请码管理服务实例
func NewInvitationCodeManageService(
	invitationCodeRepo repository.IInvitationCodeRepository,
	userRepo repository.IUserRepository,
	logger *zap.Logger,
) IInvitationCodeManageService {
	return &InvitationCodeManageService{
		invitationCodeRepo: invitationCodeRepo,
		userRepo:           userRepo,
		logger:             logger,
	}
}

// GetInvitationCodeList Get invitation code list with pagination and filters | 获取邀请码列表（分页和筛选）
func (s *InvitationCodeManageService) GetInvitationCodeList(ctx context.Context, page, pageSize int, keyword, status, mode string) ([]*ent.InvitationCode, int64, error) {
	offset := (page - 1) * pageSize

	codes, err := s.invitationCodeRepo.Query(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
		query := q.Order(ent.Desc(invitationcode.FieldCreatedAt))

		if keyword != "" {
			query = query.Where(invitationcode.CodeContains(keyword))
		}
		if status != "" {
			query = query.Where(invitationcode.StatusEQ(invitationcode.Status(status)))
		}
		if mode != "" {
			query = query.Where(invitationcode.GenerationModeEQ(invitationcode.GenerationMode(mode)))
		}

		return query.
			Offset(offset).
			Limit(pageSize)
	})
	if err != nil {
		s.logger.Error("Failed to query invitation codes | 查询邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, 0, errors.New("查询邀请码列表失败")
	}

	total, err := s.invitationCodeRepo.Count(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
		query := q
		if keyword != "" {
			query = query.Where(invitationcode.CodeContains(keyword))
		}
		if status != "" {
			query = query.Where(invitationcode.StatusEQ(invitationcode.Status(status)))
		}
		if mode != "" {
			query = query.Where(invitationcode.GenerationModeEQ(invitationcode.GenerationMode(mode)))
		}
		return query
	})
	if err != nil {
		s.logger.Error("Failed to count invitation codes | 统计邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, 0, errors.New("统计邀请码数量失败")
	}

	return codes, int64(total), nil
}

// CreateInvitationCode Create invitation code manually (admin) | 手动创建邀请码（管理员）
func (s *InvitationCodeManageService) CreateInvitationCode(ctx context.Context, code string, creatorID int, mode string, costAmount int, expiresAt *time.Time, remark string) (*ent.InvitationCode, error) {
	// Check if code already exists | 检查邀请码是否已存在
	existingCode, err := s.invitationCodeRepo.GetByCode(ctx, code)
	if err != nil {
		s.logger.Error("Failed to check invitation code existence | 检查邀请码存在性失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("检查邀请码失败")
	}
	if existingCode != nil {
		return nil, errors.New("邀请码已存在")
	}

	// Verify creator exists | 验证创建者存在
	creator, err := s.userRepo.GetByID(ctx, creatorID)
	if err != nil || creator == nil {
		return nil, errors.New("创建者不存在")
	}

	invCode, err := s.invitationCodeRepo.Create(ctx, func(c *ent.InvitationCodeCreate) *ent.InvitationCodeCreate {
		builder := c.
			SetCode(code).
			SetCreatorID(creatorID).
			SetStatus(invitationcode.StatusUnused).
			SetGenerationMode(invitationcode.GenerationMode(mode)).
			SetCostAmount(costAmount)

		if expiresAt != nil {
			builder = builder.SetExpiresAt(*expiresAt)
		}
		if remark != "" {
			builder = builder.SetRemark(remark)
		}

		return builder
	})
	if err != nil {
		s.logger.Error("Failed to create invitation code | 创建邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("创建邀请码失败")
	}

	return invCode, nil
}

// UpdateInvitationCode Update invitation code information | 更新邀请码信息
func (s *InvitationCodeManageService) UpdateInvitationCode(ctx context.Context, id int, expiresAt *time.Time, remark string) (*ent.InvitationCode, error) {
	invCode, err := s.invitationCodeRepo.GetByID(ctx, id)
	if err != nil || invCode == nil {
		return nil, errors.New("邀请码不存在")
	}

	updatedCode, err := s.invitationCodeRepo.Update(ctx, id, func(u *ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne {
		updater := u
		if expiresAt != nil {
			updater = updater.SetExpiresAt(*expiresAt)
		}
		if remark != "" {
			updater = updater.SetRemark(remark)
		}
		return updater
	})
	if err != nil {
		s.logger.Error("Failed to update invitation code | 更新邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("更新邀请码失败")
	}

	return updatedCode, nil
}

// UpdateInvitationCodeStatus Update invitation code status | 更新邀请码状态
func (s *InvitationCodeManageService) UpdateInvitationCodeStatus(ctx context.Context, id int, status string) (*ent.InvitationCode, error) {
	invCode, err := s.invitationCodeRepo.GetByID(ctx, id)
	if err != nil || invCode == nil {
		return nil, errors.New("邀请码不存在")
	}

	updatedCode, err := s.invitationCodeRepo.Update(ctx, id, func(u *ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne {
		return u.SetStatus(invitationcode.Status(status))
	})
	if err != nil {
		s.logger.Error("Failed to update invitation code status | 更新邀请码状态失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("更新邀请码状态失败")
	}

	return updatedCode, nil
}

// DeleteInvitationCode Delete invitation code | 删除邀请码
func (s *InvitationCodeManageService) DeleteInvitationCode(ctx context.Context, id int) error {
	invCode, err := s.invitationCodeRepo.GetByID(ctx, id)
	if err != nil || invCode == nil {
		return errors.New("邀请码不存在")
	}

	// Check if code is used | 检查邀请码是否已使用
	if invCode.Status == invitationcode.StatusUsed {
		return errors.New("已使用的邀请码不能删除")
	}

	err = s.invitationCodeRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete invitation code | 删除邀请码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("删除邀请码失败")
	}

	return nil
}

// GetInvitationCodeStats Get invitation code statistics | 获取邀请码统计信息
func (s *InvitationCodeManageService) GetInvitationCodeStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total count | 总数
	total, err := s.invitationCodeRepo.Count(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
		return q
	})
	if err != nil {
		s.logger.Error("Failed to count invitation codes | 统计邀请码总数失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("统计邀请码总数失败")
	}
	stats["total"] = int64(total)

	// Count by status | 按状态统计
	statuses := []invitationcode.Status{
		invitationcode.StatusUnused,
		invitationcode.StatusUsed,
		invitationcode.StatusExpired,
		invitationcode.StatusDisabled,
	}

	for _, status := range statuses {
		count, err := s.invitationCodeRepo.Count(ctx, func(q *ent.InvitationCodeQuery) *ent.InvitationCodeQuery {
			return q.Where(invitationcode.StatusEQ(status))
		})
		if err != nil {
			s.logger.Error("Failed to count invitation codes by status | 按状态统计邀请码失败",
				tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("status", status.String()))
			return nil, errors.New("统计邀请码状态失败")
		}
		stats[status.String()+"_count"] = int64(count)
	}

	return stats, nil
}
