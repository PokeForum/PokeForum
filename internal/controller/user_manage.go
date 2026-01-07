package controller

import (
	"fmt"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// UserManageController User management controller | 用户管理控制器
type UserManageController struct {
	// Injector instance for obtaining services | 注入器实例，用于获取服务
	injector *do.Injector
}

// NewUserManageController Create user management controller instance | 创建用户管理控制器实例
func NewUserManageController(injector *do.Injector) *UserManageController {
	return &UserManageController{
		injector: injector,
	}
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *UserManageController) getUserID(c *gin.Context) (int, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("未找到Authorization header")
	}

	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return sID, nil
}

// UserManageRouter User management related route registration | 用户管理相关路由注册
func (ctrl *UserManageController) UserManageRouter(router *gin.RouterGroup) {
	// User list | 用户列表
	router.GET("", ctrl.GetUserList)
	// Create user | 创建用户
	router.POST("", ctrl.CreateUser)
	// Update user information | 更新用户信息
	router.PUT("", ctrl.UpdateUser)
	// Get user details | 获取用户详情
	router.GET("/:id", ctrl.GetUserDetail)

	// User status management | 用户状态管理
	router.PUT("/status", ctrl.UpdateUserStatus)

	// User role management | 用户身份管理
	router.PUT("/role", saGin.CheckRole(user.RoleSuperAdmin.String()), ctrl.UpdateUserRole)

	// User points management | 用户积分管理
	router.PUT("/points", ctrl.UpdateUserPoints)

	// User currency management | 用户货币管理
	router.PUT("/currency", ctrl.UpdateUserCurrency)

	// Moderator category management | 版主管理版块
	router.PUT("/moderator/categories", ctrl.SetModeratorCategories)

	// Balance change logs | 余额变动记录
	router.GET("/balance/logs", ctrl.GetUserBalanceLog)
	router.GET("/balance/summary/:id", ctrl.GetUserBalanceSummary)

	// User ban/unban | 用户封禁/解封
	router.POST("/ban", ctrl.BanUser)
	router.POST("/unban", ctrl.UnbanUser)
}

