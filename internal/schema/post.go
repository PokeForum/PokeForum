package schema

// UserPostCreateRequest 创建帖子请求
type UserPostCreateRequest struct {
	// 版块ID
	CategoryID int `json:"category_id" binding:"required"`
	// 帖子标题
	Title string `json:"title" binding:"required,min=1,max=200"`
	// 帖子内容
	Content string `json:"content" binding:"required,min=1"`
	// 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
}

// UserPostCreateResponse 创建帖子响应
type UserPostCreateResponse struct {
	// 帖子ID
	ID int `json:"id"`
	// 版块ID
	CategoryID int `json:"category_id"`
	// 版块名称
	CategoryName string `json:"category_name"`
	// 帖子标题
	Title string `json:"title"`
	// 帖子内容
	Content string `json:"content"`
	// 作者用户名
	Username string `json:"username"`
	// 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// 浏览数
	ViewCount int `json:"view_count"`
	// 点赞数
	LikeCount int `json:"like_count"`
	// 点踩数
	DislikeCount int `json:"dislike_count"`
	// 收藏数
	FavoriteCount int `json:"favorite_count"`
	// 是否精华帖
	IsEssence bool `json:"is_essence"`
	// 是否置顶
	IsPinned bool `json:"is_pinned"`
	// 帖子状态
	Status string `json:"status"`
	// 创建时间
	CreatedAt string `json:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at"`
}

// UserPostUpdateRequest 更新帖子请求
type UserPostUpdateRequest struct {
	// 帖子ID
	ID int `json:"id" binding:"required"`
	// 帖子标题
	Title string `json:"title" binding:"required,min=1,max=200"`
	// 帖子内容
	Content string `json:"content" binding:"required,min=1"`
	// 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
}

// UserPostUpdateResponse 更新帖子响应
type UserPostUpdateResponse struct {
	// 帖子ID
	ID int `json:"id"`
	// 版块ID
	CategoryID int `json:"category_id"`
	// 版块名称
	CategoryName string `json:"category_name"`
	// 帖子标题
	Title string `json:"title"`
	// 帖子内容
	Content string `json:"content"`
	// 作者用户名
	Username string `json:"username"`
	// 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// 浏览数
	ViewCount int `json:"view_count"`
	// 点赞数
	LikeCount int `json:"like_count"`
	// 点踩数
	DislikeCount int `json:"dislike_count"`
	// 收藏数
	FavoriteCount int `json:"favorite_count"`
	// 是否精华帖
	IsEssence bool `json:"is_essence"`
	// 是否置顶
	IsPinned bool `json:"is_pinned"`
	// 帖子状态
	Status string `json:"status"`
	// 创建时间
	CreatedAt string `json:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at"`
}

// UserPostActionRequest 帖子操作请求
type UserPostActionRequest struct {
	// 帖子ID
	ID int `json:"id" binding:"required"`
}

// UserPostActionResponse 帖子操作响应
type UserPostActionResponse struct {
	// 操作成功
	Success bool `json:"success"`
	// 当前点赞数
	LikeCount int `json:"like_count"`
	// 当前点踩数
	DislikeCount int `json:"dislike_count"`
	// 当前收藏数
	FavoriteCount int `json:"favorite_count"`
	// 操作类型
	ActionType string `json:"action_type"`
}

// UserPostListRequest 帖子列表请求
type UserPostListRequest struct {
	// 版块ID，可选
	CategoryID int `json:"category_id,omitempty"`
	// 页码，默认1
	Page int `json:"page" binding:"min=1"`
	// 每页数量，默认20，最大100
	PageSize int `json:"page_size" binding:"min=1,max=100"`
	// 排序方式：latest(最新)、hot(热门)、essence(精华)
	Sort string `json:"sort" binding:"omitempty,oneof=latest hot essence"`
}

// UserPostListResponse 帖子列表响应
type UserPostListResponse struct {
	// 帖子列表
	Posts []UserPostCreateResponse `json:"posts"`
	// 总数
	Total int `json:"total"`
	// 当前页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 总页数
	TotalPages int `json:"total_pages"`
}

// UserPostDetailRequest 帖子详情请求
type UserPostDetailRequest struct {
	// 帖子ID
	ID int `uri:"id" binding:"required"`
}

// UserPostDetailResponse 帖子详情响应
type UserPostDetailResponse struct {
	// 帖子ID
	ID int `json:"id"`
	// 版块ID
	CategoryID int `json:"category_id"`
	// 版块名称
	CategoryName string `json:"category_name"`
	// 帖子标题
	Title string `json:"title"`
	// 帖子内容
	Content string `json:"content"`
	// 作者用户名
	Username string `json:"username"`
	// 阅读限制
	ReadPermission string `json:"read_permission,omitempty"`
	// 浏览数
	ViewCount int `json:"view_count"`
	// 点赞数
	LikeCount int `json:"like_count"`
	// 点踩数
	DislikeCount int `json:"dislike_count"`
	// 收藏数
	FavoriteCount int `json:"favorite_count"`
	// 是否精华帖
	IsEssence bool `json:"is_essence"`
	// 是否置顶
	IsPinned bool `json:"is_pinned"`
	// 帖子状态
	Status string `json:"status"`
	// 创建时间
	CreatedAt string `json:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at"`
}
