<?php
declare(strict_types=1);

namespace app\controller\tool;

use app\attribute\Log;
use app\attribute\Perm;
use app\service\GenService;
use app\service\PermissionService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use think\Request;
use think\Response;

/**
 * 代码生成控制器（对位经典版 GenController）
 *
 * 范围拍板 A（specs/11.0.0「范围拍板建议」节）：本控制器只实现**数据层 8 条路由**——
 *   GET  /tool/gen               渲染列表页
 *   POST /tool/gen/list          业务列表
 *   POST /tool/gen/db/list       库表列表
 *   POST /tool/gen/column/list   列列表
 *   GET  /tool/gen/importTable   渲染导入弹窗页
 *   POST /tool/gen/importTable   导入表结构
 *   POST /tool/gen/edit          编辑保存
 *   POST /tool/gen/remove        删除
 * 模板生成链（GET edit 渲染 / preview / download / genCode / batchGenCode / createTable /
 * synchDb）有意排除：**路由不注册即 404**（deviations #21），故本类无对应方法。
 */
class GenController extends \app\BaseController
{
    /** 列表可排序字段白名单（对位 gen.html 的 sortable 列：表名称/表描述/实体类名称/创建时间/更新时间） */
    private const SORTABLE_FIELDS = ['table_name', 'table_comment', 'class_name', 'create_time', 'update_time'];

    /** GET /tool/gen（tool:gen:view） */
    #[Perm('tool:gen:view')]
    public function index(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        return Response::create('gen/index', 'view')->assign([
            // 工具栏「创建」按钮按**角色**显隐（对位 shiro:hasRole="admin"；check_perm 是权限串，管不了角色）
            'isAdmin' => PermissionService::isAdmin($session),
        ]);
    }

