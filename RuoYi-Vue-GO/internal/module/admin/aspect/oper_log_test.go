package aspect

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"

	"ruoyi-vue-go/internal/common/enums"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/pkg/response"
)

// newTestDB sqlite 内存库（每测试独立命名库 + 单连接，保证异步 goroutine
// 与测试看到同一内存库，且测试之间互不污染）。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
	dsn := fmt.Sprintf("file:operlog_%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&do.SysOperLog{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	return db
}

// waitFirstLog 轮询等待异步落库完成并返回首条记录（对位 AsyncManager 的异步语义）。
func waitFirstLog(t *testing.T, db *gorm.DB) do.SysOperLog {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var m do.SysOperLog
		if err := db.First(&m).Error; err == nil {
			return m
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("操作日志未落库")
	return do.SysOperLog{}
}

// fakeAuth 注入模拟登录会话（真实链路由 middleware.Auth 完成，测试不依赖 Redis）。
func fakeAuth(lu *model.LoginUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.CtxKeyLoginUser, lu)
		c.Next()
	}
}

// testLoginUser 会话样例（User 为原始 JSON，与 Java FastJson 形态一致）
var testLoginUser = &model.LoginUser{
	UserId: 1,
	User:   json.RawMessage(`{"userId":1,"userName":"admin","dept":{"deptId":100,"deptName":"研发部门"},"roles":[{"roleKey":"admin"}]}`),
}

// TestSanitizeJSON 敏感参数脱敏：内置名单顶层+嵌套全剔除、额外名单生效、非 JSON 返回空。
func TestSanitizeJSON(t *testing.T) {
	body := `{"userName":"u1","password":"p1","oldPassword":"p0","newPassword":"p2","confirmPassword":"p3","profile":{"phone":"13800000000","password":"inner"}}`
	got := sanitizeJSON([]byte(body), nil)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("脱敏结果不是合法 JSON: %v", err)
	}
	for _, k := range EXCLUDE_PROPERTIES {
		if _, ok := m[k]; ok {
			t.Errorf("顶层敏感字段 %q 未被剔除", k)
		}
	}
	profile, _ := m["profile"].(map[string]any)
	if profile == nil {
		t.Fatalf("嵌套字段丢失: %s", got)
	}
	if _, ok := profile["password"]; ok {
		t.Errorf("嵌套敏感字段 password 未被剔除: %s", got)
	}
	if profile["phone"] != "13800000000" {
		t.Errorf("非敏感字段被误删: %s", got)
	}

	// 额外剔除名单（对位 @Log(excludeParamNames=...)）
	got2 := sanitizeJSON([]byte(`{"userName":"u1","phone":"138"}`), []string{"phone"})
	if strings.Contains(got2, "phone") {
		t.Errorf("额外剔除名单未生效: %s", got2)
	}

	// 非 JSON / 空 body → 空串（对位 isFilterObject 跳过文件与流对象）
	if s := sanitizeJSON([]byte(`-----------------------------boundary`), nil); s != "" {
		t.Errorf("非 JSON body 应返回空串, got %q", s)
	}
	if s := sanitizeJSON(nil, nil); s != "" {
		t.Errorf("空 body 应返回空串, got %q", s)
	}
}

// TestTruncateRunes 截断逻辑（对位 StringUtils.substring 按字符数截断）。
func TestTruncateRunes(t *testing.T) {
	long := strings.Repeat("中", 2001)
	if got := truncateRunes(long, paramMaxLength); len([]rune(got)) != 2000 {
		t.Errorf("应截断到 2000 字符, got %d", len([]rune(got)))
	}
	if got := truncateRunes("short", 2000); got != "short" {
		t.Errorf("不超长不应改变: %q", got)
	}
}

// TestOperLogSuccess 端到端成功场景：POST JSON body 带 password → 落库 status=0、
// operParam 已脱敏、jsonResult 为响应体、操作人/部门/方法名/IP/URL/耗时齐全。
func TestOperLogSuccess(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	opt := OperLogOpt("用户管理", enums.BusinessTypeInsert)
	h := lg.OperLog(opt, func(c *gin.Context) error {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil { // 验证 body 快照已回放，handler 可正常读取
			return err
		}
		response.OkData(c, gin.H{"userName": body["userName"]}).JSON()
		return nil
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser))
	r.POST("/system/user", middleware.Wrap(h))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/system/user", strings.NewReader(`{"userName":"u1","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 响应不受日志切面影响
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp["code"].(float64) != 200 {
		t.Fatalf("响应异常: %s", w.Body.String())
	}

	m := waitFirstLog(t, db)
	if m.Title != "用户管理" {
		t.Errorf("title = %q", m.Title)
	}
	if m.BusinessType != int(enums.BusinessTypeInsert) || m.OperatorType != int(enums.OperatorTypeManage) {
		t.Errorf("枚举落库值 = businessType %d / operatorType %d", m.BusinessType, m.OperatorType)
	}
	if m.Status != int(enums.BusinessStatusSuccess) || m.ErrorMsg != "" {
		t.Errorf("成功场景 status/errorMsg = %d/%q", m.Status, m.ErrorMsg)
	}
	if m.OperName != "admin" || m.DeptName != "研发部门" {
		t.Errorf("操作人/部门 = %q/%q", m.OperName, m.DeptName)
	}
	if strings.Contains(m.OperParam, "secret") {
		t.Errorf("operParam 泄露敏感参数: %s", m.OperParam)
	}
	if !strings.Contains(m.OperParam, "u1") {
		t.Errorf("operParam 缺少非敏感参数: %s", m.OperParam)
	}
	if !strings.Contains(m.JsonResult, "u1") || !strings.Contains(m.JsonResult, `"code":200`) {
		t.Errorf("jsonResult 应为响应体: %s", m.JsonResult)
	}
	if m.RequestMethod != http.MethodPost || m.OperURL != "/system/user" {
		t.Errorf("method/url = %q/%q", m.RequestMethod, m.OperURL)
	}
	if !strings.Contains(m.Method, "TestOperLogSuccess") || !strings.HasSuffix(m.Method, "()") {
		t.Errorf("method 应为处理函数全名带括号: %q", m.Method)
	}
	if m.OperIP == "" {
		t.Errorf("operIP 未记录")
	}
	if m.CostTime < 0 {
		t.Errorf("costTime = %d", m.CostTime)
	}
	if m.OperTime.IsZero() {
		t.Errorf("operTime 未记录")
	}
}

