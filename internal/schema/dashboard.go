package schema

// DashboardStatsResponse Dashboard statistics response | 仪表盘统计响应体
type DashboardStatsResponse struct {
	// User statistics | 用户统计
	UserStats UserStats `json:"user_stats"` // User statistics info | 用户统计信息
	// Post statistics | 帖子统计
	PostStats PostStats `json:"post_stats"` // Post statistics info | 帖子统计信息
	// Comment statistics | 评论统计
	CommentStats CommentStats `json:"comment_stats"` // Comment statistics info | 评论统计信息
	// Category statistics | 版块统计
	CategoryStats CategoryStats `json:"category_stats"` // Category statistics info | 版块统计信息
	// System statistics | 系统统计
	SystemStats SystemStats `json:"system_stats"` // System statistics info | 系统统计信息
}

// UserStats User statistics | 用户统计
type UserStats struct {
	TotalUsers     int64 `json:"total_users"`     // Total users | 总用户数
	ActiveUsers    int64 `json:"active_users"`    // Active users (logged in within 30 days) | 活跃用户数（30天内登录）
	NewUsers       int64 `json:"new_users"`       // New users (today) | 新增用户数（今日）
	OnlineUsers    int64 `json:"online_users"`    // Online users | 在线用户数
	BannedUsers    int64 `json:"banned_users"`    // Banned users | 被封禁用户数
	ModeratorCount int64 `json:"moderator_count"` // Moderator count | 版主数量
}

// PostStats Post statistics | 帖子统计
type PostStats struct {
	TotalPosts     int64 `json:"total_posts"`     // Total posts | 总帖子数
	PublishedPosts int64 `json:"published_posts"` // Published posts | 已发布帖子数
	DraftPosts     int64 `json:"draft_posts"`     // Draft posts | 草稿帖子数
	LockedPosts    int64 `json:"locked_posts"`    // Locked posts | 被锁定帖子数
	EssencePosts   int64 `json:"essence_posts"`   // Essence posts | 精华帖子数
	PinnedPosts    int64 `json:"pinned_posts"`    // Pinned posts | 置顶帖子数
	TodayPosts     int64 `json:"today_posts"`     // New posts today | 今日新增帖子数
}

// CommentStats Comment statistics | 评论统计
type CommentStats struct {
	TotalComments    int64 `json:"total_comments"`    // Total comments | 总评论数
	SelectedComments int64 `json:"selected_comments"` // Selected comments | 精选评论数
	PinnedComments   int64 `json:"pinned_comments"`   // Pinned comments | 置顶评论数
	TodayComments    int64 `json:"today_comments"`    // New comments today | 今日新增评论数
}

// CategoryStats Category statistics | 版块统计
type CategoryStats struct {
	TotalCategories  int64 `json:"total_categories"`  // Total categories | 总版块数
	ActiveCategories int64 `json:"active_categories"` // Active categories | 活跃版块数
	HiddenCategories int64 `json:"hidden_categories"` // Hidden categories | 隐藏版块数
	LockedCategories int64 `json:"locked_categories"` // Locked categories | 锁定版块数
}

// SystemStats System statistics | 系统统计
type SystemStats struct {
	TotalViews   int64 `json:"total_views"`   // Total views | 总浏览量
	TodayViews   int64 `json:"today_views"`   // Today views | 今日浏览量
	TotalLikes   int64 `json:"total_likes"`   // Total likes | 总点赞数
	TodayLikes   int64 `json:"today_likes"`   // Today likes | 今日点赞数
	StorageUsed  int64 `json:"storage_used"`  // Storage used (bytes) | 存储使用量（字节）
	DatabaseSize int64 `json:"database_size"` // Database size (bytes) | 数据库大小（字节）
}

// RecentActivityResponse Recent activity response | 最近活动响应体
type RecentActivityResponse struct {
	RecentPosts    []RecentPost    `json:"recent_posts"`    // Recent posts | 最近帖子
	RecentComments []RecentComment `json:"recent_comments"` // Recent comments | 最近评论
	NewUsers       []NewUser       `json:"new_users"`       // New users | 新用户
}

// RecentPost Recent post | 最近帖子
type RecentPost struct {
	ID           int    `json:"id" example:"1"`                                  // Post ID | 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                           // Post title | 帖子标题
	Username     string `json:"username" example:"testuser"`                     // Author username | 作者用户名
	Avatar       string `json:"avatar" example:"https://example.com/avatar.png"` // Author avatar | 作者头像
	CategoryName string `json:"category_name" example:"技术讨论"`                    // Category name | 版块名称
	CreatedAt    string `json:"created_at" example:"2024-01-01 00:00:00"`        // Creation time | 创建时间
}

// RecentComment Recent comment | 最近评论
type RecentComment struct {
	ID        int    `json:"id" example:"1"`                                  // Comment ID | 评论ID
	Content   string `json:"content" example:"很有见地的评论"`                       // Comment content (truncated to first 100 characters) | 评论内容（截取前100字符）
	Username  string `json:"username" example:"testuser"`                     // Commenter username | 评论者用户名
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"` // Commenter avatar | 评论者头像
	PostTitle string `json:"post_title" example:"技术分享帖"`                      // Post title | 帖子标题
	CreatedAt string `json:"created_at" example:"2024-01-01 00:00:00"`        // Creation time | 创建时间
}

// NewUser New user | 新用户
type NewUser struct {
	ID        int    `json:"id" example:"1"`                                  // User ID | 用户ID
	Username  string `json:"username" example:"newuser"`                      // Username | 用户名
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"` // User avatar | 用户头像
	Email     string `json:"email" example:"new@example.com"`                 // Email | 邮箱
	CreatedAt string `json:"created_at" example:"2024-01-01 00:00:00"`        // Registration time | 注册时间
}

// PopularPostsResponse Popular posts response | 热门帖子响应体
type PopularPostsResponse struct {
	Posts []PopularPost `json:"posts"` // Popular posts list | 热门帖子列表
}

// PopularPost Popular post | 热门帖子
type PopularPost struct {
	ID           int    `json:"id" example:"1"`                                  // Post ID | 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                           // Post title | 帖子标题
	Username     string `json:"username" example:"testuser"`                     // Author username | 作者用户名
	Avatar       string `json:"avatar" example:"https://example.com/avatar.png"` // Author avatar | 作者头像
	CategoryName string `json:"category_name" example:"技术讨论"`                    // Category name | 版块名称
	ViewCount    int    `json:"view_count" example:"1500"`                       // View count | 浏览数
	LikeCount    int    `json:"like_count" example:"50"`                         // Like count | 点赞数
	CommentCount int    `json:"comment_count" example:"25"`                      // Comment count | 评论数
	CreatedAt    string `json:"created_at" example:"2024-01-01 00:00:00"`        // Creation time | 创建时间
}

// PopularCategoriesResponse Popular categories response | 热门版块响应体
type PopularCategoriesResponse struct {
	Categories []PopularCategory `json:"categories"` // Popular categories list | 热门版块列表
}

// PopularCategory Popular category | 热门版块
type PopularCategory struct {
	ID          int    `json:"id" example:"1"`                  // Category ID | 版块ID
	Name        string `json:"name" example:"技术讨论"`             // Category name | 版块名称
	PostCount   int    `json:"post_count" example:"500"`        // Post count | 帖子数量
	Description string `json:"description" example:"技术相关话题讨论区"` // Category description | 版块描述
}
