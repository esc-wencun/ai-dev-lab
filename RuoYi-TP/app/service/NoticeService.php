<?php
declare(strict_types=1);

namespace app\service;

use think\db\Query;
use think\facade\Db;

/**
 * 公告服务（对位经典版 SysNoticeServiceImpl——纯表操作，无缓存无级联）
 */
class NoticeService
{
    /** 按 noticeId 查单条（下划线行；对位 selectNoticeVo 含 cast(notice_content as char)） */
    public static function selectNoticeById(int $noticeId): ?array
    {
        return Db::table('sys_notice')
            ->field('notice_id,notice_title,notice_type,notice_content,status,create_by,create_time,update_by,update_time,remark')
            ->where('notice_id', $noticeId)
            ->find();
    }

    /**
     * 列表 Query（noticeTitle like / noticeType eq / createBy like）。
     * 固定 order by notice_id desc；PageQuery 排序参数叠加时合并单 order 子句
     * （orderBy 优先、notice_id desc 兜底——行为等价经典版 PageHelper 追加语义）。
     */
    public static function selectNoticeList(array $filter): Query
    {
        $query = Db::table('sys_notice')
            ->field('notice_id,notice_title,notice_type,notice_content,status,create_by,create_time,update_by,update_time,remark');
        if (($filter['noticeTitle'] ?? '') !== '') {
            $query->whereLike('notice_title', '%' . $filter['noticeTitle'] . '%');
        }
        if (($filter['noticeType'] ?? '') !== '') {
            $query->where('notice_type', $filter['noticeType']);
        }
        if (($filter['createBy'] ?? '') !== '') {
            $query->whereLike('create_by', '%' . $filter['createBy'] . '%');
        }
        if (($filter['orderBy'] ?? '') !== '') {
            $query->order($filter['orderBy'], $filter['isAsc'] ?? 'asc');
        }
        $query->order('notice_id', 'desc');
        return $query;
    }

    /** 新增（createBy/createTime 显式写；对位 insert values sysdate()） */
    public static function insertNotice(array $input, string $createBy): int
    {
        return Db::table('sys_notice')->insert(array_merge($input, [
            'create_by'   => $createBy,
            'create_time' => date('Y-m-d H:i:s'),
        ])) ? 1 : 0;
    }

    /** 修改（updateBy/updateTime；对位 update set update_time = sysdate()） */
    public static function updateNotice(array $input, string $updateBy): int
    {
        return Db::table('sys_notice')
            ->where('notice_id', $input['notice_id'])
            ->update(array_merge($input, [
                'update_by'   => $updateBy,
                'update_time' => date('Y-m-d H:i:s'),
            ]));
    }

    /** 物理删（in；对位 deleteNoticeByIds(Convert.toStrArray(ids))） */
    public static function deleteNoticeByIds(array $ids): int
    {
        if (!$ids) {
            return 0;
        }
        return Db::table('sys_notice')->where('notice_id', 'in', $ids)->delete();
    }

    /** 新增/修改输入收敛（对位 @Validated 文案 + @Xss 标题校验；富文本正文免转义直存——经典版 excludes 原样） */
    public static function noticeInput(\think\Request $request): array
    {
        $noticeTitle = trim((string)$request->post('noticeTitle', ''));
        if ($noticeTitle === '') {
            throw new \BusinessException('公告标题不能为空');
        }
        if (mb_strlen($noticeTitle) > 50) {
            throw new \BusinessException('公告标题不能超过50个字符');
        }
        if (self::containsScriptChars($noticeTitle)) {
            throw new \BusinessException('公告标题不能包含脚本字符');
        }
        return [
            'notice_title'   => $noticeTitle,
            'notice_type'    => (string)$request->post('noticeType', ''),
            'notice_content' => (string)$request->post('noticeContent', ''),
            'status'         => (string)$request->post('status', '0'),
            'remark'         => (string)$request->post('remark', ''),
        ];
    }

    /** @Xss 对位：XssValidator 同款正则（HTML 标签检测） */
    public static function containsScriptChars(string $value): bool
    {
        $pattern = '/<(\S*?)[^>]*>.*?|<.*? \/>/';
        return preg_match($pattern, $value) === 1;
    }
}
