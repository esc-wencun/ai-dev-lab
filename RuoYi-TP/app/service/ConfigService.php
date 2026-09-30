<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;
use think\facade\Db;

/**
 * sys_config 参数服务（对位经典版 ConfigService/ISysConfigService）
 *
 * 读表 + Redis `config:` 缓存；6.0.0 参数管理改值时调 refresh 清缓存。
 */
final class ConfigService
{
    /** 取配置值；不存在返回 $default（对位经典版 selectConfigByKey 语义） */
    public static function get(string $configKey, string $default = ''): string
    {
        $cacheKey = TpConstant::PREFIX_CONFIG . $configKey;
        $cached = RedisCache::get($cacheKey);
        if ($cached !== null) {
            return (string)$cached;
        }
        $row = Db::table('sys_config')->where('config_key', $configKey)->field('config_value')->find();
        $value = $row['config_value'] ?? $default;
        RedisCache::set($cacheKey, $value, 0);
        return (string)$value;
    }

    public static function getBool(string $configKey, bool $default = false): bool
    {
        $v = self::get($configKey, $default ? 'true' : 'false');
        return strtolower($v) === 'true' || $v === '1';
    }

    /** 清单条或全部配置缓存（参数管理模块修改后调用） */
    public static function refresh(string $configKey = ''): void
    {
        if ($configKey !== '') {
            RedisCache::delete(TpConstant::PREFIX_CONFIG . $configKey);
            return;
        }
        foreach (RedisCache::keysScan(TpConstant::PREFIX_CONFIG . '*') as $key) {
            RedisCache::delete($key, true);
        }
    }

    /* ================= 6.0.0 收编：管理端联动 ================= */

    /** 单键写缓存（对位 CacheUtils.put；参数新增/修改后调用） */
    public static function set(string $configKey, string $value): void
    {
        RedisCache::set(TpConstant::PREFIX_CONFIG . $configKey, $value, 0);
    }

    /** 全量预热（对位 resetConfigCache 语义：全表 sys_config 逐行回填） */
    public static function loadAll(): void
    {
        $rows = Db::table('sys_config')->field('config_key,config_value')->select()->toArray();
        foreach ($rows as $row) {
            RedisCache::set(TpConstant::PREFIX_CONFIG . $row['config_key'], $row['config_value'], 0);
        }
    }

    /** 清全部 + 全量预热（对位 resetConfigCache；refreshCache 端点消费） */
    public static function reset(): void
    {
        self::refresh();
        self::loadAll();
    }
}
