"""
防重复提交 + 限流（对应java版@RepeatSubmit + @RateLimiter）

用法：
    @router.post('/add', dependencies=[Depends(prevent_repeat_submit())])
    @router.get('/list', dependencies=[Depends(rate_limiter(count=10, time_seconds=60))])
"""
import json
import hashlib
from typing import Optional
from fastapi import Request
from common.constant import CacheConstants
from module_admin.service.login_service import get_redis_cache
from exceptions.exception import LoginException
from utils.response_util import ResponseUtil
from utils.log_util import logger

# 限流Lua脚本（对齐java版RedisConfig.limitScriptText）
RATE_LIMIT_LUA = """
local key = KEYS[1]
local count = tonumber(ARGV[1])
local time = tonumber(ARGV[2])
local current = redis.call('get', key);
if current and tonumber(current) > count then
    return tonumber(current);
end
current = redis.call('incr', key)
if tonumber(current) == 1 then
    redis.call('expire', key, time)
end
return tonumber(current);
"""


async def _get_request_digest(request: Request, body: Optional[bytes]) -> str:
    """
    请求指纹：url + token + body摘要
    """
    token = ''
    auth = request.headers.get('Authorization') or ''
    if auth.startswith('Bearer '):
        token = auth[7:]
    raw = f"{request.url.path}|{token}|{body.decode('utf-8', errors='replace') if body else ''}"
    return hashlib.md5(raw.encode('utf-8')).hexdigest()


def prevent_repeat_submit(interval_seconds: int = 5, message: str = '不允许重复提交，请稍候再试'):
    """
    防重复提交依赖（对应java @RepeatSubmit(interval=...)，java为毫秒，此处秒）
    interval窗口内同url+同token+同body视为重复
    """
    from common.constant import Constants

    async def _dependency(request: Request):
        if request.method not in ('POST', 'PUT', 'DELETE'):
            return
        body = await request.body()
        request.state.cached_body = body.decode('utf-8', errors='replace')
        digest = await _get_request_digest(request, body)
        cache = get_redis_cache(request)
        key = cache.build_key(CacheConstants.REPEAT_SUBMIT_KEY, digest)
        cache_raw = cache._redis
        # SET NX EX：首次成功，窗口内重复被拒
        ok = await cache_raw.set(key, '1', ex=interval_seconds, nx=True)
        if not ok:
            raise LoginException(message=message)
    return _dependency


def rate_limiter(count: int, time_seconds: int, limit_type: str = 'DEFAULT'):
    """
    限流依赖（对应java @RateLimiter(count, time, limitType)）
    :param count: 窗口内最大次数
    :param time_seconds: 窗口秒数
    :param limit_type: DEFAULT全局 / IP按客户端
    """
    async def _dependency(request: Request):
        if limit_type == 'IP':
            client = request.client.host if request.client else 'unknown'
            suffix = client
        else:
            suffix = 'global'
        cache = get_redis_cache(request)
        key = cache.build_key(CacheConstants.RATE_LIMIT_KEY,
                              request.url.path.replace('/', '_'), suffix)
        try:
            current = await cache._redis.eval(RATE_LIMIT_LUA, 1, key, count, time_seconds)
            if current and int(current) > count:
                raise LoginException(message='访问过于频繁，请稍候再试')
        except LoginException:
            raise
        except Exception as e:
            # redis故障不限流（对齐java侧容错策略）
            logger.warning(f'限流器异常（已放行）：{e}')
    return _dependency
