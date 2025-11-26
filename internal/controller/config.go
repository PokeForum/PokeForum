package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// ConfigController 公开配置控制器
type ConfigController struct {
	injector *do.Injector
}

// NewConfigController 创建公开配置控制器实例
func NewConfigController(injector *do.Injector) *ConfigController {
	return &ConfigController{
		injector: injector,
	}
}

// ConfigRouter 公开配置路由注册
func (ctrl *ConfigController) ConfigRouter(router *gin.RouterGroup) {
	router.GET("", ctrl.GetPublicConfig)
}

// GetPublicConfig 获取公开配置
// @Summary 获取公开配置
// @Description 获取客户端所需的公开配置，包括常规、首页、SEO、安全、代码、评论配置
// @Tags 公开配置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PublicConfigResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /config [get]
func (ctrl *ConfigController) GetPublicConfig(c *gin.Context) {
	settingsService, err := do.Invoke[service.ISettingsService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	ctx := c.Request.Context()

	// 并行获取所有配置
	routine, err := settingsService.GetRoutineSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	home, err := settingsService.GetHomeSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	seo, err := settingsService.GetSeoSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	safe, err := settingsService.GetSafeSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	code, err := settingsService.GetCodeSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	comment, err := settingsService.GetCommentSettings(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, schema.PublicConfigResponse{
		Routine: routine,
		Home:    home,
		Seo:     seo,
		Safe:    safe,
		Code:    code,
		Comment: comment,
	})
}
