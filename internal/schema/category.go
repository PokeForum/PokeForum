package schema

// UserCategoryListItem 用户版块列表项响应体
type UserCategoryListItem struct {
	ID          int    `json:"id" example:"1"`                              // 版块ID
	Name        string `json:"name" example:"技术讨论"`                         // 版块名称
	Slug        string `json:"slug" example:"tech"`                         // 版块英文标识
	Description string `json:"description" example:"技术相关话题讨论区"`             // 版块描述
	Icon        string `json:"icon" example:"https://example.com/icon.png"` // 版块图标
	Weight      int    `json:"weight" example:"0"`                          // 权重排序
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`   // 创建时间
}

// UserCategoryResponse 用户版块响应体
type UserCategoryResponse struct {
	List []UserCategoryListItem `json:"list"` // 版块列表
}
