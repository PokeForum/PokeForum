package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// PostManageController 帖子管理控制器
type PostManageController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewPostManageController 创建帖子管理控制器实例
func NewPostManageController(injector *do.Injector) *PostManageController {
	return &PostManageController{
		injector: injector,
	}
}

// PostManageRouter 帖子管理相关路由注册
func (ctrl *PostManageController) PostManageRouter(router *gin.RouterGroup) {
	// 帖子列表
	router.GET("", ctrl.GetPostList)
	// 创建帖子
	router.POST("", ctrl.CreatePost)
	// 更新帖子信息
	router.PUT("", ctrl.UpdatePost)
	// 获取帖子详情
	router.GET("/:id", ctrl.GetPostDetail)
	// 删除帖子
	router.DELETE("/:id", ctrl.DeletePost)

	// 帖子状态管理
	router.PUT("/status", ctrl.UpdatePostStatus)

	// 帖子属性管理
	router.PUT("/essence", ctrl.SetPostEssence)
	router.PUT("/pin", ctrl.SetPostPin)
	router.PUT("/move", ctrl.MovePost)
}

// GetPostList 获取帖子列表
// @Summary 获取帖子列表
// @Description 分页获取帖子列表，支持多种筛选条件
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Param keyword query string false "搜索关键词" example("技术")
// @Param status query string false "帖子状态" example("Normal")
// @Param category_id query int false "版块ID" example("1")
// @Param user_id query int false "用户ID" example("1")
// @Param is_essence query bool false "是否精华帖" example("true")
// @Param is_pinned query bool false "是否置顶" example("false")
// @Success 200 {object} response.Data{data=schema.PostListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts [get]
func (ctrl *PostManageController) GetPostList(c *gin.Context) {
	var req schema.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postManageService.GetPostList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreatePost 创建帖子
// @Summary 创建帖子
// @Description 管理员创建新帖子
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostCreateRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts [post]
func (ctrl *PostManageController) CreatePost(c *gin.Context) {
	var req schema.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	post, err := postManageService.CreatePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             post.ID,
		UserID:         post.UserID,
		Username:       "", // 需要查询用户信息，这里简化处理
		CategoryID:     post.CategoryID,
		CategoryName:   "", // 需要查询版块信息，这里简化处理
		Title:          post.Title,
		Content:        post.Content,
		ReadPermission: post.ReadPermission,
		ViewCount:      post.ViewCount,
		LikeCount:      post.LikeCount,
		DislikeCount:   post.DislikeCount,
		FavoriteCount:  post.FavoriteCount,
		IsEssence:      post.IsEssence,
		IsPinned:       post.IsPinned,
		Status:         post.Status.String(),
		PublishIP:      post.PublishIP,
		CreatedAt:      post.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      post.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// UpdatePost 更新帖子信息
// @Summary 更新帖子信息
// @Description 更新帖子的基本信息
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostUpdateRequest true "帖子信息"
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts [put]
func (ctrl *PostManageController) UpdatePost(c *gin.Context) {
	var req schema.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	post, err := postManageService.UpdatePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             post.ID,
		UserID:         post.UserID,
		Username:       "", // 需要查询用户信息，这里简化处理
		CategoryID:     post.CategoryID,
		CategoryName:   "", // 需要查询版块信息，这里简化处理
		Title:          post.Title,
		Content:        post.Content,
		ReadPermission: post.ReadPermission,
		ViewCount:      post.ViewCount,
		LikeCount:      post.LikeCount,
		DislikeCount:   post.DislikeCount,
		FavoriteCount:  post.FavoriteCount,
		IsEssence:      post.IsEssence,
		IsPinned:       post.IsPinned,
		Status:         post.Status.String(),
		PublishIP:      post.PublishIP,
		CreatedAt:      post.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      post.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// GetPostDetail 获取帖子详情
// @Summary 获取帖子详情
// @Description 获取指定帖子的详细信息
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param id path int true "帖子ID" example("1")
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/{id} [get]
func (ctrl *PostManageController) GetPostDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := postManageService.GetPostDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeletePost 删除帖子
// @Summary 删除帖子
// @Description 软删除帖子（将状态设为封禁）
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param id path int true "帖子ID" example("1")
// @Success 200 {object} response.Data "删除成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/{id} [delete]
func (ctrl *PostManageController) DeletePost(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = postManageService.DeletePost(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdatePostStatus 更新帖子状态
// @Summary 更新帖子状态
// @Description 更新帖子的状态（正常、锁定、草稿、私有、封禁）
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostStatusUpdateRequest true "状态信息"
// @Success 200 {object} response.Data "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/status [put]
func (ctrl *PostManageController) UpdatePostStatus(c *gin.Context) {
	var req schema.PostStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = postManageService.UpdatePostStatus(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostEssence 设置帖子精华
// @Summary 设置帖子精华
// @Description 设置或取消帖子的精华状态
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostEssenceUpdateRequest true "精华信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/essence [put]
func (ctrl *PostManageController) SetPostEssence(c *gin.Context) {
	var req schema.PostEssenceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = postManageService.SetPostEssence(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostPin 设置帖子置顶
// @Summary 设置帖子置顶
// @Description 设置或取消帖子的置顶状态
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostPinUpdateRequest true "置顶信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/pin [put]
func (ctrl *PostManageController) SetPostPin(c *gin.Context) {
	var req schema.PostPinUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = postManageService.SetPostPin(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// MovePost 移动帖子到其他版块
// @Summary 移动帖子
// @Description 将帖子移动到指定的版块
// @Tags [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostMoveRequest true "移动信息"
// @Success 200 {object} response.Data "移动成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/posts/move [put]
func (ctrl *PostManageController) MovePost(c *gin.Context) {
	var req schema.PostMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	postManageService, err := do.Invoke[service.IPostManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = postManageService.MovePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
