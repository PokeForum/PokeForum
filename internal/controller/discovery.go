package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/service"
)

// DiscoveryController 发现控制器
type DiscoveryController struct {
	discoveryService service.IDiscoveryService
}

// NewDiscoveryController 创建发现控制器实例
func NewDiscoveryController(discoveryService service.IDiscoveryService) *DiscoveryController {
	return &DiscoveryController{
		discoveryService: discoveryService,
	}
}

// DiscoveryRouter 发现相关路由注册
func (ctrl *DiscoveryController) DiscoveryRouter(router *gin.RouterGroup) {
	router.GET("/fresh", ctrl.GetFreshPosts)
	router.GET("/discussions", ctrl.GetLatestDiscussions)
	router.GET("/comments", ctrl.GetInteractiveComments)
}

// GetFreshPosts 获取新鲜发布的帖子
// @Summary Get fresh posts | 获取新鲜发布的帖子
// @Description Get the latest published posts, sorted by creation time descending | 获取最新发布的帖子，按创建时间降序排列
// @Tags [User]Discovery | [用户]发现
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.DiscoveryFreshResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /discovery/fresh [get]
func (ctrl *DiscoveryController) GetFreshPosts(c *gin.Context) {
	result, err := ctrl.discoveryService.GetFreshPosts(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取新鲜发布帖子失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetLatestDiscussions 获取最新讨论
// @Summary Get latest discussions | 获取最新讨论
// @Description Get posts sorted by latest reply time | 获取按最近回复时间排序的帖子
// @Tags [User]Discovery | [用户]发现
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.DiscoveryLatestDiscussionResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /discovery/discussions [get]
func (ctrl *DiscoveryController) GetLatestDiscussions(c *gin.Context) {
	result, err := ctrl.discoveryService.GetLatestDiscussions(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取最新讨论失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}

// GetInteractiveComments 获取互动评论
// @Summary Get interactive comments | 获取互动评论
// @Description Get the latest comments | 获取最新的评论
// @Tags [User]Discovery | [用户]发现
// @Accept json
// @Produce json
// @Success 200 {object} response.Data{data=schema.DiscoveryCommentsResponse} "Success | 获取成功"
// @Failure 500 {object} response.Data "Server internal error | 服务器内部错误"
// @Router /discovery/comments [get]
func (ctrl *DiscoveryController) GetInteractiveComments(c *gin.Context) {
	result, err := ctrl.discoveryService.GetInteractiveComments(c.Request.Context())
	if err != nil {
		response.ResErrorWithMsg(c, 500, "获取互动评论失败", err.Error())
		return
	}

	response.ResSuccess(c, result)
}
