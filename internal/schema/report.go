package schema

// ========== User-side DTOs | 用户侧DTO ==========

// ReportCreateRequest Create report request | 创建举报请求
type ReportCreateRequest struct {
	// Target ID | 被举报目标ID
	TargetID int `json:"target_id" binding:"required,min=1" example:"1"`
	// Target type: post/comment/user | 目标类型
	TargetType string `json:"target_type" binding:"required,oneof=post comment user" example:"post"`
	// Report type: spam/harassment/inappropriate/illegal/plagiarism/other | 举报类型
	ReportType string `json:"report_type" binding:"required,oneof=spam harassment inappropriate illegal plagiarism other" example:"spam"`
	// Detailed reason (supports Markdown) | 详细原因（支持 Markdown）
	Reason string `json:"reason" binding:"max=5000" example:"This post contains spam content. See details: [link](url)"`
}

// ReportCreateResponse Create report response | 创建举报响应
type ReportCreateResponse struct {
	// Report ID | 举报ID
	ID int `json:"id" example:"1"`
	// Report status | 举报状态
	Status string `json:"status" example:"pending"`
	// Created time | 创建时间
	CreatedAt string `json:"created_at" example:"2024-01-01 12:00:00"`
}

// UserReportItem User report list item | 用户举报列表项
type UserReportItem struct {
	// Report ID | 举报ID
	ID int `json:"id" example:"1"`
	// Target ID | 目标ID
	TargetID int `json:"target_id" example:"100"`
	// Target type | 目标类型
	TargetType string `json:"target_type" example:"post"`
	// Target title (post title or comment preview) | 目标标题（帖子标题或评论摘要）
	TargetTitle string `json:"target_title,omitempty" example:"Post title"`
	// Report type | 举报类型
	ReportType string `json:"report_type" example:"spam"`
	// Reason | 原因
	Reason string `json:"reason" example:"Spam content"`
	// Status | 状态
	Status string `json:"status" example:"resolved"`
	// Handle result | 处理结果
	HandleResult string `json:"handle_result,omitempty" example:"Content deleted"`
	// Created time | 创建时间
	CreatedAt string `json:"created_at" example:"2024-01-01 12:00:00"`
	// Handled time | 处理时间
	HandledAt string `json:"handled_at,omitempty" example:"2024-01-02 10:00:00"`
}

// UserReportListResponse User report list response | 用户举报列表响应
type UserReportListResponse struct {
	// Report list | 举报列表
	List []UserReportItem `json:"list"`
	// Total count | 总数
	Total int64 `json:"total"`
	// Page number | 页码
	Page int `json:"page"`
	// Page size | 每页数量
	PageSize int `json:"page_size"`
	// Total pages | 总页数
	TotalPages int `json:"total_pages"`
}

// ========== Admin-side DTOs | 管理侧DTO ==========

// ReportListQuery Report list query parameters | 举报列表查询参数
type ReportListQuery struct {
	// Page number | 页码
	Page int `form:"page" example:"1"`
	// Page size | 每页数量
	PageSize int `form:"page_size" example:"20"`
	// Status: pending/processing/resolved/ignored | 状态
	Status string `form:"status" example:"pending"`
	// Target type: post/comment/user | 目标类型
	TargetType string `form:"target_type" example:"post"`
	// Report type | 举报类型
	ReportType string `form:"report_type" example:"spam"`
	// Reporter ID (0 for all) | 举报者ID（0表示全部）
	ReporterID int `form:"reporter_id" example:"0"`
	// Target author ID (0 for all) | 被举报作者ID（0表示全部）
	TargetAuthorID int `form:"target_author_id" example:"0"`
}

