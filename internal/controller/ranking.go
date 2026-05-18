package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// RankingController 排行榜控制器
type RankingController struct {
	rankingService service.IRankingService
}

// NewRankingController 创建排行榜控制器实例
func NewRankingController(rankingService service.IRankingService) *RankingController {
	return &RankingController{
		rankingService: rankingService,
	}
}

// RankingRouter 排行榜相关路由注册
func (ctrl *RankingController) RankingRouter(router *gin.RouterGroup) {
	router.GET("/reading", ctrl.GetReadingRanking)
	router.GET("/post-count", ctrl.GetPostCountRanking)
	router.GET("/comment-count", ctrl.GetCommentCountRanking)
	router.GET("/follower", ctrl.GetFollowerRanking)
	router.GET("/points", ctrl.GetPointsRanking)
	router.GET("/currency", ctrl.GetCurrencyRanking)
}

// GetReadingRanking 获取阅读排行榜
// @Summary 获取阅读排行榜
// @Description 根据时间范围获取阅读量最高的帖子列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.ReadingRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/reading [get]
func (ctrl *RankingController) GetReadingRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetReadingRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取阅读排行榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPostCountRanking 获取帖子数排行榜
// @Summary 获取帖子数排行榜
// @Description 根据时间范围获取发帖数最多的用户列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.PostCountRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/post-count [get]
func (ctrl *RankingController) GetPostCountRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetPostCountRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取帖子数排行榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetCommentCountRanking 获取评论数排行榜
// @Summary 获取评论数排行榜
// @Description 根据时间范围获取评论数最多的用户列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.CommentCountRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/comment-count [get]
func (ctrl *RankingController) GetCommentCountRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetCommentCountRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取评论数排行榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetFollowerRanking 获取名人榜（被关注数）
// @Summary 获取名人榜
// @Description 获取粉丝数最多的用户列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.FollowerRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/follower [get]
func (ctrl *RankingController) GetFollowerRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetFollowerRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取名人榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetPointsRanking 获取积分榜
// @Summary 获取积分榜
// @Description 获取积分最高的用户列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.PointsRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/points [get]
func (ctrl *RankingController) GetPointsRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetPointsRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取积分榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetCurrencyRanking 获取财富榜（货币）
// @Summary 获取财富榜
// @Description 获取货币最多的用户列表（固定返回前100条）
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Success 200 {object} response.Data{data=schema.CurrencyRankingResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking/currency [get]
func (ctrl *RankingController) GetCurrencyRanking(c *gin.Context) {
	var req schema.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "请求参数错误", err.Error())
		return
	}

	result, err := ctrl.rankingService.GetCurrencyRanking(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, "获取财富榜失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}
