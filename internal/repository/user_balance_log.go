package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/userbalancelog"
)

// IUserBalanceLogRepository User balance log repository interface | 用户余额日志仓储接口
type IUserBalanceLogRepository interface {
	// Create Create balance log | 创建余额日志
	Create(ctx context.Context, builderFunc func(*ent.UserBalanceLogCreate) *ent.UserBalanceLogCreate) (*ent.UserBalanceLog, error)
	// Query Query balance logs with condition | 条件查询余额日志
	Query(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) ([]*ent.UserBalanceLog, error)
	// Count Count balance logs with condition | 条件统计余额日志数
	Count(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) (int, error)
	// AggregateSum Aggregate sum of amount | 聚合求和
	AggregateSum(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) (int, error)
}

// UserBalanceLogRepository User balance log repository implementation | 用户余额日志仓储实现
type UserBalanceLogRepository struct {
	db *ent.Client
}

// NewUserBalanceLogRepository Create user balance log repository instance | 创建用户余额日志仓储实例
func NewUserBalanceLogRepository(db *ent.Client) IUserBalanceLogRepository {
	return &UserBalanceLogRepository{db: db}
}

// Create Create balance log | 创建余额日志
func (r *UserBalanceLogRepository) Create(ctx context.Context, builderFunc func(*ent.UserBalanceLogCreate) *ent.UserBalanceLogCreate) (*ent.UserBalanceLog, error) {
	creator := r.db.UserBalanceLog.Create()
	creator = builderFunc(creator)
	log, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建余额日志失败: %w", err)
	}
	return log, nil
}

// Query Query balance logs with condition | 条件查询余额日志
func (r *UserBalanceLogRepository) Query(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) ([]*ent.UserBalanceLog, error) {
	query := r.db.UserBalanceLog.Query()
	query = conditionFunc(query)
	logs, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询余额日志失败: %w", err)
	}
	return logs, nil
}

// Count Count balance logs with condition | 条件统计余额日志数
func (r *UserBalanceLogRepository) Count(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) (int, error) {
	query := r.db.UserBalanceLog.Query()
	query = conditionFunc(query)
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("统计余额日志数失败: %w", err)
	}
	return count, nil
}

// AggregateSum Aggregate sum of amount | 聚合求和
func (r *UserBalanceLogRepository) AggregateSum(ctx context.Context, conditionFunc func(*ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery) (int, error) {
	query := r.db.UserBalanceLog.Query()
	query = conditionFunc(query)
	sum, err := query.Aggregate(ent.Sum(userbalancelog.FieldAmount)).Int(ctx)
	if err != nil {
		return 0, fmt.Errorf("聚合求和失败: %w", err)
	}
	return sum, nil
}
