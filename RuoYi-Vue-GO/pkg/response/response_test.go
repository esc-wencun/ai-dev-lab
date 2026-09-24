package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// body 解析响应体为 map（信封顶层字段校验用）
func body(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v，body=%s", err, w.Body.String())
	}
	return m
}

func TestOk(t *testing.T) {
	c, w := newTestContext(t)
	Ok(c).JSON()
	if w.Code != 200 {
		t.Fatalf("HTTP 状态码应为 200，实际 %d", w.Code)
	}
	m := body(t, w)
	if m["code"] != float64(CodeSuccess) {
		t.Errorf("code 应为 %d，实际 %v", CodeSuccess, m["code"])
	}
	if m["msg"] != "操作成功" {
		t.Errorf("msg 应为 操作成功，实际 %v", m["msg"])
	}
	if _, ok := m["data"]; ok {
		t.Error("未设置 Data 时不应输出 data 字段")
	}
}

func TestOkData(t *testing.T) {
	c, w := newTestContext(t)
	OkData(c, gin.H{"id": 1}).JSON()
	m := body(t, w)
	if m["code"] != float64(CodeSuccess) {
		t.Errorf("code 应为 %d", CodeSuccess)
	}
	data, ok := m["data"].(map[string]any)
	if !ok || data["id"] != float64(1) {
		t.Errorf("data 字段输出不符: %v", m["data"])
	}
}

func TestError(t *testing.T) {
	c, w := newTestContext(t)
	Error(c, "用户不存在/密码错误").JSON()
	m := body(t, w)
	if m["code"] != float64(CodeError) {
		t.Errorf("code 应为 %d", CodeError)
	}
	if m["msg"] != "用户不存在/密码错误" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

func TestErrorDefaultMsg(t *testing.T) {
	c, w := newTestContext(t)
	Error(c, "").JSON()
	m := body(t, w)
	if m["msg"] != "操作失败" {
		t.Errorf("空文案应输出默认 操作失败，实际 %v", m["msg"])
	}
}

func TestWarn(t *testing.T) {
	c, w := newTestContext(t)
	Warn(c, "验证码已失效").JSON()
	m := body(t, w)
	if m["code"] != float64(CodeWarn) {
		t.Errorf("code 应为 %d", CodeWarn)
	}
	if m["msg"] != "验证码已失效" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

func TestWarnDefaultMsg(t *testing.T) {
	c, w := newTestContext(t)
	Warn(c, "").JSON()
	m := body(t, w)
	if m["msg"] != "操作警告" {
		t.Errorf("空文案应输出默认 操作警告，实际 %v", m["msg"])
	}
}

func TestForbidden(t *testing.T) {
	c, w := newTestContext(t)
	Forbidden(c).JSON()
	m := body(t, w)
	if m["code"] != float64(CodeForbidden) {
		t.Errorf("code 应为 %d", CodeForbidden)
	}
	if m["msg"] != "没有权限，请联系管理员授权" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

func TestUnauthorized(t *testing.T) {
	c, w := newTestContext(t)
	Unauthorized(c).JSON()
	m := body(t, w)
	if m["code"] != float64(CodeUnauthorized) {
		t.Errorf("code 应为 %d", CodeUnauthorized)
	}
	if m["msg"] != "请求访问：认证失败，无法访问系统资源" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

func TestMsgOverride(t *testing.T) {
	c, w := newTestContext(t)
	Ok(c).Msg("登录成功").JSON()
	m := body(t, w)
	if m["msg"] != "登录成功" {
		t.Errorf("Msg 应覆盖默认文案，实际 %v", m["msg"])
	}
}

// TestPutFreeFields 自由字段置于顶层（对位 AjaxResult.put：token/uuid/img/rows/total）
func TestPutFreeFields(t *testing.T) {
	c, w := newTestContext(t)
	Ok(c).Put("token", "abc123").Put("rows", []int{1, 2}).Put("total", 2).JSON()
	m := body(t, w)
	if m["token"] != "abc123" {
		t.Errorf("顶层应含 token 字段: %v", m)
	}
	rows, ok := m["rows"].([]any)
	if !ok || len(rows) != 2 {
		t.Errorf("顶层应含 rows 字段: %v", m["rows"])
	}
	if m["total"] != float64(2) {
		t.Errorf("顶层应含 total 字段: %v", m["total"])
	}
	if m["code"] != float64(CodeSuccess) {
		t.Errorf("自由字段不应挤掉 code: %v", m["code"])
	}
}

// TestPutOverride 同 key 后写覆盖先写
func TestPutOverride(t *testing.T) {
	c, w := newTestContext(t)
	Ok(c).Put("captchaEnabled", true).Put("captchaEnabled", false).JSON()
	m := body(t, w)
	if m["captchaEnabled"] != false {
		t.Errorf("重复 Put 应后写覆盖先写，实际 %v", m["captchaEnabled"])
	}
}

// TestAllHTTP200 契约：所有响应（含 401/403/500）HTTP 状态码恒为 200
func TestAllHTTP200(t *testing.T) {
	cases := map[string]func(c *gin.Context){
		"ok":    func(c *gin.Context) { Ok(c).JSON() },
		"error": func(c *gin.Context) { Error(c, "x").JSON() },
		"warn":  func(c *gin.Context) { Warn(c, "x").JSON() },
		"403":   func(c *gin.Context) { Forbidden(c).JSON() },
		"401":   func(c *gin.Context) { Unauthorized(c).JSON() },
	}
	for name, call := range cases {
		c, w := newTestContext(t)
		call(c)
		if w.Code != 200 {
			t.Errorf("%s: HTTP 状态码恒为 200，实际 %d", name, w.Code)
		}
	}
}
