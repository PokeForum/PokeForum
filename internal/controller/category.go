package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/service"
)

// CategoryController User-side category controller | 用户侧版块控制器
type CategoryController struct {
	categoryService service.ICategoryService
}

// NewCategoryController Create user-side category controller instance | 创建用户侧版块控制器实例
func NewCategoryController(categoryService service.ICategoryService) *CategoryController {
	return &CategoryController{
		categoryService: categoryService,
	}
}

// CategoryRouter User-side category related route registration | 用户侧版块相关路由注册
func (ctrl *CategoryController) CategoryRouter(router *gin.RouterGroup) {
	// Get category list | 获取版块列表
	router.GET("", ctrl.GetUserCategories)
}

// GetUserCategories Get list of categories visible to users | 获取用户可见的版块列表
// @Summary Get category list | 获取版块列表
// @Description Get list of categories visible to users. Normal and Locked categories are visible to everyone. LoginRequired categories are only visible to logged-in users. Hidden categories are not returned but can be accessed via direct URL | 获取用户可见的版块列表。Normal和Locked状态对所有人可见，LoginRequired仅登录用户可见，Hidden不在列表返回但可通过URL直接访问
// @Tags [User]Category | [用户]版块
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.UserCategoryResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /categories [get]
func (ctrl *CategoryController) GetUserCategories(c *gin.Context) {
	// 检查用户是否已登录 | Check if user is logged in
	isLoggedIn := satoken.IsLoggedIn(c)

	// Invoke service | 调用服务
	result, err := ctrl.categoryService.GetUserCategories(c.Request.Context(), isLoggedIn)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
