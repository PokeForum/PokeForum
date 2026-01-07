package schema

// CommentListRequest Comment list query request | 评论列表查询请求体
type CommentListRequest struct {
	Page       int    `form:"page" binding:"required,min=1" example:"1"`               // Page number | 页码
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // Items per page | 每页数量
	Keyword    string `form:"keyword" example:"技术"`                                    // Search keyword (comment content) | 搜索关键词（评论内容）
	PostID     int    `form:"post_id" example:"1"`                                     // Post ID filter | 帖子ID筛选
	UserID     int    `form:"user_id" example:"1"`                                     // User ID filter | 用户ID筛选
	ParentID   *int   `form:"parent_id" example:"1"`                                   // Parent comment ID filter (nil for top-level comments) | 父评论ID筛选（nil表示顶级评论）
	IsSelected *bool  `form:"is_selected" example:"true"`                              // Selected comment filter | 是否精选评论筛选
	IsPinned   *bool  `form:"is_pinned" example:"true"`                                // Pinned comment filter | 是否置顶评论筛选
	ReplyToID  int    `form:"reply_to_id" example:"1"`                                 // Reply target user ID filter | 回复目标用户ID筛选
}

// CommentCreateRequest Create comment request | 创建评论请求体
type CommentCreateRequest struct {
	PostID        int    `json:"post_id" binding:"required" example:"1"`                      // Post ID | 帖子ID
	UserID        int    `json:"user_id" binding:"required" example:"1"`                      // User ID | 用户ID
	ParentID      *int   `json:"parent_id" example:"1"`                                       // Parent comment ID | 父评论ID
	ReplyToUserID *int   `json:"reply_to_user_id" example:"2"`                                // Reply target user ID | 回复目标用户ID
	Content       string `json:"content" binding:"required,min=1,max=1000" example:"很有见地的评论"` // Comment content | 评论内容
	CommenterIP   string `json:"commenter_ip" example:"192.168.1.1"`                          // Commenter IP | 评论者IP
	DeviceInfo    string `json:"device_info" example:"Chrome/Windows"`                        // Device info | 设备信息
}

// CommentUpdateRequest Update comment request | 更新评论请求体
type CommentUpdateRequest struct {
	ID      int    `json:"id" binding:"required" example:"1"`                            // Comment ID | 评论ID
	Content string `json:"content" binding:"required,min=1,max=1000" example:"更新后的评论内容"` // Comment content | 评论内容
}

// CommentSelectedUpdateRequest Set comment selected request | 设置评论精选请求体
type CommentSelectedUpdateRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"` // Comment ID | 评论ID
	IsSelected bool   `json:"is_selected" example:"true"`        // Whether selected | 是否精选
	Reason     string `json:"reason" example:"优质评论"`             // Operation reason | 操作原因
}

// CommentPinUpdateRequest Set comment pinned request | 设置评论置顶请求体
type CommentPinUpdateRequest struct {
	ID       int    `json:"id" binding:"required" example:"1"` // Comment ID | 评论ID
	IsPinned bool   `json:"is_pinned" example:"true"`          // Whether pinned | 是否置顶
	Reason   string `json:"reason" example:"重要评论"`             // Operation reason | 操作原因
}

// CommentListItem Comment list item response | 评论列表项响应体
type CommentListItem struct {
	ID              int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	PostID          int    `json:"post_id" example:"1"`                      // Post ID | 帖子ID
	PostTitle       string `json:"post_title" example:"技术分享帖"`               // Post title | 帖子标题
	UserID          int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username        string `json:"username" example:"testuser"`              // Username | 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                    // Parent comment ID | 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`             // Reply target user ID | 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`   // Reply target username | 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                // Comment content | 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                  // Like count | 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                // Dislike count | 点踩数
	IsSelected      bool   `json:"is_selected" example:"true"`               // Whether selected | 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	CommenterIP     string `json:"commenter_ip" example:"192.168.1.1"`       // Commenter IP | 评论者IP
	DeviceInfo      string `json:"device_info" example:"Chrome/Windows"`     // Device info | 设备信息
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}

// CommentListResponse Comment list response | 评论列表响应体
type CommentListResponse struct {
	List     []CommentListItem `json:"list"`      // Comment list | 评论列表
	Total    int64             `json:"total"`     // Total count | 总数量
	Page     int               `json:"page"`      // Current page number | 当前页码
	PageSize int               `json:"page_size"` // Items per page | 每页数量
}

// CommentDetailResponse Comment detail response | 评论详情响应体
type CommentDetailResponse struct {
	ID              int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	PostID          int    `json:"post_id" example:"1"`                      // Post ID | 帖子ID
	PostTitle       string `json:"post_title" example:"技术分享帖"`               // Post title | 帖子标题
	UserID          int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username        string `json:"username" example:"testuser"`              // Username | 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                    // Parent comment ID | 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`             // Reply target user ID | 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`   // Reply target username | 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                // Comment content | 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                  // Like count | 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                // Dislike count | 点踩数
	IsSelected      bool   `json:"is_selected" example:"true"`               // Whether selected | 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	CommenterIP     string `json:"commenter_ip" example:"192.168.1.1"`       // Commenter IP | 评论者IP
	DeviceInfo      string `json:"device_info" example:"Chrome/Windows"`     // Device info | 设备信息
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}
