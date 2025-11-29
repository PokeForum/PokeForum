package schema

// CategoryListRequest 版块列表查询请求体
type CategoryListRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`               // 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // 每页数量
	Keyword  string `form:"keyword" example:"技术"`                                    // 搜索关键词（版块名称或描述）
	Status   string `form:"status" example:"Normal"`                                 // 版块状态筛选
}

// CategoryCreateRequest 创建版块请求体
type CategoryCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50" example:"技术讨论"`                                 // 版块名称
	Slug        string `json:"slug" binding:"required,min=2,max=50" example:"tech"`                                 // 版块英文标识
	Description string `json:"description" binding:"max=500" example:"技术相关话题讨论区"`                                   // 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"`                                         // 版块图标
	Weight      int    `json:"weight" binding:"min=0" example:"0"`                                                  // 权重排序
	Status      string `json:"status" binding:"required,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // 版块状态
}

// CategoryUpdateRequest 更新版块请求体
type CategoryUpdateRequest struct {
	ID          int    `json:"id" binding:"required" example:"1"`                                                    // 版块ID
	Name        string `json:"name" binding:"omitempty,min=2,max=50" example:"技术讨论"`                                 // 版块名称
	Slug        string `json:"slug" binding:"omitempty,min=2,max=50" example:"tech"`                                 // 版块英文标识
	Description string `json:"description" binding:"omitempty,max=500" example:"技术相关话题讨论区"`                          // 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"`                                          // 版块图标
	Weight      int    `json:"weight" binding:"omitempty,min=0" example:"0"`                                         // 权重排序
	Status      string `json:"status" binding:"omitempty,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // 版块状态
}

// CategoryStatusUpdateRequest 更新版块状态请求体
type CategoryStatusUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                                   // 版块ID
	Status string `json:"status" binding:"required,oneof=Normal LoginRequired Hidden Locked" example:"Normal"` // 版块状态
	Reason string `json:"reason" example:"版块调整"`                                                               // 操作原因
}

// CategoryListItem 版块列表项响应体
type CategoryListItem struct {
	ID          int    `json:"id" example:"1"`                              // 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // 版块名称
	Slug        string `json:"slug" example:"tech"`                         // 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // 版块图标
	Weight      int    `json:"weight" example:"0"`                          // 权重排序
	Status      string `json:"status" example:"Normal"`                     // 版块状态
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"`    // 创建时间
	UpdatedAt   string `json:"updated_at" example:"2024-01-01 00:00:00"`    // 更新时间
}

// CategoryListResponse 版块列表响应体
type CategoryListResponse struct {
	List     []CategoryListItem `json:"list"`      // 版块列表
	Total    int64              `json:"total"`     // 总数量
	Page     int                `json:"page"`      // 当前页码
	PageSize int                `json:"page_size"` // 每页数量
}

// CategoryDetailResponse 版块详情响应体
type CategoryDetailResponse struct {
	ID          int    `json:"id" example:"1"`                              // 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // 版块名称
	Slug        string `json:"slug" example:"tech"`                         // 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // 版块图标
	Weight      int    `json:"weight" example:"0"`                          // 权重排序
	Status      string `json:"status" example:"Normal"`                     // 版块状态
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"`    // 创建时间
	UpdatedAt   string `json:"updated_at" example:"2024-01-01 00:00:00"`    // 更新时间
}

// CategoryModeratorRequest 设置版块版主请求体
type CategoryModeratorRequest struct {
	CategoryID int    `json:"category_id" binding:"required" example:"1"` // 版块ID
	UserID     int    `json:"user_id" binding:"required" example:"10"`    // 版主用户ID
	Reason     string `json:"reason" example:"版主任命"`                      // 操作原因
}
