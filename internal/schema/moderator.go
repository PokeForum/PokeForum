package schema

// PostBanRequest Ban post request | 封禁帖子请求体
type PostBanRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	Reason string `json:"reason" example:"违规内容"`             // Ban reason | 封禁原因
}

// PostEditRequest Edit post request | 编辑帖子请求体
type PostEditRequest struct {
	ID      int    `json:"id" binding:"required" example:"1"`                           // Post ID | 帖子ID
	Title   string `json:"title" binding:"required,min=1,max=100" example:"修改后的标题"`     // Post title | 帖子标题
	Content string `json:"content" binding:"required,min=1,max=10000" example:"修改后的内容"` // Post content | 帖子内容
}

// PostMoveRequest Move post request | 移动帖子请求体
// type PostMoveRequest struct {
//	ID               int `json:"id" binding:"required" example:"1"`                 // Post ID | 帖子ID
//	TargetCategoryID int `json:"target_category_id" binding:"required" example:"2"` // Target category ID | 目标版块ID
//}

// PostEssenceRequest Set post essence request | 设置帖子精华请求体
type PostEssenceRequest struct {
	ID        int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	IsEssence bool   `json:"is_essence" example:"true"`         // Whether essence | 是否精华
	Reason    string `json:"reason" example:"优质内容"`             // Operation reason | 操作原因
}

// PostLockRequest Lock post request | 锁定帖子请求体
type PostLockRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	IsLock bool   `json:"is_lock" example:"true"`            // Whether locked | 是否锁定
	Reason string `json:"reason" example:"违规讨论"`             // Operation reason | 操作原因
}

// PostPinRequest Pin post request | 置顶帖子请求体
type PostPinRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"` // Post ID | 帖子ID
	IsPin  bool   `json:"is_pin" example:"true"`             // Whether pinned | 是否置顶
	Reason string `json:"reason" example:"重要公告"`             // Operation reason | 操作原因
}

// CategoryEditRequest Edit category request | 编辑版块请求体
type CategoryEditRequest struct {
	ID          int    `json:"id" binding:"required" example:"1"`                   // Category ID | 版块ID
	Name        string `json:"name" binding:"required,min=1,max=50" example:"技术讨论"` // Category name | 版块名称
	Description string `json:"description" example:"技术相关话题讨论区"`                     // Category description | 版块描述
	Icon        string `json:"icon" example:"tech"`                                 // Category icon | 版块图标
}

// CategoryAnnouncementRequest Category announcement request | 版块公告请求体
type CategoryAnnouncementRequest struct {
	CategoryID int    `json:"category_id" binding:"required" example:"1"`                 // Category ID | 版块ID
	Title      string `json:"title" binding:"required,min=1,max=100" example:"版块公告标题"`    // Announcement title | 公告标题
	Content    string `json:"content" binding:"required,min=1,max=1000" example:"版块公告内容"` // Announcement content | 公告内容
	IsPinned   bool   `json:"is_pinned" example:"true"`                                   // Whether pinned announcement | 是否置顶公告
}

// CategoryAnnouncementResponse Category announcement response | 版块公告响应体
type CategoryAnnouncementResponse struct {
	ID         int    `json:"id" example:"1"`                           // Announcement ID | 公告ID
	CategoryID int    `json:"category_id" example:"1"`                  // Category ID | 版块ID
	Title      string `json:"title" example:"版块公告标题"`                   // Announcement title | 公告标题
	Content    string `json:"content" example:"版块公告内容"`                 // Announcement content | 公告内容
	IsPinned   bool   `json:"is_pinned" example:"true"`                 // Whether pinned | 是否置顶
	Username   string `json:"username" example:"moderator"`             // Publisher username | 发布者用户名
	CreatedAt  string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt  string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}

// ModeratorCategoriesResponse Moderator managed categories list response | 版主管理的版块列表响应体
type ModeratorCategoriesResponse struct {
	Categories []ModeratorCategory `json:"categories"` // Moderator managed categories list | 版主管理的版块列表
}

// ModeratorCategory Moderator managed category info | 版主管理的版块信息
type ModeratorCategory struct {
	ID          int    `json:"id" example:"1"`                           // Category ID | 版块ID
	Name        string `json:"name" example:"技术讨论"`                      // Category name | 版块名称
	Slug        string `json:"slug" example:"tech"`                      // Category slug | 版块标识
	Description string `json:"description" example:"技术相关话题讨论区"`          // Category description | 版块描述
	Icon        string `json:"icon" example:"tech"`                      // Category icon | 版块图标
	Status      string `json:"status" example:"Normal"`                  // Category status | 版块状态
	PostCount   int    `json:"post_count" example:"100"`                 // Post count | 帖子数量
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
}

// ModeratorPostResponse Moderator post operation response | 版主帖子操作响应体
type ModeratorPostResponse struct {
	ID           int    `json:"id" example:"1"`                           // Post ID | 帖子ID
	Title        string `json:"title" example:"技术分享帖"`                    // Post title | 帖子标题
	Content      string `json:"content" example:"很有见地的内容"`                // Post content | 帖子内容
	Username     string `json:"username" example:"testuser"`              // Author username | 作者用户名
	CategoryID   int    `json:"category_id" example:"1"`                  // Category ID | 版块ID
	CategoryName string `json:"category_name" example:"技术讨论"`             // Category name | 版块名称
	Status       string `json:"status" example:"Normal"`                  // Post status | 帖子状态
	IsEssence    bool   `json:"is_essence" example:"true"`                // Whether essence | 是否精华
	IsPinned     bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	ViewCount    int    `json:"view_count" example:"1500"`                // View count | 浏览数
	LikeCount    int    `json:"like_count" example:"50"`                  // Like count | 点赞数
	CommentCount int    `json:"comment_count" example:"25"`               // Comment count | 评论数
	CreatedAt    string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt    string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}
