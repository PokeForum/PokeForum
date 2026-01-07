package controller

import (
	"fmt"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// BlacklistController Blacklist controller | 黑名单控制器
type BlacklistController struct {
	blacklistService service.IBlacklistService
}

// NewBlacklistController Create blacklist controller instance | 创建黑名单控制器实例
func NewBlacklistController(blacklistService service.IBlacklistService) *BlacklistController {
	return &BlacklistController{
		blacklistService: blacklistService,
	}
}

// BlacklistRouter Blacklist related route registration | 黑名单相关路由注册
func (ctrl *BlacklistController) BlacklistRouter(router *gin.RouterGroup) {
	router.Use(saGin.CheckRole(user.RoleUser.String()))

	// Get blacklist list | 获取黑名单列表
	router.GET("/list", ctrl.GetBlacklistList)
	// Add user to blacklist | 添加用户到黑名单
	router.POST("/add", ctrl.AddToBlacklist)
	// Remove user from blacklist | 从黑名单移除用户
	router.DELETE("/remove", ctrl.RemoveFromBlacklist)
}

// getUserID Get token from header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *BlacklistController) getUserID(c *gin.Context) (int, error) {
	// Get token from header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("authorization header not found | 未找到Authorization header")
	}

	// Use stputil to get logged-in user ID | 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	// Convert string to int | String转Int
	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return sID, nil
}

// GetBlacklistList Get user blacklist list | 获取用户黑名单列表
// @Summary Get user blacklist list | 获取用户黑名单列表
// @Description Get the current user's blacklist, supports pagination | 获取当前用户的黑名单列表,支持分页
// @Tags [User]Blacklist Management | [用户]黑名单管理
// @Accept json
// @Produce json
// @Param page query int false "Page number | 页码" default(1)
// @Param page_size query int false "Items per page | 每页数量" default(20)
// @Success 200 {object} response.Data{data=schema.UserBlacklistListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/blacklist/list [get]
func (ctrl *BlacklistController) GetBlacklistList(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败", err.Error())
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
	result, err := ctrl.blacklistService.GetUserBlacklist(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get blacklist | 获取黑名单列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// AddToBlacklist Add user to blacklist | 添加用户到黑名单
// @Summary Add user to blacklist | 添加用户到黑名单
// @Description Add a specified user to the current user's blacklist | 将指定用户添加到当前用户的黑名单中
// @Tags [User]Blacklist Management | [用户]黑名单管理
// @Accept json
// @Produce json
// @Param request body schema.UserBlacklistAddRequest true "Add to blacklist request | 添加黑名单请求"
// @Success 200 {object} response.Data{data=schema.UserBlacklistAddResponse} "Added successfully | 添加成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Forbidden operation | 禁止操作"
// @Failure 404 {object} response.Data "User not found | 用户不存在"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/blacklist/add [post]
func (ctrl *BlacklistController) AddToBlacklist(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserBlacklistAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Cannot block yourself | 不能拉黑自己
	if req.BlockedUserID == userID {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Forbidden operation | 禁止操作", "Cannot add yourself to blacklist | 不能将自己添加到黑名单")
		return
	}

	// Call service to add to blacklist | 调用服务添加黑名单
	result, err := ctrl.blacklistService.AddToBlacklist(c.Request.Context(), userID, req.BlockedUserID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to add to blacklist | 添加黑名单失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// RemoveFromBlacklist Remove user from blacklist | 从黑名单移除用户
// @Summary Remove user from blacklist | 从黑名单移除用户
// @Description Remove a specified user from the current user's blacklist | 将指定用户从当前用户的黑名单中移除
// @Tags [User]Blacklist Management | [用户]黑名单管理
// @Accept json
// @Produce json
// @Param request body schema.UserBlacklistRemoveRequest true "Remove from blacklist request | 移除黑名单请求"
// @Success 200 {object} response.Data{data=schema.UserBlacklistRemoveResponse} "Removed successfully | 移除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 404 {object} response.Data "Blacklist record not found | 黑名单记录不存在"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/blacklist/remove [delete]
func (ctrl *BlacklistController) RemoveFromBlacklist(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserBlacklistRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Call service to remove from blacklist | 调用服务移除黑名单
	err = ctrl.blacklistService.RemoveFromBlacklist(c.Request.Context(), userID, req.BlockedUserID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to remove from blacklist | 移除黑名单失败", err.Error())
		return
	}

	result := schema.UserBlacklistRemoveResponse{
		Message: "Removed from blacklist successfully | 移除黑名单成功",
	}

	response.ResSuccess(c, result)
}
