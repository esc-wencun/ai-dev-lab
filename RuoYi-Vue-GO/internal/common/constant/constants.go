package constant

// 本文件对位 Constants.java（通用常量）。
// 未移植项（Java 平台特有，Go 无对应物）：UTF8/GBK/DEFAULT_LOCALE/WWW/HTTP/HTTPS、
// LOOKUP_RMI/LDAP/LDAPS（JNDI 防护）、JSON_WHITELIST_STR/JOB_WHITELIST_STR/JOB_ERROR_STR
// （FastJson 与 quartz 包名白名单）。其余逐一对表移植。
const (
	// 通用成功 / 失败标识
	Success = "0"
	Fail    = "1"

	// 登录成功 / 注销 / 注册 / 登录失败（操作日志 status 场景文案）
	LoginSuccess = "Success"
	Logout       = "Logout"
	Register     = "Register"
	LoginFail    = "Error"

	// 权限标识
	AllPermission = "*:*:*" // 所有权限标识
	SuperAdmin    = "admin" // 管理员角色权限标识

	// 分隔符
	RoleDelimiter       = "," // 角色权限分隔符
	PermissionDelimiter = "," // 权限标识分隔符

	// 验证码有效期（分钟）
	CaptchaExpiration = 2

	// 令牌
	Token        = "token"
	TokenPrefix  = "Bearer "
	LoginUserKey = "login_user_key"

	// JWT claims key（JwtUsername 对位 io.jsonwebtoken Claims.SUBJECT = "sub"）
	JwtUserid      = "userid"
	JwtUsername    = "sub"
	JwtAvatar      = "avatar"
	JwtCreated     = "created"
	JwtAuthorities = "authorities"

	// 资源映射路径前缀（头像/上传文件访问路径）
	ResourcePrefix = "/profile"
)

// 数据权限范围（对位 Constants.Dept 内嵌常量，sys_role.data_scope 落库值）
const (
	DataScopeAll          = "1" // 全部数据权限
	DataScopeCustom       = "2" // 自定数据权限
	DataScopeDept         = "3" // 部门数据权限
	DataScopeDeptAndChild = "4" // 部门及以下数据权限
	DataScopeSelf         = "5" // 仅本人数据权限
)
