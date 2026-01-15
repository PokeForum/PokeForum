package schema

// OAuthProviderPublicItem OAuth provider public info for frontend | OAuth提供商公开信息（前端展示用）
type OAuthProviderPublicItem struct {
	Provider  string `json:"provider" example:"GitHub"` // Provider type | 提供商类型
	SortOrder int    `json:"sort_order" example:"0"`    // Sort order | 排序顺序
}

// OAuthProviderPublicListResponse OAuth provider public list response | OAuth提供商公开列表响应
type OAuthProviderPublicListResponse struct {
	List []OAuthProviderPublicItem `json:"list"` // Provider list | 提供商列表
}

// OAuthAuthorizeRequest Get authorization URL request | 获取授权URL请求
type OAuthAuthorizeRequest struct {
	RedirectURI string `form:"redirect_uri" binding:"required,url" example:"https://example.com/oauth/callback"` // Frontend callback URL | 前端回调地址
}

// OAuthAuthorizeResponse Authorization URL response | 授权URL响应
type OAuthAuthorizeResponse struct {
	AuthorizeURL string `json:"authorize_url"` // Full authorization URL | 完整授权URL
	State        string `json:"state"`         // State parameter (frontend needs to pass back) | state参数(前端需回传)
}

// OAuthCallbackRequest OAuth callback request | OAuth回调请求
type OAuthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`  // Authorization code | 授权码
	State string `json:"state" binding:"required"` // State parameter | state参数
}

// OAuthCallbackResponse OAuth callback response | OAuth回调响应
type OAuthCallbackResponse struct {
	Action   string `json:"action"`            // Executed action: login/register/bindRequired | 执行的操作: login/register/bindRequired
	UserID   int    `json:"user_id"`           // User ID | 用户ID
	Username string `json:"username"`          // Username | 用户名
	Token    string `json:"token,omitempty"`   // Token (returned for login/register) | Token（登录/注册场景返回）
	Message  string `json:"message,omitempty"` // Message | 消息
}

// OAuthUserBindListResponse User OAuth binding list response | 用户OAuth绑定列表响应
type OAuthUserBindListResponse struct {
	List []OAuthUserBindItem `json:"list"` // Binding list | 绑定列表
}

// OAuthUserBindItem User OAuth binding item | 用户OAuth绑定项
type OAuthUserBindItem struct {
	Provider         string `json:"provider" example:"GitHub"`              // Provider type | 提供商类型
	ProviderUsername string `json:"provider_username" example:"octocat"`    // Provider username | 提供商用户名
	ProviderAvatar   string `json:"provider_avatar"`                        // Provider avatar | 提供商头像
	BoundAt          string `json:"bound_at" example:"2024-01-01 00:00:00"` // Binding time | 绑定时间
}

// OAuthBindCallbackRequest OAuth bind callback request | OAuth绑定回调请求
type OAuthBindCallbackRequest struct {
	Code  string `json:"code" binding:"required"`  // Authorization code | 授权码
	State string `json:"state" binding:"required"` // State parameter | state参数
}

// OAuthUnbindRequest OAuth unbind request | OAuth解绑请求
type OAuthUnbindRequest struct {
	Provider string `uri:"provider" binding:"required,oneof=QQ GitHub Google"` // Provider type | 提供商类型
}
