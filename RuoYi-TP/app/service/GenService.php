<?php
declare(strict_types=1);

namespace app\service;

use GenConstants;
use PageQuery;
use think\facade\Db;

/**
 * 代码生成服务（对位经典版 IGenTableService/GenTableServiceImpl 的**数据层部分** +
 * GenTableColumnServiceImpl + 两个 Mapper XML）
 *
 * 范围拍板 A（specs/11.0.0「范围拍板建议」节）：数据层 8 端点照常实现，模板生成链
 * （preview / downloadCode / genCode / batchGenCode / createTable / synchDb）与编辑页渲染
 * 有意排除。因此经典版仅被上述排除端点消费的三个聚合读（selectGenTableById /
 * selectGenTableByName / selectGenTableAll）**不在此实现**，避免死代码；日后若做模板生成
 * 或编辑页，按 GenTableMapper.xml 的 join 读法补齐即可（spec 实施记录已注明）。
 *
 * SQL 逐句对位经典版 Mapper XML（含 `(select database())`、`NOT LIKE 'qrtz\_%'`/`'gen\_%'`、
 * lower like、information_schema 的 case 推断）；插入/更新的动态列 `<if>` 条件同样照抄，
 * 空串/null 的字段不写库，由 DB 默认值兜底。
 */
final class GenService
{
    /** 列表查询列（对位 GenTableMapper.xml 的 selectGenTableVo，含 options 列） */
    private const TABLE_FIELDS = 'table_id, table_name, table_comment, sub_table_name, sub_table_fk_name, class_name, '
        . 'tpl_category, package_name, module_name, business_name, function_name, function_author, form_col_num, '
        . 'gen_type, gen_path, options, create_by, create_time, update_by, update_time, remark';

    /** gen_table 插入列（对位 insertGenTable 的 <if>，table_id 自增不写） */
    private const TABLE_INSERT_FIELDS = [
        'table_name', 'table_comment', 'class_name', 'tpl_category', 'package_name', 'module_name', 'business_name',
        'function_name', 'function_author', 'form_col_num', 'gen_type', 'gen_path', 'remark', 'create_by',
    ];

    /** gen_table 更新列（对位 updateGenTable 的 <if>） */
    private const TABLE_UPDATE_FIELDS = [
        'table_name', 'table_comment', 'sub_table_name', 'sub_table_fk_name', 'class_name', 'function_author',
        'form_col_num', 'gen_type', 'gen_path', 'tpl_category', 'package_name', 'module_name', 'business_name',
        'function_name', 'options', 'remark',
    ];

    /** gen_table_column 插入列（对位 insertGenTableColumn 的 <if>） */
    private const COLUMN_INSERT_FIELDS = [
        'table_id', 'column_name', 'column_comment', 'column_type', 'java_type', 'java_field', 'is_pk', 'is_increment',
        'is_required', 'is_insert', 'is_edit', 'is_list', 'is_query', 'query_type', 'html_type', 'dict_type', 'sort',
        'create_by',
    ];

    /** 代码生成配置（对位经典版 generator.yml → config/gen.php） */
    public static function config(): array
    {
        $cfg = config('gen');
        return is_array($cfg) ? $cfg : [];
    }

    // ==================== 查询 ====================

