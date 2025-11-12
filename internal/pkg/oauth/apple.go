package oauth

import (
	"context"
	"fmt"
)

// AppleProvider Apple OAuth提供商
// Apple使用Sign in with Apple，需要特殊处理
type AppleProvider struct {
	*BaseProvider
}

// NewAppleProvider 创建Apple OAuth提供商实例
func NewAppleProvider(config *Config) (IProvider, error) {
	// 设置Apple默认配置
	if config.AuthURL == "" {
		config.AuthURL = "https://appleid.apple.com/auth/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://appleid.apple.com/auth/token"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"name", "email"}
	}

	return &AppleProvider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL 获取Apple授权URL
func (a *AppleProvider) GetAuthURL(state string) string {
	extraParams := map[string]string{
		"response_mode": "form_post", // Apple推荐使用form_post
	}
	return a.BuildAuthURL(state, extraParams)
}

// ExchangeToken 使用授权码换取访问令牌
// Apple的实现需要使用JWT作为client_secret
func (a *AppleProvider) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	// Apple需要使用JWT作为client_secret
	// 这里使用标准的表单方式，但client_secret应该是预先生成的JWT
	return a.ExchangeTokenByForm(ctx, code)
}

// GetUserInfo 获取Apple用户信息
// 注意：Apple只在首次授权时返回用户信息，后续需要从ID Token中解析
func (a *AppleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// Apple不提供标准的用户信息端点
	// 用户信息需要从ID Token（JWT）中解析
	// 这里返回一个基础实现，实际使用时需要解析ID Token
	return nil, fmt.Errorf("apple provider requires parsing ID token for user info")
}

// RefreshToken 刷新访问令牌
func (a *AppleProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return a.RefreshTokenByForm(ctx, refreshToken)
}

// ValidateToken 验证访问令牌是否有效
// Apple需要验证ID Token
func (a *AppleProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	// Apple的Token验证需要验证JWT签名
	return false, fmt.Errorf("apple provider requires JWT validation")
}
