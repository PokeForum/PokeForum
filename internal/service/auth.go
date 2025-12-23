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
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	smtp "github.com/PokeForum/PokeForum/internal/pkg/email"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// IAuthService 认证服务接口
type IAuthService interface {
	// Register 用户注册
	Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error)
	// Login 用户登录
	Login(ctx context.Context, req schema.LoginRequest) (*ent.User, error)
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
	// SendForgotPasswordCode 发送找回密码验证码
	SendForgotPasswordCode(ctx context.Context, req schema.ForgotPasswordRequest) (*schema.ForgotPasswordResponse, error)
	// ResetPassword 重置密码
	ResetPassword(ctx context.Context, req schema.ResetPasswordRequest) (*schema.ResetPasswordResponse, error)
}

// AuthService 认证服务实现
type AuthService struct {
	db       *ent.Client
	cache    cache.ICacheService
	logger   *zap.Logger
	settings ISettingsService // 添加设置服务依赖
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger, settings ISettingsService) IAuthService {
	return &AuthService{
		db:       db,
		cache:    cacheService,
		logger:   logger,
		settings: settings,
	}
}

// Register 用户注册
// 验证用户名和邮箱是否已存在，然后创建新用户
func (s *AuthService) Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error) {
	// 检查系统是否允许注册
	isCloseRegister, err := s.settings.GetSettingByKey(ctx, _const.SafeIsCloseRegister, _const.SettingBoolTrue.String())
	if err != nil {
		s.logger.Error("查询注册设置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询注册设置失败: %w", err)
	}
	if isCloseRegister == _const.SettingBoolTrue.String() {
		return nil, errors.New("系统已关闭注册功能")
	}

	// 检查用户名是否已存在
	existingUser, err := s.db.User.Query().
		Where(user.UsernameEQ(req.Username)).
		First(ctx)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.db.User.Query().
		Where(user.EmailEQ(req.Email)).
		First(ctx)
	if err == nil && existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("查询邮箱失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询邮箱失败: %w", err)
	}

	// 检查邮箱白名单设置
	isEnableEmailWhitelist, err := s.settings.GetSettingByKey(ctx, _const.SafeIsEnableEmailWhitelist, _const.SettingBoolFalse.String())
	if err != nil {
		s.logger.Error("查询邮箱白名单设置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询邮箱白名单设置失败: %w", err)
	}

	// 如果开启了邮箱白名单，则进行白名单验证
	if isEnableEmailWhitelist == _const.SettingBoolTrue.String() {
		// 获取邮箱白名单列表
		emailWhitelist, err := s.settings.GetSettingByKey(ctx, _const.SafeEmailWhitelist, "")
		if err != nil {
			s.logger.Error("查询邮箱白名单失败", tracing.WithTraceIDField(ctx), zap.Error(err))
			return nil, fmt.Errorf("查询邮箱白名单失败: %w", err)
		}

		// 提取邮箱域名
		emailDomain := utils.ExtractEmailDomain(req.Email)
		if emailDomain == "" {
			return nil, errors.New("邮箱格式无效")
		}

		// 验证邮箱域名是否在白名单中
		if !isEmailDomainInWhitelist(emailDomain, emailWhitelist) {
			s.logger.Warn("邮箱域名不在白名单中", tracing.WithTraceIDField(ctx), zap.String("email", req.Email), zap.String("domain", emailDomain))
			return nil, errors.New("邮箱域名不在允许注册的白名单中")
		}
	}
	// 生成密码盐
	pwdSalt := utils.GeneratePasswordSalt()

	// 密码加盐
	req.Password = utils.CombinePasswordWithSalt(req.Password, pwdSalt)

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("密码加密失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	newUser, err := s.db.User.Create().
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPassword(hashedPassword).
		SetPasswordSalt(pwdSalt).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return newUser, nil
}

