package time_tools

import (
	"time"
)

// DateTimeFormat Standard time format for API return | API返回的标准时间格式
const DateTimeFormat = "2006-01-02 15:04:05"

// FormatDateTime Format time.Time to API standard time format | 将time.Time格式化为API标准时间格式
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// CalculateRemainingTime Calculate remaining time, return standard time format string | 计算剩余时间，返回标准时间格式字符串
// seconds: remaining time in seconds | 参数 seconds: 剩余时间的秒数
// Returns: standard time format string (e.g., 2025-10-29 12:23:31) | 返回值: 标准时间格式字符串 (例如: 2025-10-29 12:23:31)
func CalculateRemainingTime(seconds int64) string {
	// Get current timestamp (seconds) | 获取当前时间戳（秒）
	currentTimestamp := time.Now().Unix()

	// Calculate remaining timestamp | 计算剩余时间戳
	remainingTimestamp := currentTimestamp + seconds

	// Convert timestamp to time.Time object | 将时间戳转换为time.Time对象
	remainingTime := time.Unix(remainingTimestamp, 0)

	// Convert to standard time format string and return | 转换为标准时间格式字符串返回
	return remainingTime.Format("2006-01-02 15:04:05")
}
