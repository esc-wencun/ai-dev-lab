<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\NoticeReadService;
use app\service\NoticeService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 通知公告控制器（对位经典版 SysNoticeController 13 方法）
 *
 * 7.0.0 收编 2.0.0 最小版：listTop 转真实实现；补 markRead/markReadAll 路由。
 * listTop/markRead/markReadAll/view 四端点无 #[Perm] 仅登录态（普通用户读公告链路，经典版实锤）。
 */
class NoticeController extends \app\BaseController
{
    /** GET /system/notice */
    #[Perm('system:notice:view')]
    public function index(Request $request): Response
    {
        return Response::create('notice/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_notice_status'),
            'types' => \app\service\DictService::listByType('sys_notice_type'),
        ]);
    }

    /** POST /system/notice/list：TableDataInfo（selectVo 10 列驼峰；list 输出含 noticeContent 全文——经典版 selectVo 未排除，原样） */
    #[Perm('system:notice:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['create_time']);
        $filter = [
            'noticeTitle' => trim((string)$request->post('noticeTitle', '')),
            'noticeType'  => (string)$request->post('noticeType', ''),
            'createBy'    => trim((string)$request->post('createBy', '')),
        ];
        if ($pq->orderBy !== null) {
            $filter['orderBy'] = $pq->orderBy;
            $filter['isAsc'] = $pq->isAsc;
        }
        $page = NoticeService::selectNoticeList($filter)->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([$this, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** GET /system/notice/add（全屏 tab 页） */
    #[Perm('system:notice:add')]
    public function add(Request $request): Response
    {
        return Response::create('notice/add', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_notice_status'),
            'types' => \app\service\DictService::listByType('sys_notice_type'),
        ]);
    }

    /** POST /system/notice/add */
    #[Perm('system:notice:add')]
    #[Log('通知公告', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = NoticeService::noticeInput($request);
        NoticeService::insertNotice($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/notice/edit/{noticeId} */
    #[Perm('system:notice:edit')]
    public function edit(Request $request, int $noticeId): Response
    {
        $notice = NoticeService::selectNoticeById($noticeId);
        if ($notice === null) {
            throw new \BusinessException('公告不存在');
        }
        return Response::create('notice/edit', 'view')->assign([
            'notice' => $notice,
            'datas'  => \app\service\DictService::listByType('sys_notice_status'),
            'types'  => \app\service\DictService::listByType('sys_notice_type'),
        ]);
    }

    /** POST /system/notice/edit */
    #[Perm('system:notice:edit')]
    #[Log('通知公告', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = NoticeService::noticeInput($request);
        $input['notice_id'] = (int)$request->post('noticeId', 0);
        NoticeService::updateNotice($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/notice/view/{noticeId}：popupRight 弹层内容（无 #[Perm] 仅登录态） */
    public function view(Request $request, int $noticeId): Response
    {
        $notice = NoticeService::selectNoticeById($noticeId);
        if ($notice === null) {
            throw new \BusinessException('公告不存在');
        }
        return Response::create('notice/view', 'view')->assign(['notice' => $notice]);
    }

    /** GET /system/notice/listTop：收编 2.0.0 空形态 → 真实查询（响应键与主框架消费 JS 一字不差） */
    public function listTop(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)($session['user_id'] ?? $session['userId'] ?? 0);
        $list = NoticeReadService::selectNoticeListWithReadStatus($userId, 5);
        $unreadCount = count(array_filter($list, fn($n) => !$n['isRead']));
        return AjaxResult::of(TpConstant::CODE_SUCCESS, '操作成功', [
            'data'        => $list,
            'unreadCount' => $unreadCount,
        ]);
    }

    /** POST /system/notice/markRead：insert ignore 幂等（无 #[Perm] 仅登录态） */
    public function markRead(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)($session['user_id'] ?? $session['userId'] ?? 0);
        NoticeReadService::markRead((int)$request->post('noticeId', 0), $userId);
        return AjaxResult::success();
    }

    /** POST /system/notice/markReadAll：ids 逗号串（前端回传全部 5 条含已读；空数组短路） */
    public function markReadAll(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)($session['user_id'] ?? $session['userId'] ?? 0);
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        NoticeReadService::markReadBatch($userId, $ids);
        return AjaxResult::success();
    }

    /** GET /system/notice/readUsers/{noticeId}：已读用户页 */
    #[Perm('system:notice:list')]
    public function readUsers(Request $request, int $noticeId): Response
    {
        $notice = NoticeService::selectNoticeById($noticeId);
        if ($notice === null) {
            throw new \BusinessException('公告不存在');
        }
        return Response::create('notice/readUsers', 'view')->assign(['notice' => $notice]);
    }

    /** POST /system/notice/readUsers/list：TableDataInfo（6 驼峰列；read_time desc 固定） */
    #[Perm('system:notice:list')]
    public function readUsersList(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['read_time']);
        $query = NoticeReadService::selectReadUsersByNoticeId(
            (int)$request->post('noticeId', 0),
            trim((string)$request->post('searchValue', ''))
        );
        $total = count($query);
        if ($pq->orderBy !== null) {
            $orderBy = $pq->orderBy === 'readTime' ? 'read_time' : $pq->orderBy;
            usort($query, function ($a, $b) use ($orderBy, $pq) {
                $cmp = strcmp((string)($a[$orderBy] ?? ''), (string)($b[$orderBy] ?? ''));
                return $pq->isAsc === 'desc' ? -$cmp : $cmp;
            });
        }
        $rows = array_slice($query, ($pq->pageNum - 1) * $pq->pageSize, $pq->pageSize);
        return TableDataInfo::of(array_map([$this, 'toReadUserRow'], $rows), $total);
    }

    /** POST /system/notice/remove：两连删（先清已读再删公告；toAjax 按公告行数） */
    #[Perm('system:notice:remove')]
    #[Log('通知公告', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        NoticeReadService::deleteByNoticeIds($ids);
        return AjaxResult::success(NoticeService::deleteNoticeByIds($ids) > 0 ? '操作成功' : '操作失败');
    }

    /** 行 → 驼峰（selectVo 10 列；noticeContent 全文原样输出） */
    private function toResponseRow(array $r): array
    {
        return [
            'noticeId'      => (int)$r['notice_id'],
            'noticeTitle'   => $r['notice_title'],
            'noticeType'    => $r['notice_type'],
            'noticeContent' => $r['notice_content'],
            'status'        => $r['status'],
            'createBy'      => $r['create_by'] ?? '',
            'createTime'    => $r['create_time'] ?? null,
            'updateBy'      => $r['update_by'] ?? '',
            'updateTime'    => $r['update_time'] ?? null,
            'remark'        => $r['remark'] ?? '',
        ];
    }

    /** 已读用户行 → 驼峰 6 列 */
    private function toReadUserRow(array $r): array
    {
        return [
            'userId'      => (int)$r['userId'],
            'loginName'   => $r['loginName'],
            'userName'    => $r['userName'],
            'deptName'    => $r['deptName'] ?? '',
            'phonenumber' => $r['phonenumber'] ?? '',
            'readTime'    => $r['readTime'],
        ];
    }
}
