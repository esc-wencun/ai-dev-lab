<?php
declare(strict_types=1);

namespace app\model;

use think\Model;

/**
 * 岗位表模型（sys_post，对位经典版 SysPost domain）
 */
class SysPost extends Model
{
    protected $name = 'sys_post';
    protected $pk = 'post_id';

    protected $autoWriteTimestamp = false;
}
