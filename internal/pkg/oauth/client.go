package oauth

import (
	"sync"
)

// Client OAuth客户端实现
type Client struct {
	providers map[Provider]IProvider // 已注册的OAuth提供商
	mu        sync.RWMutex            // 读写锁，保护providers
}

// NewClient 创建OAuth客户端实例
func NewClient() IClient {
	return &Client{
		providers: make(map[Provider]IProvider),
	}
}

// GetProvider 获取指定的OAuth提供商
func (c *Client) GetProvider(provider Provider) (IProvider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	p, ok := c.providers[provider]
	if !ok {
		return nil, ErrProviderNotRegistered
	}

	return p, nil
}

// RegisterProvider 注册OAuth提供商
func (c *Client) RegisterProvider(provider Provider, config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已注册
	if _, ok := c.providers[provider]; ok {
		return ErrProviderAlreadyRegistered
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return err
	}

	// 根据提供商类型创建对应的实现
	var p IProvider
	var err error

	switch provider {
	case ProviderGitHub:
		p, err = NewGitHubProvider(config)
	case ProviderQQ:
		p, err = NewQQProvider(config)
	case ProviderGoogle:
		p, err = NewGoogleProvider(config)
	case ProviderApple:
		p, err = NewAppleProvider(config)
	case ProviderTelegram:
		p, err = NewTelegramProvider(config)
	case ProviderFIDO2:
		p, err = NewFIDO2Provider(config)
	default:
		return ErrProviderNotFound
	}

	if err != nil {
		return err
	}

	c.providers[provider] = p
	return nil
}

// UnregisterProvider 注销OAuth提供商
func (c *Client) UnregisterProvider(provider Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.providers[provider]; !ok {
		return ErrProviderNotRegistered
	}

	delete(c.providers, provider)
	return nil
}

// ListProviders 列出所有已注册的OAuth提供商
func (c *Client) ListProviders() []Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providers := make([]Provider, 0, len(c.providers))
	for p := range c.providers {
		providers = append(providers, p)
	}

	return providers
}

// validateConfig 验证OAuth配置
func validateConfig(config *Config) error {
	if config == nil {
		return ErrInvalidConfig
	}

	if config.ClientID == "" {
		return ErrMissingRequiredField
	}

	if config.ClientSecret == "" && config.Provider != ProviderFIDO2 {
		// FIDO2不需要ClientSecret
		return ErrMissingRequiredField
	}

	if config.AuthURL == "" && config.Provider != ProviderFIDO2 {
		return ErrMissingRequiredField
	}

	if config.TokenURL == "" && config.Provider != ProviderFIDO2 {
		return ErrMissingRequiredField
	}

	if config.RedirectURL == "" {
		return ErrMissingRequiredField
	}

	return nil
}
