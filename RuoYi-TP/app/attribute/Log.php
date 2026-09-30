<?php
declare(strict_types=1);

namespace app\attribute;

use Attribute;

/**
 * 操作日志注解（对位经典版 @Log(title, businessType)），声明在控制器方法上
 *
 * businessType 对位经典版 BusinessType 枚举序号：
 * 0其它 1新增 2修改 3删除 4授权 5导出 6导入 7强退 8清空
 */
#[\Attribute(\Attribute::TARGET_METHOD)]
final class Log
{
    public const OTHER  = 0;
    public const INSERT = 1;
    public const UPDATE = 2;
    public const DELETE = 3;
    public const GRANT  = 4;
    public const EXPORT = 5;
    public const IMPORT = 6;
    public const FORCE  = 7;
    /** 8=生成代码（经典版 GENCODE；10.0.0 勘误——原误定义为 CLEAN=8） */
    public const GENCODE = 8;
    /** 9=清空（经典版 CLEAN；10.0.0 勘误由 8 改 9，sys_oper_type 字典实锤 8=生成代码/9=清空数据） */
    public const CLEAN  = 9;

    public function __construct(
        public readonly string $title,
        public readonly int $businessType = self::OTHER,
    ) {
    }
}
