package controller

import (
	"net/http"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// ReportController Report controller | 举报控制器
type ReportController struct {
	BaseController
	reportService       service.IReportService
	reportManageService service.IReportManageService
}

// NewReportController Create report controller instance | 创建举报控制器实例
func NewReportController(reportService service.IReportService, reportManageService service.IReportManageService) *ReportController {
	return &ReportController{
		reportService:       reportService,
		reportManageService: reportManageService,
	}
}

// ReportRouter User-side report routes | 用户侧举报路由
func (ctrl *ReportController) ReportRouter(router *gin.RouterGroup) {
	// Require login | 需要登录
	router.Use(saGin.CheckRole(user.RoleUser.String()))

	// Create report | 创建举报
	router.POST("", ctrl.CreateReport)
	// Get my reports | 获取我的举报列表
	router.GET("/my", ctrl.GetMyReports)
}

// ReportManageRouter Admin report management routes | 管理员举报管理路由
func (ctrl *ReportController) ReportManageRouter(router *gin.RouterGroup) {
	// Require admin role | 需要管理员角色
	router.Use(saGin.CheckRole(user.RoleAdmin.String()))

	// Get report list | 获取举报列表
	router.GET("/list", ctrl.GetReportList)
	// Get report statistics | 获取举报统计
	router.GET("/stats", ctrl.GetReportStats)
	// Handle report | 处理举报
	router.POST("/handle", ctrl.HandleReport)
	// Batch handle reports | 批量处理举报
	router.POST("/batch-handle", ctrl.BatchHandleReports)
}

// CreateReport Create report | 创建举报
// @Summary Create report | 创建举报
// @Description Report a post, comment or user | 举报帖子、评论或用户
// @Tags [User]Report | [用户]举报
// @Accept json
// @Produce json
// @Param request body schema.ReportCreateRequest true "Create report request | 创建举报请求"
// @Success 200 {object} response.Data{data=schema.ReportCreateResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden operation | 禁止操作"
// @Failure 404 {object} response.Data "Target not found | 目标不存在"
// @Failure 409 {object} response.Data "Already reported | 已举报过"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /reports [post]
func (ctrl *ReportController) CreateReport(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID, err := ctrl.GetUserID(c)
	if err != nil {
		response.ResErrorWithHTTPStatus(c, response.ResCodeToHTTPStatus(response.CodeNeedLogin), response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败")
		return
	}

	// Parse request | 解析请求
	var req schema.ReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.reportService.CreateReport(c.Request.Context(), userID, &req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to create report | 创建举报失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetMyReports Get current user's reports | 获取当前用户的举报列表
// @Summary Get my reports | 获取我的举报列表
// @Description Get the current user's report history | 获取当前用户的举报历史
// @Tags [User]Report | [用户]举报
// @Accept json
// @Produce json
// @Param page query int false "Page number | 页码" default(1)
// @Param page_size query int false "Items per page | 每页数量" default(20)
// @Success 200 {object} response.Data{data=schema.UserReportListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /reports/my [get]
func (ctrl *ReportController) GetMyReports(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID, err := ctrl.GetUserID(c)
	if err != nil {
		response.ResErrorWithHTTPStatus(c, response.ResCodeToHTTPStatus(response.CodeNeedLogin), response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败")
		return
	}

	// Get query parameters | 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid page parameter | 页码参数错误", "page must be a positive integer | page必须是正整数")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 50 {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid page_size parameter | 每页数量参数错误", "page_size must be an integer between 1-50 | page_size必须是1-50之间的整数")
		return
	}

	// Call service | 调用服务
	result, err := ctrl.reportService.GetUserReports(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get reports | 获取举报列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetReportList Get report list (admin) | 获取举报列表（管理员）
// @Summary Get report list | 获取举报列表
// @Description Get report list with filters for administrators | 管理员获取带筛选的举报列表
// @Tags [Admin]Report Management | [管理员]举报管理
// @Accept json
// @Produce json
// @Param page query int false "Page number | 页码" default(1)
// @Param page_size query int false "Items per page | 每页数量" default(20)
// @Param status query string false "Status filter | 状态筛选" Enums(pending,processing,resolved,ignored)
// @Param target_type query string false "Target type filter | 目标类型筛选" Enums(post,comment,user)
// @Param report_type query string false "Report type filter | 举报类型筛选"
// @Param reporter_id query int false "Reporter ID filter | 举报者ID筛选"
// @Param target_author_id query int false "Target author ID filter | 被举报作者ID筛选"
// @Success 200 {object} response.Data{data=schema.ReportListResponse} "Success | 获取成功"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden | 无权限"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /manage/reports/list [get]
func (ctrl *ReportController) GetReportList(c *gin.Context) {
	var query schema.ReportListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid query parameters | 查询参数错误", err.Error())
		return
	}

	// Set default values | 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	// Call service | 调用服务
	result, err := ctrl.reportManageService.GetReportList(c.Request.Context(), &query)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get report list | 获取举报列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetReportStats Get report statistics | 获取举报统计
// @Summary Get report statistics | 获取举报统计
// @Description Get report statistics for dashboard | 获取举报统计数据用于仪表盘
// @Tags [Admin]Report Management | [管理员]举报管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.ReportStatsResponse} "Success | 获取成功"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden | 无权限"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /manage/reports/stats [get]
func (ctrl *ReportController) GetReportStats(c *gin.Context) {
	result, err := ctrl.reportManageService.GetReportStats(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get report stats | 获取举报统计失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// HandleReport Handle report | 处理举报
// @Summary Handle report | 处理举报
// @Description Handle a specific report (delete/warn/ban/ignore) | 处理指定举报（删除/警告/封禁/忽略）
// @Tags [Admin]Report Management | [管理员]举报管理
// @Accept json
// @Produce json
// @Param request body schema.ReportHandleRequest true "Handle report request | 处理举报请求"
// @Success 200 {object} response.Data{data=schema.ReportHandleResponse} "Success | 处理成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden | 无权限"
// @Failure 404 {object} response.Data "Report not found | 举报不存在"
// @Failure 409 {object} response.Data "Already handled | 已被处理"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /manage/reports/handle [post]
func (ctrl *ReportController) HandleReport(c *gin.Context) {
	// Get handler ID | 获取处理人ID
	handlerID, err := ctrl.GetUserID(c)
	if err != nil {
		response.ResErrorWithHTTPStatus(c, response.ResCodeToHTTPStatus(response.CodeNeedLogin), response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败")
		return
	}

	// Parse request | 解析请求
	var req schema.ReportHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.reportManageService.HandleReport(c.Request.Context(), handlerID, &req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to handle report | 处理举报失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// BatchHandleReports Batch handle reports | 批量处理举报
// @Summary Batch handle reports | 批量处理举报
// @Description Batch handle multiple reports at once | 一次性批量处理多个举报
// @Tags [Admin]Report Management | [管理员]举报管理
// @Accept json
// @Produce json
// @Param request body schema.ReportBatchHandleRequest true "Batch handle request | 批量处理请求"
// @Success 200 {object} response.Data "Success | 处理成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden | 无权限"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /manage/reports/batch-handle [post]
func (ctrl *ReportController) BatchHandleReports(c *gin.Context) {
	// Get handler ID | 获取处理人ID
	handlerID, err := ctrl.GetUserID(c)
	if err != nil {
		response.ResErrorWithHTTPStatus(c, http.StatusUnauthorized, response.CodeNeedLogin, "Unauthorized | 未授权")
		return
	}

	// Parse request | 解析请求
	var req schema.ReportBatchHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Call service | 调用服务
	err = ctrl.reportManageService.BatchHandleReports(c.Request.Context(), handlerID, &req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to batch handle reports | 批量处理举报失败", err.Error())
		return
	}

	response.ResSuccess(c, gin.H{"message": "Batch handled successfully | 批量处理成功"})
}
