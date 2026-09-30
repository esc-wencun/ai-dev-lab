<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;

/**
 * 会话服务：cookie(uuid) + Redis 会话（对位经典版 Shiro Session）
 *
 * 30 分钟空闲超时；剩余 <20 分钟时续满（对位 RuoYi-Vue TokenService.verifyToken 语义）。
 * 会话 JSON 内容由登录链路写入：user_id/loginName/userName/deptId/avatar/isAdmin/permissions/roles/pwdUpdateDate/rememberMe
 * （注意 user_id 是下划线——消费方取值 `$session['user_id'] ?? $session['userId']`，勿凭本注释臆断键名）。
 */
final class SessionService
{
    public const COOKIE_NAME = 'tp_session';

    public static function create(array $sessionData): string
    {
        $uuid = bin2hex(random_bytes(16));
        self::write($uuid, $sessionData);
        return $uuid;
    }

    /** 读会话；命中即按续期语义刷新 TTL，返回 [uuid, data] 或 null */
    public static function touch(?string $uuid): ?array
    {
        if ($uuid === null || $uuid === '') {
            return null;
        }
        $data = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid);
        if ($data === null) {
            return null;
        }
        $ttl = RedisCache::ttl(TpConstant::PREFIX_SESSION . $uuid);
        if ($ttl !== -1 && $ttl < TpConstant::SESSION_RENEW_THRESHOLD) {
            RedisCache::expire(TpConstant::PREFIX_SESSION . $uuid, TpConstant::SESSION_IDLE_SECONDS);
        }
        return [$uuid, $data];
    }

    public static function write(string $uuid, array $sessionData): void
    {
        RedisCache::set(TpConstant::PREFIX_SESSION . $uuid, $sessionData, TpConstant::SESSION_IDLE_SECONDS);
    }

    public static function destroy(string $uuid): void
    {
        RedisCache::delete(TpConstant::PREFIX_SESSION . $uuid);
    }
}
