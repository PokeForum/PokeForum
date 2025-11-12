package middleware

import (
	"errors"
	"net/http"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Recovery recover掉项目可能出现的panic，并使用zap记录相关日志
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// 从context中获取链路ID，用于追踪请求
		traceID := tracing.GetTraceID(c.Request.Context())

		// 尝试转换为validator.ValidationErrors
		var errs validator.ValidationErrors
		if errors.As(&errs, &recovered) {
			c.String(http.StatusBadRequest, "参数校验失败")
			return
		}

		// 尝试转换为error
		if err, ok := recovered.(error); ok {
			// 在错误日志中包含链路ID
			configs.Log.Error(err.Error(), zap.String("trace_id", traceID))
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
