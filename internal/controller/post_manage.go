package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// PostManageController Post management controller | 帖子管理控制器
type PostManageController struct {
	postManageService service.IPostManageService
}

// NewPostManageController Create post management controller instance | 创建帖子管理控制器实例
func NewPostManageController(injector *do.Injector) *PostManageController {
	return &PostManageController{
		postManageService: do.MustInvoke[service.IPostManageService](injector),
	}
}

// PostManageRouter Post management related route registration | 帖子管理相关路由注册
func (ctrl *PostManageController) PostManageRouter(router *gin.RouterGroup) {
	// Post list | 帖子列表
	router.GET("", ctrl.GetPostList)
	// Create post | 创建帖子
	router.POST("", ctrl.CreatePost)
	// Update post information | 更新帖子信息
	router.PUT("", ctrl.UpdatePost)
	// Get post detail | 获取帖子详情
	router.GET("/:id", ctrl.GetPostDetail)
	// Delete post | 删除帖子
	router.DELETE("/:id", ctrl.DeletePost)

	// Post status management | 帖子状态管理
	router.PUT("/status", ctrl.UpdatePostStatus)

	// Post property management | 帖子属性管理
	router.PUT("/essence", ctrl.SetPostEssence)
	router.PUT("/pin", ctrl.SetPostPin)
	router.PUT("/move", ctrl.MovePost)
}

// GetPostList Get post list | 获取帖子列表
// @Summary Get post list | 获取帖子列表
// @Description Get paginated post list with multiple filtering conditions | 分页获取帖子列表,支持多种筛选条件
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param keyword query string false "Search keyword | 搜索关键词" example("技术")
// @Param status query string false "Post status | 帖子状态" example("Normal")
// @Param category_id query int false "Category ID | 版块ID" example("1")
// @Param user_id query int false "User ID | 用户ID" example("1")
// @Param is_essence query bool false "Is featured post | 是否精华帖" example("true")
// @Param is_pinned query bool false "Is pinned | 是否置顶" example("false")
// @Success 200 {object} response.Data{data=schema.PostListResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts [get]
func (ctrl *PostManageController) GetPostList(c *gin.Context) {
	var req schema.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postManageService.GetPostList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreatePost Create post | 创建帖子
// @Summary Create post | 创建帖子
// @Description Admin creates new post | 管理员创建新帖子
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostCreateRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts [post]
func (ctrl *PostManageController) CreatePost(c *gin.Context) {
	var req schema.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	post, err := ctrl.postManageService.CreatePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             post.ID,
		UserID:         post.UserID,
		Username:       "", // Need to query user info, simplified here | 需要查询用户信息,这里简化处理
		CategoryID:     post.CategoryID,
		CategoryName:   "", // Need to query category info, simplified here | 需要查询版块信息,这里简化处理
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

// UpdatePost Update post information | 更新帖子信息
// @Summary Update post information | 更新帖子信息
// @Description Update basic information of a post | 更新帖子的基本信息
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostUpdateRequest true "Post information | 帖子信息"
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts [put]
func (ctrl *PostManageController) UpdatePost(c *gin.Context) {
	var req schema.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	post, err := ctrl.postManageService.UpdatePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             post.ID,
		UserID:         post.UserID,
		Username:       "", // Need to query user info, simplified here | 需要查询用户信息,这里简化处理
		CategoryID:     post.CategoryID,
		CategoryName:   "", // Need to query category info, simplified here | 需要查询版块信息,这里简化处理
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

// GetPostDetail Get post detail | 获取帖子详情
// @Summary Get post detail | 获取帖子详情
// @Description Get detailed information of the specified post | 获取指定帖子的详细信息
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param id path int true "Post ID | 帖子ID" example("1")
// @Success 200 {object} response.Data{data=schema.PostDetailResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/{id} [get]
func (ctrl *PostManageController) GetPostDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.postManageService.GetPostDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeletePost Delete post | 删除帖子
// @Summary Delete post | 删除帖子
// @Description Soft delete post (set status to banned) | 软删除帖子(将状态设为封禁)
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param id path int true "Post ID | 帖子ID" example("1")
// @Success 200 {object} response.Data "Deleted successfully | 删除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/{id} [delete]
func (ctrl *PostManageController) DeletePost(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.postManageService.DeletePost(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdatePostStatus Update post status | 更新帖子状态
// @Summary Update post status | 更新帖子状态
// @Description Update post status (normal, locked, draft, private, banned) | 更新帖子的状态(正常、锁定、草稿、私有、封禁)
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostStatusUpdateRequest true "Status information | 状态信息"
// @Success 200 {object} response.Data "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/status [put]
func (ctrl *PostManageController) UpdatePostStatus(c *gin.Context) {
	var req schema.PostStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.postManageService.UpdatePostStatus(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostEssence Set post as featured | 设置帖子精华
// @Summary Set post as featured | 设置帖子精华
// @Description Set or cancel featured status of a post | 设置或取消帖子的精华状态
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostEssenceUpdateRequest true "Featured information | 精华信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/essence [put]
func (ctrl *PostManageController) SetPostEssence(c *gin.Context) {
	var req schema.PostEssenceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.postManageService.SetPostEssence(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetPostPin Set post as pinned | 设置帖子置顶
// @Summary Set post as pinned | 设置帖子置顶
// @Description Set or cancel pinned status of a post | 设置或取消帖子的置顶状态
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostPinUpdateRequest true "Pin information | 置顶信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/pin [put]
func (ctrl *PostManageController) SetPostPin(c *gin.Context) {
	var req schema.PostPinUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.postManageService.SetPostPin(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// MovePost Move post to another category | 移动帖子到其他版块
// @Summary Move post | 移动帖子
// @Description Move post to the specified category | 将帖子移动到指定的版块
// @Tags [Admin]Post Management | [管理员]主题贴管理
// @Accept json
// @Produce json
// @Param request body schema.PostMoveRequest true "Move information | 移动信息"
// @Success 200 {object} response.Data "Moved successfully | 移动成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/posts/move [put]
func (ctrl *PostManageController) MovePost(c *gin.Context) {
	var req schema.PostMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.postManageService.MovePost(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
