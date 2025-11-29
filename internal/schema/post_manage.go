package schema

// PostListRequest 帖子列表查询请求体
type PostListRequest struct {
	Page       int    `form:"page" binding:"required,min=1" example:"1"`               // 页码
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // 每页数量
	Keyword    string `form:"keyword" example:"技术"`                                    // 搜索关键词（标题或内容）
	Status     string `form:"status" example:"Normal"`                                 // 帖子状态筛选
	CategoryID int    `form:"category_id" example:"1"`                                 // 版块ID筛选
	UserID     int    `form:"user_id" example:"1"`                                     // 用户ID筛选
	IsEssence  *bool  `form:"is_essence" example:"true"`                               // 是否精华帖筛选
	IsPinned   *bool  `form:"is_pinned" example:"true"`                                // 是否置顶筛选
}

// PostCreateRequest 创建帖子请求体
type PostCreateRequest struct {
	UserID         int    `json:"user_id" binding:"required" example:"1"`                                           // 用户ID
	CategoryID     int    `json:"category_id" binding:"required" example:"1"`                                       // 版块ID
	Title          string `json:"title" binding:"required,min=2,max=200" example:"技术分享帖"`                           // 帖子标题
	Content        string `json:"content" binding:"required,min=10" example:"## 技术分享\n这是内容"`                        // 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`                                                  // 阅读限制
	PublishIP      string `json:"publish_ip" example:"192.168.1.1"`                                                 // 发布IP
	Status         string `json:"status" binding:"required,oneof=Normal Locked Draft Private Ban" example:"Normal"` // 帖子状态
}

// PostUpdateRequest 更新帖子请求体
type PostUpdateRequest struct {
	ID             int    `json:"id" binding:"required" example:"1"`                                                 // 帖子ID
	Title          string `json:"title" binding:"omitempty,min=2,max=200" example:"技术分享帖"`                           // 帖子标题
	Content        string `json:"content" binding:"omitempty,min=10" example:"## 技术分享\n这是内容"`                        // 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`                                                   // 阅读限制
	Status         string `json:"status" binding:"omitempty,oneof=Normal Locked Draft Private Ban" example:"Normal"` // 帖子状态
}

// PostStatusUpdateRequest 更新帖子状态请求体
type PostStatusUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                                // 帖子ID
	Status string `json:"status" binding:"required,oneof=Normal Locked Draft Private Ban" example:"Normal"` // 帖子状态
	Reason string `json:"reason" example:"违规内容"`                                                            // 操作原因
}

// PostEssenceUpdateRequest 设置帖子精华请求体
type PostEssenceUpdateRequest struct {
	ID        int    `json:"id" binding:"required" example:"1"` // 帖子ID
	IsEssence bool   `json:"is_essence" example:"true"`         // 是否精华
	Reason    string `json:"reason" example:"优质内容"`             // 操作原因
}

// PostPinUpdateRequest 设置帖子置顶请求体
type PostPinUpdateRequest struct {
	ID       int    `json:"id" binding:"required" example:"1"` // 帖子ID
	IsPinned bool   `json:"is_pinned" example:"true"`          // 是否置顶
	Reason   string `json:"reason" example:"重要公告"`             // 操作原因
}

// PostMoveRequest 移动帖子到其他版块请求体
type PostMoveRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"`          // 帖子ID
	CategoryID int    `json:"category_id" binding:"required" example:"2"` // 目标版块ID
	Reason     string `json:"reason" example:"内容更适合该版块"`                  // 操作原因
}

// PostListItem 帖子列表项响应体
type PostListItem struct {
	ID            int    `json:"id" example:"1"`                           // 帖子ID
	UserID        int    `json:"user_id" example:"1"`                      // 用户ID
	Username      string `json:"username" example:"testuser"`              // 用户名
	CategoryID    int    `json:"category_id" example:"1"`                  // 版块ID
	CategoryName  string `json:"category_name" example:"技术讨论"`             // 版块名称
	Title         string `json:"title" example:"技术分享帖"`                    // 帖子标题
	Content       string `json:"content" example:"## 技术分享\n这是内容"`          // 帖子内容（截取前100字符）
	ViewCount     int    `json:"view_count" example:"150"`                 // 浏览数
	LikeCount     int    `json:"like_count" example:"25"`                  // 点赞数
	DislikeCount  int    `json:"dislike_count" example:"2"`                // 点踩数
	FavoriteCount int    `json:"favorite_count" example:"10"`              // 收藏数
	IsEssence     bool   `json:"is_essence" example:"true"`                // 是否精华帖
	IsPinned      bool   `json:"is_pinned" example:"false"`                // 是否置顶
	Status        string `json:"status" example:"Normal"`                  // 帖子状态
	PublishIP     string `json:"publish_ip" example:"192.168.1.1"`         // 发布IP
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt     string `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
}

// PostListResponse 帖子列表响应体
type PostListResponse struct {
	List     []PostListItem `json:"list"`      // 帖子列表
	Total    int64          `json:"total"`     // 总数量
	Page     int            `json:"page"`      // 当前页码
	PageSize int            `json:"page_size"` // 每页数量
}

// PostDetailResponse 帖子详情响应体
type PostDetailResponse struct {
	ID             int    `json:"id" example:"1"`                           // 帖子ID
	UserID         int    `json:"user_id" example:"1"`                      // 用户ID
	Username       string `json:"username" example:"testuser"`              // 用户名
	CategoryID     int    `json:"category_id" example:"1"`                  // 版块ID
	CategoryName   string `json:"category_name" example:"技术讨论"`             // 版块名称
	Title          string `json:"title" example:"技术分享帖"`                    // 帖子标题
	Content        string `json:"content" example:"## 技术分享\n这是内容"`          // 帖子内容
	ReadPermission string `json:"read_permission" example:"login"`          // 阅读限制
	ViewCount      int    `json:"view_count" example:"150"`                 // 浏览数
	LikeCount      int    `json:"like_count" example:"25"`                  // 点赞数
	DislikeCount   int    `json:"dislike_count" example:"2"`                // 点踩数
	FavoriteCount  int    `json:"favorite_count" example:"10"`              // 收藏数
	IsEssence      bool   `json:"is_essence" example:"true"`                // 是否精华帖
	IsPinned       bool   `json:"is_pinned" example:"false"`                // 是否置顶
	Status         string `json:"status" example:"Normal"`                  // 帖子状态
	PublishIP      string `json:"publish_ip" example:"192.168.1.1"`         // 发布IP
	CreatedAt      string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt      string `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
}
