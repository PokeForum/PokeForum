package cache

import "time"

// ICacheService Redis缓存服务接口
// 提供统一的缓存操作方法，避免在各个Service中重复编写Redis操作代码
type ICacheService interface {
	// Get 获取缓存值
	// key: 缓存键名
	// 返回: 缓存值和错误信息
	Get(key string) (string, error)

	// Set 设置缓存值（永久有效）
	// key: 缓存键名
	// value: 缓存值
	// 返回: 错误信息
	Set(key string, value interface{}) error

	// SetEx 设置缓存值并指定过期时间
	// key: 缓存键名
	// value: 缓存值
	// expiration: 过期时间（秒）
	// 返回: 错误信息
	SetEx(key string, value interface{}, expiration int) error

	// SetExDuration 设置缓存值并指定过期时间（使用Duration）
	// key: 缓存键名
	// value: 缓存值
	// expiration: 过期时间
	// 返回: 错误信息
	SetExDuration(key string, value interface{}, expiration time.Duration) error

	// Del 删除缓存
	// keys: 要删除的缓存键名（支持多个）
	// 返回: 删除的数量和错误信息
	Del(keys ...string) (int, error)

	// Exists 检查缓存是否存在
	// key: 缓存键名
	// 返回: 是否存在和错误信息
	Exists(key string) (bool, error)

	// Expire 设置缓存过期时间
	// key: 缓存键名
	// expiration: 过期时间（秒）
	// 返回: 是否成功和错误信息
	Expire(key string, expiration int) (bool, error)

	// TTL 获取缓存剩余过期时间
	// key: 缓存键名
	// 返回: 剩余秒数和错误信息（-1表示永久，-2表示不存在）
	TTL(key string) (int, error)

	// Incr 自增操作
	// key: 缓存键名
	// 返回: 自增后的值和错误信息
	Incr(key string) (int64, error)

	// Decr 自减操作
	// key: 缓存键名
	// 返回: 自减后的值和错误信息
	Decr(key string) (int64, error)

	// HGet 获取哈希表字段值
	// key: 哈希表键名
	// field: 字段名
	// 返回: 字段值和错误信息
	HGet(key, field string) (string, error)

	// HSet 设置哈希表字段值
	// key: 哈希表键名
	// field: 字段名
	// value: 字段值
	// 返回: 错误信息
	HSet(key, field string, value interface{}) error

	// HDel 删除哈希表字段
	// key: 哈希表键名
	// fields: 要删除的字段名（支持多个）
	// 返回: 删除的数量和错误信息
	HDel(key string, fields ...string) (int, error)

	// HGetAll 获取哈希表所有字段和值
	// key: 哈希表键名
	// 返回: 字段和值的映射以及错误信息
	HGetAll(key string) (map[string]string, error)

	// HIncrBy 哈希表字段自增
	// key: 哈希表键名
	// field: 字段名
	// increment: 增量值
	// 返回: 自增后的值和错误信息
	HIncrBy(key, field string, increment int64) (int64, error)

	// HMGet 批量获取哈希表字段值
	// key: 哈希表键名
	// fields: 字段名列表
	// 返回: 字段值列表和错误信息
	HMGet(key string, fields ...string) ([]string, error)

	// HMSet 批量设置哈希表字段值
	// key: 哈希表键名
	// fieldValues: 字段和值的映射
	// 返回: 错误信息
	HMSet(key string, fieldValues map[string]interface{}) error

	// SAdd 向集合添加成员
	// key: 集合键名
	// members: 要添加的成员（支持多个）
	// 返回: 添加的数量和错误信息
	SAdd(key string, members ...interface{}) (int, error)

	// SMembers 获取集合所有成员
	// key: 集合键名
	// 返回: 成员列表和错误信息
	SMembers(key string) ([]string, error)

	// SRem 从集合移除成员
	// key: 集合键名
	// members: 要移除的成员（支持多个）
	// 返回: 移除的数量和错误信息
	SRem(key string, members ...interface{}) (int, error)

	// SCard 获取集合成员数量
	// key: 集合键名
	// 返回: 成员数量和错误信息
	SCard(key string) (int64, error)

	// SIsMember 判断成员是否在集合中
	// key: 集合键名
	// member: 要判断的成员
	// 返回: 是否存在和错误信息
	SIsMember(key string, member interface{}) (bool, error)

	// ZAdd 向有序集合添加成员
	// key: 有序集合键名
	// member: 成员值
	// score: 分数
	// 返回: 错误信息
	ZAdd(key string, member string, score float64) error

	// XAdd 向Stream添加消息
	// stream: Stream键名
	// values: 消息字段和值的映射
	// 返回: 消息ID和错误信息
	XAdd(stream string, values map[string]interface{}) (string, error)

	// XReadGroup 从消费者组读取消息
	// group: 消费者组名
	// consumer: 消费者名
	// streams: Stream键名和ID的映射
	// count: 每次读取的最大消息数量
	// 返回: 消息列表和错误信息
	XReadGroup(group, consumer string, streams map[string]string, count int64) ([]map[string]interface{}, error)

	// XAck 确认消息已处理
	// stream: Stream键名
	// group: 消费者组名
	// ids: 要确认的消息ID列表
	// 返回: 确认的消息数量和错误信息
	XAck(stream, group string, ids ...string) (int64, error)

	// XPending 查看待处理消息
	// stream: Stream键名
	// group: 消费者组名
	// 返回: 待处理消息信息和错误信息
	XPending(stream, group string) (map[string]interface{}, error)

	// XDel 删除Stream中的消息
	// stream: Stream键名
	// ids: 要删除的消息ID列表
	// 返回: 删除的消息数量和错误信息
	XDel(stream string, ids ...string) (int64, error)

	// XGroupCreate 创建消费者组
	// stream: Stream键名
	// group: 消费者组名
	// id: 起始ID（"$"表示从最新开始，"0"表示从最开始）
	// 返回: 错误信息
	XGroupCreate(stream, group, id string) error

	// XLen 获取Stream长度
	// stream: Stream键名
	// 返回: Stream长度和错误信息
	XLen(stream string) (int64, error)
}
