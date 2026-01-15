package oauth

import (
	"context"
	"fmt"
)

// GitHubProvider GitHub OAuth提供商
type GitHubProvider struct {
	*BaseProvider
}

// NewGitHubProvider 创建GitHub OAuth提供商实例
func NewGitHubProvider(config *Config) (IProvider, error) {
	// 设置GitHub默认配置
	if config.AuthURL == "" {
		config.AuthURL = "https://github.com/login/oauth/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://github.com/login/oauth/access_token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://api.github.com/user"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"user:email"}
	}

	return &GitHubProvider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL 获取GitHub授权URL
func (g *GitHubProvider) GetAuthURL(state string, redirectURL string) string {
	return g.BuildAuthURL(state, redirectURL, nil)
}

// ExchangeToken 使用授权码换取访问令牌
func (g *GitHubProvider) ExchangeToken(ctx context.Context, code string, _ string) (*TokenResponse, error) {
	return g.ExchangeTokenByForm(ctx, code)
}

// GetUserInfo 获取GitHub用户信息
func (g *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	userInfoMap, err := g.GetUserInfoByJSON(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 解析GitHub用户信息
	userInfo := &UserInfo{
		ExtraData: userInfoMap,
	}

	// GitHub用户ID
	if id, ok := userInfoMap["id"].(float64); ok {
		userInfo.ProviderUserID = fmt.Sprintf("%.0f", id)
	}

	// 用户名
	if login, ok := userInfoMap["login"].(string); ok {
		userInfo.Username = login
	}

	// 邮箱
	if email, ok := userInfoMap["email"].(string); ok {
		userInfo.Email = email
	}

	// 头像
	if avatar, ok := userInfoMap["avatar_url"].(string); ok {
		userInfo.Avatar = avatar
	}

	if userInfo.ProviderUserID == "" {
		return nil, fmt.Errorf("%w: provider_user_id is empty", ErrGetUserInfoFailed)
	}

	return userInfo, nil
}

// RefreshToken GitHub不支持刷新令牌
func (g *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return nil, fmt.Errorf("github does not support refresh token")
}

// ValidateToken 验证访问令牌是否有效
func (g *GitHubProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	return g.ValidateTokenByUserInfo(ctx, accessToken)
}
