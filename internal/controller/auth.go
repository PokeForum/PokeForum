package controller

import (
	"context"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"go.uber.org/zap"
)

// AuthController 认证控制器
type AuthController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewAuthController 创建认证控制器实例
func NewAuthController(injector *do.Injector) *AuthController {
	return &AuthController{
		injector: injector,
	}
}

// AuthRouter 认证相关路由注册
func (ctrl *AuthController) AuthRouter(router *gin.RouterGroup) {
	// 注册路由
	router.POST("/register", ctrl.Register)
	// 登录路由
	router.POST("/login", ctrl.Login)
	// 退出登录
	router.POST("/logout", saGin.CheckLogin(), ctrl.Logout)
}

// Register 用户注册接口
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body schema.RegisterRequest true "注册信息"
// @Success 200 {object} response.Data{data=schema.UserResponse} "注册成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req schema.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务进行注册
	user, err := authService.Register(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, schema.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Login 用户登录接口
// @Summary 用户登录
// @Description 用户登录获取认证信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body schema.LoginRequest true "登录信息"
// @Success 200 {object} response.Data{data=schema.LoginResponse} "登录成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 从注入器获取 AuthService
	authService, err := do.Invoke[service.IAuthService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务进行登录
	user, err := authService.Login(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 请求UA
	ua := c.GetHeader("User-Agent")
	// 创建状态Token
	token, err := saGin.Login(user.ID, ua)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}
	// 设置用户身份
	if err = stputil.SetRoles(user.ID, satoken.GetUserRole(user.Role.String())); err != nil {
		configs.Log.Warn(err.Error())
	}

	// 记录登录日志 - 使用协程异步保存
	go func() {
		// 获取客户端IP地址
		clientIP := c.ClientIP()
		// 获取设备信息
		deviceInfo := c.GetHeader("User-Agent")
		if deviceInfo == "" {
			deviceInfo = "Unknown"
		}

		// 创建登录记录
		_, err := configs.DB.UserLoginLog.Create().
			SetUserID(user.ID).
			SetIPAddress(clientIP).
			SetDeviceInfo(deviceInfo).
			SetSuccess(true).
			Save(context.Background())

		if err != nil {
			// 记录错误日志，但不影响主流程
			configs.Log.Error("保存登录日志失败",
				zap.Int("user_id", user.ID),
				zap.String("ip_address", clientIP),
				zap.Error(err))
		}
	}()

	// 返回成功响应
	response.ResSuccess(c, schema.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	})
}

// Logout 用户退出登录接口
// @Summary 用户退出登录
// @Description 用户退出登录，清除认证信息
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} response.Data "退出登录成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /auth/logout [post]
// @Security Bearer
func (ctrl *AuthController) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")

	// 执行登出操作，清除 Token
	logoutErr := saGin.LogoutByToken(token)
	if logoutErr != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, logoutErr.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, nil)
}
