<?php
declare(strict_types=1);

namespace app\service;

use PageQuery;
use think\facade\Db;

/**
 * 岗位服务（对位经典版 SysPostServiceImpl）
 *
 * 无数据权限、无 admin 保护（经典版实锤一致）。
 */
final class PostService
{
    /** 列表查询构造器（对位 selectPostList；排序由 PageQuery 白名单驱动；分页/全量由调用方决定 select 或 paginate） */
    public static function selectPostList(array $filter, ?PageQuery $pq = null): \think\db\Query
    {
        $query = Db::table('sys_post');
        if (($filter['postCode'] ?? '') !== '') {
            $query->whereLike('post_code', '%' . $filter['postCode'] . '%');
        }
        if (($filter['postName'] ?? '') !== '') {
            $query->whereLike('post_name', '%' . $filter['postName'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', $filter['status']);
        }
        if ($pq !== null && $pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        return $query;
    }

    public static function selectPostById(int $postId): ?array
    {
        return Db::table('sys_post')->where('post_id', $postId)->find();
    }

    /** 名称全局唯一 → bool 唯一 */
    public static function checkPostNameUnique(string $postName, int $postId = 0): bool
    {
        $row = Db::table('sys_post')->where('post_name', $postName)->find();
        return $row === null || (int)$row['post_id'] === $postId;
    }

    /** 编码全局唯一 → bool 唯一 */
    public static function checkPostCodeUnique(string $postCode, int $postId = 0): bool
    {
        $row = Db::table('sys_post')->where('post_code', $postCode)->find();
        return $row === null || (int)$row['post_id'] === $postId;
    }

    public static function insertPost(array $post, string $loginName): void
    {
        Db::table('sys_post')->insert([
            'post_code'   => $post['post_code'],
            'post_name'   => $post['post_name'],
            'post_sort'   => (int)$post['post_sort'],
            'status'      => $post['status'] ?? '0',
            'remark'      => $post['remark'] ?? '',
            'create_by'   => $loginName,
            'create_time' => date('Y-m-d H:i:s'),
        ]);
    }

    public static function updatePost(array $post, string $loginName): void
    {
        Db::table('sys_post')->where('post_id', (int)$post['post_id'])->update([
            'post_code'   => $post['post_code'],
            'post_name'   => $post['post_name'],
            'post_sort'   => (int)$post['post_sort'],
            'status'      => $post['status'] ?? '0',
            'remark'      => $post['remark'] ?? '',
            'update_by'   => $loginName,
            'update_time' => date('Y-m-d H:i:s'),
        ]);
    }

    /** 批量删除（对位 deletePostByIds）：先校验占用（任意一个已分配 → 整批拒绝），后物理删 */
    public static function deletePostByIds(array $ids): void
    {
        foreach ($ids as $id) {
            $id = (int)$id;
            $used = Db::table('sys_user_post')->where('post_id', $id)->count();
            if ($used > 0) {
                $post = Db::table('sys_post')->where('post_id', $id)->field('post_name')->find();
                throw new \BusinessException(($post['post_name'] ?? $id) . '已分配,不能删除');
            }
        }
        Db::table('sys_post')->whereIn('post_id', array_map('intval', $ids))->delete();
    }

    /** 全量岗位（对位 selectPostAll；新增/编辑页 select2 消费；输出驼峰键） */
    public static function selectPostAll(): array
    {
        $rows = Db::table('sys_post')
            ->field('post_id,post_code,post_name,post_sort,status')
            ->order('post_sort')
            ->select()->toArray();
        return array_map(static fn(array $r): array => [
            'postId'   => (int)$r['post_id'],
            'postCode' => $r['post_code'],
            'postName' => $r['post_name'],
            'postSort' => (int)$r['post_sort'],
            'status'   => $r['status'],
            'flag'     => false,
        ], $rows);
    }

    /** 全量岗位 + 用户已有岗位 flag=true 合并（对位 selectPostsByUserId + flag 合并） */
    public static function selectPostsByUserId(int $userId): array
    {
        $posts = self::selectPostAll();
        $owned = Db::table('sys_user_post')->where('user_id', $userId)->column('post_id');
        foreach ($posts as &$p) {
            if (in_array($p['postId'], array_map('intval', $owned), true)) {
                $p['flag'] = true;
            }
        }
        return $posts;
    }
}
