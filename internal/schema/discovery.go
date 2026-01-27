package schema

// DiscoveryPostItem 发现页帖子项
type DiscoveryPostItem struct {
	// Post ID | 帖子ID
	ID int `json:"id"`
	// Post title | 帖子标题
	Title string `json:"title"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Author ID | 作者ID
	UserID int `json:"user_id"`
	// Author username | 作者用户名
	Username string `json:"username"`
	// Author avatar | 作者头像
	Avatar string `json:"avatar"`
	// View count | 浏览数
	ViewCount int `json:"view_count"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Comment count | 评论数
	CommentCount int `json:"comment_count"`
	// Whether essence post | 是否精华帖
	IsEssence bool `json:"is_essence"`
	// Creation time | 创建时间
	CreatedAt string `json:"created_at"`
	// Update time | 更新时间
	UpdatedAt string `json:"updated_at"`
}

// DiscoveryFreshResponse 新鲜发布响应（最新发布的帖子）
type DiscoveryFreshResponse struct {
	// Post list | 帖子列表
	Posts []DiscoveryPostItem `json:"posts"`
}

// DiscoveryLatestDiscussionResponse 最新讨论响应（按最近被用户回复的帖子顺序）
type DiscoveryLatestDiscussionResponse struct {
	// Post list | 帖子列表
	Posts []DiscoveryLatestDiscussionItem `json:"posts"`
}

// DiscoveryLatestDiscussionItem 最新讨论帖子项
type DiscoveryLatestDiscussionItem struct {
	// Post ID | 帖子ID
	ID int `json:"id"`
	// Post title | 帖子标题
	Title string `json:"title"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Author ID | 作者ID
	UserID int `json:"user_id"`
	// Author username | 作者用户名
	Username string `json:"username"`
	// Author avatar | 作者头像
	Avatar string `json:"avatar"`
	// Comment count | 评论数
	CommentCount int `json:"comment_count"`
	// Last reply user ID | 最后回复用户ID
	LastReplyUserID int `json:"last_reply_user_id"`
	// Last reply username | 最后回复用户名
	LastReplyUsername string `json:"last_reply_username"`
	// Last reply time | 最后回复时间
	LastReplyAt string `json:"last_reply_at"`
}

// DiscoveryCommentItem 互动评论项
type DiscoveryCommentItem struct {
	// Comment ID | 评论ID
	ID int `json:"id"`
	// Comment content | 评论内容
	Content string `json:"content"`
	// Post ID | 帖子ID
	PostID int `json:"post_id"`
	// Post title | 帖子标题
	PostTitle string `json:"post_title"`
	// User ID | 用户ID
	UserID int `json:"user_id"`
	// Username | 用户名
	Username string `json:"username"`
	// Avatar | 头像
	Avatar string `json:"avatar"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Creation time | 创建时间
	CreatedAt string `json:"created_at"`
}

// DiscoveryCommentsResponse 互动评论响应
type DiscoveryCommentsResponse struct {
	// Comment list | 评论列表
	Comments []DiscoveryCommentItem `json:"comments"`
}
