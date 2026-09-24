package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/security"
	"ruoyi-vue-go/pkg/response"
)

// testSecret 对齐 Java/Python 默认密钥（原文 UTF-8 字节直接作 HMAC key，见 token_service.go 头注）
const testSecret = "abcdefghijklmnopqrstuvwxyz"

func newTestEnv(t *testing.T) (*cache.RedisCache, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	rc := cache.NewRedisCache(client)
	ts := security.NewTokenService(rc, "Authorization", testSecret, 30)
	r := gin.New()
	r.Use(Auth(ts))
	return rc, r
}

// addProtected 注册受 system:user:list 权限保护的演示端点（端到端场景用）
func addProtected(r *gin.Engine) {
	r.GET("/system/user/list", RequirePerm("system:user:list"), Wrap(func(c *gin.Context) error {
		response.Ok(c).Msg("ok").JSON()
		return nil
	}))
}

func signToken(uuid string) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		constant.LoginUserKey: uuid,
		constant.JwtUsername:  "ry",
	})
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		panic(err)
	}
	return signed
}

func seedSession(t *testing.T, rc *cache.RedisCache, uuid string, lu *model.LoginUser) {
	t.Helper()
	lu.Token = uuid
	if err := rc.SetObject(context.Background(), cache.BuildKey(constant.LoginTokenKey, uuid), lu, 30*time.Minute); err != nil {
		t.Fatalf("seed session: %v", err)
	}
}

// makeLoginUser 构造会话；roleKeys 写入 User 原始 JSON 的 roles（对位 SysUser.roles）
func makeLoginUser(userId int64, permissions []string, roleKeys ...string) *model.LoginUser {
	rolesJSON := "null"
	if len(roleKeys) > 0 {
		quoted := make([]string, 0, len(roleKeys))
		for _, rk := range roleKeys {
			quoted = append(quoted, fmt.Sprintf(`{"roleKey":%q}`, rk))
		}
		rolesJSON = "[" + strings.Join(quoted, ",") + "]"
	}
	userRaw := fmt.Sprintf(`{"userName":"u%s","roles":%s}`, strconv.FormatInt(userId, 10), rolesJSON)
	return &model.LoginUser{UserId: userId, DeptId: 100, Permissions: permissions, User: []byte(userRaw)}
}

func do(r *gin.Engine, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// bodyCode 断言 HTTP 恒 200（信封契约）并返回 body 中的 code 字段
func bodyCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 状态码应为 200（信封契约），实际 %d", w.Code)
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析信封失败: %v, body=%s", err, w.Body.String())
	}
	return body.Code
}

func TestRequirePermEndToEnd(t *testing.T) {
	rc, r := newTestEnv(t)
	addProtected(r)

	// admin（userId=1）放行
	seedSession(t, rc, "uuid-admin", makeLoginUser(1, nil, "admin"))
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-admin"))); got != response.CodeSuccess {
		t.Errorf("admin 应放行 200，得 %d", got)
	}

	// ry（userId=2，common 角色无 system:user:list）→ 403
	seedSession(t, rc, "uuid-ry", makeLoginUser(2, []string{"system:user:query"}, "common"))
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-ry"))); got != response.CodeForbidden {
		t.Errorf("ry 应 403，得 %d", got)
	}

	// 无 token → 401
	if got := bodyCode(t, do(r, "/system/user/list", "")); got != response.CodeUnauthorized {
		t.Errorf("无 token 应 401，得 %d", got)
	}

	// 含 *:*:* 全权标识 → 放行
	seedSession(t, rc, "uuid-all", makeLoginUser(3, []string{constant.AllPermission}, "common"))
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-all"))); got != response.CodeSuccess {
		t.Errorf("*:*:* 应放行 200，得 %d", got)
	}

	// 精确匹配 → 放行
	seedSession(t, rc, "uuid-trim", makeLoginUser(4, []string{"system:user:list"}, "common"))
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-trim"))); got != response.CodeSuccess {
		t.Errorf("精确匹配应放行 200，得 %d", got)
	}
}

func TestRequireRole(t *testing.T) {
	rc, _ := newTestEnv(t)
	r := gin.New()
	r.Use(Auth(security.NewTokenService(rc, "Authorization", testSecret, 30)))
	r.GET("/admin/demo", RequireRole("admin"), Wrap(func(c *gin.Context) error {
		response.Ok(c).Msg("ok").JSON()
		return nil
	}))
	r.GET("/common/demo", RequireRole("common"), Wrap(func(c *gin.Context) error {
		response.Ok(c).Msg("ok").JSON()
		return nil
	}))

	seedSession(t, rc, "uuid-ry", makeLoginUser(2, nil, "common"))
	if got := bodyCode(t, do(r, "/common/demo", "Bearer "+signToken("uuid-ry"))); got != response.CodeSuccess {
		t.Errorf("common 角色访问 common 端点应放行，得 %d", got)
	}
	if got := bodyCode(t, do(r, "/admin/demo", "Bearer "+signToken("uuid-ry"))); got != response.CodeForbidden {
		t.Errorf("common 角色访问 admin 端点应 403，得 %d", got)
	}
	if got := bodyCode(t, do(r, "/admin/demo", "")); got != response.CodeUnauthorized {
		t.Errorf("无 token 访问角色端点应 401，得 %d", got)
	}

	// 超级管理员 roleKey=admin 放行任意角色校验
	seedSession(t, rc, "uuid-super", makeLoginUser(3, nil, "admin"))
	if got := bodyCode(t, do(r, "/admin/demo", "Bearer "+signToken("uuid-super"))); got != response.CodeSuccess {
		t.Errorf("roleKey=admin 应放行 admin 端点，得 %d", got)
	}
}

