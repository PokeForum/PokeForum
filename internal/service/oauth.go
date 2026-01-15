package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/oauth"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// OAuthState OAuth state parameter structure | OAuth状态参数结构
type OAuthState struct {
	Nonce     string `json:"nonce"`      // Random string to prevent replay | 随机字符串，防重放
	Provider  string `json:"provider"`   // Provider type | 提供商类型
	Action    string `json:"action"`     // Action type: login/bind | 操作类型: login/bind
	UserID    int    `json:"user_id"`    // User ID for bind scenario | 绑定场景下的用户ID
	ClientIP  string `json:"client_ip"`  // Client IP | 客户端IP
	ExpiresAt int64  `json:"expires_at"` // Expiration timestamp | 过期时间戳
}

const (
	// OAuthStateKeyPrefix OAuth state Redis key prefix | OAuth state Redis键前缀
	OAuthStateKeyPrefix = "oauth:state:"
	// OAuthStateTTL OAuth state TTL (10 minutes) | OAuth state有效期（10分钟）
	OAuthStateTTL = 600

	// OAuthActionLogin Login action | 登录操作
	OAuthActionLogin = "login"
	// OAuthActionBind Bind action | 绑定操作
	OAuthActionBind = "bind"

	// OAuthResultLogin Login result | 登录结果
	OAuthResultLogin = "login"
	// OAuthResultRegister Register result | 注册结果
	OAuthResultRegister = "register"
	// OAuthResultBindRequired Bind required result | 需要绑定结果
	OAuthResultBindRequired = "bindRequired"
	// OAuthResultBind Bind result | 绑定结果
	OAuthResultBind = "bind"
)

// IOAuthService OAuth login service interface | OAuth登录服务接口
type IOAuthService interface {
	// GetEnabledProviders Get enabled providers list | 获取已启用的提供商列表
	GetEnabledProviders(ctx context.Context) (*schema.OAuthProviderPublicListResponse, error)
	// GetAuthorizeURL Get authorization URL | 获取授权URL
	GetAuthorizeURL(ctx context.Context, provider string, req schema.OAuthAuthorizeRequest, clientIP string) (*schema.OAuthAuthorizeResponse, error)
	// HandleCallback Handle OAuth callback (login/register) | 处理OAuth回调(登录/注册)
	HandleCallback(ctx context.Context, provider string, req schema.OAuthCallbackRequest, clientIP, ua string) (*schema.OAuthCallbackResponse, error)
	// GetUserBindList Get user binding list | 获取用户绑定列表
	GetUserBindList(ctx context.Context, userID int) (*schema.OAuthUserBindListResponse, error)
	// GetBindURL Get bind authorization URL | 获取绑定授权URL
	GetBindURL(ctx context.Context, userID int, provider string, req schema.OAuthAuthorizeRequest, clientIP string) (*schema.OAuthAuthorizeResponse, error)
	// HandleBindCallback Handle bind callback | 处理绑定回调
	HandleBindCallback(ctx context.Context, userID int, provider string, req schema.OAuthBindCallbackRequest, clientIP string) error
	// Unbind Unbind OAuth | 解绑OAuth
	Unbind(ctx context.Context, userID int, provider string) error
}

// OAuthService OAuth login service implementation | OAuth登录服务实现
type OAuthService struct {
	db                *ent.Client
	userRepo          repository.IUserRepository
	userOAuthRepo     repository.IUserOAuthRepository
	oauthProviderRepo repository.IOAuthProviderRepository
	cache             cache.ICacheService
	logger            *zap.Logger
	settings          ISettingsService
	oauthClient       oauth.IClient
}

// NewOAuthService Create OAuth login service instance | 创建OAuth登录服务实例
func NewOAuthService(
	db *ent.Client,
	userRepo repository.IUserRepository,
	userOAuthRepo repository.IUserOAuthRepository,
	oauthProviderRepo repository.IOAuthProviderRepository,
	cacheService cache.ICacheService,
	logger *zap.Logger,
	settings ISettingsService,
) IOAuthService {
	return &OAuthService{
		db:                db,
		userRepo:          userRepo,
		userOAuthRepo:     userOAuthRepo,
		oauthProviderRepo: oauthProviderRepo,
		cache:             cacheService,
		logger:            logger,
		settings:          settings,
		oauthClient:       oauth.NewClient(),
	}
}

