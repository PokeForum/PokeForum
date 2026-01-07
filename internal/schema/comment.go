package schema

// UserCommentCreateRequest User create comment request | 用户创建评论请求体
type UserCommentCreateRequest struct {
	PostID        int    `json:"post_id" binding:"required" example:"1"`             // Post ID | 帖子ID
	ParentID      *int   `json:"parent_id" example:"1"`                              // Parent comment ID (used when replying to comment) | 父评论ID（回复评论时使用）
	ReplyToUserID *int   `json:"reply_to_user_id" example:"2"`                       // Reply target user ID (used when replying to user) | 回复目标用户ID（回复用户时使用）
	Content       string `json:"content" binding:"required,min=1" example:"很有见地的评论"` // Comment content | 评论内容
}

// UserCommentCreateResponse User create comment response | 用户创建评论响应体
type UserCommentCreateResponse struct {
	ID              int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	PostID          int    `json:"post_id" example:"1"`                      // Post ID | 帖子ID
	ParentID        *int   `json:"parent_id" example:"1"`                    // Parent comment ID | 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`             // Reply target user ID | 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`   // Reply target username | 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                // Comment content | 评论内容
	LikeCount       int    `json:"like_count" example:"0"`                   // Like count | 点赞数
	DislikeCount    int    `json:"dislike_count" example:"0"`                // Dislike count | 点踩数
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
}

// UserCommentUpdateRequest User update comment request | 用户更新评论请求体
type UserCommentUpdateRequest struct {
	ID      int    `json:"id" binding:"required" example:"1"`                   // Comment ID | 评论ID
	Content string `json:"content" binding:"required,min=1" example:"更新后的评论内容"` // Comment content | 评论内容
}

// UserCommentUpdateResponse User update comment response | 用户更新评论响应体
type UserCommentUpdateResponse struct {
	ID        int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	Content   string `json:"content" example:"更新后的评论内容"`               // Comment content | 评论内容
	UpdatedAt string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}

// UserCommentActionRequest User comment action request | 用户评论操作请求体
type UserCommentActionRequest struct {
	ID int `json:"id" binding:"required" example:"1"` // Comment ID | 评论ID
}

// UserCommentActionResponse User comment action response | 用户评论操作响应体
type UserCommentActionResponse struct {
	ID           int `json:"id" example:"1"`            // Comment ID | 评论ID
	LikeCount    int `json:"like_count" example:"10"`   // Like count | 点赞数
	DislikeCount int `json:"dislike_count" example:"1"` // Dislike count | 点踩数
}

// UserCommentListRequest User comment list request | 用户评论列表请求体
type UserCommentListRequest struct {
	PostID   int    `form:"post_id" binding:"required" example:"1"`                 // Post ID | 帖子ID
	Page     int    `form:"page" binding:"required,min=1" example:"1"`              // Page number | 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=50" example:"20"` // Items per page | 每页数量
	SortBy   string `form:"sort_by" example:"created_at"`                           // Sort field: created_at, like_count | 排序字段：created_at, like_count
	SortDesc bool   `form:"sort_desc" example:"true"`                               // Descending order | 是否降序
}

// UserCommentListItem User comment list item response | 用户评论列表项响应体
type UserCommentListItem struct {
	ID              int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	FloorNumber     int    `json:"floor_number" example:"1"`                 // Floor number | 楼号
	UserID          int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username        string `json:"username" example:"testuser"`              // Username | 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                    // Parent comment ID | 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`             // Reply target user ID | 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`   // Reply target username | 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                // Comment content | 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                  // Like count | 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                // Dislike count | 点踩数
	UserLiked       bool   `json:"user_liked" example:"false"`               // Whether current user has liked | 当前用户是否已点赞
	UserDisliked    bool   `json:"user_disliked" example:"false"`            // Whether current user has disliked | 当前用户是否已点踩
	IsSelected      bool   `json:"is_selected" example:"true"`               // Whether selected | 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}

// UserCommentListResponse User comment list response | 用户评论列表响应体
type UserCommentListResponse struct {
	List     []UserCommentListItem `json:"list"`      // Comment list | 评论列表
	Total    int64                 `json:"total"`     // Total count | 总数量
	Page     int                   `json:"page"`      // Current page number | 当前页码
	PageSize int                   `json:"page_size"` // Items per page | 每页数量
}

// UserCommentDetailResponse User comment detail response | 用户评论详情响应体
type UserCommentDetailResponse struct {
	ID              int    `json:"id" example:"1"`                           // Comment ID | 评论ID
	PostID          int    `json:"post_id" example:"1"`                      // Post ID | 帖子ID
	UserID          int    `json:"user_id" example:"1"`                      // User ID | 用户ID
	Username        string `json:"username" example:"testuser"`              // Username | 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                    // Parent comment ID | 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`             // Reply target user ID | 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`   // Reply target username | 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                // Comment content | 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                  // Like count | 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                // Dislike count | 点踩数
	UserLiked       bool   `json:"user_liked" example:"false"`               // Whether current user has liked | 当前用户是否已点赞
	UserDisliked    bool   `json:"user_disliked" example:"false"`            // Whether current user has disliked | 当前用户是否已点踩
	IsSelected      bool   `json:"is_selected" example:"true"`               // Whether selected | 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                // Whether pinned | 是否置顶
	CreatedAt       string `json:"created_at" example:"2024-01-01 00:00:00"` // Creation time | 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01 00:00:00"` // Update time | 更新时间
}
