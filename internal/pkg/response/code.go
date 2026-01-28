package response

type ResCode int64

const (
	CodeSuccess ResCode = 20000

	CodeInvalidParam = 40000
	CodeUnauthorized = 40100 // Unauthorized | 未授权
	CodeForbidden    = 40300 // Forbidden | 禁止访问
	CodeNotFound     = 40400 // Not found | 资源不存在
	CodeNoPermission = 40101 // No read permission | 无阅读权限

	CodeGenericError    = 50000
	CodeServerBusy      = 50001
	CodeTooManyRequests = 50002
	CodeNeedLogin       = 50003
	CodeTimeout         = 50004 // Request timeout | 请求超时
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess: "Success",

	CodeInvalidParam: "Invalid request parameters | 请求参数错误",
	CodeUnauthorized: "Unauthorized | 未授权",
	CodeForbidden:    "Forbidden | 禁止访问",
	CodeNotFound:     "Not found | 资源不存在",
	CodeNoPermission: "No read permission | 无阅读权限",

	CodeGenericError:    "Error",
	CodeServerBusy:      "The system is busy, please try again later | 系统繁忙，请稍候再试",
	CodeTooManyRequests: "Too many requests, please try again later | 请求过于频繁，请稍后再试",
	CodeNeedLogin:       "Not logged in | 未登录",
	CodeTimeout:         "Request timeout | 请求超时",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