// TestOperLogError 异常场景：error 记录 status=1 + errorMsg 后原样上抛（对位 @AfterThrowing）。
func TestOperLogError(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	boom := errors.New("业务处理失败：用户已存在")
	h := lg.OperLog(OperLogOpt("用户管理", enums.BusinessTypeInsert), func(c *gin.Context) error {
		return boom
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser))
	r.POST("/system/user", middleware.Wrap(h))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/system/user", strings.NewReader(`{"userName":"u1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Wrap 把上抛错误映射为 code=500 信封（错误继续上抛未被吞掉）
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp["code"].(float64) != 500 {
		t.Fatalf("错误信封异常: %s", w.Body.String())
	}

	m := waitFirstLog(t, db)
	if m.Status != int(enums.BusinessStatusFail) {
		t.Errorf("失败场景 status = %d", m.Status)
	}
	if m.ErrorMsg != boom.Error() {
		t.Errorf("errorMsg = %q, want %q", m.ErrorMsg, boom.Error())
	}
	if m.JsonResult != "" {
		t.Errorf("失败场景不应记录 jsonResult: %q", m.JsonResult)
	}
}

// TestOperLogQueryParams query 参数记录 + 脱敏（对位 setRequestValue 的 paramsMap 分支）。
func TestOperLogQueryParams(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	h := lg.OperLog(OperLogOpt("用户管理", enums.BusinessTypeOther), func(c *gin.Context) error {
		response.Ok(c).JSON()
		return nil
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser))
	r.GET("/system/user/list", middleware.Wrap(h))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/system/user/list?pageNum=1&pageSize=10&password=x", nil))

	m := waitFirstLog(t, db)
	if !strings.Contains(m.OperParam, `"pageNum":"1"`) || !strings.Contains(m.OperParam, "pageSize") {
		t.Errorf("operParam 缺少 query 参数: %s", m.OperParam)
	}
	if strings.Contains(m.OperParam, "password") {
		t.Errorf("query 中 password 未脱敏: %s", m.OperParam)
	}
}

// TestOperLogPanic panic 也记录（status=1）后继续上抛，交由 Recovery 兜底。
func TestOperLogPanic(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	h := lg.OperLog(OperLogOpt("用户管理", enums.BusinessTypeDelete), func(c *gin.Context) error {
		panic("数据库连接中断")
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser), gin.Recovery())
	r.DELETE("/system/user/1", middleware.Wrap(h))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/system/user/1", nil))

	m := waitFirstLog(t, db)
	if m.Status != int(enums.BusinessStatusFail) || !strings.Contains(m.ErrorMsg, "数据库连接中断") {
		t.Errorf("panic 场景 status/errorMsg = %d/%q", m.Status, m.ErrorMsg)
	}
}

// TestOperLogSaveFlags 关闭保存开关后 operParam/jsonResult 均不记录。
func TestOperLogSaveFlags(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	opt := OperLogOpt("用户管理", enums.BusinessTypeExport)
	opt.IsSaveRequestData = false
	opt.IsSaveResponseData = false
	h := lg.OperLog(opt, func(c *gin.Context) error {
		response.Ok(c).JSON()
		return nil
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser))
	r.GET("/system/user/export", middleware.Wrap(h))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/system/user/export?id=1", nil))

	m := waitFirstLog(t, db)
	if m.OperParam != "" || m.JsonResult != "" {
		t.Errorf("关闭保存开关后 operParam/jsonResult = %q/%q", m.OperParam, m.JsonResult)
	}
}

// TestFormParamsRecorded form-urlencoded 表单参数按 parameter map 记录（对位 Java getParamMap）。
func TestFormParamsRecorded(t *testing.T) {
	db := newTestDB(t)
	lg := NewOperLogger(db)

	h := lg.OperLog(OperLogOpt("参数管理", enums.BusinessTypeUpdate), func(c *gin.Context) error {
		response.Ok(c).JSON()
		return nil
	})

	r := gin.New()
	r.Use(fakeAuth(testLoginUser))
	r.POST("/system/config", middleware.Wrap(h))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/system/config", strings.NewReader("configName=参数A&configKey=sys.a&password=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	m := waitFirstLog(t, db)
	if !strings.Contains(m.OperParam, "configName") {
		t.Errorf("operParam 缺少表单参数: %s", m.OperParam)
	}
	if strings.Contains(m.OperParam, "password") {
		t.Errorf("表单 password 未脱敏: %s", m.OperParam)
	}
}
