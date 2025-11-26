package schema

// PerformanceSSERequest SSE 客户端请求参数
type PerformanceSSERequest struct {
	Modules  string `form:"modules"`  // 监控模块，逗号分隔: system,pgsql,redis（默认全部）
	Interval int    `form:"interval"` // 推送间隔秒数（默认3秒，最小1秒，最大60秒）
}

// PerformanceWSResponse WebSocket 服务端推送消息
type PerformanceWSResponse struct {
	Timestamp int64              `json:"timestamp"`        // 时间戳
	System    *SystemMetrics     `json:"system,omitempty"` // 系统指标
	PgSQL     *PostgreSQLMetrics `json:"pgsql,omitempty"`  // PostgreSQL 指标
	Redis     *RedisMetrics      `json:"redis,omitempty"`  // Redis 指标
}

// ========== 系统监控指标 ==========

// SystemMetrics 系统监控指标
type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`     // CPU 指标
	Memory  MemoryMetrics  `json:"memory"`  // 内存指标
	Disk    DiskMetrics    `json:"disk"`    // 磁盘指标
	Network NetworkMetrics `json:"network"` // 网络指标
	Load    LoadMetrics    `json:"load"`    // 负载指标
}

// CPUMetrics CPU 指标
type CPUMetrics struct {
	UsagePercent  float64 `json:"usage_percent"`  // CPU 总使用率
	UserPercent   float64 `json:"user_percent"`   // 用户态 CPU
	SystemPercent float64 `json:"system_percent"` // 内核态 CPU
	IdlePercent   float64 `json:"idle_percent"`   // 空闲 CPU
	Cores         int     `json:"cores"`          // CPU 核心数
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Total        uint64  `json:"total"`         // 总内存 (bytes)
	Used         uint64  `json:"used"`          // 已用内存 (bytes)
	Available    uint64  `json:"available"`     // 可用内存 (bytes)
	UsagePercent float64 `json:"usage_percent"` // 内存使用率
	SwapTotal    uint64  `json:"swap_total"`    // 交换分区总量 (bytes)
	SwapUsed     uint64  `json:"swap_used"`     // 交换分区已用 (bytes)
	SwapPercent  float64 `json:"swap_percent"`  // 交换分区使用率
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	Total        uint64  `json:"total"`         // 磁盘总量 (bytes)
	Used         uint64  `json:"used"`          // 已用空间 (bytes)
	Free         uint64  `json:"free"`          // 可用空间 (bytes)
	UsagePercent float64 `json:"usage_percent"` // 使用率
	ReadBytes    uint64  `json:"read_bytes"`    // 读取字节数
	WriteBytes   uint64  `json:"write_bytes"`   // 写入字节数
	ReadCount    uint64  `json:"read_count"`    // 读取次数
	WriteCount   uint64  `json:"write_count"`   // 写入次数
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`   // 发送字节数
	BytesRecv   uint64 `json:"bytes_recv"`   // 接收字节数
	PacketsSent uint64 `json:"packets_sent"` // 发送包数
	PacketsRecv uint64 `json:"packets_recv"` // 接收包数
	ErrIn       uint64 `json:"err_in"`       // 接收错误数
	ErrOut      uint64 `json:"err_out"`      // 发送错误数
	DropIn      uint64 `json:"drop_in"`      // 接收丢包数
	DropOut     uint64 `json:"drop_out"`     // 发送丢包数
	Connections int    `json:"connections"`  // 连接数
}

// LoadMetrics 负载指标
type LoadMetrics struct {
	Load1  float64 `json:"load1"`  // 1分钟负载
	Load5  float64 `json:"load5"`  // 5分钟负载
	Load15 float64 `json:"load15"` // 15分钟负载
}

// ========== PostgreSQL 监控指标 ==========

// PostgreSQLMetrics PostgreSQL 监控指标
type PostgreSQLMetrics struct {
	Connections PgConnections `json:"connections"` // 连接指标
	Transaction PgTransaction `json:"transaction"` // 事务指标
	Cache       PgCache       `json:"cache"`       // 缓存指标
	Database    PgDatabase    `json:"database"`    // 数据库指标
	Locks       PgLocks       `json:"locks"`       // 锁指标
	Replication PgReplication `json:"replication"` // 复制指标
}

// PgConnections PostgreSQL 连接指标
type PgConnections struct {
	Active       int     `json:"active"`        // 活跃连接数
	Idle         int     `json:"idle"`          // 空闲连接数
	IdleInTx     int     `json:"idle_in_tx"`    // 事务中空闲连接数
	Waiting      int     `json:"waiting"`       // 等待连接数
	Total        int     `json:"total"`         // 总连接数
	MaxConn      int     `json:"max_conn"`      // 最大连接数
	UsagePercent float64 `json:"usage_percent"` // 连接使用率
}

// PgTransaction PostgreSQL 事务指标
type PgTransaction struct {
	Committed  int64   `json:"committed"`   // 已提交事务数
	RolledBack int64   `json:"rolled_back"` // 已回滚事务数
	TPS        float64 `json:"tps"`         // 每秒事务数
}

// PgCache PostgreSQL 缓存指标
type PgCache struct {
	BlocksHit  int64   `json:"blocks_hit"`  // 缓存命中块数
	BlocksRead int64   `json:"blocks_read"` // 磁盘读取块数
	HitRatio   float64 `json:"hit_ratio"`   // 缓存命中率
	TempFiles  int64   `json:"temp_files"`  // 临时文件数
	TempBytes  int64   `json:"temp_bytes"`  // 临时文件大小
}

