<?php
declare(strict_types=1);

namespace app\service;

use think\facade\Db;

/**
 * 公告已读记录服务（对位经典版 SysNoticeReadServiceImpl + SysNoticeReadMapper.xml）
 *
 * sys_notice_read 有 uk_user_notice 唯一键（user_id+notice_id）——
 * 已读写入一律 INSERT IGNORE 防重复（重复 markRead 不报错不重复插）。
 */
class NoticeReadService
{
    /** 标记已读（insert ignore 幂等） */
    public static function markRead(int $noticeId, int $userId): void
    {
        Db::execute(
            'INSERT IGNORE INTO sys_notice_read (notice_id, user_id, read_time) VALUES (?, ?, ?)',
            [$noticeId, $userId, date('Y-m-d H:i:s')]
        );
    }

    /** 批量标记已读（空数组短路 + 批量 insert ignore；经典版原样） */
    public static function markReadBatch(int $userId, array $noticeIds): void
    {
        if (!$noticeIds) {
            return;
        }
        $now = date('Y-m-d H:i:s');
        $values = [];
        $bindings = [];
        foreach ($noticeIds as $nid) {
            $values[] = '(?, ?, ?)';
            array_push($bindings, (int)$nid, $userId, $now);
        }
        Db::execute(
            'INSERT IGNORE INTO sys_notice_read (notice_id, user_id, read_time) VALUES ' . implode(',', $values),
            $bindings
        );
    }

    /** 删除公告时清理对应已读记录（controller 两连删的第一删） */
    public static function deleteByNoticeIds(array $ids): int
    {
        if (!$ids) {
            return 0;
        }
        return Db::table('sys_notice_read')->where('notice_id', 'in', $ids)->delete();
    }

    /**
     * 带已读状态的公告列表（listTop 数据源；对位 selectNoticeListWithReadStatus）。
     * status='0' order by notice_id desc limit N；行输出驼峰 + isRead 布尔（主框架消费 JS 契约键）。
     */
    public static function selectNoticeListWithReadStatus(int $userId, int $limit = 5): array
    {
        $rows = Db::query(
            'SELECT n.notice_id AS noticeId, n.notice_title AS noticeTitle, n.notice_type AS noticeType,'
            . ' n.status, n.create_by AS createBy, n.create_time AS createTime,'
            . ' CASE WHEN r.notice_id IS NOT NULL THEN TRUE ELSE FALSE END AS isRead'
            . ' FROM sys_notice n'
            . ' LEFT JOIN sys_notice_read r ON r.notice_id = n.notice_id AND r.user_id = ?'
            . ' WHERE n.status = ?'
            . ' ORDER BY n.notice_id DESC'
            . ' LIMIT ' . $limit,
            [$userId, '0']
        );
        return array_map(fn($r) => array_merge((array)$r, ['isRead' => (bool)$r['isRead']]), $rows);
    }

    /** 已阅读某公告的用户列表（三表 JOIN + searchValue 双列 like；行 6 驼峰列；read_time desc 固定） */
    public static function selectReadUsersByNoticeId(int $noticeId, string $searchValue = ''): array
    {
        $sql = 'SELECT u.user_id AS userId, u.login_name AS loginName, u.user_name AS userName,'
            . ' d.dept_name AS deptName, u.phonenumber AS phonenumber, r.read_time AS readTime'
            . ' FROM sys_notice_read r'
            . ' INNER JOIN sys_user u ON u.user_id = r.user_id AND u.del_flag = ?'
            . ' LEFT JOIN sys_dept d ON d.dept_id = u.dept_id'
            . ' WHERE r.notice_id = ?';
        $bindings = ['0', $noticeId];
        if ($searchValue !== '') {
            $sql .= ' AND (u.login_name LIKE ? OR u.user_name LIKE ?)';
            $like = '%' . $searchValue . '%';
            array_push($bindings, $like, $like);
        }
        $sql .= ' ORDER BY r.read_time DESC';
        return array_map(fn($r) => (array)$r, Db::query($sql, $bindings));
    }

    /** 未读数（经典版保留方法，无端点消费——服务层完整性） */
    public static function selectUnreadCount(int $userId): int
    {
        return (int)Db::table('sys_notice n')
            ->whereRaw("n.status = '0' and not exists (select 1 from sys_notice_read r where r.notice_id = n.notice_id and r.user_id = " . (int)$userId . ')')
            ->count();
    }
}
