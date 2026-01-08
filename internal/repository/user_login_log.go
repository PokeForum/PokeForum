package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
)

// IUserLoginLogRepository UserLoginLog repository interface | 用户登录日志仓储接口
type IUserLoginLogRepository interface {
	// Create Create login log | 创建登录日志
	Create(ctx context.Context, userID int, ipAddress, deviceInfo string, success bool) (*ent.UserLoginLog, error)
}

// UserLoginLogRepository UserLoginLog repository implementation | 用户登录日志仓储实现
type UserLoginLogRepository struct {
	db *ent.Client
}

// NewUserLoginLogRepository Create user login log repository instance | 创建用户登录日志仓储实例
func NewUserLoginLogRepository(db *ent.Client) IUserLoginLogRepository {
	return &UserLoginLogRepository{db: db}
}

// Create Create login log | 创建登录日志
func (r *UserLoginLogRepository) Create(ctx context.Context, userID int, ipAddress, deviceInfo string, success bool) (*ent.UserLoginLog, error) {
	log, err := r.db.UserLoginLog.Create().
		SetUserID(userID).
		SetIPAddress(ipAddress).
		SetDeviceInfo(deviceInfo).
		SetSuccess(success).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建登录日志失败: %w", err)
	}
	return log, nil
}
