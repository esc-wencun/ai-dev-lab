<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Perm;

use app\service\MonitorCacheService;
use AjaxResult;
use RedisCache;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 缓存监控控制器（对位经典版 CacheController 7 路由；getNames/getKeys/getValue 返回 HTML 片段非 JSON）
 */
class CacheController extends \app\BaseController
{
    /** GET /monitor/cache：三栏主页 */
    #[Perm('monitor:cache:view')]
    public function index(Request $request): Response
    {
        return Response::create('cache/cache', 'view')->assign([
            'cacheNames' => array_map(fn($name) => ['name' => $name], array_keys(MonitorCacheService::prefixes())),
        ]);
    }

    /** POST /monitor/cache/getNames：HTML 片段（缓存名 tr 列表） */
    #[Perm('monitor:cache:view')]
    public function getNames(Request $request): Response
    {
        return Response::create('cache/fragment_names', 'view')->assign([
            'cacheNames' => array_map(fn($name) => ['name' => $name], array_keys(MonitorCacheService::prefixes())),
        ]);
    }

    /** POST /monitor/cache/getKeys：HTML 片段（键列表；cacheName 空 = 全库已知前缀合并） */
    #[Perm('monitor:cache:view')]
    public function getKeys(Request $request): Response
    {
        $cacheName = (string)$request->post('cacheName', '');
        $prefix = $cacheName !== '' ? MonitorCacheService::prefixOf($cacheName) : null;
        if ($cacheName !== '' && $prefix === null) {
            $keys = [];
        } elseif ($prefix !== null) {
            $keys = RedisCache::keysScan($prefix . '*');
            sort($keys);
        } else {
            $keys = MonitorCacheService::allKnownKeys();
        }
        return Response::create('cache/fragment_keys', 'view')->assign(['cacheKeys' => $keys]);
    }

    /** POST /monitor/cache/getValue：HTML 片段（键值 pretty print） */
    #[Perm('monitor:cache:view')]
    public function getValue(Request $request): Response
    {
        $cacheName = (string)$request->post('cacheName', '');
        $cacheKey = (string)$request->post('cacheKey', '');
        return Response::create('cache/fragment_value', 'view')->assign([
            'cacheName'  => $cacheName,
            'cacheKey'   => $cacheKey,
            'cacheValue' => MonitorCacheService::prettyValue($cacheKey) ?? '（不存在或已过期）',
        ]);
    }

    /** POST /monitor/cache/clearCacheName：清某前缀全部键（JSON） */
    #[Perm('monitor:cache:view')]
    public function clearCacheName(Request $request): Response
    {
        $prefix = MonitorCacheService::prefixOf((string)$request->post('cacheName', ''));
        if ($prefix === null) {
            return AjaxResult::error('未知缓存名');
        }
        if ($prefix === TpConstant::PREFIX_SESSION) {
            return AjaxResult::error('会话缓存不允许批量清理');
        }
        MonitorCacheService::clearPrefix($prefix);
        return AjaxResult::success();
    }

    /** POST /monitor/cache/clearCacheKey：删单个完整键（JSON） */
    #[Perm('monitor:cache:view')]
    public function clearCacheKey(Request $request): Response
    {
        $cacheKey = (string)$request->post('cacheKey', '');
        if ($cacheKey === '' || str_contains($cacheKey, '..')) {
            return AjaxResult::error('参数错误');
        }
        RedisCache::delete($cacheKey, raw: true);
        return AjaxResult::success();
    }

    /** GET /monitor/cache/clearAll：清全部业务缓存（session: 排除；JSON） */
    #[Perm('monitor:cache:view')]
    public function clearAll(Request $request): Response
    {
        MonitorCacheService::clearAll();
        return AjaxResult::success();
    }
}
