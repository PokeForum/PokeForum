package schema

// PerformanceSSERequest SSE client request parameters | SSE 客户端请求参数
type PerformanceSSERequest struct {
	Modules  string `form:"modules"`  // Monitoring modules, comma-separated: system,pgsql,redis (default all) | 监控模块，逗号分隔: system,pgsql,redis（默认全部）
	Interval int    `form:"interval"` // Push interval in seconds (default 3 seconds, min 1 second, max 60 seconds) | 推送间隔秒数（默认3秒，最小1秒，最大60秒）
}

// PerformanceWSResponse WebSocket server push message | WebSocket 服务端推送消息
type PerformanceWSResponse struct {
	Timestamp int64              `json:"timestamp"`        // Timestamp | 时间戳
	System    *SystemMetrics     `json:"system,omitempty"` // System metrics | 系统指标
	PgSQL     *PostgreSQLMetrics `json:"pgsql,omitempty"`  // PostgreSQL metrics | PostgreSQL 指标
	Redis     *RedisMetrics      `json:"redis,omitempty"`  // Redis metrics | Redis 指标
}

// ========== System monitoring metrics | 系统监控指标 ==========

// SystemMetrics System monitoring metrics | 系统监控指标
type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`     // CPU metrics | CPU 指标
	Memory  MemoryMetrics  `json:"memory"`  // Memory metrics | 内存指标
	Disk    DiskMetrics    `json:"disk"`    // Disk metrics | 磁盘指标
	Network NetworkMetrics `json:"network"` // Network metrics | 网络指标
	Load    LoadMetrics    `json:"load"`    // Load metrics | 负载指标
}

// CPUMetrics CPU metrics | CPU 指标
type CPUMetrics struct {
	UsagePercent  float64 `json:"usage_percent"`  // Total CPU usage | CPU 总使用率
	UserPercent   float64 `json:"user_percent"`   // User mode CPU | 用户态 CPU
	SystemPercent float64 `json:"system_percent"` // Kernel mode CPU | 内核态 CPU
	IdlePercent   float64 `json:"idle_percent"`   // Idle CPU | 空闲 CPU
	Cores         int     `json:"cores"`          // CPU cores | CPU 核心数
}

// MemoryMetrics Memory metrics | 内存指标
type MemoryMetrics struct {
	Total        uint64  `json:"total"`         // Total memory (bytes) | 总内存 (bytes)
	Used         uint64  `json:"used"`          // Used memory (bytes) | 已用内存 (bytes)
	Available    uint64  `json:"available"`     // Available memory (bytes) | 可用内存 (bytes)
	UsagePercent float64 `json:"usage_percent"` // Memory usage percentage | 内存使用率
	SwapTotal    uint64  `json:"swap_total"`    // Swap total (bytes) | 交换分区总量 (bytes)
	SwapUsed     uint64  `json:"swap_used"`     // Swap used (bytes) | 交换分区已用 (bytes)
	SwapPercent  float64 `json:"swap_percent"`  // Swap usage percentage | 交换分区使用率
}

// DiskMetrics Disk metrics | 磁盘指标
type DiskMetrics struct {
	Total        uint64  `json:"total"`         // Disk total (bytes) | 磁盘总量 (bytes)
	Used         uint64  `json:"used"`          // Used space (bytes) | 已用空间 (bytes)
	Free         uint64  `json:"free"`          // Free space (bytes) | 可用空间 (bytes)
	UsagePercent float64 `json:"usage_percent"` // Usage percentage | 使用率
	ReadBytes    uint64  `json:"read_bytes"`    // Read bytes | 读取字节数
	WriteBytes   uint64  `json:"write_bytes"`   // Write bytes | 写入字节数
	ReadCount    uint64  `json:"read_count"`    // Read count | 读取次数
	WriteCount   uint64  `json:"write_count"`   // Write count | 写入次数
}

// NetworkMetrics Network metrics | 网络指标
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`   // Bytes sent | 发送字节数
	BytesRecv   uint64 `json:"bytes_recv"`   // Bytes received | 接收字节数
	PacketsSent uint64 `json:"packets_sent"` // Packets sent | 发送包数
	PacketsRecv uint64 `json:"packets_recv"` // Packets received | 接收包数
	ErrIn       uint64 `json:"err_in"`       // Receive errors | 接收错误数
	ErrOut      uint64 `json:"err_out"`      // Send errors | 发送错误数
	DropIn      uint64 `json:"drop_in"`      // Receive packet drops | 接收丢包数
	DropOut     uint64 `json:"drop_out"`     // Send packet drops | 发送丢包数
	Connections int    `json:"connections"`  // Connection count | 连接数
}

