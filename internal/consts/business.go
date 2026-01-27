package _const

const (
	DefaultPage     = 1        // Default page number | 默认页码
	DefaultPageSize = 20       // Default page size | 默认每页数量
	DefaultSort     = "latest" // Default sort order | 默认排序方式

	MaxDraftsPerUser        = 10   // Maximum drafts per user | 每个用户最大草稿数
	MaxContentPreviewLength = 100  // Maximum content preview length | 内容预览最大长度
	MaxPostTitleLength      = 100  // Maximum post title length | 帖子标题最大长度
	MaxCommentContentLength = 1000 // Maximum comment content length | 评论内容最大长度
	MaxEmailLength          = 100  // Maximum email length | 邮箱最大长度
	MaxUsernameLength       = 20   // Maximum username length | 用户名最大长度
	MaxPasswordLength       = 100  // Maximum password length | 密码最大长度

	VerificationCodeExpiryMinutes = 10 // Verification code expiry time in minutes | 验证码有效期（分钟）

	DefaultRankingLimit = 10  // Default ranking list limit | 默认排行榜数量限制
	MaxRankingLimit     = 100 // Maximum ranking list limit | 排行榜最大数量限制

	EditCooldownMinutes = 3 // Post edit cooldown in minutes | 帖子编辑冷却时间（分钟）
	PrivateCooldownDays = 3 // Post private cooldown in days | 帖子私有化冷却时间（天）

	ReadHeaderTimeoutSeconds = 10 // HTTP read header timeout in seconds | HTTP读取头部超时时间（秒）

	DefaultTimeWindowSeconds = 1   // Default rate limit time window in seconds | 默认限流时间窗口（秒）
	DefaultMaxRequests       = 100 // Default max requests per window | 默认每个窗口最大请求数
	APITimeWindowSeconds     = 60  // API rate limit time window in seconds | API限流时间窗口（秒）
	APIMaxRequests           = 60  // API max requests per window | API每个窗口最大请求数
	AuthTimeWindowSeconds    = 60  // Auth rate limit time window in seconds | 认证限流时间窗口（秒）
	AuthMaxRequests          = 10  // Auth max requests per window | 认证每个窗口最大请求数

	RetryAfterMultiplier = 2 // Retry after time multiplier | 重试时间倍数

	TaskManagerRetryInterval = 100 // Task manager retry interval in milliseconds | 任务管理器重试间隔（毫秒）
)
