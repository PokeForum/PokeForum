package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/stputil"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/categorymoderator"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/ent/userbalancelog"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// IUserManageService 用户管理服务接口
type IUserManageService interface {
	// GetUserList 获取用户列表
	GetUserList(ctx context.Context, req schema.UserListRequest) (*schema.UserListResponse, error)
	// CreateUser 创建用户
	CreateUser(ctx context.Context, req schema.UserCreateRequest) (*ent.User, error)
	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, req schema.UserUpdateRequest) (*ent.User, error)
	// UpdateUserStatus 更新用户状态
	UpdateUserStatus(ctx context.Context, req schema.UserStatusUpdateRequest) error
	// UpdateUserRole 更新用户身份
	UpdateUserRole(ctx context.Context, req schema.UserRoleUpdateRequest) error
	// UpdateUserPoints 更新用户积分
	UpdateUserPoints(ctx context.Context, req schema.UserPointsUpdateRequest) error
	// UpdateUserCurrency 更新用户货币
	UpdateUserCurrency(ctx context.Context, req schema.UserCurrencyUpdateRequest) error
	// SetModeratorCategories 设置版主管理版块
	SetModeratorCategories(ctx context.Context, req schema.ModeratorCategoryRequest) error
	// GetUserDetail 获取用户详情
	GetUserDetail(ctx context.Context, id int) (*schema.UserDetailResponse, error)
	// GetUserBalanceLog 获取用户余额变动记录
	GetUserBalanceLog(ctx context.Context, req schema.UserBalanceLogRequest) (*schema.UserBalanceLogResponse, error)
	// GetUserBalanceSummary 获取用户余额汇总信息
	GetUserBalanceSummary(ctx context.Context, userID int) (*schema.UserBalanceSummary, error)
	// GetUserPostCount 查询单个用户的发帖数
	GetUserPostCount(ctx context.Context, userID int) (int, error)
	// GetUserCommentCount 查询单个用户的评论数
	GetUserCommentCount(ctx context.Context, userID int) (int, error)
	// BanUser 封禁用户
	BanUser(ctx context.Context, req schema.UserBanRequest) error
	// UnbanUser 解封用户
	UnbanUser(ctx context.Context, req schema.UserUnbanRequest) error
}

// UserManageService 用户管理服务实现
type UserManageService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewUserManageService 创建用户管理服务实例
func NewUserManageService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IUserManageService {
	return &UserManageService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// checkOperatorPermission 校验操作者权限
// 管理员只能操作用户和版主，超级管理员可以操作所有身份
func (s *UserManageService) checkOperatorPermission(ctx context.Context, operatorID int, targetRole user.Role) error {
	// 获取操作者信息
	operator, err := s.db.User.Get(ctx, operatorID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("操作者不存在")
		}
		s.logger.Error("获取操作者信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取操作者信息失败: %w", err)
	}

	// 超级管理员可以操作所有身份
	if operator.Role == user.RoleSuperAdmin {
		return nil
	}

	// 管理员只能操作用户和版主
	if operator.Role == user.RoleAdmin {
		if targetRole == user.RoleAdmin || targetRole == user.RoleSuperAdmin {
			return errors.New("无权操作管理员或超级管理员")
		}
		return nil
	}

	return errors.New("无操作权限")
}

