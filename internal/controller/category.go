package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
)

// CategoryController 用户侧版块控制器
type CategoryController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewCategoryController 创建用户侧版块控制器实例
func NewCategoryController(injector *do.Injector) *CategoryController {
	return &CategoryController{
		injector: injector,
	}
}

// CategoryRouter 用户侧版块相关路由注册
func (ctrl *CategoryController) CategoryRouter(router *gin.RouterGroup) {
	// 获取版块列表
	router.GET("", ctrl.GetUserCategories)
}

// GetUserCategories 获取用户可见的版块列表
// @Summary 获取版块列表
// @Description 获取用户可见的版块列表，包括正常、登录可见和锁定状态的版块，隐藏版块不可见
// @Tags [用户]版块
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.UserCategoryResponse} "获取成功"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /categories [get]
func (ctrl *CategoryController) GetUserCategories(c *gin.Context) {
	// 获取服务
	categoryService, err := do.Invoke[service.ICategoryService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := categoryService.GetUserCategories(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