// Login 用户登录
// 根据用户名查找用户，验证密码是否正确
func (s *AuthService) Login(ctx context.Context, req schema.LoginRequest) (*ent.User, error) {
	// 根据邮箱查找用户
	u, err := s.db.User.Query().
		Where(user.EmailEQ(req.Email)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查封禁-长期
	if u.Status == user.StatusBlocked {
		return nil, errors.New("账户已被锁定使用")
	}

	// 检查封禁-短期
	if isDisabled := stputil.IsDisable(u.ID); isDisabled {
		remainingTime, _ := stputil.GetDisableTime(u.ID) // 查询剩余时间, 单位（秒）
		return nil, fmt.Errorf("账户已被限制使用, 解除时间: %s", time_tools.CalculateRemainingTime(remainingTime))
	}

	// 拼接密码和盐
	combinedPassword := utils.CombinePasswordWithSalt(req.Password, u.PasswordSalt)

	// 验证密码
	if ok := utils.CheckPasswordHash(combinedPassword, u.Password); !ok {
		return nil, errors.New("密码错误")
	}

	return u, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := s.db.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("查询用户失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// isEmailDomainInWhitelist 检查邮箱域名是否在白名单中
// emailDomain: 要检查的域名，如 "gmail.com"
// whitelist: 白名单字符串，用英文逗号分隔域名，如 "gmail.com,qq.com"
func isEmailDomainInWhitelist(emailDomain, whitelist string) bool {
	if whitelist == "" {
		return false // 白名单为空，不允许任何域名
	}

	// 将白名单按英文逗号分割
	domains := strings.Split(whitelist, ",")

	// 遍历白名单中的每个域名
	for _, domain := range domains {
		// 去除前后空格
		whitelistDomain := strings.TrimSpace(domain)
		// 跳过空行
		if whitelistDomain == "" {
			continue
		}
		// 检查是否匹配（不区分大小写）
		if strings.EqualFold(emailDomain, whitelistDomain) {
			return true
		}
	}

	return false
}

// SendForgotPasswordCode 发送找回密码验证码
func (s *AuthService) SendForgotPasswordCode(ctx context.Context, req schema.ForgotPasswordRequest) (*schema.ForgotPasswordResponse, error) {
	s.logger.Info("发送找回密码验证码", zap.String("email", req.Email), tracing.WithTraceIDField(ctx))

	// 检查邮箱是否存在
	userData, err := s.db.User.Query().
		Where(user.EmailEQ(req.Email)).
		Select(user.FieldID, user.FieldEmail).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("该邮箱未注册")
		}
		s.logger.Error("查询用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查发送频率限制
	limitKey := fmt.Sprintf("password:reset:limit:%s", req.Email)
	limitValue, err := s.cache.Get(ctx, limitKey)
	if err == nil && limitValue != "" {
		sendCount := 0
		if _, parseErr := fmt.Sscanf(limitValue, "%d", &sendCount); parseErr == nil && sendCount >= 3 {
			return nil, errors.New("发送次数过多，请1小时后再试")
		}
	}

	// 生成6位随机验证码
	code, err := generateVerifyCode()
	if err != nil {
		s.logger.Error("生成验证码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存储验证码到Redis（10分钟有效期）
	codeKey := fmt.Sprintf("password:reset:code:%s", req.Email)
	err = s.cache.SetEx(ctx, codeKey, code, 600)
	if err != nil {
		s.logger.Error("存储验证码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	// 更新发送频率限制
	newCount := 1
	if limitValue != "" {
		var val int
		if _, parseErr := fmt.Sscanf(limitValue, "%d", &val); parseErr == nil {
			newCount = val + 1
		}
	}
	if err := s.cache.SetEx(ctx, limitKey, fmt.Sprintf("%d", newCount), 3600); err != nil {
		s.logger.Warn("更新发送频率限制失败", zap.String("key", limitKey), zap.Error(err), tracing.WithTraceIDField(ctx))
	}

	// 发送重置密码邮件
	err = s.sendPasswordResetEmail(ctx, userData.Email, code)
	if err != nil {
		s.logger.Error("发送重置密码邮件失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("发送重置密码邮件失败: %w", err)
	}

	s.logger.Info("找回密码验证码发送成功", zap.String("email", req.Email), tracing.WithTraceIDField(ctx))

	return &schema.ForgotPasswordResponse{
		Sent:      true,
		Message:   "验证码已发送到您的邮箱，请查收",
		ExpiresIn: 600,
	}, nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req schema.ResetPasswordRequest) (*schema.ResetPasswordResponse, error) {
	s.logger.Info("重置密码", zap.String("email", req.Email), tracing.WithTraceIDField(ctx))

	// 获取存储的验证码
	codeKey := fmt.Sprintf("password:reset:code:%s", req.Email)
	storedCode, err := s.cache.Get(ctx, codeKey)
	if err != nil || storedCode == "" {
		return nil, errors.New("验证码不存在或已过期")
	}

	// 验证验证码
	if storedCode != req.Code {
		return nil, errors.New("验证码错误")
	}

	// 查询用户
	userData, err := s.db.User.Query().
		Where(user.EmailEQ(req.Email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("查询用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 密码加盐加密
	combinedPassword := utils.CombinePasswordWithSalt(req.NewPassword, userData.PasswordSalt)
	hashedPassword, err := utils.HashPassword(combinedPassword)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	_, err = s.db.User.UpdateOneID(userData.ID).
		SetPassword(hashedPassword).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新密码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新密码失败: %w", err)
	}

	// 删除验证码缓存
	if _, err := s.cache.Del(ctx, codeKey); err != nil {
		s.logger.Warn("删除验证码缓存失败", zap.String("key", codeKey), zap.Error(err), tracing.WithTraceIDField(ctx))
	}

	s.logger.Info("密码重置成功", zap.String("email", req.Email), tracing.WithTraceIDField(ctx))

	return &schema.ResetPasswordResponse{
		Success: true,
		Message: "密码重置成功，请使用新密码登录",
	}, nil
}

// generateVerifyCode 生成6位随机验证码
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

// sendPasswordResetEmail 发送重置密码邮件
func (s *AuthService) sendPasswordResetEmail(ctx context.Context, email, code string) error {
	// 获取网站设置
	siteConfig, err := s.settings.GetSeoSettings(ctx)
	if err != nil {
		s.logger.Error("获取网站配置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取网站配置失败: %w", err)
	}

	// 获取SMTP配置
	smtpConfig, err := s.settings.GetSMTPConfig(ctx)
	if err != nil {
		s.logger.Error("获取SMTP配置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取SMTP配置失败: %w", err)
	}

	// 创建邮件模板渲染器
	emailTemplate := smtp.NewEmailTemplate(s.settings, s.logger)

	// 渲染邮件模板
	htmlBody, err := emailTemplate.RenderPasswordResetTemplate(ctx, code, siteConfig.WebSiteName)
	if err != nil {
		s.logger.Error("渲染邮件模板失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	// 发送邮件
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
