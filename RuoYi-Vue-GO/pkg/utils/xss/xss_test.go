package xss

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// post 构造 POST JSON 请求走中间件，返回 handler 收到的 body。
func post(t *testing.T, mw gin.HandlerFunc, path, body string) (int, string) {
	t.Helper()
	var got string
	r := gin.New()
	r.Use(mw)
	r.POST("/*any", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		got = string(b)
		c.Status(200)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w.Code, got
}

// TestCleanScript 标签剥离：<script> 等标签被剥、标签间文本保留（对齐 Python 版 _TAG_RE 语义）。
func TestCleanScript(t *testing.T) {
	if got := Clean(`<script>alert(1)</script>hello`); got != "alert(1)hello" {
		t.Errorf("Clean = %q, want %q", got, "alert(1)hello")
	}
	if got := Clean(`plain text`); got != "plain text" {
		t.Errorf("纯文本被改变: %q", got)
	}
	if got := Clean(`<img src=x onerror=alert(1)>a<b>b</b>`); got != "ab" {
		t.Errorf("img/b 标签未剥净: %q", got)
	}
}

// TestMiddlewareJSONBody JSON body 嵌套字段清洗（顶层+嵌套+数组）。
func TestMiddlewareJSONBody(t *testing.T) {
	_, got := post(t, Middleware(Options{}), "/system/user",
		`{"userName":"u1","remark":"<script>x</script>备注","profile":{"hobby":["<b>球</b>棋", 3]}}`)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("清洗后非 JSON: %q", got)
	}
	if m["userName"] != "u1" {
		t.Errorf("非标签字段被误改: %v", m["userName"])
	}
	if m["remark"] != "x备注" {
		t.Errorf("script 标签未剥离（文本应保留）: %v", m["remark"])
	}
	profile := m["profile"].(map[string]any)
	hobby := profile["hobby"].([]any)
	if hobby[0] != "球棋" {
		t.Errorf("数组内标签未清洗: %v", hobby[0])
	}
	if hobby[1] != float64(3) {
		t.Errorf("数字字段应保留: %v", hobby[1])
	}
}

// TestMiddlewareExcludes 排除名单路径不清洗（富文本保留）。
func TestMiddlewareExcludes(t *testing.T) {
	body := `{"noticeTitle":"t","noticeContent":"<p>富文本<b>内容</b></p>"}`
	_, got := post(t, Middleware(Options{}), "/system/notice", body)
	if !strings.Contains(got, "<p>") {
		t.Errorf("排除路径富文本被清洗: %q", got)
	}
}

// TestMiddlewareMethodScope GET/DELETE 不过滤；非 JSON Content-Type 不过滤。
func TestMiddlewareMethodScope(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(Options{}))
	var got string
	r.GET("/*any", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		got = string(b)
		c.Status(200)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system/user?x=1", strings.NewReader(`{"a":"<b>k</b>"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if got != `{"a":"<b>k</b>"}` {
		t.Errorf("GET 应不过滤: %q", got)
	}
}

// TestMiddlewareNonJSON 非 JSON body（如 multipart）原样放行。
func TestMiddlewareNonJSON(t *testing.T) {
	body := "-----------------------------boundary\r\nContent-Disposition: form-data\r\n\r\n<b>v</b>\r\n"
	_, got := post(t, Middleware(Options{}), "/system/user", body)
	if !strings.Contains(got, "<b>v</b>") {
		t.Errorf("非 JSON body 应原样放行: %q", got)
	}
}
