package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/schema"
)

const (
	// Redis 存储 Key 前缀
	perfKeyPrefix = "perf:"
	// 原始数据保留时间 1 小时
	perfRawDataTTL = 1 * time.Hour
)

// IPerformanceService 性能监控服务接口
type IPerformanceService interface {
	// CollectSystemMetrics 采集系统指标
	CollectSystemMetrics(ctx context.Context) (*schema.SystemMetrics, error)
	// CollectPgSQLMetrics 采集 PostgreSQL 指标
	CollectPgSQLMetrics(ctx context.Context) (*schema.PostgreSQLMetrics, error)
	// CollectRedisMetrics 采集 Redis 指标
	CollectRedisMetrics(ctx context.Context) (*schema.RedisMetrics, error)
	// CollectAllMetrics 采集所有指标
	CollectAllMetrics(ctx context.Context, modules []string) (*schema.PerformanceWSResponse, error)
	// SaveMetrics 保存监控数据到 Redis
	SaveMetrics(ctx context.Context, data *schema.PerformanceWSResponse) error
	// GetHistoryMetrics 获取历史监控数据
	GetHistoryMetrics(ctx context.Context, req schema.PerformanceHistoryRequest) (*schema.PerformanceHistoryResponse, error)
}

// PerformanceService 性能监控服务实现
type PerformanceService struct {
	pgDB   *sql.DB       // PostgreSQL 原生连接
	redis  *redis.Client // Redis 客户端
	logger *zap.Logger
}

// NewPerformanceService 创建性能监控服务实例
func NewPerformanceService(pgDB *sql.DB, redisClient *redis.Client, logger *zap.Logger) IPerformanceService {
	return &PerformanceService{
		pgDB:   pgDB,
		redis:  redisClient,
		logger: logger,
	}
}

// CollectSystemMetrics 采集系统指标
func (s *PerformanceService) CollectSystemMetrics(ctx context.Context) (*schema.SystemMetrics, error) {
	metrics := &schema.SystemMetrics{}

	// 采集 CPU 指标
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPU.UsagePercent = cpuPercent[0]
	}

	cpuTimes, err := cpu.TimesWithContext(ctx, false)
	if err == nil && len(cpuTimes) > 0 {
		total := cpuTimes[0].User + cpuTimes[0].System + cpuTimes[0].Idle + cpuTimes[0].Nice +
			cpuTimes[0].Iowait + cpuTimes[0].Irq + cpuTimes[0].Softirq + cpuTimes[0].Steal
		if total > 0 {
			metrics.CPU.UserPercent = (cpuTimes[0].User / total) * 100
			metrics.CPU.SystemPercent = (cpuTimes[0].System / total) * 100
			metrics.CPU.IdlePercent = (cpuTimes[0].Idle / total) * 100
		}
	}
	metrics.CPU.Cores = runtime.NumCPU()

	// 采集内存指标
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		metrics.Memory.Total = memInfo.Total
		metrics.Memory.Used = memInfo.Used
		metrics.Memory.Available = memInfo.Available
		metrics.Memory.UsagePercent = memInfo.UsedPercent
	}

	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		metrics.Memory.SwapTotal = swapInfo.Total
		metrics.Memory.SwapUsed = swapInfo.Used
		metrics.Memory.SwapPercent = swapInfo.UsedPercent
	}

	// 采集磁盘指标
	diskUsage, err := disk.UsageWithContext(ctx, "/")
	if err == nil {
		metrics.Disk.Total = diskUsage.Total
		metrics.Disk.Used = diskUsage.Used
		metrics.Disk.Free = diskUsage.Free
		metrics.Disk.UsagePercent = diskUsage.UsedPercent
	}

	diskIO, err := disk.IOCountersWithContext(ctx)
	if err == nil {
		for _, io := range diskIO {
			metrics.Disk.ReadBytes += io.ReadBytes
			metrics.Disk.WriteBytes += io.WriteBytes
			metrics.Disk.ReadCount += io.ReadCount
			metrics.Disk.WriteCount += io.WriteCount
		}
	}

	// 采集网络指标
	netIO, err := net.IOCountersWithContext(ctx, false)
	if err == nil && len(netIO) > 0 {
		metrics.Network.BytesSent = netIO[0].BytesSent
		metrics.Network.BytesRecv = netIO[0].BytesRecv
		metrics.Network.PacketsSent = netIO[0].PacketsSent
		metrics.Network.PacketsRecv = netIO[0].PacketsRecv
		metrics.Network.ErrIn = netIO[0].Errin
		metrics.Network.ErrOut = netIO[0].Errout
		metrics.Network.DropIn = netIO[0].Dropin
		metrics.Network.DropOut = netIO[0].Dropout
	}

	// 获取网络连接数
	conns, err := net.ConnectionsWithContext(ctx, "all")
	if err == nil {
		metrics.Network.Connections = len(conns)
	}

	// 采集负载指标
	loadAvg, err := load.AvgWithContext(ctx)
	if err == nil {
		metrics.Load.Load1 = loadAvg.Load1
		metrics.Load.Load5 = loadAvg.Load5
		metrics.Load.Load15 = loadAvg.Load15
	}

	return metrics, nil
}