// GetEnabledProviders Get enabled providers list | 获取已启用的提供商列表
func (s *OAuthService) GetEnabledProviders(ctx context.Context) (*schema.OAuthProviderPublicListResponse, error) {
	enabled := true
	providers, err := s.oauthProviderRepo.List(ctx, "", &enabled)
	if err != nil {
		s.logger.Error("Failed to get enabled OAuth providers | 获取已启用OAuth提供商失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}

	list := make([]schema.OAuthProviderPublicItem, 0, len(providers))
	for _, p := range providers {
		list = append(list, schema.OAuthProviderPublicItem{
			Provider:  p.Provider.String(),
			SortOrder: p.SortOrder,
		})
	}

	return &schema.OAuthProviderPublicListResponse{List: list}, nil
}

// GetAuthorizeURL Get authorization URL | 获取授权URL
func (s *OAuthService) GetAuthorizeURL(ctx context.Context, provider string, req schema.OAuthAuthorizeRequest, clientIP string) (*schema.OAuthAuthorizeResponse, error) {
	// Get provider config | 获取提供商配置
	oauthProvider, err := s.getAndRegisterProvider(ctx, provider)
	if err != nil {
		return nil, err
	}

	// Generate state | 生成state
	state, err := s.generateState(ctx, provider, OAuthActionLogin, 0, clientIP)
	if err != nil {
		s.logger.Error("Failed to generate OAuth state | 生成OAuth state失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("生成授权参数失败")
	}

	// Get authorization URL | 获取授权URL
	p, err := s.oauthClient.GetProvider(oauth.Provider(provider))
	if err != nil {
		s.logger.Error("Failed to get OAuth provider | 获取OAuth提供商失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("provider", provider))
		return nil, errors.New("OAuth提供商未配置")
	}

	authURL := p.GetAuthURL(state, req.RedirectURI)

	s.logger.Info("Generated OAuth authorize URL | 生成OAuth授权URL",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", oauthProvider.Provider.String()),
		zap.String("client_ip", clientIP))

	return &schema.OAuthAuthorizeResponse{
		AuthorizeURL: authURL,
		State:        state,
	}, nil
}

// HandleCallback Handle OAuth callback (login/register) | 处理OAuth回调(登录/注册)
func (s *OAuthService) HandleCallback(ctx context.Context, provider string, req schema.OAuthCallbackRequest, clientIP, ua string) (*schema.OAuthCallbackResponse, error) {
	// Validate state | 验证state
	stateData, err := s.validateAndDeleteState(ctx, req.State, provider, OAuthActionLogin)
	if err != nil {
		return nil, err
	}

	// Get provider | 获取提供商
	p, err := s.oauthClient.GetProvider(oauth.Provider(provider))
	if err != nil {
		// Try to register provider | 尝试注册提供商
		if _, regErr := s.getAndRegisterProvider(ctx, provider); regErr != nil {
			return nil, errors.New("OAuth提供商未配置")
		}
		p, err = s.oauthClient.GetProvider(oauth.Provider(provider))
		if err != nil {
			return nil, errors.New("OAuth提供商未配置")
		}
	}

	// Exchange token | 换取token
	tokenResp, err := p.ExchangeToken(ctx, req.Code)
	if err != nil {
		s.logger.Error("Failed to exchange OAuth token | 换取OAuth token失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("provider", provider))
		return nil, errors.New("授权码换取失败")
	}

	// Get user info | 获取用户信息
	userInfo, err := p.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to get OAuth user info | 获取OAuth用户信息失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("provider", provider))
		return nil, errors.New("获取用户信息失败")
	}

	// Check if already bound | 检查是否已绑定
	binding, err := s.userOAuthRepo.GetByProviderUserID(ctx, provider, userInfo.ProviderUserID)
	if err != nil {
		s.logger.Error("Failed to query OAuth binding | 查询OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("查询绑定信息失败")
	}

	if binding != nil {
		// Already bound, login directly | 已绑定，直接登录
		return s.loginExistingUser(ctx, binding.UserID, provider, userInfo, ua, stateData.ClientIP)
	}

	// Not bound, check if can auto register | 未绑定，检查是否可以自动注册
	canRegister, msg := s.canAutoRegister(ctx)
	if !canRegister {
		s.logger.Info("OAuth auto register not allowed | OAuth自动注册不允许",
			tracing.WithTraceIDField(ctx),
			zap.String("provider", provider),
			zap.String("reason", msg))
		return &schema.OAuthCallbackResponse{
			Action:  OAuthResultBindRequired,
			Message: msg,
		}, nil
	}

	// Auto register new user | 自动注册新用户
	return s.registerNewUser(ctx, provider, userInfo, ua, clientIP)
}

// GetUserBindList Get user binding list | 获取用户绑定列表
func (s *OAuthService) GetUserBindList(ctx context.Context, userID int) (*schema.OAuthUserBindListResponse, error) {
	bindings, err := s.userOAuthRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user OAuth bindings | 获取用户OAuth绑定列表失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.Int("user_id", userID))
		return nil, err
	}

	list := make([]schema.OAuthUserBindItem, 0, len(bindings))
	for _, b := range bindings {
		list = append(list, schema.OAuthUserBindItem{
			Provider:         b.Provider.String(),
			ProviderUsername: b.ProviderUsername,
			ProviderAvatar:   b.ProviderAvatar,
			BoundAt:          b.CreatedAt.Format(time_tools.DateTimeFormat),
		})
	}

	return &schema.OAuthUserBindListResponse{List: list}, nil
}

// GetBindURL Get bind authorization URL | 获取绑定授权URL
func (s *OAuthService) GetBindURL(ctx context.Context, userID int, provider string, req schema.OAuthAuthorizeRequest, clientIP string) (*schema.OAuthAuthorizeResponse, error) {
	// Check if already bound | 检查是否已绑定
	exists, err := s.userOAuthRepo.ExistsByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		s.logger.Error("Failed to check OAuth binding | 检查OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("检查绑定状态失败")
	}
	if exists {
		return nil, errors.New("您已绑定该OAuth提供商")
	}

	// Get provider config | 获取提供商配置
	oauthProvider, err := s.getAndRegisterProvider(ctx, provider)
	if err != nil {
		return nil, err
	}

	// Generate state | 生成state
	state, err := s.generateState(ctx, provider, OAuthActionBind, userID, clientIP)
	if err != nil {
		s.logger.Error("Failed to generate OAuth state | 生成OAuth state失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("生成授权参数失败")
	}

	// Get authorization URL | 获取授权URL
	p, err := s.oauthClient.GetProvider(oauth.Provider(provider))
	if err != nil {
		return nil, errors.New("OAuth提供商未配置")
	}

	authURL := p.GetAuthURL(state, req.RedirectURI)

	s.logger.Info("Generated OAuth bind URL | 生成OAuth绑定URL",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", oauthProvider.Provider.String()),
		zap.Int("user_id", userID))

	return &schema.OAuthAuthorizeResponse{
		AuthorizeURL: authURL,
		State:        state,
	}, nil
}

// HandleBindCallback Handle bind callback | 处理绑定回调
func (s *OAuthService) HandleBindCallback(ctx context.Context, userID int, provider string, req schema.OAuthBindCallbackRequest, clientIP string) error {
	// Validate state | 验证state
	stateData, err := s.validateAndDeleteState(ctx, req.State, provider, OAuthActionBind)
	if err != nil {
		return err
	}

	// Verify user ID matches | 验证用户ID匹配
	if stateData.UserID != userID {
		s.logger.Warn("OAuth bind user ID mismatch | OAuth绑定用户ID不匹配",
			tracing.WithTraceIDField(ctx),
			zap.Int("state_user_id", stateData.UserID),
			zap.Int("request_user_id", userID))
		return errors.New("授权参数无效")
	}

	// Get provider | 获取提供商
	p, err := s.oauthClient.GetProvider(oauth.Provider(provider))
	if err != nil {
		if _, regErr := s.getAndRegisterProvider(ctx, provider); regErr != nil {
			return errors.New("OAuth提供商未配置")
		}
		p, err = s.oauthClient.GetProvider(oauth.Provider(provider))
		if err != nil {
			return errors.New("OAuth提供商未配置")
		}
	}

	// Exchange token | 换取token
	tokenResp, err := p.ExchangeToken(ctx, req.Code)
	if err != nil {
		s.logger.Error("Failed to exchange OAuth token | 换取OAuth token失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("授权码换取失败")
	}

	// Get user info | 获取用户信息
	userInfo, err := p.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to get OAuth user info | 获取OAuth用户信息失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("获取用户信息失败")
	}

	// Check if provider user ID already bound to another user | 检查该OAuth账号是否已被其他用户绑定
	existingBinding, err := s.userOAuthRepo.GetByProviderUserID(ctx, provider, userInfo.ProviderUserID)
	if err != nil {
		s.logger.Error("Failed to query OAuth binding | 查询OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("查询绑定信息失败")
	}
	if existingBinding != nil {
		return errors.New("该OAuth账号已被其他用户绑定")
	}

	// Check if user already bound this provider | 检查用户是否已绑定该提供商
	exists, err := s.userOAuthRepo.ExistsByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		return errors.New("检查绑定状态失败")
	}
	if exists {
		return errors.New("您已绑定该OAuth提供商")
	}

	// Create binding | 创建绑定
	_, err = s.userOAuthRepo.Create(ctx, userID, provider, userInfo.ProviderUserID,
		userInfo.Username, userInfo.Email, userInfo.Avatar, userInfo.ExtraData)
	if err != nil {
		s.logger.Error("Failed to create OAuth binding | 创建OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("绑定失败")
	}

	s.logger.Info("OAuth bind success | OAuth绑定成功",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", provider),
		zap.Int("user_id", userID),
		zap.String("provider_user_id", userInfo.ProviderUserID))

	return nil
}

// Unbind Unbind OAuth | 解绑OAuth
func (s *OAuthService) Unbind(ctx context.Context, userID int, provider string) error {
	// Get binding | 获取绑定
	binding, err := s.userOAuthRepo.GetByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		s.logger.Error("Failed to get OAuth binding | 获取OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("查询绑定信息失败")
	}
	if binding == nil {
		return errors.New("未绑定该OAuth提供商")
	}

	// Check if this is the last login method | 检查是否是最后一种登录方式
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("查询用户信息失败")
	}

	// Count OAuth bindings | 统计OAuth绑定数量
	bindCount, err := s.userOAuthRepo.CountByUserID(ctx, userID)
	if err != nil {
		return errors.New("查询绑定数量失败")
	}

	// If user has no password and only one OAuth binding, cannot unbind | 如果用户没有密码且只有一个OAuth绑定，不能解绑
	if u.Password == "" && bindCount <= 1 {
		return errors.New("不能解绑最后一种登录方式，请先设置密码")
	}

	// Delete binding | 删除绑定
	if err := s.userOAuthRepo.Delete(ctx, binding.ID); err != nil {
		s.logger.Error("Failed to delete OAuth binding | 删除OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return errors.New("解绑失败")
	}

	s.logger.Info("OAuth unbind success | OAuth解绑成功",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", provider),
		zap.Int("user_id", userID))

	return nil
}

// getAndRegisterProvider Get and register OAuth provider | 获取并注册OAuth提供商
func (s *OAuthService) getAndRegisterProvider(ctx context.Context, provider string) (*ent.OAuthProvider, error) {
	// Check if provider is enabled | 检查提供商是否启用
	enabled := true
	providers, err := s.oauthProviderRepo.List(ctx, provider, &enabled)
	if err != nil {
		s.logger.Error("Failed to query OAuth provider | 查询OAuth提供商失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("查询OAuth提供商失败")
	}
	if len(providers) == 0 {
		return nil, errors.New("该OAuth提供商未启用")
	}

	oauthProvider := providers[0]

	// Register provider to client | 注册提供商到客户端
	config := &oauth.Config{
		Provider:     oauth.Provider(provider),
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		AuthURL:      oauthProvider.AuthURL,
		TokenURL:     oauthProvider.TokenURL,
		UserInfoURL:  oauthProvider.UserInfoURL,
		Scopes:       oauthProvider.Scopes,
		ExtraConfig:  oauthProvider.ExtraConfig,
	}

	// Register provider | 注册提供商
	if err := s.oauthClient.RegisterProvider(oauth.Provider(provider), config); err != nil {
		if err == oauth.ErrProviderAlreadyRegistered {
			s.logger.Debug("OAuth provider already registered | OAuth提供商已注册",
				tracing.WithTraceIDField(ctx), zap.String("provider", provider))
		} else {
			s.logger.Error("Failed to register OAuth provider | 注册OAuth提供商失败",
				tracing.WithTraceIDField(ctx), zap.Error(err), zap.String("provider", provider))
			return nil, fmt.Errorf("注册OAuth提供商失败: %w", err)
		}
	}

	return oauthProvider, nil
}

// generateState Generate OAuth state | 生成OAuth state
func (s *OAuthService) generateState(ctx context.Context, provider, action string, userID int, clientIP string) (string, error) {
	// Generate nonce | 生成nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Create state data | 创建state数据
	stateData := OAuthState{
		Nonce:     nonce,
		Provider:  provider,
		Action:    action,
		UserID:    userID,
		ClientIP:  clientIP,
		ExpiresAt: time.Now().Add(OAuthStateTTL * time.Second).Unix(),
	}

	// Serialize state | 序列化state
	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return "", err
	}

	// Store to Redis | 存储到Redis
	key := OAuthStateKeyPrefix + nonce
	if err := s.cache.SetEx(ctx, key, string(stateJSON), OAuthStateTTL); err != nil {
		return "", err
	}

	return nonce, nil
}

// validateAndDeleteState Validate and delete OAuth state | 验证并删除OAuth state
func (s *OAuthService) validateAndDeleteState(ctx context.Context, state, provider, action string) (*OAuthState, error) {
	// Get state from Redis | 从Redis获取state
	key := OAuthStateKeyPrefix + state
	stateJSON, err := s.cache.Get(ctx, key)
	if err != nil || stateJSON == "" {
		return nil, errors.New("授权参数无效或已过期")
	}

	// Delete state (single use) | 删除state（单次使用）
	if _, delErr := s.cache.Del(ctx, key); delErr != nil {
		s.logger.Warn("Failed to delete OAuth state | 删除OAuth state失败",
			tracing.WithTraceIDField(ctx), zap.Error(delErr))
	}

	// Parse state | 解析state
	var stateData OAuthState
	if err := json.Unmarshal([]byte(stateJSON), &stateData); err != nil {
		return nil, errors.New("授权参数无效")
	}

	// Validate expiration | 验证过期
	if time.Now().Unix() > stateData.ExpiresAt {
		return nil, errors.New("授权参数已过期")
	}

	// Validate provider | 验证provider
	if stateData.Provider != provider {
		return nil, errors.New("授权参数不匹配")
	}

	// Validate action | 验证action
	if stateData.Action != action {
		return nil, errors.New("授权参数不匹配")
	}

	return &stateData, nil
}

// canAutoRegister Check if auto register is allowed | 检查是否允许自动注册
func (s *OAuthService) canAutoRegister(ctx context.Context) (bool, string) {
	// Check if registration is closed | 检查是否关闭注册
	isCloseRegister, err := s.settings.GetSettingByKey(ctx, _const.SafeIsCloseRegister, _const.SettingBoolFalse.String())
	if err != nil {
		s.logger.Error("Failed to query registration settings | 查询注册设置失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return false, "系统配置查询失败"
	}
	if isCloseRegister == _const.SettingBoolTrue.String() {
		return false, "系统已关闭注册功能，无法自动创建账户"
	}

	// TODO: Check if invitation code is required | 检查是否需要邀请码
	// isRequireInviteCode, err := s.settings.GetSettingByKey(ctx, _const.SafeIsRequireInviteCode, _const.SettingBoolFalse.String())
	// if err == nil && isRequireInviteCode == _const.SettingBoolTrue.String() {
	//     return false, "系统开启了邀请码注册，无法自动创建账户"
	// }

	return true, ""
}

// loginExistingUser Login existing user | 登录已存在用户
func (s *OAuthService) loginExistingUser(ctx context.Context, userID int, provider string, _ *oauth.UserInfo, ua, clientIP string) (*schema.OAuthCallbackResponse, error) {
	// Get user | 获取用户
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user | 获取用户失败",
			tracing.WithTraceIDField(ctx), zap.Error(err), zap.Int("user_id", userID))
		return nil, errors.New("用户不存在")
	}

	// Check user status | 检查用户状态
	if u.Status == user.StatusBlocked {
		return nil, errors.New("账户已被锁定使用")
	}

	// Check temporary ban | 检查临时封禁
	if isDisabled := stputil.IsDisable(u.ID); isDisabled {
		remainingTime, err := stputil.GetDisableTime(u.ID)
		if err != nil {
			return nil, errors.New("账户已被限制使用")
		}
		return nil, fmt.Errorf("账户已被限制使用, 解除时间: %s", time_tools.CalculateRemainingTime(remainingTime))
	}

	// Create token | 创建token
	token, err := saGin.Login(u.ID, ua)
	if err != nil {
		s.logger.Error("Failed to create login session | 创建登录会话失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("登录失败")
	}

	// Set user role | 设置用户角色
	if err = stputil.SetRoles(u.ID, satoken.GetUserRole(u.Role.String())); err != nil {
		s.logger.Warn("Failed to set user role | 设置用户角色失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
	}

	s.logger.Info("OAuth login success | OAuth登录成功",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", provider),
		zap.Int("user_id", userID),
		zap.String("client_ip", clientIP))

	return &schema.OAuthCallbackResponse{
		Action:   OAuthResultLogin,
		UserID:   u.ID,
		Username: u.Username,
		Token:    token,
	}, nil
}

// registerNewUser Register new user via OAuth | 通过OAuth注册新用户
func (s *OAuthService) registerNewUser(ctx context.Context, provider string, userInfo *oauth.UserInfo, ua, clientIP string) (*schema.OAuthCallbackResponse, error) {
	// Generate unique username | 生成唯一用户名
	username := s.generateUniqueUsername(ctx, userInfo.Username, provider)

	// Generate random password | 生成随机密码
	randomPassword, err := generateRandomPassword(16)
	if err != nil {
		s.logger.Error("Failed to generate random password | 生成随机密码失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("创建账户失败")
	}
	hashedPassword, err := utils.HashPassword(randomPassword)
	if err != nil {
		s.logger.Error("Failed to hash password | 密码加密失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("创建账户失败")
	}

	// Create user | 创建用户
	newUser, err := s.userRepo.CreateWithBuilder(ctx, func(creator *ent.UserCreate) *ent.UserCreate {
		creator.SetUsername(username).
			SetPassword(hashedPassword).
			SetEmail(fmt.Sprintf("%s_%s@oauth.local", provider, userInfo.ProviderUserID))

		if userInfo.Avatar != "" {
			creator.SetAvatar(userInfo.Avatar)
		}
		return creator
	})
	if err != nil {
		s.logger.Error("Failed to create user | 创建用户失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("创建账户失败")
	}

	// Create OAuth binding | 创建OAuth绑定
	_, err = s.userOAuthRepo.Create(ctx, newUser.ID, provider, userInfo.ProviderUserID,
		userInfo.Username, userInfo.Email, userInfo.Avatar, userInfo.ExtraData)
	if err != nil {
		s.logger.Error("Failed to create OAuth binding | 创建OAuth绑定失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		// Rollback user creation | 回滚用户创建（这里简化处理，实际可能需要事务）
		return nil, errors.New("创建账户失败")
	}

	// Create token | 创建token
	token, err := saGin.Login(newUser.ID, ua)
	if err != nil {
		s.logger.Error("Failed to create login session | 创建登录会话失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, errors.New("登录失败")
	}

	// Set user role | 设置用户角色
	if err = stputil.SetRoles(newUser.ID, satoken.GetUserRole(user.RoleUser.String())); err != nil {
		s.logger.Warn("Failed to set user role | 设置用户角色失败",
			tracing.WithTraceIDField(ctx), zap.Error(err))
	}

	s.logger.Info("OAuth register success | OAuth注册成功",
		tracing.WithTraceIDField(ctx),
		zap.String("provider", provider),
		zap.Int("user_id", newUser.ID),
		zap.String("username", username),
		zap.String("client_ip", clientIP))

	return &schema.OAuthCallbackResponse{
		Action:   OAuthResultRegister,
		UserID:   newUser.ID,
		Username: newUser.Username,
		Token:    token,
	}, nil
}

// generateUniqueUsername Generate unique username | 生成唯一用户名
func (s *OAuthService) generateUniqueUsername(ctx context.Context, baseUsername, provider string) string {
	if baseUsername == "" {
		baseUsername = provider + "_user"
	}

	username := baseUsername
	suffix := 1

	for {
		exists, err := s.userRepo.ExistsByUsername(ctx, username)
		if err != nil || !exists {
			break
		}
		username = fmt.Sprintf("%s_%d", baseUsername, suffix)
		suffix++
		if suffix > 1000 {
			// Fallback to random suffix | 回退到随机后缀
			randomBytes := make([]byte, 4)
			if _, randErr := rand.Read(randomBytes); randErr != nil {
				username = fmt.Sprintf("%s_%d", baseUsername, time.Now().UnixNano())
			} else {
				username = fmt.Sprintf("%s_%s", baseUsername, hex.EncodeToString(randomBytes))
			}
			break
		}
	}

	return username
}

// generateRandomPassword Generate random password | 生成随机密码
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
