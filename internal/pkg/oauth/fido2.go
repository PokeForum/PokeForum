package oauth

import (
	"context"
	"fmt"
)

// FIDO2Provider FIDO2 provider | FIDO2提供商
// FIDO2/WebAuthn is a passwordless authentication standard, not OAuth2.0 | FIDO2/WebAuthn是一种无密码认证标准，不是OAuth2.0
// Provide a placeholder implementation here, actual use requires integrating WebAuthn library | 这里提供一个占位实现，实际使用需要集成WebAuthn库
type FIDO2Provider struct {
	*BaseProvider
}

// NewFIDO2Provider Create FIDO2 provider instance | 创建FIDO2提供商实例
func NewFIDO2Provider(config *Config) (IProvider, error) {
	// FIDO2 does not use standard OAuth2.0 flow | FIDO2不使用标准OAuth2.0流程
	// Need to integrate WebAuthn library to implement | 需要集成WebAuthn库来实现
	return &FIDO2Provider{
		BaseProvider: NewBaseProvider(config),
	}, nil
}

// GetAuthURL FIDO2 does not use authorization URL | FIDO2不使用授权URL
func (f *FIDO2Provider) GetAuthURL(state string, redirectURL string) string {
	return ""
}

// ExchangeToken FIDO2 does not use Token exchange | FIDO2不使用Token交换
func (f *FIDO2Provider) ExchangeToken(ctx context.Context, code string, _ string) (*TokenResponse, error) {
	return nil, fmt.Errorf("fido2 does not use token exchange")
}

// GetUserInfo FIDO2 user information needs to be obtained from WebAuthn authentication result | FIDO2的用户信息需要从WebAuthn认证结果中获取
func (f *FIDO2Provider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	return nil, fmt.Errorf("fido2 user info should be obtained from webauthn assertion")
}

// RefreshToken FIDO2 does not support refresh token | FIDO2不支持刷新令牌
func (f *FIDO2Provider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return nil, fmt.Errorf("fido2 does not support refresh token")
}

// ValidateToken FIDO2 does not use access token | FIDO2不使用访问令牌
func (f *FIDO2Provider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	return false, fmt.Errorf("fido2 does not use access token")
}

// Note: Complete FIDO2 implementation requires: | 注意：完整的FIDO2实现需要：
// 1. Registration phase: generate challenge, verify attestation | 1. 注册阶段：生成challenge，验证attestation
// 2. Authentication phase: generate challenge, verify assertion | 2. 认证阶段：生成challenge，验证assertion
// 3. Store public key credentials | 3. 存储公钥凭证
// Recommend using mature WebAuthn libraries, such as: | 建议使用成熟的WebAuthn库，如：
// - github.com/go-webauthn/webauthn
// - github.com/duo-labs/webauthn
