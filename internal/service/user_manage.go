package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/stputil"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/categorymoderator"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/ent/userbalancelog"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// IUserManageService User management service interface | 用户管理服务接口
type IUserManageService interface {
	// GetUserList Get user list | 获取用户列表
	GetUserList(ctx context.Context, req schema.UserListRequest) (*schema.UserListResponse, error)
	// CreateUser Create user | 创建用户
	CreateUser(ctx context.Context, req schema.UserCreateRequest) (*ent.User, error)
	// UpdateUser Update user information | 更新用户信息
	UpdateUser(ctx context.Context, req schema.UserUpdateRequest) (*ent.User, error)
	// UpdateUserStatus Update user status | 更新用户状态
	UpdateUserStatus(ctx context.Context, req schema.UserStatusUpdateRequest) error
	// UpdateUserRole Update user role | 更新用户身份
	UpdateUserRole(ctx context.Context, req schema.UserRoleUpdateRequest) error
	// UpdateUserPoints Update user points | 更新用户积分
	UpdateUserPoints(ctx context.Context, req schema.UserPointsUpdateRequest) error
	// UpdateUserCurrency Update user currency | 更新用户货币
	UpdateUserCurrency(ctx context.Context, req schema.UserCurrencyUpdateRequest) error
	// SetModeratorCategories Set moderator categories | 设置版主管理版块
	SetModeratorCategories(ctx context.Context, req schema.ModeratorCategoryRequest) error
	// GetUserDetail Get user details | 获取用户详情
	GetUserDetail(ctx context.Context, id int) (*schema.UserDetailResponse, error)
	// GetUserBalanceLog Get user balance change log | 获取用户余额变动记录
	GetUserBalanceLog(ctx context.Context, req schema.UserBalanceLogRequest) (*schema.UserBalanceLogResponse, error)
	// GetUserBalanceSummary Get user balance summary | 获取用户余额汇总信息
	GetUserBalanceSummary(ctx context.Context, userID int) (*schema.UserBalanceSummary, error)
	// GetUserPostCount Query single user's post count | 查询单个用户的发帖数
	GetUserPostCount(ctx context.Context, userID int) (int, error)
	// GetUserCommentCount Query single user's comment count | 查询单个用户的评论数
	GetUserCommentCount(ctx context.Context, userID int) (int, error)
	// BanUser Ban user | 封禁用户
	BanUser(ctx context.Context, req schema.UserBanRequest) error
	// UnbanUser Unban user | 解封用户
	UnbanUser(ctx context.Context, req schema.UserUnbanRequest) error
}

// UserManageService User management service implementation | 用户管理服务实现
type UserManageService struct {
	db                    *ent.Client
	userRepo              repository.IUserRepository
	postRepo              repository.IPostRepository
	commentRepo           repository.ICommentRepository
	categoryRepo          repository.ICategoryRepository
	categoryModeratorRepo repository.ICategoryModeratorRepository
	userBalanceLogRepo    repository.IUserBalanceLogRepository
	cache                 cache.ICacheService
	logger                *zap.Logger
}

// NewUserManageService Create user management service instance | 创建用户管理服务实例
func NewUserManageService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) IUserManageService {
	return &UserManageService{
		db:                    db,
		userRepo:              repos.User,
		postRepo:              repos.Post,
		commentRepo:           repos.Comment,
		categoryRepo:          repos.Category,
		categoryModeratorRepo: repos.CategoryModerator,
		userBalanceLogRepo:    repos.UserBalanceLog,
		cache:                 cacheService,
		logger:                logger,
	}
}

