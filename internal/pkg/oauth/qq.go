package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// QQProvider QQ OAuth提供商
type QQProvider struct {
	*BaseProvider
}

// NewQQProvider 创建QQ OAuth提供商实例
func NewQQProvider(config *Config) (IProvider, error) {
	// 设置QQ默认配置
	if config.AuthURL == "" {
		config.AuthURL = "https://graph.qq.com/oauth2.0/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://graph.qq.com/oauth2.0/token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://graph.qq.com/user/get_user_info"
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"get_user_info"}
	}

	return &QQProvider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL 获取QQ授权URL
func (q *QQProvider) GetAuthURL(state string) string {
	return q.BuildAuthURL(state, nil)
}

// ExchangeToken 使用授权码换取访问令牌
// QQ返回的是URL编码格式，不是JSON
func (q *QQProvider) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", q.config.ClientID)
	data.Set("client_secret", q.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", q.config.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "GET", q.config.TokenURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeTokenFailed, err)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	// QQ返回格式: access_token=xxx&expires_in=7776000&refresh_token=xxx
	params, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	accessToken := params.Get("access_token")
	if accessToken == "" {
		return nil, fmt.Errorf("%w: access_token is empty, response: %s", ErrExchangeTokenFailed, string(body))
	}

	expiresIn := 7776000 // QQ默认90天
	if params.Get("expires_in") != "" {
		fmt.Sscanf(params.Get("expires_in"), "%d", &expiresIn)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: params.Get("refresh_token"),
	}, nil
}

// GetUserInfo 获取QQ用户信息
// QQ需要先获取OpenID，再获取用户信息
func (q *QQProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// 第一步：获取OpenID
	openID, err := q.getOpenID(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 第二步：获取用户信息
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("oauth_consumer_key", q.config.ClientID)
	params.Set("openid", openID)

	req, err := http.NewRequestWithContext(ctx, "GET", q.config.UserInfoURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetUserInfoFailed, err)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	var userInfoMap map[string]interface{}
	if err := json.Unmarshal(body, &userInfoMap); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	// 检查返回码
	if ret, ok := userInfoMap["ret"].(float64); ok && ret != 0 {
		return nil, fmt.Errorf("%w: ret=%v, msg=%v", ErrGetUserInfoFailed, ret, userInfoMap["msg"])
	}

	// 解析QQ用户信息
	userInfo := &UserInfo{
		ProviderUserID: openID,
		ExtraData:      userInfoMap,
	}

	// 昵称
	if nickname, ok := userInfoMap["nickname"].(string); ok {
		userInfo.Username = nickname
	}

	// 头像
	if figureurl, ok := userInfoMap["figureurl_qq_2"].(string); ok {
		userInfo.Avatar = figureurl
	} else if figureurl, ok := userInfoMap["figureurl_qq_1"].(string); ok {
		userInfo.Avatar = figureurl
	}

	return userInfo, nil
}

// getOpenID 获取QQ OpenID
func (q *QQProvider) getOpenID(ctx context.Context, accessToken string) (string, error) {
	openIDURL := "https://graph.qq.com/oauth2.0/me"
	params := url.Values{}
	params.Set("access_token", accessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", openIDURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGetUserInfoFailed, err)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	// QQ返回格式: callback( {"client_id":"YOUR_APPID","openid":"YOUR_OPENID"} );
	// 需要提取JSON部分
	re := regexp.MustCompile(`\{.*}`)
	jsonStr := re.FindString(string(body))
	if jsonStr == "" {
		return "", fmt.Errorf("%w: cannot extract json from response: %s", ErrParseResponse, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	openID, ok := result["openid"].(string)
	if !ok || openID == "" {
		return "", fmt.Errorf("%w: openid is empty", ErrGetUserInfoFailed)
	}

	return openID, nil
}

// RefreshToken 刷新访问令牌
func (q *QQProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", q.config.ClientID)
	data.Set("client_secret", q.config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "GET", q.config.TokenURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRefreshTokenFailed, err)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	params, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseResponse, err)
	}

	accessToken := params.Get("access_token")
	if accessToken == "" {
		return nil, fmt.Errorf("%w: access_token is empty", ErrRefreshTokenFailed)
	}

	expiresIn := 7776000
	if params.Get("expires_in") != "" {
		fmt.Sscanf(params.Get("expires_in"), "%d", &expiresIn)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: params.Get("refresh_token"),
	}, nil
}

// ValidateToken 验证访问令牌是否有效
func (q *QQProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	_, err := q.getOpenID(ctx, accessToken)
	if err != nil {
		if strings.Contains(err.Error(), "openid is empty") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
