package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// CategoryManageController 版块管理控制器
type CategoryManageController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewCategoryManageController 创建版块管理控制器实例
func NewCategoryManageController(injector *do.Injector) *CategoryManageController {
	return &CategoryManageController{
		injector: injector,
	}
}

// CategoryManageRouter 版块管理相关路由注册
func (ctrl *CategoryManageController) CategoryManageRouter(router *gin.RouterGroup) {
	// 版块列表
	router.GET("", ctrl.GetCategoryList)
	// 创建版块
	router.POST("", ctrl.CreateCategory)
	// 更新版块信息
	router.PUT("", ctrl.UpdateCategory)
	// 获取版块详情
	router.GET("/:id", ctrl.GetCategoryDetail)
	// 删除版块
	router.DELETE("/:id", ctrl.DeleteCategory)

	// 版块状态管理
	router.PUT("/status", ctrl.UpdateCategoryStatus)

	// 版块版主管理
	router.PUT("/moderators", ctrl.SetCategoryModerators)
}

// GetCategoryList 获取版块列表
// @Summary 获取版块列表
// @Description 分页获取版块列表，支持关键词搜索和状态筛选
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Param keyword query string false "搜索关键词" example("技术")
// @Param status query string false "版块状态" example("Normal")
// @Success 200 {object} response.Data{data=schema.CategoryListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories [get]
func (ctrl *CategoryManageController) GetCategoryList(c *gin.Context) {
	var req schema.CategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := categoryManageService.GetCategoryList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateCategory 创建版块
// @Summary 创建版块
// @Description 管理员创建新版块
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param request body schema.CategoryCreateRequest true "版块信息"
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories [post]
func (ctrl *CategoryManageController) CreateCategory(c *gin.Context) {
	var req schema.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	category, err := categoryManageService.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Weight:      category.Weight,
		Status:      category.Status.String(),
		PostCount:   0, // TODO: 添加帖子统计功能
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response.ResSuccess(c, result)
}

// UpdateCategory 更新版块信息
// @Summary 更新版块信息
// @Description 更新版块的基本信息
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param request body schema.CategoryUpdateRequest true "版块信息"
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories [put]
func (ctrl *CategoryManageController) UpdateCategory(c *gin.Context) {
	var req schema.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	category, err := categoryManageService.UpdateCategory(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Weight:      category.Weight,
		Status:      category.Status.String(),
		PostCount:   0, // TODO: 添加帖子统计功能
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response.ResSuccess(c, result)
}

// GetCategoryDetail 获取版块详情
// @Summary 获取版块详情
// @Description 获取指定版块的详细信息
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param id path int true "版块ID" example("1")
// @Success 200 {object} response.Data{data=schema.CategoryDetailResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories/{id} [get]
func (ctrl *CategoryManageController) GetCategoryDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := categoryManageService.GetCategoryDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteCategory 删除版块
// @Summary 删除版块
// @Description 软删除版块（将状态设为隐藏）
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param id path int true "版块ID" example("1")
// @Success 200 {object} response.Data "删除成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories/{id} [delete]
func (ctrl *CategoryManageController) DeleteCategory(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = categoryManageService.DeleteCategory(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateCategoryStatus 更新版块状态
// @Summary 更新版块状态
// @Description 更新版块的状态（正常、登录可见、会员可见、隐藏、锁定）
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param request body schema.CategoryStatusUpdateRequest true "状态信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories/status [put]
func (ctrl *CategoryManageController) UpdateCategoryStatus(c *gin.Context) {
	var req schema.CategoryStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = categoryManageService.UpdateCategoryStatus(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCategoryModerators 设置版块版主
// @Summary 设置版块版主
// @Description 为指定版块设置版主列表
// @Tags CategoryManage
// @Accept json
// @Produce json
// @Param request body schema.CategoryModeratorRequest true "版主信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/categories/moderators [put]
func (ctrl *CategoryManageController) SetCategoryModerators(c *gin.Context) {
	var req schema.CategoryModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	categoryManageService, err := do.Invoke[service.ICategoryManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = categoryManageService.SetCategoryModerators(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
