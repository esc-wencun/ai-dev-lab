// Package router 路由注册测试：API 文档端点兼容性（对位 Java springdoc 路径契约）。
package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSwaggerDocEndpoints 前端 系统工具→接口文档 iframe 加载链路：
// /swagger-ui/index.html（页面）、/v3/api-docs（OpenAPI schema）、/swagger-ui.html（redirect）。
func TestSwaggerDocEndpoints(t *testing.T) {
	r := New(Deps{})

	// swagger-ui 页面可加载（iframe src）
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger-ui/index.html", nil))
	if w.Code != 200 {
		t.Errorf("/swagger-ui/index.html status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "swagger-ui") {
		t.Errorf("页面内容异常（应含 swagger-ui 资源引用）: %d 字节", w.Body.Len())
	}

	// OpenAPI schema
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v3/api-docs", nil))
	if w.Code != 200 {
		t.Fatalf("/v3/api-docs status = %d", w.Code)
	}
	var schema map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &schema); err != nil {
		t.Fatalf("/v3/api-docs 非合法 JSON: %v", err)
	}
	if schema["openapi"] == nil && schema["swagger"] == nil {
		t.Errorf("缺少 openapi/swagger 版本字段: %v", schema)
	}

	// 旧路径 redirect（对位 springdoc swagger-ui.path=/swagger-ui.html）
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger-ui.html", nil))
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("/swagger-ui.html status = %d, want 301", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/swagger-ui/index.html" {
		t.Errorf("Location = %q", loc)
	}
}

// TestSwaggerWhitelistFree 无 token 访问文档端点不被拦（对位 permitAll；
// 本测试直接注册路由验证白名单前缀逻辑在 Auth 中间件外的行为基线）。
func TestSwaggerWhitelistFree(t *testing.T) {
	w := httptest.NewRecorder()
	New(Deps{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger-ui/swagger-ui-bundle.js", nil))
	if w.Code == 401 {
		t.Error("swagger 静态资源不应要求认证")
	}
}
