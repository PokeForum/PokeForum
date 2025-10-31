package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// RedisCacheService Redis缓存服务实现
type RedisCacheService struct {
	pool   *redis.Pool // Redis连接池
	logger *zap.Logger // 日志记录器
}

// NewRedisCacheService 创建Redis缓存服务实例
func NewRedisCacheService(pool *redis.Pool, logger *zap.Logger) ICacheService {
	return &RedisCacheService{
		pool:   pool,
		logger: logger,
	}
}

// getConn 获取Redis连接
func (r *RedisCacheService) getConn() redis.Conn {
	return r.pool.Get()
}

// Get 获取缓存值
func (r *RedisCacheService) Get(key string) (string, error) {
	conn := r.getConn()
	defer conn.Close()

	value, err := redis.String(conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		r.logger.Error("获取缓存失败", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}
	return value, nil
}

// Set 设置缓存值（永久有效）
func (r *RedisCacheService) Set(key string, value interface{}) error {
	conn := r.getConn()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		r.logger.Error("设置缓存失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// SetEx 设置缓存值并指定过期时间（秒）
func (r *RedisCacheService) SetEx(key string, value interface{}, expiration int) error {
	conn := r.getConn()
	defer conn.Close()

	_, err := conn.Do("SETEX", key, expiration, value)
	if err != nil {
		r.logger.Error("设置缓存失败", zap.String("key", key), zap.Int("expiration", expiration), zap.Error(err))
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// SetExDuration 设置缓存值并指定过期时间（使用Duration）
func (r *RedisCacheService) SetExDuration(key string, value interface{}, expiration time.Duration) error {
	return r.SetEx(key, value, int(expiration.Seconds()))
}

// Del 删除缓存
func (r *RedisCacheService) Del(keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	conn := r.getConn()
	defer conn.Close()

	// 将keys转换为interface{}切片
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	count, err := redis.Int(conn.Do("DEL", args...))
	if err != nil {
		r.logger.Error("删除缓存失败", zap.Strings("keys", keys), zap.Error(err))
		return 0, fmt.Errorf("删除缓存失败: %w", err)
	}
	return count, nil
}

// Exists 检查缓存是否存在
func (r *RedisCacheService) Exists(key string) (bool, error) {
	conn := r.getConn()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		r.logger.Error("检查缓存是否存在失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("检查缓存是否存在失败: %w", err)
	}
	return exists, nil
}

// Expire 设置缓存过期时间
func (r *RedisCacheService) Expire(key string, expiration int) (bool, error) {
	conn := r.getConn()
	defer conn.Close()

	success, err := redis.Bool(conn.Do("EXPIRE", key, expiration))
	if err != nil {
		r.logger.Error("设置缓存过期时间失败", zap.String("key", key), zap.Int("expiration", expiration), zap.Error(err))
		return false, fmt.Errorf("设置缓存过期时间失败: %w", err)
	}
	return success, nil
}

// TTL 获取缓存剩余过期时间
func (r *RedisCacheService) TTL(key string) (int, error) {
	conn := r.getConn()
	defer conn.Close()

	ttl, err := redis.Int(conn.Do("TTL", key))
	if err != nil {
		r.logger.Error("获取缓存过期时间失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("获取缓存过期时间失败: %w", err)
	}
	return ttl, nil
}

// Incr 自增操作
func (r *RedisCacheService) Incr(key string) (int64, error) {
	conn := r.getConn()
	defer conn.Close()

	value, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		r.logger.Error("自增操作失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("自增操作失败: %w", err)
	}
	return value, nil
}

// Decr 自减操作
func (r *RedisCacheService) Decr(key string) (int64, error) {
	conn := r.getConn()
	defer conn.Close()

	value, err := redis.Int64(conn.Do("DECR", key))
	if err != nil {
		r.logger.Error("自减操作失败", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("自减操作失败: %w", err)
	}
	return value, nil
}

// HGet 获取哈希表字段值
func (r *RedisCacheService) HGet(key, field string) (string, error) {
	conn := r.getConn()
	defer conn.Close()

	value, err := redis.String(conn.Do("HGET", key, field))
	if err != nil && err != redis.ErrNil {
		r.logger.Error("获取哈希表字段值失败", zap.String("key", key), zap.String("field", field), zap.Error(err))
		return "", fmt.Errorf("获取哈希表字段值失败: %w", err)
	}
	return value, nil
}

// HSet 设置哈希表字段值
func (r *RedisCacheService) HSet(key, field string, value interface{}) error {
	conn := r.getConn()
	defer conn.Close()

	_, err := conn.Do("HSET", key, field, value)
	if err != nil {
		r.logger.Error("设置哈希表字段值失败", zap.String("key", key), zap.String("field", field), zap.Error(err))
		return fmt.Errorf("设置哈希表字段值失败: %w", err)
	}
	return nil
}

// HDel 删除哈希表字段
func (r *RedisCacheService) HDel(key string, fields ...string) (int, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	conn := r.getConn()
	defer conn.Close()

	// 构建参数列表
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i, field := range fields {
		args[i+1] = field
	}

	count, err := redis.Int(conn.Do("HDEL", args...))
	if err != nil {
		r.logger.Error("删除哈希表字段失败", zap.String("key", key), zap.Strings("fields", fields), zap.Error(err))
		return 0, fmt.Errorf("删除哈希表字段失败: %w", err)
	}
	return count, nil
}

// HGetAll 获取哈希表所有字段和值
func (r *RedisCacheService) HGetAll(key string) (map[string]string, error) {
	conn := r.getConn()
	defer conn.Close()

	values, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		r.logger.Error("获取哈希表所有字段失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取哈希表所有字段失败: %w", err)
	}
	return values, nil
}
