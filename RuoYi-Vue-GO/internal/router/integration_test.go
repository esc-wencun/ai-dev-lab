// 01-基础设施 checklist 联动验收：临时演示 handler 把分页+权限+操作日志+导出
// 全部横切能力挂上，httptest 端到端走通（不连共用库，sqlite 内存 + miniredis）。
package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/common/enums"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/aspect"
	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/pkg/response"
	"ruoyi-vue-go/pkg/utils/excel"
	"ruoyi-vue-go/pkg/utils/page"
)

// TestInfraIntegrationDemo 联动演示：/demo/list（分页+权限+日志）与 /demo/export（导出+日志）。
func TestInfraIntegrationDemo(t *testing.T) {
	// 环境：sqlite 内存库
	db, err := gorm.Open(sqlite.Open("file:demo_infra?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite 打开失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&do.SysOperLog{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	// 演示业务表 + 种子数据（15 行，验证分页 total/页大小）
	type demoItem struct {
		ID     int64  `gorm:"column:id;primaryKey"`
		Status string `gorm:"column:status"`
	}
	if err := db.Exec("CREATE TABLE demo_item (id int primary key, status text)").Error; err != nil {
		t.Fatalf("建演示表失败: %v", err)
	}
	for i := 1; i <= 15; i++ {
		status := "0"
		if i%2 == 0 {
			status = "1"
		}
		db.Exec("INSERT INTO demo_item (id, status) VALUES (?, ?)", i, status)
	}

	logger := aspect.NewOperLogger(db)

	// 演示 handler：list（分页 TableDataInfo 响应）+ export（xlsx 流）
	listHandler := logger.OperLog(aspect.OperLogOpt("演示模块", enums.BusinessTypeOther), func(c *gin.Context) error {
		lu := middleware.GetLoginUser(c)
		if !middleware.HasPerm(lu, "demo:list") {
			response.Error(c, "无权限").JSON()
			return nil
		}
		var rows []demoItem
		total, err := page.Paginate(c.Request.Context(), db.Table("demo_item"), &rows, page.FromGin(c), true)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功", "rows": rows, "total": total})
		return nil
	})
	exportHandler := logger.OperLog(aspect.OperLogOpt("演示模块", enums.BusinessTypeExport), func(c *gin.Context) error {
		cols := []excel.Column{
			{Title: "编号", Field: "id", Width: 8},
			{Title: "状态", Field: "status", Converter: map[string]string{"0": "正常", "1": "停用"}},
		}
		rows := []map[string]any{
			{"id": int64(1), "status": "0"},
			{"id": int64(2), "status": "1"},
		}
		excel.Export(c.Writer, "演示数据", cols, rows)
		return nil
	})

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.CtxKeyLoginUser, adminSession()) }, gin.Recovery())
	r.GET("/demo/list", middleware.Wrap(listHandler))
	r.GET("/demo/export", middleware.Wrap(exportHandler))

	// 1) list：权限放行 + TableDataInfo（rows/total 顶层字段，默认 pageSize=10 → 10 行 15 总数）
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/demo/list", nil))
	var envelope map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil || envelope["code"].(float64) != 200 {
		t.Fatalf("list 响应异常: %s", w.Body.String())
	}
	if envelope["total"].(float64) != 15 {
		t.Errorf("分页 total = %v, want 15", envelope["total"])
	}
	if rows, ok := envelope["rows"].([]any); !ok || len(rows) != 10 {
		t.Errorf("分页 rows = %v, want 10 行（默认 pageSize）", envelope["rows"])
	}

	// 2) export：xlsx 头 + 下载头
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/demo/export", nil))
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "spreadsheetml") {
		t.Errorf("export Content-Type = %q", ct)
	}
	if w.Header().Get("download-filename") == "" {
		t.Error("export 缺 download-filename 头")
	}

	// 3) 无权限场景：非 admin 会话 403 信封
	r2 := gin.New()
	r2.Use(func(c *gin.Context) { c.Set(middleware.CtxKeyLoginUser, rySession()) }, gin.Recovery())
	r2.GET("/demo/list", middleware.RequirePerm("demo:list"), middleware.Wrap(listHandler))
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/demo/list", nil))
	if !strings.Contains(w.Body.String(), `"code":403`) {
		t.Errorf("无权限应 403 信封: %s", w.Body.String())
	}

	// 4) 操作日志落库：两次成功请求（list+export）+ 1 次 403（RequirePerm 在日志切面外，不落日志）
	var count int64
	for i := 0; i < 100; i++ {
		db.Table("sys_oper_log").Count(&count)
		if count >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if count < 2 {
		t.Fatalf("操作日志应落库 >=2 条, got %d", count)
	}
	var logs []do.SysOperLog
	db.Order("oper_id").Find(&logs)
	if logs[0].Title != "演示模块" || logs[0].Status != int(enums.BusinessStatusSuccess) {
		t.Errorf("日志1 = title %q status %d", logs[0].Title, logs[0].Status)
	}
	if logs[1].BusinessType != int(enums.BusinessTypeExport) {
		t.Errorf("日志2 businessType = %d, want Export", logs[1].BusinessType)
	}
	if !strings.Contains(logs[1].OperURL, "/demo/export") {
		t.Errorf("日志2 URL = %q", logs[1].OperURL)
	}
}

// adminSession admin 会话（HasPerm 对 userId=1 全放行）。
func adminSession() *model.LoginUser {
	return &model.LoginUser{UserId: 1, User: json.RawMessage(`{"userId":1,"userName":"admin","dept":{"deptName":"研发部门"}}`)}
}

// rySession 普通用户会话（无 demo 权限）。
func rySession() *model.LoginUser {
	return &model.LoginUser{
		UserId:      2,
		Permissions: []string{"system:user:list"},
		User:        json.RawMessage(`{"userId":2,"userName":"ry","roles":[]}`),
	}
}
