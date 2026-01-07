package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// OAuthProviderController OAuth provider management controller | OAuth提供商管理控制器
type OAuthProviderController struct {
	// Injector instance for obtaining services | 注入器实例,用于获取服务
	injector *do.Injector
}

// NewOAuthProviderController Create OAuth provider management controller instance | 创建OAuth提供商管理控制器实例
func NewOAuthProviderController(injector *do.Injector) *OAuthProviderController {
	return &OAuthProviderController{
		injector: injector,
	}
}

// OAuthProviderRouter OAuth provider management related route registration | OAuth提供商管理相关路由注册
func (ctrl *OAuthProviderController) OAuthProviderRouter(router *gin.RouterGroup) {
	// OAuth provider list | OAuth提供商列表
	router.GET("", ctrl.GetProviderList)
	// Create OAuth provider | 创建OAuth提供商
	router.POST("", ctrl.CreateProvider)
	// Update OAuth provider information | 更新OAuth提供商信息
	router.PUT("", ctrl.UpdateProvider)
	// Get OAuth provider details | 获取OAuth提供商详情
	router.GET("/:id", ctrl.GetProviderDetail)
	// Delete OAuth provider | 删除OAuth提供商
	router.DELETE("/:id", ctrl.DeleteProvider)

	// OAuth provider status management | OAuth提供商状态管理
	router.PUT("/status", ctrl.UpdateProviderStatus)
}

// GetProviderList Get OAuth provider list | 获取OAuth提供商列表
// @Summary Get OAuth provider list | 获取OAuth提供商列表
// @Description Get all OAuth provider list, supports filtering by provider type and enabled status | 获取所有OAuth提供商列表,支持提供商类型和启用状态筛选
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param provider query string false "Provider type | 提供商类型" example("GitHub")
// @Param enabled query bool false "Is enabled | 是否启用" example(true)
// @Success 200 {object} response.Data{data=schema.OAuthProviderListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth [get]
// @Security Bearer
func (ctrl *OAuthProviderController) GetProviderList(c *gin.Context) {
	var req schema.OAuthProviderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := oauthProviderService.GetProviderList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateProvider Create OAuth provider | 创建OAuth提供商
// @Summary Create OAuth provider | 创建OAuth提供商
// @Description Admin creates new OAuth provider configuration | 管理员创建新的OAuth提供商配置
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderCreateRequest true "OAuth provider information | OAuth提供商信息"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth [post]
// @Security Bearer
func (ctrl *OAuthProviderController) CreateProvider(c *gin.Context) {
	var req schema.OAuthProviderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	provider, err := oauthProviderService.CreateProvider(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.OAuthProviderDetailResponse{
		ID:           provider.ID,
		Provider:     provider.Provider.String(),
		ClientID:     provider.ClientID,
		ClientSecret: "***", // Don't return complete secret when creating | 创建时不返回完整密钥
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		UserInfoURL:  provider.UserInfoURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		ExtraConfig:  provider.ExtraConfig,
		Enabled:      provider.Enabled,
		SortOrder:    provider.SortOrder,
		CreatedAt:    provider.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:    provider.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// UpdateProvider Update OAuth provider information | 更新OAuth提供商信息
// @Summary Update OAuth provider information | 更新OAuth提供商信息
// @Description Update OAuth provider configuration information | 更新OAuth提供商的配置信息
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderUpdateRequest true "OAuth provider information | OAuth提供商信息"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth [put]
// @Security Bearer
func (ctrl *OAuthProviderController) UpdateProvider(c *gin.Context) {
	var req schema.OAuthProviderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	provider, err := oauthProviderService.UpdateProvider(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	maskedSecret := "***"
	if provider.ClientSecret != "" && len(provider.ClientSecret) > 3 {
		maskedSecret = provider.ClientSecret[:3] + "***"
	}

	result := &schema.OAuthProviderDetailResponse{
		ID:           provider.ID,
		Provider:     provider.Provider.String(),
		ClientID:     provider.ClientID,
		ClientSecret: maskedSecret,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		UserInfoURL:  provider.UserInfoURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		ExtraConfig:  provider.ExtraConfig,
		Enabled:      provider.Enabled,
		SortOrder:    provider.SortOrder,
		CreatedAt:    provider.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:    provider.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// GetProviderDetail Get OAuth provider details | 获取OAuth提供商详情
// @Summary Get OAuth provider details | 获取OAuth提供商详情
// @Description Get detailed information of specified OAuth provider | 获取指定OAuth提供商的详细信息
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param id path int true "OAuth provider ID | OAuth提供商ID"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 404 {object} response.Data "OAuth provider not found | OAuth提供商不存在"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth/{id} [get]
// @Security Bearer
func (ctrl *OAuthProviderController) GetProviderDetail(c *gin.Context) {
	// Get path parameter | 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid OAuth provider ID | 无效的OAuth提供商ID")
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	result, err := oauthProviderService.GetProviderDetail(c.Request.Context(), id)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteProvider Delete OAuth provider | 删除OAuth提供商
// @Summary Delete OAuth provider | 删除OAuth提供商
// @Description Delete specified OAuth provider configuration | 删除指定的OAuth提供商配置
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param id path int true "OAuth provider ID | OAuth提供商ID"
// @Success 200 {object} response.Data "Deleted successfully | 删除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 404 {object} response.Data "OAuth provider not found | OAuth提供商不存在"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth/{id} [delete]
// @Security Bearer
func (ctrl *OAuthProviderController) DeleteProvider(c *gin.Context) {
	// Get path parameter | 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid OAuth provider ID | 无效的OAuth提供商ID")
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	if err = oauthProviderService.DeleteProvider(c.Request.Context(), id); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateProviderStatus Update OAuth provider status | 更新OAuth提供商状态
// @Summary Update OAuth provider status | 更新OAuth提供商状态
// @Description Enable or disable OAuth provider | 启用或禁用OAuth提供商
// @Tags [Super Admin]OAuth Provider Management | [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderStatusUpdateRequest true "Status update information | 状态更新信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 404 {object} response.Data "OAuth provider not found | OAuth提供商不存在"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/oauth/status [put]
// @Security Bearer
func (ctrl *OAuthProviderController) UpdateProviderStatus(c *gin.Context) {
	var req schema.OAuthProviderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get service | 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Call service | 调用服务
	if err = oauthProviderService.UpdateProviderStatus(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