    /** POST /tool/gen/list（tool:gen:list）：业务列表，行键驼峰 */
    #[Perm('tool:gen:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), self::SORTABLE_FIELDS);
        $page = GenService::selectGenTableList([
            'tableName'    => trim((string)$request->post('tableName', '')),
            'tableComment' => trim((string)$request->post('tableComment', '')),
            'beginTime'    => $pq->beginTime,
            'endTime'      => $pq->endTime,
        ], $pq)->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([GenService::class, 'toTableRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /tool/gen/db/list（tool:gen:list）：库表列表（information_schema，排除 qrtz_/gen_ 与已导入表） */
    #[Perm('tool:gen:list')]
    public function dbList(Request $request): Response
    {
        $pq = PageQuery::from($request->post());
        $page = GenService::selectDbTableList([
            'tableName'    => trim((string)$request->post('tableName', '')),
            'tableComment' => trim((string)$request->post('tableComment', '')),
        ])->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([GenService::class, 'toDbTableRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /tool/gen/column/list（tool:gen:list）：手工组装 rows/total，不走分页（经典版原样） */
    #[Perm('tool:gen:list')]
    public function columnList(Request $request): Response
    {
        $rows = array_map(
            [GenService::class, 'toColumnRow'],
            GenService::selectGenTableColumnListByTableId((int)$request->post('tableId', 0))
        );
        return TableDataInfo::of($rows, count($rows));
    }

    /** GET /tool/gen/importTable（tool:gen:list）：渲染导入弹窗页 */
    #[Perm('tool:gen:list')]
    public function importTable(Request $request): Response
    {
        return Response::create('gen/importTable', 'view');
    }

    /** POST /tool/gen/importTable（tool:gen:list）：导入表结构（整批一个事务） */
    #[Perm('tool:gen:list')]
    #[Log('代码生成', Log::IMPORT)]
    public function importTableSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $tables = (string)$request->post('tables', '');
        $tableNames = array_values(array_filter(
            array_map('trim', explode(',', $tables)),
            static fn(string $v): bool => $v !== ''
        ));
        // 经典版：先按名查库表，再整批导入（导入失败文案由 service 统一给「导入失败：{msg}」）
        $tableList = GenService::selectDbTableListByNames($tableNames);
        GenService::importGenTable($tableList, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /tool/gen/edit（tool:gen:edit）：编辑保存（必填校验 → validateEdit → 主表+列事务更新） */
    #[Perm('tool:gen:edit')]
    #[Log('代码生成', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $table = $this->editInput($request);
        $this->validateRequired($table);
        GenService::validateEdit($table);
        GenService::updateGenTable($table, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /tool/gen/remove（tool:gen:remove）：主表+列两表事务删，固定 success（对位经典版不查行数） */
    #[Perm('tool:gen:remove')]
    #[Log('代码生成', Log::DELETE)]
    public function remove(Request $request): Response
    {
        GenService::deleteGenTableByIds(explode(',', (string)$request->post('ids', '')));
        return AjaxResult::success();
    }

    // ==================== 入参收敛 ====================

    /** 编辑保存入参（表单驼峰 → 下划线；字段集对位 @Validated GenTable 的数据绑定） */
    private function editInput(Request $request): array
    {
        $params = $request->post('params', []);
        $columns = $request->post('columns', []);
        return [
            'table_id'          => (int)$request->post('tableId', 0),
            'table_name'        => (string)$request->post('tableName', ''),
            'table_comment'     => (string)$request->post('tableComment', ''),
            'sub_table_name'    => $this->nullableString($request->post('subTableName')),
            'sub_table_fk_name' => $this->nullableString($request->post('subTableFkName')),
            'class_name'        => (string)$request->post('className', ''),
            'tpl_category'      => (string)$request->post('tplCategory', ''),
            'package_name'      => (string)$request->post('packageName', ''),
            'module_name'       => (string)$request->post('moduleName', ''),
            'business_name'     => (string)$request->post('businessName', ''),
            'function_name'     => (string)$request->post('functionName', ''),
            'function_author'   => (string)$request->post('functionAuthor', ''),
            // 对位 Java `private int formColNum` 的基本类型语义：表单未提交时为 0（不是 null）
            'form_col_num'      => (int)$request->post('formColNum', 0),
            'gen_type'          => $this->nullableString($request->post('genType')),
            'gen_path'          => $this->nullableString($request->post('genPath')),
            'remark'            => $this->nullableString($request->post('remark')),
            'params'            => is_array($params) ? $params : [],
            'columns'           => array_map(
                fn(mixed $c): array => $this->columnInput(is_array($c) ? $c : []),
                is_array($columns) ? $columns : []
            ),
        ];
    }

    /** 列表单列入参（对位 GenTableColumn 绑定；未提交字段即 null，更新时按经典版语义写 null） */
    private function columnInput(array $column): array
    {
        return [
            'column_id'      => (int)($column['columnId'] ?? 0),
            'column_comment' => $this->nullableString($column['columnComment'] ?? null),
            'column_type'    => $this->nullableString($column['columnType'] ?? null),
            'java_type'      => $this->nullableString($column['javaType'] ?? null),
            'java_field'     => $this->nullableString($column['javaField'] ?? null),
            'is_required'    => $this->nullableString($column['isRequired'] ?? null),
            'is_insert'      => $this->nullableString($column['isInsert'] ?? null),
            'is_edit'        => $this->nullableString($column['isEdit'] ?? null),
            'is_list'        => $this->nullableString($column['isList'] ?? null),
            'is_query'       => $this->nullableString($column['isQuery'] ?? null),
            'query_type'     => $this->nullableString($column['queryType'] ?? null),
            'html_type'      => $this->nullableString($column['htmlType'] ?? null),
            'dict_type'      => $this->nullableString($column['dictType'] ?? null),
            'sort'           => isset($column['sort']) && $column['sort'] !== '' ? (int)$column['sort'] : null,
        ];
    }

    /**
     * 必填校验（对位 @Validated 触发的 Bean Validation 文案，按字段声明顺序）
     * 表级 8 项来自 GenTable 的 @NotBlank；列级 1 项来自 GenTableColumn.javaField 的 @NotBlank。
     */
    private function validateRequired(array $table): void
    {
        $checks = [
            'table_name'      => '表名称不能为空',
            'table_comment'   => '表描述不能为空',
            'class_name'      => '实体类名称不能为空',
            'package_name'    => '生成包路径不能为空',
            'module_name'     => '生成模块名不能为空',
            'business_name'   => '生成业务名不能为空',
            'function_name'   => '生成功能名不能为空',
            'function_author' => '作者不能为空',
        ];
        foreach ($checks as $field => $message) {
            if (trim((string)($table[$field] ?? '')) === '') {
                throw new \BusinessException($message);
            }
        }
        foreach (($table['columns'] ?? []) as $column) {
            if (trim((string)($column['java_field'] ?? '')) === '') {
                throw new \BusinessException('Java属性不能为空');
            }
        }
    }

    private function nullableString(mixed $value): ?string
    {
        return $value === null ? null : (string)$value;
    }
}
