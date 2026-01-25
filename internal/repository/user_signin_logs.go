package repository

import (
	"context"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/usersigninlogs"
)

// IUserSigninLogsRepository User signin logs repository interface | 用户签到日志仓储接口
type IUserSigninLogsRepository interface {
	// Exists Check if signin log exists | 检查签到日志是否存在
	Exists(ctx context.Context, userID int64, signDate time.Time) (bool, error)
	// Create Create signin log | 创建签到日志
	Create(ctx context.Context, userID int64, signDate time.Time) error
}

// UserSigninLogsRepository User signin logs repository implementation | 用户签到日志仓储实现
type UserSigninLogsRepository struct {
	db *ent.Client
}

// NewUserSigninLogsRepository Create user signin logs repository instance | 创建用户签到日志仓储实例
func NewUserSigninLogsRepository(db *ent.Client) IUserSigninLogsRepository {
	return &UserSigninLogsRepository{db: db}
}

// Exists Check if signin log exists | 检查签到日志是否存在
func (r *UserSigninLogsRepository) Exists(ctx context.Context, userID int64, signDate time.Time) (bool, error) {
	return r.db.UserSigninLogs.Query().
		Where(
			usersigninlogs.UserID(userID),
			usersigninlogs.SignDate(signDate),
		).
		Exist(ctx)
}

// Create Create signin log | 创建签到日志
func (r *UserSigninLogsRepository) Create(ctx context.Context, userID int64, signDate time.Time) error {
	return r.db.UserSigninLogs.Create().
		SetUserID(userID).
		SetSignDate(signDate).
		Exec(ctx)
}
