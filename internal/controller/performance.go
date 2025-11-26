package controller

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"go.uber.org/zap"
)

// PerformanceController 性能监控控制器
type PerformanceController struct {
	injector *do.Injector
	logger   *zap.Logger
}

// NewPerformanceController 创建性能监控控制器实例
func NewPerformanceController(injector *do.Injector) *PerformanceController {
	logger, _ := do.Invoke[*zap.Logger](injector)
	return &PerformanceController{
		injector: injector,
		logger:   logger,
	}
}

// PerformanceRouter 注册性能监控路由
func (ctrl *PerformanceController) PerformanceRouter(router *gin.RouterGroup) {
	router.GET("/stream", ctrl.HandleSSE)
	router.GET("/history", ctrl.GetHistoryMetrics)
}

// HandleSSE 处理 SSE 连接
// @Summary 性能监控 SSE 流
// @Description 通过 SSE 实时推送性能监控数据
// @Tags [超级管理员]性能监控
// @Produce text/event-stream
// @Param modules query string false "监控模块，逗号分隔" default(system,pgsql,redis)
// @Param interval query int false "推送间隔秒数" default(3) minimum(1) maximum(60)
// @Success 200 {object} schema.PerformanceWSResponse "SSE 数据流"
// @Router /super/manage/performance/stream [get]
// @Security Bearer
func (ctrl *PerformanceController) HandleSSE(c *gin.Context) {
	// 解析请求参数
	var req schema.PerformanceSSERequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 设置默认值
	modules := []string{"system", "pgsql", "redis"}
	if req.Modules != "" {
		modules = strings.Split(req.Modules, ",")
		// 过滤有效模块
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

	// 获取服务
	perfService, err := do.Invoke[service.IPerformanceService](ctrl.injector)
	if err != nil {
		ctrl.logger.Error("获取性能监控服务失败", zap.Error(err))
		response.ResError(c, response.CodeServerBusy)
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 创建上下文用于取消
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// 监听客户端断开
	clientGone := c.Writer.CloseNotify()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 立即发送第一次数据
	ctrl.sendMetrics(c, ctx, perfService, modules)

	for {
		select {
		case <-ticker.C:
			ctrl.sendMetrics(c, ctx, perfService, modules)
		case <-clientGone:
			ctrl.logger.Debug("SSE 客户端断开连接")
			return
		case <-ctx.Done():
			return
		}
	}
}

// sendMetrics 发送监控数据
func (ctrl *PerformanceController) sendMetrics(c *gin.Context, ctx context.Context, perfService service.IPerformanceService, modules []string) {
	metrics, err := perfService.CollectAllMetrics(ctx, modules)
	if err != nil {
		ctrl.logger.Error("采集监控指标失败", zap.Error(err))
		return
	}

	// 保存到 Redis
	if err := perfService.SaveMetrics(ctx, metrics); err != nil {
		ctrl.logger.Warn("保存监控数据失败", zap.Error(err))
	}

	// 序列化数据
	data, err := json.Marshal(metrics)
	if err != nil {
		ctrl.logger.Error("序列化监控数据失败", zap.Error(err))
		return
	}

	// 发送 SSE 事件
	_, err = io.WriteString(c.Writer, "data: "+string(data)+"\n\n")
	if err != nil {
		ctrl.logger.Debug("SSE 写入失败", zap.Error(err))
		return
	}
	c.Writer.Flush()
}

// GetHistoryMetrics 获取历史监控数据
// @Summary 获取历史监控数据
// @Description 查询指定时间范围内的历史监控数据
// @Tags [超级管理员]性能监控
// @Accept json
// @Produce json
// @Param module query string false "监控模块" Enums(system, pgsql, redis) default(system)
// @Param start query int false "开始时间戳（默认1小时前）"
// @Param end query int false "结束时间戳（默认当前时间）"
// @Param interval query string false "数据间隔" Enums(1m, 5m, 1h, 1d) default(1m)
// @Success 200 {object} response.Data{data=schema.PerformanceHistoryResponse} "获取成功"
// @Failure 400 {object} response.Data "请求参数错误"
// @Failure 500 {object} response.Data "服务器错误"
// @Router /super/manage/performance/history [get]
// @Security Bearer
func (ctrl *PerformanceController) GetHistoryMetrics(c *gin.Context) {
	var req schema.PerformanceHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 设置默认值
	now := time.Now().Unix()
	if req.Module == "" {
		req.Module = "system"
	}
	if req.End == 0 {
		req.End = now
	}
	if req.Start == 0 {
		req.Start = now - 3600 // 默认1小时前
	}
	if req.Interval == "" {
		req.Interval = "1m"
	}

	// 验证时间范围不超过 30 天
	maxRange := int64(30 * 24 * 60 * 60)
	if req.End-req.Start > maxRange {
		response.ResErrorWithMsg(c, response.CodeInvalidParam, "时间范围不能超过30天")
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
