package controller

import (
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// UserFollowController User follow controller | 用户关注控制器
type UserFollowController struct {
	followService service.IUserFollowService
}

// NewUserFollowController Create user follow controller instance | 创建用户关注控制器实例
func NewUserFollowController(followService service.IUserFollowService) *UserFollowController {
	return &UserFollowController{
		followService: followService,
	}
}

// UserFollowRouter User follow related route registration | 用户关注相关路由注册
func (ctrl *UserFollowController) UserFollowRouter(router *gin.RouterGroup) {
	// Authenticated routes - require login | 认证路由 - 需要登录
	router.Use(saGin.CheckRole(user.RoleUser.String()))
	// Follow user | 关注用户
	router.POST("", ctrl.FollowUser)
	// Unfollow user | 取消关注用户
	router.DELETE("/:user_id", ctrl.UnfollowUser)
	// Get followers list | 获取粉丝列表
	router.GET("/followers", ctrl.GetFollowers)
	// Get following list | 获取关注列表
	router.GET("/following", ctrl.GetFollowing)
	// Get follow status | 获取关注状态
	router.GET("/status/:user_id", ctrl.GetFollowStatus)
}

// getCurrentUserID Get current user ID from context, return 0 if not logged in | 从上下文获取当前用户ID，未登录返回0
func (ctrl *UserFollowController) getCurrentUserID(c *gin.Context) int {
	// Get token from Header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		// Guest mode, return 0 | 游客模式，返回0
		return 0
	}

	// Use stputil to get logged-in user ID | 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0
	}

	// Convert string to int | String转Int
	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0
	}

	return sID
}

// FollowUser Follow user | 关注用户
// @Summary Follow user | 关注用户
// @Description Follow a specified user | 关注指定用户
// @Tags [User] User Follow | [用户个人中心] 用户关注
// @Accept json
// @Produce json
// @Param request body schema.UserFollowRequest true "Follow user request | 关注用户请求"
// @Success 200 {object} response.Data{data=schema.UserFollowResponse} "Followed successfully | 关注成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/follow [post]
func (ctrl *UserFollowController) FollowUser(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID := ctrl.getCurrentUserID(c)
	if userID == 0 {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Please login first | 请先登录", "")
		return
	}

	var req schema.UserFollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	result, err := ctrl.followService.FollowUser(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to follow user | 关注用户失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// UnfollowUser Unfollow user | 取消关注用户
// @Summary Unfollow user | 取消关注用户
// @Description Unfollow a specified user | 取消关注指定用户
// @Tags [User] User Follow | [用户个人中心] 用户关注
// @Accept json
// @Produce json
// @Param user_id path int true "User ID to unfollow | 要取消关注的用户ID"
// @Success 200 {object} response.Data{data=schema.UserUnfollowResponse} "Unfollowed successfully | 取消关注成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/follow/{user_id} [delete]
func (ctrl *UserFollowController) UnfollowUser(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID := ctrl.getCurrentUserID(c)
	if userID == 0 {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Please login first | 请先登录", "")
		return
	}

	// Get user_id from path parameter | 从路径参数获取user_id
	userIDStr := c.Param("user_id")
	followingID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid user_id parameter | user_id参数错误", "user_id must be an integer | user_id必须是整数")
		return
	}

	req := schema.UserUnfollowRequest{
		FollowingID: followingID,
	}

	result, err := ctrl.followService.UnfollowUser(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to unfollow user | 取消关注用户失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetFollowers Get user followers list | 获取用户粉丝列表
// @Summary Get user followers list | 获取用户粉丝列表
// @Description Get the specified user's followers list, supports pagination. If user_id is not provided, returns current user's followers | 获取指定用户的粉丝列表，支持分页。如果不提供user_id，则返回当前登录用户的粉丝列表
// @Tags [User] User Follow | [用户个人中心] 用户关注
// @Accept json
// @Produce json
// @Param user_id query int false "User ID, default to current user | 用户ID，默认为当前用户"
// @Param page query int false "Page number | 页码" default(1)
// @Param page_size query int false "Items per page | 每页数量" default(20)
// @Success 200 {object} response.Data{data=schema.UserFollowersResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/follow/followers [get]
func (ctrl *UserFollowController) GetFollowers(c *gin.Context) {
	var req schema.UserFollowersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get current user ID | 获取当前用户ID
	currentUserID := ctrl.getCurrentUserID(c)
	if currentUserID == 0 {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Please login first | 请先登录", "")
		return
	}

	// Set default values | 设置默认值
	if req.UserID == 0 {
		// If user_id not provided, use current user | 如果未提供user_id，使用当前用户
		req.UserID = currentUserID
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := ctrl.followService.GetFollowers(c.Request.Context(), currentUserID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get followers list | 获取粉丝列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetFollowing Get user following list | 获取用户关注列表
// @Summary Get user following list | 获取用户关注列表
// @Description Get the specified user's following list, supports pagination. If user_id is not provided, returns current user's following | 获取指定用户的关注列表，支持分页。如果不提供user_id，则返回当前登录用户的关注列表
// @Tags [User] User Follow | [用户个人中心] 用户关注
// @Accept json
// @Produce json
// @Param user_id query int false "User ID, default to current user | 用户ID，默认为当前用户"
// @Param page query int false "Page number | 页码" default(1)
// @Param page_size query int false "Items per page | 每页数量" default(20)
// @Success 200 {object} response.Data{data=schema.UserFollowingResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/follow/following [get]
func (ctrl *UserFollowController) GetFollowing(c *gin.Context) {
	var req schema.UserFollowingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get current user ID | 获取当前用户ID
	currentUserID := ctrl.getCurrentUserID(c)
	if currentUserID == 0 {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Please login first | 请先登录", "")
		return
	}

	// Set default values | 设置默认值
	if req.UserID == 0 {
		// If user_id not provided, use current user | 如果未提供user_id，使用当前用户
		req.UserID = currentUserID
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := ctrl.followService.GetFollowing(c.Request.Context(), currentUserID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get following list | 获取关注列表失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetFollowStatus Get follow status between two users | 获取两个用户之间的关注状态
// @Summary Get follow status | 获取关注状态
// @Description Get the follow status between current user and specified user | 获取当前用户与指定用户之间的关注状态
// @Tags [User] User Follow | [用户个人中心] 用户关注
// @Accept json
// @Produce json
// @Param user_id path int true "Target user ID | 目标用户ID"
// @Success 200 {object} response.Data{data=schema.UserFollowStatusResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /profile/follow/status/{user_id} [get]
func (ctrl *UserFollowController) GetFollowStatus(c *gin.Context) {
	// Get current user ID | 获取当前用户ID
	userID := ctrl.getCurrentUserID(c)
	if userID == 0 {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "Please login first | 请先登录", "")
		return
	}

	// Get target user ID from path parameter | 从路径参数获取目标用户ID
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid user_id parameter | user_id参数错误", "user_id must be an integer | user_id必须是整数")
		return
	}

	result, err := ctrl.followService.GetFollowStatus(c.Request.Context(), userID, targetUserID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "Failed to get follow status | 获取关注状态失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}
