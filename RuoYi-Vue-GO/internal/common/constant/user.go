package constant

// 本文件对位 UserConstants.java（用户/菜单常量）。
const (
	// 平台内系统用户的唯一标志
	SysUser = "SYS_USER"

	// 正常 / 异常状态（通用）
	Normal    = "0"
	Exception = "1"

	// 用户封禁状态
	UserDisable = "1"

	// 角色正常 / 封禁状态
	RoleNormal  = "0"
	RoleDisable = "1"

	// 部门正常 / 停用状态
	DeptNormal  = "0"
	DeptDisable = "1"

	// 字典正常状态
	DictNormal = "0"

	// 是否为系统默认（是）
	Yes = "Y"

	// 是否菜单外链（是 / 否）
	YesFrame = "0"
	NoFrame  = "1"

	// 菜单类型（目录 / 菜单 / 按钮）
	TypeDir    = "M"
	TypeMenu   = "C"
	TypeButton = "F"

	// 前端组件标识
	Layout     = "Layout"
	ParentView = "ParentView"
	InnerLink  = "InnerLink"

	// 校验是否唯一的返回标识
	Unique    = true
	NotUnique = false

	// 用户名长度限制
	UsernameMinLength = 2
	UsernameMaxLength = 20

	// 密码长度限制
	PasswordMinLength = 5
	PasswordMaxLength = 20
)
