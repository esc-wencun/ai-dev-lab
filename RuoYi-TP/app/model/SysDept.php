<?php
declare(strict_types=1);

namespace app\model;

use think\Model;

/**
 * 部门表模型（sys_dept，对位经典版 SysDept domain）
 *
 * 字段名 = DB 下划线列名（think-orm 原样）；时间输出 Y-m-d H:i:s（全局 datetime_format）。
 * 仅供 service 层数据访问；业务判断在 DeptService。
 */
class SysDept extends Model
{
    protected $name = 'sys_dept';
    protected $pk = 'dept_id';

    // 关闭自动时间戳（经典版由 base_entity 手动管理；本项目 service 显式写 create_time）
    protected $autoWriteTimestamp = false;
}
