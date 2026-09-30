<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\DictService;
use app\service\ExcelExportService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 字典类型控制器（对位经典版 SysDictTypeController 13 方法）
 */
class DictTypeController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '字典主键', 'field' => 'dictId', 'numeric' => true],
        ['name' => '字典名称', 'field' => 'dictName'],
        ['name' => '字典类型', 'field' => 'dictType'],
        ['name' => '状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
    ];

    /** GET /system/dict */
    #[Perm('system:dict:view')]
    public function index(Request $request): Response
    {
        return Response::create('dict/type/index', 'view')->assign([
            'datas' => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dict/list：TableDataInfo（selectVo 7 列驼峰） */
    #[Perm('system:dict:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['dict_id', 'dict_name', 'dict_type', 'create_time']);
        $query = DictService::selectDictTypeList($this->listFilter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([DictService::class, 'toTypeResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/dict/export */
    #[Perm('system:dict:export')]
    #[Log('字典类型', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['dict_id', 'dict_name', 'dict_type', 'create_time']);
        $query = DictService::selectDictTypeList($this->listFilter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([DictService::class, 'toTypeResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '字典类型');
        return AjaxResult::success($fileName);
    }

    /** GET /system/dict/add */
    #[Perm('system:dict:add')]
    public function add(Request $request): Response
    {
        return Response::create('dict/type/add', 'view')->assign([
            'datas' => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dict/add */
    #[Perm('system:dict:add')]
    #[Log('字典类型', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->dictTypeInput($request);
        if (!DictService::checkDictTypeUnique($input['dict_type'])) {
            return AjaxResult::error("新增字典'{$input['dict_name']}'失败，字典类型已存在");
        }
        DictService::insertDictType($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/dict/edit/{dictId} */
    #[Perm('system:dict:edit')]
    public function edit(Request $request, int $dictId): Response
    {
        $dict = DictService::selectDictTypeById($dictId);
        if ($dict === null) {
            throw new \BusinessException('字典类型不存在');
        }
        return Response::create('dict/type/edit', 'view')->assign([
            'dict'  => $dict,
            'datas' => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dict/edit（事务：dict_type 级联 + 新类型缓存重刷） */
    #[Perm('system:dict:edit')]
    #[Log('字典类型', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->dictTypeInput($request);
        $input['dict_id'] = (int)$request->post('dictId', 0);
        if (!DictService::checkDictTypeUnique($input['dict_type'], $input['dict_id'])) {
            return AjaxResult::error("修改字典'{$input['dict_name']}'失败，字典类型已存在");
        }
        DictService::updateDictType($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /system/dict/remove：占用校验 + 物理删（无事务经典原样；固定 success） */
    #[Perm('system:dict:remove')]
    #[Log('字典类型', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = $this->ids($request->post('ids', ''));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        foreach ($ids as $dictId) {
            $type = DictService::selectDictTypeById($dictId);
            if ($type === null) {
                continue;
            }
            if (DictService::countDictDataByType((string)$type['dict_type']) > 0) {
                return AjaxResult::error($type['dict_name'] . '已分配,不能删除');
            }
            DictService::deleteDictTypeById($dictId);
            DictService::removeCache((string)$type['dict_type']);
        }
        return AjaxResult::success();
    }

    /** GET /system/dict/refreshCache：GET + remove 权限（原样 quirk） */
    #[Perm('system:dict:remove')]
    #[Log('字典类型', Log::CLEAN)]
    public function refreshCache(Request $request): Response
    {
        DictService::resetCache();
        return AjaxResult::success();
    }

    /** GET /system/dict/detail/{dictId}：复用 data 列表页 */
    #[Perm('system:dict:list')]
    public function detail(Request $request, int $dictId): Response
    {
        $dict = DictService::selectDictTypeById($dictId);
        if ($dict === null) {
            throw new \BusinessException('字典类型不存在');
        }
        return Response::create('dict/data/index', 'view')->assign([
            'dict'     => $dict,
            'dictList' => DictService::selectDictTypeAll(),
            'datas'    => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dict/checkDictTypeUnique：裸 boolean（无 #[Perm]） */
    public function checkDictTypeUnique(Request $request): Response
    {
        $unique = DictService::checkDictTypeUnique(
            (string)$request->post('dictType', ''),
            (int)$request->post('dictId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** GET /system/dict/selectDictTree/{columnId}/{dictType}：辅助页（经典无调用方，原样保留） */
    public function selectDictTree(Request $request, int $columnId, string $dictType): Response
    {
        $dict = DictService::selectDictTypeByType($dictType) ?: [];
        return Response::create('dict/type/tree', 'view')->assign([
            'columnId' => $columnId,
            'dict'     => $dict,
        ]);
    }

    /** GET /system/dict/treeData：Ztree 平铺裸数组（无 #[Perm]） */
    public function treeData(Request $request): Response
    {
        return \think\Response::create(DictService::selectDictTree(), 'json');
    }

    /** 字典类型表单输入收敛（文案对位 @Validated；含正则校验） */
    private function dictTypeInput(Request $request): array
    {
        $dictName = trim((string)$request->post('dictName', ''));
        if ($dictName === '') {
            throw new \BusinessException('字典名称不能为空');
        }
        if (mb_strlen($dictName) > 100) {
            throw new \BusinessException('字典类型名称长度不能超过100个字符');
        }
        $dictType = trim((string)$request->post('dictType', ''));
        if ($dictType === '') {
            throw new \BusinessException('字典类型不能为空');
        }
        if (mb_strlen($dictType) > 100) {
            throw new \BusinessException('字典类型类型长度不能超过100个字符');
        }
        if (!preg_match('/^[a-z][a-z0-9_]*$/', $dictType)) {
            throw new \BusinessException('字典类型必须以字母开头，且只能为（小写字母，数字，下滑线）');
        }
        return [
            'dict_name' => $dictName,
            'dict_type' => $dictType,
            'status'    => (string)$request->post('status', '0'),
            'remark'    => (string)$request->post('remark', ''),
        ];
    }

    private function listFilter(Request $request): array
    {
        $params = $request->post('params', []);
        return [
            'dictName'  => trim((string)$request->post('dictName', '')),
            'dictType'  => trim((string)$request->post('dictType', '')),
            'status'    => (string)$request->post('status', ''),
            'beginTime' => (string)($params['beginTime'] ?? ''),
            'endTime'   => (string)($params['endTime'] ?? ''),
        ];
    }

    private function ids(string $comma): array
    {
        return array_values(array_filter(array_map('intval', explode(',', $comma)), fn($v) => $v > 0));
    }
}
