package controller

import (
	"fmt"
	"strconv"

	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// ModeratorController 版主控制器
type ModeratorController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewModeratorController 创建版主控制器实例
func NewModeratorController(injector *do.Injector) *ModeratorController {
	return &ModeratorController{
		injector: injector,
	}
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *ModeratorController) getUserID(c *gin.Context) (int, error) {
	// 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("未找到Authorization header")
	}

	// 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	// String转Int
	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return sID, nil
}

// ModeratorRouter 版主相关路由注册
func (ctrl *ModeratorController) ModeratorRouter(router *gin.RouterGroup) {
	// 获取版主管理的版块列表
	router.GET("/categories", ctrl.GetModeratorCategories)

	// 主题帖管理
	{
		// 封禁帖子
		router.POST("/posts/ban", ctrl.BanPost)
		// 编辑帖子
		router.PUT("/posts", ctrl.EditPost)
		// 移动帖子
		router.PUT("/posts/move", ctrl.MovePost)
		// 设置帖子精华
		router.PUT("/posts/essence", ctrl.SetPostEssence)
		// 锁定帖子
		router.PUT("/posts/lock", ctrl.LockPost)
		// 置顶帖子
		router.PUT("/posts/pin", ctrl.PinPost)
	}

	// 版块管理
	{
		// 编辑版块
		router.PUT("/categories", ctrl.EditCategory)
		// 创建版块公告
		router.POST("/categories/announcement", ctrl.CreateCategoryAnnouncement)
		// 获取版块公告列表
		router.GET("/categories/:category_id/announcements", ctrl.GetCategoryAnnouncements)
	}
}

// GetModeratorCategories 获取版主管理的版块列表
// @Summary 获取版主管理的版块列表
// @Description 获取当前版主有管理权限的所有版块列表
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.ModeratorCategoriesResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/categories [get]
func (ctrl *ModeratorController) GetModeratorCategories(c *gin.Context) {
	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := moderatorService.GetModeratorCategories(c.Request.Context(), userID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// BanPost 封禁帖子
// @Summary 封禁帖子
// @Description 版主封禁指定版块内的帖子
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostBanRequest true "封禁信息"
// @Success 200 {object} response.Data "封禁成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts/ban [post]
func (ctrl *ModeratorController) BanPost(c *gin.Context) {
	var req schema.PostBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.BanPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// EditPost 编辑帖子
// @Summary 编辑帖子
// @Description 版主编辑指定版块内的帖子内容
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostEditRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.ModeratorPostResponse} "编辑成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts [put]
func (ctrl *ModeratorController) EditPost(c *gin.Context) {
	var req schema.PostEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := moderatorService.EditPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// MovePost 移动帖子
// @Summary 移动帖子
// @Description 版主将帖子移动到自己有权限的其他版块
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostMoveRequest true "移动信息"
// @Success 200 {object} response.Data "移动成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts/move [put]
func (ctrl *ModeratorController) MovePost(c *gin.Context) {
	var req schema.PostMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.MovePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostEssence 设置帖子精华
// @Summary 设置帖子精华
// @Description 版主设置或取消帖子的精华状态
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostEssenceRequest true "精华信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts/essence [put]
func (ctrl *ModeratorController) SetPostEssence(c *gin.Context) {
	var req schema.PostEssenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.SetPostEssence(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// LockPost 锁定帖子
// @Summary 锁定帖子
// @Description 版主锁定或解锁帖子，锁定后用户无法回复
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostLockRequest true "锁定信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts/lock [put]
func (ctrl *ModeratorController) LockPost(c *gin.Context) {
	var req schema.PostLockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.LockPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// PinPost 置顶帖子
// @Summary 置顶帖子
// @Description 版主置顶或取消置顶帖子
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostPinRequest true "置顶信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/posts/pin [put]
func (ctrl *ModeratorController) PinPost(c *gin.Context) {
	var req schema.PostPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.PinPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// EditCategory 编辑版块
// @Summary 编辑版块
// @Description 版主编辑自己管理的版块信息
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryEditRequest true "版块信息"
// @Success 200 {object} response.Data "编辑成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/categories [put]
func (ctrl *ModeratorController) EditCategory(c *gin.Context) {
	var req schema.CategoryEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = moderatorService.EditCategory(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// CreateCategoryAnnouncement 创建版块公告
// @Summary 创建版块公告
// @Description 版主为自己管理的版块创建公告
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryAnnouncementRequest true "公告信息"
// @Success 200 {object} response.Data{data=schema.CategoryAnnouncementResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/categories/announcement [post]
func (ctrl *ModeratorController) CreateCategoryAnnouncement(c *gin.Context) {
	var req schema.CategoryAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := moderatorService.CreateCategoryAnnouncement(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetCategoryAnnouncements 获取版块公告列表
// @Summary 获取版块公告列表
// @Description 获取指定版块的公告列表
// @Tags [版主]版块管理
// @Accept json
// @Produce json
// @Param category_id path int true "版块ID" example("1")
// @Success 200 {object} response.Data{data=[]schema.CategoryAnnouncementResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /moderator/categories/{category_id}/announcements [get]
func (ctrl *ModeratorController) GetCategoryAnnouncements(c *gin.Context) {
	var req struct {
		CategoryID int `uri:"category_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	moderatorService, err := do.Invoke[service.IModeratorService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := moderatorService.GetCategoryAnnouncements(c.Request.Context(), userID, req.CategoryID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
