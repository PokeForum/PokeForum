package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
{
	"code": 200, 		// Error code in the program | 程序中的错误码
	"msg": "xxx", 		// Prompt message | 提示信息
	"data": {}			// Data | 数据
}
*/

type Data struct {
	Code ResCode `json:"code"`
	Msg  any     `json:"msg"`
	Data any     `json:"data"`
}

// ResError Return error information | 返回错误信息
func ResError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK,
		&Data{
			Code: code,
			Msg:  code.Msg(),
			Data: nil,
		})
}

// ResErrorWithMsg Custom error return | 自定义错误返回
func ResErrorWithMsg(c *gin.Context, code ResCode, msg any, data ...any) {
	c.JSON(http.StatusOK,
		&Data{
			Code: code,
			Msg:  msg,
			Data: data,
		})
}

// ResSuccess Return success information | 返回成功信息
func ResSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK,
		&Data{
			Code: CodeSuccess,
			Msg:  CodeSuccess.Msg(),
			Data: data,
		})
}
