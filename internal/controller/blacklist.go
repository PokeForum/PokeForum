package controller

import (
	"fmt"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// BlacklistController 黑名单控制器
type BlacklistController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewBlacklistController 创建黑名单控制器实例
func NewBlacklistController(injector *do.Injector) *BlacklistController {
	return &BlacklistController{
		injector: injector,
	}
}

// BlacklistRouter 黑名单相关路由注册
func (ctrl *BlacklistController) BlacklistRouter(router *gin.RouterGroup) {
	router.Use(saGin.CheckRole(user.RoleUser.String()))

	// 获取黑名单列表
	router.GET("/list", ctrl.GetBlacklistList)
	// 添加用户到黑名单
	router.POST("/add", ctrl.AddToBlacklist)
	// 从黑名单移除用户
	router.DELETE("/remove", ctrl.RemoveFromBlacklist)
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *BlacklistController) getUserID(c *gin.Context) (int, error) {
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

// GetBlacklistList 获取用户黑名单列表
// @Summary 获取用户黑名单列表
// @Description 获取当前用户的黑名单列表，支持分页
// @Tags [用户]黑名单管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Data{data=schema.UserBlacklistListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/blacklist/list [get]
func (ctrl *BlacklistController) GetBlacklistList(c *gin.Context) {
	// 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "获取用户信息失败", err.Error())
		return
	}

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "页码参数错误", "page必须是正整数")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 50 {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "每页数量参数错误", "page_size必须是1-50之间的整数")
		return
	}

	// 获取黑名单服务
	blacklistService := do.MustInvoke[service.IBlacklistService](ctrl.injector)

	// 调用服务获取黑名单列表
	result, err := blacklistService.GetUserBlacklist(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取黑名单列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// AddToBlacklist 添加用户到黑名单
// @Summary 添加用户到黑名单
// @Description 将指定用户添加到当前用户的黑名单中
// @Tags [用户]黑名单管理
// @Accept json
// @Produce json
// @Param request body schema.UserBlacklistAddRequest true "添加黑名单请求"
// @Success 200 {object} response.Data{data=schema.UserBlacklistAddResponse} "添加成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "禁止操作"
// @Failure 404 {object} response.Data "用户不存在"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/blacklist/add [post]
func (ctrl *BlacklistController) AddToBlacklist(c *gin.Context) {
	// 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserBlacklistAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	// 不能拉黑自己
	if req.BlockedUserID == userID {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "禁止操作", "不能将自己添加到黑名单")
		return
	}

	// 获取黑名单服务
	blacklistService := do.MustInvoke[service.IBlacklistService](ctrl.injector)

	// 调用服务添加黑名单
	result, err := blacklistService.AddToBlacklist(c.Request.Context(), userID, req.BlockedUserID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "添加黑名单失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// RemoveFromBlacklist 从黑名单移除用户
// @Summary 从黑名单移除用户
// @Description 将指定用户从当前用户的黑名单中移除
// @Tags [用户]黑名单管理
// @Accept json
// @Produce json
// @Param request body schema.UserBlacklistRemoveRequest true "移除黑名单请求"
// @Success 200 {object} response.Data{data=schema.UserBlacklistRemoveResponse} "移除成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 404 {object} response.Data "黑名单记录不存在"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/blacklist/remove [delete]
func (ctrl *BlacklistController) RemoveFromBlacklist(c *gin.Context) {
	// 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserBlacklistRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	// 获取黑名单服务
	blacklistService := do.MustInvoke[service.IBlacklistService](ctrl.injector)

	// 调用服务移除黑名单
	err = blacklistService.RemoveFromBlacklist(c.Request.Context(), userID, req.BlockedUserID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "移除黑名单失败", err.Error())
		return
	}

	result := schema.UserBlacklistRemoveResponse{
		Message: "移除黑名单成功",
	}

	response.ResSuccess(c, result)
}