// checkOperatorPermission Validate operator permission | 校验操作者权限
// Admins can only operate on users and moderators, super admins can operate on all roles | 管理员只能操作用户和版主，超级管理员可以操作所有身份
func (s *UserManageService) checkOperatorPermission(ctx context.Context, operatorID int, targetRole user.Role) error {
	// 获取操作者信息
	operator, err := s.userRepo.GetByID(ctx, operatorID)
	if err != nil {
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

// GetUserList Get user list | 获取用户列表
func (s *UserManageService) GetUserList(ctx context.Context, req schema.UserListRequest) (*schema.UserListResponse, error) {
	s.logger.Info("获取用户列表", tracing.WithTraceIDField(ctx))

	// 使用 Repository 进行条件查询
	var users []*ent.User
	var total int
	var err error

	// 构建查询条件函数
	conditionFunc := func(q *ent.UserQuery) *ent.UserQuery {
		// 关键词搜索
		if req.Keyword != "" {
			q = q.Where(
				user.Or(
					user.UsernameContains(req.Keyword),
					user.EmailContains(req.Keyword),
				),
			)
		}

		// 状态筛选
		if req.Status != "" {
			q = q.Where(user.StatusEQ(user.Status(req.Status)))
		}

		// 身份筛选
		if req.Role != "" {
			q = q.Where(user.RoleEQ(user.Role(req.Role)))
		}

		// 排序和分页
		q = q.Order(ent.Desc(user.FieldCreatedAt)).
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize)

		return q
	}

	// 先获取总数
	totalCountFunc := func(q *ent.UserQuery) *ent.UserQuery {
		// 关键词搜索
		if req.Keyword != "" {
			q = q.Where(
				user.Or(
					user.UsernameContains(req.Keyword),
					user.EmailContains(req.Keyword),
				),
			)
		}

		// 状态筛选
		if req.Status != "" {
			q = q.Where(user.StatusEQ(user.Status(req.Status)))
		}

		// 身份筛选
		if req.Role != "" {
			q = q.Where(user.RoleEQ(user.Role(req.Role)))
		}

		return q
	}

	total, err = s.userRepo.CountWithCondition(ctx, totalCountFunc)
	if err != nil {
		s.logger.Error("获取用户总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 分页查询
	users, err = s.userRepo.ListWithCondition(ctx, conditionFunc, 0)
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

// CreateUser Create user | 创建用户
func (s *UserManageService) CreateUser(ctx context.Context, req schema.UserCreateRequest) (*ent.User, error) {
	s.logger.Info("创建用户", zap.String("username", req.Username), tracing.WithTraceIDField(ctx))

	// 检查用户名是否已存在
	usernameExists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("检查用户名是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户名是否存在失败: %w", err)
	}
	if usernameExists {
		return nil, errors.New("用户名已存在")
	}

	emailExists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("检查邮箱是否存在失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查邮箱是否存在失败: %w", err)
	}
	if emailExists {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	u, err := s.userRepo.CreateWithBuilder(ctx, func(c *ent.UserCreate) *ent.UserCreate {
		return c.SetUsername(req.Username).
			SetEmail(req.Email).
			SetPassword(hashedPassword).
			SetRole(user.Role(req.Role)).
			SetAvatar(req.Avatar).
			SetSignature(req.Signature).
			SetReadme(req.Readme)
	})
	if err != nil {
		s.logger.Error("创建用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	s.logger.Info("用户创建成功", zap.Int("user_id", u.ID), tracing.WithTraceIDField(ctx))
	return u, nil
}

// UpdateUser Update user information | 更新用户信息
func (s *UserManageService) UpdateUser(ctx context.Context, req schema.UserUpdateRequest) (*ent.User, error) {
	s.logger.Info("更新用户信息", zap.Int("user_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户名和邮箱是否与其他用户冲突
	if req.Username != "" {
		usernameExists, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
			return q.Where(user.And(user.IDNEQ(req.ID), user.UsernameEQ(req.Username)))
		})
		if err != nil {
			s.logger.Error("检查用户名冲突失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查用户名冲突失败: %w", err)
		}
		if usernameExists > 0 {
			return nil, errors.New("用户名已被其他用户使用")
		}
	}

	if req.Email != "" {
		emailExists, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
			return q.Where(user.And(user.IDNEQ(req.ID), user.EmailEQ(req.Email)))
		})
		if err != nil {
			s.logger.Error("检查邮箱冲突失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查邮箱冲突失败: %w", err)
		}
		if emailExists > 0 {
			return nil, errors.New("邮箱已被其他用户使用")
		}
	}

	// 构建更新操作
	u, err := s.userRepo.Update(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
		if req.Username != "" {
			u = u.SetUsername(req.Username)
		}
		if req.Email != "" {
			u = u.SetEmail(req.Email)
		}
		if req.Avatar != "" {
			u = u.SetAvatar(req.Avatar)
		}
		if req.Signature != "" {
			u = u.SetSignature(req.Signature)
		}
		if req.Readme != "" {
			u = u.SetReadme(req.Readme)
		}
		return u
	})
	if err != nil {
		s.logger.Error("更新用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	s.logger.Info("用户信息更新成功", zap.Int("user_id", u.ID), tracing.WithTraceIDField(ctx))
	return u, nil
}

// UpdateUserStatus Update user status | 更新用户状态
func (s *UserManageService) UpdateUserStatus(ctx context.Context, req schema.UserStatusUpdateRequest) error {
	s.logger.Info("更新用户状态", zap.Int("user_id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 校验操作权限
	if err = s.checkOperatorPermission(ctx, req.OperatorID, u.Role); err != nil {
		return err
	}

	// 更新用户状态
	_, err = s.userRepo.Update(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
		return u.SetStatus(user.Status(req.Status))
	})
	if err != nil {
		s.logger.Error("更新用户状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	s.logger.Info("用户状态更新成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// UpdateUserRole Update user role | 更新用户身份
func (s *UserManageService) UpdateUserRole(ctx context.Context, req schema.UserRoleUpdateRequest) error {
	s.logger.Info("更新用户身份", zap.Int("user_id", req.ID), zap.String("role", req.Role), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 更新用户身份
	_, err = s.userRepo.Update(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
		return u.SetRole(user.Role(req.Role))
	})
	if err != nil {
		s.logger.Error("更新用户身份失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新用户身份失败: %w", err)
	}

	// 如果用户不再是版主，清除其管理的版块（通过中间表）
	if user.Role(req.Role) != user.RoleModerator {
		err = s.categoryModeratorRepo.DeleteByUserID(ctx, req.ID)
		if err != nil {
			s.logger.Error("清除用户管理版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return fmt.Errorf("清除用户管理版块失败: %w", err)
		}
	}

	s.logger.Info("用户身份更新成功", zap.Int("user_id", req.ID), zap.String("reason", req.Reason), tracing.WithTraceIDField(ctx))
	return nil
}

// UpdateUserPoints Update user points | 更新用户积分
func (s *UserManageService) UpdateUserPoints(ctx context.Context, req schema.UserPointsUpdateRequest) error {
	s.logger.Info("更新用户积分", zap.Int("user_id", req.ID), zap.Int("points", req.Points), tracing.WithTraceIDField(ctx))

	// 获取用户当前积分
	u, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 计算新积分
	newPoints := u.Points + req.Points
	if newPoints < 0 {
		return errors.New("积分不能为负数")
	}

	// 更新用户积分
	err = s.userRepo.UpdateFields(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
		return u.SetPoints(newPoints)
	})
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

// UpdateUserCurrency Update user currency | 更新用户货币
func (s *UserManageService) UpdateUserCurrency(ctx context.Context, req schema.UserCurrencyUpdateRequest) error {
	s.logger.Info("更新用户货币", zap.Int("user_id", req.ID), zap.Int("currency", req.Currency), tracing.WithTraceIDField(ctx))

	// 获取用户当前货币
	u, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 计算新货币
	newCurrency := u.Currency + req.Currency
	if newCurrency < 0 {
		return errors.New("货币不能为负数")
	}

	// 更新用户货币
	err = s.userRepo.UpdateFields(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
		return u.SetCurrency(newCurrency)
	})
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

// SetModeratorCategories Set moderator categories | 设置版主管理版块
func (s *UserManageService) SetModeratorCategories(ctx context.Context, req schema.ModeratorCategoryRequest) error {
	s.logger.Info("设置版主管理版块", zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在且是版主
	existingUser, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if existingUser.Role != user.RoleModerator {
		return errors.New("只有版主才能设置管理版块")
	}

	// 检查版块是否存在
	categories, err := s.categoryRepo.GetByIDs(ctx, req.CategoryIDs)
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

// GetUserDetail Get user details | 获取用户详情
func (s *UserManageService) GetUserDetail(ctx context.Context, id int) (*schema.UserDetailResponse, error) {
	s.logger.Info("获取用户详情", zap.Int("user_id", id), tracing.WithTraceIDField(ctx))

	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取用户详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户详情失败: %w", err)
	}

	// 通过中间表查询管理的版块
	managedCategories := make([]schema.CategoryBasicInfo, 0)
	if u.Role == user.RoleModerator {
		// 查询版主关联记录
		moderatorRecords, err := s.categoryModeratorRepo.GetByUserID(ctx, id)
		if err == nil && len(moderatorRecords) > 0 {
			// 收集版块ID
			categoryIDs := make([]int, len(moderatorRecords))
			for i, record := range moderatorRecords {
				categoryIDs[i] = record.CategoryID
			}

			// 批量查询版块信息
			categories, err := s.categoryRepo.GetByIDs(ctx, categoryIDs)
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

// createBalanceLog Create balance change log | 创建余额变动记录
func (s *UserManageService) createBalanceLog(ctx context.Context, userID int, logType userbalancelog.Type, amount, beforeAmount, afterAmount int, reason, operatorName string, operatorID *int, relatedID *int, relatedType, ipAddress, userAgent string) error {
	// 创建余额变动记录
	_, err := s.userBalanceLogRepo.Create(ctx, func(c *ent.UserBalanceLogCreate) *ent.UserBalanceLogCreate {
		c = c.SetUserID(userID).
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
			c = c.SetOperatorID(*operatorID)
		}
		if relatedID != nil {
			c = c.SetRelatedID(*relatedID)
		}
		if relatedType != "" {
			c = c.SetRelatedType(relatedType)
		}
		return c
	})
	if err != nil {
		s.logger.Error("创建余额变动记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("创建余额变动记录失败: %w", err)
	}

	return nil
}

// GetUserBalanceLog Get user balance change log | 获取用户余额变动记录
func (s *UserManageService) GetUserBalanceLog(ctx context.Context, req schema.UserBalanceLogRequest) (*schema.UserBalanceLogResponse, error) {
	s.logger.Info("获取用户余额变动记录", zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 构建查询条件函数
	conditionFunc := func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		// 用户ID筛选
		if req.UserID > 0 {
			q = q.Where(userbalancelog.UserID(req.UserID))
		}

		// 变动类型筛选
		if req.Type != "" {
			q = q.Where(userbalancelog.TypeEQ(userbalancelog.Type(req.Type)))
		}

		// 操作者ID筛选
		if req.OperatorID > 0 {
			q = q.Where(userbalancelog.OperatorID(req.OperatorID))
		}

		// 关联业务类型筛选
		if req.RelatedType != "" {
			q = q.Where(userbalancelog.RelatedType(req.RelatedType))
		}

		// 日期范围筛选
		if req.StartDate != "" {
			startTime, err := time.Parse("2006-01-02", req.StartDate)
			if err == nil {
				q = q.Where(userbalancelog.CreatedAtGTE(startTime))
			}
		}
		if req.EndDate != "" {
			endTime, err := time.Parse("2006-01-02", req.EndDate)
			if err == nil {
				endTime = endTime.Add(24 * time.Hour) // 包含整天
				q = q.Where(userbalancelog.CreatedAtLT(endTime))
			}
		}
		return q
	}

	// 获取总数
	total, err := s.userBalanceLogRepo.Count(ctx, conditionFunc)
	if err != nil {
		s.logger.Error("获取余额变动记录总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取余额变动记录总数失败: %w", err)
	}

	// 分页查询
	logs, err := s.userBalanceLogRepo.Query(ctx, func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		q = conditionFunc(q)
		return q.Order(ent.Desc(userbalancelog.FieldCreatedAt)).
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize)
	})
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
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
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

// GetUserBalanceSummary Get user balance summary | 获取用户余额汇总信息
func (s *UserManageService) GetUserBalanceSummary(ctx context.Context, userID int) (*schema.UserBalanceSummary, error) {
	s.logger.Info("获取用户余额汇总信息", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 查询积分变动汇总
	pointsIn, err := s.userBalanceLogRepo.AggregateSum(ctx, func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		return q.Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypePoints),
				userbalancelog.AmountGT(0),
			),
		)
	})
	if err != nil {
		pointsIn = 0
	}

	pointsOut, err := s.userBalanceLogRepo.AggregateSum(ctx, func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		return q.Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypePoints),
				userbalancelog.AmountLT(0),
			),
		)
	})
	if err != nil {
		pointsOut = 0
	}

	// 查询货币变动汇总
	currencyIn, err := s.userBalanceLogRepo.AggregateSum(ctx, func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		return q.Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypeCurrency),
				userbalancelog.AmountGT(0),
			),
		)
	})
	if err != nil {
		currencyIn = 0
	}

	currencyOut, err := s.userBalanceLogRepo.AggregateSum(ctx, func(q *ent.UserBalanceLogQuery) *ent.UserBalanceLogQuery {
		return q.Where(
			userbalancelog.And(
				userbalancelog.UserID(userID),
				userbalancelog.TypeEQ(userbalancelog.TypeCurrency),
				userbalancelog.AmountLT(0),
			),
		)
	})
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

// getUserPostCounts Batch query users' post counts | 批量查询用户的发帖数
func (s *UserManageService) getUserPostCounts(ctx context.Context, userIDs []int) map[int]int {
	if len(userIDs) == 0 {
		return make(map[int]int)
	}

	// 使用 Repository 进行批量查询
	result := make(map[int]int)
	for _, userID := range userIDs {
		count, err := s.postRepo.CountByUserID(ctx, userID)
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

// getUserCommentCounts Batch query users' comment counts | 批量查询用户的评论数
func (s *UserManageService) getUserCommentCounts(ctx context.Context, userIDs []int) map[int]int {
	if len(userIDs) == 0 {
		return make(map[int]int)
	}

	// 使用 Repository 进行批量查询
	result := make(map[int]int)
	for _, userID := range userIDs {
		count, err := s.commentRepo.CountByUserID(ctx, userID)
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

// GetUserPostCount Query single user's post count | 查询单个用户的发帖数
func (s *UserManageService) GetUserPostCount(ctx context.Context, userID int) (int, error) {
	count, err := s.postRepo.CountByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("查询用户发帖数失败: %w", err)
	}
	return count, nil
}

// GetUserCommentCount Query single user's comment count | 查询单个用户的评论数
func (s *UserManageService) GetUserCommentCount(ctx context.Context, userID int) (int, error) {
	count, err := s.commentRepo.CountByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("查询用户评论数失败: %w", err)
	}
	return count, nil
}

// BanUser Ban user | 封禁用户
func (s *UserManageService) BanUser(ctx context.Context, req schema.UserBanRequest) error {
	s.logger.Info("封禁用户", zap.Int("user_id", req.ID), zap.Int64("duration", req.Duration), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
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
		_, err = s.userRepo.Update(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
			return u.SetStatus(user.StatusBlocked)
		})
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

// UnbanUser Unban user | 解封用户
func (s *UserManageService) UnbanUser(ctx context.Context, req schema.UserUnbanRequest) error {
	s.logger.Info("解封用户", zap.Int("user_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	u, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 校验操作权限
	if err = s.checkOperatorPermission(ctx, req.OperatorID, u.Role); err != nil {
		return err
	}

	// 如果是永久封禁状态，恢复为正常状态
	if u.Status == user.StatusBlocked {
		_, err = s.userRepo.Update(ctx, req.ID, func(u *ent.UserUpdateOne) *ent.UserUpdateOne {
			return u.SetStatus(user.StatusNormal)
		})
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
