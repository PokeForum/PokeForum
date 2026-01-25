package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

var startTime = time.Now()

// IHealthService Health check service interface | 健康检查服务接口
type IHealthService interface {
	CheckHealth(ctx context.Context, detail bool) *schema.HealthStatus
	CheckReady(ctx context.Context) (bool, map[string]bool)
	Ping() string
}

// HealthService Health check service implementation | 健康检查服务实现
type HealthService struct {
	db       *ent.Client
	userRepo repository.IUserRepository
	cache    cache.ICacheService
}

// NewHealthService Create health check service instance | 创建健康检查服务实例
func NewHealthService(db *ent.Client, repos *repository.Repositories, cache cache.ICacheService) IHealthService {
	return &HealthService{
		db:       db,
		userRepo: repos.User,
		cache:    cache,
	}
}

// Ping Simple liveness probe | 简单存活检测
func (s *HealthService) Ping() string {
	return "pong"
}

// CheckHealth Detailed health check | 详细健康检查
func (s *HealthService) CheckHealth(ctx context.Context, detail bool) *schema.HealthStatus {
	checks := make(map[string]schema.Check)
	overallStatus := "healthy"

	// Check database connection | 检查数据库连接
	dbCheck := s.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status == "down" {
		overallStatus = "unhealthy"
	}

	// Check Redis connection | 检查Redis连接
	redisCheck := s.checkRedis(ctx)
	checks["redis"] = redisCheck
	if redisCheck.Status == "down" {
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	// Build response | 构建响应
	health := &schema.HealthStatus{
		Status:    overallStatus,
		Version:   _const.Version,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Checks:    checks,
	}

	// Add system information if detailed information is requested | 如果请求详细信息,添加系统信息
	if detail {
		health.System = s.getSystemInfo()
	}

	return health
}

// CheckReady Readiness check | 就绪检查
func (s *HealthService) CheckReady(ctx context.Context) (bool, map[string]bool) {
	// Check core dependencies | 检查核心依赖
	dbOK := s.checkDatabase(ctx).Status == "up"
	redisOK := s.checkRedis(ctx).Status == "up"

	details := map[string]bool{
		"database": dbOK,
		"redis":    redisOK,
	}

	return dbOK && redisOK, details
}

// checkDatabase Check database connection | 检查数据库连接
func (s *HealthService) checkDatabase(ctx context.Context) schema.Check {
	if s.db == nil {
		return schema.Check{
			Status:  "down",
			Message: "database client not initialized",
		}
	}

	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Execute simple query to test connection | 执行简单查询测试连接
	_, err := s.userRepo.Count(timeoutCtx)
	latency := time.Since(start)

	if err != nil {
		return schema.Check{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return schema.Check{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkRedis Check Redis connection | 检查Redis连接
func (s *HealthService) checkRedis(ctx context.Context) schema.Check {
	if s.cache == nil {
		return schema.Check{
			Status:  "down",
			Message: "redis client not initialized",
		}
	}

	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Execute PING command to test connection | 执行PING命令测试连接
	err := s.cache.Ping(timeoutCtx).Err()
	latency := time.Since(start)

	if err != nil {
		return schema.Check{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return schema.Check{
		Status:  "up",
		Latency: latency.String(),
	}
}

// getSystemInfo Get system information | 获取系统信息
func (s *HealthService) getSystemInfo() *schema.SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &schema.SystemInfo{
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
