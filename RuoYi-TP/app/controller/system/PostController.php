<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;
use PageQuery;
use app\service\ExcelExportService;
use app\service\PostService;
use AjaxResult;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 岗位管理控制器（对位经典版 SysPostController，10 方法）
 */
class PostController extends \app\BaseController
{
    /** GET /system/post：列表页 */
    #[Perm('system:post:view')]
    public function index(Request $request): Response
    {
        return Response::create('post/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/post/list：TableDataInfo 分页 */
    #[Perm('system:post:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['post_code', 'post_name', 'post_sort', 'create_time']);
        $query = PostService::selectPostList([
            'postCode' => trim((string)$request->post('postCode', '')),
            'postName' => trim((string)$request->post('postName', '')),
            'status'   => (string)$request->post('status', ''),
        ], $pq);
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        // 输出驼峰键（对位经典版 SysPost 实体 Jackson 序列化；bootstrap-table columns field: postId 等按驼峰取值）
        $rows = array_map(static fn(array $r): array => [
            'postId'     => (int)$r['post_id'],
            'postCode'   => $r['post_code'],
            'postName'   => $r['post_name'],
            'postSort'   => (int)$r['post_sort'],
            'status'     => $r['status'],
            'remark'     => $r['remark'],
            'createTime' => $r['create_time'],
        ], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/post/export：全量导出（文件名装在 msg） */
    #[Perm('system:post:export')]
    #[Log('岗位管理', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['post_code', 'post_name', 'post_sort', 'create_time']);
        $rows = PostService::selectPostList([
            'postCode' => trim((string)$request->post('postCode', '')),
            'postName' => trim((string)$request->post('postName', '')),
            'status'   => (string)$request->post('status', ''),
        ], $pq)->select()->toArray();

        $columns = [
            ['name' => '岗位序号', 'field' => 'post_id', 'numeric' => true],
            ['name' => '岗位编码', 'field' => 'post_code'],
            ['name' => '岗位名称', 'field' => 'post_name'],
            ['name' => '岗位排序', 'field' => 'post_sort', 'numeric' => true],
            ['name' => '状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
        ];
        $fileName = ExcelExportService::export($rows, $columns, '岗位数据');
        return AjaxResult::success($fileName);
    }

    /** GET /system/post/add：新增弹窗 */
    #[Perm('system:post:add')]
    public function add(Request $request): Response
    {
        return Response::create('post/add', 'view');
    }

    /** POST /system/post/add */
    #[Perm('system:post:add')]
    #[Log('岗位管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $input = $this->postInput($request);
        if (!PostService::checkPostNameUnique($input['post_name'])) {
            return AjaxResult::error("新增岗位'{$input['post_name']}'失败，岗位名称已存在");
        }
        if (!PostService::checkPostCodeUnique($input['post_code'])) {
            return AjaxResult::error("新增岗位'{$input['post_name']}'失败，岗位编码已存在");
        }
        PostService::insertPost($input, (string)($user['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/post/edit/{postId}：修改弹窗（无数据权限校验，经典版一致） */
    #[Perm('system:post:edit')]
    public function edit(Request $request, int $postId): Response
    {
        $post = PostService::selectPostById($postId);
        if ($post === null) {
            throw new \BusinessException('岗位不存在');
        }
        return Response::create('post/edit', 'view')->assign(['post' => $post]);
    }

    /** POST /system/post/edit */
    #[Perm('system:post:edit')]
    #[Log('岗位管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $input = $this->postInput($request);
        $postId = (int)$input['post_id'];
        if (!PostService::checkPostNameUnique($input['post_name'], $postId)) {
            return AjaxResult::error("修改岗位'{$input['post_name']}'失败，岗位名称已存在");
        }
        if (!PostService::checkPostCodeUnique($input['post_code'], $postId)) {
            return AjaxResult::error("修改岗位'{$input['post_name']}'失败，岗位编码已存在");
        }
        PostService::updatePost($input, (string)($user['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /system/post/remove：物理删（ids 逗号串；已分配整批拒绝 500） */
    #[Perm('system:post:remove')]
    #[Log('岗位管理', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_filter(explode(',', (string)$request->post('ids', '')), fn($v) => $v !== '');
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        PostService::deletePostByIds($ids);
        return AjaxResult::success();
    }

    /** POST /system/post/checkPostNameUnique：裸 boolean（无 #[Perm]） */
    public function checkPostNameUnique(Request $request): Response
    {
        $unique = PostService::checkPostNameUnique(
            (string)$request->post('postName', ''),
            (int)$request->post('postId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/post/checkPostCodeUnique：裸 boolean（无 #[Perm]） */
    public function checkPostCodeUnique(Request $request): Response
    {
        $unique = PostService::checkPostCodeUnique(
            (string)$request->post('postCode', ''),
            (int)$request->post('postId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** 岗位表单输入收敛（文案对位经典版 @Validated 消息） */
    private function postInput(Request $request): array
    {
        $postCode = trim((string)$request->post('postCode', ''));
        if ($postCode === '') {
            throw new \BusinessException('岗位编码不能为空');
        }
        if (mb_strlen($postCode) > 64) {
            throw new \BusinessException('岗位编码长度不能超过64个字符');
        }
        $postName = trim((string)$request->post('postName', ''));
        if ($postName === '') {
            throw new \BusinessException('岗位名称不能为空');
        }
        if (mb_strlen($postName) > 50) {
            throw new \BusinessException('岗位名称长度不能超过50个字符');
        }
        if ($request->post('postSort') === null || $request->post('postSort') === '') {
            throw new \BusinessException('显示顺序不能为空');
        }
        return [
            'post_id'   => (int)$request->post('postId', 0),
            'post_code' => $postCode,
            'post_name' => $postName,
            'post_sort' => (int)$request->post('postSort', 0),
            'status'    => (string)$request->post('status', '0'),
            'remark'    => (string)$request->post('remark', ''),
        ];
    }
}
