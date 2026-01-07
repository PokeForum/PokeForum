package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
)

// CategoryController User-side category controller | 用户侧版块控制器
type CategoryController struct {
	categoryService service.ICategoryService
}

// NewCategoryController Create user-side category controller instance | 创建用户侧版块控制器实例
func NewCategoryController(injector *do.Injector) *CategoryController {
	return &CategoryController{
		categoryService: do.MustInvoke[service.ICategoryService](injector),
	}
}

// CategoryRouter User-side category related route registration | 用户侧版块相关路由注册
func (ctrl *CategoryController) CategoryRouter(router *gin.RouterGroup) {
	// Get category list | 获取版块列表
	router.GET("", ctrl.GetUserCategories)
}

// GetUserCategories Get list of categories visible to users | 获取用户可见的版块列表
// @Summary Get category list | 获取版块列表
// @Description Get list of categories visible to users, including normal, login-visible and locked status categories, hidden categories are not visible | 获取用户可见的版块列表，包括正常、登录可见和锁定状态的版块，隐藏版块不可见
// @Tags [User]Category | [用户]版块
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.UserCategoryResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /categories [get]
func (ctrl *CategoryController) GetUserCategories(c *gin.Context) {
	// Invoke service | 调用服务
	result, err := ctrl.categoryService.GetUserCategories(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
