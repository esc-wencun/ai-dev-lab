// 会话内嵌用户 JSON 视图（对位 Java LoginUser.user 的 FastJSON 序列化形态）。
package service

import (
	"encoding/json"

	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/module/admin/dao"
)

// UserSessionView 会话内嵌用户 JSON 视图（字段名对齐 Java SysUser 序列化；
// dept/roles 内嵌、不含 password——getInfo 的 user 字段即此形态）。
func UserSessionView(ud *dao.UserDetail) map[string]any {
	u := ud.User
	view := map[string]any{
		"userId":        u.UserID,
		"deptId":        u.DeptID,
		"userName":      u.UserName,
		"nickName":      u.NickName,
		"email":         u.Email,
		"avatar":        u.Avatar,
		"phonenumber":   u.Phonenumber,
		"sex":           u.Sex,
		"status":        u.Status,
		"loginIp":       u.LoginIP,
		"loginDate":     u.LoginDate,
		"pwdUpdateDate": u.PwdUpdateDate,
		"createTime":    u.CreateTime,
		"remark":        u.Remark,
	}
	if ud.Dept != nil {
		view["dept"] = map[string]any{
			"deptId":    ud.Dept.DeptID,
			"parentId":  ud.Dept.ParentID,
			"ancestors": ud.Dept.Ancestors,
			"deptName":  ud.Dept.DeptName,
			"orderNum":  ud.Dept.OrderNum,
			"leader":    ud.Dept.Leader,
			"status":    ud.Dept.Status,
		}
	}
	if len(ud.Roles) > 0 {
		roleList := make([]map[string]any, 0, len(ud.Roles))
		for _, r := range ud.Roles {
			roleList = append(roleList, map[string]any{
				"roleId":    r.RoleID,
				"roleName":  r.RoleName,
				"roleKey":   r.RoleKey,
				"roleSort":  r.RoleSort,
				"dataScope": r.DataScope,
				"status":    r.Status,
			})
		}
		view["roles"] = roleList
	}
	return view
}

// SessionUserView 简短别名（profile_service 使用）。
func SessionUserView(ud *dao.UserDetail) map[string]any { return UserSessionView(ud) }

// BuildLoginUser 组装会话对象（对位 new LoginUser(userId, deptId, user, permissions)）。
func BuildLoginUser(ud *dao.UserDetail, roles, permissions []string) (*model.LoginUser, error) {
	data, err := json.Marshal(UserSessionView(ud))
	if err != nil {
		return nil, err
	}
	return &model.LoginUser{
		UserId:      ud.User.UserID,
		DeptId:      ud.User.DeptID,
		Permissions: permissions,
		User:        data,
	}, nil
}
