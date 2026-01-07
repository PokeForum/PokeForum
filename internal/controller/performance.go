package controller

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
)

// PerformanceController Performance monitoring controller | 性能监控控制器
type PerformanceController struct {
	injector *do.Injector
	logger   *zap.Logger
}

// NewPerformanceController Create performance monitoring controller instance | 创建性能监控控制器实例
func NewPerformanceController(injector *do.Injector) *PerformanceController {
	logger, _ := do.Invoke[*zap.Logger](injector) //nolint:errcheck // Returns nil if dependency injection fails | 依赖注入失败时返回nil
	return &PerformanceController{
		injector: injector,
		logger:   logger,
	}
}

// PerformanceRouter Register performance monitoring routes | 注册性能监控路由
func (ctrl *PerformanceController) PerformanceRouter(router *gin.RouterGroup) {
	router.GET("/stream", ctrl.HandleSSE)
	router.GET("/history", ctrl.GetHistoryMetrics)
}

// HandleSSE Handle SSE connection | 处理 SSE 连接
// @Summary Performance monitoring SSE stream | 性能监控 SSE 流
// @Description Push performance monitoring data in real-time via SSE | 通过 SSE 实时推送性能监控数据
// @Tags [Super Admin]Performance Monitoring | [超级管理员]性能监控
// @Produce text/event-stream
// @Param modules query string false "Monitoring modules, comma separated | 监控模块，逗号分隔" default(system,pgsql,redis)
// @Param interval query int false "Push interval in seconds | 推送间隔秒数" default(3) minimum(1) maximum(60)
// @Success 200 {object} schema.PerformanceWSResponse "SSE data stream | SSE 数据流"
// @Router /super/manage/performance/stream [get]
// @Security Bearer
func (ctrl *PerformanceController) HandleSSE(c *gin.Context) {
	// Parse request parameters | 解析请求参数
	var req schema.PerformanceSSERequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Set default values | 设置默认值
	modules := []string{"system", "pgsql", "redis"}
	if req.Modules != "" {
		modules = strings.Split(req.Modules, ",")
		// Filter valid modules | 过滤有效模块
		validModules := make([]string, 0)
		for _, m := range modules {
			m = strings.TrimSpace(m)
			if m == "system" || m == "pgsql" || m == "redis" {
				validModules = append(validModules, m)
			}
		}
		if len(validModules) > 0 {
			modules = validModules
		}
	}

	interval := 3
	if req.Interval > 0 {
		interval = req.Interval
		if interval < 1 {
			interval = 1
		}
		if interval > 60 {
			interval = 60
		}
	}

	// Get service | 获取服务
	perfService, err := do.Invoke[service.IPerformanceService](ctrl.injector)
	if err != nil {
		ctrl.logger.Error("Failed to get performance monitoring service | 获取性能监控服务失败", zap.Error(err))
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// Set SSE response headers | 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering | 禁用 Nginx 缓冲

	// Create context for cancellation | 创建上下文用于取消
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Listen for client disconnection | 监听客户端断开
	clientGone := c.Writer.CloseNotify()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Send first data immediately | 立即发送第一次数据
	ctrl.sendMetrics(c, ctx, perfService, modules)

	for {
		select {
		case <-ticker.C:
			ctrl.sendMetrics(c, ctx, perfService, modules)
		case <-clientGone:
			ctrl.logger.Debug("SSE client disconnected | SSE 客户端断开连接")
			return
		case <-ctx.Done():
			return
		}
	}
}

// sendMetrics Send monitoring data | 发送监控数据
func (ctrl *PerformanceController) sendMetrics(c *gin.Context, ctx context.Context, perfService service.IPerformanceService, modules []string) {
	metrics, err := perfService.CollectAllMetrics(ctx, modules)
	if err != nil {
		ctrl.logger.Error("Failed to collect monitoring metrics | 采集监控指标失败", zap.Error(err))
		return
	}

	// Save to Redis | 保存到 Redis
	if err := perfService.SaveMetrics(ctx, metrics); err != nil {
		ctrl.logger.Warn("Failed to save monitoring data | 保存监控数据失败", zap.Error(err))
	}

	// Serialize data | 序列化数据
	data, err := json.Marshal(metrics)
	if err != nil {
		ctrl.logger.Error("Failed to serialize monitoring data | 序列化监控数据失败", zap.Error(err))
		return
	}

	// Send SSE event | 发送 SSE 事件
	_, err = c.Writer.WriteString("data: " + string(data) + "\n\n")
	if err != nil {
		ctrl.logger.Debug("SSE write failed | SSE 写入失败", zap.Error(err))
		return
	}
	c.Writer.Flush()
}

// GetHistoryMetrics Get historical monitoring data | 获取历史监控数据
// @Summary Get historical monitoring data | 获取历史监控数据
// @Description Query historical monitoring data within specified time range | 查询指定时间范围内的历史监控数据
// @Tags [Super Admin]Performance Monitoring | [超级管理员]性能监控
// @Accept json
// @Produce json
// @Param module query string false "Monitoring module | 监控模块" Enums(system, pgsql, redis) default(system)
// @Param start query int false "Start timestamp (default 1 hour ago) | 开始时间戳（默认1小时前）"
// @Param end query int false "End timestamp (default current time) | 结束时间戳（默认当前时间）"
// @Param interval query string false "Data interval | 数据间隔" Enums(1m, 5m, 1h, 1d) default(1m)
// @Success 200 {object} response.Data{data=schema.PerformanceHistoryResponse} "Success | 获取成功"
// @Failure 400 {object} response.Data "Invalid request parameters | 请求参数错误"
// @Failure 500 {object} response.Data "Server error | 服务器错误"
// @Router /super/manage/performance/history [get]
// @Security Bearer
func (ctrl *PerformanceController) GetHistoryMetrics(c *gin.Context) {
	var req schema.PerformanceHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// Set default values | 设置默认值
	now := time.Now().Unix()
	if req.Module == "" {
		req.Module = "system"
	}
	if req.End == 0 {
		req.End = now
	}
	if req.Start == 0 {
		req.Start = now - 3600 // Default 1 hour ago | 默认1小时前
	}
	if req.Interval == "" {
		req.Interval = "1m"
	}

	// Validate time range does not exceed 30 days | 验证时间范围不超过 30 天
	maxRange := int64(30 * 24 * 60 * 60)
	if req.End-req.Start > maxRange {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "Time range cannot exceed 30 days | 时间范围不能超过30天")
		return
	}

	perfService, err := do.Invoke[service.IPerformanceService](ctrl.injector)
	if err != nil {
		response.ResError(c, response.CodeServerBusy)
		return
	}

	result, err := perfService.GetHistoryMetrics(c.Request.Context(), req)
	if err != nil {
		response.ResErrorWithMsg(c, response.CodeGenericError, err.Error())
		return
	}

	response.ResSuccess(c, result)
}
