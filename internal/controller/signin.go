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
	"github.com/PokeForum/PokeForum/internal/service"
)

// SigninController Sign-in controller | 签到控制器
type SigninController struct {
	signinService service.ISigninService
}

// NewSigninController Create sign-in controller instance | 创建签到控制器实例
func NewSigninController(injector *do.Injector) *SigninController {
	return &SigninController{
		signinService: do.MustInvoke[service.ISigninService](injector),
	}
}

// SigninRouter Configure sign-in system routes | 配置签到系统路由
func (ctrl *SigninController) SigninRouter(router *gin.RouterGroup) {
	// Perform sign-in | 执行签到
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.Signin)
	// Get sign-in status | 获取签到状态
	router.GET("/status", saGin.CheckRole(user.RoleUser.String()), ctrl.GetSigninStatus)
	// Get daily ranking | 获取每日排行榜
	router.GET("/ranking/daily", ctrl.GetDailyRanking)
	// Get continuous sign-in ranking | 获取连续签到排行榜
	router.GET("/ranking/continuous", ctrl.GetContinuousRanking)
}

// getUserID Get token from Header and parse user ID | 从Header中获取token并解析用户ID
func (ctrl *SigninController) getUserID(c *gin.Context) (int, error) {
	// Get token from Header | 从Header中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		return 0, fmt.Errorf("未找到Authorization header")
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

// Signin Perform sign-in | 执行签到
// @Summary Perform sign-in | 执行签到
// @Description User performs daily sign-in and receives points and experience rewards | 用户执行每日签到，获得积分和经验值奖励
// @Tags [User]Sign-in | [用户]签到
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SigninResponse} "Sign-in successful | 签到成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 403 {object} response.Data "Sign-in feature not enabled | 签到功能未启用"
// @Failure 409 {object} response.Data "Already signed in today | 今日已签到"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /signin [post]
func (ctrl *SigninController) Signin(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// Call sign-in service | 调用签到服务
	result, err := ctrl.signinService.Signin(c.Request.Context(), int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetSigninStatus Get sign-in status | 获取签到状态
// @Summary Get sign-in status | 获取签到状态
// @Description Get user's sign-in status including consecutive days, total days, etc. | 获取用户的签到状态，包括连续签到天数、总签到天数等信息
// @Tags [User]Sign-in | [用户]签到
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SigninStatus} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 401 {object} response.Data "Unauthorized | 未授权"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /signin/status [get]
func (ctrl *SigninController) GetSigninStatus(c *gin.Context) {
	// Get user ID | 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// Call sign-in service to get status | 调用签到服务获取状态
	status, err := ctrl.signinService.GetSigninStatus(c.Request.Context(), int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, status)
}

// GetDailyRanking Get daily sign-in ranking | 获取每日签到排行榜
// @Summary Get daily sign-in ranking | 获取每日签到排行榜
// @Description Get sign-in ranking for specified date, sorted by reward points descending, returns top 100 max | 获取指定日期的签到排行榜，按奖励积分从高到低排序，最多返回前100名
// @Tags [User]Sign-in | [用户]签到
// @Accept json
// @Produce json
// @Param date query string false "Query date in format YYYY-MM-DD, defaults to today if not provided | 查询日期，格式：YYYY-MM-DD，不传则查询今日"
// @Param limit query int false "Return count limit, default 10, max 100 | 返回数量限制，默认10，最大100"
// @Success 200 {object} response.Data{data=schema.SigninRankingResponse} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /signin/ranking/daily [get]
func (ctrl *SigninController) GetDailyRanking(c *gin.Context) {
	// Parse query parameters | 解析查询参数
	date := c.Query("date")
	limit := 10 // Default value | 默认值

	// Get limit count parameter | 获取限制数量参数
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "限制数量必须是1-100之间的整数")
			return
		}
		limit = parsedLimit
	}

	// Try to get user ID (optional, for getting current user's rank) | 尝试获取用户ID（可选，用于获取当前用户排名）
	userID, _ := ctrl.getUserID(c) //nolint:errcheck // User ID is optional | 用户ID是可选的

	// Call sign-in service to get ranking | 调用签到服务获取排行榜
	ranking, err := ctrl.signinService.GetDailyRanking(c.Request.Context(), date, limit, int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, ranking)
}

// GetContinuousRanking Get continuous sign-in ranking | 获取连续签到排行榜
// @Summary Get continuous sign-in ranking | 获取连续签到排行榜
// @Description Get continuous sign-in days ranking, sorted by consecutive days descending, returns top 100 max | 获取连续签到天数排行榜，按连续签到天数从高到低排序，最多返回前100名
// @Tags [User]Sign-in | [用户]签到
// @Accept json
// @Produce json
// @Param limit query int false "Return count limit, default 10, max 100 | 返回数量限制，默认10，最大100"
// @Success 200 {object} response.Data{data=schema.SigninRankingResponse} "Retrieve successful | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Internal server error | 服务器内部错误"
// @Router /signin/ranking/continuous [get]
func (ctrl *SigninController) GetContinuousRanking(c *gin.Context) {
	// Parse query parameters | 解析查询参数
	limit := 10 // Default value | 默认值

	// Get limit count parameter | 获取限制数量参数
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "限制数量必须是1-100之间的整数")
			return
		}
		limit = parsedLimit
	}

	// Try to get user ID (optional, for getting current user's rank) | 尝试获取用户ID（可选，用于获取当前用户排名）
	userID, _ := ctrl.getUserID(c) //nolint:errcheck // User ID is optional | 用户ID是可选的

	// Call sign-in service to get ranking | 调用签到服务获取排行榜
	ranking, err := ctrl.signinService.GetContinuousRanking(c.Request.Context(), limit, int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, ranking)
}
