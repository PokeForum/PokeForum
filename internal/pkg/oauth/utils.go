package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateState Generate random state string | 生成随机state字符串
// Used to prevent CSRF attacks | 用于防止CSRF攻击
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateState Validate state string | 验证state字符串
// expectedState: state stored in session | expectedState: 存储在session中的state
// actualState: state obtained from callback parameters | actualState: 从回调参数中获取的state
func ValidateState(expectedState, actualState string) bool {
	return expectedState != "" && expectedState == actualState
}

// IsProviderSupported Check if provider is supported | 检查提供商是否支持
func IsProviderSupported(provider Provider) bool {
	switch provider {
	case ProviderQQ, ProviderGitHub, ProviderApple, ProviderGoogle, ProviderTelegram, ProviderFIDO2:
		return true
	default:
		return false
	}
}

// GetProviderName Get display name of provider | 获取提供商的显示名称
func GetProviderName(provider Provider) string {
	switch provider {
	case ProviderQQ:
		return "QQ"
	case ProviderGitHub:
		return "GitHub"
	case ProviderApple:
		return "Apple"
	case ProviderGoogle:
		return "Google"
	case ProviderTelegram:
		return "Telegram"
	case ProviderFIDO2:
		return "FIDO2"
	default:
		return string(provider)
	}
}

// GetAllProviders Get all supported providers | 获取所有支持的提供商
func GetAllProviders() []Provider {
	return []Provider{
		ProviderQQ,
		ProviderGitHub,
		ProviderApple,
		ProviderGoogle,
		ProviderTelegram,
		ProviderFIDO2,
	}
}
