package schema

// UserPostCreateRequest Create post request | 创建帖子请求
type UserPostCreateRequest struct {
	// Post ID (optional, for updating draft) | 帖子ID（可选，用于更新草稿）
	ID int `json:"id,omitempty"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id" binding:"required"`
	// Post title | 帖子标题
	Title string `json:"title" binding:"required,min=1,max=200"`
	// Post content | 帖子内容
	Content string `json:"content" binding:"required,min=1"`
	// Read permission | 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
}

// UserPostCreateResponse Create post response | 创建帖子响应
type UserPostCreateResponse struct {
	// Post ID | 帖子ID
	ID int `json:"id"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Post title | 帖子标题
	Title string `json:"title"`
	// Post content | 帖子内容
	Content string `json:"content"`
	// Author username | 作者用户名
	Username string `json:"username"`
	// Read permission | 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// View count | 浏览数
	ViewCount int `json:"view_count"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Dislike count | 点踩数
	DislikeCount int `json:"dislike_count"`
	// Favorite count | 收藏数
	FavoriteCount int `json:"favorite_count"`
	// Whether current user has liked | 当前用户是否已点赞
	UserLiked bool `json:"user_liked"`
	// Whether current user has disliked | 当前用户是否已点踩
	UserDisliked bool `json:"user_disliked"`
	// Whether current user has favorited | 当前用户是否已收藏
	UserFavorited bool `json:"user_favorited"`
	// Whether essence post | 是否精华帖
	IsEssence bool `json:"is_essence"`
	// Whether pinned | 是否置顶
	IsPinned bool `json:"is_pinned"`
	// Post status | 帖子状态
	Status string `json:"status"`
	// Creation time | 创建时间
	CreatedAt string `json:"created_at"`
	// Update time | 更新时间
	UpdatedAt string `json:"updated_at"`
}

// UserPostUpdateRequest Update post request | 更新帖子请求
type UserPostUpdateRequest struct {
	// Post ID | 帖子ID
	ID int `json:"id" binding:"required"`
	// Post title | 帖子标题
	Title string `json:"title" binding:"required,min=1,max=200"`
	// Post content | 帖子内容
	Content string `json:"content" binding:"required,min=1"`
	// Read permission | 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
}

// UserPostUpdateResponse Update post response | 更新帖子响应
type UserPostUpdateResponse struct {
	// Post ID | 帖子ID
	ID int `json:"id"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Post title | 帖子标题
	Title string `json:"title"`
	// Post content | 帖子内容
	Content string `json:"content"`
	// Author username | 作者用户名
	Username string `json:"username"`
	// Read permission | 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// View count | 浏览数
	ViewCount int `json:"view_count"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Dislike count | 点踩数
	DislikeCount int `json:"dislike_count"`
	// Favorite count | 收藏数
	FavoriteCount int `json:"favorite_count"`
	// Whether current user has liked | 当前用户是否已点赞
	UserLiked bool `json:"user_liked"`
	// Whether current user has disliked | 当前用户是否已点踩
	UserDisliked bool `json:"user_disliked"`
	// Whether current user has favorited | 当前用户是否已收藏
	UserFavorited bool `json:"user_favorited"`
	// Whether essence post | 是否精华帖
	IsEssence bool `json:"is_essence"`
	// Whether pinned | 是否置顶
	IsPinned bool `json:"is_pinned"`
	// Post status | 帖子状态
	Status string `json:"status"`
	// Creation time | 创建时间
	CreatedAt string `json:"created_at"`
	// Update time | 更新时间
	UpdatedAt string `json:"updated_at"`
}

// UserPostActionRequest Post action request | 帖子操作请求
type UserPostActionRequest struct {
	// Post ID | 帖子ID
	ID int `json:"id" binding:"required"`
}

