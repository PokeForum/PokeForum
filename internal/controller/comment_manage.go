package controller

import (
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// CommentManageController Comment management controller | 评论管理控制器
type CommentManageController struct {
	commentManageService service.ICommentManageService
}

// NewCommentManageController Create comment management controller instance | 创建评论管理控制器实例
func NewCommentManageController(injector *do.Injector) *CommentManageController {
	return &CommentManageController{
		commentManageService: do.MustInvoke[service.ICommentManageService](injector),
	}
}

// CommentManageRouter Comment management related route registration | 评论管理相关路由注册
func (ctrl *CommentManageController) CommentManageRouter(router *gin.RouterGroup) {
	// Comment list | 评论列表
	router.GET("", ctrl.GetCommentList)
	// Create comment | 创建评论
	router.POST("", ctrl.CreateComment)
	// Update comment information | 更新评论信息
	router.PUT("", ctrl.UpdateComment)
	// Get comment detail | 获取评论详情
	router.GET("/:id", ctrl.GetCommentDetail)
	// Delete comment | 删除评论
	router.DELETE("/:id", ctrl.DeleteComment)

	// Comment property management | 评论属性管理
	router.PUT("/selected", ctrl.SetCommentSelected)
	router.PUT("/pin", ctrl.SetCommentPin)
}

// GetCommentList Get comment list | 获取评论列表
// @Summary Get comment list | 获取评论列表
// @Description Get paginated comment list with multiple filtering conditions | 分页获取评论列表,支持多种筛选条件
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param keyword query string false "Search keyword | 搜索关键词" example("技术")
// @Param post_id query int false "Post ID | 帖子ID" example("1")
// @Param user_id query int false "User ID | 用户ID" example("1")
// @Param parent_id query int false "Parent comment ID | 父评论ID" example("1")
// @Param is_selected query bool false "Is featured comment | 是否精选评论" example("true")
// @Param is_pinned query bool false "Is pinned comment | 是否置顶评论" example("false")
// @Param reply_to_id query int false "Reply target user ID | 回复目标用户ID" example("1")
// @Success 200 {object} response.Data{data=schema.CommentListResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments [get]
func (ctrl *CommentManageController) GetCommentList(c *gin.Context) {
	var req schema.CommentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.commentManageService.GetCommentList(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// CreateComment Create comment | 创建评论
// @Summary Create comment | 创建评论
// @Description Admin creates new comment | 管理员创建新评论
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentCreateRequest true "Comment information | 评论信息"
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments [post]
func (ctrl *CommentManageController) CreateComment(c *gin.Context) {
	var req schema.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	comment, err := ctrl.commentManageService.CreateComment(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CommentDetailResponse{
		ID:              comment.ID,
		PostID:          comment.PostID,
		PostTitle:       "", // Need to query post info, simplified here | 需要查询帖子信息,这里简化处理
		UserID:          comment.UserID,
		Username:        "", // Need to query user info, simplified here | 需要查询用户信息,这里简化处理
		ParentID:        &comment.ParentID,
		ReplyToUserID:   &comment.ReplyToUserID,
		ReplyToUsername: "", // Need to query reply target user info, simplified here | 需要查询回复目标用户信息,这里简化处理
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

// UpdateComment Update comment information | 更新评论信息
// @Summary Update comment information | 更新评论信息
// @Description Update comment content information | 更新评论的内容信息
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentUpdateRequest true "Comment information | 评论信息"
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments [put]
func (ctrl *CommentManageController) UpdateComment(c *gin.Context) {
	var req schema.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	comment, err := ctrl.commentManageService.UpdateComment(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Convert to response format | 转换为响应格式
	result := &schema.CommentDetailResponse{
		ID:              comment.ID,
		PostID:          comment.PostID,
		PostTitle:       "", // Need to query post info, simplified here | 需要查询帖子信息,这里简化处理
		UserID:          comment.UserID,
		Username:        "", // Need to query user info, simplified here | 需要查询用户信息,这里简化处理
		ParentID:        &comment.ParentID,
		ReplyToUserID:   &comment.ReplyToUserID,
		ReplyToUsername: "", // Need to query reply target user info, simplified here | 需要查询回复目标用户信息,这里简化处理
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

// GetCommentDetail Get comment detail | 获取评论详情
// @Summary Get comment detail | 获取评论详情
// @Description Get detailed information of the specified comment | 获取指定评论的详细信息
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param id path int true "Comment ID | 评论ID" example("1")
// @Success 200 {object} response.Data{data=schema.CommentDetailResponse} "Retrieved successfully | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments/{id} [get]
func (ctrl *CommentManageController) GetCommentDetail(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	result, err := ctrl.commentManageService.GetCommentDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// DeleteComment Delete comment | 删除评论
// @Summary Delete comment | 删除评论
// @Description Delete the specified comment | 删除指定的评论
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param id path int true "Comment ID | 评论ID" example("1")
// @Success 200 {object} response.Data "Deleted successfully | 删除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments/{id} [delete]
func (ctrl *CommentManageController) DeleteComment(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.commentManageService.DeleteComment(c.Request.Context(), req.ID)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCommentSelected Set comment as featured | 设置评论精选
// @Summary Set comment as featured | 设置评论精选
// @Description Set or cancel featured status of a comment | 设置或取消评论的精选状态
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentSelectedUpdateRequest true "Featured information | 精选信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments/selected [put]
func (ctrl *CommentManageController) SetCommentSelected(c *gin.Context) {
	var req schema.CommentSelectedUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.commentManageService.SetCommentSelected(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// SetCommentPin Set comment as pinned | 设置评论置顶
// @Summary Set comment as pinned | 设置评论置顶
// @Description Set or cancel pinned status of a comment | 设置或取消评论的置顶状态
// @Tags [Admin]Comment Management | [管理员]评论管理
// @Accept json
// @Produce json
// @Param request body schema.CommentPinUpdateRequest true "Pin information | 置顶信息"
// @Success 200 {object} response.Data "Set successfully | 设置成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/comments/pin [put]
func (ctrl *CommentManageController) SetCommentPin(c *gin.Context) {
	var req schema.CommentPinUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Call service | 调用服务
	err := ctrl.commentManageService.SetCommentPin(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}
