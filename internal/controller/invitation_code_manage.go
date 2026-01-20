package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// InvitationCodeManageController Invitation code management controller | 邀请码管理控制器
type InvitationCodeManageController struct {
	invitationCodeManageService service.IInvitationCodeManageService
}

// NewInvitationCodeManageController Create invitation code management controller instance | 创建邀请码管理控制器实例
func NewInvitationCodeManageController(invitationCodeManageService service.IInvitationCodeManageService) *InvitationCodeManageController {
	return &InvitationCodeManageController{
		invitationCodeManageService: invitationCodeManageService,
	}
}

// InvitationCodeManageRouter Invitation code management related route registration | 邀请码管理相关路由注册
func (ctrl *InvitationCodeManageController) InvitationCodeManageRouter(router *gin.RouterGroup) {
	// Invitation code list | 邀请码列表
	router.GET("", ctrl.GetInvitationCodeList)
	// Create invitation code | 创建邀请码
	router.POST("", ctrl.CreateInvitationCode)
	// Update invitation code information | 更新邀请码信息
	router.PUT("", ctrl.UpdateInvitationCode)
	// Get invitation code details | 获取邀请码详情
	router.GET("/:id", ctrl.GetInvitationCodeDetail)
	// Delete invitation code | 删除邀请码
	router.DELETE("/:id", ctrl.DeleteInvitationCode)

	// Invitation code status management | 邀请码状态管理
	router.PUT("/status", ctrl.UpdateInvitationCodeStatus)

	// Invitation code statistics | 邀请码统计
	router.GET("/stats", ctrl.GetInvitationCodeStats)
}

// GetInvitationCodeList Get invitation code list | 获取邀请码列表
// @Summary Get invitation code list | 获取邀请码列表
// @Description Get paginated invitation code list with support for keyword search and status filtering | 分页获取邀请码列表，支持关键词搜索和状态筛选
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Param keyword query string false "Search keyword (invitation code) | 搜索关键词（邀请码）" example("abc123")
// @Param status query string false "Status filter: unused, used, expired, disabled | 状态筛选" example("unused")
// @Param mode query string false "Generation mode filter | 生成方式筛选" example("direct")
// @Success 200 {object} response.Data{data=schema.InvitationCodeListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes [get]
func (ctrl *InvitationCodeManageController) GetInvitationCodeList(c *gin.Context) {
	var req schema.InvitationCodeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	codes, total, err := ctrl.invitationCodeManageService.GetInvitationCodeList(c.Request.Context(), req.Page, req.PageSize, req.Keyword, req.Status, req.Mode)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	list := make([]schema.InvitationCodeListItem, 0, len(codes))
	for _, code := range codes {
		item := schema.InvitationCodeListItem{
			ID:             code.ID,
			Code:           code.Code,
			CreatorID:      code.CreatorID,
			Status:         code.Status.String(),
			GenerationMode: code.GenerationMode.String(),
			CostAmount:     code.CostAmount,
			Remark:         code.Remark,
			CreatedAt:      code.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:      code.UpdatedAt.Format(time_tools.DateTimeFormat),
		}

		if code.UsedByID != nil {
			item.UsedByID = code.UsedByID
		}
		if !code.UsedAt.IsZero() {
			item.UsedAt = code.UsedAt.Format(time_tools.DateTimeFormat)
		}
		if !code.ExpiresAt.IsZero() {
			item.ExpiresAt = code.ExpiresAt.Format(time_tools.DateTimeFormat)
		}
		if code.UsedIP != "" {
			item.UsedIP = code.UsedIP
		}
		if code.UsedUserAgent != "" {
			item.UsedUserAgent = code.UsedUserAgent
		}

		list = append(list, item)
	}

	result := &schema.InvitationCodeListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	response.ResSuccess(c, result)
}

// CreateInvitationCode Create invitation code | 创建邀请码
// @Summary Create invitation code | 创建邀请码
// @Description Admin creates invitation code manually | 管理员手动创建邀请码
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param request body schema.InvitationCodeCreateRequest true "Invitation code information | 邀请码信息"
// @Success 200 {object} response.Data{data=schema.InvitationCodeDetailResponse} "Created successfully | 创建成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes [post]
func (ctrl *InvitationCodeManageController) CreateInvitationCode(c *gin.Context) {
	var req schema.InvitationCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.ParseInLocation(time_tools.DateTimeFormat, req.ExpiresAt, time.Local)
		if err != nil {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "过期时间格式错误")
			return
		}
		expiresAt = &t
	}

	invCode, err := ctrl.invitationCodeManageService.CreateInvitationCode(c.Request.Context(), req.Code, req.CreatorID, req.Mode, req.CostAmount, expiresAt, req.Remark)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	result := ctrl.convertToDetailResponse(invCode)
	response.ResSuccess(c, result)
}

