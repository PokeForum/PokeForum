package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PokeForum/PokeForum/internal/pkg/errcode"
)

/*
{
	"code": 200, 		// Error code in the program | 程序中的错误码
	"err_code": "SUCCESS", // Business error code | 业务错误码
	"msg": "xxx", 		// Prompt message | 提示信息
	"data": {}			// Data | 数据
}
*/

var resCodeToErrCode = map[ResCode]string{
	CodeSuccess:         string(errcode.Success),
	CodeInvalidParam:    string(errcode.InvalidParam),
	CodeUnauthorized:    string(errcode.NeedLogin),
	CodeForbidden:       string(errcode.NoPermission),
	CodeNotFound:        string(errcode.GenericError),
	CodeNoPermission:    string(errcode.NoPermission),
	CodeGenericError:    string(errcode.GenericError),
	CodeServerBusy:      string(errcode.ServerBusy),
	CodeTooManyRequests: string(errcode.TooManyRequests),
	CodeNeedLogin:       string(errcode.NeedLogin),
	CodeTimeout:         string(errcode.ServerBusy),
}

type Data struct {
	Code    ResCode `json:"code"`
	ErrCode string  `json:"err_code"`
	Msg     any     `json:"msg"`
	Data    any     `json:"data"`
}

// ResError Return error information | 返回错误信息
func ResError(c *gin.Context, code ResCode) {
	errCode, ok := resCodeToErrCode[code]
	if !ok {
		errCode = string(errcode.GenericError)
	}
	c.JSON(http.StatusOK,
		&Data{
			Code:    code,
			ErrCode: errCode,
			Msg:     code.Msg(),
			Data:    nil,
		})
}

// ResErrorWithMsg Custom error return | 自定义错误返回
func ResErrorWithMsg(c *gin.Context, code ResCode, msg any, data ...any) {
	errCode, ok := resCodeToErrCode[code]
	if !ok {
		errCode = string(errcode.GenericError)
	}
	c.JSON(http.StatusOK,
		&Data{
			Code:    code,
			ErrCode: errCode,
			Msg:     msg,
			Data:    data,
		})
}

// ResErrorWithErrCode Custom error return with specific error code | 自定义错误返回，指定错误码
func ResErrorWithErrCode(c *gin.Context, code ResCode, businessErrCode errcode.ErrCode) {
	c.JSON(http.StatusOK,
		&Data{
			Code:    code,
			ErrCode: string(businessErrCode),
			Msg:     businessErrCode.Msg(),
			Data:    nil,
		})
}

// ResErrorWithErrCodeAndMsg Custom error return with specific error code and message | 自定义错误返回，指定错误码和消息
func ResErrorWithErrCodeAndMsg(c *gin.Context, code ResCode, businessErrCode errcode.ErrCode, msg any) {
	c.JSON(http.StatusOK,
		&Data{
			Code:    code,
			ErrCode: string(businessErrCode),
			Msg:     msg,
			Data:    nil,
		})
}

// ResSuccess Return success information | 返回成功信息
func ResSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK,
		&Data{
			Code:    CodeSuccess,
			ErrCode: string(errcode.Success),
			Msg:     CodeSuccess.Msg(),
			Data:    data,
		})
}

// ResCodeToHTTPStatus 将业务码转换为 HTTP 状态码
func ResCodeToHTTPStatus(code ResCode) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidParam:
		return http.StatusBadRequest
	case CodeUnauthorized, CodeNeedLogin:
		return http.StatusUnauthorized
	case CodeForbidden, CodeNoPermission:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeGenericError, CodeServerBusy:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ResErrorWithHTTPStatus 返回错误信息，并指定 HTTP 状态码
func ResErrorWithHTTPStatus(c *gin.Context, httpStatus int, code ResCode, msg any) {
	errCode, ok := resCodeToErrCode[code]
	if !ok {
		errCode = string(errcode.GenericError)
	}
	c.JSON(httpStatus,
		&Data{
			Code:    code,
			ErrCode: errCode,
			Msg:     msg,
			Data:    nil,
		})
}
