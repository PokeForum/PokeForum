package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// SettingsController Settings controller | 设置控制器
type SettingsController struct {
	settingsService service.ISettingsService
}

// NewSettingsController Create settings controller instance | 创建设置控制器实例
func NewSettingsController(injector *do.Injector) *SettingsController {
	return &SettingsController{
		settingsService: do.MustInvoke[service.ISettingsService](injector),
	}
}

// SettingsRouter Settings related route registration | 设置相关路由注册
// Register all settings module route endpoints | 注册所有设置模块的路由端点
func (ctrl *SettingsController) SettingsRouter(router *gin.RouterGroup) {
	// Routine settings | 常规设置
	routineGroup := router.Group("/routine")
	{
		routineGroup.GET("", ctrl.GetRoutineSettings)
		routineGroup.POST("", ctrl.UpdateRoutineSettings)
	}

	// Home settings | 首页设置
	homeGroup := router.Group("/home")
	{
		homeGroup.GET("", ctrl.GetHomeSettings)
		homeGroup.POST("", ctrl.UpdateHomeSettings)
	}

	// Comment settings | 评论设置
	commentGroup := router.Group("/comment")
	{
		commentGroup.GET("", ctrl.GetCommentSettings)
		commentGroup.POST("", ctrl.UpdateCommentSettings)
	}

	// SEO settings | SEO设置
	seoGroup := router.Group("/seo")
	{
		seoGroup.GET("", ctrl.GetSeoSettings)
		seoGroup.POST("", ctrl.UpdateSeoSettings)
	}

	// Code configuration | 代码配置
	codeGroup := router.Group("/code")
	{
		codeGroup.GET("", ctrl.GetCodeSettings)
		codeGroup.POST("", ctrl.UpdateCodeSettings)
	}

	// Security settings | 安全设置
	safeGroup := router.Group("/safe")
	{
		safeGroup.GET("", ctrl.GetSafeSettings)
		safeGroup.POST("", ctrl.UpdateSafeSettings)
	}

	// Email settings | 邮箱设置
	emailGroup := router.Group("/email")
	{
		emailGroup.GET("", ctrl.GetEmailSettings)
		emailGroup.POST("", ctrl.UpdateEmailSettings)
		emailGroup.POST("/test", ctrl.SendTestEmail)
	}

	// Sign-in settings | 签到设置
	signinGroup := router.Group("/signin")
	{
		signinGroup.GET("", ctrl.GetSigninSettings)
		signinGroup.POST("", ctrl.UpdateSigninSettings)
	}
}

// GetRoutineSettings Get routine settings | 获取常规设置
// @Summary Get routine settings | 获取常规设置
// @Description Get website routine configuration including Logo, Icon, filing number, etc. | 获取网站的常规配置，包括Logo、Icon、备案号等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.RoutineSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/routine [get]
// @Security Bearer
func (ctrl *SettingsController) GetRoutineSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetRoutineSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateRoutineSettings Update routine settings | 更新常规设置
// @Summary Update routine settings | 更新常规设置
// @Description Update website routine configuration | 更新网站的常规配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.RoutineSettingsRequest true "Routine settings information | 常规设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/routine [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateRoutineSettings(c *gin.Context) {
	var req schema.RoutineSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateRoutineSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetHomeSettings Get home settings | 获取首页设置
// @Summary Get home settings | 获取首页设置
// @Description Get home page configuration including slides, friend links, etc. | 获取首页的配置，包括幻灯片、友情链接等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.HomeSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/home [get]
// @Security Bearer
func (ctrl *SettingsController) GetHomeSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetHomeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateHomeSettings Update home settings | 更新首页设置
// @Summary Update home settings | 更新首页设置
// @Description Update home page configuration | 更新首页的配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.HomeSettingsRequest true "Home settings information | 首页设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/home [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateHomeSettings(c *gin.Context) {
	var req schema.HomeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateHomeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetCommentSettings Get comment settings | 获取评论设置
// @Summary Get comment settings | 获取评论设置
// @Description Get comment related configuration including review, blacklist, etc. | 获取评论相关的配置，包括审核、黑名单等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.CommentSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/comment [get]
// @Security Bearer
func (ctrl *SettingsController) GetCommentSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetCommentSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateCommentSettings Update comment settings | 更新评论设置
// @Summary Update comment settings | 更新评论设置
// @Description Update comment related configuration | 更新评论相关的配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.CommentSettingsRequest true "Comment settings information | 评论设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/comment [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateCommentSettings(c *gin.Context) {
	var req schema.CommentSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateCommentSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetSeoSettings Get SEO settings | 获取SEO设置
