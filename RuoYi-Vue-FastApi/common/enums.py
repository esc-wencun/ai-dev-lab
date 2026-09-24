"""
业务枚举定义（对应java版ruoyi-common/enums包）
Python枚举范式：带字段枚举用元组成员 + __init__，一比一复刻java写法
"""
from enum import Enum


class UserStatus(Enum):
    """
    用户状态（对应java版UserStatus）
    """
    OK = ("0", "正常")
    DISABLE = ("1", "停用")
    DELETED = ("2", "已删除")

    def __init__(self, code, info):
        self.code = code
        self.info = info


class BusinessType(Enum):
    """
    业务操作类型（对应java版BusinessType）
    code即java @Log注解business_type落库数字（java枚举ordinal 0-9），勿调整对应关系
    注意：python枚举没有java的ordinal()方法，因此显式携带code字段
    """
    OTHER = (0, "其它")
    INSERT = (1, "新增")
    UPDATE = (2, "修改")
    DELETE = (3, "删除")
    GRANT = (4, "授权")
    EXPORT = (5, "导出")
    IMPORT = (6, "导入")
    FORCE = (7, "强退")
    GENCODE = (8, "生成代码")
    CLEAN = (9, "清空数据")

    def __init__(self, code, info):
        self.code = code
        self.info = info


class OperatorType(Enum):
    """
    操作人类别（对应java版OperatorType；java中OTHER=0, MANAGE=1, MOBILE=2，用显式value对齐）
    """
    OTHER = (0, "其它")
    MANAGE = (1, "后台用户")
    MOBILE = (2, "手机端用户")

    def __init__(self, code, info):
        self.code = code
        self.info = info


class DataSourceType(Enum):
    """
    数据源（对应java版DataSource）
    """
    MASTER = ("master", "主库")
    SLAVE = ("slave", "从库")

    def __init__(self, name_, info):
        self.name_ = name_
        self.info = info


class HttpMethod(Enum):
    """
    http方法（对应java版HttpMethod）
    """
    GET = ("GET", "查")
    POST = ("POST", "增")
    PUT = ("PUT", "改")
    DELETE = ("DELETE", "删")

    def __init__(self, method, action):
        self.method = method
        self.action = action
