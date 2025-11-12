package controller

import (
	"strconv"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// OAuthProviderController OAuth提供商管理控制器
type OAuthProviderController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewOAuthProviderController 创建OAuth提供商管理控制器实例
func NewOAuthProviderController(injector *do.Injector) *OAuthProviderController {
	return &OAuthProviderController{
		injector: injector,
	}
}

// OAuthProviderRouter OAuth提供商管理相关路由注册
func (ctrl *OAuthProviderController) OAuthProviderRouter(router *gin.RouterGroup) {
	// OAuth提供商列表
	router.GET("", ctrl.GetProviderList)
	// 创建OAuth提供商
	router.POST("", ctrl.CreateProvider)
	// 更新OAuth提供商信息
	router.PUT("", ctrl.UpdateProvider)
	// 获取OAuth提供商详情
	router.GET("/:id", ctrl.GetProviderDetail)
	// 删除OAuth提供商
	router.DELETE("/:id", ctrl.DeleteProvider)

	// OAuth提供商状态管理
	router.PUT("/status", ctrl.UpdateProviderStatus)
}

// GetProviderList 获取OAuth提供商列表
// @Summary 获取OAuth提供商列表
// @Description 获取所有OAuth提供商列表，支持提供商类型和启用状态筛选
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param provider query string false "提供商类型" example("GitHub")
// @Param enabled query bool false "是否启用" example(true)
// @Success 200 {object} response.Data{data=schema.OAuthProviderListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth [get]
// @Security Bearer
func (ctrl *OAuthProviderController) GetProviderList(c *gin.Context) {
	var req schema.OAuthProviderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := oauthProviderService.GetProviderList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateProvider 创建OAuth提供商
// @Summary 创建OAuth提供商
// @Description 管理员创建新的OAuth提供商配置
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderCreateRequest true "OAuth提供商信息"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth [post]
// @Security Bearer
func (ctrl *OAuthProviderController) CreateProvider(c *gin.Context) {
	var req schema.OAuthProviderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	provider, err := oauthProviderService.CreateProvider(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.OAuthProviderDetailResponse{
		ID:           provider.ID,
		Provider:     provider.Provider.String(),
		ClientID:     provider.ClientID,
		ClientSecret: "***", // 创建时不返回完整密钥
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		UserInfoURL:  provider.UserInfoURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		ExtraConfig:  provider.ExtraConfig,
		Enabled:      provider.Enabled,
		SortOrder:    provider.SortOrder,
		CreatedAt:    provider.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    provider.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response.ResSuccess(c, result)
}

// UpdateProvider 更新OAuth提供商信息
// @Summary 更新OAuth提供商信息
// @Description 更新OAuth提供商的配置信息
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderUpdateRequest true "OAuth提供商信息"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth [put]
// @Security Bearer
func (ctrl *OAuthProviderController) UpdateProvider(c *gin.Context) {
	var req schema.OAuthProviderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	provider, err := oauthProviderService.UpdateProvider(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
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
		CreatedAt:    provider.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    provider.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response.ResSuccess(c, result)
}

// GetProviderDetail 获取OAuth提供商详情
// @Summary 获取OAuth提供商详情
// @Description 获取指定OAuth提供商的详细信息
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param id path int true "OAuth提供商ID"
// @Success 200 {object} response.Data{data=schema.OAuthProviderDetailResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 404 {object} response.Data "OAuth提供商不存在"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth/{id} [get]
// @Security Bearer
func (ctrl *OAuthProviderController) GetProviderDetail(c *gin.Context) {
	// 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "无效的OAuth提供商ID")
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := oauthProviderService.GetProviderDetail(c.Request.Context(), id)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteProvider 删除OAuth提供商
// @Summary 删除OAuth提供商
// @Description 删除指定的OAuth提供商配置
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param id path int true "OAuth提供商ID"
// @Success 200 {object} response.Data "删除成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 404 {object} response.Data "OAuth提供商不存在"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth/{id} [delete]
// @Security Bearer
func (ctrl *OAuthProviderController) DeleteProvider(c *gin.Context) {
	// 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "无效的OAuth提供商ID")
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	if err = oauthProviderService.DeleteProvider(c.Request.Context(), id); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateProviderStatus 更新OAuth提供商状态
// @Summary 更新OAuth提供商状态
// @Description 启用或禁用OAuth提供商
// @Tags [超级管理员]OAuth提供商管理
// @Accept json
// @Produce json
// @Param request body schema.OAuthProviderStatusUpdateRequest true "状态更新信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 404 {object} response.Data "OAuth提供商不存在"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/oauth/status [put]
// @Security Bearer
func (ctrl *OAuthProviderController) UpdateProviderStatus(c *gin.Context) {
	var req schema.OAuthProviderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	oauthProviderService, err := do.Invoke[service.IOAuthProviderService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	if err = oauthProviderService.UpdateProviderStatus(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
