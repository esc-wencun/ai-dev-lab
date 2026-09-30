<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;

/**
 * 缓存监控服务（对位经典版 CacheService；对象从 ehcache 七缓存名换成 Redis 六前缀枚举）
 *
 * 验证码无独立键（存会话 captcha 字段）——枚举无 captcha 项。
 * clearAll 排除 session: 前缀（防全站踢线——经典版 SYS_AUTH_CACHE 排除先例）。
 */
class MonitorCacheService
{
    /** 前缀注册表（键=展示名，值=TpConstant 前缀常量；新增前缀只改这里 + TpConstant） */
    public static function prefixes(): array
    {
        return [
            'session'       => TpConstant::PREFIX_SESSION,
            'pwd_retry'     => TpConstant::PREFIX_PWD_RETRY,
            'repeat_submit' => TpConstant::PREFIX_REPEAT_SUBMIT,
            'rate_limit'    => TpConstant::PREFIX_RATE_LIMIT,
            'config'        => TpConstant::PREFIX_CONFIG,
            'dict'          => TpConstant::PREFIX_DICT,
        ];
    }

    /** 某展示名对应的前缀；未知名返回 null */
    public static function prefixOf(string $name): ?string
    {
        return self::prefixes()[$name] ?? null;
    }

    /** 全库已知前缀的键合并（cacheName 空时对位经典版 getCacheKeys('') 刷新钮） */
    public static function allKnownKeys(): array
    {
        $keys = [];
        foreach (self::prefixes() as $prefix) {
            foreach (RedisCache::keysScan($prefix . '*') as $k) {
                $keys[] = $k;
            }
        }
        sort($keys);
        return $keys;
    }

    /** 键值 pretty print（JSON 解码美化；键不存在 → null） */
    public static function prettyValue(string $fullKey): ?string
    {
        $raw = RedisCache::client()->get($fullKey);
        if ($raw === null) {
            return null;
        }
        $decoded = json_decode($raw, true);
        return $decoded === null ? (string)$raw : json_encode($decoded, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
    }

    /** 清空某前缀全部键 */
    public static function clearPrefix(string $prefix): int
    {
        $n = 0;
        foreach (RedisCache::keysScan($prefix . '*') as $k) {
            RedisCache::delete($k, raw: true);
            $n++;
        }
        return $n;
    }

    /** 清空全部业务缓存（session: 排除） */
    public static function clearAll(): int
    {
        $n = 0;
        foreach (self::prefixes() as $name => $prefix) {
            if ($name === 'session') {
                continue;
            }
            $n += self::clearPrefix($prefix);
        }
        return $n;
    }
}
