package enums

// 本文件对位 UserStatus.java（用户状态，sys_user.status / del_flag 场景使用 code 落库）。
type UserStatus string

const (
	UserStatusOK      UserStatus = "0" // 正常
	UserStatusDisable UserStatus = "1" // 停用
	UserStatusDeleted UserStatus = "2" // 删除
)

// Info 返回 Java 枚举的 info 文案（正常 / 停用 / 删除）
func (s UserStatus) Info() string {
	switch s {
	case UserStatusOK:
		return "正常"
	case UserStatusDisable:
		return "停用"
	case UserStatusDeleted:
		return "删除"
	default:
		return "未知状态"
	}
}
