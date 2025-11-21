package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisLock 创建Redis分布式锁实例
func NewRedisLock(client *redis.Client, logger *zap.Logger) *RedisLock {
	return &RedisLock{
		client: client,
		logger: logger,
	}
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

	// 如果设置了超时时间，则进行重试
	if options.Timeout > 0 {
		startTime := time.Now()
		for {
			// 尝试获取锁
			success, err := l.client.SetNX(ctx, key, value, options.Expiration).Result()
			if err != nil {
				l.logger.Error("获取分布式锁失败", zap.String("key", key), zap.Error(err))
				return false, fmt.Errorf("获取分布式锁失败: %w", err)
			}

			// 如果获取成功
			if success {
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
		success, err := l.client.SetNX(ctx, key, value, options.Expiration).Result()
		if err != nil {
			l.logger.Error("获取分布式锁失败", zap.String("key", key), zap.Error(err))
			return false, fmt.Errorf("获取分布式锁失败: %w", err)
		}

		if success {
			l.logger.Debug("获取分布式锁成功", zap.String("key", key), zap.String("value", value))
			return true, nil
		}

		return false, nil
	}
}

// Unlock 释放分布式锁
// 使用Lua脚本确保只有锁的持有者才能释放锁
func (l *RedisLock) Unlock(ctx context.Context, key string, value string) error {
	// Lua脚本：只有当锁的值匹配时才删除
	// 这样可以确保只有锁的持有者才能释放锁
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, luaScript, []string{key}, value).Int()
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
	count, err := l.client.Exists(ctx, key).Result()
	if err != nil {
		l.logger.Error("检查锁状态失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("检查锁状态失败: %w", err)
	}

	return count > 0, nil
}
