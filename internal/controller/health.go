package controller

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/configs"
	_const "github.com/PokeForum/PokeForum/internal/consts"
)

// HealthController 健康检查控制器
type HealthController struct{}

// NewHealthController 创建健康检查控制器实例
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthStatus 健康状态响应
type HealthStatus struct {
	Status    string           `json:"status"`           // 整体状态: healthy, degraded, unhealthy
	Version   string           `json:"version"`          // 应用版本
	Timestamp string           `json:"timestamp"`        // 检查时间
	Uptime    string           `json:"uptime,omitempty"` // 运行时间
	Checks    map[string]Check `json:"checks"`           // 各组件检查结果
	System    *SystemInfo      `json:"system,omitempty"` // 系统信息（仅详细模式）
}

// Check 单个组件检查结果
type Check struct {
	Status  string `json:"status"`            // 状态: up, down
	Message string `json:"message,omitempty"` // 额外信息
	Latency string `json:"latency,omitempty"` // 响应延迟
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAlloc     string `json:"mem_alloc"`
	MemSys       string `json:"mem_sys"`
}

var startTime = time.Now()

// HealthRouter 注册健康检查路由
func (ctrl *HealthController) HealthRouter(router *gin.Engine) {
	// 简单存活检测（用于负载均衡器）
	router.GET("/ping", ctrl.Ping)

	// 详细健康检查（用于监控系统）
	router.GET("/health", ctrl.Health)

	// 就绪检查（用于Kubernetes等编排系统）
	router.GET("/ready", ctrl.Ready)
}

// Ping 简单存活检测
// @Summary 存活检测
// @Description 简单的存活检测，返回pong
// @Tags 健康检查
// @Produce plain
// @Success 200 {string} string "pong"
// @Router /ping [get]
func (ctrl *HealthController) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// Health 详细健康检查
// @Summary 详细健康检查
// @Description 检查所有依赖服务的健康状态
// @Tags 健康检查
// @Produce json
// @Param detail query bool false "是否返回系统详细信息"
// @Success 200 {object} HealthStatus "服务健康"
// @Success 503 {object} HealthStatus "服务不健康"
// @Router /health [get]
func (ctrl *HealthController) Health(c *gin.Context) {
	checks := make(map[string]Check)
	overallStatus := "healthy"

	// 检查数据库连接
	dbCheck := ctrl.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status == "down" {
		overallStatus = "unhealthy"
	}

	// 检查Redis连接
	redisCheck := ctrl.checkRedis()
	checks["redis"] = redisCheck
	if redisCheck.Status == "down" {
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	// 构建响应
	health := HealthStatus{
		Status:    overallStatus,
		Version:   _const.Version,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Checks:    checks,
	}

	// 如果请求详细信息，添加系统信息
	if c.Query("detail") == "true" {
		health.System = ctrl.getSystemInfo()
	}

	// 根据状态返回不同的HTTP状态码
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// Ready 就绪检查
// @Summary 就绪检查
// @Description 检查服务是否准备好接收流量
// @Tags 健康检查
// @Produce json
// @Success 200 {object} map[string]string "服务就绪"
// @Success 503 {object} map[string]string "服务未就绪"
// @Router /ready [get]
func (ctrl *HealthController) Ready(c *gin.Context) {
	// 检查核心依赖
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

// checkDatabase 检查数据库连接
func (ctrl *HealthController) checkDatabase() Check {
	if configs.DB == nil {
		return Check{
			Status:  "down",
			Message: "database client not initialized",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行简单查询测试连接
	_, err := configs.DB.User.Query().Limit(1).Count(ctx)
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

// checkRedis 检查Redis连接
func (ctrl *HealthController) checkRedis() Check {
	if configs.Cache == nil {
		return Check{
			Status:  "down",
			Message: "redis client not initialized",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行PING命令测试连接
	err := configs.Cache.Ping(ctx).Err()
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

// getSystemInfo 获取系统信息
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

// formatBytes 格式化字节数
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
