package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	bizerrors "ruoyi-vue-go/internal/common/errors"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// parseBody 解析响应体信封
func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v，body=%s", err, w.Body.String())
	}
	return m
}

// TestWrapBusinessError handler 返回 BusinessError → 按其 Code/Msg 输出信封
func TestWrapBusinessError(t *testing.T) {
	r := newTestRouter()
	r.GET("/biz", Wrap(func(c *gin.Context) error {
		return bizerrors.NewWithCode(601, "验证码已失效")
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/biz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 恒 200，实际 %d", w.Code)
	}
	m := parseBody(t, w)
	if m["code"] != float64(601) {
		t.Errorf("code 应为 601，实际 %v", m["code"])
	}
	if m["msg"] != "验证码已失效" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

// TestWrapDefaultBusinessError BusinessError 默认 code=500
func TestWrapDefaultBusinessError(t *testing.T) {
	r := newTestRouter()
	r.GET("/biz", Wrap(func(c *gin.Context) error {
		return bizerrors.New("用户不存在/密码错误")
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/biz", nil)
	r.ServeHTTP(w, req)

	m := parseBody(t, w)
	if m["code"] != float64(500) {
		t.Errorf("默认 code 应为 500，实际 %v", m["code"])
	}
	if m["msg"] != "用户不存在/密码错误" {
		t.Errorf("msg 不符: %v", m["msg"])
	}
}

// TestWrapPlainError 泛化 error → 500 + 错误文本（对位 Java error(e.getMessage())）
func TestWrapPlainError(t *testing.T) {
	r := newTestRouter()
	r.GET("/plain", Wrap(func(c *gin.Context) error {
		return errors.New("数据库连接失败")
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/plain", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 恒 200，实际 %d", w.Code)
	}
	m := parseBody(t, w)
	if m["code"] != float64(500) {
		t.Errorf("code 应为 500，实际 %v", m["code"])
	}
	if m["msg"] != "数据库连接失败" {
		t.Errorf("msg 应为错误文本: %v", m["msg"])
	}
}

// TestWrapNoError handler 返回 nil → Wrap 不写响应
func TestWrapNoError(t *testing.T) {
	r := newTestRouter()
	r.GET("/ok", Wrap(func(c *gin.Context) error {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
		return nil
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	m := parseBody(t, w)
	if m["code"] != float64(200) {
		t.Errorf("handler 正常返回时信封不应被改写: %v", m)
	}
}

// TestRecoveryPanicError panic 值为 error → 500 信封 + 异常文本
func TestRecoveryPanicError(t *testing.T) {
	r := newTestRouter()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic(errors.New("空指针"))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 恒 200，实际 %d", w.Code)
	}
	m := parseBody(t, w)
	if m["code"] != float64(500) {
		t.Errorf("code 应为 500，实际 %v", m["code"])
	}
	if m["msg"] != "空指针" {
		t.Errorf("msg 应取 err.Error(): %v", m["msg"])
	}
}

// TestRecoveryPanicString panic 值为非 error → fmt.Sprint 文本
func TestRecoveryPanicString(t *testing.T) {
	r := newTestRouter()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic(fmt.Sprintf("非法状态 %d", 42))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	m := parseBody(t, w)
	if m["code"] != float64(500) {
		t.Errorf("code 应为 500，实际 %v", m["code"])
	}
	if m["msg"] != "非法状态 42" {
		t.Errorf("msg 应为 panic 文本: %v", m["msg"])
	}
}

// TestRecoveryPassThrough 不 panic 时请求正常通过
func TestRecoveryPassThrough(t *testing.T) {
	r := newTestRouter()
	r.Use(Recovery())
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("不 panic 时应正常响应，实际 %d", w.Code)
	}
}
