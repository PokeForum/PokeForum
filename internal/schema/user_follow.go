package schema

// UserFollowRequest Follow user request | 关注用户请求
type UserFollowRequest struct {
	FollowingID int `json:"following_id" binding:"required,min=1" example:"123"`
}

// UserUnfollowRequest Unfollow user request | 取消关注请求
type UserUnfollowRequest struct {
	FollowingID int `uri:"user_id" binding:"required,min=1" example:"123"`
}

// UserFollowResponse Follow user response | 关注用户响应
type UserFollowResponse struct {
	Success     bool   `json:"success" example:"true"`
	Message     string `json:"message" example:"关注成功"`
	FollowingID int    `json:"following_id" example:"123"`
}

// UserUnfollowResponse Unfollow user response | 取消关注响应
type UserUnfollowResponse struct {
	Success     bool   `json:"success" example:"true"`
	Message     string `json:"message" example:"取消关注成功"`
	FollowingID int    `json:"following_id" example:"123"`
}

// UserFollowersRequest Get user followers request | 获取用户粉丝列表请求
type UserFollowersRequest struct {
	UserID   int `form:"user_id" example:"123"`  // User ID, query current logged in user if not provided | 用户ID，不传则查询当前登录用户
	Page     int `form:"page" example:"1"`       // Page number, default 1 | 页码，默认1
	PageSize int `form:"page_size" example:"20"` // Items per page, default 20 | 每页数量，默认20
}

// UserFollowingRequest Get user following request | 获取用户关注列表请求
type UserFollowingRequest struct {
	UserID   int `form:"user_id" example:"123"`  // User ID, query current logged in user if not provided | 用户ID，不传则查询当前登录用户
	Page     int `form:"page" example:"1"`       // Page number, default 1 | 页码，默认1
	PageSize int `form:"page_size" example:"20"` // Items per page, default 20 | 每页数量，默认20
}

// UserFollowItem Follow/Follower item | 关注/粉丝项
type UserFollowItem struct {
	UserID      int    `json:"user_id" example:"123"`
	Username    string `json:"username" example:"username"`
	Avatar      string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Signature   string `json:"signature" example:"这是个性签名"`
	IsFollowing bool   `json:"is_following" example:"true"` // 当前用户是否关注此人
	IsMutual    bool   `json:"is_mutual" example:"false"`   // 是否互相关注
	FollowedAt  string `json:"followed_at" example:"2024-01-20 12:00:00"`
}

// UserFollowersResponse Get user followers response | 获取用户粉丝列表响应
type UserFollowersResponse struct {
	List     []UserFollowItem `json:"list"`
	Total    int64            `json:"total" example:"100"`
	Page     int              `json:"page" example:"1"`
	PageSize int              `json:"page_size" example:"20"`
}

// UserFollowingResponse Get user following response | 获取用户关注列表响应
type UserFollowingResponse struct {
	List     []UserFollowItem `json:"list"`
	Total    int64            `json:"total" example:"50"`
	Page     int              `json:"page" example:"1"`
	PageSize int              `json:"page_size" example:"20"`
}

// UserFollowStatusResponse Get follow status response | 获取关注状态响应
type UserFollowStatusResponse struct {
	IsFollowing bool `json:"is_following" example:"true"` // 当前用户是否关注目标用户
	IsFollower  bool `json:"is_follower" example:"false"` // 目标用户是否关注当前用户
	IsMutual    bool `json:"is_mutual" example:"false"`   // 是否互相关注
}
