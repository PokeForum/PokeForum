package oauth

import (
	"context"
	"time"
)

// Provider OAuth提供商类型
type Provider string

const (
	ProviderQQ       Provider = "QQ"
	ProviderGitHub   Provider = "GitHub"
	ProviderApple    Provider = "Apple"
	ProviderGoogle   Provider = "Google"
	ProviderTelegram Provider = "Telegram"
	ProviderFIDO2    Provider = "FIDO2"
)

// Config OAuth配置信息
type Config struct {
	Provider     Provider               // OAuth提供商
	ClientID     string                 // 客户端ID
	ClientSecret string                 // 客户端密钥
	AuthURL      string                 // 授权URL
	TokenURL     string                 // Token获取URL
	UserInfoURL  string                 // 用户信息获取URL
	RedirectURL  string                 // 回调URL
	Scopes       []string               // 请求范围
	ExtraConfig  map[string]interface{} // 额外配置参数
}

// TokenResponse OAuth Token响应
type TokenResponse struct {
	AccessToken  string    // 访问令牌
	TokenType    string    // 令牌类型（通常是Bearer）
	ExpiresIn    int       // 过期时间（秒）
	RefreshToken string    // 刷新令牌
	Scope        string    // 授权范围
	ExpiresAt    time.Time // 过期时间点
}

// UserInfo OAuth用户信息
type UserInfo struct {
	ProviderUserID string                 // 第三方平台的用户唯一标识
	Username       string                 // 用户名
	Email          string                 // 邮箱
	Avatar         string                 // 头像URL
	ExtraData      map[string]interface{} // 额外的用户信息
}

// IProvider OAuth提供商接口
// 每个OAuth提供商需要实现此接口
type IProvider interface {
	// GetAuthURL 获取授权URL
	// state: 用于防止CSRF攻击的随机字符串
	// 返回用户应该访问的授权URL
	GetAuthURL(state string) string

	// ExchangeToken 使用授权码换取访问令牌
	// ctx: 上下文
	// code: 授权码
	// 返回Token响应或错误
	ExchangeToken(ctx context.Context, code string) (*TokenResponse, error)

	// GetUserInfo 获取用户信息
	// ctx: 上下文
	// accessToken: 访问令牌
	// 返回用户信息或错误
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

	// RefreshToken 刷新访问令牌
	// ctx: 上下文
	// refreshToken: 刷新令牌
	// 返回新的Token响应或错误
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// ValidateToken 验证访问令牌是否有效
	// ctx: 上下文
	// accessToken: 访问令牌
	// 返回是否有效及错误
	ValidateToken(ctx context.Context, accessToken string) (bool, error)
}

// IClient OAuth客户端接口
type IClient interface {
	// GetProvider 获取指定的OAuth提供商
	GetProvider(provider Provider) (IProvider, error)

	// RegisterProvider 注册OAuth提供商
	RegisterProvider(provider Provider, config *Config) error

	// UnregisterProvider 注销OAuth提供商
	UnregisterProvider(provider Provider) error

	// ListProviders 列出所有已注册的OAuth提供商
	ListProviders() []Provider
}
