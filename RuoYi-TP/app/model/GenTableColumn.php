<?php
declare(strict_types=1);

namespace app\model;

use think\Model;

/**
 * 代码生成业务表字段模型（gen_table_column，对位经典版 GenTableColumn domain）
 */
class GenTableColumn extends Model
{
    protected $name = 'gen_table_column';
    protected $pk = 'column_id';

    protected $autoWriteTimestamp = false;
}
