package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/pkg/response"
)

// newRateTestEnv miniredis + RedisCache（Lua 脚本在 miniredis 上真实执行）。
func newRateTestEnv(t *testing.T) (*cache.RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rc := cache.NewRedisCache(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	return rc, mr
}

// simpleHandler 返回 200 信封的兜底 handler。
func simpleHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Ok(c).JSON()
	}
}

// TestPreventRepeatSubmitJSON 同参数 POST：第二次 500 信封"不允许重复提交"。
func TestPreventRepeatSubmitJSON(t *testing.T) {
	rc, _ := newRateTestEnv(t)

	r := gin.New()
	r.POST("/system/user", PreventRepeatSubmit(rc, RepeatSubmitOptions{}), simpleHandler())

	post := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/system/user", strings.NewReader(`{"userName":"u1"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("token", "tok-1")
		r.ServeHTTP(w, req)
		return w
	}

	if w := post(); w.Code != 200 || !strings.Contains(w.Body.String(), `"code":200`) {
		t.Fatalf("首次请求应放行: %s", w.Body.String())
	}
	w := post()
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"code":500`) {
		t.Fatalf("重复请求应 code=500: %s", w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["msg"] != "不允许重复提交，请稍候再试" {
		t.Errorf("msg = %v, want 默认文案", body["msg"])
	}
}

// TestPreventRepeatSubmitDifferentParams 不同参数放行、间隔过后放行、不同 token 放行。
func TestPreventRepeatSubmitDifferentParams(t *testing.T) {
	rc, mr := newRateTestEnv(t)

	r := gin.New()
	r.POST("/system/user", PreventRepeatSubmit(rc, RepeatSubmitOptions{}), simpleHandler())

	post := func(body, token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/system/user", strings.NewReader(body))
		req.Header.Set("token", token)
		r.ServeHTTP(w, req)
		return w
	}

	if w := post(`{"userName":"u1"}`, "tok-1"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Fatalf("首次应放行: %s", w.Body.String())
	}
	// 不同参数：放行
	if w := post(`{"userName":"u2"}`, "tok-1"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("不同参数应放行: %s", w.Body.String())
	}
	// 不同 token：放行
	if w := post(`{"userName":"u1"}`, "tok-2"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("不同 token 应放行: %s", w.Body.String())
	}
	// 时间窗口过后：同参数也放行（快进 miniredis 时钟）
	mr.FastForward(6 * time.Second)
	if w := post(`{"userName":"u1"}`, "tok-1"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("间隔后同参数应放行: %s", w.Body.String())
	}
}

// TestPreventRepeatSubmitQueryParams GET query 参数同样判重。
func TestPreventRepeatSubmitQueryParams(t *testing.T) {
	rc, _ := newRateTestEnv(t)

	r := gin.New()
	r.GET("/system/user/list", PreventRepeatSubmit(rc, RepeatSubmitOptions{}), simpleHandler())

	get := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/system/user/list?"+query, nil))
		return w
	}
	if w := get("pageNum=1&pageSize=10"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Fatalf("首次 GET 应放行: %s", w.Body.String())
	}
	if w := get("pageNum=1&pageSize=10"); !strings.Contains(w.Body.String(), `"code":500`) {
		t.Errorf("同 query 重复应拦截: %s", w.Body.String())
	}
	if w := get("pageNum=2&pageSize=10"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("不同 query 应放行: %s", w.Body.String())
	}
}

// TestRateLimiterLuaCount Lua 计数：窗口内超过 count 拦截，超限文案正确。
func TestRateLimiterLuaCount(t *testing.T) {
	rc, _ := newRateTestEnv(t)

	r := gin.New()
	r.GET("/captchaImage", RateLimiter(rc, "TestRateLimiterLuaCount", RateLimiterOptions{Time: 60, Count: 3}), simpleHandler())

	var last *httptest.ResponseRecorder
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/captchaImage", nil))
		last = w
		if i < 3 {
			if !strings.Contains(w.Body.String(), `"code":200`) {
				t.Fatalf("第 %d 次应放行: %s", i+1, w.Body.String())
			}
		} else {
			if !strings.Contains(w.Body.String(), `"code":500`) {
				t.Errorf("第 %d 次应拦截: %s", i+1, w.Body.String())
			}
		}
	}
	var body map[string]any
	_ = json.Unmarshal(last.Body.Bytes(), &body)
	if body["msg"] != "访问过于频繁，请稍候再试" {
		t.Errorf("msg = %v", body["msg"])
	}
}

// TestRateLimiterIPDimension IP 维度：不同 IP 互不影响。
func TestRateLimiterIPDimension(t *testing.T) {
	rc, _ := newRateTestEnv(t)

	r := gin.New()
	r.GET("/sms", RateLimiter(rc, "TestRateLimiterIP", RateLimiterOptions{Time: 60, Count: 1, LimitType: LimitIP}), simpleHandler())

	hit := func(ip string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/sms", nil)
		req.Header.Set("X-Real-IP", ip) // 配合 gin TrustedPlatform 不可用时 ClientIP 回退链
		req.RemoteAddr = ip + ":1234"
		r.ServeHTTP(w, req)
		return w
	}
	if w := hit("1.2.3.4"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("IP1 首次应放行: %s", w.Body.String())
	}
	if w := hit("1.2.3.4"); !strings.Contains(w.Body.String(), `"code":500`) {
		t.Errorf("IP1 第二次应拦截: %s", w.Body.String())
	}
	if w := hit("5.6.7.8"); !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("IP2 应独立计数放行: %s", w.Body.String())
	}
}

// TestRateLimiterWindowReset 窗口过后计数重置（expire 生效）。
func TestRateLimiterWindowReset(t *testing.T) {
	rc, mr := newRateTestEnv(t)

	r := gin.New()
	r.GET("/x", RateLimiter(rc, "TestRateLimiterWindowReset", RateLimiterOptions{Time: 2, Count: 1}), simpleHandler())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !strings.Contains(w.Body.String(), `"code":200`) {
		t.Fatalf("首次应放行: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !strings.Contains(w.Body.String(), `"code":500`) {
		t.Fatalf("窗口内第二次应拦截: %s", w.Body.String())
	}
	mr.FastForward(3 * time.Second)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !strings.Contains(w.Body.String(), `"code":200`) {
		t.Errorf("窗口过后应放行: %s", w.Body.String())
	}
}