// CollectPgSQLMetrics 采集 PostgreSQL 指标
func (s *PerformanceService) CollectPgSQLMetrics(ctx context.Context) (*schema.PostgreSQLMetrics, error) {
	metrics := &schema.PostgreSQLMetrics{}

	if s.pgDB == nil {
		return metrics, fmt.Errorf("数据库连接不可用")
	}

	// 采集连接指标
	s.collectPgConnections(ctx, metrics)
	// 采集事务指标
	s.collectPgTransactions(ctx, metrics)
	// 采集缓存指标
	s.collectPgCache(ctx, metrics)
	// 采集数据库指标
	s.collectPgDatabase(ctx, metrics)
	// 采集锁指标
	s.collectPgLocks(ctx, metrics)
	// 采集复制指标
	s.collectPgReplication(ctx, metrics)

	return metrics, nil
}

// collectPgConnections 采集 PostgreSQL 连接指标
func (s *PerformanceService) collectPgConnections(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	// 查询连接状态分布
	rows, err := s.pgDB.QueryContext(ctx, `
		SELECT state, COUNT(*) 
		FROM pg_stat_activity 
		WHERE state IS NOT NULL 
		GROUP BY state
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var state string
			var count int
			if err := rows.Scan(&state, &count); err == nil {
				switch state {
				case "active":
					metrics.Connections.Active = count
				case "idle":
					metrics.Connections.Idle = count
				case "idle in transaction":
					metrics.Connections.IdleInTx = count
				}
				metrics.Connections.Total += count
			}
		}
	}

	// 查询等待连接数
	var waiting int
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type IS NOT NULL
	`).Scan(&waiting)
	if err == nil {
		metrics.Connections.Waiting = waiting
	}

	// 查询最大连接数
	var maxConnStr string
	err = s.pgDB.QueryRowContext(ctx, `SHOW max_connections`).Scan(&maxConnStr)
	if err == nil {
		maxConn, _ := strconv.Atoi(maxConnStr)
		metrics.Connections.MaxConn = maxConn
		if maxConn > 0 {
			metrics.Connections.UsagePercent = float64(metrics.Connections.Total) / float64(maxConn) * 100
		}
	}
}

