<?php
declare(strict_types=1);

namespace app\service;

use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../app/common/GenConstants.php';
require_once __DIR__ . '/../../app/service/GenUtils.php';

/**
 * GenUtils 纯逻辑单测（对位经典版 GenUtils 全规则，逐条实测）
 *
 * 不启动 TP 容器、不连库：配置值以常量数组形式传入。
 * 夹具取 sys_notice 真实列形态（ry-tp 库实测 column_type：int / varchar(50) / char(1) /
 * longblob / datetime / varchar(255)），另补合成类型覆盖四条分支。
 */
final class GenUtilsTest extends TestCase
{
    /** 对位经典版 generator.yml 原值 */
    private const CFG = [
        'author'         => 'ruoyi',
        'packageName'    => 'com.ruoyi.system',
        'autoRemovePre'  => false,
        'tablePrefix'    => 'sys_',
        'allowOverwrite' => false,
    ];

    // ---------- 基础字符串工具 ----------

    public function testGetDbType(): void
    {
        $this->assertSame('varchar', GenUtils::getDbType('varchar(50)'));
        $this->assertSame('int', GenUtils::getDbType('int'));
        $this->assertSame('decimal', GenUtils::getDbType('decimal(10,2)'));
        $this->assertSame('longblob', GenUtils::getDbType('longblob'));
    }

    public function testGetColumnLength(): void
    {
        $this->assertSame(500, GenUtils::getColumnLength('varchar(500)'));
        $this->assertSame(1, GenUtils::getColumnLength('char(1)'));
        $this->assertSame(0, GenUtils::getColumnLength('text'));
        $this->assertSame(0, GenUtils::getColumnLength('int'));
    }

    public function testToCamelCaseKeepsPrefixAndPlainName(): void
    {
        // 经典版 quirk：无下划线原样返回（不做首字母小写）
        $this->assertSame('status', GenUtils::toCamelCase('status'));
        $this->assertSame('ID', GenUtils::toCamelCase('ID'));
        // 不去表前缀
        $this->assertSame('sysUser', GenUtils::toCamelCase('sys_user'));
        $this->assertSame('noticeTitle', GenUtils::toCamelCase('notice_title'));
        $this->assertSame('createTime', GenUtils::toCamelCase('create_time'));
    }

    public function testConvertToCamelCase(): void
    {
        $this->assertSame('SysNotice', GenUtils::convertToCamelCase('sys_notice'));
        $this->assertSame('GenTableColumn', GenUtils::convertToCamelCase('gen_table_column'));
        $this->assertSame('Notice', GenUtils::convertToCamelCase('notice'));
        $this->assertSame('SysNotice', GenUtils::convertToCamelCase('SYS_NOTICE'));
        $this->assertSame('', GenUtils::convertToCamelCase(''));
    }

    public function testGetModuleNameAndBusinessName(): void
    {
        $this->assertSame('system', GenUtils::getModuleName('com.ruoyi.system'));
        $this->assertSame('ruoyi', GenUtils::getModuleName('com.ruoyi'));
        $this->assertSame('user', GenUtils::getBusinessName('sys_user'));
        $this->assertSame('notice', GenUtils::getBusinessName('notice'));
    }

    public function testReplaceTextStripsTableWordAndRuoyi(): void
    {
        $this->assertSame('通知公告', GenUtils::replaceText('通知公告表'));
        $this->assertSame('测试', GenUtils::replaceText('若依测试表'));
        $this->assertSame('用户信息', GenUtils::replaceText('用户信息'));
        $this->assertSame('', GenUtils::replaceText('表'));
    }

    public function testConvertClassNameHonoursAutoRemovePre(): void
    {
        // 经典版默认 autoRemovePre=false → 前缀保留在类名里
        $this->assertSame('SysNotice', GenUtils::convertClassName('sys_notice', false, 'sys_'));
        $this->assertSame('Notice', GenUtils::convertClassName('sys_notice', true, 'sys_'));
        // 多前缀逗号分隔
        $this->assertSame('Notice', GenUtils::convertClassName('sys_notice', true, 'biz_,sys_'));
        // 命中第一个匹配前缀即去前缀（经典版 replaceFirst + break），故类名不含 biz
        $this->assertSame('Order', GenUtils::convertClassName('biz_order', true, 'biz_,sys_'));
    }

