"""
invoke_target 解析与校验（spec-09 补齐，对齐java版JobInvokeUtil + SysJobController校验链）

目标语法与java一致：
- ryTask.ryNoParams                  （无参）
- ryTask.ryParams('ry')              （单字符串参）
- ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)
参数类型推断规则（对齐java getMethodParams）：
  'x'/"x" -> str；true/false -> bool；123L -> int(long)；1.5D -> float(double)；其他 -> int
"""
import re
from typing import List, Tuple, Any
from common.constant import Constants


def _smart_split(param_str: str) -> List[str]:
    """
    引号感知的逗号切分（等价java正则 split(",(?=...)")语义：引号内的逗号不切分）
    """
    parts, buf, quote = [], '', None
    for ch in param_str:
        if quote:
            buf += ch
            if ch == quote:
                quote = None
        elif ch in ('"', "'"):
            quote = ch
            buf += ch
        elif ch == ',':
            parts.append(buf)
            buf = ''
        else:
            buf += ch
    parts.append(buf)
    return parts


def parse_target(invoke_target: str) -> Tuple[str, str, List[Any]]:
    """
    解析目标字符串
    :return: (bean_name, method_name, params)；无参时 params=[]
    :raise ValueError: 语法不合法
    """
    target = (invoke_target or '').strip()
    if not target:
        raise ValueError('目标字符串为空')
    if '(' not in target or not target.endswith(')'):
        # 无参形式 xxx.yyy
        if not re.fullmatch(r'[A-Za-z_][\w.]*', target):
            raise ValueError(f'目标字符串语法不合法: {target}')
        bean, method = target.rsplit('.', 1)
        return bean, method, []

    bean_method = target[:target.rindex('(')]
    if '.' not in bean_method:
        raise ValueError(f'目标字符串缺少方法限定: {target}')
    bean, method = bean_method.rsplit('.', 1)
    param_str = target[target.rindex('(') + 1:target.rindex(')')].strip()

    params: List[Any] = []
    if param_str:
        for raw in _smart_split(param_str):
            params.append(_parse_param(raw.strip()))
    return bean, method, params


def _parse_param(s: str) -> Any:
    """
    单个参数字面量转python值（对齐java类型推断规则）
    """
    if not s:
        raise ValueError('存在空参数')
    # 字符串：'x' 或 "x"
    if (s.startswith("'") and s.endswith("'")) or (s.startswith('"') and s.endswith('"')):
        if len(s) < 2:
            raise ValueError(f'字符串参数不合法: {s}')
        return s[1:-1]
    low = s.lower()
    # 布尔
    if low == 'true':
        return True
    if low == 'false':
        return False
    # long（java L后缀）
    if s.endswith('L'):
        return int(s[:-1])
    # double（java D后缀）
    if s.endswith('D'):
        return float(s[:-1])
    # 整数
    if re.fullmatch(r'-?\d+', s):
        return int(s)
    # 浮点（宽松：316.50 这类）
    if re.fullmatch(r'-?\d+\.\d+', s):
        return float(s)
    raise ValueError(f'参数类型无法识别: {s}')


# ---- 校验链（对齐java SysJobController新增/修改时的invokeTarget校验） ----

# 违禁内容（对齐java Constants.JOB_ERROR_STR + LOOKUP_RMI/LDAP/HTTP）
_FORBIDDEN_SUBSTRINGS = [
    'rmi:', 'ldap:', 'ldaps:', 'http://', 'https://',
    'java.net.URL', 'javax.naming.InitialContext', 'org.yaml.snakeyaml',
    'org.springframework', 'org.apache', 'com.ruoyi.common.utils.file',
    'com.ruoyi.common.config', 'com.ruoyi.generator',
]


def validate_target(invoke_target: str, registered_checker) -> str:
    """
    完整校验链（顺序与文案对齐java SysJobController）
    :param registered_checker: callable(bean_name, method_name) -> bool，注册表检查
    :return: 错误文案；空串=通过
    """
    t = invoke_target or ''
    # java先做黑名单再解析
    for forbidden in _FORBIDDEN_SUBSTRINGS:
        if forbidden.lower() in t.lower():
            if forbidden in ('rmi:',):
                return f"目标字符串不允许'rmi'调用"
            if forbidden in ('ldap:', 'ldaps:'):
                return f"目标字符串不允许'ldap(s)'调用"
            if forbidden in ('http://', 'https://'):
                return f"目标字符串不允许'http(s)'调用"
            return '目标字符串存在违规'
    try:
        bean, method, _ = parse_target(t)
    except ValueError as e:
        return str(e)
    # 注册表检查（bean.method 全名），对齐java白名单语义但更严格
    if not registered_checker(f'{bean}.{method}'):
        return '目标字符串不在白名单内'
    return ''