// UserPostActionResponse Post action response | 帖子操作响应
type UserPostActionResponse struct {
	// Operation success | 操作成功
	Success bool `json:"success"`
	// Current like count | 当前点赞数
	LikeCount int `json:"like_count"`
	// Current dislike count | 当前点踩数
	DislikeCount int `json:"dislike_count"`
	// Current favorite count | 当前收藏数
	FavoriteCount int `json:"favorite_count"`
	// Action type | 操作类型
	ActionType string `json:"action_type"`
}

// UserPostListRequest Post list request | 帖子列表请求
type UserPostListRequest struct {
	// Category ID, optional | 版块ID，可选
	CategoryID int `form:"category_id,omitempty"`
	// Category slug, optional | 版块slug，可选
	Slug string `form:"slug,omitempty"`
	// Keyword for title search, optional | 标题关键词搜索，可选
	Keyword string `form:"keyword,omitempty"`
	// Page number, default 1 | 页码，默认1
	Page int `form:"page" binding:"min=1"`
	// Items per page, default 20, max 100 | 每页数量，默认20，最大100
	PageSize int `form:"page_size" binding:"min=1,max=100"`
	// Sort method: latest (newest), hot (popular), essence (featured) | 排序方式：latest(最新)、hot(热门)、essence(精华)
	Sort string `form:"sort" binding:"omitempty,oneof=latest hot essence"`
}

// UserPostListResponse Post list response | 帖子列表响应
type UserPostListResponse struct {
	// Post list | 帖子列表
	Posts []UserPostCreateResponse `json:"posts"`
	// Total count | 总数
	Total int `json:"total"`
	// Current page number | 当前页码
	Page int `json:"page"`
	// Items per page | 每页数量
	PageSize int `json:"page_size"`
	// Total pages | 总页数
	TotalPages int `json:"total_pages"`
}

// UserPostDetailRequest Post detail request | 帖子详情请求
type UserPostDetailRequest struct {
	// Post ID | 帖子ID
	ID int `uri:"id" binding:"required"`
}

// UserPostDetailResponse Post detail response | 帖子详情响应
type UserPostDetailResponse struct {
	// Post ID | 帖子ID
	ID int `json:"id"`
	// Category ID | 版块ID
	CategoryID int `json:"category_id"`
	// Category name | 版块名称
	CategoryName string `json:"category_name"`
	// Post title | 帖子标题
	Title string `json:"title"`
	// Post content | 帖子内容
	Content string `json:"content"`
	// Author ID | 作者 ID
	UserID int `json:"user_id"`
	// Author username | 作者用户名
	Username string `json:"username"`
	// Read permission | 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// View count | 浏览数
	ViewCount int `json:"view_count"`
	// Like count | 点赞数
	LikeCount int `json:"like_count"`
	// Dislike count | 点踩数
	DislikeCount int `json:"dislike_count"`
	// Favorite count | 收藏数
	FavoriteCount int `json:"favorite_count"`
	// Whether current user has liked | 当前用户是否已点赞
	UserLiked bool `json:"user_liked"`
	// Whether current user has disliked | 当前用户是否已点踩
	UserDisliked bool `json:"user_disliked"`
	// Whether current user has favorited | 当前用户是否已收藏
	UserFavorited bool `json:"user_favorited"`
	// Whether essence post | 是否精华帖
	IsEssence bool `json:"is_essence"`
	// Whether pinned | 是否置顶
	IsPinned bool `json:"is_pinned"`
	// Post status | 帖子状态
	Status string `json:"status"`
	// Creation time | 创建时间
	CreatedAt string `json:"created_at"`
	// Update time | 更新时间
	UpdatedAt string `json:"updated_at"`
}

// UserDraftListRequest Draft list request | 草稿列表请求
type UserDraftListRequest struct {
	// Page number, default 1 | 页码，默认1
	Page int `form:"page" binding:"min=1"`
	// Items per page, default 20, max 100 | 每页数量，默认20，最大100
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// UserDraftDeleteRequest Delete draft request | 删除草稿请求
type UserDraftDeleteRequest struct {
	// Draft ID | 草稿ID
	ID int `json:"id" binding:"required"`
}
