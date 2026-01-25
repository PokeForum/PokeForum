package schema

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
	GoVersion    string `json:"go_version"`    // Go version | Go版本
	NumGoroutine int    `json:"num_goroutine"` // Number of goroutines | 协程数量
	NumCPU       int    `json:"num_cpu"`       // Number of CPU cores | CPU核心数
	MemAlloc     string `json:"mem_alloc"`     // Memory allocation | 内存分配
	MemSys       string `json:"mem_sys"`       // System memory | 系统内存
}
