package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	pool   *redis.Pool
	logger *zap.Logger
}

// NewRedisLock 创建Redis分布式锁实例
func NewRedisLock(pool *redis.Pool, logger *zap.Logger) *RedisLock {
	return &RedisLock{
		pool:   pool,
		logger: logger,
	}
}

// getConn 获取Redis连接
func (l *RedisLock) getConn() redis.Conn {
	return l.pool.Get()
}

// LockOptions 锁选项
type LockOptions struct {
	// 锁的过期时间，默认10秒
	Expiration time.Duration
	// 获取锁的重试间隔，默认100毫秒
	RetryInterval time.Duration
	// 获取锁的超时时间，默认0（不重试）
	Timeout time.Duration
}

// DefaultLockOptions 默认锁选项
func DefaultLockOptions() *LockOptions {
	return &LockOptions{
		Expiration:    10 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		Timeout:       0, // 默认不重试
	}
}

// Lock 获取分布式锁
// key: 锁的键名
// value: 锁的值（通常使用唯一标识）
// options: 锁选项
// 返回：是否获取成功、错误信息
func (l *RedisLock) Lock(ctx context.Context, key string, value string, options *LockOptions) (bool, error) {
	if options == nil {
		options = DefaultLockOptions()
	}

	conn := l.getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			l.logger.Error("关闭Redis连接失败", zap.Error(err))
		}
	}()

	// 使用SET命令的NX和EX选项实现分布式锁
	// SET key value NX EX seconds
	expirationSeconds := int(options.Expiration.Seconds())

	// 如果设置了超时时间，则进行重试
	if options.Timeout > 0 {
		startTime := time.Now()
		for {
			// 尝试获取锁
			result, err := redis.String(conn.Do("SET", key, value, "NX", "EX", expirationSeconds))
			if err != nil && !errors.Is(err, redis.ErrNil) {
				l.logger.Error("获取分布式锁失败", zap.String("key", key), zap.Error(err))
				return false, fmt.Errorf("获取分布式锁失败: %w", err)
			}

			// 如果获取成功
			if result == "OK" {
				l.logger.Debug("获取分布式锁成功", zap.String("key", key), zap.String("value", value))
				return true, nil
			}

			// 检查是否超时
			if time.Since(startTime) >= options.Timeout {
				l.logger.Warn("获取分布式锁超时", zap.String("key", key), zap.Duration("timeout", options.Timeout))
				return false, nil
			}

			// 等待重试
			select {
			case <-ctx.Done():
				l.logger.Info("获取分布式锁被取消", zap.String("key", key))
				return false, ctx.Err()
			case <-time.After(options.RetryInterval):
				// 继续重试
			}
		}
	} else {
		// 不重试，只尝试一次
		result, err := redis.String(conn.Do("SET", key, value, "NX", "EX", expirationSeconds))
		if err != nil && !errors.Is(err, redis.ErrNil) {
			l.logger.Error("获取分布式锁失败", zap.String("key", key), zap.Error(err))
			return false, fmt.Errorf("获取分布式锁失败: %w", err)
		}

		if result == "OK" {
			l.logger.Debug("获取分布式锁成功", zap.String("key", key), zap.String("value", value))
			return true, nil
		}

		return false, nil
	}
}

// Unlock 释放分布式锁
// 使用Lua脚本确保只有锁的持有者才能释放锁
func (l *RedisLock) Unlock(ctx context.Context, key string, value string) error {
	conn := l.getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			l.logger.Error("关闭Redis连接失败", zap.Error(err))
		}
	}()

	// Lua脚本：只有当锁的值匹配时才删除
	// 这样可以确保只有锁的持有者才能释放锁
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := redis.Int(conn.Do("EVAL", luaScript, 1, key, value))
	if err != nil {
		l.logger.Error("释放分布式锁失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("释放分布式锁失败: %w", err)
	}

	if result == 1 {
		l.logger.Debug("释放分布式锁成功", zap.String("key", key), zap.String("value", value))
	} else {
		l.logger.Warn("分布式锁不存在或值不匹配", zap.String("key", key), zap.String("value", value))
	}

	return nil
}

// TryLock 尝试获取锁（不重试）
func (l *RedisLock) TryLock(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	options := &LockOptions{
		Expiration:    expiration,
		RetryInterval: 0,
		Timeout:       0,
	}
	return l.Lock(ctx, key, value, options)
}

// IsLocked 检查锁是否存在
func (l *RedisLock) IsLocked(ctx context.Context, key string) (bool, error) {
	conn := l.getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			l.logger.Error("关闭Redis连接失败", zap.Error(err))
		}
	}()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		l.logger.Error("检查锁状态失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("检查锁状态失败: %w", err)
	}

	return exists, nil
}
