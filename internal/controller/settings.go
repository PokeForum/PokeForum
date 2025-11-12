package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// SettingsController 设置控制器
type SettingsController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewSettingsController 创建设置控制器实例
func NewSettingsController(injector *do.Injector) *SettingsController {
	return &SettingsController{
		injector: injector,
	}
}

// SettingsRouter 设置相关路由注册
// 注册所有设置模块的路由端点
func (ctrl *SettingsController) SettingsRouter(router *gin.RouterGroup) {
	// 常规设置
	routineGroup := router.Group("/routine")
	{
		routineGroup.GET("", ctrl.GetRoutineSettings)
		routineGroup.POST("", ctrl.UpdateRoutineSettings)
	}

	// 首页设置
	homeGroup := router.Group("/home")
	{
		homeGroup.GET("", ctrl.GetHomeSettings)
		homeGroup.POST("", ctrl.UpdateHomeSettings)
	}

	// 评论设置
	commentGroup := router.Group("/comment")
	{
		commentGroup.GET("", ctrl.GetCommentSettings)
		commentGroup.POST("", ctrl.UpdateCommentSettings)
	}

	// SEO设置
	seoGroup := router.Group("/seo")
	{
		seoGroup.GET("", ctrl.GetSeoSettings)
		seoGroup.POST("", ctrl.UpdateSeoSettings)
	}

	// 代码配置
	codeGroup := router.Group("/code")
	{
		codeGroup.GET("", ctrl.GetCodeSettings)
		codeGroup.POST("", ctrl.UpdateCodeSettings)
	}

	// 安全设置
	safeGroup := router.Group("/safe")
	{
		safeGroup.GET("", ctrl.GetSafeSettings)
		safeGroup.POST("", ctrl.UpdateSafeSettings)
	}

	// 邮箱设置
	emailGroup := router.Group("/email")
	{
		emailGroup.GET("", ctrl.GetEmailSettings)
		emailGroup.POST("", ctrl.UpdateEmailSettings)
		emailGroup.POST("/test", ctrl.SendTestEmail)
	}
}

// GetRoutineSettings 获取常规设置
// @Summary 获取常规设置
// @Description 获取网站的常规配置，包括Logo、Icon、备案号等
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.RoutineSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/routine [get]
// @Security Bearer
func (ctrl *SettingsController) GetRoutineSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetRoutineSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateRoutineSettings 更新常规设置
// @Summary 更新常规设置
// @Description 更新网站的常规配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.RoutineSettingsRequest true "常规设置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/routine [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateRoutineSettings(c *gin.Context) {
	var req schema.RoutineSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateRoutineSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetHomeSettings 获取首页设置
// @Summary 获取首页设置
// @Description 获取首页的配置，包括幻灯片、友情链接等
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.HomeSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/home [get]
// @Security Bearer
func (ctrl *SettingsController) GetHomeSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetHomeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateHomeSettings 更新首页设置
// @Summary 更新首页设置
// @Description 更新首页的配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.HomeSettingsRequest true "首页设置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/home [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateHomeSettings(c *gin.Context) {
	var req schema.HomeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateHomeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetCommentSettings 获取评论设置
// @Summary 获取评论设置
// @Description 获取评论相关的配置，包括审核、黑名单等
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.CommentSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/comment [get]
// @Security Bearer
func (ctrl *SettingsController) GetCommentSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetCommentSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateCommentSettings 更新评论设置
// @Summary 更新评论设置
// @Description 更新评论相关的配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.CommentSettingsRequest true "评论设置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/comment [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateCommentSettings(c *gin.Context) {
	var req schema.CommentSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateCommentSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetSeoSettings 获取SEO设置
// @Summary 获取SEO设置
// @Description 获取网站SEO相关配置，包括网站名称、关键词、描述等
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SeoSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/seo [get]
// @Security Bearer
func (ctrl *SettingsController) GetSeoSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetSeoSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateSeoSettings 更新SEO设置
// @Summary 更新SEO设置
// @Description 更新网站SEO相关配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.SeoSettingsRequest true "SEO设置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/seo [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateSeoSettings(c *gin.Context) {
	var req schema.SeoSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateSeoSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetCodeSettings 获取代码配置
// @Summary 获取代码配置
// @Description 获取自定义代码配置，包括页头、页脚代码和自定义CSS
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.CodeSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/code [get]
// @Security Bearer
func (ctrl *SettingsController) GetCodeSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetCodeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateCodeSettings 更新代码配置
// @Summary 更新代码配置
// @Description 更新自定义代码配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.CodeSettingsRequest true "代码配置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/code [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateCodeSettings(c *gin.Context) {
	var req schema.CodeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateCodeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetSafeSettings 获取安全设置
// @Summary 获取安全设置
// @Description 获取安全相关配置，包括注册控制、邮箱白名单等
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SafeSettingsResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/safe [get]
// @Security Bearer
func (ctrl *SettingsController) GetSafeSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetSafeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateSafeSettings 更新安全设置
// @Summary 更新安全设置
// @Description 更新安全相关配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.SafeSettingsRequest true "安全设置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/safe [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateSafeSettings(c *gin.Context) {
	var req schema.SafeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateSafeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetEmailSettings 获取邮箱设置
// @Summary 获取邮箱设置
// @Description 获取SMTP邮箱服务配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.EmailSMTPConfigResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/email [get]
// @Security Bearer
func (ctrl *SettingsController) GetEmailSettings(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	config, err := settingsService.GetSMTPConfig(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateEmailSettings 更新邮箱设置
// @Summary 更新邮箱设置
// @Description 更新SMTP邮箱服务配置
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.EmailSMTPConfigRequest true "SMTP配置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/email [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateEmailSettings(c *gin.Context) {
	var req schema.EmailSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	if err = settingsService.UpdateSMTPConfig(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SendTestEmail 发送测试邮件
// @Summary 发送测试邮件
// @Description 使用当前SMTP配置发送一封测试邮件
// @Tags [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.EmailTestRequest true "收件人邮箱"
// @Success 200 {object} response.Data{data=schema.EmailTestResponse} "发送成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/settings/email/test [post]
// @Security Bearer
func (ctrl *SettingsController) SendTestEmail(c *gin.Context) {
	var req schema.EmailTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 从gin.Context获取用户ID并设置到context中
	ctx := tracing.ContextWithUserID(c, c.Request.Context())

	if err = settingsService.SendTestEmail(ctx, req.ToEmail); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, schema.EmailTestResponse{
		Success: true,
		Message: "测试邮件已发送",
	})
}
