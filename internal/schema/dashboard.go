package schema

// DashboardStatsResponse 仪表盘统计响应体
type DashboardStatsResponse struct {
	// 用户统计
	UserStats UserStats `json:"user_stats"` // 用户统计信息
	// 帖子统计
	PostStats PostStats `json:"post_stats"` // 帖子统计信息
	// 评论统计
	CommentStats CommentStats `json:"comment_stats"` // 评论统计信息
	// 版块统计
	CategoryStats CategoryStats `json:"category_stats"` // 版块统计信息
	// 系统统计
	SystemStats SystemStats `json:"system_stats"` // 系统统计信息
}

// UserStats 用户统计
type UserStats struct {
	TotalUsers     int64 `json:"total_users"`     // 总用户数
	ActiveUsers    int64 `json:"active_users"`    // 活跃用户数（30天内登录）
	NewUsers       int64 `json:"new_users"`       // 新增用户数（今日）
	OnlineUsers    int64 `json:"online_users"`    // 在线用户数
	BannedUsers    int64 `json:"banned_users"`    // 被封禁用户数
	ModeratorCount int64 `json:"moderator_count"` // 版主数量
}

// PostStats 帖子统计
type PostStats struct {
	TotalPosts     int64 `json:"total_posts"`     // 总帖子数
	PublishedPosts int64 `json:"published_posts"` // 已发布帖子数
	DraftPosts     int64 `json:"draft_posts"`     // 草稿帖子数
	LockedPosts    int64 `json:"locked_posts"`    // 被锁定帖子数
	EssencePosts   int64 `json:"essence_posts"`   // 精华帖子数
	PinnedPosts    int64 `json:"pinned_posts"`    // 置顶帖子数
	TodayPosts     int64 `json:"today_posts"`     // 今日新增帖子数
}

// CommentStats 评论统计
type CommentStats struct {
	TotalComments    int64 `json:"total_comments"`    // 总评论数
	SelectedComments int64 `json:"selected_comments"` // 精选评论数
	PinnedComments   int64 `json:"pinned_comments"`   // 置顶评论数
	TodayComments    int64 `json:"today_comments"`    // 今日新增评论数
}

// CategoryStats 版块统计
type CategoryStats struct {
	TotalCategories  int64 `json:"total_categories"`  // 总版块数
	ActiveCategories int64 `json:"active_categories"` // 活跃版块数
	HiddenCategories int64 `json:"hidden_categories"` // 隐藏版块数
	LockedCategories int64 `json:"locked_categories"` // 锁定版块数
}

// SystemStats 系统统计
type SystemStats struct {
	TotalViews   int64 `json:"total_views"`   // 总浏览量
	TodayViews   int64 `json:"today_views"`   // 今日浏览量
	TotalLikes   int64 `json:"total_likes"`   // 总点赞数
	TodayLikes   int64 `json:"today_likes"`   // 今日点赞数
	StorageUsed  int64 `json:"storage_used"`  // 存储使用量（字节）
	DatabaseSize int64 `json:"database_size"` // 数据库大小（字节）
}

// RecentActivityResponse 最近活动响应体
type RecentActivityResponse struct {
	RecentPosts    []RecentPost    `json:"recent_posts"`    // 最近帖子
	RecentComments []RecentComment `json:"recent_comments"` // 最近评论
	NewUsers       []NewUser       `json:"new_users"`       // 新用户
}

// RecentPost 最近帖子
type RecentPost struct {
	ID           int    `json:"id" example:"1"`                                  // 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                           // 帖子标题
	Username     string `json:"username" example:"testuser"`                     // 作者用户名
	Avatar       string `json:"avatar" example:"https://example.com/avatar.png"` // 作者头像
	CategoryName string `json:"category_name" example:"技术讨论"`                    // 版块名称
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"`       // 创建时间
}

// RecentComment 最近评论
type RecentComment struct {
	ID        int    `json:"id" example:"1"`                                  // 评论ID
	Content   string `json:"content" example:"很有见地的评论"`                       // 评论内容（截取前100字符）
	Username  string `json:"username" example:"testuser"`                     // 评论者用户名
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"` // 评论者头像
	PostTitle string `json:"post_title" example:"技术分享帖"`                      // 帖子标题
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`       // 创建时间
}

// NewUser 新用户
type NewUser struct {
	ID        int    `json:"id" example:"1"`                                  // 用户ID
	Username  string `json:"username" example:"newuser"`                      // 用户名
	Avatar    string `json:"avatar" example:"https://example.com/avatar.png"` // 用户头像
	Email     string `json:"email" example:"new@example.com"`                 // 邮箱
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`       // 注册时间
}

// PopularPostsResponse 热门帖子响应体
type PopularPostsResponse struct {
	Posts []PopularPost `json:"posts"` // 热门帖子列表
}

// PopularPost 热门帖子
type PopularPost struct {
	ID           int    `json:"id" example:"1"`                            // 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                     // 帖子标题
	Username     string `json:"username" example:"testuser"`               // 作者用户名
	CategoryName string `json:"category_name" example:"技术讨论"`              // 版块名称
	ViewCount    int    `json:"view_count" example:"1500"`                 // 浏览数
	LikeCount    int    `json:"like_count" example:"50"`                   // 点赞数
	CommentCount int    `json:"comment_count" example:"25"`                // 评论数
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
}

// PopularCategoriesResponse 热门版块响应体
type PopularCategoriesResponse struct {
	Categories []PopularCategory `json:"categories"` // 热门版块列表
}

// PopularCategory 热门版块
type PopularCategory struct {
	ID          int    `json:"id" example:"1"`                  // 版块ID
	Name        string `json:"name" example:"技术讨论"`             // 版块名称
	PostCount   int    `json:"post_count" example:"500"`        // 帖子数量
	Description string `json:"description" example:"技术相关话题讨论区"` // 版块描述
}
