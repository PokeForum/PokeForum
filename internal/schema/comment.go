package schema

// UserCommentCreateRequest 用户创建评论请求体
type UserCommentCreateRequest struct {
	PostID        int    `json:"post_id" binding:"required" example:"1"`                      // 帖子ID
	ParentID      *int   `json:"parent_id" example:"1"`                                       // 父评论ID（回复评论时使用）
	ReplyToUserID *int   `json:"reply_to_user_id" example:"2"`                                // 回复目标用户ID（回复用户时使用）
	Content       string `json:"content" binding:"required,min=1,max=1000" example:"很有见地的评论"` // 评论内容
}

// UserCommentCreateResponse 用户创建评论响应体
type UserCommentCreateResponse struct {
	ID              int    `json:"id" example:"1"`                            // 评论ID
	PostID          int    `json:"post_id" example:"1"`                       // 帖子ID
	ParentID        *int   `json:"parent_id" example:"1"`                     // 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`              // 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`    // 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                 // 评论内容
	LikeCount       int    `json:"like_count" example:"0"`                    // 点赞数
	DislikeCount    int    `json:"dislike_count" example:"0"`                 // 点踩数
	CreatedAt       string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
}

// UserCommentUpdateRequest 用户更新评论请求体
type UserCommentUpdateRequest struct {
	ID      int    `json:"id" binding:"required" example:"1"`                            // 评论ID
	Content string `json:"content" binding:"required,min=1,max=1000" example:"更新后的评论内容"` // 评论内容
}

// UserCommentUpdateResponse 用户更新评论响应体
type UserCommentUpdateResponse struct {
	ID        int    `json:"id" example:"1"`                            // 评论ID
	Content   string `json:"content" example:"更新后的评论内容"`                // 评论内容
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"` // 更新时间
}

// UserCommentActionRequest 用户评论操作请求体
type UserCommentActionRequest struct {
	ID int `json:"id" binding:"required" example:"1"` // 评论ID
}

// UserCommentActionResponse 用户评论操作响应体
type UserCommentActionResponse struct {
	ID           int `json:"id" example:"1"`            // 评论ID
	LikeCount    int `json:"like_count" example:"10"`   // 点赞数
	DislikeCount int `json:"dislike_count" example:"1"` // 点踩数
}

// UserCommentListRequest 用户评论列表请求体
type UserCommentListRequest struct {
	PostID   int    `form:"post_id" binding:"required" example:"1"`                 // 帖子ID
	Page     int    `form:"page" binding:"required,min=1" example:"1"`              // 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=50" example:"20"` // 每页数量
	SortBy   string `form:"sort_by" example:"created_at"`                           // 排序字段：created_at, like_count
	SortDesc bool   `form:"sort_desc" example:"true"`                               // 是否降序
}

// UserCommentListItem 用户评论列表项响应体
type UserCommentListItem struct {
	ID              int    `json:"id" example:"1"`                            // 评论ID
	FloorNumber     int    `json:"floor_number" example:"1"`                  // 楼号
	UserID          int    `json:"user_id" example:"1"`                       // 用户ID
	Username        string `json:"username" example:"testuser"`               // 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                     // 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`              // 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`    // 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                 // 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                   // 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                 // 点踩数
	IsSelected      bool   `json:"is_selected" example:"true"`                // 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                 // 是否置顶
	CreatedAt       string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01T00:00:00Z"` // 更新时间
}

// UserCommentListResponse 用户评论列表响应体
type UserCommentListResponse struct {
	List     []UserCommentListItem `json:"list"`      // 评论列表
	Total    int64                 `json:"total"`     // 总数量
	Page     int                   `json:"page"`      // 当前页码
	PageSize int                   `json:"page_size"` // 每页数量
}

// UserCommentDetailResponse 用户评论详情响应体
type UserCommentDetailResponse struct {
	ID              int    `json:"id" example:"1"`                            // 评论ID
	PostID          int    `json:"post_id" example:"1"`                       // 帖子ID
	UserID          int    `json:"user_id" example:"1"`                       // 用户ID
	Username        string `json:"username" example:"testuser"`               // 用户名
	ParentID        *int   `json:"parent_id" example:"1"`                     // 父评论ID
	ReplyToUserID   *int   `json:"reply_to_user_id" example:"2"`              // 回复目标用户ID
	ReplyToUsername string `json:"reply_to_username" example:"targetuser"`    // 回复目标用户名
	Content         string `json:"content" example:"很有见地的评论"`                 // 评论内容
	LikeCount       int    `json:"like_count" example:"10"`                   // 点赞数
	DislikeCount    int    `json:"dislike_count" example:"1"`                 // 点踩数
	IsSelected      bool   `json:"is_selected" example:"true"`                // 是否精选
	IsPinned        bool   `json:"is_pinned" example:"false"`                 // 是否置顶
	CreatedAt       string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
	UpdatedAt       string `json:"updated_at" example:"2024-01-01T00:00:00Z"` // 更新时间
}