func TestAuthFailScenarios(t *testing.T) {
	_, r := newTestEnv(t)
	addProtected(r)

	// 非法签名 → 401
	badTok, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{constant.LoginUserKey: "uuid-x"}).SignedString([]byte("wrong-secret"))
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+badTok)); got != response.CodeUnauthorized {
		t.Errorf("非法签名应 401，得 %d", got)
	}

	// 算法伪造（none）→ 401（WithValidMethods 仅允许 HS512）
	noneSigned, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{constant.LoginUserKey: "uuid-x"}).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+noneSigned)); got != response.CodeUnauthorized {
		t.Errorf("none 算法应 401，得 %d", got)
	}

	// JWT 合法但 Redis 会话不存在（过期/被踢）→ 401
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-missing"))); got != response.CodeUnauthorized {
		t.Errorf("无会话应 401，得 %d", got)
	}

	// 垃圾 token → 401（Auth 不拦截，由权限中间件返回）
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer garbage")); got != response.CodeUnauthorized {
		t.Errorf("垃圾 token 应 401，得 %d", got)
	}
}

func TestAuthWhitelistSkip(t *testing.T) {
	_, r := newTestEnv(t)
	r.GET("/login", Wrap(func(c *gin.Context) error {
		response.Ok(c).Msg("login page").JSON()
		return nil
	}))
	// 白名单路径无 token 直接到达 handler
	if got := bodyCode(t, do(r, "/login", "")); got != response.CodeSuccess {
		t.Errorf("白名单 /login 应免认证放行，得 %d", got)
	}
}

func TestAuthFastJsonCompat(t *testing.T) {
	rc, r := newTestEnv(t)
	r.GET("/system/user/list", RequirePerm("system:user:query"), Wrap(func(c *gin.Context) error {
		lu := GetLoginUser(c)
		response.Ok(c).Put("username", lu.Username()).Put("roles", lu.RoleKeys()).JSON()
		return nil
	}))

	// Java FastJson 写入格式：顶层与 user/roles 均带 @type 头
	raw := `{"@type":"com.ruoyi.common.core.domain.model.LoginUser","userId":2,"deptId":101,` +
		`"token":"uuid-ry","loginTime":1700000000000,"expireTime":4102444800000,` +
		`"ipaddr":"127.0.0.1","loginLocation":"内网","browser":"Chrome","os":"Windows 10",` +
		`"permissions":["system:user:query"],"user":{"@type":"com.ruoyi.common.core.domain.entity.SysUser",` +
		`"userName":"ry","roles":[{"@type":"com.ruoyi.common.core.domain.entity.SysRole","roleKey":"common"}]}}`
	if err := rc.SetRaw(context.Background(), cache.BuildKey(constant.LoginTokenKey, "uuid-ry"), raw, 30*time.Minute); err != nil {
		t.Fatalf("seed raw: %v", err)
	}

	w := do(r, "/system/user/list", "Bearer "+signToken("uuid-ry"))
	if got := bodyCode(t, w); got != response.CodeSuccess {
		t.Fatalf("Java FastJson 会话应可解析并放行，得 %d", got)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"username":"ry"`) || !strings.Contains(body, `"roles":["common"]`) {
		t.Errorf("会话字段解析不符: %s", body)
	}
}

func TestVerifyTokenRefresh(t *testing.T) {
	rc, r := newTestEnv(t)
	addProtected(r)
	ctx := context.Background()

	// 剩余 < 20 分钟：请求触发续期（loginTime/expireTime 更新）
	lu := makeLoginUser(2, []string{"system:user:list"}, "common")
	now := time.Now().UnixMilli()
	lu.LoginTime = now - 30*60*1000
	lu.ExpireTime = now + 10*60*1000 // 剩余 10 分钟
	seedSession(t, rc, "uuid-refresh", lu)
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-refresh"))); got != response.CodeSuccess {
		t.Fatalf("续期场景请求应成功，得 %d", got)
	}
	var refreshed model.LoginUser
	if err := rc.GetObject(ctx, cache.BuildKey(constant.LoginTokenKey, "uuid-refresh"), &refreshed); err != nil {
		t.Fatalf("读取续期后会话: %v", err)
	}
	if refreshed.LoginTime <= lu.LoginTime {
		t.Errorf("续期后 loginTime 应更新: old=%d new=%d", lu.LoginTime, refreshed.LoginTime)
	}
	if refreshed.ExpireTime <= lu.ExpireTime {
		t.Errorf("续期后 expireTime 应为 now+30min: old=%d new=%d", lu.ExpireTime, refreshed.ExpireTime)
	}

	// 剩余 > 20 分钟：不续期
	lu2 := makeLoginUser(2, []string{"system:user:list"}, "common")
	lu2.LoginTime = now - 1*60*1000
	lu2.ExpireTime = now + 29*60*1000 // 剩余 29 分钟
	seedSession(t, rc, "uuid-fresh", lu2)
	if got := bodyCode(t, do(r, "/system/user/list", "Bearer "+signToken("uuid-fresh"))); got != response.CodeSuccess {
		t.Fatalf("不续期场景请求应成功，得 %d", got)
	}
	var untouched model.LoginUser
	if err := rc.GetObject(ctx, cache.BuildKey(constant.LoginTokenKey, "uuid-fresh"), &untouched); err != nil {
		t.Fatalf("读取未续期会话: %v", err)
	}
	if untouched.LoginTime != lu2.LoginTime {
		t.Errorf("剩余充足时不应续期: old=%d new=%d", lu2.LoginTime, untouched.LoginTime)
	}
}
