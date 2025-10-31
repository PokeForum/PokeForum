package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// RandomString 生成指定长度的随机字符串
// 包含字符：0-9, a-z, A-Z
// length: 要生成的字符串长度
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// 定义字符集：数字、小写字母、大写字母
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetSize := big.NewInt(int64(len(charset)))

	// 创建字节切片存储结果
	result := make([]byte, length)

	// 生成随机字符
	for i := 0; i < length; i++ {
		// 生成随机索引
		randomIndex, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			// 如果随机数生成失败，使用时间作为备选方案
			// 这种情况极少发生，但为了代码健壮性需要处理
			panic("随机数生成失败: " + err.Error())
		}
		// 将对应字符添加到结果中
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result)
}

// ExtractEmailDomain 从邮箱地址中提取域名
// 例如: "user@gmail.com" -> "gmail.com"
// 如果邮箱格式无效，返回空字符串
func ExtractEmailDomain(email string) string {
	// 找到最后一个@符号的位置
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 || atIndex == len(email)-1 {
		return "" // 没有找到@符号或者@符号是最后一个字符
	}
	
	// 提取@符号后面的域名部分
	domain := email[atIndex+1:]
	
	// 去除前后空格
	domain = strings.TrimSpace(domain)
	
	// 检查域名是否包含点号
	if !strings.Contains(domain, ".") {
		return "" // 域名格式无效
	}
	
	return domain
}
