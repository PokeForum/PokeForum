package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/const"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
	"github.com/click33/sa-token-go/stputil"
	"go.uber.org/zap"
)

// IAuthService 认证服务接口
type IAuthService interface {
	// Register 用户注册
	Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error)
	// Login 用户登录
	Login(ctx context.Context, req schema.LoginRequest) (*ent.User, error)
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
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
// whitelist: 白名单字符串，每行一个域名，如 "gmail.com\nqq.com"
func isEmailDomainInWhitelist(emailDomain, whitelist string) bool {
	if whitelist == "" {
		return false // 白名单为空，不允许任何域名
	}

	// 将白名单按行分割
	domains := strings.Split(whitelist, "\n")

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
