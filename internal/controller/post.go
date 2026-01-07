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

// PostController Post Controller | 帖子控制器
type PostController struct {
	postService service.IPostService
}

// NewPostController Create post controller instance | 创建帖子控制器实例
func NewPostController(injector *do.Injector) *PostController {
	return &PostController{
		postService: do.MustInvoke[service.IPostService](injector),
	}
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *PostController) getUserID(c *gin.Context) (int, error) {
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

// PostRouter Post-related route registration | 帖子相关路由注册
func (ctrl *PostController) PostRouter(router *gin.RouterGroup) {
	// Publish new post | 发布新帖
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.CreatePost)
	// Save draft | 保存草稿
	router.POST("/draft", saGin.CheckRole(user.RoleUser.String()), ctrl.SaveDraft)
	// Edit post | 编辑帖子
	router.PUT("", saGin.CheckRole(user.RoleUser.String()), ctrl.UpdatePost)
	// Set post as private | 设置帖子私有
	router.PUT("/private", saGin.CheckRole(user.RoleUser.String()), ctrl.SetPostPrivate)
	// Like post | 点赞帖子
	router.POST("/like", saGin.CheckRole(user.RoleUser.String()), ctrl.LikePost)
	// Dislike post | 点踩帖子
	router.POST("/dislike", saGin.CheckRole(user.RoleUser.String()), ctrl.DislikePost)
	// Favorite post | 收藏帖子
	router.POST("/favorite", saGin.CheckRole(user.RoleUser.String()), ctrl.FavoritePost)
	// Get post list | 获取帖子列表
	router.GET("", ctrl.GetPostList)
	// Get post detail | 获取帖子详情
	router.GET("/:id", ctrl.GetPostDetail)
}

// CreatePost Publish new post | 发布新帖
// @Summary Publish new post | 发布新帖
// @Description User publishes a new topic post | 用户发布新的主题帖
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostCreateRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostCreateResponse} "Published successfully | 发布成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts [post]
func (ctrl *PostController) CreatePost(c *gin.Context) {
	var req schema.UserPostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.CreatePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// SaveDraft Save draft | 保存草稿
// @Summary Save draft | 保存草稿
// @Description User saves post draft | 用户保存帖子草稿
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostCreateRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostCreateResponse} "Saved successfully | 保存成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/draft [post]
func (ctrl *PostController) SaveDraft(c *gin.Context) {
	var req schema.UserPostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.SaveDraft(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// UpdatePost Edit post | 编辑帖子
// @Summary Edit post | 编辑帖子
// @Description User edits their own post (can be operated once every three minutes) | 用户编辑自己的帖子(每三分钟可操作一次)
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostUpdateRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostUpdateResponse} "Edited successfully | 编辑成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 403 {object} response.Data "Insufficient permissions or too frequent operations | 权限不足或操作过于频繁"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts [put]
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	var req schema.UserPostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.UpdatePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// SetPostPrivate Set post as private | 设置帖子私有
// @Summary Set post as private | 设置帖子私有
// @Description User sets post as private or public (can be operated once every three days) | 用户设置帖子为私有或公开(每三日可操作一次)
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 403 {object} response.Data "Insufficient permissions or too frequent operations | 权限不足或操作过于频繁"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/private [put]
func (ctrl *PostController) SetPostPrivate(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.SetPostPrivate(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// LikePost Like post | 点赞帖子
// @Summary Like post | 点赞帖子
// @Description User likes a post (one-way, cannot cancel like) | 用户点赞帖子(单向,不可取消点赞)
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "Liked successfully | 点赞成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 403 {object} response.Data "Already liked | 已经点赞过"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/like [post]
func (ctrl *PostController) LikePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.LikePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DislikePost Dislike post | 点踩帖子
// @Summary Dislike post | 点踩帖子
// @Description User dislikes a post (one-way, cannot cancel dislike) | 用户点踩帖子(单向,不可取消点踩)
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "Disliked successfully | 点踩成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 403 {object} response.Data "Already disliked | 已经点踩过"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/dislike [post]
func (ctrl *PostController) DislikePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.DislikePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// FavoritePost Favorite post | 收藏帖子
// @Summary Favorite post | 收藏帖子
// @Description User favorites or unfavorites a post (two-way operation) | 用户收藏或取消收藏帖子(双向操作)
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param request body schema.UserPostActionRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.UserPostActionResponse} "Operation successful | 操作成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Not logged in | 未登录"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/favorite [post]
func (ctrl *PostController) FavoritePost(c *gin.Context) {
	var req schema.UserPostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResError(c, response.CodeNeedLogin)
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.FavoritePost(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPostList Get post list | 获取帖子列表
// @Summary Get post list | 获取帖子列表
// @Description Get post list with pagination and sorting support | 获取帖子列表,支持分页和排序
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param category_id query int false "Category ID | 版块ID"
// @Param page query int false "Page number, default 1 | 页码,默认1" default(1)
// @Param page_size query int false "Items per page, default 20, max 100 | 每页数量,默认20,最大100" default(20)
// @Param sort query string false "Sort method: latest (newest), hot (popular), essence (featured) | 排序方式:latest(最新)、hot(热门)、essence(精华)" default(latest)
// @Success 200 {object} response.Data{data=schema.UserPostListResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts [get]
func (ctrl *PostController) GetPostList(c *gin.Context) {
	var req schema.UserPostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.GetPostList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPostDetail Get post detail | 获取帖子详情
// @Summary Get post detail | 获取帖子详情
// @Description Get detailed information of the specified post and increment view count | 获取指定帖子的详细信息,并增加浏览数
// @Tags [User]Topic Posts | [用户]主题贴
// @Accept json
// @Produce json
// @Param id path int true "Post ID | 帖子ID"
// @Success 200 {object} response.Data{data=schema.UserPostDetailResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 404 {object} response.Data "Post not found | 帖子不存在"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /posts/{id} [get]
func (ctrl *PostController) GetPostDetail(c *gin.Context) {
	var req schema.UserPostDetailRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postService.GetPostDetail(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
