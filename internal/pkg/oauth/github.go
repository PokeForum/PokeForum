package oauth

import (
	"context"
	"fmt"
)

// GitHubProvider GitHub OAuth provider | GitHub OAuth提供商
type GitHubProvider struct {
	*BaseProvider
}

// NewGitHubProvider Create GitHub OAuth provider instance | 创建GitHub OAuth提供商实例
func NewGitHubProvider(config *Config) (IProvider, error) {
	// Set GitHub default configuration | 设置GitHub默认配置
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

// GetAuthURL Get GitHub authorization URL | 获取GitHub授权URL
func (g *GitHubProvider) GetAuthURL(state string, redirectURL string) string {
	return g.BuildAuthURL(state, redirectURL, nil)
}

// ExchangeToken Exchange authorization code for access token | 使用授权码换取访问令牌
func (g *GitHubProvider) ExchangeToken(ctx context.Context, code string, redirectURI string) (*TokenResponse, error) {
	return g.ExchangeTokenByForm(ctx, code, redirectURI)
}

// GetUserInfo Get GitHub user information | 获取GitHub用户信息
func (g *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	userInfoMap, err := g.GetUserInfoByJSON(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Parse GitHub user information | 解析GitHub用户信息
	userInfo := &UserInfo{
		ExtraData: userInfoMap,
	}

	// GitHub user ID | GitHub用户ID
	if id, ok := userInfoMap["id"].(float64); ok {
		userInfo.ProviderUserID = fmt.Sprintf("%.0f", id)
	}

	// Username | 用户名
	if login, ok := userInfoMap["login"].(string); ok {
		userInfo.Username = login
	}

	// Email | 邮箱
	if email, ok := userInfoMap["email"].(string); ok {
		userInfo.Email = email
	}

	// Avatar | 头像
	if avatar, ok := userInfoMap["avatar_url"].(string); ok {
		userInfo.Avatar = avatar
	}

	if userInfo.ProviderUserID == "" {
		return nil, fmt.Errorf("%w: provider_user_id is empty", ErrGetUserInfoFailed)
	}

	return userInfo, nil
}

// RefreshToken GitHub does not support refresh token | GitHub不支持刷新令牌
func (g *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return nil, fmt.Errorf("github does not support refresh token")
}

// ValidateToken Validate if access token is valid | 验证访问令牌是否有效
func (g *GitHubProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	return g.ValidateTokenByUserInfo(ctx, accessToken)
}
