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
 * 字典数据控制器（对位经典版 SysDictDataController 8 方法）
 */
class DictDataController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '字典编码', 'field' => 'dictCode', 'numeric' => true],
        ['name' => '字典排序', 'field' => 'dictSort', 'numeric' => true],
        ['name' => '字典标签', 'field' => 'dictLabel'],
        ['name' => '字典键值', 'field' => 'dictValue'],
        ['name' => '字典类型', 'field' => 'dictType'],
        ['name' => '字典样式', 'field' => 'cssClass'],
        ['name' => '是否默认', 'field' => 'isDefault', 'convert' => 'Y=是,N=否'],
        ['name' => '状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
    ];

    /** GET /system/dict/data：经典版不传变量——TP 兜底空（原样） */
    #[Perm('system:dict:view')]
    public function index(Request $request): Response
    {
        return Response::create('dict/data/index', 'view')->assign([
            'dictList' => [],
            'dict'     => null,
            'datas'    => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dict/data/list：TableDataInfo（selectVo 12 列驼峰） */
    #[Perm('system:dict:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['dict_sort', 'create_time']);
        $query = DictService::selectDictDataList([
            'dictType' => (string)$request->post('dictType', ''),
            'dictLabel' => trim((string)$request->post('dictLabel', '')),
            'status'   => (string)$request->post('status', ''),
        ]);
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([DictService::class, 'toDataResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/dict/data/export */
    #[Perm('system:dict:export')]
    #[Log('字典数据', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['dict_sort', 'create_time']);
        $query = DictService::selectDictDataList([
            'dictType' => (string)$request->post('dictType', ''),
            'dictLabel' => trim((string)$request->post('dictLabel', '')),
            'status'   => (string)$request->post('status', ''),
        ]);
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([DictService::class, 'toDataResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '字典数据');
        return AjaxResult::success($fileName);
    }

    /** GET /system/dict/data/add/{dictType} */
    #[Perm('system:dict:add')]
    public function add(Request $request, string $dictType): Response
    {
        return Response::create('dict/data/add', 'view')->assign([
            'dictType'   => $dictType,
            'datas'      => DictService::listByType('sys_normal_disable'),
            'yesNoDatas' => DictService::listByType('sys_yes_no'),
        ]);
    }

    /** POST /system/dict/data/add：成功后 setCache 该类型 */
    #[Perm('system:dict:add')]
    #[Log('字典数据', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->dictDataInput($request);
        DictService::insertDictData($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/dict/data/edit/{dictCode} */
    #[Perm('system:dict:edit')]
    public function edit(Request $request, int $dictCode): Response
    {
        $dict = DictService::selectDictDataById($dictCode);
        if ($dict === null) {
            throw new \BusinessException('字典数据不存在');
        }
        return Response::create('dict/data/edit', 'view')->assign([
            'dict'       => $dict,
            'datas'      => DictService::listByType('sys_normal_disable'),
            'yesNoDatas' => DictService::listByType('sys_yes_no'),
        ]);
    }

    /** POST /system/dict/data/edit */
    #[Perm('system:dict:edit')]
    #[Log('字典数据', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->dictDataInput($request);
        $input['dict_code'] = (int)$request->post('dictCode', 0);
        DictService::updateDictData($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /system/dict/data/remove：逐个删 + 缓存联动（固定 success） */
    #[Perm('system:dict:remove')]
    #[Log('字典数据', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        DictService::deleteDictDataByIds($ids);
        return AjaxResult::success();
    }

    /** 字典数据表单输入收敛（文案对位 @Validated） */
    private function dictDataInput(Request $request): array
    {
        $dictLabel = trim((string)$request->post('dictLabel', ''));
        if ($dictLabel === '') {
            throw new \BusinessException('字典标签不能为空');
        }
        if (mb_strlen($dictLabel) > 100) {
            throw new \BusinessException('字典标签长度不能超过100个字符');
        }
        $dictValue = trim((string)$request->post('dictValue', ''));
        if ($dictValue === '') {
            throw new \BusinessException('字典键值不能为空');
        }
        if (mb_strlen($dictValue) > 100) {
            throw new \BusinessException('字典键值长度不能超过100个字符');
        }
        $dictType = trim((string)$request->post('dictType', ''));
        if ($dictType === '') {
            throw new \BusinessException('字典类型不能为空');
        }
        if (mb_strlen($dictType) > 100) {
            throw new \BusinessException('字典类型长度不能超过100个字符');
        }
        $cssClass = (string)$request->post('cssClass', '');
        if (mb_strlen($cssClass) > 100) {
            throw new \BusinessException('样式属性长度不能超过100个字符');
        }
        return [
            'dict_label' => $dictLabel,
            'dict_value' => $dictValue,
            'dict_type'  => $dictType,
            'css_class'  => $cssClass,
            'dict_sort'  => (int)$request->post('dictSort', 0),
            'list_class' => (string)$request->post('listClass', ''),
            'is_default' => (string)$request->post('isDefault', 'N'),
            'status'     => (string)$request->post('status', '0'),
            'remark'     => (string)$request->post('remark', ''),
        ];
    }
}
