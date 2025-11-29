package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// CommentManageController 评论管理控制器
type CommentManageController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewCommentManageController 创建评论管理控制器实例
func NewCommentManageController(injector *do.Injector) *CommentManageController {
	return &CommentManageController{
		injector: injector,
	}
}

// CommentManageRouter 评论管理相关路由注册
func (ctrl *CommentManageController) CommentManageRouter(router *gin.RouterGroup) {
	// 评论列表
	router.GET("", ctrl.GetCommentList)
	// 创建评论
	router.POST("", ctrl.CreateComment)
	// 更新评论信息
	router.PUT("", ctrl.UpdateComment)
	// 获取评论详情
	router.GET("/:id", ctrl.GetCommentDetail)
	// 删除评论
	router.DELETE("/:id", ctrl.DeleteComment)

	// 评论属性管理
	router.PUT("/selected", ctrl.SetCommentSelected)
	router.PUT("/pin", ctrl.SetCommentPin)
}

// GetCommentList 获取评论列表
// @Summary 获取评论列表
// @Description 分页获取评论列表，支持多种筛选条件
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Param keyword query string false "搜索关键词" example("技术")
// @Param post_id query int false "帖子ID" example("1")
// @Param user_id query int false "用户ID" example("1")
// @Param parent_id query int false "父评论ID" example("1")
// @Param is_selected query bool false "是否精选评论" example("true")
// @Param is_pinned query bool false "是否置顶评论" example("false")
// @Param reply_to_id query int false "回复目标用户ID" example("1")
// @Success 200 {object} response.Data{data=schema.CommentListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments [get]
func (ctrl *CommentManageController) GetCommentList(c *gin.Context) {
	var req schema.CommentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := commentManageService.GetCommentList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 管理员创建新评论
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentCreateRequest true "评论信息"
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "创建成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments [post]
func (ctrl *CommentManageController) CreateComment(c *gin.Context) {
	var req schema.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	comment, err := commentManageService.CreateComment(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.CommentDetailResponse{
		ID:              comment.ID,
		PostID:          comment.PostID,
		PostTitle:       "", // 需要查询帖子信息，这里简化处理
		UserID:          comment.UserID,
		Username:        "", // 需要查询用户信息，这里简化处理
		ParentID:        &comment.ParentID,
		ReplyToUserID:   &comment.ReplyToUserID,
		ReplyToUsername: "", // 需要查询回复目标用户信息，这里简化处理
		Content:         comment.Content,
		LikeCount:       comment.LikeCount,
		DislikeCount:    comment.DislikeCount,
		IsSelected:      comment.IsSelected,
		IsPinned:        comment.IsPinned,
		CommenterIP:     comment.CommenterIP,
		DeviceInfo:      comment.DeviceInfo,
		CreatedAt:       comment.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:       comment.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// UpdateComment 更新评论信息
// @Summary 更新评论信息
// @Description 更新评论的内容信息
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentUpdateRequest true "评论信息"
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "更新成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments [put]
func (ctrl *CommentManageController) UpdateComment(c *gin.Context) {
	var req schema.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	comment, err := commentManageService.UpdateComment(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// 转换为响应格式
	result := &schema.CommentDetailResponse{
		ID:              comment.ID,
		PostID:          comment.PostID,
		PostTitle:       "", // 需要查询帖子信息，这里简化处理
		UserID:          comment.UserID,
		Username:        "", // 需要查询用户信息，这里简化处理
		ParentID:        &comment.ParentID,
		ReplyToUserID:   &comment.ReplyToUserID,
		ReplyToUsername: "", // 需要查询回复目标用户信息，这里简化处理
		Content:         comment.Content,
		LikeCount:       comment.LikeCount,
		DislikeCount:    comment.DislikeCount,
		IsSelected:      comment.IsSelected,
		IsPinned:        comment.IsPinned,
		CommenterIP:     comment.CommenterIP,
		DeviceInfo:      comment.DeviceInfo,
		CreatedAt:       comment.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:       comment.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	response.ResSuccess(c, result)
}

// GetCommentDetail 获取评论详情
// @Summary 获取评论详情
// @Description 获取指定评论的详细信息
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param id path int true "评论ID" example("1")
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments/{id} [get]
func (ctrl *CommentManageController) GetCommentDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	result, err := commentManageService.GetCommentDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteComment 删除评论
// @Summary 删除评论
// @Description 删除指定的评论
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param id path int true "评论ID" example("1")
// @Success 200 {object} response.Data "删除成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments/{id} [delete]
func (ctrl *CommentManageController) DeleteComment(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = commentManageService.DeleteComment(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCommentSelected 设置评论精选
// @Summary 设置评论精选
// @Description 设置或取消评论的精选状态
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentSelectedUpdateRequest true "精选信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments/selected [put]
func (ctrl *CommentManageController) SetCommentSelected(c *gin.Context) {
	var req schema.CommentSelectedUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = commentManageService.SetCommentSelected(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCommentPin 设置评论置顶
// @Summary 设置评论置顶
// @Description 设置或取消评论的置顶状态
// @Tags [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentPinUpdateRequest true "置顶信息"
// @Success 200 {object} response.Data "设置成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /manage/comments/pin [put]
func (ctrl *CommentManageController) SetCommentPin(c *gin.Context) {
	var req schema.CommentPinUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 获取服务
	commentManageService, err := do.Invoke[service.ICommentManageService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 调用服务
	err = commentManageService.SetCommentPin(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
