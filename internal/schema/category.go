package schema

// UserCategoryListItem User category list item response | 用户版块列表项响应体
type UserCategoryListItem struct {
	ID          int    `json:"id" example:"1"`                              // Category ID | 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // Category name | 版块名称
	Slug        string `json:"slug" example:"tech"`                         // Category slug | 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // Category description | 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // Category icon | 版块图标
	Weight      int    `json:"weight" example:"0"`                          // Sort weight | 权重排序
	CreatedAt   string `json:"created_at" example:"2024-01-01 00:00:00"`    // Creation time | 创建时间
}

// UserCategoryResponse User category response | 用户版块响应体
type UserCategoryResponse struct {
	List []UserCategoryListItem `json:"list"` // Category list | 版块列表
}
