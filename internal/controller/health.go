package controller

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/ent"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/samber/do"
)

// HealthController Health check controller | 健康检查控制器
type HealthController struct {
	db    *ent.Client
	cache cache.ICacheService
}

// NewHealthController Create health check controller instance | 创建健康检查控制器实例
func NewHealthController(injector *do.Injector) *HealthController {
	return &HealthController{
		db:    do.MustInvoke[*ent.Client](injector),
		cache: do.MustInvoke[cache.ICacheService](injector),
	}
}

// HealthStatus Health status response | 健康状态响应
type HealthStatus struct {
	Status    string           `json:"status"`           // Overall status: healthy, degraded, unhealthy | 整体状态: healthy, degraded, unhealthy
	Version   string           `json:"version"`          // Application version | 应用版本
	Timestamp string           `json:"timestamp"`        // Check time | 检查时间
	Uptime    string           `json:"uptime,omitempty"` // Uptime | 运行时间
	Checks    map[string]Check `json:"checks"`           // Component check results | 各组件检查结果
	System    *SystemInfo      `json:"system,omitempty"` // System information (detail mode only) | 系统信息(仅详细模式)
}

// Check Single component check result | 单个组件检查结果
type Check struct {
	Status  string `json:"status"`            // Status: up, down | 状态: up, down
	Message string `json:"message,omitempty"` // Additional information | 额外信息
	Latency string `json:"latency,omitempty"` // Response latency | 响应延迟
}

// SystemInfo System information | 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAlloc     string `json:"mem_alloc"`
	MemSys       string `json:"mem_sys"`
}

var startTime = time.Now()

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
// @Success 200 {object} HealthStatus "Service healthy | 服务健康"
// @Success 503 {object} HealthStatus "Service unhealthy | 服务不健康"
// @Router /health [get]
func (ctrl *HealthController) Health(c *gin.Context) {
	checks := make(map[string]Check)
	overallStatus := "healthy"

	// Check database connection | 检查数据库连接
	dbCheck := ctrl.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status == "down" {
		overallStatus = "unhealthy"
	}

	// Check Redis connection | 检查Redis连接
	redisCheck := ctrl.checkRedis()
	checks["redis"] = redisCheck
	if redisCheck.Status == "down" {
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	// Build response | 构建响应
	health := HealthStatus{
		Status:    overallStatus,
		Version:   _const.Version,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Checks:    checks,
	}

	// Add system information if detailed information is requested | 如果请求详细信息,添加系统信息
	if c.Query("detail") == "true" {
		health.System = ctrl.getSystemInfo()
	}

	// Return different HTTP status codes based on status | 根据状态返回不同的HTTP状态码
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
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
	// Check core dependencies | 检查核心依赖
	dbOK := ctrl.checkDatabase().Status == "up"
	redisOK := ctrl.checkRedis().Status == "up"

	if dbOK && redisOK {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
		return
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"status":   "not_ready",
		"database": dbOK,
		"redis":    redisOK,
	})
}

// checkDatabase Check database connection | 检查数据库连接
func (ctrl *HealthController) checkDatabase() Check {
	if ctrl.db == nil {
		return Check{
			Status:  "down",
			Message: "database client not initialized",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute simple query to test connection | 执行简单查询测试连接
	_, err := ctrl.db.User.Query().Limit(1).Count(ctx)
	latency := time.Since(start)

	if err != nil {
		return Check{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return Check{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkRedis Check Redis connection | 检查Redis连接
func (ctrl *HealthController) checkRedis() Check {
	if ctrl.cache == nil {
		return Check{
			Status:  "down",
			Message: "redis client not initialized",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute PING command to test connection | 执行PING命令测试连接
	err := ctrl.cache.Ping(ctx).Err()
	latency := time.Since(start)

	if err != nil {
		return Check{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return Check{
		Status:  "up",
		Latency: latency.String(),
	}
}

// getSystemInfo Get system information | 获取系统信息
func (ctrl *HealthController) getSystemInfo() *SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAlloc:     formatBytes(m.Alloc),
		MemSys:       formatBytes(m.Sys),
	}
}

// formatBytes Format byte count | 格式化字节数
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
