package schema

// InvitationCodeListRequest Invitation code list query request | 邀请码列表查询请求体
type InvitationCodeListRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`               // Page number | 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // Items per page | 每页数量
	Keyword  string `form:"keyword" example:"abc123"`                                // Search keyword (invitation code) | 搜索关键词（邀请码）
	Status   string `form:"status" example:"unused"`                                 // Status filter: unused, used, expired, disabled | 状态筛选
	Mode     string `form:"mode" example:"direct"`                                   // Generation mode filter | 生成方式筛选
}

// InvitationCodeCreateRequest Create invitation code request (admin) | 创建邀请码请求体（管理员）
type InvitationCodeCreateRequest struct {
	Code       string `json:"code" binding:"required,min=6,max=32" example:"abc123def456"`           // Invitation code | 邀请码
	CreatorID  int    `json:"creator_id" binding:"required,min=1" example:"1"`                       // Creator user ID | 创建者用户ID
	Mode       string `json:"mode" binding:"required,oneof=direct points currency" example:"direct"` // Generation mode | 生成方式
	CostAmount int    `json:"cost_amount" binding:"min=0" example:"0"`                               // Cost amount | 消耗数量
	Remark     string `json:"remark" binding:"max=500" example="管理员手动创建"`                            // Remark | 备注
}

// InvitationCodeUpdateRequest Update invitation code request | 更新邀请码请求体
type InvitationCodeUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`       // Invitation code ID | 邀请码ID
	Remark string `json:"remark" binding:"max=500" example="更新备注"` // Remark | 备注
}

// InvitationCodeStatusUpdateRequest Update invitation code status request | 更新邀请码状态请求体
type InvitationCodeStatusUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                               // Invitation code ID | 邀请码ID
	Status string `json:"status" binding:"required,oneof=unused used expired disabled" example:"disabled"` // Status | 状态
	Reason string `json:"reason" example:"违规操作"`                                                           // Operation reason | 操作原因
}

// InvitationCodeListItem Invitation code list item response | 邀请码列表项响应体
type InvitationCodeListItem struct {
	ID             int    `json:"id" example:"1"`                                  // Invitation code ID | 邀请码ID
	Code           string `json:"code" example:"abc123def456"`                     // Invitation code | 邀请码
	CreatorID      int    `json:"creator_id" example:"1"`                          // Creator user ID | 创建者用户ID
	UsedByID       *int   `json:"used_by_id,omitempty" example:"2"`                // Used by user ID | 使用者用户ID
	Status         string `json:"status" example:"unused"`                         // Status | 状态
	GenerationMode string `json:"generation_mode" example:"direct"`                // Generation mode | 生成方式
	CostAmount     int    `json:"cost_amount" example:"0"`                         // Cost amount | 消耗数量
	UsedAt         string `json:"used_at,omitempty" example:"2024-01-01 12:00:00"` // Used at timestamp | 使用时间
	UsedIP         string `json:"used_ip,omitempty" example:"192.168.1.1"`         // Used IP address | 使用时的IP地址
	UsedUserAgent  string `json:"used_user_agent,omitempty" example:"Mozilla/5.0"` // Used user agent | 使用时的用户代理
	Remark         string `json:"remark" example:"管理员手动创建"`                        // Remark | 备注
	CreatedAt      string `json:"created_at" example:"2024-01-01 00:00:00"`        // Creation time | 创建时间
	UpdatedAt      string `json:"updated_at" example:"2024-01-01 00:00:00"`        // Update time | 更新时间
}

// InvitationCodeListResponse Invitation code list response | 邀请码列表响应体
type InvitationCodeListResponse struct {
	List     []InvitationCodeListItem `json:"list"`      // Invitation code list | 邀请码列表
	Total    int64                    `json:"total"`     // Total count | 总数量
	Page     int                      `json:"page"`      // Current page number | 当前页码
	PageSize int                      `json:"page_size"` // Items per page | 每页数量
}

// InvitationCodeDetailResponse Invitation code detail response | 邀请码详情响应体
type InvitationCodeDetailResponse struct {
	ID             int    `json:"id" example:"1"`                                  // Invitation code ID | 邀请码ID
	Code           string `json:"code" example:"abc123def456"`                     // Invitation code | 邀请码
	CreatorID      int    `json:"creator_id" example:"1"`                          // Creator user ID | 创建者用户ID
	UsedByID       *int   `json:"used_by_id,omitempty" example:"2"`                // Used by user ID | 使用者用户ID
	Status         string `json:"status" example:"unused"`                         // Status | 状态
	GenerationMode string `json:"generation_mode" example:"direct"`                // Generation mode | 生成方式
	CostAmount     int    `json:"cost_amount" example:"0"`                         // Cost amount | 消耗数量
	UsedAt         string `json:"used_at,omitempty" example:"2024-01-01 12:00:00"` // Used at timestamp | 使用时间
	UsedIP         string `json:"used_ip,omitempty" example:"192.168.1.1"`         // Used IP address | 使用时的IP地址
	UsedUserAgent  string `json:"used_user_agent,omitempty" example:"Mozilla/5.0"` // Used user agent | 使用时的用户代理
	Remark         string `json:"remark" example:"管理员手动创建"`                        // Remark | 备注
	CreatedAt      string `json:"created_at" example:"2024-01-01 00:00:00"`        // Creation time | 创建时间
	UpdatedAt      string `json:"updated_at" example:"2024-01-01 00:00:00"`        // Update time | 更新时间
}

// InvitationCodeStatsResponse Invitation code statistics response | 邀请码统计响应体
type InvitationCodeStatsResponse struct {
	TotalCount     int64 `json:"total_count" example:"100"`     // Total codes | 总邀请码数
	UnusedCount    int64 `json:"unused_count" example:"50"`     // Unused codes | 未使用数量
	UsedCount      int64 `json:"used_count" example:"40"`       // Used codes | 已使用数量
	ExpiredCount   int64 `json:"expired_count" example:"8"`     // Expired codes | 已过期数量
	DisabledCount  int64 `json:"disabled_count" example:"2"`    // Disabled codes | 已禁用数量
	TotalUsedCount int64 `json:"total_used_count" example:"40"` // Total used count | 总使用次数
}
