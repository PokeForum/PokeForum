package schema

// OAuthProviderListRequest OAuth provider list query request | OAuth提供商列表查询请求体
type OAuthProviderListRequest struct {
	Provider string `form:"provider" example:"GitHub"` // Provider type filter | 提供商类型筛选
	Enabled  *bool  `form:"enabled" example:"true"`    // Enabled status filter | 是否启用筛选
}

// OAuthProviderCreateRequest Create OAuth provider request | 创建OAuth提供商请求体
type OAuthProviderCreateRequest struct {
	Provider     string                 `json:"provider" binding:"required,oneof=QQ GitHub Apple Google Telegram FIDO2" example:"GitHub"` // Provider type | 提供商类型
	ClientID     string                 `json:"client_id" binding:"required" example:"your_client_id"`                                    // Client ID | 客户端ID
	ClientSecret string                 `json:"client_secret" binding:"required" example:"your_client_secret"`                            // Client secret | 客户端密钥
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`                              // Authorization URL | 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"`                          // Token URL | Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`                                      // User info URL | 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`                                 // Callback URL | 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                                              // Request scopes | 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                                             // Extra config parameters | 额外配置参数
	Enabled      bool                   `json:"enabled" example:"true"`                                                                   // Whether enabled | 是否启用
	SortOrder    int                    `json:"sort_order" binding:"min=0" example:"0"`                                                   // Sort order | 排序顺序
}

// OAuthProviderUpdateRequest Update OAuth provider request | 更新OAuth提供商请求体
type OAuthProviderUpdateRequest struct {
	ID           int                    `json:"id" binding:"required" example:"1"`                               // Provider ID | 提供商ID
	ClientID     string                 `json:"client_id" example:"your_client_id"`                              // Client ID | 客户端ID
	ClientSecret string                 `json:"client_secret" example:"your_client_secret"`                      // Client secret (empty to skip update) | 客户端密钥（为空则不更新）
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // Authorization URL | 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token URL | Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`             // User info URL | 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`        // Callback URL | 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                     // Request scopes | 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                    // Extra config parameters | 额外配置参数
	Enabled      *bool                  `json:"enabled" example:"true"`                                          // Whether enabled | 是否启用
	SortOrder    *int                   `json:"sort_order" binding:"omitempty,min=0" example:"0"`                // Sort order | 排序顺序
}

// OAuthProviderStatusUpdateRequest Update OAuth provider status request | 更新OAuth提供商状态请求体
type OAuthProviderStatusUpdateRequest struct {
	ID      int  `json:"id" binding:"required" example:"1"` // Provider ID | 提供商ID
	Enabled bool `json:"enabled" example:"true"`            // Whether enabled | 是否启用
}

// OAuthProviderListItem OAuth provider list item response | OAuth提供商列表项响应体
type OAuthProviderListItem struct {
	ID          int      `json:"id" example:"1"`                                                  // Provider ID | 提供商ID
	Provider    string   `json:"provider" example:"GitHub"`                                       // Provider type | 提供商类型
	ClientID    string   `json:"client_id" example:"your_client_id"`                              // Client ID | 客户端ID
	AuthURL     string   `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // Authorization URL | 授权URL
	TokenURL    string   `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token URL | Token获取URL
	UserInfoURL string   `json:"user_info_url" example:"https://api.github.com/user"`             // User info URL | 用户信息获取URL
	RedirectURL string   `json:"redirect_url" example:"https://example.com/auth/callback"`        // Callback URL | 回调URL
	Scopes      []string `json:"scopes" example:"user:email"`                                     // Request scopes | 请求范围
	Enabled     bool     `json:"enabled" example:"true"`                                          // Whether enabled | 是否启用
	SortOrder   int      `json:"sort_order" example:"0"`                                          // Sort order | 排序顺序
	CreatedAt   string   `json:"created_at" example:"2024-01-01 00:00:00"`                        // Creation time | 创建时间
	UpdatedAt   string   `json:"updated_at" example:"2024-01-01 00:00:00"`                        // Update time | 更新时间
}

// OAuthProviderListResponse OAuth provider list response | OAuth提供商列表响应体
type OAuthProviderListResponse struct {
	List []OAuthProviderListItem `json:"list"` // Provider list | 提供商列表
}

// OAuthProviderDetailResponse OAuth provider detail response | OAuth提供商详情响应体
type OAuthProviderDetailResponse struct {
	ID           int                    `json:"id" example:"1"`                                                  // Provider ID | 提供商ID
	Provider     string                 `json:"provider" example:"GitHub"`                                       // Provider type | 提供商类型
	ClientID     string                 `json:"client_id" example:"your_client_id"`                              // Client ID | 客户端ID
	ClientSecret string                 `json:"client_secret" example:"***"`                                     // Client secret (masked) | 客户端密钥（脱敏）
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // Authorization URL | 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token URL | Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`             // User info URL | 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`        // Callback URL | 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                     // Request scopes | 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                    // Extra config parameters | 额外配置参数
	Enabled      bool                   `json:"enabled" example:"true"`                                          // Whether enabled | 是否启用
	SortOrder    int                    `json:"sort_order" example:"0"`                                          // Sort order | 排序顺序
	CreatedAt    string                 `json:"created_at" example:"2024-01-01 00:00:00"`                        // Creation time | 创建时间
	UpdatedAt    string                 `json:"updated_at" example:"2024-01-01 00:00:00"`                        // Update time | 更新时间
}
