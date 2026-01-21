package controller

import (
	"fmt"
	"strconv"

	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// InvitationCodeController Invitation code controller | 邀请码控制器
type InvitationCodeController struct {
	invitationCodeService service.IInvitationCodeService
	userRepo              repository.IUserRepository
}

// NewInvitationCodeController Create invitation code controller instance | 创建邀请码控制器实例
func NewInvitationCodeController(
	invitationCodeService service.IInvitationCodeService,
	userRepo repository.IUserRepository,
) *InvitationCodeController {
	return &InvitationCodeController{
		invitationCodeService: invitationCodeService,
		userRepo:              userRepo,
	}
}

// InvitationCodeRouter Configure invitation code routes | 配置邀请码路由
func (ctrl *InvitationCodeController) InvitationCodeRouter(router *gin.RouterGroup) {
	// Generate invitation code | 生成邀请码
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.GenerateInvitationCode)
	// Get my invitation codes | 获取我的邀请码列表
	router.GET("/my", saGin.CheckRole(user.RoleUser.String()), ctrl.GetMyInvitationCodes)
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *InvitationCodeController) getUserID(c *gin.Context) (int, error) {
	// Get token from Header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("authorization header not found | 未找到Authorization header")
	}

	// Use stputil to get logged-in user ID | 使用stputil获取登录用户ID
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

// GenerateInvitationCode Generate invitation code | 生成邀请码
// @Summary Generate invitation code | 生成邀请码
// @Description Generate an invitation code for the current user. May cost points or currency depending on system settings | 为当前用户生成邀请码。根据系统设置可能需要消耗积分或货币
// @Tags [User]Invitation Code | [用户]邀请码
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.UserInvitationCodeDetail} "Invitation code generated successfully | 邀请码生成成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Invitation code feature not enabled | 邀请码功能未启用"
// @Failure 409 {object} response.Data "Insufficient points/currency or maximum generation count reached | 积分/货币不足或已达到最大生成数量"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /invitation-codes [post]
func (ctrl *InvitationCodeController) GenerateInvitationCode(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Get user info | 获取用户信息
	userInfo, err := ctrl.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || userInfo == nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "User does not exist | 用户不存在")
		return
	}

	// Call service to generate invitation code | 调用服务生成邀请码
	invCode, err := ctrl.invitationCodeService.GenerateCode(c.Request.Context(), userID, userInfo.Username)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Build response | 构建响应
	result := &schema.UserInvitationCodeDetail{
		Code:           invCode.Code,
		Status:         invCode.Status.String(),
		GenerationMode: invCode.GenerationMode.String(),
		CostAmount:     invCode.CostAmount,
		UsedAt:         invCode.UsedAt,
		CreatedAt:      invCode.CreatedAt,
	}

	response.ResSuccess(c, result)
}

// GetMyInvitationCodes Get my invitation codes | 获取我的邀请码列表
// @Summary Get my invitation codes | 获取我的邀请码列表
// @Description Get paginated list of invitation codes created by the current user | 获取当前用户创建的邀请码分页列表
// @Tags [User]Invitation Code | [用户]邀请码
// @Accept json
// @Produce json
// @Param page query int true "Page number | 页码" minimum(1)
// @Param page_size query int true "Items per page | 每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Data{data=schema.UserInvitationCodeListData} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /invitation-codes/my [get]
func (ctrl *InvitationCodeController) GetMyInvitationCodes(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "Failed to get user information | 获取用户信息失败", err.Error())
		return
	}

	// Parse query parameters | 解析查询参数
	var req schema.MyInvitationCodesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Set default values | 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// Call service to get invitation codes | 调用服务获取邀请码列表
	codes, total, err := ctrl.invitationCodeService.GetUserCodes(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	// Build response | 构建响应
	list := make([]*schema.UserInvitationCodeListItem, 0, len(codes))
	for _, code := range codes {
		list = append(list, &schema.UserInvitationCodeListItem{
			ID:             code.ID,
			Code:           code.Code,
			Status:         code.Status.String(),
			GenerationMode: code.GenerationMode.String(),
			CostAmount:     code.CostAmount,
			UsedAt:         code.UsedAt,
			CreatedAt:      code.CreatedAt,
		})
	}

	result := &schema.UserInvitationCodeListData{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	response.ResSuccess(c, result)
}