    // ---------- initTable ----------

    public function testInitTableOnSysNotice(): void
    {
        $row = GenUtils::initTable([
            'table_name'    => 'sys_notice',
            'table_comment' => '通知公告表',
            'create_time'   => '2026-09-01 10:00:00',
            'update_time'   => '2026-09-01 10:00:00',
        ], 'admin', self::CFG);

        $this->assertSame('SysNotice', $row['class_name']);
        $this->assertSame('com.ruoyi.system', $row['package_name']);
        $this->assertSame('system', $row['module_name']);
        $this->assertSame('notice', $row['business_name']);
        $this->assertSame('通知公告', $row['function_name']);
        $this->assertSame('ruoyi', $row['function_author']);
        $this->assertSame('admin', $row['create_by']);
    }

    // ---------- initColumnField：类型映射四族 ----------

    /** @return array 单列夹具（模拟 information_schema.columns 行） */
    private function col(string $name, string $type, string $isPk = '0', string $isNullable = 'YES', string $extra = '', int $sort = 1): array
    {
        return [
            'column_name'    => $name,
            'column_type'    => $type,
            'column_comment' => $name . ' 注释',
            'is_required'    => ($isNullable === 'NO' && $isPk !== '1') ? '1' : null,
            'is_pk'          => $isPk,
            'is_increment'   => $extra === 'auto_increment' ? '1' : '0',
            'sort'           => $sort,
        ];
    }

    private function infer(array $column): array
    {
        return GenUtils::initColumnField($column, ['table_id' => 7, 'create_by' => 'admin']);
    }

    public function testStringFamilyMapping(): void
    {
        $c = $this->infer($this->col('notice_title', 'varchar(50)'));
        $this->assertSame('noticeTitle', $c['java_field']);
        $this->assertSame('String', $c['java_type']);
        $this->assertSame('input', $c['html_type']);
        $this->assertSame(7, $c['table_id']);
        $this->assertSame('admin', $c['create_by']);

        // 长度 ≥500 → 文本域
        $this->assertSame('textarea', $this->infer($this->col('big_text', 'varchar(500)'))['html_type']);
        $this->assertSame('input', $this->infer($this->col('small_text', 'varchar(499)'))['html_type']);

        // 文本族 → 文本域（无括号，长度 0 也走 textarea）
        foreach (['tinytext', 'text', 'mediumtext', 'longtext'] as $type) {
            $this->assertSame('textarea', $this->infer($this->col('x_col', $type))['html_type'], $type);
            $this->assertSame('String', $this->infer($this->col('x_col', $type))['java_type'], $type);
        }
    }

    public function testTimeFamilyMapping(): void
    {
        foreach (['datetime', 'time', 'date', 'timestamp'] as $type) {
            $c = $this->infer($this->col('create_time', $type));
            $this->assertSame('Date', $c['java_type'], $type);
            $this->assertSame('datetime', $c['html_type'], $type);
        }
    }

    public function testNumberFamilyMapping(): void
    {
        // 标度 > 0 → BigDecimal
        $this->assertSame('BigDecimal', $this->infer($this->col('price', 'decimal(10,2)'))['java_type']);
        // 单段宽度 ≤10 → Integer
        $this->assertSame('Integer', $this->infer($this->col('flag', 'int(1)'))['java_type']);
        // 单段宽度 >10 → Long
        $this->assertSame('Long', $this->infer($this->col('count', 'int(11)'))['java_type']);
        // 无括号（MySQL8 的 int/bigint 实况）→ Long
        $this->assertSame('Long', $this->infer($this->col('notice_id', 'int'))['java_type']);
        $this->assertSame('Long', $this->infer($this->col('user_id', 'bigint'))['java_type']);
        // 标度为 0 的两段 → Long（经典版 else 分支）
        $this->assertSame('Long', $this->infer($this->col('amount', 'decimal(10,0)'))['java_type']);
        // 数值族 htmlType 恒 input（随后可能被特判覆盖）
        $this->assertSame('input', $this->infer($this->col('count', 'int(11)'))['html_type']);
    }

