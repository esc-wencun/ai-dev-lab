"""
业务常量定义（对应java版ruoyi-common/constant包）
"""


class CacheConstants:
    """
    缓存的key前缀常量（对应java版CacheConstants）
    """
    # 登录用户 redis key
    LOGIN_TOKEN_KEY = "login_tokens:"
    # 验证码 redis key
    CAPTCHA_CODE_KEY = "captcha_codes:"
    # 参数管理 cache key
    SYS_CONFIG_KEY = "sys_config:"
    # 字典管理 cache key
    SYS_DICT_KEY = "sys_dict:"
    # 防重提交 redis key
    REPEAT_SUBMIT_KEY = "repeat_submit:"
    # 限流 redis key
    RATE_LIMIT_KEY = "rate_limit:"
    # 登录账户密码错误次数 redis key
    PWD_ERR_CNT_KEY = "pwd_err_cnt:"


class UserConstants:
    """
    用户常量信息（对应java版UserConstants）
    """
    # 是/否外链（java版：0是外链 1否）
    YES_FRAME = "0"
    NO_FRAME = "1"

    # 菜单类型（目录/菜单/按钮）
    TYPE_DIR = "M"
    TYPE_MENU = "C"
    TYPE_BUTTON = "F"

    # 组件布局常量
    LAYOUT = "Layout"
    PARENT_VIEW = "ParentView"
    INNER_LINK = "InnerLink"

    # 唯一/不唯一
    UNIQUE = True
    NOT_UNIQUE = False

    # 用户名长度限制
    USERNAME_MIN_LENGTH = 2
    USERNAME_MAX_LENGTH = 20
    # 密码长度限制
    PASSWORD_MIN_LENGTH = 5
    PASSWORD_MAX_LENGTH = 20

    # 正常状态
    NORMAL = "0"
    # 禁用状态
    DISABLE = "1"
    # 逻辑删除标志
    DEL_FLAG_DELETED = "2"


class Constants:
    """
    通用常量（对应java版Constants）
    """
    # 通用成功标识
    SUCCESS = "0"
    # 通用失败标识
    FAIL = "1"

    # 登录成功/失败/注销/注册（登录日志status）
    LOGIN_SUCCESS = "Success"
    LOGIN_FAIL = "Error"
    LOGOUT = "Logout"
    REGISTER = "Register"

    # 所有权限标识
    ALL_PERMISSION = "*:*:*"
    # 管理员角色权限标识
    SUPER_ADMIN = "admin"

    # 令牌
    TOKEN = "token"
    # 令牌前缀
    TOKEN_PREFIX = "Bearer "
    # JWT会话键（java版Constants.LOGIN_USER_KEY）
    LOGIN_USER_KEY = "login_user_key"
    # JWT用户名键（java版 Claims.SUBJECT）
    JWT_USERNAME = "sub"

    # 验证码有效期（分钟）
    CAPTCHA_EXPIRATION = 2

    # 资源映射路径前缀
    RESOURCE_PREFIX = "/profile"


class SysConfig:
    """
    系统配置参数键名（对应java版代码中使用的sys.account.*等config key）
    """
    # 验证码开关
    CAPTCHA_ENABLED = "sys.account.captchaEnabled"
    # 用户注册开关
    REGISTER_USER = "sys.account.registerUser"
    # 用户初始密码
    USER_INIT_PASSWORD = "sys.user.initPassword"
    # IP黑名单
    LOGIN_BLACK_IP_LIST = "sys.login.blackIPList"
    # 初始密码修改策略
    INIT_PASSWORD_MODIFY = "sys.account.initPasswordModify"
    # 密码更新周期
    PASSWORD_VALIDATE_DAYS = "sys.account.passwordValidateDays"
    # 密码字符范围
    ACCOUNT_CHRTYPE = "sys.account.chrtype"


class DataScope:
    """
    数据权限范围（对应java版Constants.Dept）
    """
    # 全部数据权限
    ALL = "1"
    # 自定数据权限
    CUSTOM = "2"
    # 部门数据权限
    DEPT = "3"
    # 部门及以下数据权限
    DEPT_AND_CHILD = "4"
    # 仅本人数据权限
    SELF = "5"


class HttpStatus:
    """
    返回状态码（对应java版HttpStatus，用于ResponseUtil）
    """
    SUCCESS = 200
    BAD_REQUEST = 400
    UNAUTHORIZED = 401
    FORBIDDEN = 403
    NOT_FOUND = 404
    ERROR = 500
    WARN = 601
