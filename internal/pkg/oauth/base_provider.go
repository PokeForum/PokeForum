package oauth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// BaseProvider OAuth base provider | OAuth基础提供商
// Provide common HTTP request methods and Token handling logic | 提供通用的HTTP请求方法和Token处理逻辑
type BaseProvider struct {
	config     *Config       // OAuth配置
	httpClient *resty.Client // HTTP客户端
}

// NewBaseProvider Create base provider instance | 创建基础提供商实例
func NewBaseProvider(config *Config) *BaseProvider {
	return &BaseProvider{
		config: config,
		httpClient: resty.New().
			SetTimeout(30*time.Second).
			SetHeader("Accept", "application/json"),
	}
}

// GetConfig Get configuration | 获取配置
func (b *BaseProvider) GetConfig() *Config {
	return b.config
}

// BuildAuthURL Build authorization URL | 构建授权URL
// Common authorization URL building logic | 通用的授权URL构建逻辑
func (b *BaseProvider) BuildAuthURL(state string, redirectURL string, extraParams map[string]string) string {
	params := url.Values{}
	params.Set("client_id", b.config.ClientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("state", state)
	params.Set("response_type", "code")

	// Add scopes | 添加scopes
	if len(b.config.Scopes) > 0 {
		params.Set("scope", strings.Join(b.config.Scopes, " "))
	}

	// Add extra parameters | 添加额外参数
	for k, v := range extraParams {
		params.Set(k, v)
	}

	return fmt.Sprintf("%s?%s", b.config.AuthURL, params.Encode())
}

// doTokenExchange Execute common logic for Token exchange/refresh | 执行Token交换/刷新的通用逻辑
func (b *BaseProvider) doTokenExchange(ctx context.Context, data url.Values, errOnFailed error) (*TokenResponse, error) {
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	resp, err := b.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(data.Encode()).
		SetResult(&tokenResp).
		Post(b.config.TokenURL)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%w: status=%d, body=%s", errOnFailed, resp.StatusCode(), resp.String())
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: access_token is empty", errOnFailed)
	}

	return &TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		RefreshToken: tokenResp.RefreshToken,
		Scope:        tokenResp.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

// ExchangeTokenByForm Exchange Token via form | 通过表单方式交换Token
// Suitable for most OAuth providers | 适用于大多数OAuth提供商
func (b *BaseProvider) ExchangeTokenByForm(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", b.config.ClientID)
	data.Set("client_secret", b.config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	return b.doTokenExchange(ctx, data, ErrExchangeTokenFailed)
}

// RefreshTokenByForm Refresh Token via form | 通过表单方式刷新Token
func (b *BaseProvider) RefreshTokenByForm(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", b.config.ClientID)
	data.Set("client_secret", b.config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	return b.doTokenExchange(ctx, data, ErrRefreshTokenFailed)
}

// GetUserInfoByJSON Get user information via JSON | 通过JSON方式获取用户信息
// Send GET request with Bearer Token | 发送带Bearer Token的GET请求
func (b *BaseProvider) GetUserInfoByJSON(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	var userInfo map[string]interface{}

	resp, err := b.httpClient.R().
		SetContext(ctx).
		SetAuthToken(accessToken).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&userInfo).
		Get(b.config.UserInfoURL)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrGetUserInfoFailed, resp.StatusCode(), resp.String())
	}

	return userInfo, nil
}

// ValidateTokenByUserInfo Validate Token by getting user information | 通过获取用户信息来验证Token
// If user information can be obtained successfully, Token is valid | 如果能成功获取用户信息，说明Token有效
func (b *BaseProvider) ValidateTokenByUserInfo(ctx context.Context, accessToken string) (bool, error) {
	_, err := b.GetUserInfoByJSON(ctx, accessToken)
	if err != nil {
		return false, nil
	}
	return true, nil
}
