package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCacheService Redis缓存服务实现
type RedisCacheService struct {
	client *redis.Client // Redis客户端
	logger *zap.Logger   // 日志记录器
}

// NewRedisCacheService 创建Redis缓存服务实例
func NewRedisCacheService(client *redis.Client, logger *zap.Logger) ICacheService {
	return &RedisCacheService{
		client: client,
		logger: logger,
	}
}

// Get 获取缓存值
func (r *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.logger.Error("获取缓存失败", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}
	return value, nil
}

// Set 设置缓存值（永久有效）
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}) error {
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		r.logger.Error("设置缓存失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// SetEx 设置缓存值并指定过期时间（秒）
func (r *RedisCacheService) SetEx(ctx context.Context, key string, value interface{}, expiration int) error {
	err := r.client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
	if err != nil {
		r.logger.Error("设置缓存失败", zap.String("key", key), zap.Int("expiration", expiration), zap.Error(err))
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// SetExDuration 设置缓存值并指定过期时间（使用Duration）
func (r *RedisCacheService) SetExDuration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.SetEx(ctx, key, value, int(expiration.Seconds()))
}

// Del 删除缓存
func (r *RedisCacheService) Del(ctx context.Context, keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	count, err := r.client.Del(ctx, keys...).Result()
	if err != nil {
		r.logger.Error("删除缓存失败", zap.Strings("keys", keys), zap.Error(err))
		return 0, fmt.Errorf("删除缓存失败: %w", err)
	}
	return int(count), nil
}

// Exists 检查缓存是否存在
func (r *RedisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("检查缓存是否存在失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("检查缓存是否存在失败: %w", err)
	}
	return count > 0, nil
}

// Expire 设置缓存过期时间
func (r *RedisCacheService) Expire(ctx context.Context, key string, expiration int) (bool, error) {
	success, err := r.client.Expire(ctx, key, time.Duration(expiration)*time.Second).Result()
	if err != nil {
		r.logger.Error("设置缓存过期时间失败", zap.String("key", key), zap.Int("expiration", expiration), zap.Error(err))
		return false, fmt.Errorf("设置缓存过期时间失败: %w", err)
	}
	return success, nil
}

// TTL 获取缓存剩余过期时间
func (r *RedisCacheService) TTL(ctx context.Context, key string) (int, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.logger.Error("获取缓存过期时间失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("获取缓存过期时间失败: %w", err)
	}
	return int(ttl.Seconds()), nil
}

// Incr 自增操作
func (r *RedisCacheService) Incr(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		r.logger.Error("自增操作失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("自增操作失败: %w", err)
	}
	return value, nil
}

// Decr 自减操作
func (r *RedisCacheService) Decr(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		r.logger.Error("自减操作失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("自减操作失败: %w", err)
	}
	return value, nil
}

// HGet 获取哈希表字段值
func (r *RedisCacheService) HGet(ctx context.Context, key, field string) (string, error) {
	value, err := r.client.HGet(ctx, key, field).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.logger.Error("获取哈希表字段值失败", zap.String("key", key), zap.String("field", field), zap.Error(err))
		return "", fmt.Errorf("获取哈希表字段值失败: %w", err)
	}
	return value, nil
}

// HSet 设置哈希表字段值
func (r *RedisCacheService) HSet(ctx context.Context, key, field string, value interface{}) error {
	err := r.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		r.logger.Error("设置哈希表字段值失败", zap.String("key", key), zap.String("field", field), zap.Error(err))
		return fmt.Errorf("设置哈希表字段值失败: %w", err)
	}
	return nil
}

// HDel 删除哈希表字段
func (r *RedisCacheService) HDel(ctx context.Context, key string, fields ...string) (int, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	count, err := r.client.HDel(ctx, key, fields...).Result()
	if err != nil {
		r.logger.Error("删除哈希表字段失败", zap.String("key", key), zap.Strings("fields", fields), zap.Error(err))
		return 0, fmt.Errorf("删除哈希表字段失败: %w", err)
	}
	return int(count), nil
}

// HGetAll 获取哈希表所有字段和值
func (r *RedisCacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		r.logger.Error("获取哈希表所有字段失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取哈希表所有字段失败: %w", err)
	}
	return values, nil
}