// GetUserList Get user list | 获取用户列表
// @Summary Get user list | 获取用户列表
// @Description Get paginated user list with keyword search and status filtering support | 分页获取用户列表，支持关键词搜索和状态筛选
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param keyword query string false "Search keyword | 搜索关键词" example("test")
// @Param status query string false "User status | 用户状态" example("Normal")
// @Param role query string false "User role | 用户身份" example("User")
// @Success 200 {object} response.Data{data=schema.UserListResponse} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users [get]
func (ctrl *UserManageController) GetUserList(c *gin.Context) {
	var req schema.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := userManageService.GetUserList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateUser Create user | 创建用户
// @Summary Create user | 创建用户
// @Description Administrator creates new user account | 管理员创建新用户账户
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserCreateRequest true "User information | 用户信息"
// @Success 200 {object} response.Data{data=schema.UserDetailResponse} "Creation successful | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users [post]
func (ctrl *UserManageController) CreateUser(c *gin.Context) {
	var req schema.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	u, err := userManageService.CreateUser(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Query user's post count in real-time | 实时查询用户的发帖数和评论数
	postCount, err := userManageService.GetUserPostCount(c.Request.Context(), u.ID)
	if err != nil {
		// Post count should be 0 for newly created user, use default value on query failure | 新创建用户的发帖数应该是0，查询失败时使用默认值
		postCount = 0
	}

	commentCount, err := userManageService.GetUserCommentCount(c.Request.Context(), u.ID)
	if err != nil {
		// Comment count should be 0 for newly created user, use default value on query failure | 新创建用户的评论数应该是0，查询失败时使用默认值
		commentCount = 0
	}

	// Convert to response format | 转换为响应格式
	result := &schema.UserDetailResponse{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Signature:     u.Signature,
		Readme:        u.Readme,
		EmailVerified: u.EmailVerified,
		Points:        u.Points,
		Currency:      u.Currency,
		PostCount:     postCount,
		CommentCount:  commentCount,
		Status:        u.Status.String(),
		Role:          u.Role.String(),
		CreatedAt:     u.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:     u.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// UpdateUser Update user information | 更新用户信息
// @Summary Update user information | 更新用户信息
// @Description Update user's basic information | 更新用户的基本信息
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserUpdateRequest true "User information | 用户信息"
// @Success 200 {object} response.Data{data=schema.UserDetailResponse} "Update successful | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users [put]
func (ctrl *UserManageController) UpdateUser(c *gin.Context) {
	var req schema.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	u, err := userManageService.UpdateUser(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Query user's post count in real-time | 实时查询用户的发帖数和评论数
	postCount, err := userManageService.GetUserPostCount(c.Request.Context(), u.ID)
	if err != nil {
		// Use default value 0 on query failure | 查询失败时使用默认值0
		postCount = 0
	}

	commentCount, err := userManageService.GetUserCommentCount(c.Request.Context(), u.ID)
	if err != nil {
		// Use default value 0 on query failure | 查询失败时使用默认值0
		commentCount = 0
	}

	// Convert to response format | 转换为响应格式
	result := &schema.UserDetailResponse{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Signature:     u.Signature,
		Readme:        u.Readme,
		EmailVerified: u.EmailVerified,
		Points:        u.Points,
		Currency:      u.Currency,
		PostCount:     postCount,
		CommentCount:  commentCount,
		Status:        u.Status.String(),
		Role:          u.Role.String(),
		CreatedAt:     u.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:     u.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// GetUserDetail Get user details | 获取用户详情
// @Summary Get user details | 获取用户详情
// @Description Get detailed information for specified user | 获取指定用户的详细信息
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param id path int true "User ID | 用户ID" example("1")
// @Success 200 {object} response.Data{data=schema.UserDetailResponse} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/{id} [get]
func (ctrl *UserManageController) GetUserDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := userManageService.GetUserDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// UpdateUserStatus Update user status | 更新用户状态
// @Summary Update user status | 更新用户状态
// @Description Update user's status (Normal, Muted, Banned, etc.) | 更新用户的状态（正常、禁言、封禁等）
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserStatusUpdateRequest true "Status information | 状态信息"
// @Success 200 {object} response.Data "Update successful | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/status [put]
func (ctrl *UserManageController) UpdateUserStatus(c *gin.Context) {
	var req schema.UserStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get operator ID | 获取操作者ID
	operatorID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}
	req.OperatorID = operatorID

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	err = userManageService.UpdateUserStatus(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateUserRole Update user role | 更新用户身份
// @Summary Update user role | 更新用户身份
// @Description Update user's role permissions (Regular User, Moderator, Admin, etc.) | 更新用户的身份权限（普通用户、版主、管理员等）
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserRoleUpdateRequest true "Role information | 身份信息"
// @Success 200 {object} response.Data "Update successful | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/role [put]
func (ctrl *UserManageController) UpdateUserRole(c *gin.Context) {
	var req schema.UserRoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	err = userManageService.UpdateUserRole(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateUserPoints Update user points | 更新用户积分
// @Summary Update user points | 更新用户积分
// @Description Add or deduct points for user | 为用户增加或减少积分
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserPointsUpdateRequest true "Points information | 积分信息"
// @Success 200 {object} response.Data "Update successful | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/points [put]
func (ctrl *UserManageController) UpdateUserPoints(c *gin.Context) {
	var req schema.UserPointsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	err = userManageService.UpdateUserPoints(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateUserCurrency Update user currency | 更新用户货币
// @Summary Update user currency | 更新用户货币
// @Description Add or deduct currency for user | 为用户增加或减少货币
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserCurrencyUpdateRequest true "Currency information | 货币信息"
// @Success 200 {object} response.Data "Update successful | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/currency [put]
func (ctrl *UserManageController) UpdateUserCurrency(c *gin.Context) {
	var req schema.UserCurrencyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	err = userManageService.UpdateUserCurrency(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetModeratorCategories Set moderator categories | 设置版主管理版块
// @Summary Set moderator categories | 设置版主管理版块
// @Description Set the list of categories managed by specified moderator | 为指定版主设置其管理的版块列表
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.ModeratorCategoryRequest true "Category information | 版块信息"
// @Success 200 {object} response.Data "Setting successful | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/moderator/categories [put]
func (ctrl *UserManageController) SetModeratorCategories(c *gin.Context) {
	var req schema.ModeratorCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	err = userManageService.SetModeratorCategories(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetUserBalanceLog Get user balance change logs | 获取用户余额变动记录
// @Summary Get user balance change logs | 获取用户余额变动记录
// @Description Get paginated user balance change logs with various filtering options | 分页获取用户余额变动记录，支持多种筛选条件
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param user_id query int false "User ID filter | 用户ID筛选" example("1")
// @Param type query string false "Change type filter | 变动类型筛选" example("points")
// @Param start_date query string false "Start date | 开始日期" example("2024-01-01")
// @Param end_date query string false "End date | 结束日期" example("2024-12-31")
// @Param operator_id query int false "Operator ID filter | 操作者ID筛选" example("2")
// @Param related_type query string false "Related business type filter | 关联业务类型筛选" example("post")
// @Success 200 {object} response.Data{data=schema.UserBalanceLogResponse} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/balance/logs [get]
func (ctrl *UserManageController) GetUserBalanceLog(c *gin.Context) {
	var req schema.UserBalanceLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := userManageService.GetUserBalanceLog(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetUserBalanceSummary Get user balance summary information | 获取用户余额汇总信息
// @Summary Get user balance summary information | 获取用户余额汇总信息
// @Description Get balance summary statistics for specified user | 获取指定用户的余额汇总统计信息
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param id path int true "User ID | 用户ID" example("1")
// @Success 200 {object} response.Data{data=schema.UserBalanceSummary} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/balance/summary/{id} [get]
func (ctrl *UserManageController) GetUserBalanceSummary(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := userManageService.GetUserBalanceSummary(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// BanUser Ban user | 封禁用户
// @Summary Ban user | 封禁用户
// @Description Ban specified user, supports temporary and permanent bans | 封禁指定用户，支持短期封禁和永久封禁
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserBanRequest true "Ban information | 封禁信息"
// @Success 200 {object} response.Data "Ban successful | 封禁成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/ban [post]
func (ctrl *UserManageController) BanUser(c *gin.Context) {
	var req schema.UserBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get operator ID | 获取操作者ID
	operatorID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}
	req.OperatorID = operatorID

	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	err = userManageService.BanUser(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UnbanUser Unban user | 解封用户
// @Summary Unban user | 解封用户
// @Description Remove ban status from specified user | 解除指定用户的封禁状态
// @Tags [Admin]User Management | [管理员]用户管理
// @Accept json
// @Produce json
// @Param request body schema.UserUnbanRequest true "Unban information | 解封信息"
// @Success 200 {object} response.Data "Unban successful | 解封成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/users/unban [post]
func (ctrl *UserManageController) UnbanUser(c *gin.Context) {
	var req schema.UserUnbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get operator ID | 获取操作者ID
	operatorID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}
	req.OperatorID = operatorID

	userManageService, err := do.Invoke[service.IUserManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	err = userManageService.UnbanUser(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
