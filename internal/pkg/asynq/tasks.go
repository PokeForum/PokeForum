package asynq

// 任务类型常量
const (
	// TypeSigninPersist 签到数据持久化任务
	TypeSigninPersist = "signin:persist"

	// TypeStatsSync 统计数据同步任务
	TypeStatsSync = "stats:sync"
)

// 队列名称常量
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)
