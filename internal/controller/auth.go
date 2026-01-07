package controller

import (
	"context"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// AuthController Authentication controller | 认证控制器
type AuthController struct {
	// Injector instance for obtaining services | 注入器实例，用于获取服务
	injector *do.Injector
}

// NewAuthController Create authentication controller instance | 创建认证控制器实例
func NewAuthController(injector *do.Injector) *AuthController {
	return &AuthController{
		injector: injector,
	}
}

// AuthRouter Authentication-related route registration | 认证相关路由注册
func (ctrl *AuthController) AuthRouter(router *gin.RouterGroup) {
	// Register route | 注册路由
	router.POST("/register", ctrl.Register)
	// Login route | 登录路由
	router.POST("/login", ctrl.Login)
	// Logout route | 退出登录
	router.POST("/logout", saGin.CheckLogin(), ctrl.Logout)
	// Send forgot password verification code | 发送找回密码验证码
	router.POST("/forgot-password", ctrl.ForgotPassword)
	// Reset password | 重置密码
	router.POST("/reset-password", ctrl.ResetPassword)
}

// Register User registration endpoint | 用户注册接口
// @Summary User registration | 用户注册
// @Description Create new user account | 创建新用户账户
// @Tags Authentication | 认证
// @Accept json
// @Produce json
// @Param request body schema.RegisterRequest true "Registration information | 注册信息"
// @Success 200 {object} response.Data{data=schema.UserResponse} "Registration successful | 注册成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req schema.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get AuthService from injector | 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service to register | 调用服务进行注册
	user, err := authService.Register(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, schema.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Login User login endpoint | 用户登录接口
// @Summary User login | 用户登录
// @Description User login to obtain authentication information | 用户登录获取认证信息
// @Tags Authentication | 认证
// @Accept json
// @Produce json
// @Param request body schema.LoginRequest true "Login information | 登录信息"
// @Success 200 {object} response.Data{data=schema.LoginResponse} "Login successful | 登录成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get AuthService from injector | 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service to login | 调用服务进行登录
	user, err := authService.Login(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Request UA | 请求UA
	ua := c.GetHeader("User-Agent")
	// Create status Token | 创建状态Token
	token, err := saGin.Login(user.ID, ua)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}
	// Set user identity | 设置用户身份
	if err = stputil.SetRoles(user.ID, satoken.GetUserRole(user.Role.String())); err != nil {
		configs.Log.Warn(err.Error())
	}

	// Record login log - asynchronously save using goroutine | 记录登录日志 - 使用协程异步保存
	go func() {
		// Get client IP address | 获取客户端IP地址
		clientIP := c.ClientIP()
		// Get device information | 获取设备信息
		deviceInfo := c.GetHeader("User-Agent")
		if deviceInfo == "" {
			deviceInfo = "Unknown"
		}

		// Create login record | 创建登录记录
		_, err := configs.DB.UserLoginLog.Create().
			SetUserID(user.ID).
			SetIPAddress(clientIP).
			SetDeviceInfo(deviceInfo).
			SetSuccess(true).
			Save(context.Background())

		if err != nil {
			// Log error without affecting main flow | 记录错误日志，但不影响主流程
			configs.Log.Error("Failed to save login log | 保存登录日志失败",
				zap.Int("user_id", user.ID),
				zap.String("ip_address", clientIP),
				zap.Error(err))
		}
	}()

	// Return success response | 返回成功响应
	response.ResSuccess(c, schema.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	})
}

// Logout User logout endpoint | 用户退出登录接口
// @Summary User logout | 用户退出登录
// @Description User logout and clear authentication information | 用户退出登录，清除认证信息
// @Tags Authentication | 认证
// @Accept json
// @Produce json
// @Success 200 {object} response.Data "Logout successful | 退出登录成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/logout [post]
// @Security Bearer
func (ctrl *AuthController) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")

	// Perform logout operation, clear Token | 执行登出操作，清除 Token
	logoutErr := saGin.LogoutByToken(token)
	if logoutErr != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, logoutErr.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, nil)
}

// ForgotPassword Send forgot password verification code | 发送找回密码验证码
// @Summary Send forgot password verification code | 发送找回密码验证码
// @Description Send forgot password verification code to user email | 向用户邮箱发送找回密码验证码
// @Tags Authentication | 认证
// @Accept json
// @Produce json
// @Param request body schema.ForgotPasswordRequest true "Forgot password request | 找回密码请求"
// @Success 200 {object} response.Data{data=schema.ForgotPasswordResponse} "Send successful | 发送成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 429 {object} response.Data "Too many requests | 发送频率过高"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/forgot-password [post]
func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var req schema.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get AuthService from injector | 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service to send verification code | 调用服务发送验证码
	result, err := authService.SendForgotPasswordCode(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "发送次数过多，请1小时后再试" {
			response.ResErrorWithMsg(c, response.CodeTooManyRequests, err.Error())
		} else {
			response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		}
		return
	}

	response.ResSuccess(c, result)
}

// ResetPassword Reset password | 重置密码
// @Summary Reset password | 重置密码
// @Description Reset user password through verification code | 通过验证码重置用户密码
// @Tags Authentication | 认证
// @Accept json
// @Produce json
// @Param request body schema.ResetPasswordRequest true "Reset password request | 重置密码请求"
// @Success 200 {object} response.Data{data=schema.ResetPasswordResponse} "Reset successful | 重置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/reset-password [post]
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req schema.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get AuthService from injector | 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service to reset password | 调用服务重置密码
	result, err := authService.ResetPassword(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
