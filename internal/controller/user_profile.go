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

// UserProfileController 用户个人中心控制器
type UserProfileController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewUserProfileController 创建用户个人中心控制器实例
func NewUserProfileController(injector *do.Injector) *UserProfileController {
	return &UserProfileController{
		injector: injector,
	}
}

// UserProfileRouter 用户个人中心相关路由注册
func (ctrl *UserProfileController) UserProfileRouter(router *gin.RouterGroup) {
	router.Use(saGin.CheckRole(user.RoleUser.String()))

	// 获取个人中心概览
	router.GET("/overview", ctrl.GetProfileOverview)
	// 获取用户主题帖列表
	router.GET("/posts", ctrl.GetUserPosts)
	// 获取用户评论列表
	router.GET("/comments", ctrl.GetUserComments)
	// 获取用户收藏列表
	router.GET("/favorites", ctrl.GetUserFavorites)
	// 修改密码
	router.PUT("/password", ctrl.UpdatePassword)
	// 修改头像
	router.PUT("/avatar", ctrl.UpdateAvatar)
	// 修改用户名
	router.PUT("/username", ctrl.UpdateUsername)
	// 发送邮箱验证码
	router.POST("/email/verify-code", ctrl.SendEmailVerifyCode)
	// 验证邮箱
	router.POST("/email/verify", ctrl.VerifyEmail)
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *UserProfileController) getUserID(c *gin.Context) (int, error) {
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

// GetProfileOverview 获取用户个人中心概览
// @Summary 获取用户个人中心概览
// @Description 获取指定用户的个人信息和统计数据，不传user_id则获取当前登录用户信息
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID，不传则查询当前登录用户" example("1")
// @Success 200 {object} response.Data{data=schema.UserProfileOverviewResponse} "获取成功"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/overview [get]
func (ctrl *UserProfileController) GetProfileOverview(c *gin.Context) {
	// 获取当前登录用户ID
	currentUserID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析查询参数中的用户ID
	var req schema.UserProfileOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 确定要查询的用户ID
	targetUserID := currentUserID
	if req.UserID > 0 {
		targetUserID = req.UserID
	}

	// 判断是否为本人
	isOwner := targetUserID == currentUserID

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务获取个人中心概览
	result, err := profileService.GetProfileOverview(c.Request.Context(), targetUserID, isOwner)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取个人中心概览失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// GetUserPosts 获取用户主题帖列表
// @Summary 获取用户主题帖列表
// @Description 获取指定用户发布的主题帖列表，支持分页和状态筛选，不传user_id则获取当前登录用户的帖子
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID，不传则查询当前登录用户" example("1")
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Param status query string false "帖子状态筛选：Normal、Draft、Private" example("Normal")
// @Success 200 {object} response.Data{data=schema.UserProfilePostsResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/posts [get]
func (ctrl *UserProfileController) GetUserPosts(c *gin.Context) {
	// 获取当前登录用户ID
	currentUserID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserProfilePostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 确定要查询的用户ID
	targetUserID := currentUserID
	if req.UserID > 0 {
		targetUserID = req.UserID
	}

	// 判断是否为本人
	isOwner := targetUserID == currentUserID

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务获取用户主题帖列表
	result, err := profileService.GetUserPosts(c.Request.Context(), targetUserID, req, isOwner)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取用户主题帖列表失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// GetUserComments 获取用户评论列表
// @Summary 获取用户评论列表
// @Description 获取指定用户发布的评论列表，支持分页，不传user_id则获取当前登录用户的评论
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID，不传则查询当前登录用户" example("1")
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Success 200 {object} response.Data{data=schema.UserProfileCommentsResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/comments [get]
func (ctrl *UserProfileController) GetUserComments(c *gin.Context) {
	// 获取当前登录用户ID
	currentUserID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserProfileCommentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 确定要查询的用户ID
	targetUserID := currentUserID
	if req.UserID > 0 {
		targetUserID = req.UserID
	}

	// 判断是否为本人
	isOwner := targetUserID == currentUserID

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务获取用户评论列表
	result, err := profileService.GetUserComments(c.Request.Context(), targetUserID, req, isOwner)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取用户评论列表失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// GetUserFavorites 获取用户收藏列表
// @Summary 获取用户收藏列表
// @Description 获取指定用户收藏的帖子列表，支持分页，不传user_id则获取当前登录用户的收藏
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID，不传则查询当前登录用户" example("1")
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Success 200 {object} response.Data{data=schema.UserProfileFavoritesResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/favorites [get]
func (ctrl *UserProfileController) GetUserFavorites(c *gin.Context) {
	// 获取当前登录用户ID
	currentUserID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserProfileFavoritesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 确定要查询的用户ID
	targetUserID := currentUserID
	if req.UserID > 0 {
		targetUserID = req.UserID
	}

	// 判断是否为本人
	isOwner := targetUserID == currentUserID

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务获取用户收藏列表
	result, err := profileService.GetUserFavorites(c.Request.Context(), targetUserID, req, isOwner)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取用户收藏列表失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// UpdatePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param request body schema.UserUpdatePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Data{data=schema.UserUpdatePasswordResponse} "修改成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/password [put]
func (ctrl *UserProfileController) UpdatePassword(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserUpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务修改密码
	result, err := profileService.UpdatePassword(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "修改密码失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// UpdateAvatar 修改头像
// @Summary 修改头像
// @Description 修改当前登录用户的头像
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param request body schema.UserUpdateAvatarRequest true "修改头像请求"
// @Success 200 {object} response.Data{data=schema.UserUpdateAvatarResponse} "修改成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/avatar [put]
func (ctrl *UserProfileController) UpdateAvatar(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserUpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务修改头像
	result, err := profileService.UpdateAvatar(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "修改头像失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// UpdateUsername 修改用户名
// @Summary 修改用户名
// @Description 修改当前登录用户的用户名（每七日可操作一次）
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param request body schema.UserUpdateUsernameRequest true "修改用户名请求"
// @Success 200 {object} response.Data{data=schema.UserUpdateUsernameResponse} "修改成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "操作过于频繁"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/username [put]
func (ctrl *UserProfileController) UpdateUsername(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.UserUpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	profileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务修改用户名
	result, err := profileService.UpdateUsername(c.Request.Context(), userID, req)
	if err != nil {
		response.ResErrorWithMsg(c, 500, "修改用户名失败", err.Error())
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}

// SendEmailVerifyCode 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 向用户注册邮箱发送验证码，用于邮箱验证
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.EmailVerifyCodeResponse} "发送成功"
// @Failure 401 {object} response.Data "未授权"
// @Failure 429 {object} response.Data "发送频率过高"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/email/verify-code [post]
func (ctrl *UserProfileController) SendEmailVerifyCode(c *gin.Context) {
	// 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "获取用户信息失败", err.Error())
		return
	}

	// 获取用户个人中心服务
	userProfileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务发送验证码
	result, err := userProfileService.SendEmailVerifyCode(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "发送次数过多，请1小时后再试" {
			response.ResErrorWithMsg(c, response.CodeTooManyRequests, "发送频率过高", err.Error())
		} else {
			response.ResErrorWithMsg(c, response.CodeGenericError, "发送验证码失败", err.Error())
		}
		return
	}

	response.ResSuccess(c, result)
}

// VerifyEmail 验证邮箱
// @Summary 验证邮箱
// @Description 通过验证码验证用户邮箱真实性
// @Tags [用户]个人中心
// @Accept json
// @Produce json
// @Param request body schema.EmailVerifyRequest true "验证邮箱请求"
// @Success 200 {object} response.Data{data=schema.EmailVerifyResponse} "验证成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 404 {object} response.Data "验证码不存在或已过期"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /profile/email/verify [post]
func (ctrl *UserProfileController) VerifyEmail(c *gin.Context) {
	// 获取当前用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeNeedLogin, "获取用户信息失败", err.Error())
		return
	}

	// 解析请求参数
	var req schema.EmailVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	// 获取用户个人中心服务
	userProfileService := do.MustInvoke[service.IUserProfileService](ctrl.injector)

	// 调用服务验证邮箱
	result, err := userProfileService.VerifyEmail(c.Request.Context(), userID, req)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "验证码不存在或已过期" ||
			err.Error() == "验证码错误" ||
			err.Error() == "邮箱地址不匹配" {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "验证失败", err.Error())
		} else {
			response.ResErrorWithMsg(c, response.CodeGenericError, "验证邮箱失败", err.Error())
		}
		return
	}

	response.ResSuccess(c, result)
}
