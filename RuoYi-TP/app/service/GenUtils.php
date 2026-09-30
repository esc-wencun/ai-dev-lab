<?php
declare(strict_types=1);

namespace app\service;

use GenConstants;

/**
 * 代码生成推断工具（对位经典版 com.ruoyi.generator.util.GenUtils）
 *
 * 全部为**纯函数**：不碰 Db / Redis / config，配置值（author/packageName/autoRemovePre/
 * tablePrefix）由调用方 GenService 从 config/gen.php 读出后传入。这样 PHPUnit 可直接
 * require 本文件做全规则断言，不启动 TP 容器、不连库（对位 MenuServiceTest 范式）。
 *
 * 逐条对齐经典版源码（含三处刻意照抄的 quirk，勿"顺手改进"）：
 *   1. toCamelCase 不含下划线时**原样返回**（不做首字母小写），且不去表前缀；
 *   2. 类型映射四族不含 blob 族与未列类型 → 落默认 String + input，再由 htmlType 特判覆盖；
 *   3. htmlType 特判是 else-if 链（status 优先于 type/sex 优先于 file 优先于 content）。
 */
final class GenUtils
{
    /** 数组包含指定值（对位 GenUtils.arraysContains：精确匹配，故 "int unsigned" 不命中 "int"） */
    public static function arraysContains(array $arr, string $targetValue): bool
    {
        return in_array($targetValue, $arr, true);
    }

    /** 取数据库类型（截断括号部分；对位 GenUtils.getDbType） */
    public static function getDbType(string $columnType): string
    {
        $pos = strpos($columnType, '(');
        return ($pos !== false && $pos > 0) ? substr($columnType, 0, $pos) : $columnType;
    }

    /** 取字段长度（无括号返回 0；对位 GenUtils.getColumnLength） */
    public static function getColumnLength(string $columnType): int
    {
        $pos = strpos($columnType, '(');
        if ($pos === false || $pos <= 0) {
            return 0;
        }
        $end = strpos($columnType, ')', $pos);
        if ($end === false) {
            return 0;
        }
        $length = substr($columnType, $pos + 1, $end - $pos - 1);
        // 经典版此处 Integer.valueOf(length)：字符串族括号内必为单个长度，故行为一致；
        // 非数字（如 decimal(10,2) 的 "10,2"）在经典版会抛 NumberFormatException，
        // 但该调用只在字符串/文本族可达（数值族走 columnScaleParts），此处取 0 兜底。
        return is_numeric($length) ? (int)$length : 0;
    }

    /** java 字段名驼峰（对位 StringUtils.toCamelCase：不去前缀；无下划线原样返回） */
    public static function toCamelCase(string $s): string
    {
        if (strpos($s, '_') === false) {
            return $s;
        }
        $s = strtolower($s);
        $out = '';
        $upper = false;
        $len = strlen($s);
        for ($i = 0; $i < $len; $i++) {
            $c = $s[$i];
            if ($c === '_') {
                $upper = true;
            } elseif ($upper) {
                $out .= strtoupper($c);
                $upper = false;
            } else {
                $out .= $c;
            }
        }
        return $out;
    }

    /** 表名转类名驼峰（对位 StringUtils.convertToCamelCase：每段首字母大写 + 其余小写；无下划线仅首字母大写） */
    public static function convertToCamelCase(string $name): string
    {
        if ($name === '') {
            return '';
        }
        if (strpos($name, '_') === false) {
            return strtoupper(substr($name, 0, 1)) . substr($name, 1);
        }
        $out = '';
        foreach (explode('_', $name) as $camel) {
            if ($camel === '') {
                continue; // 跳过开头/结尾下划线或连续下划线
            }
            $out .= strtoupper(substr($camel, 0, 1)) . strtolower(substr($camel, 1));
        }
        return $out;
    }