    public function testUnlistedTypeFallsBackToDefaultStringInput(): void
    {
        // 经典版 quirk：blob 族不在四族内 → 默认 String；html_type 由特判链决定（无特判则遗留 null）
        $c = $this->infer($this->col('notice_content', 'longblob'));
        $this->assertSame('String', $c['java_type']);
        // content 结尾 → summernote（特判链覆盖）
        $this->assertSame('summernote', $c['html_type']);

        // 无任何特判的未列类型 → html_type 保持 null（照抄经典版：分支外不赋默认值）
        $this->assertNull($this->infer($this->col('raw_data', 'longblob'))['html_type']);
        $this->assertSame('String', $this->infer($this->col('raw_data', 'json'))['java_type']);
    }

    public function testHtmlTypeSpecialCases(): void
    {
        $this->assertSame('radio', $this->infer($this->col('status', 'char(1)'))['html_type']);
        $this->assertSame('select', $this->infer($this->col('notice_type', 'char(1)'))['html_type']);
        $this->assertSame('select', $this->infer($this->col('user_sex', 'char(1)'))['html_type']);
        $this->assertSame('upload', $this->infer($this->col('attach_file', 'varchar(100)'))['html_type']);
        $this->assertSame('summernote', $this->infer($this->col('notice_content', 'longblob'))['html_type']);
        // 后缀判定大小写不敏感
        $this->assertSame('radio', $this->infer($this->col('STATUS', 'char(1)'))['html_type']);
        // 特判链优先级：status 覆盖内容类后缀
        $this->assertSame('radio', $this->infer($this->col('file_status', 'char(1)'))['html_type']);
    }

    public function testQueryTypeLikeOnlyForNameSuffix(): void
    {
        $this->assertSame('LIKE', $this->infer($this->col('user_name', 'varchar(30)'))['query_type']);
        $this->assertSame('LIKE', $this->infer($this->col('NAME', 'varchar(30)'))['query_type']);
        $this->assertSame('EQ', $this->infer($this->col('notice_title', 'varchar(50)'))['query_type']);
        $this->assertSame('EQ', $this->infer($this->col('status', 'char(1)'))['query_type']);
    }

    public function testInsertEditListQueryToggles(): void
    {
        // 主键：is_insert 恒 1；edit/list/query 三者皆 null
        $pk = $this->infer($this->col('notice_id', 'int', '1', 'NO', 'auto_increment', 1));
        $this->assertSame('1', $pk['is_insert']);
        $this->assertNull($pk['is_edit']);
        $this->assertNull($pk['is_list']);
        $this->assertNull($pk['is_query']);

        // 普通业务列：四态全 1
        $title = $this->infer($this->col('notice_title', 'varchar(50)', '0', 'NO', '', 2));
        $this->assertSame('1', $title['is_insert']);
        $this->assertSame('1', $title['is_edit']);
        $this->assertSame('1', $title['is_list']);
        $this->assertSame('1', $title['is_query']);

        // create_by / create_time：not_edit + not_list + not_query → 仅 is_insert
        foreach (['create_by', 'create_time'] as $name) {
            $c = $this->infer($this->col($name, 'varchar(64)'));
            $this->assertSame('1', $c['is_insert'], $name);
            $this->assertNull($c['is_edit'], $name);
            $this->assertNull($c['is_list'], $name);
            $this->assertNull($c['is_query'], $name);
        }

        // update_by / update_time：可编辑、不进列表与查询
        foreach (['update_by', 'update_time'] as $name) {
            $c = $this->infer($this->col($name, 'varchar(64)'));
            $this->assertSame('1', $c['is_edit'], $name);
            $this->assertNull($c['is_list'], $name);
            $this->assertNull($c['is_query'], $name);
        }

        // remark：可编辑可列表，但不作查询条件
        $remark = $this->infer($this->col('remark', 'varchar(255)'));
        $this->assertSame('1', $remark['is_edit']);
        $this->assertSame('1', $remark['is_list']);
        $this->assertNull($remark['is_query']);

        // 名为 id 的非主键列：落入 NOT_EDIT/NOT_LIST/NOT_QUERY
        $idCol = $this->infer($this->col('id', 'bigint'));
        $this->assertNull($idCol['is_edit']);
        $this->assertNull($idCol['is_list']);
        $this->assertNull($idCol['is_query']);
    }

