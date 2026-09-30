<?php
declare(strict_types=1);

namespace app\middleware;

use app\service\SessionService;
use Closure;
use RedisCache;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 防重复提交中间件（对位经典版 SameUrlDataInterceptor）
 *
 * 语义：同 URL + 同参数摘要 + 间隔 < 1s → 判重复，抛「不允许重复提交，请稍候再试」。
 * 经典版存 session，这里等价换 Redis 键（repeat_submit:<uuid>:<摘要>，TTL 1s）。
 */
class RepeatSubmit
{
    public function handle(Request $request, Closure $next): Response
    {
        if (strtoupper($request->method()) !== 'POST') {
            return $next($request);
        }

        $uuid = (string)($request->middleware('session_uuid') ?? 'anon');
        $params = $request->param();
        ksort($params);
        $digest = md5($request->pathinfo() . '|' . json_encode($params, JSON_UNESCAPED_UNICODE));
        $key = TpConstant::PREFIX_REPEAT_SUBMIT . $uuid . ':' . $digest;

        if (RedisCache::has($key)) {
            throw new \BusinessException('不允许重复提交，请稍候再试');
        }
        RedisCache::set($key, 1, TpConstant::REPEAT_SUBMIT_INTERVAL);
        return $next($request);
    }
}
