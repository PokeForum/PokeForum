package schema

// UserBlacklistAddRequest 用户添加黑名单请求体
type UserBlacklistAddRequest struct {
	BlockedUserID int `json:"blocked_user_id" binding:"required" example:"2"` // 被拉黑用户ID
}

// UserBlacklistAddResponse 用户添加黑名单响应体
type UserBlacklistAddResponse struct {
	ID            int    `json:"id" example:"1"`                           // 黑名单记录ID
	UserID        int    `json:"user_id" example:"1"`                      // 执行拉黑的用户ID
	BlockedUserID int    `json:"blocked_user_id" example:"2"`              // 被拉黑用户ID
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
}

// UserBlacklistRemoveRequest 用户移除黑名单请求体
type UserBlacklistRemoveRequest struct {
	BlockedUserID int `json:"blocked_user_id" binding:"required" example:"2"` // 被拉黑用户ID
}

// UserBlacklistRemoveResponse 用户移除黑名单响应体
type UserBlacklistRemoveResponse struct {
	Message string `json:"message" example:"移除黑名单成功"` // 操作结果消息
}

// UserBlacklistItem 黑名单项目
type UserBlacklistItem struct {
	ID              int    `json:"id" example:"1"`                                          // 黑名单记录ID
	BlockedUserID   int    `json:"blocked_user_id" example:"2"`                             // 被拉黑用户ID
	BlockedUsername string `json:"blocked_username" example:"targetuser"`                   // 被拉黑用户名
	BlockedAvatar   string `json:"blocked_avatar" example:"https://example.com/avatar.jpg"` // 被拉黑用户头像
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"`                // 创建时间
}

// UserBlacklistListRequest 用户黑名单列表请求体
type UserBlacklistListRequest struct {
	Page     int `json:"page" binding:"min=1" example:"1"`              // 页码
	PageSize int `json:"page_size" binding:"min=1,max=50" example:"20"` // 每页数量
}

// UserBlacklistListResponse 用户黑名单列表响应体
type UserBlacklistListResponse struct {
	List       []UserBlacklistItem `json:"list"`        // 黑名单列表
	Total      int64               `json:"total"`       // 总数量
	Page       int                 `json:"page"`        // 当前页码
	PageSize   int                 `json:"page_size"`   // 每页数量
	TotalPages int                 `json:"total_pages"` // 总页数
}
