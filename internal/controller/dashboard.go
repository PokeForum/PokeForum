package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
)

// DashboardController Dashboard controller | 仪表盘控制器
type DashboardController struct {
	dashboardService service.IDashboardService
}

// NewDashboardController Create dashboard controller instance | 创建仪表盘控制器实例
func NewDashboardController(dashboardService service.IDashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// DashboardRouter Dashboard related route registration | 仪表盘相关路由注册
func (ctrl *DashboardController) DashboardRouter(router *gin.RouterGroup) {
	// Get dashboard statistics | 获取仪表盘统计数据
	router.GET("/stats", ctrl.GetDashboardStats)
	// Get recent activity | 获取最近活动
	router.GET("/activity", ctrl.GetRecentActivity)
	// Get popular posts | 获取热门帖子
	router.GET("/popular-posts", ctrl.GetPopularPosts)
	// Get popular categories | 获取热门版块
	router.GET("/popular-categories", ctrl.GetPopularCategories)
}

// GetDashboardStats Get dashboard statistics | 获取仪表盘统计数据
// @Summary Get dashboard statistics | 获取仪表盘统计数据
// @Description Get system statistics including users, posts, comments, categories and system stats | 获取系统各项统计数据，包括用户、帖子、评论、版块和系统统计
// @Tags [Admin]Dashboard | [管理员]仪表盘
// @Accept json
// @Produce json
// @Param start_date query string false "Start date | 开始日期" example("2024-01-01")
// @Param end_date query string false "End date | 结束日期" example("2024-12-31")
// @Success 200 {object} response.Data{data=schema.DashboardStatsResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/dashboard/stats [get]
func (ctrl *DashboardController) GetDashboardStats(c *gin.Context) {
	// Invoke service | 调用服务
	result, err := ctrl.dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetRecentActivity Get recent activity | 获取最近活动
// @Summary Get recent activity | 获取最近活动
// @Description Get recent system activity including recent posts, comments and new users | 获取系统最近的活动，包括最近帖子、评论和新用户
// @Tags [Admin]Dashboard | [管理员]仪表盘
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.RecentActivityResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/dashboard/activity [get]
func (ctrl *DashboardController) GetRecentActivity(c *gin.Context) {
	// Invoke service | 调用服务
	result, err := ctrl.dashboardService.GetRecentActivity(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPopularPosts Get popular posts | 获取热门帖子
// @Summary Get popular posts | 获取热门帖子
// @Description Get list of popular posts with highest view counts | 获取浏览量最高的热门帖子列表
// @Tags [Admin]Dashboard | [管理员]仪表盘
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PopularPostsResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/dashboard/popular-posts [get]
func (ctrl *DashboardController) GetPopularPosts(c *gin.Context) {
	// Invoke service | 调用服务
	result, err := ctrl.dashboardService.GetPopularPosts(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPopularCategories Get popular categories | 获取热门版块
// @Summary Get popular categories | 获取热门版块
// @Description Get list of popular categories with most posts | 获取帖子数量最多的热门版块列表
// @Tags [Admin]Dashboard | [管理员]仪表盘
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.PopularCategoriesResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/dashboard/popular-categories [get]
func (ctrl *DashboardController) GetPopularCategories(c *gin.Context) {
	// Invoke service | 调用服务
	result, err := ctrl.dashboardService.GetPopularCategories(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
