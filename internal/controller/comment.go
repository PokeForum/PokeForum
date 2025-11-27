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

// CommentController 评论控制器
type CommentController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewCommentController 创建评论控制器实例
func NewCommentController(injector *do.Injector) *CommentController {
	return &CommentController{
		injector: injector,
	}
}

// CommentRouter 评论相关路由注册
func (ctrl *CommentController) CommentRouter(router *gin.RouterGroup) {
	// 发布评论
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.CreateComment)
	// 编辑评论
	router.PUT("", saGin.CheckRole(user.RoleUser.String()), ctrl.UpdateComment)
	// 获取评论列表
	router.GET("", ctrl.GetCommentList)
	// 点赞评论
	router.POST("/like", saGin.CheckRole(user.RoleUser.String()), ctrl.LikeComment)
	// 点踩评论
	router.POST("/dislike", saGin.CheckRole(user.RoleUser.String()), ctrl.DislikeComment)
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *CommentController) getUserID(c *gin.Context) (int, error) {
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

// CreateComment 创建评论
// @Summary 创建评论
// @Description 用户创建新评论，支持回复评论和回复用户
// @Tags [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentCreateRequest true "创建评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentCreateResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /comments [post]
func (ctrl *CommentController) CreateComment(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserCommentCreateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// 获取客户端IP和设备信息
	clientIP := c.ClientIP()
	deviceInfo := c.GetHeader("User-Agent")
	if deviceInfo == "" {
		deviceInfo = "Unknown"
	}

	// 调用服务创建评论
	result, err := commentService.CreateComment(c.Request.Context(), userID, clientIP, deviceInfo, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "创建评论失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// UpdateComment 更新评论
// @Summary 更新评论
// @Description 用户更新自己的评论内容
// @Tags [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentUpdateRequest true "更新评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentUpdateResponse} "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "权限不足"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /comments [put]
func (ctrl *CommentController) UpdateComment(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserCommentUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// 调用服务更新评论
	result, err := commentService.UpdateComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "更新评论失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// LikeComment 点赞评论
// @Summary 点赞评论
// @Description 用户点赞评论，单向操作不可取消
// @Tags [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentActionRequest true "点赞评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentActionResponse} "点赞成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /comments/like [post]
func (ctrl *CommentController) LikeComment(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserCommentActionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// 调用服务点赞评论
	result, err := commentService.LikeComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "点赞评论失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// DislikeComment 点踩评论
// @Summary 点踩评论
// @Description 用户点踩评论，单向操作不可取消
// @Tags [用户]评论
// @Accept json
// @Produce json
// @Param request body schema.UserCommentActionRequest true "点踩评论请求"
// @Success 200 {object} response.Data{data=schema.UserCommentActionResponse} "点踩成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /comments/dislike [post]
func (ctrl *CommentController) DislikeComment(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserCommentActionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// 调用服务点踩评论
	result, err := commentService.DislikeComment(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "点踩评论失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// GetCommentList 获取评论列表
// @Summary 获取评论列表
// @Description 分页获取指定帖子的评论列表，支持排序
// @Tags [用户]评论
// @Accept json
// @Produce json
// @Param post_id query int true "帖子ID"
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Param sort_by query string false "排序字段：created_at, like_count" example("created_at")
// @Param sort_desc query bool false "是否降序" example("true")
// @Success 200 {object} response.Data{data=schema.UserCommentListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /comments [get]
func (ctrl *CommentController) GetCommentList(c *gin.Context) {
	// 解析请求参数
	var req schema.UserCommentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	commentService := do.MustInvoke[service.ICommentService](ctrl.injector)

	// 调用服务获取评论列表
	result, err := commentService.GetCommentList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取评论列表失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}
