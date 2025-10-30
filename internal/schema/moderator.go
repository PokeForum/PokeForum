package schema

// PostBanRequest 封禁帖子请求体
type PostBanRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // 帖子ID
	Reason string `json:"reason" example:"违规内容"`             // 封禁原因
}

// PostEditRequest 编辑帖子请求体
type PostEditRequest struct {
	ID      int    `json:"id" binding:"required" example:"1"`                           // 帖子ID
	Title   string `json:"title" binding:"required,min=1,max=100" example:"修改后的标题"`     // 帖子标题
	Content string `json:"content" binding:"required,min=1,max=10000" example:"修改后的内容"` // 帖子内容
}

// PostMoveRequest 移动帖子请求体
type PostMoveRequest struct {
	ID               int `json:"id" binding:"required" example:"1"`                 // 帖子ID
	TargetCategoryID int `json:"target_category_id" binding:"required" example:"2"` // 目标版块ID
}

// PostEssenceRequest 设置帖子精华请求体
type PostEssenceRequest struct {
	ID        int    `json:"id" binding:"required" example:"1"` // 帖子ID
	IsEssence bool   `json:"is_essence" example:"true"`         // 是否精华
	Reason    string `json:"reason" example:"优质内容"`             // 操作原因
}

// PostLockRequest 锁定帖子请求体
type PostLockRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // 帖子ID
	IsLock bool   `json:"is_lock" example:"true"`            // 是否锁定
	Reason string `json:"reason" example:"违规讨论"`             // 操作原因
}

// PostPinRequest 置顶帖子请求体
type PostPinRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // 帖子ID
	IsPin  bool   `json:"is_pin" example:"true"`             // 是否置顶
	Reason string `json:"reason" example:"重要公告"`             // 操作原因
}

// CategoryEditRequest 编辑版块请求体
type CategoryEditRequest struct {
	ID          int    `json:"id" binding:"required" example:"1"`                   // 版块ID
	Name        string `json:"name" binding:"required,min=1,max=50" example:"技术讨论"` // 版块名称
	Description string `json:"description" example:"技术相关话题讨论区"`                     // 版块描述
	Icon        string `json:"icon" example="tech"`                                 // 版块图标
}

// CategoryAnnouncementRequest 版块公告请求体
type CategoryAnnouncementRequest struct {
	CategoryID int    `json:"category_id" binding:"required" example:"1"`                 // 版块ID
	Title      string `json:"title" binding:"required,min=1,max=100" example:"版块公告标题"`    // 公告标题
	Content    string `json:"content" binding:"required,min=1,max=1000" example:"版块公告内容"` // 公告内容
	IsPinned   bool   `json:"is_pinned" example:"true"`                                   // 是否置顶公告
}

// CategoryAnnouncementResponse 版块公告响应体
type CategoryAnnouncementResponse struct {
	ID         int    `json:"id" example:"1"`                            // 公告ID
	CategoryID int    `json:"category_id" example:"1"`                   // 版块ID
	Title      string `json:"title" example:"版块公告标题"`                    // 公告标题
	Content    string `json:"content" example:"版块公告内容"`                  // 公告内容
	IsPinned   bool   `json:"is_pinned" example:"true"`                  // 是否置顶
	Username   string `json:"username" example:"moderator"`              // 发布者用户名
	CreatedAt  string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
	UpdatedAt  string `json:"updated_at" example:"2024-01-01T00:00:00Z"` // 更新时间
}

// ModeratorCategoriesResponse 版主管理的版块列表响应体
type ModeratorCategoriesResponse struct {
	Categories []ModeratorCategory `json:"categories"` // 版主管理的版块列表
}

// ModeratorCategory 版主管理的版块信息
type ModeratorCategory struct {
	ID          int    `json:"id" example:"1"`                            // 版块ID
	Name        string `json:"name" example:"技术讨论"`                       // 版块名称
	Slug        string `json:"slug" example="tech"`                       // 版块标识
	Description string `json:"description" example:"技术相关话题讨论区"`           // 版块描述
	Icon        string `json:"icon" example:"tech"`                       // 版块图标
	Status      string `json:"status" example:"Normal"`                   // 版块状态
	PostCount   int    `json:"post_count" example:"100"`                  // 帖子数量
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
}

// ModeratorPostResponse 版主帖子操作响应体
type ModeratorPostResponse struct {
	ID           int    `json:"id" example:"1"`                            // 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                     // 帖子标题
	Content      string `json:"content" example:"很有见地的内容"`                 // 帖子内容
	Username     string `json:"username" example:"testuser"`               // 作者用户名
	CategoryID   int    `json:"category_id" example:"1"`                   // 版块ID
	CategoryName string `json:"category_name" example:"技术讨论"`              // 版块名称
	Status       string `json:"status" example:"Normal"`                   // 帖子状态
	IsEssence    bool   `json:"is_essence" example:"true"`                 // 是否精华
	IsPinned     bool   `json:"is_pinned" example:"false"`                 // 是否置顶
	ViewCount    int    `json:"view_count" example:"1500"`                 // 浏览数
	LikeCount    int    `json:"like_count" example:"50"`                   // 点赞数
	CommentCount int    `json:"comment_count" example:"25"`                // 评论数
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
	UpdatedAt    string `json:"updated_at" example:"2024-01-01T00:00:00Z"` // 更新时间
}
