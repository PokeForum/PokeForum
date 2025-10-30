package response

type ResCode int64

const (
	CodeSuccess ResCode = 20000

	CodeInvalidParam = 40000

	CodeGenericError    = 50000
	CodeServerBusy      = 50001
	CodeTooManyRequests = 50002
	CodeNeedLogin       = 50003
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess: "Success",

	CodeInvalidParam: "请求参数错误",

	CodeGenericError:    "Error",
	CodeServerBusy:      "系统繁忙，请稍候再试",
	CodeTooManyRequests: "请求过于频繁，请稍后再试",
	CodeNeedLogin:       "未登录",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