// GetUserList 获取用户列表
func (s *UserManageService) GetUserList(ctx context.Context, req schema.UserListRequest) (*schema.UserListResponse, error) {
	s.logger.Info("获取用户列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.User.Query()

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where(
			user.Or(
				user.UsernameContains(req.Keyword),
				user.EmailContains(req.Keyword),
			),
		)
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where(user.StatusEQ(user.Status(req.Status)))
	}

	// 身份筛选
	if req.Role != "" {
		query = query.Where(user.RoleEQ(user.Role(req.Role)))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 分页查询
	users, err := query.
		Order(ent.Desc(user.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 转换响应数据
	list := make([]schema.UserListItem, len(users))
	if len(users) == 0 {
		return &schema.UserListResponse{
			List:     list,
			Total:    int64(total),
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	// 收集用户ID
	userIDs := make([]int, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// 批量查询用户的发帖数和评论数
	postCounts := s.getUserPostCounts(ctx, userIDs)
	commentCounts := s.getUserCommentCounts(ctx, userIDs)

	// 构建响应数据
	for i, u := range users {
		list[i] = schema.UserListItem{
			ID:            u.ID,
			Username:      u.Username,
			Email:         u.Email,
			Avatar:        u.Avatar,
			Signature:     u.Signature,
			EmailVerified: u.EmailVerified,
			Points:        u.Points,
			Currency:      u.Currency,
			PostCount:     postCounts[u.ID],
			CommentCount:  commentCounts[u.ID],
			Status:        u.Status.String(),
			Role:          u.Role.String(),
			CreatedAt:     u.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:     u.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.UserListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateUser 创建用户
func (s *UserManageService) CreateUser(ctx context.Context, req schema.UserCreateRequest) (*ent.User, error) {
	s.logger.Info("创建用户", zap.String("username", req.Username), tracing.WithTraceIDField(ctx))

	// 检查用户名是否已存在
	existingUser, err := s.db.User.Query().
		Where(user.Or(user.UsernameEQ(req.Username), user.EmailEQ(req.Email))).
		First(ctx)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名或邮箱已存在")
	}
	if !ent.IsNotFound(err) {
		s.logger.Error("检查用户是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	// 生成密码盐和加密密码
	passwordSalt := utils.GeneratePasswordSalt()
	combinedPassword := utils.CombinePasswordWithSalt(req.Password, passwordSalt)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(combinedPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	u, err := s.db.User.Create().
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPassword(string(hashedPassword)).
		SetPasswordSalt(passwordSalt).
		SetRole(user.Role(req.Role)).
		SetAvatar(req.Avatar).
		SetSignature(req.Signature).
		SetReadme(req.Readme).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	s.logger.Info("用户创建成功", zap.Int("user_id", u.ID), tracing.WithTraceIDField(ctx))
	return u, nil
}

// UpdateUser 更新用户信息
func (s *UserManageService) UpdateUser(ctx context.Context, req schema.UserUpdateRequest) (*ent.User, error) {
	s.logger.Info("更新用户信息", zap.Int("user_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	existingUser, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户名和邮箱是否与其他用户冲突
	if req.Username != "" || req.Email != "" {
		conflictQuery := s.db.User.Query().
			Where(
				user.And(
					user.IDNEQ(req.ID),
					user.Or(
						user.UsernameEQ(req.Username),
						user.EmailEQ(req.Email),
					),
				),
			)
		if conflictUser, err := conflictQuery.First(ctx); err == nil && conflictUser != nil {
			return nil, errors.New("用户名或邮箱已被其他用户使用")
		} else if !ent.IsNotFound(err) {
			s.logger.Error("检查用户名邮箱冲突失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查用户名邮箱冲突失败: %w", err)
		}
	}

	// 构建更新操作
	update := s.db.User.UpdateOne(existingUser)
	if req.Username != "" {
		update.SetUsername(req.Username)
	}
	if req.Email != "" {
		update.SetEmail(req.Email)
	}
	if req.Avatar != "" {
		update.SetAvatar(req.Avatar)
	}
	if req.Signature != "" {
		update.SetSignature(req.Signature)
	}
	if req.Readme != "" {
		update.SetReadme(req.Readme)
	}

	u, err := update.Save(ctx)
	if err != nil {
		s.logger.Error("更新用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	s.logger.Info("用户信息更新成功", zap.Int("user_id", u.ID), tracing.WithTraceIDField(ctx))
	return u, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserManageService) UpdateUserStatus(ctx context.Context, req schema.UserStatusUpdateRequest) error {
	s.logger.Info("更新用户状态", zap.Int("user_id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 校验操作权限
	if err = s.checkOperatorPermission(ctx, req.OperatorID, u.Role); err != nil {
		return err
	}

	// 更新用户状态
	_, err = s.db.User.UpdateOneID(req.ID).
		SetStatus(user.Status(req.Status)).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新用户状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	s.logger.Info("用户状态更新成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// UpdateUserRole 更新用户身份
func (s *UserManageService) UpdateUserRole(ctx context.Context, req schema.UserRoleUpdateRequest) error {
	s.logger.Info("更新用户身份", zap.Int("user_id", req.ID), zap.String("role", req.Role), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	existingUser, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 更新用户身份
	_, err = s.db.User.UpdateOne(existingUser).
		SetRole(user.Role(req.Role)).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新用户身份失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户身份失败: %w", err)
	}

	// 如果用户不再是版主，清除其管理的版块（通过中间表）
	if user.Role(req.Role) != user.RoleModerator {
		_, err = s.db.CategoryModerator.Delete().
			Where(categorymoderator.UserIDEQ(req.ID)).
			Exec(ctx)
		if err != nil {
			s.logger.Error("清除用户管理版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("清除用户管理版块失败: %w", err)
		}
	}

	s.logger.Info("用户身份更新成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// UpdateUserPoints 更新用户积分
func (s *UserManageService) UpdateUserPoints(ctx context.Context, req schema.UserPointsUpdateRequest) error {
	s.logger.Info("更新用户积分", zap.Int("user_id", req.ID), zap.Int("points", req.Points), tracing.WithTraceIDField(ctx))

	// 获取用户当前积分
	u, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 计算新积分
	newPoints := u.Points + req.Points
	if newPoints < 0 {
		return errors.New("积分不能为负数")
	}

	// 更新用户积分
	err = s.db.User.UpdateOneID(req.ID).
		SetPoints(newPoints).
		Exec(ctx)
	if err != nil {
		s.logger.Error("更新用户积分失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户积分失败: %w", err)
	}

	// 创建余额变动记录
	err = s.createBalanceLog(
		ctx,
		req.ID,
		userbalancelog.TypePoints,
		req.Points,
		u.Points,
		newPoints,
		req.Reason,
		"",  // 操作者名称，这里可以后续从context中获取
		nil, // 操作者ID，这里可以后续从context中获取
		nil, // 关联业务ID
		"",  // 关联业务类型
		"",  // IP地址
		"",  // 用户代理
	)
	if err != nil {
		s.logger.Error("创建积分变动记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 不影响主流程，只记录错误
	}

	s.logger.Info("用户积分更新成功", zap.Int("old_points", u.Points), zap.Int("new_points", newPoints), tracing.WithTraceIDField(ctx))
	return nil
}

// UpdateUserCurrency 更新用户货币
func (s *UserManageService) UpdateUserCurrency(ctx context.Context, req schema.UserCurrencyUpdateRequest) error {
	s.logger.Info("更新用户货币", zap.Int("user_id", req.ID), zap.Int("currency", req.Currency), tracing.WithTraceIDField(ctx))

	// 获取用户当前货币
	u, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 计算新货币
	newCurrency := u.Currency + req.Currency
	if newCurrency < 0 {
		return errors.New("货币不能为负数")
	}

	// 更新用户货币
	err = s.db.User.UpdateOneID(req.ID).
		SetCurrency(newCurrency).
		Exec(ctx)
	if err != nil {
		s.logger.Error("更新用户货币失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户货币失败: %w", err)
	}

	// 创建余额变动记录
	err = s.createBalanceLog(
		ctx,
		req.ID,
		userbalancelog.TypeCurrency,
		req.Currency,
		u.Currency,
		newCurrency,
		req.Reason,
		"",  // 操作者名称，这里可以后续从context中获取
		nil, // 操作者ID，这里可以后续从context中获取
		nil, // 关联业务ID
		"",  // 关联业务类型
		"",  // IP地址
		"",  // 用户代理
	)
	if err != nil {
		s.logger.Error("创建货币变动记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 不影响主流程，只记录错误
	}

	s.logger.Info("用户货币更新成功", zap.Int("old_currency", u.Currency), zap.Int("new_currency", newCurrency), tracing.WithTraceIDField(ctx))
	return nil
}

// SetModeratorCategories 设置版主管理版块
func (s *UserManageService) SetModeratorCategories(ctx context.Context, req schema.ModeratorCategoryRequest) error {
	s.logger.Info("设置版主管理版块", zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在且是版主
	existingUser, err := s.db.User.Get(ctx, req.UserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if existingUser.Role != user.RoleModerator {
		return errors.New("只有版主才能设置管理版块")
	}

	// 检查版块是否存在
	categories, err := s.db.Category.Query().
		Where(category.IDIn(req.CategoryIDs...)).
		All(ctx)
	if err != nil {
		s.logger.Error("获取版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取版块信息失败: %w", err)
	}

	if len(categories) != len(req.CategoryIDs) {
		return errors.New("部分版块不存在")
	}

	// 使用事务更新版主管理的版块（通过中间表）
	tx, err := s.db.Tx(ctx)
	if err != nil {
		s.logger.Error("开启事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("开启事务失败: %w", err)
	}

	// 删除旧的版主关联记录
	_, err = tx.CategoryModerator.Delete().
		Where(categorymoderator.UserIDEQ(req.UserID)).
		Exec(ctx)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		s.logger.Error("删除旧版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除旧版主关联记录失败: %w", err)
	}

	// 批量插入新的版主关联记录
	if len(req.CategoryIDs) > 0 {
		bulk := make([]*ent.CategoryModeratorCreate, len(req.CategoryIDs))
		for i, categoryID := range req.CategoryIDs {
			bulk[i] = tx.CategoryModerator.Create().
				SetCategoryID(categoryID).
				SetUserID(req.UserID)
		}

		_, err = tx.CategoryModerator.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return err
			}
			s.logger.Error("批量插入版主关联记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("批量插入版主关联记录失败: %w", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		s.logger.Error("提交事务失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Info("版主管理版块设置成功", zap.Int("user_id", req.UserID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// GetUserDetail 获取用户详情
func (s *UserManageService) GetUserDetail(ctx context.Context, id int) (*schema.UserDetailResponse, error) {
	s.logger.Info("获取用户详情", zap.Int("user_id", id), tracing.WithTraceIDField(ctx))

	// 获取用户信息
	u, err := s.db.User.Query().
		Where(user.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户详情失败: %w", err)
	}

	// 通过中间表查询管理的版块
	managedCategories := make([]schema.CategoryBasicInfo, 0)
	if u.Role == user.RoleModerator {
		// 查询版主关联记录
		moderatorRecords, err := s.db.CategoryModerator.Query().
			Where(categorymoderator.UserIDEQ(id)).
			All(ctx)
		if err == nil && len(moderatorRecords) > 0 {
			// 收集版块ID
			categoryIDs := make([]int, len(moderatorRecords))
			for i, record := range moderatorRecords {
				categoryIDs[i] = record.CategoryID
			}

			// 批量查询版块信息
			categories, err := s.db.Category.Query().
				Where(category.IDIn(categoryIDs...)).
				All(ctx)
			if err == nil {
				for _, cat := range categories {
					managedCategories = append(managedCategories, schema.CategoryBasicInfo{
						ID:   cat.ID,
						Name: cat.Name,
						Slug: cat.Slug,
					})
				}
			}
		}
	}

	// 实时查询用户的发帖数和评论数
	postCount, err := s.GetUserPostCount(ctx, u.ID)
	if err != nil {
		s.logger.Error("查询用户发帖数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		postCount = 0
	}

	commentCount, err := s.GetUserCommentCount(ctx, u.ID)
	if err != nil {
		s.logger.Error("查询用户评论数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		commentCount = 0
	}

	// 构建响应数据
	response := &schema.UserDetailResponse{
		ID:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		Avatar:            u.Avatar,
		Signature:         u.Signature,
		Readme:            u.Readme,
		EmailVerified:     u.EmailVerified,
		Points:            u.Points,
		Currency:          u.Currency,
		PostCount:         postCount,
		CommentCount:      commentCount,
		Status:            u.Status.String(),
		Role:              u.Role.String(),
		CreatedAt:         u.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:         u.UpdatedAt.Format(time_tools.DateTimeFormat),
		ManagedCategories: managedCategories,
	}

	return response, nil
}

// createBalanceLog 创建余额变动记录
func (s *UserManageService) createBalanceLog(ctx context.Context, userID int, logType userbalancelog.Type, amount, beforeAmount, afterAmount int, reason, operatorName string, operatorID *int, relatedID *int, relatedType, ipAddress, userAgent string) error {
	// 创建余额变动记录
	createBuilder := s.db.UserBalanceLog.Create().
		SetUserID(userID).
		SetType(logType).
		SetAmount(amount).
		SetBeforeAmount(beforeAmount).
		SetAfterAmount(afterAmount).
		SetReason(reason).
		SetOperatorName(operatorName).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent)

	// 设置可选字段
	if operatorID != nil {
		createBuilder.SetOperatorID(*operatorID)
	}
	if relatedID != nil {
		createBuilder.SetRelatedID(*relatedID)
	}
	if relatedType != "" {
		createBuilder.SetRelatedType(relatedType)
	}

	_, err := createBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("创建余额变动记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("创建余额变动记录失败: %w", err)
	}

	return nil
}

// GetUserBalanceLog 获取用户余额变动记录
func (s *UserManageService) GetUserBalanceLog(ctx context.Context, req schema.UserBalanceLogRequest) (*schema.UserBalanceLogResponse, error) {
	s.logger.Info("获取用户余额变动记录", zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.UserBalanceLog.Query()

	// 用户ID筛选
	if req.UserID > 0 {
		query = query.Where(userbalancelog.UserID(req.UserID))
	}

	// 变动类型筛选
	if req.Type != "" {
		query = query.Where(userbalancelog.TypeEQ(userbalancelog.Type(req.Type)))
	}

	// 操作者ID筛选
	if req.OperatorID > 0 {
		query = query.Where(userbalancelog.OperatorID(req.OperatorID))
	}

	// 关联业务类型筛选
	if req.RelatedType != "" {
		query = query.Where(userbalancelog.RelatedType(req.RelatedType))
	}

	// 日期范围筛选
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where(userbalancelog.CreatedAtGTE(startTime))
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			endTime = endTime.Add(24 * time.Hour) // 包含整天
			query = query.Where(userbalancelog.CreatedAtLT(endTime))
		}
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取余额变动记录总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取余额变动记录总数失败: %w", err)
	}

	// 分页查询
	logs, err := query.
		Order(ent.Desc(userbalancelog.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取余额变动记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取余额变动记录失败: %w", err)
	}

	// 收集用户ID
	userIDs := make(map[int]bool)
	for _, log := range logs {
		userIDs[log.UserID] = true
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername).
		All(ctx)
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err))
	}
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 转换响应数据
	list := make([]schema.UserBalanceLogItem, len(logs))
	for i, log := range logs {
		item := schema.UserBalanceLogItem{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     userMap[log.UserID],
			Type:         log.Type.String(),
			Amount:       log.Amount,
			BeforeAmount: log.BeforeAmount,
			AfterAmount:  log.AfterAmount,
			Reason:       log.Reason,
			OperatorID:   log.OperatorID,
			OperatorName: log.OperatorName,
			RelatedID:    log.RelatedID,
			RelatedType:  log.RelatedType,
			IPAddress:    log.IPAddress,
			CreatedAt:    log.CreatedAt.Format(time_tools.DateTimeFormat),
		}

		list[i] = item
	}

	return &schema.UserBalanceLogResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserBalanceSummary 获取用户余额汇总信息
func (s *UserManageService) GetUserBalanceSummary(ctx context.Context, userID int) (*schema.UserBalanceSummary, error) {
	s.logger.Info("获取用户余额汇总信息", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 获取用户信息
	u, err := s.db.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 查询积分变动汇总
	pointsIn, err := s.db.UserBalanceLog.Query().
		Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypePoints),
				userbalancelog.AmountGT(0),
			),
		).
		Aggregate(ent.Sum(userbalancelog.FieldAmount)).
		Int(ctx)
	if err != nil {
		pointsIn = 0
	}

	pointsOut, err := s.db.UserBalanceLog.Query().
		Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypePoints),
				userbalancelog.AmountLT(0),
			),
		).
		Aggregate(ent.Sum(userbalancelog.FieldAmount)).
		Int(ctx)
	if err != nil {
		pointsOut = 0
	}

	// 查询货币变动汇总
	currencyIn, err := s.db.UserBalanceLog.Query().
		Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypeCurrency),
				userbalancelog.AmountGT(0),
			),
		).
		Aggregate(ent.Sum(userbalancelog.FieldAmount)).
		Int(ctx)
	if err != nil {
		currencyIn = 0
	}

	currencyOut, err := s.db.UserBalanceLog.Query().
		Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypeCurrency),
				userbalancelog.AmountLT(0),
			),
		).
		Aggregate(ent.Sum(userbalancelog.FieldAmount)).
		Int(ctx)
	if err != nil {
		currencyOut = 0
	}

	// 构建响应数据
	summary := &schema.UserBalanceSummary{
		UserID:           u.ID,
		Username:         u.Username,
		CurrentPoints:    u.Points,
		CurrentCurrency:  u.Currency,
		TotalPointsIn:    pointsIn,
		TotalPointsOut:   -pointsOut, // 转为正数
		TotalCurrencyIn:  currencyIn,
		TotalCurrencyOut: -currencyOut, // 转为正数
	}

	return summary, nil
}

// getUserPostCounts 批量查询用户的发帖数
func (s *UserManageService) getUserPostCounts(ctx context.Context, userIDs []int) map[int]int {
	if len(userIDs) == 0 {
		return make(map[int]int)
	}

	// 使用Ent ORM进行批量查询，避免原生SQL的复杂性
	// 分别查询每个用户的发帖数
	result := make(map[int]int)
	for _, userID := range userIDs {
		count, err := s.db.Post.Query().
			Where(post.UserIDEQ(userID)).
			Count(ctx)
		if err != nil {
			s.logger.Error("查询用户发帖数失败", zap.Int("user_id", userID), zap.Error(err), tracing.WithTraceIDField(ctx))
			// 失败时使用默认值0，继续处理其他用户
			result[userID] = 0
			continue
		}
		result[userID] = count
	}

	return result
}

// getUserCommentCounts 批量查询用户的评论数
func (s *UserManageService) getUserCommentCounts(ctx context.Context, userIDs []int) map[int]int {
	if len(userIDs) == 0 {
		return make(map[int]int)
	}

	// 使用Ent ORM进行批量查询，避免原生SQL的复杂性
	// 分别查询每个用户的评论数
	result := make(map[int]int)
	for _, userID := range userIDs {
		count, err := s.db.Comment.Query().
			Where(comment.UserIDEQ(userID)).
			Count(ctx)
		if err != nil {
			s.logger.Error("查询用户评论数失败", zap.Int("user_id", userID), zap.Error(err), tracing.WithTraceIDField(ctx))
			// 失败时使用默认值0，继续处理其他用户
			result[userID] = 0
			continue
		}
		result[userID] = count
	}

	return result
}

// GetUserPostCount 查询单个用户的发帖数
func (s *UserManageService) GetUserPostCount(ctx context.Context, userID int) (int, error) {
	count, err := s.db.Post.Query().
		Where(post.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询用户发帖数失败: %w", err)
	}
	return count, nil
}

// GetUserCommentCount 查询单个用户的评论数
func (s *UserManageService) GetUserCommentCount(ctx context.Context, userID int) (int, error) {
	count, err := s.db.Comment.Query().
		Where(comment.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询用户评论数失败: %w", err)
	}
	return count, nil
}

// BanUser 封禁用户
func (s *UserManageService) BanUser(ctx context.Context, req schema.UserBanRequest) error {
	s.logger.Info("封禁用户", zap.Int("user_id", req.ID), zap.Int64("duration", req.Duration), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 校验操作权限
	if err := s.checkOperatorPermission(ctx, req.OperatorID, u.Role); err != nil {
		return err
	}

	// 检查用户是否已被永久封禁
	if u.Status == user.StatusBlocked {
		return errors.New("用户已被永久封禁")
	}

	if req.Duration == 0 {
		// 永久封禁：更新数据库状态
		_, err = s.db.User.UpdateOneID(req.ID).
			SetStatus(user.StatusBlocked).
			Save(ctx)
		if err != nil {
			s.logger.Error("永久封禁用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("永久封禁用户失败: %w", err)
		}
		// 同时调用sa-token进行封禁
		if err = stputil.Disable(req.ID, 0); err != nil {
			s.logger.Warn("stputil 永久封禁失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		}
	} else {
		// 短期封禁：使用sa-token
		if err = stputil.Disable(req.ID, time.Duration(req.Duration)*time.Second); err != nil {
			s.logger.Warn("stputil 临时封禁失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		}
	}

	// 踢出用户所有已登录设备
	if err = stputil.Kickout(req.ID); err != nil {
		s.logger.Warn("stputil 踢出设备下线失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}

	s.logger.Info("用户封禁成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// UnbanUser 解封用户
func (s *UserManageService) UnbanUser(ctx context.Context, req schema.UserUnbanRequest) error {
	s.logger.Info("解封用户", zap.Int("user_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.db.User.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 校验操作权限
	if err = s.checkOperatorPermission(ctx, req.OperatorID, u.Role); err != nil {
		return err
	}

	// 如果是永久封禁状态，恢复为正常状态
	if u.Status == user.StatusBlocked {
		_, err = s.db.User.UpdateOneID(req.ID).
			SetStatus(user.StatusNormal).
			Save(ctx)
		if err != nil {
			s.logger.Error("解封用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("解封用户失败: %w", err)
		}
	}

	// 解除sa-token封禁
	if err = stputil.Untie(req.ID); err != nil {
		s.logger.Warn("stputil 移除封禁失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("解封用户失败: %w", err)
	}

	s.logger.Info("用户解封成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}
