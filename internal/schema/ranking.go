package schema

// RankingRequest 排行榜通用请求参数
type RankingRequest struct {
	// Time range: all (overall ranking), month (monthly ranking), week (weekly ranking) | 时间范围：all(总榜), month(月榜), week(周榜)
	TimeRange string `json:"time_range" form:"time_range" binding:"required,oneof=all month week"`
}

// ReadingRankingResponse 阅读榜响应
type ReadingRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []ReadingRankingItem `json:"items"`
}

// ReadingRankingItem 阅读榜项目
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
	// Created time | 创建时间
	CreatedAt string `json:"created_at"`
}

// CommentRankingResponse 评论榜响应
type CommentRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []CommentRankingItem `json:"items"`
}

// CommentRankingItem 评论榜项目
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

// PostCountRankingResponse 帖子数排行榜响应
type PostCountRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []PostCountRankingItem `json:"items"`
}

// PostCountRankingItem 帖子数排行榜项目
type PostCountRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Total posts | 帖子总数
	TotalPosts int `json:"total_posts"`
	// Total views | 总阅读数
	TotalViews int `json:"total_views"`
	// Registration time | 注册时间
	RegisteredAt string `json:"registered_at"`
}

// CommentCountRankingResponse 评论数排行榜响应
type CommentCountRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []CommentCountRankingItem `json:"items"`
}

// CommentCountRankingItem 评论数排行榜项目
type CommentCountRankingItem struct {
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

// FollowerRankingResponse 名人榜（被关注数）响应
type FollowerRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []FollowerRankingItem `json:"items"`
}

// FollowerRankingItem 名人榜项目
type FollowerRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Total followers | 粉丝总数
	TotalFollowers int `json:"total_followers"`
	// Total posts | 帖子总数
	TotalPosts int `json:"total_posts"`
	// Registration time | 注册时间
	RegisteredAt string `json:"registered_at"`
}

// PointsRankingResponse 积分榜响应
type PointsRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []PointsRankingItem `json:"items"`
}

// PointsRankingItem 积分榜项目
type PointsRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Points | 积分
	Points int `json:"points"`
	// Experience | 经验值
	Experience int `json:"experience"`
	// Registration time | 注册时间
	RegisteredAt string `json:"registered_at"`
}

// CurrencyRankingResponse 财富榜（货币）响应
type CurrencyRankingResponse struct {
	// Time range | 时间范围
	TimeRange string `json:"time_range"`
	// Ranking item list | 排行榜项目列表
	Items []CurrencyRankingItem `json:"items"`
}

// CurrencyRankingItem 财富榜项目
type CurrencyRankingItem struct {
	// Rank | 排名
	Rank int `json:"rank"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Currency | 货币
	Currency int `json:"currency"`
	// Points | 积分
	Points int `json:"points"`
	// Registration time | 注册时间
	RegisteredAt string `json:"registered_at"`
}
