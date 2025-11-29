package schema

// OAuthProviderListRequest OAuth提供商列表查询请求体
type OAuthProviderListRequest struct {
	Provider string `form:"provider" example:"GitHub"` // 提供商类型筛选
	Enabled  *bool  `form:"enabled" example:"true"`    // 是否启用筛选
}

// OAuthProviderCreateRequest 创建OAuth提供商请求体
type OAuthProviderCreateRequest struct {
	Provider     string                 `json:"provider" binding:"required,oneof=QQ GitHub Apple Google Telegram FIDO2" example:"GitHub"` // 提供商类型
	ClientID     string                 `json:"client_id" binding:"required" example:"your_client_id"`                                    // 客户端ID
	ClientSecret string                 `json:"client_secret" binding:"required" example:"your_client_secret"`                            // 客户端密钥
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`                              // 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"`                          // Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`                                      // 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`                                 // 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                                              // 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                                             // 额外配置参数
	Enabled      bool                   `json:"enabled" example:"true"`                                                                   // 是否启用
	SortOrder    int                    `json:"sort_order" binding:"min=0" example:"0"`                                                   // 排序顺序
}

// OAuthProviderUpdateRequest 更新OAuth提供商请求体
type OAuthProviderUpdateRequest struct {
	ID           int                    `json:"id" binding:"required" example:"1"`                               // 提供商ID
	ClientID     string                 `json:"client_id" example:"your_client_id"`                              // 客户端ID
	ClientSecret string                 `json:"client_secret" example:"your_client_secret"`                      // 客户端密钥（为空则不更新）
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`             // 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`        // 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                     // 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                    // 额外配置参数
	Enabled      *bool                  `json:"enabled" example:"true"`                                          // 是否启用
	SortOrder    *int                   `json:"sort_order" binding:"omitempty,min=0" example:"0"`                // 排序顺序
}

// OAuthProviderStatusUpdateRequest 更新OAuth提供商状态请求体
type OAuthProviderStatusUpdateRequest struct {
	ID      int  `json:"id" binding:"required" example:"1"` // 提供商ID
	Enabled bool `json:"enabled" example:"true"`            // 是否启用
}

// OAuthProviderListItem OAuth提供商列表项响应体
type OAuthProviderListItem struct {
	ID          int      `json:"id" example:"1"`                                                  // 提供商ID
	Provider    string   `json:"provider" example:"GitHub"`                                       // 提供商类型
	ClientID    string   `json:"client_id" example:"your_client_id"`                              // 客户端ID
	AuthURL     string   `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // 授权URL
	TokenURL    string   `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token获取URL
	UserInfoURL string   `json:"user_info_url" example:"https://api.github.com/user"`             // 用户信息获取URL
	RedirectURL string   `json:"redirect_url" example:"https://example.com/auth/callback"`        // 回调URL
	Scopes      []string `json:"scopes" example:"user:email"`                                     // 请求范围
	Enabled     bool     `json:"enabled" example:"true"`                                          // 是否启用
	SortOrder   int      `json:"sort_order" example:"0"`                                          // 排序顺序
	CreatedAt   string   `json:"created_at" example:"2024-01-01 00:00:00"`                        // 创建时间
	UpdatedAt   string   `json:"updated_at" example:"2024-01-01 00:00:00"`                        // 更新时间
}

// OAuthProviderListResponse OAuth提供商列表响应体
type OAuthProviderListResponse struct {
	List []OAuthProviderListItem `json:"list"` // 提供商列表
}

// OAuthProviderDetailResponse OAuth提供商详情响应体
type OAuthProviderDetailResponse struct {
	ID           int                    `json:"id" example:"1"`                                                  // 提供商ID
	Provider     string                 `json:"provider" example:"GitHub"`                                       // 提供商类型
	ClientID     string                 `json:"client_id" example:"your_client_id"`                              // 客户端ID
	ClientSecret string                 `json:"client_secret" example:"***"`                                     // 客户端密钥（脱敏）
	AuthURL      string                 `json:"auth_url" example:"https://github.com/login/oauth/authorize"`     // 授权URL
	TokenURL     string                 `json:"token_url" example:"https://github.com/login/oauth/access_token"` // Token获取URL
	UserInfoURL  string                 `json:"user_info_url" example:"https://api.github.com/user"`             // 用户信息获取URL
	RedirectURL  string                 `json:"redirect_url" example:"https://example.com/auth/callback"`        // 回调URL
	Scopes       []string               `json:"scopes" example:"user:email"`                                     // 请求范围
	ExtraConfig  map[string]interface{} `json:"extra_config"`                                                    // 额外配置参数
	Enabled      bool                   `json:"enabled" example:"true"`                                          // 是否启用
	SortOrder    int                    `json:"sort_order" example:"0"`                                          // 排序顺序
	CreatedAt    string                 `json:"created_at" example:"2024-01-01 00:00:00"`                        // 创建时间
	UpdatedAt    string                 `json:"updated_at" example:"2024-01-01 00:00:00"`                        // 更新时间
}
