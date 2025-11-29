package schema

// UserProfileOverviewResponse 用户个人中心概览响应体
type UserProfileOverviewResponse struct {
	ID            int    `json:"id" example:"1"`                                  // 用户ID
	Username      string `json:"username" example:"testuser"`                     // 用户名
	Email         string `json:"email" example:"test@example.com"`                // 邮箱
	Avatar        string `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
	Signature     string `json:"signature" example:"这是我的个性签名"`                    // 签名
	Readme        string `json:"readme" example:"# 关于我\n这是我的自我介绍"`                // README
	EmailVerified bool   `json:"email_verified" example:"true"`                   // 邮箱是否已验证
	Points        int    `json:"points" example:"100"`                            // 积分
	Currency      int    `json:"currency" example:"50"`                           // 货币
	PostCount     int    `json:"post_count" example:"10"`                         // 帖子数
	CommentCount  int    `json:"comment_count" example:"20"`                      // 评论数
	Status        string `json:"status" example:"Normal"`                         // 用户状态
	Role          string `json:"role" example:"User"`                             // 用户身份
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"`        // 创建时间
}

// UserProfileOverviewRequest 用户个人中心概览请求体
type UserProfileOverviewRequest struct {
	UserID int `form:"user_id" example:"1"` // 用户ID，不传则查询当前登录用户
}

// UserProfilePostsRequest 用户主题帖列表请求体
type UserProfilePostsRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`              // 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=50" example:"20"` // 每页数量
	Status   string `form:"status" example:"Normal"`                                // 帖子状态筛选：Normal、Draft、Private
	UserID   int    `form:"user_id" example:"1"`                                    // 用户ID，不传则查询当前登录用户
}

// UserProfilePostItem 用户主题帖列表项
type UserProfilePostItem struct {
	ID            int    `json:"id" example:"1"`                           // 帖子ID
	CategoryID    int    `json:"category_id" example:"1"`                  // 版块ID
	CategoryName  string `json:"category_name" example:"技术讨论"`             // 版块名称
	Title         string `json:"title" example:"我的第一个帖子"`                  // 帖子标题
	ViewCount     int    `json:"view_count" example:"100"`                 // 浏览数
	LikeCount     int    `json:"like_count" example:"10"`                  // 点赞数
	DislikeCount  int    `json:"dislike_count" example:"1"`                // 点踩数
	FavoriteCount int    `json:"favorite_count" example:"5"`               // 收藏数
	IsEssence     bool   `json:"is_essence" example:"false"`               // 是否精华帖
	IsPinned      bool   `json:"is_pinned" example:"false"`                // 是否置顶
	Status        string `json:"status" example:"Normal"`                  // 帖子状态
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
}

// UserProfilePostsResponse 用户主题帖列表响应体
type UserProfilePostsResponse struct {
	List     []UserProfilePostItem `json:"list"`      // 帖子列表
	Total    int64                 `json:"total"`     // 总数量
	Page     int                   `json:"page"`      // 当前页码
	PageSize int                   `json:"page_size"` // 每页数量
}

// UserProfileCommentsRequest 用户评论列表请求体
type UserProfileCommentsRequest struct {
	Page     int `form:"page" binding:"required,min=1" example:"1"`              // 页码
	PageSize int `form:"page_size" binding:"required,min=1,max=50" example:"20"` // 每页数量
	UserID   int `form:"user_id" example:"1"`                                    // 用户ID，不传则查询当前登录用户
}

// UserProfileCommentItem 用户评论列表项
type UserProfileCommentItem struct {
	ID           int    `json:"id" example:"1"`                           // 评论ID
	PostID       int    `json:"post_id" example:"1"`                      // 帖子ID
	PostTitle    string `json:"post_title" example:"我的第一个帖子"`             // 帖子标题
	Content      string `json:"content" example:"很有见地的评论"`                // 评论内容
	LikeCount    int    `json:"like_count" example:"10"`                  // 点赞数
	DislikeCount int    `json:"dislike_count" example:"1"`                // 点踩数
	CreatedAt    string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
}

