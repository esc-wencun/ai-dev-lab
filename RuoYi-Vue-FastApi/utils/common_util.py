"""
序列化工具（对齐java版Jackson驼峰序列化 + 参考项目CamelCaseUtil）

规范：所有模块列表/详情返回一律 ORM -> CamelCaseUtil.transform_result，禁止手写字段映射
敏感字段（password）永不序列化，对齐java版SysUser @JsonProperty(WRITE_ONLY)
"""
from datetime import datetime, date
from typing import Any, List

# 永不序列化的敏感字段
EXCLUDED_COLUMNS = {'password'}


def snake_to_camel(name: str) -> str:
    """
    下划线转驼峰：user_name -> userName
    """
    if '_' not in name:
        return name
    head, *rest = name.split('_')
    return head + ''.join(word.title() for word in rest)


def _clean_java_cache(value: str) -> str:
    """
    清理java FastJson写入redis的缓存值中的java特有语法：
    - @type 注解（"@type":"com.ruoyi..."）
    - long字面量 123L
    仅对疑似java格式的文本做保守修正，失败返回原文本
    """
    import re
    if not isinstance(value, str) or '@type' not in value:
        return value
    cleaned = re.sub(r'\s*"\$\??"?@type"\s*:\s*"[^"]*"\s*,?', lambda m: '' if m.group(0).rstrip().endswith(',') else '', value)
    cleaned = re.sub(r'"@type":"[^"]*",?', '', cleaned)
    cleaned = re.sub(r'(\d+)L', r'\1', cleaned)
    cleaned = cleaned.replace('true', 'true')  # no-op 占位，保持可读
    return cleaned


def _parse_cache_text(value: str):
    """
    解析缓存文本为python对象；兼容java FastJson格式（含@type/1L）
    """
    import json as _json
    if not isinstance(value, str):
        return value
    try:
        return _json.loads(value)
    except (ValueError, TypeError):
        pass
    try:
        return _json.loads(_clean_java_cache(value))
    except (ValueError, TypeError):
        return None  # 无法解析（java专有格式），由调用方回源


def _format_value(value: Any) -> Any:
    """
    值转换：datetime -> 'yyyy-MM-dd HH:mm:ss'，date -> 'yyyy-MM-dd'
    """
    if isinstance(value, datetime):
        return value.strftime('%Y-%m-%d %H:%M:%S')
    if isinstance(value, date):
        return value.strftime('%Y-%m-%d')
    return value


def _transform_row(row) -> Any:
    """
    转换单行：SQLAlchemy模型对象 / Row / dict / 标量
    """
    # SQLAlchemy ORM 模型对象
    if hasattr(row, '__table__'):
        result = {}
        for column in row.__table__.columns:
            if column.name in EXCLUDED_COLUMNS:
                continue
            result[snake_to_camel(column.name)] = _format_value(getattr(row, column.name))
        # 附带瞬时属性（如菜单树构建时的children）
        for key in ('children',):
            if hasattr(row, key):
                result[key] = transform_result(getattr(row, key))
        return result

    # sqlalchemy Row（多表join查询结果：命名元组式）
    if hasattr(row, '_mapping'):
        result = {}
        for key, value in row._mapping.items():
            # key可能是表名.列名（标签形式）或纯列名
            short_key = key.split('.')[-1] if isinstance(key, str) and '.' in key else key
            camel_key = snake_to_camel(short_key) if isinstance(short_key, str) else short_key
            result[camel_key] = transform_result(value) if hasattr(value, '__table__') or hasattr(value, '_mapping') else _format_value(value)
        return result

    # dict
    if isinstance(row, dict):
        return {snake_to_camel(k): _format_value(v) for k, v in row.items()}

    # 标量
    return _format_value(row)


def transform_result(data: Any) -> Any:
    """
    将SQLAlchemy查询结果转换为驼峰命名的dict列表/对象（对应java Jackson默认驼峰输出）
    :param data: ORM对象/Row/dict/标量 或 其列表
    :return: 驼峰dict（列表输入返回列表，单对象输入返回dict）
    """
    if data is None:
        return None
    if isinstance(data, (list, tuple)):
        return [_transform_row(row) for row in data]
    return _transform_row(data)
