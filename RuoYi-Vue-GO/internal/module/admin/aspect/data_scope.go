// Package aspect 业务切面（对位 Java com.ruoyi.framework.aspectj）。
//
// 本文件对位 DataScopeAspect + @DataScope 注解：按会话用户的角色 data_scope
// 生成数据权限过滤条件。注入方式为 GORM scope 函数，业务 dao 调用时声明：
//
//	db.Scopes(aspect.DataScope(lu, aspect.DataScopeOptions{UserAlias: "u"}))
//
// 条件语义与 Java dataScopeFilter 逐条对齐（MySQL 与 Java 同库，find_in_set 直接可用）：
//   - admin（userId=1）或任一角色 scope=1：不加条件（不过滤）；
//   - scope=2 自定：dept IN (SELECT dept_id FROM sys_role_dept WHERE role_id = ...)；
//   - scope=3 本部门：dept = 用户 deptId；
//   - scope=4 本部门及以下：dept IN (SELECT dept_id FROM sys_dept WHERE dept_id = ? or find_in_set(?, ancestors))；
//   - scope=5 仅本人：有用户别名时 user = userId，否则 dept = 0（查不到数据）；
//   - 角色按 scope 值去重（首个生效）；停用角色、权限字符不匹配的角色跳过；
//   - 无任何角色生效：强制 dept = 0（不查询任何数据，对位 conditions 空判断）。
package aspect

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/model"
)

// DataScopeOptions 对位 Java @DataScope 注解属性。
//
// 别名约定：DeptAlias 默认 d（对齐 Java 调用点惯例）；UserAlias 默认空串
// （对齐 Java 注解默认值——默认 u 会让部门列表等无用户表的查询在 scope=5
// 时引用不存在的别名报错，Java 此场景就是 dept=0 空集）。需要用户别名过滤的
// 调用点（如用户列表）显式传 UserAlias: "u"，与 Java 写法一一对应。
type DataScopeOptions struct {
	DeptAlias  string // 部门表别名（默认 d）
	UserAlias  string // 用户表别名（默认空，SELF 场景查空集）
	DeptField  string // 部门字段名（默认 dept_id）
	UserField  string // 用户字段名（默认 user_id）
	Permission string // 权限字符，多个逗号分隔（对位 permission()，空则不过滤角色）
}

func (o DataScopeOptions) withDefaults() DataScopeOptions {
	if o.DeptAlias == "" {
		o.DeptAlias = "d"
	}
	if o.DeptField == "" {
		o.DeptField = "dept_id"
	}
	if o.UserField == "" {
		o.UserField = "user_id"
	}
	return o
}

// dsRoleView 会话内嵌 SysUser.roles[] 中数据权限所需字段（对位 SysRole；
// 登录时 Java getMenuPermission 已按角色填充 permissions，会话 JSON 里就有）。
type dsRoleView struct {
	RoleID      int64    `json:"roleId"`
	DataScope   string   `json:"dataScope"`
	Status      string   `json:"status"`
	Permissions []string `json:"permissions"`
}

// dsUserView 会话内嵌 SysUser 中数据权限所需字段。
type dsUserView struct {
	UserID int64        `json:"userId"`
	DeptID int64        `json:"deptId"`
	Roles  []dsRoleView `json:"roles"`
}

// DataScope 数据权限过滤 scope（对位 DataScopeAspect.dataScopeFilter + BaseEntity.params.dataScope 注入）。
// lu 为当前登录会话（middleware.GetLoginUser 取得），由调用方保证已过认证。
func DataScope(lu *model.LoginUser, opt DataScopeOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cond, vals := buildDataScope(lu, opt)
		if cond == "" {
			return db
		}
		return db.Where(gorm.Expr(cond, vals...))
	}
}

// buildDataScope 生成过滤条件（对位 dataScopeFilter 的 SQL 拼接逻辑）。
// 返回空串 = 不过滤（admin / 任一角色全部权限 / 会话缺用户信息）；
// 返回 "(dept = 0)" 形态 = 查询不到任何数据。
func buildDataScope(lu *model.LoginUser, opt DataScopeOptions) (string, []any) {
	opt = opt.withDefaults()
	if lu == nil || lu.UserId == 1 {
		return "", nil // 对位 isAdmin 不过滤
	}
	var u dsUserView
	if len(lu.User) == 0 || json.Unmarshal(lu.User, &u) != nil {
		return "", nil // 对位 currentUser 为 null 不过滤（实际链路不会发生）
	}

	var perms []string
	if p := strings.TrimSpace(opt.Permission); p != "" {
		perms = strings.Split(p, ",")
	}

	// 第一遍：收集参与过滤的自定权限角色 id（对位 scopeCustomIds）
	var customIDs []int64
	for _, r := range u.Roles {
		if r.DataScope == constant.DataScopeCustom && r.Status != constant.RoleDisable && dsPermMatch(perms, r.Permissions) {
			customIDs = append(customIDs, r.RoleID)
		}
	}

	deptCol := opt.DeptAlias + "." + opt.DeptField
	userCol := opt.UserAlias + "." + opt.UserField

	var parts []string
	var vals []any
	seen := map[string]bool{}
	sawRole := false
	for _, r := range u.Roles {
		if seen[r.DataScope] || r.Status == constant.RoleDisable { // 对位 conditions.contains + 停用跳过
			continue
		}
		if len(perms) > 0 && !dsPermMatch(perms, r.Permissions) { // 对位权限字符不匹配跳过
			continue
		}
		seen[r.DataScope] = true
		sawRole = true
		switch r.DataScope {
		case constant.DataScopeAll:
			return "", nil // 全部权限：清空已拼条件并短路（对位 reset + break）
		case constant.DataScopeCustom:
			if len(customIDs) > 1 { // 多个自定权限合并 IN，避免重复拼接（对位同名分支）
				ph := strings.TrimSuffix(strings.Repeat("?,", len(customIDs)), ",")
				parts = append(parts, deptCol+" IN ( SELECT dept_id FROM sys_role_dept WHERE role_id in ("+ph+") )")
				for _, id := range customIDs {
					vals = append(vals, id)
				}
			} else {
				parts = append(parts, deptCol+" IN ( SELECT dept_id FROM sys_role_dept WHERE role_id = ? )")
				vals = append(vals, r.RoleID)
			}
		case constant.DataScopeDept:
			parts = append(parts, deptCol+" = ?")
			vals = append(vals, u.DeptID)
		case constant.DataScopeDeptAndChild:
			parts = append(parts, deptCol+" IN ( SELECT dept_id FROM sys_dept WHERE dept_id = ? or find_in_set( ? , ancestors ) )")
			vals = append(vals, u.DeptID, u.DeptID)
		case constant.DataScopeSelf:
			if opt.UserAlias != "" {
				parts = append(parts, userCol+" = ?")
				vals = append(vals, u.UserID)
			} else {
				// 仅本人且没有用户别名：不查询任何数据（对位同名分支）
				parts = append(parts, deptCol+" = 0")
			}
		}
	}
	if !sawRole { // 角色都不匹配：不查询任何数据（对位 conditions 空判断）
		parts = append(parts, deptCol+" = 0")
	}
	if len(parts) == 0 {
		return "", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", vals
}

// dsPermMatch 权限字符匹配（对位 StringUtils.containsAny(collection, array)）：
// 未指定权限字符所有角色参与；角色 permissions 为空/缺失视为不匹配。
func dsPermMatch(perms, rolePerms []string) bool {
	if len(perms) == 0 {
		return true
	}
	for _, rp := range rolePerms {
		for _, p := range perms {
			if rp == p {
				return true
			}
		}
	}
	return false
}
