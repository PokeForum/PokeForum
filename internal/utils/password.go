package utils

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword encrypts the password using hash algorithm | 对密码进行哈希加密
func HashPassword(password string) (string, error) {
	// Generate hash value using bcrypt's DefaultCost | 使用 bcrypt 的 DefaultCost 生成哈希值
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash verifies if the password matches the hash value | 验证密码是否匹配哈希值
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidateStrongPassword validates strong password rules | 验证强密码规则
// Must meet: at least 8 characters, contains uppercase letters, lowercase letters, numbers, and punctuation | 必须满足：至少8位，包含大写字母、小写字母、数字、标点符号
func ValidateStrongPassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于8位")
	}

	// Check if contains uppercase letters | 检查是否包含大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}

	// Check if contains lowercase letters | 检查是否包含小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}

	// Check if contains numbers | 检查是否包含数字
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	// Check if contains punctuation | 检查是否包含标点符号
	hasPunctuation := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)
	if !hasPunctuation {
		return fmt.Errorf("密码必须包含至少一个标点符号")
	}

	return nil
}

// GetPasswordStrengthTips returns password strength tips | 获取密码强度提示
func GetPasswordStrengthTips() string {
	return "密码必须满足以下要求：\n" +
		"• 至少8位字符\n" +
		"• 包含至少一个大写字母 (A-Z)\n" +
		"• 包含至少一个小写字母 (a-z)\n" +
		"• 包含至少一个数字 (0-9)\n" +
		"• 包含至少一个标点符号 (!@#$%^&*等)"
}
