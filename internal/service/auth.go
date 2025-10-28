package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

// IAuthService 认证服务接口
type IAuthService interface {
	// Register 用户注册
	Register(ctx context.Context, username, email, password string) (*ent.User, error)
	// Login 用户登录
	Login(ctx context.Context, username, password string) (*ent.User, error)
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
}

// AuthService 认证服务实现
type AuthService struct {
	db    *ent.Client
	cache *redis.Pool
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *ent.Client, cache *redis.Pool) IAuthService {
	return &AuthService{
		db:    db,
		cache: cache,
	}
}

// Register 用户注册
// 验证用户名和邮箱是否已存在，然后创建新用户
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*ent.User, error) {
	// 检查用户名是否已存在
	existingUser, err := s.db.User.Query().
		Where(user.UsernameEQ(username)).
		First(ctx)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.db.User.Query().
		Where(user.EmailEQ(email)).
		First(ctx)
	if err == nil && existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询邮箱失败: %w", err)
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	newUser, err := s.db.User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword(string(hashedPassword)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return newUser, nil
}

// Login 用户登录
// 根据用户名查找用户，验证密码是否正确
func (s *AuthService) Login(ctx context.Context, username, password string) (*ent.User, error) {
	// 根据用户名查找用户
	u, err := s.db.User.Query().
		Where(user.UsernameEQ(username)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
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
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}
