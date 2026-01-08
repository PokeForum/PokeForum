package repository

import (
	"context"
	"errors"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/usersigninstatus"
)

// IUserSigninStatusRepository User signin status repository interface | 用户签到状态仓储接口
type IUserSigninStatusRepository interface {
	// GetByUserID Get user signin status by user ID | 根据用户ID获取签到状态
	GetByUserID(ctx context.Context, userID int64) (*ent.UserSigninStatus, error)
	// Create Create user signin status | 创建用户签到状态
	Create(ctx context.Context, userID int64, signDate time.Time, continuousDays, totalDays int) error
	// Update Update user signin status | 更新用户签到状态
	Update(ctx context.Context, status *ent.UserSigninStatus, signDate time.Time, continuousDays, totalDays int) error
}

// UserSigninStatusRepository User signin status repository implementation | 用户签到状态仓储实现
type UserSigninStatusRepository struct {
	db *ent.Client
}

// NewUserSigninStatusRepository Create user signin status repository instance | 创建用户签到状态仓储实例
func NewUserSigninStatusRepository(db *ent.Client) IUserSigninStatusRepository {
	return &UserSigninStatusRepository{db: db}
}

// GetByUserID Get user signin status by user ID | 根据用户ID获取签到状态
func (r *UserSigninStatusRepository) GetByUserID(ctx context.Context, userID int64) (*ent.UserSigninStatus, error) {
	status, err := r.db.UserSigninStatus.Query().
		Where(usersigninstatus.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("签到状态不存在")
		}
		return nil, err
	}
	return status, nil
}

// Create Create user signin status | 创建用户签到状态
func (r *UserSigninStatusRepository) Create(ctx context.Context, userID int64, signDate time.Time, continuousDays, totalDays int) error {
	return r.db.UserSigninStatus.Create().
		SetUserID(userID).
		SetLastSigninDate(signDate).
		SetContinuousDays(continuousDays).
		SetTotalDays(totalDays).
		Exec(ctx)
}

// Update Update user signin status | 更新用户签到状态
func (r *UserSigninStatusRepository) Update(ctx context.Context, status *ent.UserSigninStatus, signDate time.Time, continuousDays, totalDays int) error {
	_, err := r.db.UserSigninStatus.UpdateOne(status).
		SetLastSigninDate(signDate).
		SetContinuousDays(continuousDays).
		SetTotalDays(totalDays).
		Save(ctx)
	return err
}
