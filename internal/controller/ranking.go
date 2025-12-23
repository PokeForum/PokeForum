package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// RankingController 排行榜控制器
type RankingController struct {
	// 注入器实例，用于获取服务
	injector *do.Injector
}

// NewRankingController 创建排行榜控制器实例
func NewRankingController(injector *do.Injector) *RankingController {
	return &RankingController{
		injector: injector,
	}
}

// RankingRouter 排行榜相关路由注册
func (ctrl *RankingController) RankingRouter(router *gin.RouterGroup) {
	// 获取排行榜列表
	router.GET("", ctrl.GetRankingList)
}

// GetRankingList 获取排行榜列表
// @Summary 获取排行榜列表
// @Description 根据排行榜类型和时间范围获取排行榜数据，支持阅读榜和评论榜
// @Tags [用户]排行榜
// @Accept json
// @Produce json
// @Param type query string true "排行榜类型：reading(阅读榜), comment(评论榜)" example("reading")
// @Param time_range query string true "时间范围：all(总榜), month(月榜), week(周榜)" example("all")
// @Param page query int true "页码" example("1")
// @Param page_size query int true "每页数量" example("20")
// @Success 200 {object} response.Data{data=schema.UserRankingListResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器内部错误"
// @Router /ranking [get]
func (ctrl *RankingController) GetRankingList(c *gin.Context) {
	// 解析请求参数
	var req schema.UserRankingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, 400, "请求参数错误", err.Error())
		return
	}

	// 获取服务实例
	rankingService := do.MustInvoke[service.IRankingService](ctrl.injector)

	// 根据排行榜类型调用不同的服务方法
	var result *schema.UserRankingListResponse
	var err error

	switch req.Type {
	case "reading":
		// 获取阅读排行榜
		result, err = rankingService.GetReadingRanking(c.Request.Context(), req)
		if err != nil {
			response.ResErrorWithMsg(c, 500, "获取阅读排行榜失败", err.Error())
			return
		}
	case "comment":
		// 获取评论排行榜
		result, err = rankingService.GetCommentRanking(c.Request.Context(), req)
		if err != nil {
			response.ResErrorWithMsg(c, 500, "获取评论排行榜失败", err.Error())
			return
		}
	default:
		response.ResErrorWithMsg(c, 400, "不支持的排行榜类型", "仅支持 reading(阅读榜) 和 comment(评论榜)")
		return
	}

	// 返回成功响应
	response.ResSuccess(c, result)
}
