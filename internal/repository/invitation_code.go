package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/invitationcode"
)

// IInvitationCodeRepository Invitation code repository interface | 邀请码仓储接口
type IInvitationCodeRepository interface {
	// Create Create invitation code | 创建邀请码
	Create(ctx context.Context, builderFunc func(*ent.InvitationCodeCreate) *ent.InvitationCodeCreate) (*ent.InvitationCode, error)
	// GetByCode Get invitation code by code string | 根据邀请码字符串获取邀请码
	GetByCode(ctx context.Context, code string) (*ent.InvitationCode, error)
	// GetByID Get invitation code by ID | 根据ID获取邀请码
	GetByID(ctx context.Context, id int) (*ent.InvitationCode, error)
	// Update Update invitation code | 更新邀请码
	Update(ctx context.Context, id int, builderFunc func(*ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne) (*ent.InvitationCode, error)
	// Delete Delete invitation code | 删除邀请码
	Delete(ctx context.Context, id int) error
	// Query Query invitation codes with condition | 条件查询邀请码
	Query(ctx context.Context, conditionFunc func(*ent.InvitationCodeQuery) *ent.InvitationCodeQuery) ([]*ent.InvitationCode, error)
	// Count Count invitation codes with condition | 条件统计邀请码数
	Count(ctx context.Context, conditionFunc func(*ent.InvitationCodeQuery) *ent.InvitationCodeQuery) (int, error)
	// CountByCreatorID Count invitation codes by creator ID | 统计用户创建的邀请码数量
	CountByCreatorID(ctx context.Context, creatorID int) (int, error)
}

// InvitationCodeRepository Invitation code repository implementation | 邀请码仓储实现
type InvitationCodeRepository struct {
	db *ent.Client
}

// NewInvitationCodeRepository Create invitation code repository instance | 创建邀请码仓储实例
func NewInvitationCodeRepository(db *ent.Client) IInvitationCodeRepository {
	return &InvitationCodeRepository{db: db}
}

// Create Create invitation code | 创建邀请码
func (r *InvitationCodeRepository) Create(ctx context.Context, builderFunc func(*ent.InvitationCodeCreate) *ent.InvitationCodeCreate) (*ent.InvitationCode, error) {
	creator := r.db.InvitationCode.Create()
	creator = builderFunc(creator)
	code, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建邀请码失败: %w", err)
	}
	return code, nil
}

// GetByCode Get invitation code by code string | 根据邀请码字符串获取邀请码
func (r *InvitationCodeRepository) GetByCode(ctx context.Context, code string) (*ent.InvitationCode, error) {
	result, err := r.db.InvitationCode.Query().
		Where(invitationcode.Code(code)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询邀请码失败: %w", err)
	}
	return result, nil
}

// GetByID Get invitation code by ID | 根据ID获取邀请码
func (r *InvitationCodeRepository) GetByID(ctx context.Context, id int) (*ent.InvitationCode, error) {
	result, err := r.db.InvitationCode.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询邀请码失败: %w", err)
	}
	return result, nil
}

// Update Update invitation code | 更新邀请码
func (r *InvitationCodeRepository) Update(ctx context.Context, id int, builderFunc func(*ent.InvitationCodeUpdateOne) *ent.InvitationCodeUpdateOne) (*ent.InvitationCode, error) {
	updater := r.db.InvitationCode.UpdateOneID(id)
	updater = builderFunc(updater)
	code, err := updater.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新邀请码失败: %w", err)
	}
	return code, nil
}

// Delete Delete invitation code | 删除邀请码
func (r *InvitationCodeRepository) Delete(ctx context.Context, id int) error {
	err := r.db.InvitationCode.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除邀请码失败: %w", err)
	}
	return nil
}

// Query Query invitation codes with condition | 条件查询邀请码
func (r *InvitationCodeRepository) Query(ctx context.Context, conditionFunc func(*ent.InvitationCodeQuery) *ent.InvitationCodeQuery) ([]*ent.InvitationCode, error) {
	query := r.db.InvitationCode.Query()
	query = conditionFunc(query)
	codes, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询邀请码失败: %w", err)
	}
	return codes, nil
}

// Count Count invitation codes with condition | 条件统计邀请码数
func (r *InvitationCodeRepository) Count(ctx context.Context, conditionFunc func(*ent.InvitationCodeQuery) *ent.InvitationCodeQuery) (int, error) {
	query := r.db.InvitationCode.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计邀请码数失败: %w", err)
	}
	return count, nil
}

// CountByCreatorID Count invitation codes by creator ID | 统计用户创建的邀请码数量
func (r *InvitationCodeRepository) CountByCreatorID(ctx context.Context, creatorID int) (int, error) {
	count, err := r.db.InvitationCode.Query().
		Where(invitationcode.CreatorID(creatorID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计用户邀请码数失败: %w", err)
	}
	return count, nil
}
