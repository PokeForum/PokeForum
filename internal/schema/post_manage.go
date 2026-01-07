package schema

// PostListRequest Post list query request | 帖子列表查询请求体
type PostListRequest struct {
	Page       int    `form:"page" binding:"required,min=1" example:"1"`               // Page number | 页码
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // Items per page | 每页数量
	Keyword    string `form:"keyword" example:"技术"`                                    // Search keyword (title or content) | 搜索关键词（标题或内容）
	Status     string `form:"status" example:"Normal"`                                 // Post status filter | 帖子状态筛选
	CategoryID int    `form:"category_id" example:"1"`                                 // Category ID filter | 版块ID筛选
	UserID     int    `form:"user_id" example:"1"`                                     // User ID filter | 用户ID筛选
	IsEssence  *bool  `form:"is_essence" example:"true"`                               // Essence post filter | 是否精华帖筛选
	IsPinned   *bool  `form:"is_pinned" example:"true"`                                // Pinned post filter | 是否置顶筛选
}

// PostCreateRequest Create post request | 创建帖子请求体
type PostCreateRequest struct {
	UserID         int    `json:"user_id" binding:"required" example:"1"`                                           // User ID | 用户ID
	CategoryID     int    `json:"category_id" binding:"required" example:"1"`                                       // Category ID | 版块ID
	Title          string `json:"title" binding:"required,min=2,max=200" example:"技术分享帖"`                           // Post title | 帖子标题
	Content        string `json:"content" binding:"required,min=10" example:"## 技术分享\n这是内容"`                        // Post content | 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`                                                  // Read permission | 阅读限制
	PublishIP      string `json:"publish_ip" example:"192.168.1.1"`                                                 // Publish IP | 发布IP
	Status         string `json:"status" binding:"required,oneof=Normal Locked Draft Private Ban" example:"Normal"` // Post status | 帖子状态
}

// PostUpdateRequest Update post request | 更新帖子请求体
type PostUpdateRequest struct {
	ID             int    `json:"id" binding:"required" example:"1"`                                                 // Post ID | 帖子ID
	Title          string `json:"title" binding:"omitempty,min=2,max=200" example:"技术分享帖"`                           // Post title | 帖子标题
	Content        string `json:"content" binding:"omitempty,min=10" example:"## 技术分享\n这是内容"`                        // Post content | 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`                                                   // Read permission | 阅读限制
	Status         string `json:"status" binding:"omitempty,oneof=Normal Locked Draft Private Ban" example:"Normal"` // Post status | 帖子状态
}

// PostStatusUpdateRequest Update post status request | 更新帖子状态请求体
type PostStatusUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                                // Post ID | 帖子ID
	Status string `json:"status" binding:"required,oneof=Normal Locked Draft Private Ban" example:"Normal"` // Post status | 帖子状态
	Reason string `json:"reason" example:"违规内容"`                                                            // Operation reason | 操作原因
}

// PostEssenceUpdateRequest Set post essence request | 设置帖子精华请求体
type PostEssenceUpdateRequest struct {
	ID        int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	IsEssence bool   `json:"is_essence" example:"true"`         // Whether essence | 是否精华
	Reason    string `json:"reason" example:"优质内容"`             // Operation reason | 操作原因
}

// PostPinUpdateRequest Set post pinned request | 设置帖子置顶请求体
type PostPinUpdateRequest struct {
	ID       int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	IsPinned bool   `json:"is_pinned" example:"true"`          // Whether pinned | 是否置顶
	Reason   string `json:"reason" example:"重要公告"`             // Operation reason | 操作原因
}

// PostMoveRequest Move post to another category request | 移动帖子到其他版块请求体
type PostMoveRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"`          // Post ID | 帖子ID
	CategoryID int    `json:"category_id" binding:"required" example:"2"` // Target category ID | 目标版块ID
	Reason     string `json:"reason" example:"内容更适合该版块"`                  // Operation reason | 操作原因
}

// PostListItem Post list item response | 帖子列表项响应体
type PostListItem struct {
	ID            int    `json:"id" example:"1"`                           // Post ID | 帖子ID
	UserID        int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username      string `json:"username" example:"testuser"`              // Username | 用户名
	CategoryID    int    `json:"category_id" example:"1"`                  // Category ID | 版块ID
	CategoryName  string `json:"category_name" example:"技术讨论"`             // Category name | 版块名称
	Title         string `json:"title" example:"技术分享帖"`                    // Post title | 帖子标题
	Content       string `json:"content" example:"## 技术分享\n这是内容"`          // Post content (first 100 characters) | 帖子内容（截取前100字符）
	ViewCount     int    `json:"view_count" example:"150"`                 // View count | 浏览数
	LikeCount     int    `json:"like_count" example:"25"`                  // Like count | 点赞数
	DislikeCount  int    `json:"dislike_count" example:"2"`                // Dislike count | 点踩数
	FavoriteCount int    `json:"favorite_count" example:"10"`              // Favorite count | 收藏数
	IsEssence     bool   `json:"is_essence" example:"true"`                // Whether essence post | 是否精华帖
	IsPinned      bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	Status        string `json:"status" example:"Normal"`                  // Post status | 帖子状态
	PublishIP     string `json:"publish_ip" example:"192.168.1.1"`         // Publish IP | 发布IP
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt     string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}

// PostListResponse Post list response | 帖子列表响应体
type PostListResponse struct {
	List     []PostListItem `json:"list"`      // Post list | 帖子列表
	Total    int64          `json:"total"`     // Total count | 总数量
	Page     int            `json:"page"`      // Current page number | 当前页码
	PageSize int            `json:"page_size"` // Items per page | 每页数量
}

// PostDetailResponse Post detail response | 帖子详情响应体
type PostDetailResponse struct {
	ID             int    `json:"id" example:"1"`                           // Post ID | 帖子ID
	UserID         int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username       string `json:"username" example:"testuser"`              // Username | 用户名
	CategoryID     int    `json:"category_id" example:"1"`                  // Category ID | 版块ID
	CategoryName   string `json:"category_name" example:"技术讨论"`             // Category name | 版块名称
	Title          string `json:"title" example:"技术分享帖"`                    // Post title | 帖子标题
	Content        string `json:"content" example:"## 技术分享\n这是内容"`          // Post content | 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`          // Read permission | 阅读限制
	ViewCount      int    `json:"view_count" example:"150"`                 // View count | 浏览数
	LikeCount      int    `json:"like_count" example:"25"`                  // Like count | 点赞数
	DislikeCount   int    `json:"dislike_count" example:"2"`                // Dislike count | 点踩数
	FavoriteCount  int    `json:"favorite_count" example:"10"`              // Favorite count | 收藏数
	IsEssence      bool   `json:"is_essence" example:"true"`                // Whether essence post | 是否精华帖
	IsPinned       bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	Status         string `json:"status" example:"Normal"`                  // Post status | 帖子状态
	PublishIP      string `json:"publish_ip" example:"192.168.1.1"`         // Publish IP | 发布IP
	CreatedAt      string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt      string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}
