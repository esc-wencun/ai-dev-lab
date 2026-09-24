// Package message 面向用户的文案字典（对位 Java i18n messages.properties）。
//
// 规则：面向用户的文案必须引用本包常量，代码不留裸中文（日志除外）；
// key 命名对齐 messages.properties 的 properties key（user.jcaptcha.error → UserJcaptchaError）。
// {0}/{min} 占位符改为 fmt.Sprintf 的 %s/%d 占位，使用时按位传参。
package message

const (
	NotNull = "* 必须填写"

	UserJcaptchaError    = "验证码错误"
	UserJcaptchaExpire   = "验证码已失效"
	UserNotExists        = "用户不存在/密码错误"
	UserPasswordNotMatch = "用户不存在/密码错误"

	// fmt.Sprintf(UserPasswordRetryLimitCount, n)
	UserPasswordRetryLimitCount  = "密码输入错误%d次"
	UserPasswordRetryLimitExceed = "密码输入错误%d次，帐户锁定%d分钟"
	UserPasswordDelete           = "对不起，您的账号已被删除"
	UserBlocked                  = "用户已封禁，请联系管理员"
	RoleBlocked                  = "角色已封禁，请联系管理员"
	LoginBlocked                 = "很遗憾，访问IP已被列入系统黑名单"
	UserLogoutSuccess            = "退出成功"

	LengthNotValid = "长度必须在%d到%d个字符之间"

	UserUsernameNotValid = "* 2到20个汉字、字母、数字或下划线组成，且必须以非数字开头"
	UserPasswordNotValid = "* 5-50个字符"

	UserEmailNotValid             = "邮箱格式错误"
	UserMobilePhoneNumberNotValid = "手机号格式错误"
	UserLoginSuccess              = "登录成功"
	UserRegisterSuccess           = "注册成功"
	UserNotfound                  = "请重新登录"
	UserForcelogout               = "管理员强制退出，请重新登录"
	UserUnknownError              = "未知错误，请重新登录"

	// 文件上传消息（保留 Java 版原文，含 HTML 换行）
	UploadExceedMaxSize        = "上传的文件大小超出限制的文件大小！<br/>允许的文件最大大小是：%dMB！"
	UploadFilenameExceedLength = "上传的文件名最长%d个字符"

	// 权限（fmt.Sprintf 拼接权限标识）
	NoPermission       = "您没有数据的权限，请联系管理员添加权限 [%s]"
	NoCreatePermission = "您没有创建数据的权限，请联系管理员添加权限 [%s]"
	NoUpdatePermission = "您没有修改数据的权限，请联系管理员添加权限 [%s]"
	NoDeletePermission = "您没有删除数据的权限，请联系管理员添加权限 [%s]"
	NoExportPermission = "您没有导出数据的权限，请联系管理员添加权限 [%s]"
	NoViewPermission   = "您没有查看数据的权限，请联系管理员添加权限 [%s]"
)
