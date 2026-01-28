package controller

import (
	"github.com/gin-gonic/gin"

	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
)

// BaseController 基础控制器，提供公共方法
type BaseController struct{}

// GetUserID 从 Cookie 获取用户ID，未登录返回错误
func (b *BaseController) GetUserID(c *gin.Context) (int, error) {
	return satoken.GetUserIDFromCookie(c)
}

// GetUserIDOrZero 从 Cookie 获取用户ID，未登录返回0（游客模式）
func (b *BaseController) GetUserIDOrZero(c *gin.Context) int {
	return satoken.GetUserIDFromCookieOrZero(c)
}