// LoadMetrics Load metrics | 负载指标
type LoadMetrics struct {
	Load1  float64 `json:"load1"`  // 1 minute load | 1分钟负载
	Load5  float64 `json:"load5"`  // 5 minute load | 5分钟负载
	Load15 float64 `json:"load15"` // 15 minute load | 15分钟负载
}

// ========== PostgreSQL monitoring metrics | PostgreSQL 监控指标 ==========

// PostgreSQLMetrics PostgreSQL monitoring metrics | PostgreSQL 监控指标
type PostgreSQLMetrics struct {
	Connections PgConnections `json:"connections"` // Connection metrics | 连接指标
	Transaction PgTransaction `json:"transaction"` // Transaction metrics | 事务指标
	Cache       PgCache       `json:"cache"`       // Cache metrics | 缓存指标
	Database    PgDatabase    `json:"database"`    // Database metrics | 数据库指标
	Locks       PgLocks       `json:"locks"`       // Lock metrics | 锁指标
	Replication PgReplication `json:"replication"` // Replication metrics | 复制指标
}

// PgConnections PostgreSQL connection metrics | PostgreSQL 连接指标
type PgConnections struct {
	Active       int     `json:"active"`        // Active connections | 活跃连接数
	Idle         int     `json:"idle"`          // Idle connections | 空闲连接数
	IdleInTx     int     `json:"idle_in_tx"`    // Idle in transaction connections | 事务中空闲连接数
	Waiting      int     `json:"waiting"`       // Waiting connections | 等待连接数
	Total        int     `json:"total"`         // Total connections | 总连接数
	MaxConn      int     `json:"max_conn"`      // Max connections | 最大连接数
	UsagePercent float64 `json:"usage_percent"` // Connection usage percentage | 连接使用率
}

// PgTransaction PostgreSQL transaction metrics | PostgreSQL 事务指标
type PgTransaction struct {
	Committed  int64   `json:"committed"`   // Committed transactions | 已提交事务数
	RolledBack int64   `json:"rolled_back"` // Rolled back transactions | 已回滚事务数
	TPS        float64 `json:"tps"`         // Transactions per second | 每秒事务数
}

// PgCache PostgreSQL cache metrics | PostgreSQL 缓存指标
type PgCache struct {
	BlocksHit  int64   `json:"blocks_hit"`  // Cache hit blocks | 缓存命中块数
	BlocksRead int64   `json:"blocks_read"` // Disk read blocks | 磁盘读取块数
	HitRatio   float64 `json:"hit_ratio"`   // Cache hit ratio | 缓存命中率
	TempFiles  int64   `json:"temp_files"`  // Temporary files | 临时文件数
	TempBytes  int64   `json:"temp_bytes"`  // Temporary file size | 临时文件大小
}

// PgDatabase PostgreSQL database metrics | PostgreSQL 数据库指标
type PgDatabase struct {
	Size       int64 `json:"size"`        // Database size (bytes) | 数据库大小 (bytes)
	TableCount int   `json:"table_count"` // Table count | 表数量
	IndexSize  int64 `json:"index_size"`  // Index size (bytes) | 索引大小 (bytes)
	TableSize  int64 `json:"table_size"`  // Table size (bytes) | 表大小 (bytes)
}

// PgLocks PostgreSQL lock metrics | PostgreSQL 锁指标
type PgLocks struct {
	Total       int   `json:"total"`        // Total locks | 锁总数
	Waiting     int   `json:"waiting"`      // Waiting locks | 等待锁数
	AccessShare int   `json:"access_share"` // AccessShare locks | AccessShare 锁数
	RowShare    int   `json:"row_share"`    // RowShare locks | RowShare 锁数
	RowExcl     int   `json:"row_excl"`     // RowExclusive locks | RowExclusive 锁数
	Deadlocks   int64 `json:"deadlocks"`    // Deadlock count | 死锁数
}

// PgReplication PostgreSQL replication metrics | PostgreSQL 复制指标
type PgReplication struct {
	IsReplica    bool  `json:"is_replica"`    // Whether is replica | 是否为副本
	ReplicaCount int   `json:"replica_count"` // Replica count | 副本数量
	LagBytes     int64 `json:"lag_bytes"`     // Replication lag (bytes) | 复制延迟 (bytes)
}

// ========== Redis monitoring metrics | Redis 监控指标 ==========

// RedisMetrics Redis monitoring metrics | Redis 监控指标
type RedisMetrics struct {
	Memory      RedisMemory      `json:"memory"`      // Memory metrics | 内存指标
	Connections RedisConnections `json:"connections"` // Connection metrics | 连接指标
	Operations  RedisOperations  `json:"operations"`  // Operation metrics | 操作指标
	Keyspace    RedisKeyspace    `json:"keyspace"`    // Keyspace metrics | 键空间指标
	Persistence RedisPersistence `json:"persistence"` // Persistence metrics | 持久化指标
	Replication RedisReplication `json:"replication"` // Replication metrics | 复制指标
}

