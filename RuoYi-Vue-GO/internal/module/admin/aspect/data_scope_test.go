package aspect

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/model"
)

// dsUser 构造会话内嵌用户 JSON（roles 含数据权限所需字段，形态与 Java 会话一致）。
func dsUser(t *testing.T, userID, deptID int64, roles []dsRoleView) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(dsUserView{UserID: userID, DeptID: deptID, Roles: roles})
	if err != nil {
		t.Fatalf("构造用户 JSON 失败: %v", err)
	}
	return data
}

// dsLoginUser 构造会话。
func dsLoginUser(t *testing.T, userID, deptID int64, roles []dsRoleView) *model.LoginUser {
	return &model.LoginUser{
		UserId: userID,
		DeptId: deptID,
		User:   dsUser(t, userID, deptID, roles),
	}
}

// assertScope 断言生成的条件与参数（want="" 表示不过滤）。
func assertScope(t *testing.T, name, got string, vals []any, want string, wantVals ...any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\n got %q\nwant %q", name, got, want)
		return
	}
	if len(got) == 0 {
		return
	}
	if len(vals) != len(wantVals) {
		t.Errorf("%s: 参数个数 = %d, want %d", name, len(vals), len(wantVals))
		return
	}
	for i := range vals {
		if !sameVal(vals[i], wantVals[i]) {
			t.Errorf("%s: vals[%d] = %v(%T), want %v", name, i, vals[i], vals[i], wantVals[i])
		}
	}
}

// sameVal 数值跨 int/int64 比较，其余走字符串比较。
func sameVal(a, b any) bool {
	ai, aok := toInt64(a)
	bi, bok := toInt64(b)
	if aok && bok {
		return ai == bi
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int64:
		return x, true
	}
	return 0, false
}

// TestAdminNoFilter admin（userId=1）不过滤（对位 isAdmin 分支）。
func TestAdminNoFilter(t *testing.T) {
	lu := dsLoginUser(t, 1, 100, []dsRoleView{{RoleID: 2, DataScope: constant.DataScopeSelf, Status: constant.RoleNormal}})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "admin", got, vals, "")
}

// TestNilSessionNoFilter 无会话/无用户 JSON 不过滤（防御分支）。
func TestNilSessionNoFilter(t *testing.T) {
	got, _ := buildDataScope(nil, DataScopeOptions{})
	assertScope(t, "nil 会话", got, nil, "")
	lu := &model.LoginUser{UserId: 2}
	got, _ = buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "缺 user JSON", got, nil, "")
}

// TestScopeAll 任一角色全部权限 → 不过滤（对位 reset + break 短路）。
func TestScopeAll(t *testing.T) {
	lu := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
		{RoleID: 3, DataScope: constant.DataScopeAll, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "全部权限", got, vals, "")
}

// TestScopeCustom 自定权限：单角色 role_id = ?，多角色合并 IN（对位 scopeCustomIds）。
func TestScopeCustom(t *testing.T) {
	lu := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 3, DataScope: constant.DataScopeCustom, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "单自定",
		got, vals,
		"(d.dept_id IN ( SELECT dept_id FROM sys_role_dept WHERE role_id = ? ))",
		int64(3))

	lu2 := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 3, DataScope: constant.DataScopeCustom, Status: constant.RoleNormal},
		{RoleID: 4, DataScope: constant.DataScopeCustom, Status: constant.RoleNormal},
	})
	got2, vals2 := buildDataScope(lu2, DataScopeOptions{})
	assertScope(t, "多自定合并 IN",
		got2, vals2,
		"(d.dept_id IN ( SELECT dept_id FROM sys_role_dept WHERE role_id in (?,?) ))",
		int64(3), int64(4))
}

// TestScopeDept 本部门：dept = 用户 deptId。
func TestScopeDept(t *testing.T) {
	lu := dsLoginUser(t, 2, 105, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "本部门", got, vals, "(d.dept_id = ?)", 105)
}

// TestScopeDeptAndChild 本部门及以下：find_in_set 子查询。
func TestScopeDeptAndChild(t *testing.T) {
	lu := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDeptAndChild, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "本部门及以下",
		got, vals,
		"(d.dept_id IN ( SELECT dept_id FROM sys_dept WHERE dept_id = ? or find_in_set( ? , ancestors ) ))",
		100, 100)
}

// TestScopeSelf 仅本人：有用户别名 user = userId；无用户别名 dept = 0 空集。
func TestScopeSelf(t *testing.T) {
	lu := dsLoginUser(t, 7, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeSelf, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{UserAlias: "u"})
	assertScope(t, "仅本人带别名", got, vals, "(u.user_id = ?)", 7)

	got2, _ := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "仅本人无别名空集", got2, nil, "(d.dept_id = 0)")
}