// ReportListItem Report list item (admin) | 举报列表项（管理侧）
type ReportListItem struct {
	// Report ID | 举报ID
	ID int `json:"id" example:"1"`
	// Reporter ID | 举报者ID
	ReporterID int `json:"reporter_id" example:"10"`
	// Reporter name | 举报者名称
	ReporterName string `json:"reporter_name" example:"user123"`
	// Target ID | 目标ID
	TargetID int `json:"target_id" example:"100"`
	// Target type | 目标类型
	TargetType string `json:"target_type" example:"post"`
	// Target title | 目标标题
	TargetTitle string `json:"target_title,omitempty" example:"Post title"`
	// Target content preview | 目标内容预览
	TargetContent string `json:"target_content,omitempty" example:"Content preview..."`
	// Target author ID | 目标作者ID
	TargetAuthorID int `json:"target_author_id" example:"20"`
	// Target author name | 目标作者名称
	TargetAuthorName string `json:"target_author_name" example:"author456"`
	// Report type | 举报类型
	ReportType string `json:"report_type" example:"spam"`
	// Reason (supports Markdown) | 原因（支持 Markdown）
	Reason string `json:"reason" example:"Spam content with [evidence](url)"`
	// Status | 状态
	Status string `json:"status" example:"pending"`
	// Handler ID | 处理人ID
	HandlerID int `json:"handler_id,omitempty" example:"5"`
	// Handler name | 处理人名称
	HandlerName string `json:"handler_name,omitempty" example:"admin"`
	// Handle result | 处理结果
	HandleResult string `json:"handle_result,omitempty"`
	// Handle action | 处理动作
	HandleAction string `json:"handle_action,omitempty"`
	// Created time | 创建时间
	CreatedAt string `json:"created_at" example:"2024-01-01 12:00:00"`
	// Handled time | 处理时间
	HandledAt string `json:"handled_at,omitempty"`
}

// ReportListResponse Report list response (admin) | 举报列表响应（管理侧）
type ReportListResponse struct {
	// Report list | 举报列表
	List []ReportListItem `json:"list"`
	// Total count | 总数
	Total int64 `json:"total"`
	// Page number | 页码
	Page int `json:"page"`
	// Page size | 每页数量
	PageSize int `json:"page_size"`
	// Total pages | 总页数
	TotalPages int `json:"total_pages"`
}

// ReportHandleRequest Handle report request | 处理举报请求
type ReportHandleRequest struct {
	// Report ID | 举报ID
	ID int `json:"id" binding:"required,min=1" example:"1"`
	// Action: delete/warn/ban/ignore | 动作
	Action string `json:"action" binding:"required,oneof=delete warn ban ignore" example:"delete"`
	// Handle result description | 处理结果说明
	HandleResult string `json:"handle_result" binding:"required,max=500" example:"Content violates community guidelines"`
	// Whether to notify the reporter | 是否通知举报者
	NotifyReporter bool `json:"notify_reporter" example:"true"`
}

// ReportHandleResponse Handle report response | 处理举报响应
type ReportHandleResponse struct {
	// Report ID | 举报ID
	ID int `json:"id" example:"1"`
	// Status | 状态
	Status string `json:"status" example:"resolved"`
	// Handled time | 处理时间
	HandledAt string `json:"handled_at" example:"2024-01-01 14:00:00"`
}

// ReportBatchHandleRequest Batch handle reports request | 批量处理举报请求
type ReportBatchHandleRequest struct {
	// Report IDs | 举报ID列表
	IDs []int `json:"ids" binding:"required,min=1" example:"[1,2,3]"`
	// Action: delete/warn/ban/ignore | 动作
	Action string `json:"action" binding:"required,oneof=delete warn ban ignore" example:"ignore"`
	// Handle result description | 处理结果说明
	HandleResult string `json:"handle_result" binding:"required,max=500"`
}

// ReportStatsItem Report statistics item | 举报统计项
type ReportStatsItem struct {
	// Type | 类型
	Type string `json:"type"`
	// Count | 数量
	Count int64 `json:"count"`
}

// ReportStatsResponse Report statistics response | 举报统计响应
type ReportStatsResponse struct {
	// Total reports | 总举报数
	TotalReports int64 `json:"total_reports"`
	// Pending count | 待处理数
	PendingCount int64 `json:"pending_count"`
	// Processing count | 处理中数
	ProcessingCount int64 `json:"processing_count"`
	// Resolved count | 已处理数
	ResolvedCount int64 `json:"resolved_count"`
	// Ignored count | 已忽略数
	IgnoredCount int64 `json:"ignored_count"`
	// Today count | 今日举报数
	TodayCount int64 `json:"today_count"`
	// Statistics by type | 按类型统计
	ByTypeStats []ReportStatsItem `json:"by_type_stats"`
}

// ReportAssignRequest Assign report to moderator request | 分配举报给版主请求
type ReportAssignRequest struct {
	// Report ID | 举报ID
	ReportID int `json:"report_id" binding:"required,min=1"`
}
