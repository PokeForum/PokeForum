package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaseProvider OAuth基础提供商
// 提供通用的HTTP请求方法和Token处理逻辑
type BaseProvider struct {
	config     *Config      // OAuth配置
	httpClient *http.Client // HTTP客户端
}

// NewBaseProvider 创建基础提供商实例
func NewBaseProvider(config *Config) *BaseProvider {
	return &BaseProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetConfig 获取配置
func (b *BaseProvider) GetConfig() *Config {
	return b.config
}

// BuildAuthURL 构建授权URL
// 通用的授权URL构建逻辑
func (b *BaseProvider) BuildAuthURL(state string, extraParams map[string]string) string {
	params := url.Values{}
	params.Set("client_id", b.config.ClientID)
	params.Set("redirect_uri", b.config.RedirectURL)
	params.Set("state", state)
	params.Set("response_type", "code")

	// 添加scopes
	if len(b.config.Scopes) > 0 {
		params.Set("scope", strings.Join(b.config.Scopes, " "))
	}

	// 添加额外参数
	for k, v := range extraParams {
		params.Set(k, v)
	}

	return fmt.Sprintf("%s?%s", b.config.AuthURL, params.Encode())
}

// ExchangeTokenByForm 通过表单方式交换Token
// 适用于大多数OAuth提供商
func (b *BaseProvider) ExchangeTokenByForm(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", b.config.ClientID)
	data.Set("client_secret", b.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", b.config.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", b.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeTokenFailed, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrExchangeTokenFailed, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: access_token is empty", ErrExchangeTokenFailed)
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

// GetUserInfoByJSON 通过JSON方式获取用户信息
// 发送带Bearer Token的GET请求
func (b *BaseProvider) GetUserInfoByJSON(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.config.UserInfoURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetUserInfoFailed, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrGetUserInfoFailed, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	return userInfo, nil
}

// RefreshTokenByForm 通过表单方式刷新Token
func (b *BaseProvider) RefreshTokenByForm(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", b.config.ClientID)
	data.Set("client_secret", b.config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", b.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRefreshTokenFailed, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrRefreshTokenFailed, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: access_token is empty", ErrRefreshTokenFailed)
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

// ValidateTokenByUserInfo 通过获取用户信息来验证Token
// 如果能成功获取用户信息，说明Token有效
func (b *BaseProvider) ValidateTokenByUserInfo(ctx context.Context, accessToken string) (bool, error) {
	_, err := b.GetUserInfoByJSON(ctx, accessToken)
	if err != nil {
		return false, nil
	}
	return true, nil
}