// collectPgTransactions 采集 PostgreSQL 事务指标
func (s *PerformanceService) collectPgTransactions(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	var committed, rolledBack sql.NullInt64
	err := s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(xact_commit), SUM(xact_rollback) 
		FROM pg_stat_database
	`).Scan(&committed, &rolledBack)
	if err == nil {
		metrics.Transaction.Committed = committed.Int64
		metrics.Transaction.RolledBack = rolledBack.Int64
	}
}

// collectPgCache 采集 PostgreSQL 缓存指标
func (s *PerformanceService) collectPgCache(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	var blksHit, blksRead sql.NullInt64
	err := s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(blks_hit), SUM(blks_read) 
		FROM pg_stat_database
	`).Scan(&blksHit, &blksRead)
	if err == nil {
		metrics.Cache.BlocksHit = blksHit.Int64
		metrics.Cache.BlocksRead = blksRead.Int64
		total := blksHit.Int64 + blksRead.Int64
		if total > 0 {
			metrics.Cache.HitRatio = float64(blksHit.Int64) / float64(total) * 100
		}
	}

	// 临时文件统计
	var tempFiles, tempBytes sql.NullInt64
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(temp_files), SUM(temp_bytes) 
		FROM pg_stat_database
	`).Scan(&tempFiles, &tempBytes)
	if err == nil {
		metrics.Cache.TempFiles = tempFiles.Int64
		metrics.Cache.TempBytes = tempBytes.Int64
	}
}

// collectPgDatabase 采集 PostgreSQL 数据库指标
func (s *PerformanceService) collectPgDatabase(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	// 数据库大小
	var dbSize int64
	err := s.pgDB.QueryRowContext(ctx, `
		SELECT pg_database_size(current_database())
	`).Scan(&dbSize)
	if err == nil {
		metrics.Database.Size = dbSize
	}

	// 表数量
	var tableCount int
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`).Scan(&tableCount)
	if err == nil {
		metrics.Database.TableCount = tableCount
	}

	// 索引大小
	var indexSize sql.NullInt64
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(pg_indexes_size(quote_ident(schemaname) || '.' || quote_ident(tablename))) 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`).Scan(&indexSize)
	if err == nil {
		metrics.Database.IndexSize = indexSize.Int64
	}

	// 表大小
	var tableSize sql.NullInt64
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(pg_table_size(quote_ident(schemaname) || '.' || quote_ident(tablename))) 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`).Scan(&tableSize)
	if err == nil {
		metrics.Database.TableSize = tableSize.Int64
	}
}

// collectPgLocks 采集 PostgreSQL 锁指标
func (s *PerformanceService) collectPgLocks(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	// 锁统计
	rows, err := s.pgDB.QueryContext(ctx, `
		SELECT mode, COUNT(*), SUM(CASE WHEN granted THEN 0 ELSE 1 END) as waiting
		FROM pg_locks 
		GROUP BY mode
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mode string
			var count, waiting int
			if err := rows.Scan(&mode, &count, &waiting); err == nil {
				metrics.Locks.Total += count
				metrics.Locks.Waiting += waiting
				switch mode {
				case "AccessShareLock":
					metrics.Locks.AccessShare = count
				case "RowShareLock":
					metrics.Locks.RowShare = count
				case "RowExclusiveLock":
					metrics.Locks.RowExcl = count
				}
			}
		}
	}

	// 死锁统计
	var deadlocks sql.NullInt64
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT SUM(deadlocks) FROM pg_stat_database
	`).Scan(&deadlocks)
	if err == nil {
		metrics.Locks.Deadlocks = deadlocks.Int64
	}
}