// PgDatabase PostgreSQL 数据库指标
type PgDatabase struct {
	Size       int64 `json:"size"`        // 数据库大小 (bytes)
	TableCount int   `json:"table_count"` // 表数量
	IndexSize  int64 `json:"index_size"`  // 索引大小 (bytes)
	TableSize  int64 `json:"table_size"`  // 表大小 (bytes)
}

// PgLocks PostgreSQL 锁指标
type PgLocks struct {
	Total       int   `json:"total"`        // 锁总数
	Waiting     int   `json:"waiting"`      // 等待锁数
	AccessShare int   `json:"access_share"` // AccessShare 锁数
	RowShare    int   `json:"row_share"`    // RowShare 锁数
	RowExcl     int   `json:"row_excl"`     // RowExclusive 锁数
	Deadlocks   int64 `json:"deadlocks"`    // 死锁数
}

// PgReplication PostgreSQL 复制指标
type PgReplication struct {
	IsReplica    bool  `json:"is_replica"`    // 是否为副本
	ReplicaCount int   `json:"replica_count"` // 副本数量
	LagBytes     int64 `json:"lag_bytes"`     // 复制延迟 (bytes)
}

// ========== Redis 监控指标 ==========

// RedisMetrics Redis 监控指标
type RedisMetrics struct {
	Memory      RedisMemory      `json:"memory"`      // 内存指标
	Connections RedisConnections `json:"connections"` // 连接指标
	Operations  RedisOperations  `json:"operations"`  // 操作指标
	Keyspace    RedisKeyspace    `json:"keyspace"`    // 键空间指标
	Persistence RedisPersistence `json:"persistence"` // 持久化指标
	Replication RedisReplication `json:"replication"` // 复制指标
}

// RedisMemory Redis 内存指标
type RedisMemory struct {
	Used               uint64  `json:"used"`                // 已用内存 (bytes)
	UsedPeak           uint64  `json:"used_peak"`           // 内存峰值 (bytes)
	UsedRSS            uint64  `json:"used_rss"`            // RSS 内存 (bytes)
	FragmentationRatio float64 `json:"fragmentation_ratio"` // 内存碎片率
	MaxMemory          uint64  `json:"max_memory"`          // 最大内存 (bytes)
	UsagePercent       float64 `json:"usage_percent"`       // 内存使用率
}

// RedisConnections Redis 连接指标
type RedisConnections struct {
	Connected        int   `json:"connected"`         // 已连接客户端数
	Blocked          int   `json:"blocked"`           // 阻塞客户端数
	MaxClients       int   `json:"max_clients"`       // 最大客户端数
	TotalConnections int64 `json:"total_connections"` // 历史总连接数
	RejectedConns    int64 `json:"rejected_conns"`    // 拒绝连接数
}

// RedisOperations Redis 操作指标
type RedisOperations struct {
	OpsPerSec     int64   `json:"ops_per_sec"`    // 每秒操作数
	TotalCommands int64   `json:"total_commands"` // 命令总数
	Hits          int64   `json:"hits"`           // 命中次数
	Misses        int64   `json:"misses"`         // 未命中次数
	HitRate       float64 `json:"hit_rate"`       // 命中率
	ExpiredKeys   int64   `json:"expired_keys"`   // 过期键数
	EvictedKeys   int64   `json:"evicted_keys"`   // 驱逐键数
}

// RedisKeyspace Redis 键空间指标
type RedisKeyspace struct {
	TotalKeys   int64 `json:"total_keys"`   // 键总数
	ExpiresKeys int64 `json:"expires_keys"` // 设置过期的键数
	AvgTTL      int64 `json:"avg_ttl"`      // 平均 TTL (ms)
}

// RedisPersistence Redis 持久化指标
type RedisPersistence struct {
	RDBLastSaveTime  int64 `json:"rdb_last_save_time"`  // 上次 RDB 保存时间
	RDBChangesSince  int64 `json:"rdb_changes_since"`   // 上次保存后的变更数
	AOFEnabled       bool  `json:"aof_enabled"`         // AOF 是否启用
	AOFRewriteInProg bool  `json:"aof_rewrite_in_prog"` // AOF 重写是否进行中
	AOFCurrentSize   int64 `json:"aof_current_size"`    // AOF 当前大小
}

// RedisReplication Redis 复制指标
type RedisReplication struct {
	Role             string `json:"role"`               // 角色 (master/slave)
	ConnectedSlaves  int    `json:"connected_slaves"`   // 已连接从节点数
	MasterLinkStatus string `json:"master_link_status"` // 主节点连接状态
	MasterLastIO     int64  `json:"master_last_io"`     // 上次与主节点通信时间
}

// ========== 历史数据查询 ==========

// PerformanceHistoryRequest 历史数据查询请求
type PerformanceHistoryRequest struct {
	Module   string `form:"module" binding:"required,oneof=system pgsql redis"` // 模块
	Start    int64  `form:"start" binding:"required"`                           // 开始时间戳
	End      int64  `form:"end" binding:"required"`                             // 结束时间戳
	Interval string `form:"interval" binding:"required,oneof=1m 5m 1h 1d"`      // 数据间隔
}

// PerformanceHistoryResponse 历史数据查询响应
type PerformanceHistoryResponse struct {
	Module    string        `json:"module"`     // 模块
	StartTime int64         `json:"start_time"` // 开始时间
	EndTime   int64         `json:"end_time"`   // 结束时间
	Interval  string        `json:"interval"`   // 数据间隔
	Data      []interface{} `json:"data"`       // 历史数据
}
