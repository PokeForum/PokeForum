package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/service"
)

// HealthController Health check controller | 健康检查控制器
type HealthController struct {
	healthService service.IHealthService
}

// NewHealthController Create health check controller instance | 创建健康检查控制器实例
func NewHealthController(healthService service.IHealthService) *HealthController {
	return &HealthController{
		healthService: healthService,
	}
}

// HealthRouter Register health check routes | 注册健康检查路由
func (ctrl *HealthController) HealthRouter(router *gin.Engine) {
	// Simple liveness probe (for load balancers) | 简单存活检测(用于负载均衡器)
	router.GET("/api/v1/ping", ctrl.Ping)

	// Detailed health check (for monitoring systems) | 详细健康检查(用于监控系统)
	router.GET("/api/v1/health", ctrl.Health)

	// Readiness check (for Kubernetes and other orchestration systems) | 就绪检查(用于Kubernetes等编排系统)
	router.GET("/api/v1/ready", ctrl.Ready)
}

// Ping Simple liveness probe | 简单存活检测
// @Summary Liveness probe | 存活检测
// @Description Simple liveness probe, returns pong | 简单的存活检测,返回pong
// @Tags Health Check | 健康检查
// @Produce plain
// @Success 200 {string} string "pong"
// @Router /ping [get]
func (ctrl *HealthController) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// Health Detailed health check | 详细健康检查
// @Summary Detailed health check | 详细健康检查
// @Description Check health status of all dependent services | 检查所有依赖服务的健康状态
// @Tags Health Check | 健康检查
// @Produce json
// @Param detail query bool false "Whether to return detailed system information | 是否返回系统详细信息"
// @Success 200 {object} schema.HealthStatus "Service healthy | 服务健康"
// @Success 503 {object} schema.HealthStatus "Service unhealthy | 服务不健康"
// @Router /health [get]
func (ctrl *HealthController) Health(c *gin.Context) {
	detail := c.Query("detail") == "true"
	health := ctrl.healthService.CheckHealth(c.Request.Context(), detail)

	// Return different HTTP status codes based on status | 根据状态返回不同的HTTP状态码
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// Ready Readiness check | 就绪检查
// @Summary Readiness check | 就绪检查
// @Description Check if service is ready to receive traffic | 检查服务是否准备好接收流量
// @Tags Health Check | 健康检查
// @Produce json
// @Success 200 {object} map[string]string "Service ready | 服务就绪"
// @Success 503 {object} map[string]string "Service not ready | 服务未就绪"
// @Router /ready [get]
func (ctrl *HealthController) Ready(c *gin.Context) {
	isReady, details := ctrl.healthService.CheckReady(c.Request.Context())

	if isReady {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
		return
	}

	response := gin.H{
		"status": "not_ready",
	}
	for k, v := range details {
		response[k] = v
	}

	c.JSON(http.StatusServiceUnavailable, response)
}