// UpdateInvitationCode Update invitation code information | 更新邀请码信息
// @Summary Update invitation code information | 更新邀请码信息
// @Description Update basic information of invitation code | 更新邀请码的基本信息
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param request body schema.InvitationCodeUpdateRequest true "Invitation code information | 邀请码信息"
// @Success 200 {object} response.Data{data=schema.InvitationCodeDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes [put]
func (ctrl *InvitationCodeManageController) UpdateInvitationCode(c *gin.Context) {
	var req schema.InvitationCodeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.ParseInLocation(time_tools.DateTimeFormat, req.ExpiresAt, time.Local)
		if err != nil {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "过期时间格式错误")
			return
		}
		expiresAt = &t
	}

	invCode, err := ctrl.invitationCodeManageService.UpdateInvitationCode(c.Request.Context(), req.ID, expiresAt, req.Remark)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	result := ctrl.convertToDetailResponse(invCode)
	response.ResSuccess(c, result)
}

// GetInvitationCodeDetail Get invitation code details | 获取邀请码详情
// @Summary Get invitation code details | 获取邀请码详情
// @Description Get detailed information of invitation code | 获取邀请码的详细信息
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param id path int true "Invitation code ID | 邀请码ID" example("1")
// @Success 200 {object} response.Data{data=schema.InvitationCodeDetailResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes/{id} [get]
func (ctrl *InvitationCodeManageController) GetInvitationCodeDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "邀请码ID格式错误")
		return
	}

	// Get invitation code list and find by ID | 获取邀请码列表并按ID查找
	codes, _, err := ctrl.invitationCodeManageService.GetInvitationCodeList(c.Request.Context(), 1, 1, "", "", "")
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	var targetInvCode *ent.InvitationCode
	for _, code := range codes {
		if code.ID == id {
			targetInvCode = code
			break
		}
	}

	if targetInvCode == nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "邀请码不存在")
		return
	}

	result := ctrl.convertToDetailResponse(targetInvCode)
	response.ResSuccess(c, result)
}

// DeleteInvitationCode Delete invitation code | 删除邀请码
// @Summary Delete invitation code | 删除邀请码
// @Description Delete invitation code (unused codes only) | 删除邀请码（仅限未使用的邀请码）
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param id path int true "Invitation code ID | 邀请码ID" example("1")
// @Success 200 {object} response.Data "Deleted successfully | 删除成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes/{id} [delete]
func (ctrl *InvitationCodeManageController) DeleteInvitationCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "邀请码ID格式错误")
		return
	}

	err = ctrl.invitationCodeManageService.DeleteInvitationCode(c.Request.Context(), id)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, nil)
}

// UpdateInvitationCodeStatus Update invitation code status | 更新邀请码状态
// @Summary Update invitation code status | 更新邀请码状态
// @Description Update status of invitation code | 更新邀请码的状态
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Param request body schema.InvitationCodeStatusUpdateRequest true "Status information | 状态信息"
// @Success 200 {object} response.Data{data=schema.InvitationCodeDetailResponse} "Updated successfully | 更新成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes/status [put]
func (ctrl *InvitationCodeManageController) UpdateInvitationCodeStatus(c *gin.Context) {
	var req schema.InvitationCodeStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	invCode, err := ctrl.invitationCodeManageService.UpdateInvitationCodeStatus(c.Request.Context(), req.ID, req.Status)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	result := ctrl.convertToDetailResponse(invCode)
	response.ResSuccess(c, result)
}

// GetInvitationCodeStats Get invitation code statistics | 获取邀请码统计信息
// @Summary Get invitation code statistics | 获取邀请码统计信息
// @Description Get statistics of invitation codes including counts by status | 获取邀请码统计信息，包括按状态统计的数量
// @Tags [Admin]Invitation Code Management | [管理员]邀请码管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.InvitationCodeStatsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /manage/invitation-codes/stats [get]
func (ctrl *InvitationCodeManageController) GetInvitationCodeStats(c *gin.Context) {
	stats, err := ctrl.invitationCodeManageService.GetInvitationCodeStats(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	result := &schema.InvitationCodeStatsResponse{
		TotalCount:    stats["total"],
		UnusedCount:   stats["unused_count"],
		UsedCount:     stats["used_count"],
		ExpiredCount:  stats["expired_count"],
		DisabledCount: stats["disabled_count"],
	}

	response.ResSuccess(c, result)
}

// convertToDetailResponse Convert entity to detail response | 转换为详情响应
func (ctrl *InvitationCodeManageController) convertToDetailResponse(invCode *ent.InvitationCode) *schema.InvitationCodeDetailResponse {
	result := &schema.InvitationCodeDetailResponse{
		ID:             invCode.ID,
		Code:           invCode.Code,
		CreatorID:      invCode.CreatorID,
		Status:         invCode.Status.String(),
		GenerationMode: invCode.GenerationMode.String(),
		CostAmount:     invCode.CostAmount,
		Remark:         invCode.Remark,
		CreatedAt:      invCode.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      invCode.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	if invCode.UsedByID != nil {
		result.UsedByID = invCode.UsedByID
	}
	if !invCode.UsedAt.IsZero() {
		result.UsedAt = invCode.UsedAt.Format(time_tools.DateTimeFormat)
	}
	if !invCode.ExpiresAt.IsZero() {
		result.ExpiresAt = invCode.ExpiresAt.Format(time_tools.DateTimeFormat)
	}
	if invCode.UsedIP != "" {
		result.UsedIP = invCode.UsedIP
	}
	if invCode.UsedUserAgent != "" {
		result.UsedUserAgent = invCode.UsedUserAgent
	}

	return result
}
