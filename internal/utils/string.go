package utils

import (
	"crypto/rand"
	"math/big"
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
