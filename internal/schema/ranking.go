package schema

// UserRankingListRequest Get ranking list request | 获取排行榜列表请求
type UserRankingListRequest struct {
	// Ranking type: reading (reading ranking), comment (comment ranking) | 排行榜类型：reading(阅读榜), comment(评论榜)
	Type string `json:"type" binding:"required,oneof=reading comment"`
	// Time range: all (overall ranking), month (monthly ranking), week (weekly ranking) | 时间范围：all(总榜), month(月榜), week(周榜)
	TimeRange string `json:"time_range" binding:"required,oneof=all month week"`
	// Page number | 页码
	Page int `json:"page" binding:"required,min=1" example:"1"`
	// Items per page | 每页数量
	PageSize int `json:"page_size" binding:"required,min=1,max=100" example:"20"`
}

// UserRankingListResponse Get ranking list response | 获取排行榜列表响应
type UserRankingListResponse struct {
	// Ranking type | 排行榜类型
	Type string `json:"type"`
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Total count | 总数量
	Total int `json:"total"`
	// Current page number | 当前页码
	Page int `json:"page"`
	// Items per page | 每页数量
	PageSize int `json:"page_size"`
	// Total pages | 总页数
	TotalPages int `json:"total_pages"`
	// Ranking item list | 排行榜项目列表
	Items []RankingItem `json:"items"`
}

// RankingItem Ranking item | 排行榜项目
type RankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// Post ID (reading ranking) | 帖子ID（阅读榜）
	PostID *int `json:"post_id,omitempty"`
	// Post title (reading ranking) | 帖子标题（阅读榜）
	PostTitle *string `json:"post_title,omitempty"`
	// User ID (comment ranking) | 用户ID（评论榜）
	UserID *int `json:"user_id,omitempty"`
	// Username (comment ranking) | 用户名（评论榜）
	Username *string `json:"username,omitempty"`
	// Statistical value | 统计数值
	Count int `json:"count"`
	// Created time | 创建时间
	CreatedAt string `json:"created_at"`
}

// ReadingRankingItem Reading ranking item | 阅读榜项目
type ReadingRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// Post ID | 帖子ID
	PostID int `json:"post_id"`
	// Post title | 帖子标题
	PostTitle string `json:"post_title"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Author username | 作者用户名
	AuthorUsername string `json:"author_username"`
	// View count | 阅读数
	ViewCount int `json:"view_count"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Comment count | 评论数
	CommentCount int `json:"comment_count"`
	// Created time | 创建时间
	CreatedAt string `json:"created_at"`
}

// CommentRankingItem Comment ranking item | 评论榜项目
type CommentRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Total comments | 评论总数
	TotalComments int `json:"total_comments"`
	// Total likes | 获赞总数
	TotalLikes int `json:"total_likes"`
	// Registration time | 注册时间
	RegisteredAt string `json:"registered_at"`
}