    /** 取模块名（packageName 最后一个点之后；对位 GenUtils.getModuleName） */
    public static function getModuleName(string $packageName): string
    {
        $pos = strrpos($packageName, '.');
        return $pos === false ? $packageName : substr($packageName, $pos + 1);
    }

    /** 取业务名（表名最后一个下划线之后；对位 GenUtils.getBusinessName） */
    public static function getBusinessName(string $tableName): string
    {
        $pos = strrpos($tableName, '_');
        return $pos === false ? $tableName : substr($tableName, $pos + 1);
    }

    /** 关键字替换：去掉「表」与「若依」（对位 RegExUtils.replaceAll(text, "(?:表|若依)", "")） */
    public static function replaceText(string $text): string
    {
        return (string)preg_replace('/(?:表|若依)/u', '', $text);
    }

    /** 表名转类名（对位 GenUtils.convertClassName：autoRemovePre 为真且前缀非空时先去前缀，再转驼峰） */
    public static function convertClassName(string $tableName, bool $autoRemovePre, string $tablePrefix): string
    {
        if ($autoRemovePre && $tablePrefix !== '') {
            foreach (explode(',', $tablePrefix) as $search) {
                if ($search !== '' && str_starts_with($tableName, $search)) {
                    $tableName = substr($tableName, strlen($search));
                    break;
                }
            }
        }
        return self::convertToCamelCase($tableName);
    }

    /**
     * 初始化表信息（对位 GenUtils.initTable）
     *
     * @param array $table  从 information_schema 读出的行（含 table_name / table_comment / create_time / update_time）
     * @param array $config config/gen.php 的五项配置
     */
    public static function initTable(array $table, string $operName, array $config): array
    {
        $tableName = (string)$table['table_name'];
        $packageName = (string)($config['packageName'] ?? '');
        $table['class_name'] = self::convertClassName(
            $tableName,
            (bool)($config['autoRemovePre'] ?? false),
            (string)($config['tablePrefix'] ?? '')
        );
        $table['package_name'] = $packageName;
        $table['module_name'] = self::getModuleName($packageName);
        $table['business_name'] = self::getBusinessName($tableName);
        $table['function_name'] = self::replaceText((string)($table['table_comment'] ?? ''));
        $table['function_author'] = (string)($config['author'] ?? '');
        $table['create_by'] = $operName;
        return $table;
    }