// HIncrBy 哈希表字段自增
func (r *RedisCacheService) HIncrBy(ctx context.Context, key, field string, increment int64) (int64, error) {
	value, err := r.client.HIncrBy(ctx, key, field, increment).Result()
	if err != nil {
		r.logger.Error("哈希表字段自增失败", zap.String("key", key), zap.String("field", field), zap.Int64("increment", increment), zap.Error(err))
		return 0, fmt.Errorf("哈希表字段自增失败: %w", err)
	}
	return value, nil
}

// HMGet 批量获取哈希表字段值
func (r *RedisCacheService) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	if len(fields) == 0 {
		return []string{}, nil
	}

	values, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		r.logger.Error("批量获取哈希表字段值失败", zap.String("key", key), zap.Strings("fields", fields), zap.Error(err))
		return nil, fmt.Errorf("批量获取哈希表字段值失败: %w", err)
	}

	// 转换为[]string
	result := make([]string, len(values))
	for i, v := range values {
		if v != nil {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
	}
	return result, nil
}

// HMSet 批量设置哈希表字段值
func (r *RedisCacheService) HMSet(ctx context.Context, key string, fieldValues map[string]interface{}) error {
	if len(fieldValues) == 0 {
		return nil
	}

	err := r.client.HMSet(ctx, key, fieldValues).Err()
	if err != nil {
		r.logger.Error("批量设置哈希表字段值失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("批量设置哈希表字段值失败: %w", err)
	}
	return nil
}

// SAdd 向集合添加成员
func (r *RedisCacheService) SAdd(ctx context.Context, key string, members ...interface{}) (int, error) {
	if len(members) == 0 {
		return 0, nil
	}

	count, err := r.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		r.logger.Error("向集合添加成员失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("向集合添加成员失败: %w", err)
	}
	return int(count), nil
}

// SMembers 获取集合所有成员
func (r *RedisCacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		r.logger.Error("获取集合成员失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取集合成员失败: %w", err)
	}
	return members, nil
}

// SRem 从集合移除成员
func (r *RedisCacheService) SRem(ctx context.Context, key string, members ...interface{}) (int, error) {
	if len(members) == 0 {
		return 0, nil
	}

	count, err := r.client.SRem(ctx, key, members...).Result()
	if err != nil {
		r.logger.Error("从集合移除成员失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("从集合移除成员失败: %w", err)
	}
	return int(count), nil
}

// SCard 获取集合成员数量
func (r *RedisCacheService) SCard(ctx context.Context, key string) (int64, error) {
	count, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		r.logger.Error("获取集合成员数量失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("获取集合成员数量失败: %w", err)
	}
	return count, nil
}

// SIsMember 判断成员是否在集合中
func (r *RedisCacheService) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	exists, err := r.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		r.logger.Error("判断成员是否在集合中失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("判断成员是否在集合中失败: %w", err)
	}
	return exists, nil
}

// ZAdd 向有序集合添加成员
func (r *RedisCacheService) ZAdd(ctx context.Context, key string, member string, score float64) error {
	err := r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
	if err != nil {
		r.logger.Error("向有序集合添加成员失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("向有序集合添加成员失败: %w", err)
	}
	return nil
}

// ZRevRangeWithScores 按分数从高到低获取有序集合成员
func (r *RedisCacheService) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error) {
	result, err := r.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		r.logger.Error("获取有序集合成员失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取有序集合成员失败: %w", err)
	}

	members := make([]ZMember, len(result))
	for i, z := range result {
		member := ""
		if str, ok := z.Member.(string); ok {
			member = str
		}
		members[i] = ZMember{
			Member: member,
			Score:  z.Score,
		}
	}
	return members, nil
}

// ZRevRank 获取成员在有序集合中的排名（从高到低）
func (r *RedisCacheService) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	rank, err := r.client.ZRevRank(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, nil
		}
		r.logger.Error("获取成员排名失败", zap.String("key", key), zap.String("member", member), zap.Error(err))
		return -1, fmt.Errorf("获取成员排名失败: %w", err)
	}
	return rank, nil
}

// XAdd 向Stream添加消息
func (r *RedisCacheService) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	messageID, err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
	if err != nil {
		r.logger.Error("向Stream添加消息失败", zap.String("stream", stream), zap.Error(err))
		return "", fmt.Errorf("向Stream添加消息失败: %w", err)
	}

	r.logger.Debug("向Stream添加消息成功", zap.String("stream", stream), zap.String("message_id", messageID))
	return messageID, nil
}

