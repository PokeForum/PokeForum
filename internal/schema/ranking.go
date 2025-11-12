package schema

// UserRankingListRequest 获取排行榜列表请求
type UserRankingListRequest struct {
	// 排行榜类型：reading(阅读榜), comment(评论榜)
	Type string `json:"type" binding:"required,oneof=reading comment"`
	// 时间范围：all(总榜), month(月榜), week(周榜)
	TimeRange string `json:"time_range" binding:"required,oneof=all month week"`
	// 页码
	Page int `json:"page" binding:"required,min=1" example:"1"`
	// 每页数量
	PageSize int `json:"page_size" binding:"required,min=1,max=100" example:"20"`
}

// UserRankingListResponse 获取排行榜列表响应
type UserRankingListResponse struct {
	// 排行榜类型
	Type string `json:"type"`
	// 时间范围
	TimeRange string `json:"time_range"`
	// 总数量
	Total int `json:"total"`
	// 当前页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 总页数
	TotalPages int `json:"total_pages"`
	// 排行榜项目列表
	Items []RankingItem `json:"items"`
}

// RankingItem 排行榜项目
type RankingItem struct {
	// 排名
	Rank int `json:"rank"`
	// 帖子ID（阅读榜）
	PostID *int `json:"post_id,omitempty"`
	// 帖子标题（阅读榜）
	PostTitle *string `json:"post_title,omitempty"`
	// 用户ID（评论榜）
	UserID *int `json:"user_id,omitempty"`
	// 用户名（评论榜）
	Username *string `json:"username,omitempty"`
	// 统计数值
	Count int `json:"count"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// ReadingRankingItem 阅读榜项目
type ReadingRankingItem struct {
	// 排名
	Rank int `json:"rank"`
	// 帖子ID
	PostID int `json:"post_id"`
	// 帖子标题
	PostTitle string `json:"post_title"`
	// 版块ID
	CategoryID int `json:"category_id"`
	// 版块名称
	CategoryName string `json:"category_name"`
	// 作者用户名
	AuthorUsername string `json:"author_username"`
	// 阅读数
	ViewCount int `json:"view_count"`
	// 点赞数
	LikeCount int `json:"like_count"`
	// 评论数
	CommentCount int `json:"comment_count"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// CommentRankingItem 评论榜项目
type CommentRankingItem struct {
	// 排名
	Rank int `json:"rank"`
	// 用户ID
	UserID int `json:"user_id"`
	// 用户名
	Username string `json:"username"`
	// 头像
	Avatar string `json:"avatar"`
	// 评论总数
	TotalComments int `json:"total_comments"`
	// 获赞总数
	TotalLikes int `json:"total_likes"`
	// 注册时间
	RegisteredAt string `json:"registered_at"`
}
