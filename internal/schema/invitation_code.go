package schema

import "time"

// GenerateInvitationCodeRequest Generate invitation code request | 生成邀请码请求
type GenerateInvitationCodeRequest struct {
	// User ID (obtained from JWT token, used for documentation only) | 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int `json:"user_id" example:"1001"`
}

// GenerateInvitationCodeResponse Generate invitation code response | 生成邀请码响应
type GenerateInvitationCodeResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"邀请码生成成功"`
	// Response data | 响应数据
	Data *UserInvitationCodeDetail `json:"data"`
}

// MyInvitationCodesRequest Get my invitation codes request | 获取我的邀请码列表请求
type MyInvitationCodesRequest struct {
	// User ID (obtained from JWT token, used for documentation only) | 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int `json:"user_id" example:"1001"`
	// Page number, default 1 | 页码，默认1
	Page int `form:"page" binding:"min=1" example:"1"`
	// Items per page, default 20 | 每页数量，默认20
	PageSize int `form:"page_size" binding:"min=1,max=100" example:"20"`
}

// MyInvitationCodesResponse Get my invitation codes response | 获取我的邀请码列表响应
type MyInvitationCodesResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"获取成功"`
	// Response data | 响应数据
	Data *UserInvitationCodeListData `json:"data"`
}

// UserInvitationCodeDetail User invitation code detail | 用户邀请码详情
type UserInvitationCodeDetail struct {
	// Invitation code | 邀请码
	Code string `json:"code" example:"abc123def456"`
	// Status: unused, used, disabled | 状态
	Status string `json:"status" example:"unused"`
	// Generation mode: direct, points, currency | 生成方式
	GenerationMode string `json:"generation_mode" example:"direct"`
	// Cost amount | 消耗数量
	CostAmount int `json:"cost_amount" example:"0"`
	// Used at timestamp | 使用时间
	UsedAt *time.Time `json:"used_at,omitempty" example:"2024-01-01 12:00:00"`
	// Created at timestamp | 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01 00:00:00"`
}

// UserInvitationCodeListData User invitation code list data | 用户邀请码列表数据
type UserInvitationCodeListData struct {
	// Invitation code list | 邀请码列表
	List []*UserInvitationCodeListItem `json:"list"`
	// Total count | 总数量
	Total int `json:"total" example:"10"`
	// Current page number | 当前页码
	Page int `json:"page" example:"1"`
	// Items per page | 每页数量
	PageSize int `json:"page_size" example:"20"`
}

// UserInvitationCodeListItem User invitation code list item | 用户邀请码列表项
type UserInvitationCodeListItem struct {
	// Invitation code ID | 邀请码ID
	ID int `json:"id" example:"1"`
	// Invitation code | 邀请码
	Code string `json:"code" example:"abc123def456"`
	// Status: unused, used, disabled | 状态
	Status string `json:"status" example:"unused"`
	// Generation mode: direct, points, currency | 生成方式
	GenerationMode string `json:"generation_mode" example:"direct"`
	// Cost amount | 消耗数量
	CostAmount int `json:"cost_amount" example:"0"`
	// Used at timestamp | 使用时间
	UsedAt *time.Time `json:"used_at,omitempty" example:"2024-01-01 12:00:00"`
	// Created at timestamp | 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01 00:00:00"`
}
