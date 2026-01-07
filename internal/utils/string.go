package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// RandomString generates a random string of specified length | 生成指定长度的随机字符串
// Contains characters: 0-9, a-z, A-Z | 包含字符：0-9, a-z, A-Z
// length: the length of the string to generate | length: 要生成的字符串长度
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// Define character set: numbers, lowercase letters, uppercase letters | 定义字符集：数字、小写字母、大写字母
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetSize := big.NewInt(int64(len(charset)))

	// Create byte slice to store result | 创建字节切片存储结果
	result := make([]byte, length)

	// Generate random characters | 生成随机字符
	for i := 0; i < length; i++ {
		// Generate random index | 生成随机索引
		randomIndex, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			// If random number generation fails, use time as fallback | 如果随机数生成失败，使用时间作为备选方案
			// This rarely happens, but needs to be handled for code robustness | 这种情况极少发生，但为了代码健壮性需要处理
			panic("随机数生成失败: " + err.Error())
		}
		// Add corresponding character to result | 将对应字符添加到结果中
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result)
}

// ExtractEmailDomain extracts domain from email address | 从邮箱地址中提取域名
// Example: "user@gmail.com" -> "gmail.com" | 例如: "user@gmail.com" -> "gmail.com"
// Returns empty string if email format is invalid | 如果邮箱格式无效，返回空字符串
func ExtractEmailDomain(email string) string {
	// Find the position of the last @ symbol | 找到最后一个@符号的位置
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 || atIndex == len(email)-1 {
		return "" // @ symbol not found or @ is the last character | 没有找到@符号或者@符号是最后一个字符
	}

	// Extract domain part after @ symbol | 提取@符号后面的域名部分
	domain := email[atIndex+1:]

	// Trim leading and trailing spaces | 去除前后空格
	domain = strings.TrimSpace(domain)

	// Check if domain contains dot | 检查域名是否包含点号
	if !strings.Contains(domain, ".") {
		return "" // Invalid domain format | 域名格式无效
	}

	return domain
}
