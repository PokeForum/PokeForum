package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// EmailController 邮箱控制器
type EmailController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewEmailController 创建邮箱控制器实例
func NewEmailController(injector *do.Injector) *EmailController {
	return &EmailController{
		injector: injector,
	}
}

// EmailRouter 邮箱相关路由注册
// 注册邮箱服务的所有路由端点
func (ctrl *EmailController) EmailRouter(router *gin.RouterGroup) {
	// 获取SMTP配置，GET请求
	router.GET("/smtp", ctrl.GetSMTPConfig)
	// 更新SMTP配置，POST请求
	router.POST("/smtp", ctrl.UpdateSMTPConfig)
	// 发送测试邮件，POST请求
	router.POST("/test", ctrl.SendTestEmail)
}

// GetSMTPConfig 获取SMTP配置接口
// @Summary 获取SMTP配置
// @Description 获取当前系统的SMTP邮箱服务配置
// @Tags Email
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.EmailSMTPConfigResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/email/smtp [get]
// @Security Bearer
func (ctrl *EmailController) GetSMTPConfig(c *gin.Context) {
	// 从注入器获取 EmailService 实例
	emailService, err := do.Invoke[service.IEmailService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务获取SMTP配置，返回配置对象
	config, err := emailService.GetSMTPConfig(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 返回成功响应，包含SMTP配置信息
	response.ResSuccess(c, config)
}

// UpdateSMTPConfig 更新SMTP配置接口
// @Summary 更新SMTP配置
// @Description 更新系统的SMTP邮箱服务配置
// @Tags Email
// @Accept json
// @Produce json
// @Param request body schema.EmailSMTPConfigRequest true "SMTP配置信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/email/smtp [post]
// @Security Bearer
func (ctrl *EmailController) UpdateSMTPConfig(c *gin.Context) {
	// 解析请求体中的SMTP配置信息
	var req schema.EmailSMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 从注入器获取 EmailService 实例
	emailService, err := do.Invoke[service.IEmailService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务更新SMTP配置到数据库
	if err := emailService.UpdateSMTPConfig(c.Request.Context(), req); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, nil)
}

// SendTestEmail 发送测试邮件接口
// @Summary 发送测试邮件
// @Description 使用当前SMTP配置发送一封测试邮件
// @Tags Email
// @Accept json
// @Produce json
// @Param request body schema.EmailTestRequest true "收件人邮箱"
// @Success 200 {object} response.Data{data=schema.EmailTestResponse} "发送成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/email/test [post]
// @Security Bearer
func (ctrl *EmailController) SendTestEmail(c *gin.Context) {
	// 解析请求体中的收件人邮箱地址
	var req schema.EmailTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 从注入器获取 EmailService 实例
	emailService, err := do.Invoke[service.IEmailService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务发送测试邮件到指定邮箱
	if err := emailService.SendTestEmail(c.Request.Context(), req.ToEmail); err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 返回成功响应，包含发送结果信息
	response.ResSuccess(c, schema.EmailTestResponse{
		Success: true,
		Message: "测试邮件已发送",
	})
}
