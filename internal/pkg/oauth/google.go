package oauth

import (
	"context"
	"fmt"
)

// GoogleProvider Google OAuth提供商
type GoogleProvider struct {
	*BaseProvider
}

// NewGoogleProvider 创建Google OAuth提供商实例
func NewGoogleProvider(config *Config) (IProvider, error) {
	// 设置Google默认配置
	if config.AuthURL == "" {
		config.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://oauth2.googleapis.com/token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile", "email"}
	}

	return &GoogleProvider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL 获取Google授权URL
func (g *GoogleProvider) GetAuthURL(state string, redirectURL string) string {
	extraParams := map[string]string{
		"access_type": "offline", // 获取refresh_token
		"prompt":      "consent",
	}
	return g.BuildAuthURL(state, redirectURL, extraParams)
}

// ExchangeToken 使用授权码换取访问令牌
func (g *GoogleProvider) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	return g.ExchangeTokenByForm(ctx, code)
}

// GetUserInfo 获取Google用户信息
func (g *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	userInfoMap, err := g.GetUserInfoByJSON(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 解析Google用户信息
	userInfo := &UserInfo{
		ExtraData: userInfoMap,
	}

	// Google用户ID
	if id, ok := userInfoMap["id"].(string); ok {
		userInfo.ProviderUserID = id
	}

	// 用户名
	if name, ok := userInfoMap["name"].(string); ok {
		userInfo.Username = name
	}

	// 邮箱
	if email, ok := userInfoMap["email"].(string); ok {
		userInfo.Email = email
	}

	// 头像
	if picture, ok := userInfoMap["picture"].(string); ok {
		userInfo.Avatar = picture
	}

	if userInfo.ProviderUserID == "" {
		return nil, fmt.Errorf("%w: provider_user_id is empty", ErrGetUserInfoFailed)
	}

	return userInfo, nil
}

// RefreshToken 刷新访问令牌
func (g *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return g.RefreshTokenByForm(ctx, refreshToken)
}

// ValidateToken 验证访问令牌是否有效
func (g *GoogleProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	return g.ValidateTokenByUserInfo(ctx, accessToken)
}
