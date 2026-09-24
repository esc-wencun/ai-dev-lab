package page

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	bizerrors "ruoyi-vue-go/internal/common/errors"
)

// 模拟业务表（对位"分页查询带 where/order 的实体列表"）
type demoDO struct {
	ID     int64  `gorm:"primaryKey"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 内存库失败: %v", err)
	}
	if err := db.AutoMigrate(&demoDO{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	for i := 1; i <= 15; i++ {
		status := "0"
		if i%3 == 0 {
			status = "1"
		}
		if err := db.Create(&demoDO{ID: int64(i), Name: fmt.Sprintf("name-%02d", i), Status: status}).Error; err != nil {
			t.Fatalf("造数失败: %v", err)
		}
	}
	return db
}

func ginCtx(t *testing.T, query string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/demo"+query, nil)
	c.Request = req
	return c
}

// TestCamelToUnderScore 驼峰转下划线（对位 StringUtils.toUnderScoreCase 常见输入）
func TestCamelToUnderScore(t *testing.T) {
	cases := map[string]string{
		"userId":     "user_id",
		"createTime": "create_time",
		"UserName":   "user_name",
		"deptID":     "dept_i_d",
		"status":     "status",
		"":           "",
	}
	for in, want := range cases {
		if got := camelToUnderScore(in); got != want {
			t.Errorf("camelToUnderScore(%q)=%q, 期望 %q", in, got, want)
		}
	}
}

// TestToOrderBy 排序片段生成 + 前端 ascending/descending 兼容
func TestToOrderBy(t *testing.T) {
	d := Domain{OrderByColumn: "createTime", IsAsc: "descending"}
	got, err := d.ToOrderBy()
	if err != nil {
		t.Fatalf("合法排序不应报错: %v", err)
	}
	if got != "create_time desc" {
		t.Errorf("排序片段=%q, 期望 %q", got, "create_time desc")
	}

	d2 := Domain{OrderByColumn: "userId", IsAsc: "ascending"}
	got2, err := d2.ToOrderBy()
	if err != nil {
		t.Fatalf("ascending 兼容失败: %v", err)
	}
	if got2 != "user_id asc" {
		t.Errorf("排序片段=%q, 期望 %q", got2, "user_id asc")
	}

	d3 := Domain{}
	got3, err := d3.ToOrderBy()
	if err != nil || got3 != "asc" {
		t.Errorf("空排序列应仅剩方向: got=%q err=%v", got3, err)
	}
}

// TestToOrderByRejectInjection 非法排序字段/方向/超长必须拒绝（白名单防注入）
func TestToOrderByRejectInjection(t *testing.T) {
	badColumns := []string{
		"userId; drop table x", // SQL 注入
		"userId--",
		"create_time'",
		"userId or 1=1",
	}
	for _, col := range badColumns {
		d := Domain{OrderByColumn: col}
		_, err := d.ToOrderBy()
		var be *bizerrors.BusinessError
		if !asBusiness(err, &be) || be.Code != 500 {
			t.Errorf("排序字段 %q 应被拒绝为 BusinessError(500)，实际 %v", col, err)
		}
	}

	// 未知排序方向（非 asc/desc/ascending/descending）被白名单拒绝
	d := Domain{OrderByColumn: "userId", IsAsc: "delete"}
	if _, err := d.ToOrderBy(); !asBusinessErr(err) {
		t.Errorf("非法 isAsc 应被拒绝，实际 %v", err)
	}

	// 超过 500 长度
	long := make([]byte, 501)
	for i := range long {
		long[i] = 'a'
	}
	d2 := Domain{OrderByColumn: string(long)}
	_, err := d2.ToOrderBy()
	var be *bizerrors.BusinessError
	if !asBusiness(err, &be) || be.Msg == "" {
		t.Errorf("超长排序应拒绝，实际 %v", err)
	}
}

// TestFromGin 分页参数读取：默认值、自定义值、非法值回退默认
func TestFromGin(t *testing.T) {
	c := ginCtx(t, "")
	d := FromGin(c)
	if d.PageNum != 1 || d.PageSize != 10 || d.IsAsc != "asc" {
		t.Errorf("默认值错误: %+v", d)
	}

	c2 := ginCtx(t, "?pageNum=3&pageSize=20&orderByColumn=userName&isAsc=descending")
	d2 := FromGin(c2)
	if d2.PageNum != 3 || d2.PageSize != 20 || d2.OrderByColumn != "userName" || d2.IsAsc != "descending" {
		t.Errorf("自定义值错误: %+v", d2)
	}

	c3 := ginCtx(t, "?pageNum=abc&pageSize=-1")
	d3 := FromGin(c3)
	if d3.PageNum != 1 || d3.PageSize != 10 {
		t.Errorf("非法值应回退默认: %+v", d3)
	}
}

// TestPaginate 分页执行：total 计算、页大小、偏移、排序方向（sqlite 内存库）
func TestPaginate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	var rows []demoDO

	// 第 1 页：total=15，rows=10
	total, err := Paginate(ctx, db.Model(&demoDO{}), &rows, Domain{PageNum: 1, PageSize: 10}, true)
	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}
	if total != 15 || len(rows) != 10 {
		t.Errorf("第1页 total=%d rows=%d, 期望 15/10", total, len(rows))
	}

	// 第 2 页：rows=5
	rows = nil
	total2, err := Paginate(ctx, db.Model(&demoDO{}), &rows, Domain{PageNum: 2, PageSize: 10}, true)
	if err != nil {
		t.Fatalf("第2页查询失败: %v", err)
	}
	if total2 != 15 || len(rows) != 5 {
		t.Errorf("第2页 total=%d rows=%d, 期望 15/5", total2, len(rows))
	}

	// 带排序倒序：第一行应为 name-15
	rows = nil
	_, err = Paginate(ctx, db.Model(&demoDO{}), &rows, Domain{PageNum: 1, PageSize: 10, OrderByColumn: "name", IsAsc: "desc"}, true)
	if err != nil {
		t.Fatalf("带排序查询失败: %v", err)
	}
	if len(rows) == 0 || rows[0].Name != "name-15" {
		t.Errorf("倒序第一行应为 name-15，实际 %+v", rows[0])
	}

	// 不分页：全量 15 行
	rows = nil
	total3, err := Paginate(ctx, db.Model(&demoDO{}), &rows, Domain{}, false)
	if err != nil {
		t.Fatalf("全量查询失败: %v", err)
	}
	if total3 != 15 || len(rows) != 15 {
		t.Errorf("全量 total=%d rows=%d, 期望 15/15", total3, len(rows))
	}

	// 带 where 条件的分页 total 只数满足条件的
	rows = nil
	total4, err := Paginate(ctx, db.Model(&demoDO{}).Where("status = ?", "1"), &rows, Domain{PageNum: 1, PageSize: 2}, true)
	if err != nil {
		t.Fatalf("带条件分页失败: %v", err)
	}
	if total4 != 5 || len(rows) != 2 {
		t.Errorf("带条件分页 total=%d rows=%d, 期望 5/2", total4, len(rows))
	}

	// 非法排序字段在 Paginate 中报 BusinessError
	rows = nil
	_, err = Paginate(ctx, db.Model(&demoDO{}), &rows, Domain{PageNum: 1, PageSize: 10, OrderByColumn: "1=1"}, true)
	if !asBusinessErr(err) {
		t.Errorf("非法排序应报错，实际 %v", err)
	}
}

func asBusiness(err error, target **bizerrors.BusinessError) bool {
	if be, ok := err.(*bizerrors.BusinessError); ok {
		*target = be
		return true
	}
	return false
}

func asBusinessErr(err error) bool {
	var be *bizerrors.BusinessError
	return asBusiness(err, &be)
}
