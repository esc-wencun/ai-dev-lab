<?php
declare(strict_types=1);

namespace app\model;

use think\Model;

/**
 * 用户表模型（sys_user，对位经典版 SysUser domain）
 *
 * sys_user_role / sys_user_post 复合主键无自增——不建模型，UserService 内 Db::table 直操作。
 */
class SysUser extends Model
{
    protected $name = 'sys_user';
    protected $pk = 'user_id';

    protected $autoWriteTimestamp = false;
}
