<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;

/**
 * 在线用户服务（对位 ISysUserOnlineService 语义；数据源 = Redis session:* SCAN，sys_user_online 表不读写）
 *
 * 会话解析口径：含非空 loginName 键 = TP 登录会话；
 * 无 loginName（匿名）跳过；JSON 解析失败 = 外来会话 → 字段显示「未知」且**禁止强退**（第一安全约束）。
 */
class OnlineService
{
    /** 扫描并解析全部在线会话（过滤匿名；外来会话跳过） */
    public static function scan(): array
    {
        $rows = [];
        foreach (RedisCache::keysScan(TpConstant::PREFIX_SESSION . '*') as $key) {
            $data = RedisCache::get($key, raw: true);
            if (!is_array($data) || !isset($data['loginName']) || $data['loginName'] === '') {
                continue; // 匿名会话/外来非法 JSON——列表不展示
            }
            $uuid = substr((string)$key, strlen(TpConstant::PREFIX_SESSION));
            $rows[] = self::toRow($uuid, $data);
        }
        return $rows;
    }

    /** 读取单个会话（强退前校验用）：返回行；匿名/非法 JSON 返回 ['unknown' => true] 供拒绝判断 */
    public static function resolve(string $uuid): array
    {
        $data = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid, raw: true);
        if (!is_array($data) || !isset($data['loginName']) || $data['loginName'] === '') {
            return ['unknown' => true];
        }
        return self::toRow($uuid, $data);
    }

    /** 强退（删除会话键）；键不存在返回 false */
    public static function forceLogout(string $uuid): bool
    {
        if (!RedisCache::has(TpConstant::PREFIX_SESSION . $uuid, raw: true)) {
            return false;
        }
        RedisCache::delete(TpConstant::PREFIX_SESSION . $uuid, raw: true);
        return true;
    }

    /**
     * 过滤（ipaddr/loginName like，PHP 侧）+ 排序 + 手工分页。
     * $orderBy 传驼峰列名（loginName/startTimestamp/lastAccessTime），白名单外不排序。
     */
    public static function filterPage(array $rows, array $filter, string $orderBy, string $isAsc, int $pageNum, int $pageSize): array
    {
        $ip = trim((string)($filter['ipaddr'] ?? ''));
        $name = trim((string)($filter['loginName'] ?? ''));
        if ($ip !== '') {
            $rows = array_values(array_filter($rows, fn($r) => str_contains((string)$r['ipaddr'], $ip)));
        }
        if ($name !== '') {
            $rows = array_values(array_filter($rows, fn($r) => str_contains((string)$r['loginName'], $name)));
        }
        $allow = ['loginName', 'startTimestamp', 'lastAccessTime'];
        if ($orderBy !== '' && in_array($orderBy, $allow, true)) {
            usort($rows, function ($a, $b) use ($orderBy, $isAsc) {
                $cmp = strcmp((string)$a[$orderBy], (string)$b[$orderBy]);
                return $isAsc === 'desc' ? -$cmp : $cmp;
            });
        }
        $total = count($rows);
        $pageRows = array_slice($rows, ($pageNum - 1) * $pageSize, $pageSize);
        return [$pageRows, $total];
    }

    /** 行 → TableDataInfo 驼峰（对位 SysUserOnlineOnlineVo） */
    private static function toRow(string $uuid, array $data): array
    {
        return [
            'sessionId'      => $uuid,
            'loginName'      => (string)($data['loginName'] ?? '未知'),
            'deptName'       => (string)($data['deptName'] ?? ($data['userName'] ?? '未知')),
            'ipaddr'         => (string)($data['ip'] ?? '未知'),
            'loginLocation'  => '内网',
            'browser'        => (string)($data['browser'] ?? '未知'),
            'os'             => (string)($data['os'] ?? '未知'),
            'status'         => 'on_line',
            'startTimestamp' => (string)($data['loginTime'] ?? '未知'),
            'lastAccessTime' => (string)($data['lastAccessTime'] ?? '未知'),
        ];
    }
}
