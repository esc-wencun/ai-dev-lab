<?php
declare(strict_types=1);

/**
 * 代码生成通用常量（对位经典版 com.ruoyi.common.constant.GenConstants）
 *
 * 纯常量类、零依赖：供 GenUtils 纯函数与 PHPUnit 直接 require（不走容器）。
 */
final class GenConstants
{
    /** 单表（增删改查） */
    public const TPL_CRUD = 'crud';

    /** 树表（增删改查） */
    public const TPL_TREE = 'tree';

    /** 主子表（增删改查） */
    public const TPL_SUB = 'sub';

    /** 树编码字段（options JSON 键） */
    public const TREE_CODE = 'treeCode';

    /** 树父编码字段（options JSON 键） */
    public const TREE_PARENT_CODE = 'treeParentCode';

    /** 树名称字段（options JSON 键） */
    public const TREE_NAME = 'treeName';

    /** 上级菜单ID字段（options JSON 键） */
    public const PARENT_MENU_ID = 'parentMenuId';

    /** 上级菜单名称字段（options JSON 键） */
    public const PARENT_MENU_NAME = 'parentMenuName';

    /** 生成详情页开关（options JSON 键） */
    public const GEN_VIEW = 'genView';

    /** 数据库字符串类型 */
    public const COLUMNTYPE_STR = ['char', 'varchar', 'nvarchar', 'varchar2'];

    /** 数据库文本类型 */
    public const COLUMNTYPE_TEXT = ['tinytext', 'text', 'mediumtext', 'longtext'];

    /** 数据库时间类型 */
    public const COLUMNTYPE_TIME = ['datetime', 'time', 'date', 'timestamp'];

    /** 数据库数字类型 */
    public const COLUMNTYPE_NUMBER = [
        'tinyint', 'smallint', 'mediumint', 'int', 'number', 'integer',
        'bit', 'bigint', 'float', 'double', 'decimal',
    ];

    /** 页面不需要编辑字段 */
    public const COLUMNNAME_NOT_EDIT = ['id', 'create_by', 'create_time', 'del_flag'];

    /** 页面不需要显示的列表字段 */
    public const COLUMNNAME_NOT_LIST = ['id', 'create_by', 'create_time', 'del_flag', 'update_by', 'update_time'];

    /** 页面不需要查询字段 */
    public const COLUMNNAME_NOT_QUERY = ['id', 'create_by', 'create_time', 'del_flag', 'update_by', 'update_time', 'remark'];

    /** 文本框 */
    public const HTML_INPUT = 'input';

    /** 文本域 */
    public const HTML_TEXTAREA = 'textarea';

    /** 下拉框 */
    public const HTML_SELECT = 'select';

    /** 单选框 */
    public const HTML_RADIO = 'radio';

    /** 复选框 */
    public const HTML_CHECKBOX = 'checkbox';

    /** 日期控件 */
    public const HTML_DATETIME = 'datetime';

    /** 上传控件 */
    public const HTML_UPLOAD = 'upload';

    /** 富文本控件 */
    public const HTML_SUMMERNOTE = 'summernote';

    /** 字符串类型（java_type 列照存经典版 Java 语义值——DB 列即契约，不做 PHP 类型映射） */
    public const TYPE_STRING = 'String';

    /** 整型 */
    public const TYPE_INTEGER = 'Integer';

    /** 长整型 */
    public const TYPE_LONG = 'Long';

    /** 高精度计算类型 */
    public const TYPE_BIGDECIMAL = 'BigDecimal';

    /** 时间类型 */
    public const TYPE_DATE = 'Date';

    /** 模糊查询 */
    public const QUERY_LIKE = 'LIKE';

    /** 相等查询 */
    public const QUERY_EQ = 'EQ';

    /** 需要（勾选态 '1'） */
    public const REQUIRE = '1';

    // 说明：经典版 GenConstants 另有 BASE_ENTITY / TREE_ENTITY / TYPE_DOUBLE 三组常量，
    // 仅供 Velocity 模板层使用。本模块范围拍板 A（specs/11.0.0）模板生成整段排除，
    // 故不照搬；若日后做模板生成再按经典版原文补齐。

    private function __construct()
    {
    }
}
