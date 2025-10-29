package time_tools

import (
	"time"
)

// CalculateRemainingTime 计算剩余时间，返回标准时间格式字符串
// 参数 seconds: 剩余时间的秒数
// 返回值: 标准时间格式字符串 (例如: 2025-10-29 12:23:31)
func CalculateRemainingTime(seconds int64) string {
	// 获取当前时间戳（秒）
	currentTimestamp := time.Now().Unix()

	// 计算剩余时间戳
	remainingTimestamp := currentTimestamp + seconds

	// 将时间戳转换为time.Time对象
	remainingTime := time.Unix(remainingTimestamp, 0)

	// 转换为标准时间格式字符串返回
	return remainingTime.Format("2006-01-02 15:04:05")
}
