"""
Redis缓存门面（对应java版RedisCache + RedisTemplate）

规范：业务代码禁止直接 request.app.state.redis，一律通过本类读写；
key前缀只允许来自 common.constant.CacheConstants；序列化统一JSON。
"""
import json
from typing import Any, Optional, List
from redis.asyncio import Redis
from common.constant import CacheConstants


class RedisCache:
    """
    Redis缓存操作封装
    """

    def __init__(self, redis: Redis):
        self._redis = redis

    # ---- key 构建 ----

    @staticmethod
    def build_key(prefix: str, *parts) -> str:
        """
        统一key拼装：prefix + part1:part2...
        :param prefix: 前缀，必须来自CacheConstants（以冒号结尾）
        :param parts: 键的组成部分
        """
        return prefix + ":".join(str(p) for p in parts)

    # ---- 纯字符串缓存（验证码答案等简单值，不做JSON包装）----

    async def set_string(self, key: str, value: str, expire_minutes: Optional[int] = None):
        """
        写入原字符串（对应java redisTemplate.opsForValue().set的字符串值语义）
        """
        if expire_minutes is not None:
            await self._redis.set(key, value, ex=expire_minutes * 60)
        else:
            await self._redis.set(key, value)

    async def get_string(self, key: str) -> Optional[str]:
        """
        读取原字符串，不存在返回None
        """
        value = await self._redis.get(key)
        return value

    # ---- 对象缓存（JSON序列化，分钟级TTL对齐java习惯）----

    async def set_cache_object(self, key: str, value: Any, expire_minutes: Optional[int] = None):
        """
        写入对象缓存
        :param key: 完整key（建议经build_key构建）
        :param value: 任意可JSON序列化对象
        :param expire_minutes: 有效期（分钟），None表示不过期
        """
        payload = json.dumps(value, ensure_ascii=False, default=str)
        if expire_minutes is not None:
            await self._redis.set(key, payload, ex=expire_minutes * 60)
        else:
            await self._redis.set(key, payload)

    async def get_cache_object(self, key: str) -> Optional[Any]:
        """
        读取对象缓存，不存在返回None
        兼容处理：java FastJson格式（@type/1L）尝试保守修正解析；
        修正后仍无法解析则**原样返回文本**（不做删除——共享redis下可能破坏java侧会话/缓存）
        """
        from utils.common_util import _parse_cache_text
        value = await self._redis.get(key)
        if value is None:
            return None
        parsed = _parse_cache_text(value)
        if parsed is None:
            return value  # 原样返回文本，不删除
        return parsed

    async def delete_object(self, *keys: str):
        """
        删除一个或多个key
        """
        if keys:
            await self._redis.delete(*keys)

    async def expire(self, key: str, minutes: int):
        """
        设置key的有效期（分钟）
        """
        await self._redis.expire(key, minutes * 60)

    async def has_key(self, key: str) -> bool:
        """
        判断key是否存在
        """
        return bool(await self._redis.exists(key))

    # ---- 计数器（密码错误次数等）----

    async def increment(self, key: str, expire_minutes: Optional[int] = None) -> int:
        """
        自增计数器，可附带有效期（秒级精度由redis保证）
        """
        count = await self._redis.incr(key)
        if expire_minutes is not None and count == 1:
            await self._redis.expire(key, expire_minutes * 60)
        return count

    async def get_counter(self, key: str) -> int:
        """
        读取计数器，不存在返回0
        """
        value = await self._redis.get(key)
        return int(value) if value else 0

    # ---- 键扫描（生产共享redis禁用KEYS，统一SCAN）----

    async def keys_by_prefix(self, prefix: str) -> List[str]:
        """
        按前缀扫描全部key（SCAN迭代，非阻塞，供在线用户/缓存监控使用）
        """
        pattern = prefix if prefix.endswith('*') else prefix + '*'
        result = []
        async for key in self._redis.scan_iter(match=pattern, count=200):
            result.append(key)
        return result

    # ---- 登录会话专用（收编Phase 0散落在TokenService的序列化逻辑）----

    async def save_login_user(self, session_uuid: str, login_user: dict, expire_minutes: int):
        """
        保存登录会话（key前缀CacheConstants.LOGIN_TOKEN_KEY）
        """
        key = self.build_key(CacheConstants.LOGIN_TOKEN_KEY, session_uuid)
        await self.set_cache_object(key, login_user, expire_minutes)

    async def get_login_user(self, session_uuid: str) -> Optional[dict]:
        """
        读取登录会话
        """
        key = self.build_key(CacheConstants.LOGIN_TOKEN_KEY, session_uuid)
        data = await self.get_cache_object(key)
        return data if isinstance(data, dict) else None

    async def delete_login_user(self, session_uuid: str):
        """
        删除登录会话
        """
        key = self.build_key(CacheConstants.LOGIN_TOKEN_KEY, session_uuid)
        await self.delete_object(key)
