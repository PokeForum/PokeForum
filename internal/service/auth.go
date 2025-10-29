package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	db     *ent.Client
	cache  *redis.Pool
	logger *zap.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *ent.Client, cache *redis.Pool, logger *zap.Logger) IAuthService {
	return &AuthService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// Register 用户注册
// 验证用户名和邮箱是否已存在，然后创建新用户
func (s *AuthService) Register(ctx context.Context, req schema.RegisterRequest) (*ent.User, error) {
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

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	newUser, err := s.db.User.Create().
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPassword(string(hashedPassword)).
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
