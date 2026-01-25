package schema

// UserBlacklistAddRequest User add to blacklist request | 用户添加黑名单请求体
type UserBlacklistAddRequest struct {
	BlockedUserID int `json:"blocked_user_id" binding:"required" example:"2"` // Blocked user ID | 被拉黑用户ID
}

// UserBlacklistAddResponse User add to blacklist response | 用户添加黑名单响应体
type UserBlacklistAddResponse struct {
	ID            int    `json:"id" example:"1"`                           // Blacklist record ID | 黑名单记录ID
	UserID        int    `json:"user_id" example:"1"`                      // User ID who performed the block | 执行拉黑的用户ID
	BlockedUserID int    `json:"blocked_user_id" example:"2"`              // Blocked user ID | 被拉黑用户ID
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
}

// UserBlacklistRemoveRequest User remove from blacklist request | 用户移除黑名单请求体
type UserBlacklistRemoveRequest struct {
	BlockedUserID int `json:"blocked_user_id" binding:"required" example:"2"` // Blocked user ID | 被拉黑用户ID
}

// UserBlacklistRemoveResponse User remove from blacklist response | 用户移除黑名单响应体
type UserBlacklistRemoveResponse struct {
	Message string `json:"message" example:"移除黑名单成功"` // Operation result message | 操作结果消息
}

// UserBlacklistItem Blacklist item | 黑名单项目
type UserBlacklistItem struct {
	ID              int    `json:"id" example:"1"`                                          // Blacklist record ID | 黑名单记录ID
	BlockedUserID   int    `json:"blocked_user_id" example:"2"`                             // Blocked user ID | 被拉黑用户ID
	BlockedUsername string `json:"blocked_username" example:"targetuser"`                   // Blocked username | 被拉黑用户名
	BlockedAvatar   string `json:"blocked_avatar" example:"https://example.com/avatar.jpg"` // Blocked user avatar | 被拉黑用户头像
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"`                // Creation time | 创建时间
}

// UserBlacklistListRequest User blacklist list request | 用户黑名单列表请求体
type UserBlacklistListRequest struct {
	Page     int `json:"page" binding:"min=1" example:"1"`              // Page number | 页码
	PageSize int `json:"page_size" binding:"min=1,max=50" example:"20"` // Items per page | 每页数量
}

// UserBlacklistListResponse User blacklist list response | 用户黑名单列表响应体
type UserBlacklistListResponse struct {
	List       []UserBlacklistItem `json:"list"`        // Blacklist list | 黑名单列表
	Total      int64               `json:"total"`       // Total count | 总数量
	Page       int                 `json:"page"`        // Current page number | 当前页码
	PageSize   int                 `json:"page_size"`   // Items per page | 每页数量
	TotalPages int                 `json:"total_pages"` // Total pages | 总页数
}
