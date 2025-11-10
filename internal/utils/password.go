package utils

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行哈希加密
func HashPassword(password string) (string, error) {
	// 使用 bcrypt 的 DefaultCost 生成哈希值
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash 验证密码是否匹配哈希值
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GeneratePasswordSalt 生成密码盐
func GeneratePasswordSalt() string {
	// 生成8位随机字符串作为盐
	return RandomString(8)
}

// CombinePasswordWithSalt 将密码和盐进行拼接，增加密码强度
// 采用 salt + password + salt 的方式，提高密码安全性
// 参数：password 原始密码，salt 盐值
// 返回：拼接后的密码字符串
func CombinePasswordWithSalt(password, salt string) string {
	// 使用 salt + password + salt 的方式进行拼接
	// 这样可以防止彩虹表攻击，增加破解难度
	return salt + password + salt
}

// ValidateStrongPassword 验证强密码规则
// 必须满足：至少8位，包含大写字母、小写字母、数字、标点符号
func ValidateStrongPassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于8位")
	}

	// 检查是否包含大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}

	// 检查是否包含小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}

	// 检查是否包含数字
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	// 检查是否包含标点符号
	hasPunctuation := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)
	if !hasPunctuation {
		return fmt.Errorf("密码必须包含至少一个标点符号")
	}

	return nil
}

// GetPasswordStrengthTips 获取密码强度提示
func GetPasswordStrengthTips() string {
	return "密码必须满足以下要求：\n" +
		"• 至少8位字符\n" +
		"• 包含至少一个大写字母 (A-Z)\n" +
		"• 包含至少一个小写字母 (a-z)\n" +
		"• 包含至少一个数字 (0-9)\n" +
		"• 包含至少一个标点符号 (!@#$%^&*等)"
}
