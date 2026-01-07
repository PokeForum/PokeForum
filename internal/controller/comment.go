package controller

import (
	"fmt"
	"strconv"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// CommentController Comment controller | 评论控制器
type CommentController struct {
	// Injector instance for retrieving services | 注入器实例,用于获取服务
	injector *do.Injector
}

// NewCommentController Create comment controller instance | 创建评论控制器实例
func NewCommentController(injector *do.Injector) *CommentController {
	return &CommentController{
		injector: injector,
	}
}

// CommentRouter Comment related route registration | 评论相关路由注册
func (ctrl *CommentController) CommentRouter(router *gin.RouterGroup) {
	// Publish comment | 发布评论
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.CreateComment)
	// Edit comment | 编辑评论
	router.PUT("", saGin.CheckRole(user.RoleUser.String()), ctrl.UpdateComment)
	// Get comment list | 获取评论列表
	router.GET("", ctrl.GetCommentList)
	// Like comment | 点赞评论
	router.POST("/like", saGin.CheckRole(user.RoleUser.String()), ctrl.LikeComment)
	// Dislike comment | 点踩评论
	router.POST("/dislike", saGin.CheckRole(user.RoleUser.String()), ctrl.DislikeComment)
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *CommentController) getUserID(c *gin.Context) (int, error) {
	// Get token from Header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("Authorization header not found | 未找到Authorization header")
	}

	// Get logged-in user ID using stputil | 使用stputil获取登录用户ID
	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		return 0, err
	}

	// Convert String to Int | String转Int
	sID, err := strconv.Atoi(loginID)
	if err != nil {
		return 0, err
	}

	return sID, nil
}

// CreateComment Create comment | 创建评论
// @Summary Create comment | 创建评论
// @Description User creates new comment, supports replying to comments and users | 用户创建新评论,支持回复评论和回复用户
// @Tags [User]Comments | [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentCreateRequest true "Create comment request | 创建评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentCreateResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /comments [post]
func (ctrl *CommentController) CreateComment(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserCommentCreateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get service instance | 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// Get client IP and device info | 获取客户端IP和设备信息
	clientIP := c.ClientIP()
	deviceInfo := c.GetHeader("User-Agent")
	if deviceInfo == "" {
		deviceInfo = "Unknown"
	}

	// Call service to create comment | 调用服务创建评论
	result, err := commentService.CreateComment(c.Request.Context(), userID, clientIP, deviceInfo, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "Failed to create comment | 创建评论失败", err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}

// UpdateComment Update comment | 更新评论
// @Summary Update comment | 更新评论
// @Description User updates their own comment content | 用户更新自己的评论内容
// @Tags [User]Comments | [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentUpdateRequest true "Update comment request | 更新评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentUpdateResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Insufficient permissions | 权限不足"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /comments [put]
func (ctrl *CommentController) UpdateComment(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserCommentUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get service instance | 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// Call service to update comment | 调用服务更新评论
	result, err := commentService.UpdateComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "Failed to update comment | 更新评论失败", err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}

// LikeComment Like comment | 点赞评论
// @Summary Like comment | 点赞评论
// @Description User likes comment, one-way operation cannot be cancelled | 用户点赞评论,单向操作不可取消
// @Tags [User]Comments | [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentActionRequest true "Like comment request | 点赞评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentActionResponse} "Liked successfully | 点赞成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /comments/like [post]
func (ctrl *CommentController) LikeComment(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserCommentActionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get service instance | 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// Call service to like comment | 调用服务点赞评论
	result, err := commentService.LikeComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "Failed to like comment | 点赞评论失败", err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}

// DislikeComment Dislike comment | 点踩评论
// @Summary Dislike comment | 点踩评论
// @Description User dislikes comment, one-way operation cannot be cancelled | 用户点踩评论,单向操作不可取消
// @Tags [User]Comments | [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentActionRequest true "Dislike comment request | 点踩评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentActionResponse} "Disliked successfully | 点踩成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /comments/dislike [post]
func (ctrl *CommentController) DislikeComment(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse request parameters | 解析请求参数
	var req schema.UserCommentActionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get service instance | 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// Call service to dislike comment | 调用服务点踩评论
	result, err := commentService.DislikeComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "Failed to dislike comment | 点踩评论失败", err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}

// GetCommentList Get comment list | 获取评论列表
// @Summary Get comment list | 获取评论列表
// @Description Get paginated comment list for specified post with sorting support | 分页获取指定帖子的评论列表,支持排序
// @Tags [User]Comments | [用户]评论
// @Accept json
// @Produce json
// @Param post_id query int true "Post ID | 帖子ID"
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param sort_by query string false "Sort field: created_at, like_count | 排序字段:created_at, like_count" example("created_at")
// @Param sort_desc query bool false "Is descending order | 是否降序" example("true")
// @Success 200 {object} response.Data{data=schema.UserCommentListResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /comments [get]
func (ctrl *CommentController) GetCommentList(c *gin.Context) {
	// Parse request parameters | 解析请求参数
	var req schema.UserCommentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Get service instance | 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// Call service to get comment list | 调用服务获取评论列表
	result, err := commentService.GetCommentList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "Failed to get comment list | 获取评论列表失败", err.Error())
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}
