package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
)

// ConfigController Public configuration controller | 公开配置控制器
type ConfigController struct {
	settingsService service.ISettingsService
}

// NewConfigController Create public configuration controller instance | 创建公开配置控制器实例
func NewConfigController(settingsService service.ISettingsService) *ConfigController {
	return &ConfigController{
		settingsService: settingsService,
	}
}

// ConfigRouter Public configuration route registration | 公开配置路由注册
func (ctrl *ConfigController) ConfigRouter(router *gin.RouterGroup) {
	router.GET("", ctrl.GetPublicConfig)
}

// GetPublicConfig Get public configuration | 获取公开配置
// @Summary Get public configuration | 获取公开配置
// @Description Get public configuration required by client, including routine, home, SEO, security, code, and comment settings | 获取客户端所需的公开配置,包括常规、首页、SEO、安全、代码、评论配置
// @Tags Public Configuration | 公开配置
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PublicConfigResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /config [get]
func (ctrl *ConfigController) GetPublicConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Get public configuration (with 30-day cache) | 获取公开配置（30天缓存）
	result, err := ctrl.settingsService.GetPublicConfig(ctx)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
