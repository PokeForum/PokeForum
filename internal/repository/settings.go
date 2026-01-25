package repository

import (
	"context"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/settings"
)

// ISettingsRepository Settings repository interface | 设置仓储接口
type ISettingsRepository interface {
	// GetByKey Get setting by key | 根据key获取设置
	GetByKey(ctx context.Context, key string) (*ent.Settings, error)
	// GetByModule Get settings by module | 根据模块获取设置列表
	GetByModule(ctx context.Context, module settings.Module) ([]*ent.Settings, error)
	// Upsert Update or insert setting | 更新或插入设置
	Upsert(ctx context.Context, module settings.Module, key, value string, valueType settings.ValueType) error
	// BatchUpsert Batch update or insert settings | 批量更新或插入设置
	BatchUpsert(ctx context.Context, module settings.Module, items map[string]string, valueType settings.ValueType) error
}

// SettingsRepository Settings repository implementation | 设置仓储实现
type SettingsRepository struct {
	db *ent.Client
}

// NewSettingsRepository Create settings repository instance | 创建设置仓储实例
func NewSettingsRepository(db *ent.Client) ISettingsRepository {
	return &SettingsRepository{db: db}
}

// GetByKey Get setting by key | 根据key获取设置
func (r *SettingsRepository) GetByKey(ctx context.Context, key string) (*ent.Settings, error) {
	setting, err := r.db.Settings.Query().
		Where(settings.KeyEQ(key)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设置失败: %w", err)
	}
	return setting, nil
}

// GetByModule Get settings by module | 根据模块获取设置列表
func (r *SettingsRepository) GetByModule(ctx context.Context, module settings.Module) ([]*ent.Settings, error) {
	configs, err := r.db.Settings.Query().
		Where(settings.ModuleEQ(module)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	return configs, nil
}

// Upsert Update or insert setting | 更新或插入设置
func (r *SettingsRepository) Upsert(ctx context.Context, module settings.Module, key, value string, valueType settings.ValueType) error {
	existing, err := r.db.Settings.Query().
		Where(
			settings.ModuleEQ(module),
			settings.KeyEQ(key),
		).
		First(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("查询配置失败: %w", err)
	}

	if existing != nil {
		_, err := r.db.Settings.UpdateOne(existing).
			SetValue(value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("更新配置失败: %w", err)
		}
	} else {
		_, err := r.db.Settings.Create().
			SetModule(module).
			SetKey(key).
			SetValue(value).
			SetValueType(valueType).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("创建配置失败: %w", err)
		}
	}

	return nil
}

// BatchUpsert Batch update or insert settings | 批量更新或插入设置
func (r *SettingsRepository) BatchUpsert(ctx context.Context, module settings.Module, items map[string]string, valueType settings.ValueType) error {
	for key, value := range items {
		if err := r.Upsert(ctx, module, key, value, valueType); err != nil {
			return err
		}
	}
	return nil
}
