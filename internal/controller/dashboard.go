package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// DashboardController 仪表盘控制器
type DashboardController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewDashboardController 创建仪表盘控制器实例
func NewDashboardController(injector *do.Injector) *DashboardController {
	return &DashboardController{
		injector: injector,
	}
}

// DashboardRouter 仪表盘相关路由注册
func (ctrl *DashboardController) DashboardRouter(router *gin.RouterGroup) {
	// 获取仪表盘统计数据
	router.GET("/stats", ctrl.GetDashboardStats)
	// 获取最近活动
	router.GET("/activity", ctrl.GetRecentActivity)
	// 获取热门帖子
	router.GET("/popular-posts", ctrl.GetPopularPosts)
	// 获取热门版块
	router.GET("/popular-categories", ctrl.GetPopularCategories)
}

// GetDashboardStats 获取仪表盘统计数据
// @Summary 获取仪表盘统计数据
// @Description 获取系统各项统计数据，包括用户、帖子、评论、版块和系统统计
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期" example("2024-01-01")
// @Param end_date query string false "结束日期" example("2024-12-31")
// @Success 200 {object} response.Data{data=schema.DashboardStatsResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/dashboard/stats [get]
func (ctrl *DashboardController) GetDashboardStats(c *gin.Context) {
	// 获取服务
	dashboardService, err := do.Invoke[service.IDashboardService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetRecentActivity 获取最近活动
// @Summary 获取最近活动
// @Description 获取系统最近的活动，包括最近帖子、评论和新用户
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.RecentActivityResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/dashboard/activity [get]
func (ctrl *DashboardController) GetRecentActivity(c *gin.Context) {
	// 获取服务
	dashboardService, err := do.Invoke[service.IDashboardService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := dashboardService.GetRecentActivity(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPopularPosts 获取热门帖子
// @Summary 获取热门帖子
// @Description 获取浏览量最高的热门帖子列表
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PopularPostsResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/dashboard/popular-posts [get]
func (ctrl *DashboardController) GetPopularPosts(c *gin.Context) {
	// 获取服务
	dashboardService, err := do.Invoke[service.IDashboardService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := dashboardService.GetPopularPosts(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPopularCategories 获取热门版块
// @Summary 获取热门版块
// @Description 获取帖子数量最多的热门版块列表
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PopularCategoriesResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/dashboard/popular-categories [get]
func (ctrl *DashboardController) GetPopularCategories(c *gin.Context) {
	// 获取服务
	dashboardService, err := do.Invoke[service.IDashboardService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := dashboardService.GetPopularCategories(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
