package controller

import (
	"fmt"
	"strconv"

	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// OAuthController OAuth login controller | OAuth登录控制器
type OAuthController struct {
	oauthService service.IOAuthService
}

// NewOAuthController Create OAuth login controller instance | 创建OAuth登录控制器实例
func NewOAuthController(oauthService service.IOAuthService) *OAuthController {
	return &OAuthController{
		oauthService: oauthService,
	}
}

// OAuthPublicRouter OAuth public routes (no login required) | OAuth公开路由（无需登录）
func (ctrl *OAuthController) OAuthPublicRouter(router *gin.RouterGroup) {
	// Get enabled OAuth providers list | 获取已启用的OAuth提供商列表
	router.GET("/providers", ctrl.GetEnabledProviders)
	// Get OAuth authorization URL | 获取OAuth授权URL
	router.GET("/:provider/authorize", ctrl.GetAuthorizeURL)
	// OAuth callback (login/register) | OAuth回调（登录/注册）
	router.POST("/:provider/callback", ctrl.HandleCallback)
}

// OAuthUserRouter OAuth user routes | OAuth用户路由
func (ctrl *OAuthController) OAuthUserRouter(router *gin.RouterGroup) {
	// Get user OAuth binding list | 获取用户OAuth绑定列表
	router.GET("/bindlist", ctrl.GetUserBindList)
	// Get OAuth bind authorization URL | 获取OAuth绑定授权URL
	router.GET("/:provider/bindurl", ctrl.GetBindURL)
	// OAuth bind callback | OAuth绑定回调
	router.POST("/:provider/bindcallback", ctrl.HandleBindCallback)
	// Unbind OAuth | 解绑OAuth
	router.DELETE("/:provider", ctrl.Unbind)
}

// GetEnabledProviders Get enabled OAuth providers list | 获取已启用的OAuth提供商列表
// @Summary Get enabled OAuth providers list | 获取已启用的OAuth提供商列表
// @Description Get all enabled OAuth providers for frontend display | 获取所有已启用的OAuth提供商，用于前端展示
// @Tags OAuth | OAuth登录
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.OAuthProviderPublicListResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/oauth/providers [get]
func (ctrl *OAuthController) GetEnabledProviders(c *gin.Context) {
	result, err := ctrl.oauthService.GetEnabledProviders(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetAuthorizeURL Get OAuth authorization URL | 获取OAuth授权URL
// @Summary Get OAuth authorization URL | 获取OAuth授权URL
// @Description Get authorization URL for specified OAuth provider | 获取指定OAuth提供商的授权URL
// @Tags OAuth | OAuth登录
// @Accept json
// @Produce json
// @Param provider path string true "Provider type | 提供商类型" Enums(QQ, GitHub, Google)
// @Param redirect_uri query string true "Frontend callback URL | 前端回调地址"
// @Success 200 {object} response.Data{data=schema.OAuthAuthorizeResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/oauth/{provider}/authorize [get]
func (ctrl *OAuthController) GetAuthorizeURL(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "provider参数不能为空")
		return
	}

	var req schema.OAuthAuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	clientIP := c.ClientIP()
	result, err := ctrl.oauthService.GetAuthorizeURL(c.Request.Context(), provider, req, clientIP)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// HandleCallback Handle OAuth callback (login/register) | 处理OAuth回调（登录/注册）
// @Summary Handle OAuth callback | 处理OAuth回调
// @Description Handle OAuth callback, auto login or register | 处理OAuth回调，自动登录或注册
// @Tags OAuth | OAuth登录
// @Accept json
// @Produce json
// @Param provider path string true "Provider type | 提供商类型" Enums(QQ, GitHub, Google)
// @Param request body schema.OAuthCallbackRequest true "OAuth callback parameters | OAuth回调参数"
// @Success 200 {object} response.Data{data=schema.OAuthCallbackResponse} "Success | 处理成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /auth/oauth/{provider}/callback [post]
func (ctrl *OAuthController) HandleCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "provider参数不能为空")
		return
	}

	var req schema.OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	clientIP := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	result, err := ctrl.oauthService.HandleCallback(c.Request.Context(), provider, req, clientIP, ua)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetUserBindList Get user OAuth binding list | 获取用户OAuth绑定列表