// collectPgReplication 采集 PostgreSQL 复制指标
func (s *PerformanceService) collectPgReplication(ctx context.Context, metrics *schema.PostgreSQLMetrics) {
	// 检查是否为副本
	var isReplica bool
	err := s.pgDB.QueryRowContext(ctx, `SELECT pg_is_in_recovery()`).Scan(&isReplica)
	if err == nil {
		metrics.Replication.IsReplica = isReplica
	}

	// 副本数量
	var replicaCount int
	err = s.pgDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_stat_replication
	`).Scan(&replicaCount)
	if err == nil {
		metrics.Replication.ReplicaCount = replicaCount
	}
}

// CollectRedisMetrics 采集 Redis 指标
func (s *PerformanceService) CollectRedisMetrics(ctx context.Context) (*schema.RedisMetrics, error) {
	metrics := &schema.RedisMetrics{}

	// 获取 Redis INFO
	info, err := s.redis.Info(ctx, "all").Result()
	if err != nil {
		return metrics, fmt.Errorf("获取 Redis INFO 失败: %w", err)
	}

	// 解析 INFO 信息
	infoMap := parseRedisInfo(info)

	// 内存指标
	metrics.Memory.Used = parseUint64(infoMap["used_memory"])
	metrics.Memory.UsedPeak = parseUint64(infoMap["used_memory_peak"])
	metrics.Memory.UsedRSS = parseUint64(infoMap["used_memory_rss"])
	metrics.Memory.FragmentationRatio = parseFloat64(infoMap["mem_fragmentation_ratio"])
	metrics.Memory.MaxMemory = parseUint64(infoMap["maxmemory"])
	if metrics.Memory.MaxMemory > 0 {
		metrics.Memory.UsagePercent = float64(metrics.Memory.Used) / float64(metrics.Memory.MaxMemory) * 100
	}

	// 连接指标
	metrics.Connections.Connected = parseInt(infoMap["connected_clients"])
	metrics.Connections.Blocked = parseInt(infoMap["blocked_clients"])
	metrics.Connections.MaxClients = parseInt(infoMap["maxclients"])
	metrics.Connections.TotalConnections = parseInt64(infoMap["total_connections_received"])
	metrics.Connections.RejectedConns = parseInt64(infoMap["rejected_connections"])

	// 操作指标
	metrics.Operations.OpsPerSec = parseInt64(infoMap["instantaneous_ops_per_sec"])
	metrics.Operations.TotalCommands = parseInt64(infoMap["total_commands_processed"])
	metrics.Operations.Hits = parseInt64(infoMap["keyspace_hits"])
	metrics.Operations.Misses = parseInt64(infoMap["keyspace_misses"])
	totalHits := metrics.Operations.Hits + metrics.Operations.Misses
	if totalHits > 0 {
		metrics.Operations.HitRate = float64(metrics.Operations.Hits) / float64(totalHits) * 100
	}
	metrics.Operations.ExpiredKeys = parseInt64(infoMap["expired_keys"])
	metrics.Operations.EvictedKeys = parseInt64(infoMap["evicted_keys"])

	// 键空间指标
	metrics.Keyspace.TotalKeys = s.getRedisKeyCount(infoMap)
	metrics.Keyspace.ExpiresKeys = s.getRedisExpiresCount(infoMap)

	// 持久化指标
	metrics.Persistence.RDBLastSaveTime = parseInt64(infoMap["rdb_last_save_time"])
	metrics.Persistence.RDBChangesSince = parseInt64(infoMap["rdb_changes_since_last_save"])
	metrics.Persistence.AOFEnabled = infoMap["aof_enabled"] == "1"
	metrics.Persistence.AOFRewriteInProg = infoMap["aof_rewrite_in_progress"] == "1"
	metrics.Persistence.AOFCurrentSize = parseInt64(infoMap["aof_current_size"])

	// 复制指标
	metrics.Replication.Role = infoMap["role"]
	metrics.Replication.ConnectedSlaves = parseInt(infoMap["connected_slaves"])
	metrics.Replication.MasterLinkStatus = infoMap["master_link_status"]
	metrics.Replication.MasterLastIO = parseInt64(infoMap["master_last_io_seconds_ago"])

	return metrics, nil
}

// getRedisKeyCount 获取 Redis 键总数
func (s *PerformanceService) getRedisKeyCount(infoMap map[string]string) int64 {
	var total int64
	for key, value := range infoMap {
		if strings.HasPrefix(key, "db") {
			parts := strings.Split(value, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "keys=") {
					count, _ := strconv.ParseInt(strings.TrimPrefix(part, "keys="), 10, 64)
					total += count
				}
			}
		}
	}
	return total
}

// getRedisExpiresCount 获取设置过期的键数
func (s *PerformanceService) getRedisExpiresCount(infoMap map[string]string) int64 {
	var total int64
	for key, value := range infoMap {
		if strings.HasPrefix(key, "db") {
			parts := strings.Split(value, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "expires=") {
					count, _ := strconv.ParseInt(strings.TrimPrefix(part, "expires="), 10, 64)
					total += count
				}
			}
		}
	}
	return total
}

// CollectAllMetrics 采集所有指标
func (s *PerformanceService) CollectAllMetrics(ctx context.Context, modules []string) (*schema.PerformanceWSResponse, error) {
	response := &schema.PerformanceWSResponse{
		Timestamp: time.Now().Unix(),
	}

	for _, module := range modules {
		switch module {
		case "system":
			metrics, err := s.CollectSystemMetrics(ctx)
			if err != nil {
				s.logger.Error("采集系统指标失败", zap.Error(err))
			} else {
				response.System = metrics
			}
		case "pgsql":
			metrics, err := s.CollectPgSQLMetrics(ctx)
			if err != nil {
				s.logger.Error("采集 PostgreSQL 指标失败", zap.Error(err))
			} else {
				response.PgSQL = metrics
			}
		case "redis":
			metrics, err := s.CollectRedisMetrics(ctx)
			if err != nil {
				s.logger.Error("采集 Redis 指标失败", zap.Error(err))
			} else {
				response.Redis = metrics
			}
		}
	}

	return response, nil
}

// SaveMetrics 保存监控数据到 Redis
func (s *PerformanceService) SaveMetrics(ctx context.Context, data *schema.PerformanceWSResponse) error {
	timestamp := data.Timestamp
	pipe := s.redis.Pipeline()

	// 保存系统指标
	if data.System != nil {
		key := fmt.Sprintf("%ssystem:%d", perfKeyPrefix, timestamp)
		jsonData, _ := json.Marshal(data.System)
		pipe.Set(ctx, key, jsonData, perfRawDataTTL)
	}

	// 保存 PostgreSQL 指标
	if data.PgSQL != nil {
		key := fmt.Sprintf("%spgsql:%d", perfKeyPrefix, timestamp)
		jsonData, _ := json.Marshal(data.PgSQL)
		pipe.Set(ctx, key, jsonData, perfRawDataTTL)
	}

	// 保存 Redis 指标
	if data.Redis != nil {
		key := fmt.Sprintf("%sredis:%d", perfKeyPrefix, timestamp)
		jsonData, _ := json.Marshal(data.Redis)
		pipe.Set(ctx, key, jsonData, perfRawDataTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetHistoryMetrics 获取历史监控数据
func (s *PerformanceService) GetHistoryMetrics(ctx context.Context, req schema.PerformanceHistoryRequest) (*schema.PerformanceHistoryResponse, error) {
	response := &schema.PerformanceHistoryResponse{
		Module:    req.Module,
		StartTime: req.Start,
		EndTime:   req.End,
		Interval:  req.Interval,
		Data:      make([]interface{}, 0),
	}

	// 根据间隔确定步长
	var step int64
	switch req.Interval {
	case "1m":
		step = 60
	case "5m":
		step = 300
	case "1h":
		step = 3600
	case "1d":
		step = 86400
	default:
		step = 60
	}

	// 遍历时间范围获取数据
	for ts := req.Start; ts <= req.End; ts += step {
		key := fmt.Sprintf("%s%s:%d", perfKeyPrefix, req.Module, ts)
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil && data != "" {
			var metrics interface{}
			switch req.Module {
			case "system":
				var m schema.SystemMetrics
				if json.Unmarshal([]byte(data), &m) == nil {
					metrics = map[string]interface{}{
						"timestamp": ts,
						"data":      m,
					}
				}
			case "pgsql":
				var m schema.PostgreSQLMetrics
				if json.Unmarshal([]byte(data), &m) == nil {
					metrics = map[string]interface{}{
						"timestamp": ts,
						"data":      m,
					}
				}
			case "redis":
				var m schema.RedisMetrics
				if json.Unmarshal([]byte(data), &m) == nil {
					metrics = map[string]interface{}{
						"timestamp": ts,
						"data":      m,
					}
				}
			}
			if metrics != nil {
				response.Data = append(response.Data, metrics)
			}
		}
	}

	return response, nil
}

// ========== 辅助函数 ==========

// parseRedisInfo 解析 Redis INFO 输出
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func parseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
