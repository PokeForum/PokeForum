package oauth

import (
	"context"
	"time"
)

// Provider OAuth provider type | OAuth提供商类型
type Provider string

const (
	ProviderQQ       Provider = "QQ"
	ProviderGitHub   Provider = "GitHub"
	ProviderGoogle   Provider = "Google"
	ProviderTelegram Provider = "Telegram"
	ProviderFIDO2    Provider = "FIDO2"
)

// Config OAuth configuration information | OAuth配置信息
type Config struct {
	Provider     Provider               // OAuth provider | OAuth提供商
	ClientID     string                 // Client ID | 客户端ID
	ClientSecret string                 // Client secret | 客户端密钥
	AuthURL      string                 // Authorization URL | 授权URL
	TokenURL     string                 // Token acquisition URL | Token获取URL
	UserInfoURL  string                 // User info acquisition URL | 用户信息获取URL
	Scopes       []string               // Request scope | 请求范围
	ExtraConfig  map[string]interface{} // Extra configuration parameters | 额外配置参数
}

// TokenResponse OAuth Token response | OAuth Token响应
type TokenResponse struct {
	AccessToken  string    // Access token | 访问令牌
	TokenType    string    // Token type (usually Bearer) | 令牌类型（通常是Bearer）
	ExpiresIn    int       // Expiration time (seconds) | 过期时间（秒）
	RefreshToken string    // Refresh token | 刷新令牌
	Scope        string    // Authorization scope | 授权范围
	ExpiresAt    time.Time // Expiration time point | 过期时间点
}

// UserInfo OAuth user information | OAuth用户信息
type UserInfo struct {
	ProviderUserID string                 // Unique user identifier from third-party platform | 第三方平台的用户唯一标识
	Username       string                 // Username | 用户名
	Email          string                 // Email | 邮箱
	Avatar         string                 // Avatar URL | 头像URL
	ExtraData      map[string]interface{} // Extra user information | 额外的用户信息
}

// IProvider OAuth provider interface | OAuth提供商接口
// Each OAuth provider needs to implement this interface | 每个OAuth提供商需要实现此接口
type IProvider interface {
	// GetAuthURL Get authorization URL | 获取授权URL
	// state: random string to prevent CSRF attacks | state: 用于防止CSRF攻击的随机字符串
	// redirectURL: callback URL from frontend | redirectURL: 前端传入的回调URL
	// Returns the authorization URL that users should visit | 返回用户应该访问的授权URL
	GetAuthURL(state string, redirectURL string) string

	// ExchangeToken Exchange authorization code for access token | 使用授权码换取访问令牌
	// ctx: context | ctx: 上下文
	// code: authorization code | code: 授权码
	// redirectURI: redirect URI used in authorization request | redirectURI: 授权请求中使用的重定向URI
	// Returns Token response or error | 返回Token响应或错误
	ExchangeToken(ctx context.Context, code string, redirectURI string) (*TokenResponse, error)

	// GetUserInfo Get user information | 获取用户信息
	// ctx: context | ctx: 上下文
	// accessToken: access token | accessToken: 访问令牌
	// Returns user information or error | 返回用户信息或错误
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

	// RefreshToken Refresh access token | 刷新访问令牌
	// ctx: context | ctx: 上下文
	// refreshToken: refresh token | refreshToken: 刷新令牌
	// Returns new Token response or error | 返回新的Token响应或错误
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// ValidateToken Validate if access token is valid | 验证访问令牌是否有效
	// ctx: context | ctx: 上下文
	// accessToken: access token | accessToken: 访问令牌
	// Returns whether valid and error | 返回是否有效及错误
	ValidateToken(ctx context.Context, accessToken string) (bool, error)
}

// IClient OAuth client interface | OAuth客户端接口
type IClient interface {
	// GetProvider Get specified OAuth provider | 获取指定的OAuth提供商
	GetProvider(provider Provider) (IProvider, error)

	// RegisterProvider Register OAuth provider | 注册OAuth提供商
	RegisterProvider(provider Provider, config *Config) error

	// UnregisterProvider Unregister OAuth provider | 注销OAuth提供商
	UnregisterProvider(provider Provider) error

	// ListProviders List all registered OAuth providers | 列出所有已注册的OAuth提供商
	ListProviders() []Provider
}
