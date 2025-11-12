package oauth

import "errors"

var (
	// ErrProviderNotFound OAuth提供商未找到
	ErrProviderNotFound = errors.New("oauth provider not found")

	// ErrProviderNotRegistered OAuth提供商未注册
	ErrProviderNotRegistered = errors.New("oauth provider not registered")

	// ErrProviderAlreadyRegistered OAuth提供商已注册
	ErrProviderAlreadyRegistered = errors.New("oauth provider already registered")

	// ErrInvalidConfig 无效的OAuth配置
	ErrInvalidConfig = errors.New("invalid oauth config")

	// ErrInvalidAuthCode 无效的授权码
	ErrInvalidAuthCode = errors.New("invalid authorization code")

	// ErrInvalidAccessToken 无效的访问令牌
	ErrInvalidAccessToken = errors.New("invalid access token")

	// ErrInvalidRefreshToken 无效的刷新令牌
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = errors.New("token expired")

	// ErrExchangeTokenFailed 交换令牌失败
	ErrExchangeTokenFailed = errors.New("exchange token failed")

	// ErrGetUserInfoFailed 获取用户信息失败
	ErrGetUserInfoFailed = errors.New("get user info failed")

	// ErrRefreshTokenFailed 刷新令牌失败
	ErrRefreshTokenFailed = errors.New("refresh token failed")

	// ErrValidateTokenFailed 验证令牌失败
	ErrValidateTokenFailed = errors.New("validate token failed")

	// ErrNetworkRequest 网络请求失败
	ErrNetworkRequest = errors.New("network request failed")

	// ErrParseResponse 解析响应失败
	ErrParseResponse = errors.New("parse response failed")

	// ErrMissingRequiredField 缺少必需字段
	ErrMissingRequiredField = errors.New("missing required field")
)
