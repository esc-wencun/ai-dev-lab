"""
操作日志装饰器（对应java版@Log注解 + LogAspect）

用法：
    @log_decorator(title='用户管理', business_type=BusinessType.INSERT)
    async def add_user(request: Request, ...):
规范：所有写操作接口（POST/PUT/DELETE）必须挂此装饰器；查询接口不挂。
"""
import time
import json
import asyncio
from functools import wraps
from typing import Callable, Optional
from fastapi import Request
from common.enums import BusinessType, OperatorType
from common.constant import Constants
from config.database import AsyncSessionLocal
from module_admin.entity.do.entity import SysOperLog
from module_admin.service.login_service import TokenService, get_redis_cache
from utils.log_util import logger

# 敏感参数不记录（对齐java版LogAspect的PASSWORD字段过滤）
SENSITIVE_KEYS = {'password', 'oldPassword', 'newPassword', 'confirmPassword', 'code'}


def _sanitize(obj, depth=0):
    """
    递归脱敏：敏感键替换为*（对齐java版正则替换password字段的语义）
    """
    if depth > 5:
        return obj
    if isinstance(obj, dict):
        return {k: ('*' if k in SENSITIVE_KEYS else _sanitize(v, depth + 1)) for k, v in obj.items()}
    if isinstance(obj, list):
        return [_sanitize(i, depth + 1) for i in obj]
    return obj


def _truncate(text: Optional[str], limit: int = 2000) -> str:
    """
    截断（对齐java版oper_param/json_result varchar(2000)）
    """
    if not text:
        return ''
    return text[:limit]


async def _build_oper_log(request: Request, title: str, business_type: BusinessType,
                          cost_ms: int, error_msg: str = '', is_error: bool = False):
    """
    组装操作日志实体（对应java版LogAspect记录的各字段）
    """
    login_user = await TokenService.get_login_user(request)
    oper_name = login_user.get('user_name') if login_user else ''

    # 请求参数：优先取body，GET参数兜底
    oper_param = ''
    try:
        body = getattr(request.state, 'cached_body', None)
        if body:
            try:
                parsed = json.loads(body)
                oper_param = json.dumps(_sanitize(parsed), ensure_ascii=False)
            except (ValueError, TypeError):
                oper_param = str(body)
        elif request.query_params:
            oper_param = str(dict(request.query_params))
    except Exception:
        oper_param = ''

    return SysOperLog(
        title=title,
        business_type=business_type.code,
        method=request.url.path,
        request_method=request.method,
        operator_type=OperatorType.MANAGE.code,
        oper_name=oper_name,
        oper_url=str(request.url.path),
        oper_ip=request.client.host if request.client else '',
        oper_param=_truncate(oper_param),
        json_result='' if is_error else '成功',
        status=1 if is_error else 0,
        error_msg=_truncate(error_msg),
        oper_time=__import__('datetime').datetime.now(),
        cost_time=cost_ms,
    )


def log_decorator(title: str, business_type: BusinessType = BusinessType.OTHER):
    """
    操作日志装饰器工厂（对应java版@Log(title, businessType)）

    请求body在装饰器内读取后缓存到request.state（供后续handler重新解析），
    记录动作在独立session中异步执行，不占用请求的事务。
    """

    def decorator(func: Callable):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            request: Optional[Request] = None
            for a in args:
                if isinstance(a, Request):
                    request = a
                    break
            if request is None:
                for v in kwargs.values():
                    if isinstance(v, Request):
                        request = v
                        break

            # 预读body并缓存（FastAPI的body只能读一次）
            # multipart上传（含文件）跳过body预读，避免流被消费导致"Stream consumed"
            if request is not None and request.method in ('POST', 'PUT', 'DELETE'):
                content_type = request.headers.get('content-type', '')
                if 'multipart/form-data' not in content_type:
                    body = await request.body()
                    request.state.cached_body = body.decode('utf-8', errors='replace')

            start = time.perf_counter()
            try:
                response = await func(*args, **kwargs)
                cost_ms = int((time.perf_counter() - start) * 1000)
                if request is not None:
                    async def _save():
                        async with AsyncSessionLocal() as session:
                            session.add(await _build_oper_log(request, title, business_type, cost_ms))
                            await session.commit()
                    asyncio.create_task(_guard(_save()))
                return response
            except Exception as e:
                cost_ms = int((time.perf_counter() - start) * 1000)
                if request is not None:
                    async def _save_err():
                        async with AsyncSessionLocal() as session:
                            session.add(await _build_oper_log(request, title, business_type,
                                                              cost_ms, str(e), is_error=True))
                            await session.commit()
                    asyncio.create_task(_guard(_save_err()))
                raise

        return wrapper

    async def _guard(coro):
        try:
            await coro
        except Exception as e:
            logger.error(f'操作日志写入失败：{e}')

    return decorator
