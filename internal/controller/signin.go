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

// SigninController 签到控制器
type SigninController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewSigninController 创建签到控制器实例
func NewSigninController(injector *do.Injector) *SigninController {
	return &SigninController{
		injector: injector,
	}
}

// SigninRouter 配置签到系统路由
func (ctrl *SigninController) SigninRouter(router *gin.RouterGroup) {
	// 执行签到
	router.POST("", saGin.CheckRole(user.RoleUser.String()), ctrl.Signin)
	// 获取签到状态
	router.GET("/status", saGin.CheckRole(user.RoleUser.String()), ctrl.GetSigninStatus)
	// 获取每日排行榜
	router.GET("/ranking/daily", ctrl.GetDailyRanking)
	// 获取连续签到排行榜
	router.GET("/ranking/continuous", ctrl.GetContinuousRanking)
}

// getUserID 从Header中获取token并解析用户ID
func (ctrl *SigninController) getUserID(c *gin.Context) (int, error) {
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

// Signin 执行签到
// @Summary 执行签到
// @Description 用户执行每日签到，获得积分和经验值奖励
// @Tags [用户]签到
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SigninResponse} "签到成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 403 {object} response.Data "签到功能未启用"
// @Failure 409 {object} response.Data "今日已签到"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /signin [post]
func (ctrl *SigninController) Signin(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 获取服务
	signinService := do.MustInvoke[service.ISigninService](ctrl.injector)

	// 调用签到服务
	result, err := signinService.Signin(c.Request.Context(), int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetSigninStatus 获取签到状态
// @Summary 获取签到状态
// @Description 获取用户的签到状态，包括连续签到天数、总签到天数等信息
// @Tags [用户]签到
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.SigninStatus} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 401 {object} response.Data "未授权"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /signin/status [get]
func (ctrl *SigninController) GetSigninStatus(c *gin.Context) {
	// 获取用户ID
	userID, err := ctrl.getUserID(c)
	if err != nil {
		response.ResErrorWithMsg(c, 401, "获取用户信息失败", err.Error())
		return
	}

	// 获取服务
	signinService := do.MustInvoke[service.ISigninService](ctrl.injector)

	// 调用签到服务获取状态
	status, err := signinService.GetSigninStatus(c.Request.Context(), int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, status)
}

// GetDailyRanking 获取每日签到排行榜
// @Summary 获取每日签到排行榜
// @Description 获取指定日期的签到排行榜，按奖励积分从高到低排序，最多返回前100名
// @Tags [用户]签到
// @Accept json
// @Produce json
// @Param date query string false "查询日期，格式：YYYY-MM-DD，不传则查询今日"
// @Param limit query int false "返回数量限制，默认10，最大100"
// @Success 200 {object} response.Data{data=schema.SigninRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /signin/ranking/daily [get]
func (ctrl *SigninController) GetDailyRanking(c *gin.Context) {
	// 解析查询参数
	date := c.Query("date")
	limit := 10 // 默认值

	// 获取限制数量参数
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "限制数量必须是1-100之间的整数")
			return
		}
		limit = parsedLimit
	}

	// 尝试获取用户ID（可选，用于获取当前用户排名）
	userID, _ := ctrl.getUserID(c) //nolint:errcheck // 用户ID是可选的

	// 获取服务
	signinService := do.MustInvoke[service.ISigninService](ctrl.injector)

	// 调用签到服务获取排行榜
	ranking, err := signinService.GetDailyRanking(c.Request.Context(), date, limit, int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, ranking)
}

// GetContinuousRanking 获取连续签到排行榜
// @Summary 获取连续签到排行榜
// @Description 获取连续签到天数排行榜，按连续签到天数从高到低排序，最多返回前100名
// @Tags [用户]签到
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制，默认10，最大100"
// @Success 200 {object} response.Data{data=schema.SigninRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /signin/ranking/continuous [get]
func (ctrl *SigninController) GetContinuousRanking(c *gin.Context) {
	// 解析查询参数
	limit := 10 // 默认值

	// 获取限制数量参数
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			response.ResErrorWithMsg(c, response.CodeInvalidParam, "限制数量必须是1-100之间的整数")
			return
		}
		limit = parsedLimit
	}

	// 尝试获取用户ID（可选，用于获取当前用户排名）
	userID, _ := ctrl.getUserID(c) //nolint:errcheck // 用户ID是可选的

	// 获取服务
	signinService := do.MustInvoke[service.ISigninService](ctrl.injector)

	// 调用签到服务获取排行榜
	ranking, err := signinService.GetContinuousRanking(c.Request.Context(), limit, int64(userID))
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, ranking)
}