// RedisMemory Redis memory metrics | Redis 内存指标
type RedisMemory struct {
	Used               uint64  `json:"used"`                // Used memory (bytes) | 已用内存 (bytes)
	UsedPeak           uint64  `json:"used_peak"`           // Memory peak (bytes) | 内存峰值 (bytes)
	UsedRSS            uint64  `json:"used_rss"`            // RSS memory (bytes) | RSS 内存 (bytes)
	FragmentationRatio float64 `json:"fragmentation_ratio"` // Memory fragmentation ratio | 内存碎片率
	MaxMemory          uint64  `json:"max_memory"`          // Max memory (bytes) | 最大内存 (bytes)
	UsagePercent       float64 `json:"usage_percent"`       // Memory usage percentage | 内存使用率
}

// RedisConnections Redis connection metrics | Redis 连接指标
type RedisConnections struct {
	Connected        int   `json:"connected"`         // Connected clients | 已连接客户端数
	Blocked          int   `json:"blocked"`           // Blocked clients | 阻塞客户端数
	MaxClients       int   `json:"max_clients"`       // Max clients | 最大客户端数
	TotalConnections int64 `json:"total_connections"` // Total connections in history | 历史总连接数
	RejectedConns    int64 `json:"rejected_conns"`    // Rejected connections | 拒绝连接数
}

// RedisOperations Redis operation metrics | Redis 操作指标
type RedisOperations struct {
	OpsPerSec     int64   `json:"ops_per_sec"`    // Operations per second | 每秒操作数
	TotalCommands int64   `json:"total_commands"` // Total commands | 命令总数
	Hits          int64   `json:"hits"`           // Hit count | 命中次数
	Misses        int64   `json:"misses"`         // Miss count | 未命中次数
	HitRate       float64 `json:"hit_rate"`       // Hit rate | 命中率
	ExpiredKeys   int64   `json:"expired_keys"`   // Expired keys | 过期键数
	EvictedKeys   int64   `json:"evicted_keys"`   // Evicted keys | 驱逐键数
}

// RedisKeyspace Redis keyspace metrics | Redis 键空间指标
type RedisKeyspace struct {
	TotalKeys   int64 `json:"total_keys"`   // Total keys | 键总数
	ExpiresKeys int64 `json:"expires_keys"` // Keys with expiration set | 设置过期的键数
	AvgTTL      int64 `json:"avg_ttl"`      // Average TTL (ms) | 平均 TTL (ms)
}

// RedisPersistence Redis persistence metrics | Redis 持久化指标
type RedisPersistence struct {
	RDBLastSaveTime  int64 `json:"rdb_last_save_time"`  // Last RDB save time | 上次 RDB 保存时间
	RDBChangesSince  int64 `json:"rdb_changes_since"`   // Changes since last save | 上次保存后的变更数
	AOFEnabled       bool  `json:"aof_enabled"`         // Whether AOF is enabled | AOF 是否启用
	AOFRewriteInProg bool  `json:"aof_rewrite_in_prog"` // Whether AOF rewrite is in progress | AOF 重写是否进行中
	AOFCurrentSize   int64 `json:"aof_current_size"`    // AOF current size | AOF 当前大小
}

// RedisReplication Redis replication metrics | Redis 复制指标
type RedisReplication struct {
	Role             string `json:"role"`               // Role (master/slave) | 角色 (master/slave)
	ConnectedSlaves  int    `json:"connected_slaves"`   // Connected slaves | 已连接从节点数
	MasterLinkStatus string `json:"master_link_status"` // Master link status | 主节点连接状态
	MasterLastIO     int64  `json:"master_last_io"`     // Last communication with master | 上次与主节点通信时间
}

// ========== Historical data query | 历史数据查询 ==========

// PerformanceHistoryRequest Historical data query request | 历史数据查询请求
type PerformanceHistoryRequest struct {
	Module   string `form:"module" binding:"required,oneof=system pgsql redis"` // Module | 模块
	Start    int64  `form:"start" binding:"required"`                           // Start timestamp | 开始时间戳
	End      int64  `form:"end" binding:"required"`                             // End timestamp | 结束时间戳
	Interval string `form:"interval" binding:"required,oneof=1m 5m 1h 1d"`      // Data interval | 数据间隔
}

// PerformanceHistoryResponse Historical data query response | 历史数据查询响应
type PerformanceHistoryResponse struct {
	Module    string        `json:"module"`     // Module | 模块
	StartTime int64         `json:"start_time"` // Start time | 开始时间
	EndTime   int64         `json:"end_time"`   // End time | 结束时间
	Interval  string        `json:"interval"`   // Data interval | 数据间隔
	Data      []interface{} `json:"data"`       // Historical data | 历史数据
}
