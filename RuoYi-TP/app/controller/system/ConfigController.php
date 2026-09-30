<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\ConfigService;
use app\service\ExcelExportService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use TpConstant;
use think\facade\Db;
use think\Request;
use think\Response;

/**
 * 参数配置控制器（对位经典版 SysConfigController 10 方法）
 */
class ConfigController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '参数主键', 'field' => 'configId', 'numeric' => true],
        ['name' => '参数名称', 'field' => 'configName'],
        ['name' => '参数键名', 'field' => 'configKey'],
        ['name' => '参数键值', 'field' => 'configValue'],
        ['name' => '系统内置', 'field' => 'configType', 'convert' => 'Y=是,N=否'],
    ];

    /** GET /system/config */
    #[Perm('system:config:view')]
    public function index(Request $request): Response
    {
        return Response::create('config/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_yes_no'),
        ]);
    }

    /** POST /system/config/list：TableDataInfo（selectVo 10 列驼峰） */
    #[Perm('system:config:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['config_id']);
        $query = $this->buildListQuery($request);
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([$this, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/config/export */
    #[Perm('system:config:export')]
    #[Log('参数管理', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['config_id']);
        $query = $this->buildListQuery($request);
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([$this, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '参数数据');
        return AjaxResult::success($fileName);
    }

    /** GET /system/config/add */
    #[Perm('system:config:add')]
    public function add(Request $request): Response
    {
        return Response::create('config/add', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_yes_no'),
        ]);
    }

    /** POST /system/config/add：成功后 ConfigService::set */
    #[Perm('system:config:add')]
    #[Log('参数管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->configInput($request);
        if (!$this->checkKeyUnique($input['config_key'])) {
            return AjaxResult::error("新增参数'{$input['config_name']}'失败，参数键名已存在");
        }
        Db::table('sys_config')->insert(array_merge($input, [
            'create_by'   => (string)($session['loginName'] ?? ''),
            'create_time' => date('Y-m-d H:i:s'),
        ]));
        ConfigService::set($input['config_key'], $input['config_value']);
        return AjaxResult::success();
    }

    /** GET /system/config/edit/{configId} */
    #[Perm('system:config:edit')]
    public function edit(Request $request, int $configId): Response
    {
        $config = Db::table('sys_config')->where('config_id', $configId)->find();
        if ($config === null) {
            throw new \BusinessException('参数不存在');
        }
        return Response::create('config/edit', 'view')->assign([
            'config' => $config,
            'datas'  => \app\service\DictService::listByType('sys_yes_no'),
        ]);
    }

    /** POST /system/config/edit：改键名联动 refresh(old) + set(new) */
    #[Perm('system:config:edit')]
    #[Log('参数管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $configId = (int)$request->post('configId', 0);
        $input = $this->configInput($request);
        $input['config_id'] = $configId;
        if (!$this->checkKeyUnique($input['config_key'], $configId)) {
            return AjaxResult::error("修改参数'{$input['config_name']}'失败，参数键名已存在");
        }
        $old = Db::table('sys_config')->where('config_id', $configId)->field('config_key,config_value')->find();
        if ($old === null) {
            throw new \BusinessException('参数不存在');
        }
        // 键名变更 → 先清旧键缓存
        if ((string)$old['config_key'] !== $input['config_key']) {
            ConfigService::refresh((string)$old['config_key']);
        }
        Db::table('sys_config')->where('config_id', $configId)->update(array_merge($input, [
            'update_by'   => (string)($session['loginName'] ?? ''),
            'update_time' => date('Y-m-d H:i:s'),
        ]));
        ConfigService::set($input['config_key'], $input['config_value']);
        return AjaxResult::success();
    }

    /** POST /system/config/remove：内置不可删（尾随空格原样）+ refresh 键；固定 success */
    #[Perm('system:config:remove')]
    #[Log('参数管理', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        foreach ($ids as $configId) {
            $row = Db::table('sys_config')->where('config_id', $configId)->find();
            if ($row === null) {
                continue;
            }
            if ((string)$row['config_type'] === 'Y') {
                return AjaxResult::error('内置参数【' . $row['config_key'] . '】不能删除 ');
            }
            Db::table('sys_config')->where('config_id', $configId)->delete();
            ConfigService::refresh((string)$row['config_key']);
        }
        return AjaxResult::success();
    }

    /** GET /system/config/refreshCache：GET + remove 权限（原样 quirk）；reset = 清 + 全量预热 */
    #[Perm('system:config:remove')]
    #[Log('参数管理', Log::CLEAN)]
    public function refreshCache(Request $request): Response
    {
        ConfigService::reset();
        return AjaxResult::success();
    }

    /** POST /system/config/checkConfigKeyUnique：裸 boolean（无 #[Perm]） */
    public function checkConfigKeyUnique(Request $request): Response
    {
        $unique = $this->checkKeyUnique(
            (string)$request->post('configKey', ''),
            (int)$request->post('configId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** config_key 全局唯一（自身放行） */
    private function checkKeyUnique(string $configKey, int $configId = 0): bool
    {
        $row = Db::table('sys_config')->where('config_key', $configKey)->find();
        return $row === null || (int)$row['config_id'] === $configId;
    }

    /** 列表查询构造器（对位 selectConfigList） */
    private function buildListQuery(Request $request): \think\db\Query
    {
        $query = Db::table('sys_config')
            ->field('config_id,config_name,config_key,config_value,config_type,create_by,create_time,update_by,update_time,remark');
        $configName = trim((string)$request->post('configName', ''));
        if ($configName !== '') {
            $query->whereLike('config_name', '%' . $configName . '%');
        }
        $configKey = trim((string)$request->post('configKey', ''));
        if ($configKey !== '') {
            $query->whereLike('config_key', '%' . $configKey . '%');
        }
        $configType = (string)$request->post('configType', '');
        if ($configType !== '') {
            $query->where('config_type', $configType);
        }
        $params = $request->post('params', []);
        $begin = (string)($params['beginTime'] ?? '');
        $end = (string)($params['endTime'] ?? '');
        if ($begin !== '') {
            $query->whereTime('create_time', '>=', date('Y-m-d 00:00:00', strtotime($begin)));
        }
        if ($end !== '') {
            $query->whereTime('create_time', '<=', date('Y-m-d 23:59:59', strtotime($end)));
        }
        return $query;
    }

    /** 行 → 驼峰（selectVo 10 列） */
    private function toResponseRow(array $r): array
    {
        return [
            'configId'   => (int)$r['config_id'],
            'configName' => $r['config_name'],
            'configKey'  => $r['config_key'],
            'configValue' => $r['config_value'],
            'configType' => $r['config_type'],
            'createBy'   => $r['create_by'] ?? '',
            'createTime' => $r['create_time'] ?? null,
            'updateBy'   => $r['update_by'] ?? '',
            'updateTime' => $r['update_time'] ?? null,
            'remark'     => $r['remark'] ?? '',
        ];
    }

    /** 参数表单输入收敛（文案对位 @Validated） */
    private function configInput(Request $request): array
    {
        $configName = trim((string)$request->post('configName', ''));
        if ($configName === '') {
            throw new \BusinessException('参数名称不能为空');
        }
        if (mb_strlen($configName) > 100) {
            throw new \BusinessException('参数名称不能超过100个字符');
        }
        $configKey = trim((string)$request->post('configKey', ''));
        if ($configKey === '') {
            throw new \BusinessException('参数键名长度不能为空');
        }
        if (mb_strlen($configKey) > 100) {
            throw new \BusinessException('参数键名长度不能超过100个字符');
        }
        $configValue = trim((string)$request->post('configValue', ''));
        if ($configValue === '') {
            throw new \BusinessException('参数键值不能为空');
        }
        if (mb_strlen($configValue) > 500) {
            throw new \BusinessException('参数键值长度不能超过500个字符');
        }
        return [
            'config_name'  => $configName,
            'config_key'   => $configKey,
            'config_value' => $configValue,
            'config_type'  => (string)$request->post('configType', 'Y'),
            'remark'       => (string)$request->post('remark', ''),
        ];
    }
}