// XReadGroup 从消费者组读取消息
func (r *RedisCacheService) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string, count int64, block time.Duration) ([]map[string]interface{}, error) {
	// 构建streams参数
	streamKeys := make([]string, 0, len(streams))
	for key := range streams {
		streamKeys = append(streamKeys, key)
	}

	result, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  append(streamKeys, ">"),
		Count:    count,
		Block:    block,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.logger.Error("从消费者组读取消息失败", zap.String("group", group), zap.String("consumer", consumer), zap.Error(err))
		return nil, fmt.Errorf("从消费者组读取消息失败: %w", err)
	}

	var messages []map[string]interface{}
	for _, stream := range result {
		for _, message := range stream.Messages {
			fieldsInterface := make(map[string]interface{})
			for k, v := range message.Values {
				fieldsInterface[k] = v
			}
			fieldsInterface["message_id"] = message.ID
			messages = append(messages, fieldsInterface)
		}
	}

	r.logger.Debug("从消费者组读取消息成功",
		zap.String("group", group),
		zap.String("consumer", consumer),
		zap.Int("message_count", len(messages)))

	return messages, nil
}

// XAck 确认消息已处理
func (r *RedisCacheService) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	count, err := r.client.XAck(ctx, stream, group, ids...).Result()
	if err != nil {
		r.logger.Error("确认消息失败", zap.String("stream", stream), zap.String("group", group), zap.Error(err))
		return 0, fmt.Errorf("确认消息失败: %w", err)
	}

	r.logger.Debug("确认消息成功",
		zap.String("stream", stream),
		zap.String("group", group),
		zap.Int64("ack_count", count))

	return count, nil
}

// XPending 查看待处理消息
func (r *RedisCacheService) XPending(ctx context.Context, stream, group string) (map[string]interface{}, error) {
	result, err := r.client.XPending(ctx, stream, group).Result()
	if err != nil {
		r.logger.Error("查询待处理消息失败", zap.String("stream", stream), zap.String("group", group), zap.Error(err))
		return nil, fmt.Errorf("查询待处理消息失败: %w", err)
	}

	consumers := make([]map[string]interface{}, 0, len(result.Consumers))
	for consumerName, count := range result.Consumers {
		consumers = append(consumers, map[string]interface{}{
			"consumer":      consumerName,
			"pending_count": count,
		})
	}

	return map[string]interface{}{
		"total":     result.Count,
		"min_id":    result.Lower,
		"max_id":    result.Higher,
		"consumers": consumers,
	}, nil
}

// XDel 删除Stream中的消息
func (r *RedisCacheService) XDel(ctx context.Context, stream string, ids ...string) (int64, error) {
	count, err := r.client.XDel(ctx, stream, ids...).Result()
	if err != nil {
		r.logger.Error("删除Stream消息失败", zap.String("stream", stream), zap.Error(err))
		return 0, fmt.Errorf("删除Stream消息失败: %w", err)
	}

	r.logger.Debug("删除Stream消息成功",
		zap.String("stream", stream),
		zap.Int64("deleted_count", count))

	return count, nil
}

// XGroupCreate 创建消费者组
func (r *RedisCacheService) XGroupCreate(ctx context.Context, stream, group, id string) error {
	err := r.client.XGroupCreateMkStream(ctx, stream, group, id).Err()
	if err != nil {
		// 如果消费者组已存在，则忽略错误
		if strings.Contains(err.Error(), "BUSYGROUP Consumer Group name already exists") {
			r.logger.Debug("消费者组已存在", zap.String("stream", stream), zap.String("group", group))
			return nil
		}
		r.logger.Error("创建消费者组失败", zap.String("stream", stream), zap.String("group", group), zap.Error(err))
		return fmt.Errorf("创建消费者组失败: %w", err)
	}

	r.logger.Info("创建消费者组成功", zap.String("stream", stream), zap.String("group", group))
	return nil
}

// XLen 获取Stream长度
func (r *RedisCacheService) XLen(ctx context.Context, stream string) (int64, error) {
	length, err := r.client.XLen(ctx, stream).Result()
	if err != nil {
		r.logger.Error("获取Stream长度失败", zap.String("stream", stream), zap.Error(err))
		return 0, fmt.Errorf("获取Stream长度失败: %w", err)
	}

	return length, nil
}

// Ping Check Redis connection | 检查Redis连接
func (r *RedisCacheService) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}