// @Summary Get user OAuth binding list | 获取用户OAuth绑定列表
// @Description Get current user's all OAuth bindings | 获取当前用户的所有OAuth绑定
// @Tags [User]OAuth | [用户]OAuth登录
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.OAuthUserBindListResponse} "Success | 获取成功"
// @Failure 401 {object} response.Data "Unauthorized | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /user/oauth/bindlist [get]
// @Security Bearer
func (ctrl *OAuthController) GetUserBindList(c *gin.Context) {
	userID, err := ctrl.getCurrentUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}

	result, err := ctrl.oauthService.GetUserBindList(c.Request.Context(), userID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetBindURL Get OAuth bind authorization URL | 获取OAuth绑定授权URL
// @Summary Get OAuth bind authorization URL | 获取OAuth绑定授权URL
// @Description Get authorization URL for binding OAuth account | 获取绑定OAuth账号的授权URL
// @Tags [User]OAuth | [用户]OAuth登录
// @Accept json
// @Produce json
// @Param provider path string true "Provider type | 提供商类型" Enums(QQ, GitHub, Google)
// @Param redirect_uri query string true "Frontend callback URL | 前端回调地址"
// @Success 200 {object} response.Data{data=schema.OAuthAuthorizeResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /user/oauth/{provider}/bindurl [get]
// @Security Bearer
func (ctrl *OAuthController) GetBindURL(c *gin.Context) {
	userID, err := ctrl.getCurrentUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}

	provider := c.Param("provider")
	if provider == "" {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "provider参数不能为空")
		return
	}

	var req schema.OAuthAuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	clientIP := c.ClientIP()
	result, err := ctrl.oauthService.GetBindURL(c.Request.Context(), userID, provider, req, clientIP)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// HandleBindCallback Handle OAuth bind callback | 处理OAuth绑定回调
// @Summary Handle OAuth bind callback | 处理OAuth绑定回调
// @Description Handle OAuth bind callback, bindOAuth account to current user | 处理OAuth绑定回调，将OAuth账号绑定到当前用户
// @Tags [User]OAuth | [用户]OAuth登录
// @Accept json
// @Produce json
// @Param provider path string true "Provider type | 提供商类型" Enums(QQ, GitHub, Google)
// @Param request body schema.OAuthBindCallbackRequest true "OAuth bind callback parameters | OAuth绑定回调参数"
// @Success 200 {object} response.Data "Success | 绑定成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /user/oauth/{provider}/bindcallback [post]
// @Security Bearer
func (ctrl *OAuthController) HandleBindCallback(c *gin.Context) {
	userID, err := ctrl.getCurrentUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}

	provider := c.Param("provider")
	if provider == "" {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "provider参数不能为空")
		return
	}

	var req schema.OAuthBindCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	clientIP := c.ClientIP()
	if err := ctrl.oauthService.HandleBindCallback(c.Request.Context(), userID, provider, req, clientIP); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// Unbind Unbind OAuth account | 解绑OAuth账号
// @Summary Unbind OAuth account | 解绑OAuth账号
// @Description Unbind OAuth account from current user | 从当前用户解绑OAuth账号
// @Tags [User]OAuth | [用户]OAuth登录
// @Accept json
// @Produce json
// @Param provider path string true "Provider type | 提供商类型" Enums(QQ, GitHub, Google)
// @Success 200 {object} response.Data "Success | 解绑成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /user/oauth/{provider} [delete]
// @Security Bearer
func (ctrl *OAuthController) Unbind(c *gin.Context) {
	userID, err := ctrl.getCurrentUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, err.Error())
		return
	}

	provider := c.Param("provider")
	if provider == "" {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "provider参数不能为空")
		return
	}

	if err := ctrl.oauthService.Unbind(c.Request.Context(), userID, provider); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// getCurrentUserID Get current logged-in user ID | 获取当前登录用户ID
func (ctrl *OAuthController) getCurrentUserID(c *gin.Context) (int, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("未找到Authorization header")
	}

	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	// Convert String to Int | String转Int
	userID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
