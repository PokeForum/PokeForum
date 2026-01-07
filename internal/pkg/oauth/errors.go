package oauth

import "errors"

var (
	// ErrProviderNotFound OAuth provider not found | OAuth提供商未找到
	ErrProviderNotFound = errors.New("oauth provider not found")

	// ErrProviderNotRegistered OAuth provider not registered | OAuth提供商未注册
	ErrProviderNotRegistered = errors.New("oauth provider not registered")

	// ErrProviderAlreadyRegistered OAuth provider already registered | OAuth提供商已注册
	ErrProviderAlreadyRegistered = errors.New("oauth provider already registered")

	// ErrInvalidConfig Invalid OAuth configuration | 无效的OAuth配置
	ErrInvalidConfig = errors.New("invalid oauth config")

	// ErrInvalidAuthCode Invalid authorization code | 无效的授权码
	ErrInvalidAuthCode = errors.New("invalid authorization code")

	// ErrInvalidAccessToken Invalid access token | 无效的访问令牌
	ErrInvalidAccessToken = errors.New("invalid access token")

	// ErrInvalidRefreshToken Invalid refresh token | 无效的刷新令牌
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrTokenExpired Token expired | 令牌已过期
	ErrTokenExpired = errors.New("token expired")

	// ErrExchangeTokenFailed Exchange token failed | 交换令牌失败
	ErrExchangeTokenFailed = errors.New("exchange token failed")

	// ErrGetUserInfoFailed Get user info failed | 获取用户信息失败
	ErrGetUserInfoFailed = errors.New("get user info failed")

	// ErrRefreshTokenFailed Refresh token failed | 刷新令牌失败
	ErrRefreshTokenFailed = errors.New("refresh token failed")

	// ErrValidateTokenFailed Validate token failed | 验证令牌失败
	ErrValidateTokenFailed = errors.New("validate token failed")

	// ErrNetworkRequest Network request failed | 网络请求失败
	ErrNetworkRequest = errors.New("network request failed")

	// ErrParseResponse Parse response failed | 解析响应失败
	ErrParseResponse = errors.New("parse response failed")

	// ErrMissingRequiredField Missing required field | 缺少必需字段
	ErrMissingRequiredField = errors.New("missing required field")
)