// TestScopeDedupDisabledPerm 同 scope 去重（首个生效）、停用角色跳过、权限字符不匹配跳过。
func TestScopeDedupDisabledPerm(t *testing.T) {
	// 同 scope 两个角色只拼一次
	lu := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
		{RoleID: 3, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "同 scope 去重", got, vals, "(d.dept_id = ?)", 100)

	// 停用角色跳过 → 无生效角色 → 空集
	lu2 := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleDisable},
	})
	got2, _ := buildDataScope(lu2, DataScopeOptions{})
	assertScope(t, "停用角色后空集", got2, nil, "(d.dept_id = 0)")

	// 指定权限字符、角色无该权限 → 跳过 → 空集
	lu3 := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal, Permissions: []string{"system:role:list"}},
	})
	got3, _ := buildDataScope(lu3, DataScopeOptions{Permission: "system:user:list"})
	assertScope(t, "权限不匹配空集", got3, nil, "(d.dept_id = 0)")

	// 指定权限字符、角色持有其中之一 → 正常拼接
	lu4 := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal, Permissions: []string{"system:user:list"}},
	})
	got4, vals4 := buildDataScope(lu4, DataScopeOptions{Permission: "system:user:list,system:user:query"})
	assertScope(t, "权限匹配拼接", got4, vals4, "(d.dept_id = ?)", 100)
}

// TestScopeMultiOR 多个不同 scope 用 OR 合并。
func TestScopeMultiOR(t *testing.T) {
	lu := dsLoginUser(t, 7, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
		{RoleID: 3, DataScope: constant.DataScopeSelf, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{UserAlias: "u"})
	assertScope(t, "多 scope OR",
		got, vals,
		"(d.dept_id = ? OR u.user_id = ?)",
		100, 7)
}

// TestAliasDefaults 别名与字段默认值/覆盖（对位注解 deptAlias/userAlias/deptField/userField）。
func TestAliasDefaults(t *testing.T) {
	lu := dsLoginUser(t, 2, 100, []dsRoleView{
		{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal},
	})
	got, vals := buildDataScope(lu, DataScopeOptions{})
	assertScope(t, "默认别名 d.dept_id", got, vals, "(d.dept_id = ?)", 100)

	got2, vals2 := buildDataScope(lu, DataScopeOptions{DeptAlias: "dp", DeptField: "parent_id"})
	assertScope(t, "自定义别名", got2, vals2, "(dp.parent_id = ?)", 100)
}

// TestDataScopeGORMIntegration scope 函数挂到 GORM 查询验证 Where 注入真实生效
// （sqlite 内存库；MySQL 真库联查由 5.0.0-用户管理端到端覆盖）。
func TestDataScopeGORMIntegration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:datascope_gorm?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	if err := db.Exec("CREATE TABLE dept_mock (dept_id int)").Error; err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	db.Exec("INSERT INTO dept_mock (dept_id) VALUES (100),(105),(200)")

	count := func(lu *model.LoginUser, opt DataScopeOptions) int64 {
		var n int64
		if err := db.Table("dept_mock d").Scopes(DataScope(lu, opt)).Count(&n).Error; err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		return n
	}

	// scope=3 本部门（deptId=105）：只见 1 行
	luDept := dsLoginUser(t, 7, 105, []dsRoleView{{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleNormal}})
	if n := count(luDept, DataScopeOptions{}); n != 1 {
		t.Errorf("本部门应命中 1 行, got %d", n)
	}
	// scope=1 全部：3 行
	luAll := dsLoginUser(t, 7, 105, []dsRoleView{{RoleID: 2, DataScope: constant.DataScopeAll, Status: constant.RoleNormal}})
	if n := count(luAll, DataScopeOptions{}); n != 3 {
		t.Errorf("全部权限应命中 3 行, got %d", n)
	}
	// 无生效角色：空集
	luNone := dsLoginUser(t, 7, 105, []dsRoleView{{RoleID: 2, DataScope: constant.DataScopeDept, Status: constant.RoleDisable}})
	if n := count(luNone, DataScopeOptions{}); n != 0 {
		t.Errorf("无生效角色应为空集, got %d", n)
	}
	// admin：不过滤 3 行
	if n := count(dsLoginUser(t, 1, 105, nil), DataScopeOptions{}); n != 3 {
		t.Errorf("admin 应命中全部 3 行, got %d", n)
	}
}
