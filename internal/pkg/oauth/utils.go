package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
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
	case ProviderQQ, ProviderGitHub, ProviderGoogle, ProviderFIDO2:
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
	case ProviderGoogle:
		return "Google"
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
		ProviderGoogle,
		ProviderFIDO2,
	}
}

// NormalizeProvider Normalize provider string to correct enum value | 将 provider 字符串标准化为正确的枚举值
// Supports case-insensitive matching | 支持大小写不敏感匹配
func NormalizeProvider(provider string) Provider {
	switch strings.ToLower(provider) {
	case "qq":
		return ProviderQQ
	case "github":
		return ProviderGitHub
	case "google":
		return ProviderGoogle
	case "fido2":
		return ProviderFIDO2
	default:
		return Provider(provider)
	}
}
