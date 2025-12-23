package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/response"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Recovery recover掉项目可能出现的panic，并使用zap记录相关日志
// 注意：不向客户端暴露内部错误详情，仅返回通用错误信息
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// 从context中获取链路ID，用于追踪请求
		traceID := tracing.GetTraceID(c.Request.Context())

		// 尝试转换为validator.ValidationErrors
		var errs validator.ValidationErrors
		if errors.As(&errs, &recovered) {
			response.ResError(c, response.CodeInvalidParam)
			return
		}

		// 记录详细错误信息到日志（包含堆栈），但不返回给客户端
		var errMsg string
		switch v := recovered.(type) {
		case error:
			errMsg = v.Error()
		case string:
			errMsg = v
		default:
			errMsg = fmt.Sprintf("%v", v)
		}

		// 记录完整的错误信息和堆栈，便于排查问题
		configs.Log.Error("服务器内部错误",
			zap.String("trace_id", traceID),
			zap.String("error", errMsg),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
			zap.String("stack", string(debug.Stack())),
		)

		// 返回通用错误信息，不暴露内部细节
		// 客户端可通过响应头中的 trace_id 联系管理员排查
		response.ResError(c, response.CodeServerBusy)
	})
}
