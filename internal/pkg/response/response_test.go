package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResErrorWithMsgUsesBusinessCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ResErrorWithMsg(c, CodeInvalidParam, "请求参数错误")

	var data Data
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if data.Code != CodeInvalidParam {
		t.Fatalf("业务码应为 %d，实际为 %d", CodeInvalidParam, data.Code)
	}
	if data.ErrCode != resCodeToErrCode[CodeInvalidParam] {
		t.Fatalf("错误码应为 %s，实际为 %s", resCodeToErrCode[CodeInvalidParam], data.ErrCode)
	}
}
