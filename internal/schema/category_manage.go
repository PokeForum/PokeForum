package schema

// CategoryListRequest Category list query request | 版块列表查询请求体
type CategoryListRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`               // Page number | 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // Items per page | 每页数量
	Keyword  string `form:"keyword" example:"技术"`                                    // Search keyword (category name or description) | 搜索关键词（版块名称或描述）
	Status   string `form:"status" example:"Normal"`                                 // Category status filter | 版块状态筛选
}

// CategoryCreateRequest Create category request | 创建版块请求体
type CategoryCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50" example:"技术讨论"`                                 // Category name | 版块名称
	Slug        string `json:"slug" binding:"required,min=2,max=50" example:"tech"`                                 // Category slug | 版块英文标识
	Description string `json:"description" binding:"max=500" example:"技术相关话题讨论区"`                                   // Category description | 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"`                                         // Category icon | 版块图标
	Weight      int    `json:"weight" binding:"min=0" example:"0"`                                                  // Sort weight | 权重排序
	Status      string `json:"status" binding:"required,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // Category status | 版块状态
}

// CategoryUpdateRequest Update category request | 更新版块请求体
type CategoryUpdateRequest struct {
	ID          int    `json:"id" binding:"required" example:"1"`                                                    // Category ID | 版块ID
	Name        string `json:"name" binding:"omitempty,min=2,max=50" example:"技术讨论"`                                 // Category name | 版块名称
	Slug        string `json:"slug" binding:"omitempty,min=2,max=50" example:"tech"`                                 // Category slug | 版块英文标识
	Description string `json:"description" binding:"omitempty,max=500" example:"技术相关话题讨论区"`                          // Category description | 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"`                                          // Category icon | 版块图标
	Weight      int    `json:"weight" binding:"omitempty,min=0" example:"0"`                                         // Sort weight | 权重排序
	Status      string `json:"status" binding:"omitempty,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // Category status | 版块状态
}

// CategoryStatusUpdateRequest Update category status request | 更新版块状态请求体
type CategoryStatusUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                                   // Category ID | 版块ID
	Status string `json:"status" binding:"required,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // Category status | 版块状态
	Reason string `json:"reason" example:"版块调整"`                                                               // Operation reason | 操作原因
}

// CategoryListItem Category list item response | 版块列表项响应体
type CategoryListItem struct {
	ID          int    `json:"id" example:"1"`                              // Category ID | 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // Category name | 版块名称
	Slug        string `json:"slug" example:"tech"`                         // Category slug | 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // Category description | 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // Category icon | 版块图标
	Weight      int    `json:"weight" example:"0"`                          // Sort weight | 权重排序
	Status      string `json:"status" example:"Normal"`                     // Category status | 版块状态
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"`    // Creation time | 创建时间
	UpdatedAt   string `json:"updated_at" example:"2024-01-01 00:00:00"`    // Update time | 更新时间
}

// CategoryListResponse Category list response | 版块列表响应体
type CategoryListResponse struct {
	List     []CategoryListItem `json:"list"`      // Category list | 版块列表
	Total    int64              `json:"total"`     // Total count | 总数量
	Page     int                `json:"page"`      // Current page number | 当前页码
	PageSize int                `json:"page_size"` // Items per page | 每页数量
}

// CategoryDetailResponse Category detail response | 版块详情响应体
type CategoryDetailResponse struct {
	ID          int    `json:"id" example:"1"`                              // Category ID | 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // Category name | 版块名称
	Slug        string `json:"slug" example:"tech"`                         // Category slug | 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // Category description | 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // Category icon | 版块图标
	Weight      int    `json:"weight" example:"0"`                          // Sort weight | 权重排序
	Status      string `json:"status" example:"Normal"`                     // Category status | 版块状态
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"`    // Creation time | 创建时间
	UpdatedAt   string `json:"updated_at" example:"2024-01-01 00:00:00"`    // Update time | 更新时间
}

// CategoryModeratorRequest Set category moderator request | 设置版块版主请求体
type CategoryModeratorRequest struct {
	CategoryID int    `json:"category_id" binding:"required" example:"1"` // Category ID | 版块ID
	UserID     int    `json:"user_id" binding:"required" example:"10"`    // Moderator user ID | 版主用户ID
	Reason     string `json:"reason" example:"版主任命"`                      // Operation reason | 操作原因
}
