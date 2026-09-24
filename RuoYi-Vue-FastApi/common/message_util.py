"""
面向用户的提示文案管理（对应java版MessageUtils + messages.properties）

约定：面向用户的提示文案一律定义在此处，代码中不留裸中文字符串（日志内容除外）
占位符风格对齐java MessageFormat：{0} {1} ...
"""

MESSAGES = {
    # ---- 登录/账号（java messages.properties）----
    "not.null": "* 必须填写",
    "user.jcaptcha.error": "验证码错误",
    "user.jcaptcha.expire": "验证码已失效",
    "user.not.exists": "用户不存在/密码错误",
    "user.password.not.match": "用户不存在/密码错误",
    "user.password.retry.limit.count": "密码输入错误{0}次",
    "user.password.retry.limit.exceed": "密码输入错误{0}次，帐户锁定{1}分钟",
    "user.password.delete": "对不起，您的账号已被删除",
    "user.blocked": "用户已封禁，请联系管理员",
    "role.blocked": "角色已封禁，请联系管理员",
    "login.blocked": "很遗憾，访问IP已被列入系统黑名单",
    "user.logout.success": "退出成功",
    "user.login.success": "登录成功",
    "user.register.success": "注册成功",
    "user.notfound": "请重新登录",

    # ---- 注册（java SysRegisterService）----
    "register.username.empty": "用户名不能为空",
    "register.password.empty": "用户密码不能为空",
    "register.username.length": "账户长度必须在2到20个字符之间",
    "register.password.length": "密码长度必须在5到20个字符之间",
    "register.username.exists": "保存用户'{0}'失败，注册账号已存在",

    # ---- 权限（java messages.properties）----
    "no.permission": "您没有数据的权限，请联系管理员添加权限 [{0}]",
    "no.view.permission": "您没有查看数据的权限，请联系管理员添加权限 [{0}]",

    # ---- 密码策略（java SysProfileController等）----
    "user.password.old.error": "修改密码失败，旧密码错误",
    "user.password.same": "新密码不能与旧密码相同",
}


def message(key: str, *args) -> str:
    """
    获取文案并做占位符替换（对应java MessageUtils.message(key, args)）
    :param key: 文案键
    :param args: {0}{1}...占位参数
    """
    template = MESSAGES.get(key)
    if template is None:
        return key
    result = template
    for i, arg in enumerate(args):
        result = result.replace(f"{{{i}}}", str(arg))
    return result