    /**
     * sys_notice 全 10 列端到端推断（ry-tp 库实测形态）
     * 这是导入链路的核心断言：逐列 java_type / html_type / 勾选态 / query_type。
     */
    public function testSysNoticeFullColumnInference(): void
    {
        $table = GenUtils::initTable([
            'table_name'    => 'sys_notice',
            'table_comment' => '通知公告表',
        ], 'admin', self::CFG);
        $table['table_id'] = 1;

        $fixture = [
            // name,             type,          isPk, nullable, extra,            sort
            ['notice_id',      'int',          '1',  'NO',     'auto_increment', 1],
            ['notice_title',   'varchar(50)',  '0',  'NO',     '',               2],
            ['notice_type',    'char(1)',      '0',  'NO',     '',               3],
            ['notice_content', 'longblob',     '0',  'YES',    '',               4],
            ['status',         'char(1)',      '0',  'YES',    '',               5],
            ['create_by',      'varchar(64)',  '0',  'YES',    '',               6],
            ['create_time',    'datetime',     '0',  'YES',    '',               7],
            ['update_by',      'varchar(64)',  '0',  'YES',    '',               8],
            ['update_time',    'datetime',     '0',  'YES',    '',               9],
            ['remark',         'varchar(255)', '0',  'YES',    '',               10],
        ];

        $inferred = [];
        foreach ($fixture as [$name, $type, $isPk, $nullable, $extra, $sort]) {
            $inferred[$name] = GenUtils::initColumnField(
                $this->col($name, $type, $isPk, $nullable, $extra, $sort),
                $table
            );
        }

        // 逐列关键属性（期望值由经典版 GenUtils 规则手工推导）
        $expect = [
            'notice_id'      => ['noticeId',      'Long',   'input',     null, null, null],
            'notice_title'   => ['noticeTitle',   'String', 'input',     '1',  '1',  '1'],
            'notice_type'    => ['noticeType',    'String', 'select',    '1',  '1',  '1'],
            'notice_content' => ['noticeContent', 'String', 'summernote','1',  '1',  '1'],
            'status'         => ['status',        'String', 'radio',     '1',  '1',  '1'],
            'create_by'      => ['createBy',      'String', 'input',     null, null, null],
            'create_time'    => ['createTime',    'Date',   'datetime',  null, null, null],
            'update_by'      => ['updateBy',      'String', 'input',     '1',  null, null],
            'update_time'    => ['updateTime',    'Date',   'datetime',  '1',  null, null],
            'remark'         => ['remark',        'String', 'input',     '1',  '1',  null],
        ];

        foreach ($expect as $name => [$field, $javaType, $htmlType, $isEdit, $isList, $isQuery]) {
            $c = $inferred[$name];
            $this->assertSame($field, $c['java_field'], "$name.java_field");
            $this->assertSame($javaType, $c['java_type'], "$name.java_type");
            $this->assertSame($htmlType, $c['html_type'], "$name.html_type");
            $this->assertSame('1', $c['is_insert'], "$name.is_insert");
            $this->assertSame($isEdit, $c['is_edit'], "$name.is_edit");
            $this->assertSame($isList, $c['is_list'], "$name.is_list");
            $this->assertSame($isQuery, $c['is_query'], "$name.is_query");
        }

        // information_schema 推断出的标记原样带下来（供 insert）
        $this->assertSame('1', $inferred['notice_id']['is_pk']);
        $this->assertSame('1', $inferred['notice_id']['is_increment']);
        $this->assertNull($inferred['notice_id']['is_required']);   // PRI 列 is_required 为 null
        $this->assertSame('1', $inferred['notice_title']['is_required']);
        $this->assertNull($inferred['status']['is_required']);
        $this->assertSame(1, $inferred['notice_id']['sort']);

        // 查询方式：仅 name 结尾为 LIKE，其余 EQ
        foreach ($inferred as $name => $c) {
            $this->assertSame('EQ', $c['query_type'], "$name.query_type");
        }
    }
}