    /** 业务列表查询构造器（对位 selectGenTableList；分页由调用方 paginate，排序列由 PageQuery 白名单驱动） */
    public static function selectGenTableList(array $filter, ?PageQuery $pq = null): \think\db\Query
    {
        $query = Db::table('gen_table')->field(self::TABLE_FIELDS);
        if (($filter['tableName'] ?? '') !== '') {
            $query->whereRaw('lower(table_name) like lower(?)', ['%' . $filter['tableName'] . '%']);
        }
        if (($filter['tableComment'] ?? '') !== '') {
            $query->whereRaw('lower(table_comment) like lower(?)', ['%' . $filter['tableComment'] . '%']);
        }
        // 表时间范围（对位 params.beginTime/endTime 的 date_format 比较，日粒度）
        if (($filter['beginTime'] ?? null) !== null && $filter['beginTime'] !== '') {
            $query->whereRaw("date_format(create_time,'%Y%m%d') >= date_format(?, '%Y%m%d')", [$filter['beginTime']]);
        }
        if (($filter['endTime'] ?? null) !== null && $filter['endTime'] !== '') {
            $query->whereRaw("date_format(create_time,'%Y%m%d') <= date_format(?, '%Y%m%d')", [$filter['endTime']]);
        }
        if ($pq !== null && $pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        return $query;
    }

    /**
     * 数据库表列表查询构造器（对位 selectDbTableList）
     * information_schema 查询：排除 qrtz_/gen_ 前缀与 gen_table 已导入表。
     */
    public static function selectDbTableList(array $filter): \think\db\Query
    {
        $query = Db::table('information_schema.tables')
            ->field('table_name, table_comment, create_time, update_time')
            ->whereRaw('table_schema = (select database())')
            ->whereRaw("table_name NOT LIKE 'qrtz\\_%' AND table_name NOT LIKE 'gen\\_%'")
            ->whereRaw('table_name NOT IN (select table_name from gen_table)');
        if (($filter['tableName'] ?? '') !== '') {
            $query->whereRaw('lower(table_name) like lower(?)', ['%' . $filter['tableName'] . '%']);
        }
        if (($filter['tableComment'] ?? '') !== '') {
            $query->whereRaw('lower(table_comment) like lower(?)', ['%' . $filter['tableComment'] . '%']);
        }
        $query->order('create_time', 'desc');
        return $query;
    }

    /** 按表名批量取库表（对位 selectDbTableListByNames：无「已导入」过滤，含 qrtz_/gen_ 排除） */
    public static function selectDbTableListByNames(array $tableNames): array
    {
        if ($tableNames === []) {
            return [];
        }
        $rows = Db::table('information_schema.tables')
            ->field('table_name, table_comment, create_time, update_time')
            ->whereRaw("table_name NOT LIKE 'qrtz\\_%' AND table_name NOT LIKE 'gen\\_%' AND table_schema = (select database())")
            ->whereIn('table_name', $tableNames)
            ->select()->toArray();
        return array_map(static fn(array $row): array => self::normalizeRow($row), $rows);
    }

    /**
     * 按表名取库表列（对位 selectDbTableColumnsByName）
     * is_required / is_pk / sort / is_increment 由 information_schema 的 case 表达式推断，SQL 原样照抄。
     */
    public static function selectDbTableColumnsByName(string $tableName): array
    {
        $rows = Db::query(
            "select column_name, "
            . "(case when (is_nullable = 'no' && column_key != 'PRI') then '1' else null end) as is_required, "
            . "(case when column_key = 'PRI' then '1' else '0' end) as is_pk, "
            . 'ordinal_position as sort, column_comment, '
            . "(case when extra = 'auto_increment' then '1' else '0' end) as is_increment, "
            . 'column_type from information_schema.columns '
            . 'where table_schema = (select database()) and table_name = (?) order by ordinal_position',
            [$tableName]
        );
        return array_map(static fn(array $row): array => self::normalizeRow($row), $rows);
    }

    /** 按表 ID 取列列表（对位 selectGenTableColumnListByTableId：where table_id order by sort） */
    public static function selectGenTableColumnListByTableId(int $tableId): array
    {
        return Db::table('gen_table_column')
            ->where('table_id', $tableId)
            ->order('sort')
            ->select()->toArray();
    }

    // ==================== 写入 ====================

    /**
     * 导入表结构（对位 importGenTable：**整批一个事务**，异常统一「导入失败：{msg}」）
     *
     * @param array  $tables   库表行（含 table_name / table_comment）
     * @param string $operName 操作人登录名（写入 create_by）
     */
    public static function importGenTable(array $tables, string $operName): void
    {
        try {
            Db::transaction(function () use ($tables, $operName) {
                $cfg = self::config();
                foreach ($tables as $table) {
                    $row = GenUtils::initTable($table, $operName, $cfg);
                    $data = self::filterInsertRow($row, self::TABLE_INSERT_FIELDS, ['table_name', 'form_col_num']);
                    $data['create_time'] = date('Y-m-d H:i:s');
                    $tableId = (int)Db::table('gen_table')->insertGetId($data);
                    if ($tableId > 0) {
                        // 保存列信息：先取库表列，再逐列推断后插入
                        $columns = self::selectDbTableColumnsByName((string)$row['table_name']);
                        foreach ($columns as $column) {
                            $inferred = GenUtils::initColumnField($column, [
                                'table_id'  => $tableId,
                                'create_by' => $row['create_by'] ?? null,
                            ]);
                            $colData = self::filterInsertRow($inferred, self::COLUMN_INSERT_FIELDS, ['sort']);
                            $colData['create_time'] = date('Y-m-d H:i:s');
                            Db::table('gen_table_column')->insert($colData);
                        }
                    }
                }
            });
        } catch (\Throwable $e) {
            throw new \BusinessException('导入失败：' . $e->getMessage());
        }
    }

    /**
     * 修改保存代码生成业务（对位 updateGenTable：主表 + 逐列，**一个事务**）
     *
     * 照抄经典版语义：主表更新影响行数 > 0 才继续更新列（MySQL 无实际变化时 affected=0 → 列不更新）。
     *
     * @param array $table 已收敛为下划线键的表数据（含 params / columns）
     */
    public static function updateGenTable(array $table, string $operName): void
    {
        $table['options'] = json_encode($table['params'] ?? [], JSON_UNESCAPED_UNICODE);
        $data = self::filterInsertRow($table, self::TABLE_UPDATE_FIELDS, ['table_name', 'sub_table_name', 'sub_table_fk_name', 'form_col_num', 'remark']);
        $data['update_by'] = $operName;
        $data['update_time'] = date('Y-m-d H:i:s');
        $tableId = (int)($table['table_id'] ?? 0);

        Db::transaction(function () use ($table, $data, $tableId, $operName) {
            $row = Db::table('gen_table')->where('table_id', $tableId)->update($data);
            if ($row > 0) {
                foreach (($table['columns'] ?? []) as $column) {
                    $columnId = (int)($column['column_id'] ?? 0);
                    if ($columnId <= 0) {
                        continue;
                    }
                    // 对位 updateGenTableColumn：全部列无条件写（表单未提交的字段写 null）
                    Db::table('gen_table_column')->where('column_id', $columnId)->update([
                        'column_comment' => $column['column_comment'] ?? null,
                        'java_type'      => $column['java_type'] ?? null,
                        'column_type'    => $column['column_type'] ?? null,
                        'java_field'     => $column['java_field'] ?? null,
                        'is_insert'      => $column['is_insert'] ?? null,
                        'is_edit'        => $column['is_edit'] ?? null,
                        'is_list'        => $column['is_list'] ?? null,
                        'is_query'       => $column['is_query'] ?? null,
                        'is_required'    => $column['is_required'] ?? null,
                        'query_type'     => $column['query_type'] ?? null,
                        'html_type'      => $column['html_type'] ?? null,
                        'dict_type'      => $column['dict_type'] ?? null,
                        'sort'           => $column['sort'] ?? null,
                        'update_by'      => $operName,
                        'update_time'    => date('Y-m-d H:i:s'),
                    ]);
                }
            }
        });
    }

    /** 删除业务对象（对位 deleteGenTableByIds：主表 + 列两表事务；固定 success 不查行数） */
    public static function deleteGenTableByIds(array $ids): void
    {
        $ids = array_values(array_filter(array_map('intval', $ids), static fn(int $v): bool => $v > 0));
        if ($ids === []) {
            // 经典版此处会拼出 `in ()` 触发 SQL 语法错误；TP 版按无操作处理（防误报 500）
            return;
        }
        Db::transaction(function () use ($ids) {
            Db::table('gen_table')->whereIn('table_id', $ids)->delete();
            Db::table('gen_table_column')->whereIn('table_id', $ids)->delete();
        });
    }

    /**
     * 修改保存前的参数校验（对位 validateEdit，四条文案照抄）
     */
    public static function validateEdit(array $table): void
    {
        $tplCategory = (string)($table['tpl_category'] ?? '');
        if ($tplCategory === GenConstants::TPL_TREE) {
            $params = is_array($table['params'] ?? null) ? $table['params'] : [];
            if ((string)($params[GenConstants::TREE_CODE] ?? '') === '') {
                throw new \BusinessException('树编码字段不能为空');
            }
            if ((string)($params[GenConstants::TREE_PARENT_CODE] ?? '') === '') {
                throw new \BusinessException('树父编码字段不能为空');
            }
            if ((string)($params[GenConstants::TREE_NAME] ?? '') === '') {
                throw new \BusinessException('树名称字段不能为空');
            }
        } elseif ($tplCategory === GenConstants::TPL_SUB) {
            if ((string)($table['sub_table_name'] ?? '') === '') {
                throw new \BusinessException('关联子表的表名不能为空');
            }
            if ((string)($table['sub_table_fk_name'] ?? '') === '') {
                throw new \BusinessException('子表关联的外键名不能为空');
            }
        }
    }

    // ==================== 响应行组装（驼峰） ====================

    /** 业务列表行（对位 GenTable 实体 Jackson 序列化：键集 = selectGenTableVo 的 21 列） */
    public static function toTableRow(array $row): array
    {
        return [
            'tableId'        => isset($row['table_id']) ? (int)$row['table_id'] : null,
            'tableName'      => $row['table_name'] ?? null,
            'tableComment'   => $row['table_comment'] ?? null,
            'subTableName'   => $row['sub_table_name'] ?? null,
            'subTableFkName' => $row['sub_table_fk_name'] ?? null,
            'className'      => $row['class_name'] ?? null,
            'tplCategory'    => $row['tpl_category'] ?? null,
            'packageName'    => $row['package_name'] ?? null,
            'moduleName'     => $row['module_name'] ?? null,
            'businessName'   => $row['business_name'] ?? null,
            'functionName'   => $row['function_name'] ?? null,
            'functionAuthor' => $row['function_author'] ?? null,
            'formColNum'     => isset($row['form_col_num']) && $row['form_col_num'] !== null ? (int)$row['form_col_num'] : null,
            'genType'        => $row['gen_type'] ?? null,
            'genPath'        => $row['gen_path'] ?? null,
            'options'        => $row['options'] ?? null,
            'createBy'       => $row['create_by'] ?? null,
            'createTime'     => $row['create_time'] ?? null,
            'updateBy'       => $row['update_by'] ?? null,
            'updateTime'     => $row['update_time'] ?? null,
            'remark'         => $row['remark'] ?? null,
        ];
    }

    /** 库表列表行（importTable 弹窗只用四列；不产出 GenTable 实体的其余 null 键） */
    public static function toDbTableRow(array $row): array
    {
        $row = self::normalizeRow($row);
        return [
            'tableName'    => $row['table_name'] ?? null,
            'tableComment' => $row['table_comment'] ?? null,
            'createTime'   => $row['create_time'] ?? null,
            'updateTime'   => $row['update_time'] ?? null,
        ];
    }

    /** 列列表行（对位 GenTableColumn 实体 Jackson 序列化：全 22 列驼峰） */
    public static function toColumnRow(array $row): array
    {
        return [
            'columnId'      => isset($row['column_id']) ? (int)$row['column_id'] : null,
            'tableId'       => isset($row['table_id']) ? (int)$row['table_id'] : null,
            'columnName'    => $row['column_name'] ?? null,
            'columnComment' => $row['column_comment'] ?? null,
            'columnType'    => $row['column_type'] ?? null,
            'javaType'      => $row['java_type'] ?? null,
            'javaField'     => $row['java_field'] ?? null,
            'isPk'          => $row['is_pk'] ?? null,
            'isIncrement'   => $row['is_increment'] ?? null,
            'isRequired'    => $row['is_required'] ?? null,
            'isInsert'      => $row['is_insert'] ?? null,
            'isEdit'        => $row['is_edit'] ?? null,
            'isList'        => $row['is_list'] ?? null,
            'isQuery'       => $row['is_query'] ?? null,
            'queryType'     => $row['query_type'] ?? null,
            'htmlType'      => $row['html_type'] ?? null,
            'dictType'      => $row['dict_type'] ?? null,
            'sort'          => isset($row['sort']) && $row['sort'] !== null ? (int)$row['sort'] : null,
            'createBy'      => $row['create_by'] ?? null,
            'createTime'    => $row['create_time'] ?? null,
            'updateBy'      => $row['update_by'] ?? null,
            'updateTime'    => $row['update_time'] ?? null,
        ];
    }

    // ==================== 内部工具 ====================

    /**
     * 行键小写化：MySQL 8 的 information_schema 列标签为大写（TABLE_NAME / COLUMN_TYPE …），
     * 而经典版 MyBatis resultMap 的 column 匹配大小写不敏感——此处等价对齐（只用于
     * information_schema 查询结果；gen_table/gen_table_column 是普通表，标签本就小写）。
     */
    private static function normalizeRow(array $row): array
    {
        return array_change_key_case($row, CASE_LOWER);
    }

    /**
     * 插入/更新的动态列过滤（对位 Mapper XML 的 `<if>` 条件：null/空串不写，DB 默认值兜底）
     *
     * @param array $row          源数据
     * @param array $fields       允许写入的列（白名单，防把 params/columns 误写进 SQL）
     * @param array $nullableOnly 仅判 null 的列（其余列 null 与空串都不写）
     */
    private static function filterInsertRow(array $row, array $fields, array $nullableOnly): array
    {
        $data = [];
        foreach ($fields as $field) {
            if (!array_key_exists($field, $row)) {
                continue;
            }
            $value = $row[$field];
            if (in_array($field, $nullableOnly, true)) {
                if ($value !== null) {
                    $data[$field] = $value;
                }
                continue;
            }
            if ($value !== null && $value !== '') {
                $data[$field] = $value;
            }
        }
        return $data;
    }
}
