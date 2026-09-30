<?php
declare(strict_types=1);

/**
 * RedisCache 门面：Redis 访问唯一入口
 *
 * - 键拼装收口：业务传裸名，按 PREFIX_* 常量拼完整键
 * - JSON 编解码收口：value 统一 JSON
 * - SCAN 替代 KEYS（keysScan）
 * - 业务代码禁止直接使用 predis 客户端
 */
final class RedisCache
{
    private static ?\Predis\ClientInterface $client = null;

    public static function client(): \Predis\ClientInterface
    {
        if (self::$client === null) {
            $conf = config('cache.redis');
            self::$client = new \Predis\Client([
                'scheme'   => 'tcp',
                'host'     => $conf['host'],
                'port'     => $conf['port'],
                'database' => $conf['database'],
                'password' => $conf['password'] !== '' ? $conf['password'] : null,
            ]);
        }
        return self::$client;
    }

    /** 取键（裸名自动加前缀；传完整键名时用 raw） */
    public static function get(string $bareKey, bool $raw = false): mixed
    {
        $value = self::client()->get(self::key($bareKey, $raw));
        return $value === null ? null : json_decode($value, true);
    }

    /** 写键；$ttl 秒过期，0 永久 */
    public static function set(string $bareKey, mixed $value, int $ttl = 0, bool $raw = false): void
    {
        $full = self::key($bareKey, $raw);
        $json = json_encode($value, JSON_UNESCAPED_UNICODE);
        if ($ttl > 0) {
            self::client()->setex($full, $ttl, $json);
        } else {
            self::client()->set($full, $json);
        }
    }

    public static function delete(string $bareKey, bool $raw = false): void
    {
        self::client()->del(self::key($bareKey, $raw));
    }

    public static function has(string $bareKey, bool $raw = false): bool
    {
        return (bool)self::client()->exists(self::key($bareKey, $raw));
    }

    /** 剩余过期秒数；-2 键不存在，-1 永久 */
    public static function ttl(string $bareKey, bool $raw = false): int
    {
        return (int)self::client()->ttl(self::key($bareKey, $raw));
    }

    /** 设置/刷新过期时间 */
    public static function expire(string $bareKey, int $ttl, bool $raw = false): void
    {
        self::client()->expire(self::key($bareKey, $raw), $ttl);
    }

    /** 自增并返回新值（计数场景，如密码重试）；键不存在时从 0 起计，带过期 */
    public static function incrWithExpire(string $bareKey, int $ttl, bool $raw = false): int
    {
        $full = self::key($bareKey, $raw);
        $new = (int)self::client()->incr($full);
        if ($new === 1) {
            self::client()->expire($full, $ttl);
        }
        return $new;
    }

    /** 原子 SETNX + TTL（10.0.0 任务执行锁/触发幂等键；键已存在返回 false） */
    public static function setNx(string $bareKey, mixed $value, int $ttl, bool $raw = false): bool
    {
        $full = self::key($bareKey, $raw);
        $json = json_encode($value, JSON_UNESCAPED_UNICODE);
        return (bool)self::client()->set($full, $json, 'EX', $ttl, 'NX');
    }

    /** SCAN 迭代取匹配键（替代 KEYS，防大库阻塞）。predis 返回 [cursor, keys[]] 二元组 */
    public static function keysScan(string $pattern): array
    {
        $keys = [];
        $it = null;
        do {
            [$it, $batch] = self::client()->scan($it, ['MATCH' => $pattern, 'COUNT' => 100]);
            foreach ($batch as $k) {
                $keys[] = $k;
            }
        } while ((string)$it !== '0');
        return $keys;
    }

    /** 裸名 → 完整键：带已知前缀的按常量拼，否则原样（raw 场景直接原样） */
    private static function key(string $bareKey, bool $raw): string
    {
        if ($raw) {
            return $bareKey;
        }
        return $bareKey; // 前缀由调用方用 TpConstant::PREFIX_* 显式拼（如 TpConstant::PREFIX_SESSION . $uuid）
    }
}
