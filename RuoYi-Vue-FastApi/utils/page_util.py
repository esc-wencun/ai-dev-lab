"""
分页工具（对应java版PageHelper startPage + BaseController.getDataTable）

规范：所有列表接口经 paginate 返回 {code, msg, rows, total}，与java版TableDataInfo一致
"""
import math
from typing import Any, Optional
from fastapi import Request
from sqlalchemy import Select, select, func, text
from sqlalchemy.ext.asyncio import AsyncSession
from utils.response_util import ResponseUtil
from utils.common_util import transform_result


# 允许排序的字符白名单校验（防SQL注入，字段名经驼峰转下划线后需匹配此模式）
def _safe_order_column(column: str) -> Optional[str]:
    """
    驼峰转下划线并校验合法性：userId -> user_id；非法字符返回None
    """
    if not column:
        return None
    # 仅允许字母数字，拒绝一切特殊符号（含括号、引号、空格）
    import re
    if not re.fullmatch(r'[A-Za-z0-9_]+', column):
        return None
    result = []
    for ch in column:
        if ch.isupper():
            result.append('_')
            result.append(ch.lower())
        else:
            result.append(ch)
    return ''.join(result)


class PageDomain:
    """
    分页参数（对应java版TableSupport.buildPageRequest）
    """

    def __init__(self, page_num: int = 1, page_size: int = 10,
                 order_by_column: Optional[str] = None, is_asc: Optional[str] = None):
        self.page_num = page_num
        self.page_size = page_size
        self.order_by_column = order_by_column
        self.is_asc = is_asc

    @property
    def offset(self) -> int:
        return (self.page_num - 1) * self.page_size

    @property
    def order_clause(self) -> Optional[str]:
        """
        生成排序子串（已校验安全），如 'user_id asc'；无排序参数返回None
        """
        col = _safe_order_column(self.order_by_column or '')
        if not col:
            return None
        direction = 'desc' if (self.is_asc or '').lower() == 'desc' else 'asc'
        return f'{col} {direction}'


def get_page_domain(request: Request) -> PageDomain:
    """
    从query params解析分页参数（对应java版TableSupport：pageNum/pageSize/orderByColumn/isAsc）
    """
    params = request.query_params
    try:
        page_num = max(1, int(params.get('pageNum', 1)))
    except ValueError:
        page_num = 1
    try:
        page_size = min(500, max(1, int(params.get('pageSize', 10))))
    except ValueError:
        page_size = 10
    return PageDomain(
        page_num=page_num,
        page_size=page_size,
        order_by_column=params.get('orderByColumn'),
        is_asc=params.get('isAsc'),
    )


async def paginate(db: AsyncSession, query: Select, request: Request,
                   transformer=transform_result):
    """
    分页查询并组装TableDataInfo响应（对应java版startPage + getDataTable）
    :param db: orm会话
    :param query: SQLAlchemy select语句（不含limit）
    :param request: 用于读取分页参数
    :param transformer: 行转换器（默认驼峰转换）
    :return: ResponseUtil.success({rows, total})格式的JSONResponse
    """
    domain = get_page_domain(request)

    # 总数
    total = (await db.execute(
        select(func.count('*')).select_from(query.order_by(None).subquery())
    )).scalar() or 0

    # 排序（column已过白名单正则校验，text()安全）
    ordered = query
    clause = domain.order_clause
    if clause:
        ordered = query.order_by(None).order_by(text(clause))

    # 分页执行
    rows = (await db.execute(
        ordered.offset(domain.offset).limit(domain.page_size)
    )).all()

    # 单列（纯ORM实体）解包，多列（join行）保留Row交由transformer处理
    data = []
    for row in rows:
        data.append(row[0] if row and len(row) == 1 else row)

    return ResponseUtil.success(
        msg='查询成功',
        dict_content={'rows': transformer(data), 'total': int(total)}
    )
