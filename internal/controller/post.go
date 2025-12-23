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

// PostController 帖子控制器
type PostController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewPostController 创建帖子控制器实例
func NewPostController(injector *do.Injector) *PostController {
	return &PostController{
		injector: injector,
	}
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *PostController) getUserID(c *gin.Context) (int, error) {
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

// PostRouter 帖子相关路由注册
func (ctrl *PostController) PostRouter(router *gin.RouterGroup) {
	// 发布新帖
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.CreatePost)
	// 保存草稿
	router.POST("/draft", saGin.CheckRole(user.RoleUser.String()), ctrl.SaveDraft)
	// 编辑帖子
	router.PUT("", saGin.CheckRole(user.RoleUser.String()), ctrl.UpdatePost)
	// 设置帖子私有
	router.PUT("/private", saGin.CheckRole(user.RoleUser.String()), ctrl.SetPostPrivate)
	// 点赞帖子
	router.POST("/like", saGin.CheckRole(user.RoleUser.String()), ctrl.LikePost)
	// 点踩帖子
	router.POST("/dislike", saGin.CheckRole(user.RoleUser.String()), ctrl.DislikePost)
	// 收藏帖子
	router.POST("/favorite", saGin.CheckRole(user.RoleUser.String()), ctrl.FavoritePost)
	// 获取帖子列表
	router.GET("", ctrl.GetPostList)
	// 获取帖子详情
	router.GET("/:id", ctrl.GetPostDetail)
}

// CreatePost 发布新帖
// @Summary 发布新帖
// @Description 用户发布新的主题帖
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostCreateRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostCreateResponse} "发布成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts [post]
func (ctrl *PostController) CreatePost(c *gin.Context) {
	var req schema.UserPostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.CreatePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// SaveDraft 保存草稿
// @Summary 保存草稿
// @Description 用户保存帖子草稿
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostCreateRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostCreateResponse} "保存成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/draft [post]
func (ctrl *PostController) SaveDraft(c *gin.Context) {
	var req schema.UserPostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.SaveDraft(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// UpdatePost 编辑帖子
// @Summary 编辑帖子
// @Description 用户编辑自己的帖子（每三分钟可操作一次）
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostUpdateRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostUpdateResponse} "编辑成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 403 {object} response.Data "权限不足或操作过于频繁"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts [put]
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	var req schema.UserPostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.UpdatePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// SetPostPrivate 设置帖子私有
// @Summary 设置帖子私有
// @Description 用户设置帖子为私有或公开（每三日可操作一次）
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 403 {object} response.Data "权限不足或操作过于频繁"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/private [put]
func (ctrl *PostController) SetPostPrivate(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.SetPostPrivate(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// LikePost 点赞帖子
// @Summary 点赞帖子
// @Description 用户点赞帖子（单向，不可取消点赞）
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "点赞成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 403 {object} response.Data "已经点赞过"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/like [post]
func (ctrl *PostController) LikePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.LikePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DislikePost 点踩帖子
// @Summary 点踩帖子
// @Description 用户点踩帖子（单向，不可取消点踩）
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "点踩成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 403 {object} response.Data "已经点踩过"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/dislike [post]
func (ctrl *PostController) DislikePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.DislikePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// FavoritePost 收藏帖子
// @Summary 收藏帖子
// @Description 用户收藏或取消收藏帖子（双向操作）
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "操作成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未登录"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/favorite [post]
func (ctrl *PostController) FavoritePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.FavoritePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPostList 获取帖子列表
// @Summary 获取帖子列表
// @Description 获取帖子列表，支持分页和排序
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param category_id query int false "版块ID"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20，最大100" default(20)
// @Param sort query string false "排序方式：latest(最新)、hot(热门)、essence(精华)" default(latest)
// @Success 200 {object} response.Data{data=schema.UserPostListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts [get]
func (ctrl *PostController) GetPostList(c *gin.Context) {
	var req schema.UserPostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.GetPostList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPostDetail 获取帖子详情
// @Summary 获取帖子详情
// @Description 获取指定帖子的详细信息，并增加浏览数
// @Tags [用户]主题贴
// @Accept json
// @Produce json
// @Param id path int true "帖子ID"
// @Success 200 {object} response.Data{data=schema.UserPostDetailResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 404 {object} response.Data "帖子不存在"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /posts/{id} [get]
func (ctrl *PostController) GetPostDetail(c *gin.Context) {
	var req schema.UserPostDetailRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postService, err := do.Invoke[service.IPostService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postService.GetPostDetail(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
