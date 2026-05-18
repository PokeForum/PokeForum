package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/click33/sa-token-go/stputil"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/asynq"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	smtp "github.com/PokeForum/PokeForum/internal/pkg/email"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// IAuthService Authentication service interface | 认证服务接口
type IAuthService interface {
	// Register User registration | 用户注册
	Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error)
	// Login User login | 用户登录
	Login(ctx context.Context, req schema.LoginRequest) (*ent.User, error)
	// GetUserByID Get user by ID | 根据ID获取用户
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
	// SendForgotPasswordCode Send password reset verification code | 发送找回密码验证码
	SendForgotPasswordCode(ctx context.Context, req schema.ForgotPasswordRequest) (*schema.ForgotPasswordResponse, error)
	// ResetPassword Reset password | 重置密码
	ResetPassword(ctx context.Context, req schema.ResetPasswordRequest) (*schema.ResetPasswordResponse, error)
	// RecordLoginLog Record login log | 记录登录日志
	RecordLoginLog(ctx context.Context, userID int, ip, ua string)
}

// AuthService Authentication service implementation | 认证服务实现
type AuthService struct {
	userRepo          repository.IUserRepository
	loginLogRepo      repository.IUserLoginLogRepository
	cache             cache.ICacheService
	logger            *zap.Logger
	settings          ISettingsService
	invitationCodeSvc IInvitationCodeService
	taskManager       *asynq.TaskManager
}

// NewAuthService Create authentication service instance | 创建认证服务实例
func NewAuthService(userRepo repository.IUserRepository, loginLogRepo repository.IUserLoginLogRepository, cacheService cache.ICacheService, logger *zap.Logger, settings ISettingsService, invitationCodeSvc IInvitationCodeService, taskManager *asynq.TaskManager) IAuthService {
	return &AuthService{
		userRepo:          userRepo,
		loginLogRepo:      loginLogRepo,
		cache:             cacheService,
		logger:            logger,
		settings:          settings,
		invitationCodeSvc: invitationCodeSvc,
		taskManager:       taskManager,
	}
}

