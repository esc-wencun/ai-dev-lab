<?php
declare(strict_types=1);

namespace app\attribute;

use Attribute;

/**
 * 权限校验注解（对位经典版 @RequiresPermissions），声明在控制器方法上
 */
#[\Attribute(\Attribute::TARGET_METHOD | \Attribute::IS_REPEATABLE)]
final class Perm
{
    public function __construct(public readonly string $value)
    {
    }
}
