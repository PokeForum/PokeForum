package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// TelegramProvider Telegram OAuth提供商
// Telegram使用Telegram Login Widget，不是标准OAuth2.0
type TelegramProvider struct {
	*BaseProvider
}

// NewTelegramProvider 创建Telegram OAuth提供商实例
func NewTelegramProvider(config *Config) (IProvider, error) {
	// Telegram不使用标准OAuth2.0流程
	// 使用Telegram Login Widget
	return &TelegramProvider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL 获取Telegram授权URL
// Telegram使用Widget，不需要授权URL
func (t *TelegramProvider) GetAuthURL(state string) string {
	// Telegram使用Widget嵌入到页面中
	// 返回Widget的配置信息
	return fmt.Sprintf("https://telegram.org/auth?bot_id=%s&origin=%s&request_access=write",
		t.config.ClientID, t.config.RedirectURL)
}

// ExchangeToken Telegram不需要交换Token
// 用户信息直接通过回调参数返回
func (t *TelegramProvider) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	return nil, fmt.Errorf("telegram does not use token exchange")
}

// GetUserInfo 从Telegram回调参数中获取用户信息
// Telegram的用户信息通过回调URL参数返回
func (t *TelegramProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	return nil, fmt.Errorf("telegram user info should be obtained from callback parameters")
}

// ValidateTelegramAuth 验证Telegram回调数据
// data: 回调参数map，包含id、first_name、username、photo_url、auth_date、hash等
func (t *TelegramProvider) ValidateTelegramAuth(data map[string]string) (*UserInfo, error) {
	// 获取hash值
	hash, ok := data["hash"]
	if !ok {
		return nil, fmt.Errorf("%w: missing hash", ErrInvalidAuthCode)
	}

	// 检查auth_date（授权时间）
	authDateStr, ok := data["auth_date"]
	if !ok {
		return nil, fmt.Errorf("%w: missing auth_date", ErrInvalidAuthCode)
	}

	// 构建验证字符串
	var dataCheckArr []string
	for k, v := range data {
		if k != "hash" {
			dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// 计算secret_key = SHA256(bot_token)
	secretKey := sha256.Sum256([]byte(t.config.ClientSecret))

	// 计算HMAC-SHA256
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// 验证hash
	if calculatedHash != hash {
		return nil, fmt.Errorf("%w: invalid hash", ErrInvalidAuthCode)
	}

	// 验证auth_date（可选：检查是否在有效期内，例如24小时）
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid auth_date", ErrInvalidAuthCode)
	}
	_ = authDate // 可以添加时间验证逻辑

	// 构建用户信息
	userInfo := &UserInfo{
		ProviderUserID: data["id"],
		Username:       data["username"],
		Avatar:         data["photo_url"],
		ExtraData: map[string]interface{}{
			"first_name": data["first_name"],
			"last_name":  data["last_name"],
			"auth_date":  authDate,
		},
	}

	if userInfo.ProviderUserID == "" {
		return nil, fmt.Errorf("%w: missing user id", ErrGetUserInfoFailed)
	}

	return userInfo, nil
}

// RefreshToken Telegram不支持刷新令牌
func (t *TelegramProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return nil, fmt.Errorf("telegram does not support refresh token")
}

// ValidateToken Telegram不使用访问令牌
func (t *TelegramProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	return false, fmt.Errorf("telegram does not use access token")
}