// UserProfileCommentsResponse 用户评论列表响应体
type UserProfileCommentsResponse struct {
	List     []UserProfileCommentItem `json:"list"`      // 评论列表
	Total    int64                    `json:"total"`     // 总数量
	Page     int                      `json:"page"`      // 当前页码
	PageSize int                      `json:"page_size"` // 每页数量
}

// UserProfileFavoritesRequest 用户收藏列表请求体
type UserProfileFavoritesRequest struct {
	Page     int `form:"page" binding:"required,min=1" example:"1"`              // 页码
	PageSize int `form:"page_size" binding:"required,min=1,max=50" example:"20"` // 每页数量
	UserID   int `form:"user_id" example:"1"`                                    // 用户ID，不传则查询当前登录用户
}

// UserProfileFavoriteItem 用户收藏列表项
type UserProfileFavoriteItem struct {
	ID            int    `json:"id" example:"1"`                             // 帖子ID
	CategoryID    int    `json:"category_id" example:"1"`                    // 版块ID
	CategoryName  string `json:"category_name" example:"技术讨论"`               // 版块名称
	Title         string `json:"title" example:"我的第一个帖子"`                    // 帖子标题
	Username      string `json:"username" example:"testuser"`                // 作者用户名
	ViewCount     int    `json:"view_count" example:"100"`                   // 浏览数
	LikeCount     int    `json:"like_count" example:"10"`                    // 点赞数
	FavoriteCount int    `json:"favorite_count" example:"5"`                 // 收藏数
	CreatedAt     string `json:"created_at" example:"2024-01-01 00:00:00"`   // 帖子创建时间
	FavoritedAt   string `json:"favorited_at" example:"2024-01-02 00:00:00"` // 收藏时间
}

// UserProfileFavoritesResponse 用户收藏列表响应体
type UserProfileFavoritesResponse struct {
	List     []UserProfileFavoriteItem `json:"list"`      // 收藏列表
	Total    int64                     `json:"total"`     // 总数量
	Page     int                       `json:"page"`      // 当前页码
	PageSize int                       `json:"page_size"` // 每页数量
}

// UserUpdatePasswordRequest 修改密码请求体
type UserUpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=64" example:"oldpass123"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=64" example:"newpass123"` // 新密码
}

// UserUpdatePasswordResponse 修改密码响应体
type UserUpdatePasswordResponse struct {
	Success bool   `json:"success" example:"true"`         // 是否成功
	Message string `json:"message" example:"密码修改成功，请重新登录"` // 提示信息
}

// UserUpdateAvatarRequest 修改头像请求体
type UserUpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url" binding:"required,url" example:"https://example.com/avatar.jpg"` // 头像URL
}

// UserUpdateAvatarResponse 修改头像响应体
type UserUpdateAvatarResponse struct {
	Success   bool   `json:"success" example:"true"`                              // 是否成功
	AvatarURL string `json:"avatar_url" example:"https://example.com/avatar.jpg"` // 新的头像URL
}

// UserUpdateUsernameRequest 修改用户名请求体
type UserUpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=2,max=20" example:"newusername"` // 新用户名
}

// UserUpdateUsernameResponse 修改用户名响应体
type UserUpdateUsernameResponse struct {
	Success  bool   `json:"success" example:"true"`         // 是否成功
	Username string `json:"username" example:"newusername"` // 新用户名
}

// EmailVerifyCodeResponse 发送邮箱验证码响应体
type EmailVerifyCodeResponse struct {
	Sent      bool   `json:"sent" example:"true"`               // 验证码发送状态
	Message   string `json:"message" example:"验证码已发送到您的邮箱，请查收"` // 提示信息
	ExpiresIn int    `json:"expires_in" example:"600"`          // 验证码有效期（秒）
}

// EmailVerifyRequest 验证邮箱请求体
type EmailVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6" example:"123456"` // 验证码
}

// EmailVerifyResponse 验证邮箱响应体
type EmailVerifyResponse struct {
	Verified bool   `json:"verified" example:"true"`  // 验证状态
	Message  string `json:"message" example:"邮箱验证成功"` // 提示信息
}
