// Package model 通用领域模型（对位 Java com.ruoyi.common.core.domain.model）。
package model

import (
	"encoding/json"
)

// LoginUser 登录用户会话（与 Java 版共享 Redis 键 login_tokens:{uuid}，JSON 字段名逐字对齐）。
//
// 序列化兼容：
//   - Java FastJson 写入的 JSON 带 "@type" 头，encoding/json 对未知字段默认忽略，可无损读取；
//   - User 用 json.RawMessage 原样保留：Go 侧续期会话时整体回写 Redis，
//     RawMessage 保证不因 Go 结构体缺字段而丢弃 SysUser 数据（也无须在此定义完整 SysUser）。
type LoginUser struct {
	UserId        int64           `json:"userId"`
	DeptId        int64           `json:"deptId"`
	Token         string          `json:"token"` // 会话 uuid，即 Redis key login_tokens: 的后缀
	LoginTime     int64           `json:"loginTime"`
	ExpireTime    int64           `json:"expireTime"`
	Ipaddr        string          `json:"ipaddr"`
	LoginLocation string          `json:"loginLocation"`
	Browser       string          `json:"browser"`
	Os            string          `json:"os"`
	Permissions   []string        `json:"permissions"`
	User          json.RawMessage `json:"user"` // SysUser 原始 JSON
}

// Username 从 User 原始 JSON 取登录账号（对位 LoginUser.getUsername() → SysUser.userName）。
func (u *LoginUser) Username() string {
	if len(u.User) == 0 {
		return ""
	}
	var view struct {
		UserName string `json:"userName"`
	}
	if err := json.Unmarshal(u.User, &view); err != nil {
		return ""
	}
	return view.UserName
}

// DeptName 从 User 原始 JSON 取部门名称（SysUser.dept.deptName，
// 对位 LogAspect 中 operLog.setDeptName(currentUser.getDept().getDeptName())）。
func (u *LoginUser) DeptName() string {
	if len(u.User) == 0 {
		return ""
	}
	var view struct {
		Dept struct {
			DeptName string `json:"deptName"`
		} `json:"dept"`
	}
	if err := json.Unmarshal(u.User, &view); err != nil {
		return ""
	}
	return view.Dept.DeptName
}

// RoleKeys 返回用户持有的角色权限字符列表（SysRole.roleKey，对位 getUser().getRoles()）。
// 只做只读视图解析，不回写，保证续期回写时 User 字段不丢数据。
func (u *LoginUser) RoleKeys() []string {
	if len(u.User) == 0 {
		return nil
	}
	var view struct {
		Roles []struct {
			RoleKey string `json:"roleKey"`
		} `json:"roles"`
	}
	if err := json.Unmarshal(u.User, &view); err != nil {
		return nil
	}
	keys := make([]string, 0, len(view.Roles))
	for _, r := range view.Roles {
		keys = append(keys, r.RoleKey)
	}
	return keys
}
