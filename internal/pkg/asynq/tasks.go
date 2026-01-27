package asynq

// 任务类型常量 | Task type constants
const (
	// TypeSigninPersist Sign-in data persistence task | 签到数据持久化任务
	TypeSigninPersist = "signin:persist"

	// TypeStatsSync Statistics data synchronization task | 统计数据同步任务
	TypeStatsSync = "stats:sync"

	// TypeRankingRefresh 排行榜缓存刷新任务
	TypeRankingRefresh = "ranking:refresh"
)

// 队列名称常量 | Queue name constants
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)
