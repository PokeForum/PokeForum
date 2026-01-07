package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// CategoryManageController Category management controller | 版块管理控制器
type CategoryManageController struct {
	categoryManageService service.ICategoryManageService
}

// NewCategoryManageController Create category management controller instance | 创建版块管理控制器实例
func NewCategoryManageController(categoryManageService service.ICategoryManageService) *CategoryManageController {
	return &CategoryManageController{
		categoryManageService: categoryManageService,
	}
}

// CategoryManageRouter Category management related route registration | 版块管理相关路由注册
func (ctrl *CategoryManageController) CategoryManageRouter(router *gin.RouterGroup) {
	// Category list | 版块列表
	router.GET("", ctrl.GetCategoryList)
	// Create category | 创建版块
	router.POST("", ctrl.CreateCategory)
	// Update category information | 更新版块信息
	router.PUT("", ctrl.UpdateCategory)
	// Get category details | 获取版块详情
	router.GET("/:id", ctrl.GetCategoryDetail)
	// Delete category | 删除版块
	router.DELETE("/:id", ctrl.DeleteCategory)

	// Category status management | 版块状态管理
	router.PUT("/status", ctrl.UpdateCategoryStatus)

	// Category moderator management | 版块版主管理
	router.PUT("/moderators", ctrl.SetCategoryModerators)
}

// GetCategoryList Get category list | 获取版块列表
// @Summary Get category list | 获取版块列表
// @Description Get paginated category list with support for keyword search and status filtering | 分页获取版块列表，支持关键词搜索和状态筛选
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param keyword query string false "Search keyword | 搜索关键词" example("技术")
// @Param status query string false "Category status | 版块状态" example("Normal")
// @Success 200 {object} response.Data{data=schema.CategoryListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories [get]
func (ctrl *CategoryManageController) GetCategoryList(c *gin.Context) {
	var req schema.CategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.categoryManageService.GetCategoryList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateCategory Create category | 创建版块
// @Summary Create category | 创建版块
// @Description Admin creates new category | 管理员创建新版块
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryCreateRequest true "Category information | 版块信息"
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories [post]
func (ctrl *CategoryManageController) CreateCategory(c *gin.Context) {
	var req schema.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	category, err := ctrl.categoryManageService.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Weight:      category.Weight,
		Status:      category.Status.String(),
		CreatedAt:   category.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:   category.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// UpdateCategory Update category information | 更新版块信息
// @Summary Update category information | 更新版块信息
// @Description Update basic information of category | 更新版块的基本信息
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryUpdateRequest true "Category information | 版块信息"
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories [put]
func (ctrl *CategoryManageController) UpdateCategory(c *gin.Context) {
	var req schema.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	category, err := ctrl.categoryManageService.UpdateCategory(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Weight:      category.Weight,
		Status:      category.Status.String(),
		CreatedAt:   category.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:   category.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// GetCategoryDetail Get category details | 获取版块详情
// @Summary Get category details | 获取版块详情
// @Description Get detailed information of specified category | 获取指定版块的详细信息
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param id path int true "Category ID | 版块ID" example("1")
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories/{id} [get]
func (ctrl *CategoryManageController) GetCategoryDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.categoryManageService.GetCategoryDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteCategory Delete category | 删除版块
// @Summary Delete category | 删除版块
// @Description Soft delete category (set status to hidden) | 软删除版块（将状态设为隐藏）
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param id path int true "Category ID | 版块ID" example("1")
// @Success 200 {object} response.Data "Deleted successfully | 删除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories/{id} [delete]
func (ctrl *CategoryManageController) DeleteCategory(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	err := ctrl.categoryManageService.DeleteCategory(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateCategoryStatus Update category status | 更新版块状态
// @Summary Update category status | 更新版块状态
// @Description Update category status (normal, login-visible, member-visible, hidden, locked) | 更新版块的状态（正常、登录可见、会员可见、隐藏、锁定）
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryStatusUpdateRequest true "Status information | 状态信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories/status [put]
func (ctrl *CategoryManageController) UpdateCategoryStatus(c *gin.Context) {
	var req schema.CategoryStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	err := ctrl.categoryManageService.UpdateCategoryStatus(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCategoryModerators Set category moderators | 设置版块版主
// @Summary Set category moderators | 设置版块版主
// @Description Set moderator list for specified category | 为指定版块设置版主列表
// @Tags [Admin]Category Management | [管理员]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryModeratorRequest true "Moderator information | 版主信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/categories/moderators [put]
func (ctrl *CategoryManageController) SetCategoryModerators(c *gin.Context) {
	var req schema.CategoryModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Invoke service | 调用服务
	err := ctrl.categoryManageService.SetCategoryModerators(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