// Register User registration | 用户注册
// Verify if username and email already exist, then create new user | 验证用户名和邮箱是否已存在，然后创建新用户
func (s *AuthService) Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error) {
	// Check if system allows registration | 检查系统是否允许注册
	isCloseRegister, err := s.settings.GetSettingByKey(ctx, _const.SafeIsCloseRegister, _const.SettingBoolTrue.String())
	if err != nil {
		s.logger.Error("Failed to query registration settings | 查询注册设置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询注册设置失败: %w", err)
	}
	if isCloseRegister == _const.SettingBoolTrue.String() {
		return nil, errors.New("系统已关闭注册功能")
	}

	// Check if invitation code is required | 检查是否需要邀请码
	isInvitationCodeEnabled, err := s.invitationCodeSvc.IsInvitationCodeEnabled(ctx)
	if err != nil {
		s.logger.Error("Failed to check invitation code setting | 检查邀请码设置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("检查邀请码设置失败: %w", err)
	}
	if isInvitationCodeEnabled {
		if req.InvitationCode == "" {
			return nil, errors.New("请输入邀请码")
		}
		// Validate invitation code | 验证邀请码
		if _, err := s.invitationCodeSvc.ValidateCode(ctx, req.InvitationCode); err != nil {
			return nil, err
		}
	}

	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to query user | 查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	existingEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to query email | 查询邮箱失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	if existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// Check email whitelist settings | 检查邮箱白名单设置
	isEnableEmailWhitelist, err := s.settings.GetSettingByKey(ctx, _const.SafeIsEnableEmailWhitelist, _const.SettingBoolFalse.String())
	if err != nil {
		s.logger.Error("Failed to query email whitelist settings | 查询邮箱白名单设置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询邮箱白名单设置失败: %w", err)
	}

	// If email whitelist is enabled, perform whitelist verification | 如果开启了邮箱白名单，则进行白名单验证
	if isEnableEmailWhitelist == _const.SettingBoolTrue.String() {
		// Get email whitelist | 获取邮箱白名单列表
		emailWhitelist, err := s.settings.GetSettingByKey(ctx, _const.SafeEmailWhitelist, "")
		if err != nil {
			s.logger.Error("Failed to query email whitelist | 查询邮箱白名单失败", tracing.WithTraceIDField(ctx), zap.Error(err))
			return nil, fmt.Errorf("查询邮箱白名单失败: %w", err)
		}

		// Extract email domain | 提取邮箱域名
		emailDomain := utils.ExtractEmailDomain(req.Email)
		if emailDomain == "" {
			return nil, errors.New("邮箱格式无效")
		}

		// Verify if email domain is in whitelist | 验证邮箱域名是否在白名单中
		if !isEmailDomainInWhitelist(emailDomain, emailWhitelist) {
			s.logger.Warn("Email domain not in whitelist | 邮箱域名不在白名单中", tracing.WithTraceIDField(ctx), zap.String("email", req.Email), zap.String("domain", emailDomain))
			return nil, errors.New("邮箱域名不在允许注册的白名单中")
		}
	}

	// Hash password | 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password | 密码加密失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	newUser, err := s.userRepo.Create(ctx, req.Username, req.Email, hashedPassword)
	if err != nil {
		s.logger.Error("Failed to create user | 创建用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}

	// Use invitation code if provided | 如果提供了邀请码，使用邀请码
	if isInvitationCodeEnabled && req.InvitationCode != "" {
		if err := s.invitationCodeSvc.UseCode(ctx, req.InvitationCode, newUser.ID, "", ""); err != nil {
			s.logger.Warn("Failed to use invitation code | 使用邀请码失败",
				tracing.WithTraceIDField(ctx),
				zap.Error(err),
				zap.String("code", req.InvitationCode),
				zap.Int("user_id", newUser.ID))
		}
	}

	return newUser, nil
}

// Login User login | 用户登录
// Find user by username and verify password | 根据用户名查找用户，验证密码是否正确
func (s *AuthService) Login(ctx context.Context, req schema.LoginRequest) (*ent.User, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to query user | 查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	if u == nil {
		return nil, errors.New("用户不存在")
	}

	// Check permanent ban | 检查封禁-长期
	if u.Status == user.StatusBlocked {
		return nil, errors.New("账户已被锁定使用")
	}

	// Check temporary ban | 检查封禁-短期
	if isDisabled := stputil.IsDisable(u.ID); isDisabled {
		remainingTime, err := stputil.GetDisableTime(u.ID) // Query remaining time in seconds | 查询剩余时间, 单位（秒）
		if err != nil {
			return nil, errors.New("账户已被限制使用")
		}
		return nil, fmt.Errorf("账户已被限制使用, 解除时间: %s", time_tools.CalculateRemainingTime(remainingTime))
	}

	// Verify password | 验证密码
	if ok := utils.CheckPasswordHash(req.Password, u.Password); !ok {
		return nil, errors.New("密码错误")
	}

	return u, nil
}

// GetUserByID Get user by ID | 根据ID获取用户
func (s *AuthService) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to query user | 查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	return u, nil
}

// isEmailDomainInWhitelist Check if email domain is in whitelist | 检查邮箱域名是否在白名单中
// emailDomain: Domain to check, e.g., "gmail.com" | 要检查的域名，如 "gmail.com"
// whitelist: Whitelist string, comma-separated domains, e.g., "gmail.com,qq.com" | 白名单字符串，用英文逗号分隔域名，如 "gmail.com,qq.com"
func isEmailDomainInWhitelist(emailDomain, whitelist string) bool {
	if whitelist == "" {
		return false // Whitelist is empty, no domain allowed | 白名单为空，不允许任何域名
	}

	// Split whitelist by comma | 将白名单按英文逗号分割
	domains := strings.Split(whitelist, ",")

	// Iterate through each domain in whitelist | 遍历白名单中的每个域名
	for _, domain := range domains {
		// Trim whitespace | 去除前后空格
		whitelistDomain := strings.TrimSpace(domain)
		// Skip empty lines | 跳过空行
		if whitelistDomain == "" {
			continue
		}
		// Check if matched (case-insensitive) | 检查是否匹配（不区分大小写）
		if strings.EqualFold(emailDomain, whitelistDomain) {
			return true
		}
	}

	return false
}

// SendForgotPasswordCode Send password recovery verification code | 发送找回密码验证码
func (s *AuthService) SendForgotPasswordCode(ctx context.Context, req schema.ForgotPasswordRequest) (*schema.ForgotPasswordResponse, error) {
	s.logger.Info("Sending password recovery verification code | 发送找回密码验证码", tracing.WithTraceIDField(ctx), zap.String("email", req.Email))

	userData, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to query user | 查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	if userData == nil {
		return nil, errors.New("该邮箱未注册")
	}

	// Check and update sending frequency limit atomically | 原子性地检查和更新发送频率限制
	limitKey := fmt.Sprintf("password:reset:limit:%s", req.Email)

	luaScript := `
		local key = KEYS[1]
		local max_count = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])
		
		local new_count = redis.call('INCR', key)
		
		if new_count == 1 then
			redis.call('EXPIRE', key, ttl)
		end
		
		if new_count > max_count then
			return {0, new_count}
		end
		
		return {1, new_count}
	`

	result, err := s.cache.Eval(ctx, luaScript, []string{limitKey}, 3, 3600)
	if err != nil {
		s.logger.Error("Failed to check frequency limit | 检查频率限制失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("检查频率限制失败: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 2 {
		s.logger.Error("Invalid result from Lua script | Lua 脚本返回结果无效", tracing.WithTraceIDField(ctx), zap.Any("result", result))
		return nil, errors.New("系统错误，请稍后重试")
	}

	allowed, ok := resultSlice[0].(int64)
	if !ok {
		s.logger.Error("Invalid allowed value from Lua script | Lua 脚本返回的 allowed 值无效", tracing.WithTraceIDField(ctx), zap.Any("result", result))
		return nil, errors.New("系统错误，请稍后重试")
	}

	if allowed == 0 {
		s.logger.Warn("Password reset request exceeds frequency limit | 找回密码请求超过频率限制", tracing.WithTraceIDField(ctx), zap.String("email", req.Email))
		return nil, errors.New("发送次数过多，请1小时后再试")
	}

	sendCount, ok := resultSlice[1].(int64)
	if !ok {
		s.logger.Error("Invalid sendCount value from Lua script | Lua 脚本返回的 sendCount 值无效", tracing.WithTraceIDField(ctx), zap.Any("result", result))
		return nil, errors.New("系统错误，请稍后重试")
	}

	// Generate 6-digit random verification code | 生成6位随机验证码
	code, err := generateVerifyCode()
	if err != nil {
		s.logger.Error("Failed to generate verification code | 生成验证码失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}

	// Store verification code to Redis (10 minutes expiry) | 存储验证码到Redis（10分钟有效期）
	codeKey := fmt.Sprintf("password:reset:code:%s", req.Email)
	err = s.cache.SetEx(ctx, codeKey, code, 600)
	if err != nil {
		s.logger.Error("Failed to store verification code | 存储验证码失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	s.logger.Info("Password reset frequency limit updated | 找回密码频率限制已更新", tracing.WithTraceIDField(ctx), zap.String("email", req.Email), zap.Int64("count", sendCount))

	// Send password reset email | 发送重置密码邮件
	err = s.sendPasswordResetEmail(ctx, userData.Email, code)
	if err != nil {
		s.logger.Error("Failed to send password reset email | 发送重置密码邮件失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("发送重置密码邮件失败: %w", err)
	}

	s.logger.Info("Password recovery verification code sent successfully | 找回密码验证码发送成功", tracing.WithTraceIDField(ctx), zap.String("email", req.Email))

	return &schema.ForgotPasswordResponse{
		Sent:      true,
		Message:   "验证码已发送到您的邮箱，请查收",
		ExpiresIn: 600,
	}, nil
}

// ResetPassword Reset password | 重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req schema.ResetPasswordRequest) (*schema.ResetPasswordResponse, error) {
	s.logger.Info("Resetting password | 重置密码", tracing.WithTraceIDField(ctx), zap.String("email", req.Email))

	// Get stored verification code | 获取存储的验证码
	codeKey := fmt.Sprintf("password:reset:code:%s", req.Email)
	storedCode, err := s.cache.Get(ctx, codeKey)
	if err != nil || storedCode == "" {
		return nil, errors.New("验证码不存在或已过期")
	}

	// Verify verification code | 验证验证码
	if storedCode != req.Code {
		return nil, errors.New("验证码错误")
	}

	userData, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to query user | 查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}
	if userData == nil {
		return nil, errors.New("用户不存在")
	}

	// Hash password | 密码加密
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash password | 密码加密失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	err = s.userRepo.UpdatePassword(ctx, userData.ID, hashedPassword)
	if err != nil {
		s.logger.Error("Failed to update password | 更新密码失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, err
	}

	// Delete verification code cache | 删除验证码缓存
	if _, err := s.cache.Del(ctx, codeKey); err != nil {
		s.logger.Warn("Failed to delete verification code cache | 删除验证码缓存失败", tracing.WithTraceIDField(ctx), zap.String("key", codeKey), zap.Error(err))
	}

	s.logger.Info("Password reset successfully | 密码重置成功", tracing.WithTraceIDField(ctx), zap.String("email", req.Email))

	return &schema.ResetPasswordResponse{
		Success: true,
		Message: "密码重置成功，请使用新密码登录",
	}, nil
}

// generateVerifyCode Generate 6-digit random verification code | 生成6位随机验证码
func generateVerifyCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}

// sendPasswordResetEmail Send password reset email | 发送重置密码邮件
func (s *AuthService) sendPasswordResetEmail(ctx context.Context, email, code string) error {
	// Get website settings | 获取网站设置
	siteConfig, err := s.settings.GetSeoSettings(ctx)
	if err != nil {
		s.logger.Error("Failed to get website configuration | 获取网站配置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("获取网站配置失败: %w", err)
	}

	// Get SMTP configuration | 获取SMTP配置
	smtpConfig, err := s.settings.GetSMTPConfig(ctx)
	if err != nil {
		s.logger.Error("Failed to get SMTP configuration | 获取SMTP配置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("获取SMTP配置失败: %w", err)
	}

	// Create email template renderer | 创建邮件模板渲染器
	emailTemplate := smtp.NewEmailTemplate(s.settings, s.logger)

	// Render email template | 渲染邮件模板
	htmlBody, err := emailTemplate.RenderPasswordResetTemplate(ctx, code, siteConfig.WebSiteName)
	if err != nil {
		s.logger.Error("Failed to render email template | 渲染邮件模板失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	// Send email | 发送邮件
	sp := smtp.NewSMTPPool(smtp.SMTPConfig{
		Name:       siteConfig.WebSiteName,
		Address:    smtpConfig.Address,
		Host:       smtpConfig.Host,
		Port:       smtpConfig.Port,
		User:       smtpConfig.Username,
		Password:   smtpConfig.Password,
		Encryption: smtpConfig.ForcedSSL,
		Keepalive:  smtpConfig.ConnectionValidity,
	}, s.logger)
	defer sp.Close()

	if err = sp.Send(ctx, email, fmt.Sprintf("【%s】密码重置验证码", siteConfig.WebSiteName), htmlBody); err != nil {
		return err
	}

	return nil
}

// RecordLoginLog Record login log | 记录登录日志
func (s *AuthService) RecordLoginLog(ctx context.Context, userID int, ip, ua string) {
	payload := &LoginLogPayload{
		UserID:  userID,
		IP:      ip,
		TraceID: tracing.GetTraceID(ctx),
	}

	task, err := NewLoginLogTask(payload)
	if err != nil {
		s.logger.Error("创建登录日志任务失败",

			tracing.WithTraceIDField(ctx), zap.Int("user_id", userID),
			zap.String("ip_address", ip),

			zap.Error(err))
		return
	}

	_, err = s.taskManager.EnqueueContext(ctx, task)
	if err != nil {
		s.logger.Error("提交登录日志任务失败",

			tracing.WithTraceIDField(ctx), zap.Int("user_id", userID),
			zap.String("ip_address", ip),

			zap.Error(err))
	}
}