// @Summary Get SEO settings | 获取SEO设置
// @Description Get website SEO related configuration including site name, keywords, description, etc. | 获取网站SEO相关配置，包括网站名称、关键词、描述等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SeoSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/seo [get]
// @Security Bearer
func (ctrl *SettingsController) GetSeoSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetSeoSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateSeoSettings Update SEO settings | 更新SEO设置
// @Summary Update SEO settings | 更新SEO设置
// @Description Update website SEO related configuration | 更新网站SEO相关配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.SeoSettingsRequest true "SEO settings information | SEO设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/seo [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateSeoSettings(c *gin.Context) {
	var req schema.SeoSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateSeoSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetCodeSettings Get code configuration | 获取代码配置
// @Summary Get code configuration | 获取代码配置
// @Description Get custom code configuration including header, footer code and custom CSS | 获取自定义代码配置，包括页头、页脚代码和自定义CSS
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.CodeSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/code [get]
// @Security Bearer
func (ctrl *SettingsController) GetCodeSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetCodeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateCodeSettings Update code configuration | 更新代码配置
// @Summary Update code configuration | 更新代码配置
// @Description Update custom code configuration | 更新自定义代码配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.CodeSettingsRequest true "Code configuration information | 代码配置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/code [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateCodeSettings(c *gin.Context) {
	var req schema.CodeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateCodeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetSafeSettings Get security settings | 获取安全设置
// @Summary Get security settings | 获取安全设置
// @Description Get security related configuration including registration control, email whitelist, etc. | 获取安全相关配置，包括注册控制、邮箱白名单等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SafeSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/safe [get]
// @Security Bearer
func (ctrl *SettingsController) GetSafeSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetSafeSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateSafeSettings Update security settings | 更新安全设置
// @Summary Update security settings | 更新安全设置
// @Description Update security related configuration | 更新安全相关配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.SafeSettingsRequest true "Security settings information | 安全设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/safe [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateSafeSettings(c *gin.Context) {
	var req schema.SafeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateSafeSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// GetEmailSettings Get email settings | 获取邮箱设置
// @Summary Get email settings | 获取邮箱设置
// @Description Get SMTP email service configuration | 获取SMTP邮箱服务配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.EmailSMTPConfigResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/email [get]
// @Security Bearer
func (ctrl *SettingsController) GetEmailSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetSMTPConfig(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateEmailSettings Update email settings | 更新邮箱设置
// @Summary Update email settings | 更新邮箱设置
// @Description Update SMTP email service configuration | 更新SMTP邮箱服务配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.EmailSMTPConfigRequest true "SMTP configuration information | SMTP配置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/email [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateEmailSettings(c *gin.Context) {
	var req schema.EmailSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateSMTPConfig(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SendTestEmail Send test email | 发送测试邮件
// @Summary Send test email | 发送测试邮件
// @Description Send a test email using current SMTP configuration | 使用当前SMTP配置发送一封测试邮件
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.EmailTestRequest true "Recipient email | 收件人邮箱"
// @Success 200 {object} response.Data{data=schema.EmailTestResponse} "Sent successfully | 发送成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/email/test [post]
// @Security Bearer
func (ctrl *SettingsController) SendTestEmail(c *gin.Context) {
	var req schema.EmailTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID from gin.Context and set to context | 从gin.Context获取用户ID并设置到context中
	ctx := tracing.ContextWithUserID(c, c.Request.Context())

	if err := ctrl.settingsService.SendTestEmail(ctx, req.ToEmail); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, schema.EmailTestResponse{
		Success: true,
		Message: "Test email sent | 测试邮件已发送",
	})
}

// GetSigninSettings Get sign-in settings | 获取签到设置
// @Summary Get sign-in settings | 获取签到设置
// @Description Get sign-in feature related configuration including reward rules, mode, etc. | 获取签到功能相关配置，包括奖励规则、模式等
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SigninSettingsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/signin [get]
// @Security Bearer
func (ctrl *SettingsController) GetSigninSettings(c *gin.Context) {
	config, err := ctrl.settingsService.GetSigninSettings(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, config)
}

// UpdateSigninSettings Update sign-in settings | 更新签到设置
// @Summary Update sign-in settings | 更新签到设置
// @Description Update sign-in feature related configuration | 更新签到功能相关配置
// @Tags [Super Admin]System Settings | [超级管理员]系统设置
// @Accept json
// @Produce json
// @Param request body schema.SigninSettingsRequest true "Sign-in settings information | 签到设置信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/settings/signin [post]
// @Security Bearer
func (ctrl *SettingsController) UpdateSigninSettings(c *gin.Context) {
	var req schema.SigninSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := ctrl.settingsService.UpdateSigninSettings(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
