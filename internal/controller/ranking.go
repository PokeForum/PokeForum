package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// RankingController Ranking controller | 排行榜控制器
type RankingController struct {
	rankingService service.IRankingService
}

// NewRankingController Create ranking controller instance | 创建排行榜控制器实例
func NewRankingController(injector *do.Injector) *RankingController {
	return &RankingController{
		rankingService: do.MustInvoke[service.IRankingService](injector),
	}
}

// RankingRouter Ranking related route registration | 排行榜相关路由注册
func (ctrl *RankingController) RankingRouter(router *gin.RouterGroup) {
	// Get ranking list | 获取排行榜列表
	router.GET("", ctrl.GetRankingList)
}

// GetRankingList Get ranking list | 获取排行榜列表
// @Summary Get ranking list | 获取排行榜列表
// @Description Get ranking data by ranking type and time range, supports reading and comment rankings | 根据排行榜类型和时间范围获取排行榜数据，支持阅读榜和评论榜
// @Tags [User]Ranking | [用户]排行榜
// @Accept json
// @Produce json
// @Param type query string true "Ranking type: reading(reading ranking), comment(comment ranking) | 排行榜类型：reading(阅读榜), comment(评论榜)" example("reading")
// @Param time_range query string true "Time range: all(all-time), month(monthly), week(weekly) | 时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Param page query int true "Page number | 页码" example("1")
// @Param page_size query int true "Items per page | 每页数量" example("20")
// @Success 200 {object} response.Data{data=schema.UserRankingListResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /ranking [get]
func (ctrl *RankingController) GetRankingList(c *gin.Context) {
	// Parse request parameters | 解析请求参数
	var req schema.UserRankingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "Invalid request parameters | 请求参数错误", err.Error())
		return
	}

	// Call different service methods based on ranking type | 根据排行榜类型调用不同的服务方法
	var result *schema.UserRankingListResponse
	var err error

	switch req.Type {
	case "reading":
		// Get reading ranking | 获取阅读排行榜
		result, err = ctrl.rankingService.GetReadingRanking(c.Request.Context(), req)
		if err != nil {
			response.ResErrorWithMsg(c, 500, "Failed to get reading ranking | 获取阅读排行榜失败", err.Error())
			return
		}
	case "comment":
		// Get comment ranking | 获取评论排行榜
		result, err = ctrl.rankingService.GetCommentRanking(c.Request.Context(), req)
		if err != nil {
			response.ResErrorWithMsg(c, 500, "Failed to get comment ranking | 获取评论排行榜失败", err.Error())
			return
		}
	default:
		response.ResErrorWithMsg(c, 400, "Unsupported ranking type | 不支持的排行榜类型", "Only supports reading and comment rankings | 仅支持 reading(阅读榜) 和 comment(评论榜)")
		return
	}

	// Return success response | 返回成功响应
	response.ResSuccess(c, result)
}
