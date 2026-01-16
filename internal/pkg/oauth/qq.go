package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// QQProvider QQ OAuth provider | QQ OAuth提供商
type QQProvider struct {
	*BaseProvider
}

const (
	// unknownParseErrMsg unknown parse error message | 未知解析错误消息
	unknownParseErrMsg = "unknown error"
)

// NewQQProvider Create QQ OAuth provider instance | 创建QQ OAuth提供商实例
func NewQQProvider(config *Config) (IProvider, error) {
	// Set QQ default configuration | 设置QQ默认配置
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

// GetAuthURL Get QQ authorization URL | 获取QQ授权URL
func (q *QQProvider) GetAuthURL(state string, redirectURL string) string {
	return q.BuildAuthURL(state, redirectURL, nil)
}

// ExchangeToken Exchange authorization code for access token | 使用授权码换取访问令牌
// QQ returns URL encoded format by default, but can also return JSON format | QQ默认返回URL编码格式，但也可能返回JSON格式
func (q *QQProvider) ExchangeToken(ctx context.Context, code string, redirectURI string) (*TokenResponse, error) {
	resp, err := q.httpClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"client_id":     q.config.ClientID,
			"client_secret": q.config.ClientSecret,
			"code":          code,
			"grant_type":    "authorization_code",
			"redirect_uri":  redirectURI,
			"fmt":           "json", // Specify JSON format return | 指定返回JSON格式
		}).
		Get(q.config.TokenURL)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrExchangeTokenFailed, resp.StatusCode(), resp.String())
	}

	return q.parseTokenResponse(resp.Body(), ErrExchangeTokenFailed)
}

// GetUserInfo Get QQ user information | 获取QQ用户信息
// QQ needs to get OpenID first, then get user information | QQ需要先获取OpenID，再获取用户信息
func (q *QQProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// Step 1: Get OpenID | 第一步：获取OpenID
	openID, err := q.getOpenID(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Step 2: Get user information | 第二步：获取用户信息
	var userInfoMap map[string]interface{}

	resp, err := q.httpClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"access_token":       accessToken,
			"oauth_consumer_key": q.config.ClientID,
			"openid":             openID,
		}).
		SetResult(&userInfoMap).
		Get(q.config.UserInfoURL)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrGetUserInfoFailed, resp.StatusCode(), resp.String())
	}

	// Check return code | 检查返回码
	if ret, ok := userInfoMap["ret"].(float64); ok && ret != 0 {
		return nil, fmt.Errorf("%w: ret=%v, msg=%v", ErrGetUserInfoFailed, ret, userInfoMap["msg"])
	}

	// Parse QQ user information | 解析QQ用户信息
	userInfo := &UserInfo{
		ProviderUserID: openID,
		ExtraData:      userInfoMap,
	}

	// Nickname | 昵称
	if nickname, ok := userInfoMap["nickname"].(string); ok {
		userInfo.Username = nickname
	}

	// Avatar | 头像
	if figureurl, ok := userInfoMap["figureurl_qq_2"].(string); ok {
		userInfo.Avatar = figureurl
	} else if figureurl, ok := userInfoMap["figureurl_qq_1"].(string); ok {
		userInfo.Avatar = figureurl
	}

	return userInfo, nil
}

// getOpenID Get QQ OpenID | 获取QQ OpenID
func (q *QQProvider) getOpenID(ctx context.Context, accessToken string) (string, error) {
	openIDURL := "https://graph.qq.com/oauth2.0/me"

	resp, err := q.httpClient.R().
		SetContext(ctx).
		SetQueryParam("access_token", accessToken).
		Get(openIDURL)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("%w: status=%d, body=%s", ErrGetUserInfoFailed, resp.StatusCode(), resp.String())
	}

	body := resp.String()

	// QQ return format: callback( {"client_id":"YOUR_APPID","openid":"YOUR_OPENID"} )
	// Need to extract JSON part | 需要提取JSON部分
	re := regexp.MustCompile(`\{.*}`)
	jsonStr := re.FindString(body)
	if jsonStr == "" {
		return "", fmt.Errorf("%w: cannot extract json from response: %s", ErrParseResponse, body)
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

// RefreshToken Refresh access token | 刷新访问令牌
func (q *QQProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	resp, err := q.httpClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"client_id":     q.config.ClientID,
			"client_secret": q.config.ClientSecret,
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
			"fmt":           "json", // Specify JSON format return | 指定返回JSON格式
		}).
		Get(q.config.TokenURL)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkRequest, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrRefreshTokenFailed, resp.StatusCode(), resp.String())
	}

	return q.parseTokenResponse(resp.Body(), ErrRefreshTokenFailed)
}

// parseTokenResponse Parse token response from QQ API | 解析QQ API的token响应
// QQ returns JSON format or URL encoded format | QQ返回JSON格式或URL编码格式
func (q *QQProvider) parseTokenResponse(bodyBytes []byte, errToWrap error) (*TokenResponse, error) {
	// Try JSON format first | 先尝试JSON格式
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    string `json:"expires_in"` // QQ returns string type | QQ返回字符串类型
		RefreshToken string `json:"refresh_token"`
	}

	jsonErr := json.Unmarshal(bodyBytes, &tokenResp)
	if jsonErr == nil && tokenResp.AccessToken != "" {
		// JSON parsing successful | JSON解析成功
		expiresIn := 7776000 // QQ default 90 days | QQ默认90天
		if tokenResp.ExpiresIn != "" {
			if _, err := fmt.Sscanf(tokenResp.ExpiresIn, "%d", &expiresIn); err != nil || expiresIn == 0 {
				expiresIn = 7776000 // Keep default value if parsing fails | 解析失败保持默认值
			}
		}

		return &TokenResponse{
			AccessToken:  tokenResp.AccessToken,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
			RefreshToken: tokenResp.RefreshToken,
		}, nil
	}

	// JSON parsing failed or AccessToken is empty, try URL encoded format
	// JSON解析失败或AccessToken为空,尝试URL编码格式
	bodyString := string(bodyBytes)
	values, err := url.ParseQuery(bodyString)
	if err != nil {
		parseErrMsg := unknownParseErrMsg
		if jsonErr != nil {
			parseErrMsg = fmt.Sprintf("json parse error: %v", jsonErr)
		}
		return nil, fmt.Errorf("%w: url parse failed (%v), %s, body: %s", errToWrap, err, parseErrMsg, bodyString)
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		parseErrMsg := unknownParseErrMsg
		if jsonErr != nil {
			parseErrMsg = fmt.Sprintf("json parse error: %v", jsonErr)
		}
		return nil, fmt.Errorf("%w: access_token is empty in both json and url encoded response, %s, body: %s", errToWrap, parseErrMsg, bodyString)
	}

	expiresIn := 7776000 // QQ default 90 days | QQ默认90天
	if expiresInStr := values.Get("expires_in"); expiresInStr != "" {
		if _, err := fmt.Sscanf(expiresInStr, "%d", &expiresIn); err != nil || expiresIn == 0 {
			expiresIn = 7776000 // Keep default value if parsing fails | 解析失败保持默认值
		}
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: values.Get("refresh_token"),
	}, nil
}

// ValidateToken Validate if access token is valid | 验证访问令牌是否有效
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