    /**
     * 初始化列属性（对位 GenUtils.initColumnField，规则逐条照抄）
     *
     * 未被规则命中的属性显式置 null（对位经典版实体属性为 null → 插入时该列不写、落 DB 默认值）。
     * 输入行来自 information_schema.columns（column_name / column_type / column_comment /
     * is_required / is_pk / is_increment / sort）。
     *
     * @param array $table 需含 table_id 与 create_by（由 initTable 后的主表行提供）
     */
    public static function initColumnField(array $column, array $table): array
    {
        $columnType = (string)$column['column_type'];
        $dataType = self::getDbType($columnType);
        $columnName = (string)$column['column_name'];

        $column['table_id'] = $table['table_id'];
        $column['create_by'] = $table['create_by'];

        // java 字段名（驼峰，不去前缀）
        $column['java_field'] = self::toCamelCase($columnName);
        // 默认类型
        $column['java_type'] = GenConstants::TYPE_STRING;
        $column['query_type'] = GenConstants::QUERY_EQ;
        $column['html_type'] = null;

        if (
            self::arraysContains(GenConstants::COLUMNTYPE_STR, $dataType)
            || self::arraysContains(GenConstants::COLUMNTYPE_TEXT, $dataType)
        ) {
            // 字符串长度 ≥500 或文本类型 → 文本域
            $columnLength = self::getColumnLength($columnType);
            $isText = self::arraysContains(GenConstants::COLUMNTYPE_TEXT, $dataType);
            $column['html_type'] = ($columnLength >= 500 || $isText)
                ? GenConstants::HTML_TEXTAREA
                : GenConstants::HTML_INPUT;
        } elseif (self::arraysContains(GenConstants::COLUMNTYPE_TIME, $dataType)) {
            $column['java_type'] = GenConstants::TYPE_DATE;
            $column['html_type'] = GenConstants::HTML_DATETIME;
        } elseif (self::arraysContains(GenConstants::COLUMNTYPE_NUMBER, $dataType)) {
            $column['html_type'] = GenConstants::HTML_INPUT;

            // 浮点型统一 BigDecimal；标度缺失时按宽度：≤10 → Integer，否则 Long
            $parts = self::columnScaleParts($columnType);
            if ($parts !== null && count($parts) === 2 && (int)$parts[1] > 0) {
                $column['java_type'] = GenConstants::TYPE_BIGDECIMAL;
            } elseif ($parts !== null && count($parts) === 1 && (int)$parts[0] <= 10) {
                $column['java_type'] = GenConstants::TYPE_INTEGER;
            } else {
                $column['java_type'] = GenConstants::TYPE_LONG;
            }
        }

        // 插入字段（默认所有字段都需要插入）
        $column['is_insert'] = GenConstants::REQUIRE;
        // 编辑字段
        $column['is_edit'] = (!self::arraysContains(GenConstants::COLUMNNAME_NOT_EDIT, $columnName) && !self::isPk($column))
            ? GenConstants::REQUIRE : null;
        // 列表字段
        $column['is_list'] = (!self::arraysContains(GenConstants::COLUMNNAME_NOT_LIST, $columnName) && !self::isPk($column))
            ? GenConstants::REQUIRE : null;
        // 查询字段
        $column['is_query'] = (!self::arraysContains(GenConstants::COLUMNNAME_NOT_QUERY, $columnName) && !self::isPk($column))
            ? GenConstants::REQUIRE : null;

        // 查询字段类型：name 结尾 → LIKE
        if (self::endsWithIgnoreCase($columnName, 'name')) {
            $column['query_type'] = GenConstants::QUERY_LIKE;
        }

        // html 类型特判（else-if 链：status > type/sex > file > content，照抄勿改为独立 if）
        if (self::endsWithIgnoreCase($columnName, 'status')) {
            $column['html_type'] = GenConstants::HTML_RADIO;
        } elseif (self::endsWithIgnoreCase($columnName, 'type') || self::endsWithIgnoreCase($columnName, 'sex')) {
            $column['html_type'] = GenConstants::HTML_SELECT;
        } elseif (self::endsWithIgnoreCase($columnName, 'file')) {
            $column['html_type'] = GenConstants::HTML_UPLOAD;
        } elseif (self::endsWithIgnoreCase($columnName, 'content')) {
            $column['html_type'] = GenConstants::HTML_SUMMERNOTE;
        }

        return $column;
    }

    /** 主键判据（对位 GenTableColumn.isPk()：仅 '1' 为真） */
    public static function isPk(array $column): bool
    {
        return ($column['is_pk'] ?? null) === '1';
    }

    /** 大小写不敏感后缀判定（对位 StringUtils.endsWithIgnoreCase） */
    public static function endsWithIgnoreCase(string $haystack, string $needle): bool
    {
        return str_ends_with(strtolower($haystack), strtolower($needle));
    }

    /**
     * 数值类型的括号内精度/标度切分
     * （对位 StringUtils.split(StringUtils.substringBetween(columnType, "(", ")"), ",")：
     *   无括号 → null；空串 → 空数组）
     */
    public static function columnScaleParts(string $columnType): ?array
    {
        $pos = strpos($columnType, '(');
        if ($pos === false) {
            return null;
        }
        $end = strpos($columnType, ')', $pos);
        if ($end === false) {
            return null;
        }
        $inner = substr($columnType, $pos + 1, $end - $pos - 1);
        return $inner === '' ? [] : explode(',', $inner);
    }
}
