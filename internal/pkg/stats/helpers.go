package stats

import (
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/pkg/cache"
)

// Helper 统计助手,封装通用的统计操作
type Helper struct {
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewStatsHelper 创建统计助手实例
func NewStatsHelper(cacheService cache.ICacheService, logger *zap.Logger) *Helper {
	return &Helper{
		cache:  cacheService,
		logger: logger,
	}
}

// IncrStats 增加统计数据
// key: Redis Hash键
// field: Hash字段名
// increment: 增量值
func (h *Helper) IncrStats(ctx context.Context, key, field string, increment int64) error {
	_, err := h.cache.HIncrBy(ctx, key, field, increment)
	if err != nil {
		h.logger.Error("增加统计数据失败",
			zap.String("key", key),
			zap.String("field", field),
			zap.Int64("increment", increment),
			zap.Error(err))
		return err
	}
	return nil
}

// GetStats 获取统计数据
// key: Redis Hash键
// fields: 要获取的字段列表
// 返回: 字段值映射
func (h *Helper) GetStats(ctx context.Context, key string, fields []string) (map[string]int, error) {
	values, err := h.cache.HMGet(ctx, key, fields...)
	if err != nil {
		h.logger.Error("获取统计数据失败",
			zap.String("key", key),
			zap.Strings("fields", fields),
			zap.Error(err))
		return nil, err
	}

	result := make(map[string]int)
	for i, field := range fields {
		if i < len(values) && values[i] != "" {
			val, _ := strconv.Atoi(values[i]) //nolint:errcheck // 解析失败返回0是预期行为
			result[field] = val
		} else {
			result[field] = 0
		}
	}
	return result, nil
}

// SetStats 设置统计数据
// key: Redis Hash键
// fieldValues: 字段和值的映射
func (h *Helper) SetStats(ctx context.Context, key string, fieldValues map[string]int) error {
	// 将int转换为interface{}
	data := make(map[string]interface{})
	for field, value := range fieldValues {
		data[field] = value
	}

	err := h.cache.HMSet(ctx, key, data)
	if err != nil {
		h.logger.Error("设置统计数据失败",
			zap.String("key", key),
			zap.Error(err))
		return err
	}
	return nil
}

// SetUserAction 设置用户操作状态
// key: Redis Hash键
// actionType: 操作类型
func (h *Helper) SetUserAction(ctx context.Context, key string, actionType ActionType) error {
	err := h.cache.HSet(ctx, key, string(actionType), 1)
	if err != nil {
		h.logger.Error("设置用户操作状态失败",
			zap.String("key", key),
			zap.String("action_type", string(actionType)),
			zap.Error(err))
		return err
	}
	return nil
}

// RemoveUserAction 移除用户操作状态
// key: Redis Hash键
// actionType: 操作类型
func (h *Helper) RemoveUserAction(ctx context.Context, key string, actionType ActionType) error {
	_, err := h.cache.HDel(ctx, key, string(actionType))
	if err != nil {
		h.logger.Error("移除用户操作状态失败",
			zap.String("key", key),
			zap.String("action_type", string(actionType)),
			zap.Error(err))
		return err
	}
	return nil
}

// GetUserActions 获取用户操作状态
// key: Redis Hash键
// 返回: 操作类型到状态的映射
func (h *Helper) GetUserActions(ctx context.Context, key string) (map[string]bool, error) {
	values, err := h.cache.HGetAll(ctx, key)
	if err != nil && !errors.Is(err, redis.Nil) {
		h.logger.Error("获取用户操作状态失败",
			zap.String("key", key),
			zap.Error(err))
		return nil, err
	}

	result := make(map[string]bool)
	for field := range values {
		result[field] = true
	}
	return result, nil
}

// MarkDirty 标记对象为脏数据
// dirtySetKey: 脏数据集合键
// objectID: 对象ID
func (h *Helper) MarkDirty(ctx context.Context, dirtySetKey string, objectID int) error {
	_, err := h.cache.SAdd(ctx, dirtySetKey, objectID)
	if err != nil {
		h.logger.Error("标记脏数据失败",
			zap.String("dirty_set_key", dirtySetKey),
			zap.Int("object_id", objectID),
			zap.Error(err))
		return err
	}
	return nil
}

// GetDirtyIDs 获取所有脏数据ID
// dirtySetKey: 脏数据集合键
func (h *Helper) GetDirtyIDs(ctx context.Context, dirtySetKey string) ([]int, error) {
	members, err := h.cache.SMembers(ctx, dirtySetKey)
	if err != nil {
		h.logger.Error("获取脏数据ID失败",
			zap.String("dirty_set_key", dirtySetKey),
			zap.Error(err))
		return nil, err
	}

	ids := make([]int, 0, len(members))
	for _, member := range members {
		id, err := strconv.Atoi(member)
		if err != nil {
			h.logger.Warn("解析脏数据ID失败",
				zap.String("member", member),
				zap.Error(err))
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// RemoveDirtyIDs 移除脏数据ID
// dirtySetKey: 脏数据集合键
// objectIDs: 对象ID列表
func (h *Helper) RemoveDirtyIDs(ctx context.Context, dirtySetKey string, objectIDs []int) error {
	if len(objectIDs) == 0 {
		return nil
	}

	members := make([]interface{}, len(objectIDs))
	for i, id := range objectIDs {
		members[i] = id
	}

	_, err := h.cache.SRem(ctx, dirtySetKey, members...)
	if err != nil {
		h.logger.Error("移除脏数据ID失败",
			zap.String("dirty_set_key", dirtySetKey),
			zap.Error(err))
		return err
	}
	return nil
}

// GetDirtyCount 获取脏数据数量
// dirtySetKey: 脏数据集合键
func (h *Helper) GetDirtyCount(ctx context.Context, dirtySetKey string) (int64, error) {
	count, err := h.cache.SCard(ctx, dirtySetKey)
	if err != nil {
		h.logger.Error("获取脏数据数量失败",
			zap.String("dirty_set_key", dirtySetKey),
			zap.Error(err))
		return 0, err
	}
	return count, nil
}

// DeleteStatsCache 删除统计缓存
// key: Redis键
func (h *Helper) DeleteStatsCache(ctx context.Context, key string) error {
	_, err := h.cache.Del(ctx, key)
	if err != nil {
		h.logger.Error("删除统计缓存失败",
			zap.String("key", key),
			zap.Error(err))
		return err
	}
	return nil
}
