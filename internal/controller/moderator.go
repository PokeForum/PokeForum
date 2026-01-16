package controller

import (
	"fmt"
	"strconv"

	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// ModeratorController Moderator controller | 版主控制器
type ModeratorController struct {
	moderatorService service.IModeratorService
}

// NewModeratorController Create moderator controller instance | 创建版主控制器实例
func NewModeratorController(moderatorService service.IModeratorService) *ModeratorController {
	return &ModeratorController{
		moderatorService: moderatorService,
	}
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *ModeratorController) getUserID(c *gin.Context) (int, error) {
	// Get token from Header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("authorization header not found | 未找到Authorization header")
	}

	// Use stputil to get login user ID | 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	// String to Int | String转Int
	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return sID, nil
}

// ModeratorRouter Moderator related route registration | 版主相关路由注册
func (ctrl *ModeratorController) ModeratorRouter(router *gin.RouterGroup) {
	// Get list of categories managed by moderator | 获取版主管理的版块列表
	router.GET("/categories", ctrl.GetModeratorCategories)

	// Topic post management | 主题帖管理
	{
		// Ban post | 封禁帖子
		router.POST("/posts/ban", ctrl.BanPost)
		// Edit post | 编辑帖子
		router.PUT("/posts", ctrl.EditPost)
		// Move post | 移动帖子
		router.PUT("/posts/move", ctrl.MovePost)
		// Set post essence | 设置帖子精华
		router.PUT("/posts/essence", ctrl.SetPostEssence)
		// Lock post | 锁定帖子
		router.PUT("/posts/lock", ctrl.LockPost)
		// Pin post | 置顶帖子
		router.PUT("/posts/pin", ctrl.PinPost)
	}

	// Category management | 版块管理
	{
		// Edit category | 编辑版块
		router.PUT("/categories", ctrl.EditCategory)
		// Create category announcement | 创建版块公告
		router.POST("/categories/announcement", ctrl.CreateCategoryAnnouncement)
		// Get category announcement list | 获取版块公告列表
		router.GET("/categories/:category_id/announcements", ctrl.GetCategoryAnnouncements)
	}
}

// GetModeratorCategories Get list of categories managed by moderator | 获取版主管理的版块列表
// @Summary Get list of categories managed by moderator | 获取版主管理的版块列表
// @Description Get list of all categories that the current moderator has permission to manage | 获取当前版主有管理权限的所有版块列表
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.ModeratorCategoriesResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/categories [get]
func (ctrl *ModeratorController) GetModeratorCategories(c *gin.Context) {
	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.moderatorService.GetModeratorCategories(c.Request.Context(), userID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// BanPost Ban post | 封禁帖子
// @Summary Ban post | 封禁帖子
// @Description Moderator bans a post within specified category | 版主封禁指定版块内的帖子
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostBanRequest true "Ban information | 封禁信息"
// @Success 200 {object} response.Data "Banned successfully | 封禁成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts/ban [post]
func (ctrl *ModeratorController) BanPost(c *gin.Context) {
	var req schema.PostBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.BanPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// EditPost Edit post | 编辑帖子
// @Summary Edit post | 编辑帖子
// @Description Moderator edits post content within specified category | 版主编辑指定版块内的帖子内容
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostEditRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.ModeratorPostResponse} "Edited successfully | 编辑成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts [put]
func (ctrl *ModeratorController) EditPost(c *gin.Context) {
	var req schema.PostEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.moderatorService.EditPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// MovePost Move post | 移动帖子
// @Summary Move post | 移动帖子
// @Description Moderator moves post to other categories they have permission for | 版主将帖子移动到自己有权限的其他版块
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostMoveRequest true "Move information | 移动信息"
// @Success 200 {object} response.Data "Moved successfully | 移动成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts/move [put]
func (ctrl *ModeratorController) MovePost(c *gin.Context) {
	var req schema.PostMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.MovePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostEssence Set post essence | 设置帖子精华
// @Summary Set post essence | 设置帖子精华
// @Description Moderator sets or cancels post essence status | 版主设置或取消帖子的精华状态
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostEssenceRequest true "Essence information | 精华信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts/essence [put]
func (ctrl *ModeratorController) SetPostEssence(c *gin.Context) {
	var req schema.PostEssenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.SetPostEssence(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// LockPost Lock post | 锁定帖子
// @Summary Lock post | 锁定帖子
// @Description Moderator locks or unlocks post, users cannot reply after locking | 版主锁定或解锁帖子，锁定后用户无法回复
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostLockRequest true "Lock information | 锁定信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts/lock [put]
func (ctrl *ModeratorController) LockPost(c *gin.Context) {
	var req schema.PostLockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.LockPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// PinPost Pin post | 置顶帖子
// @Summary Pin post | 置顶帖子
// @Description Moderator pins or unpins post | 版主置顶或取消置顶帖子
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.PostPinRequest true "Pin information | 置顶信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/posts/pin [put]
func (ctrl *ModeratorController) PinPost(c *gin.Context) {
	var req schema.PostPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.PinPost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// EditCategory Edit category | 编辑版块
// @Summary Edit category | 编辑版块
// @Description Moderator edits information of categories they manage | 版主编辑自己管理的版块信息
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryEditRequest true "Category information | 版块信息"
// @Success 200 {object} response.Data "Edited successfully | 编辑成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/categories [put]
func (ctrl *ModeratorController) EditCategory(c *gin.Context) {
	var req schema.CategoryEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	err = ctrl.moderatorService.EditCategory(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// CreateCategoryAnnouncement Create category announcement | 创建版块公告
// @Summary Create category announcement | 创建版块公告
// @Description Moderator creates announcement for categories they manage | 版主为自己管理的版块创建公告
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param request body schema.CategoryAnnouncementRequest true "Announcement information | 公告信息"
// @Success 200 {object} response.Data{data=schema.CategoryAnnouncementResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/categories/announcement [post]
func (ctrl *ModeratorController) CreateCategoryAnnouncement(c *gin.Context) {
	var req schema.CategoryAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.moderatorService.CreateCategoryAnnouncement(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetCategoryAnnouncements Get category announcement list | 获取版块公告列表
// @Summary Get category announcement list | 获取版块公告列表
// @Description Get announcement list of specified category | 获取指定版块的公告列表
// @Tags [Moderator]Category Management | [版主]版块管理
// @Accept json
// @Produce json
// @Param category_id path int true "Category ID | 版块ID" example("1")
// @Success 200 {object} response.Data{data=[]schema.CategoryAnnouncementResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /moderator/categories/{category_id}/announcements [get]
func (ctrl *ModeratorController) GetCategoryAnnouncements(c *gin.Context) {
	var req struct {
		CategoryID int `uri:"category_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID (moderator identity verified by other middleware) | 获取用户ID（通过其他中间件验证版主身份）
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Invoke service | 调用服务
	result, err := ctrl.moderatorService.GetCategoryAnnouncements(c.Request.Context(), userID, req.CategoryID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
