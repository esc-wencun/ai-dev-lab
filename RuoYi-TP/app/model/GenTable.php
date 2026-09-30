<?php
declare(strict_types=1);

namespace app\model;

use think\Model;

/**
 * 代码生成业务表模型（gen_table，对位经典版 GenTable domain）
 *
 * 薄模型：查询与写入统一走 GenService 的 Db 查询构造器（与 PostService/DictService 同范式），
 * 本类只声明表名与主键，供需要模型语义的场景使用。
 */
class GenTable extends Model
{
    protected $name = 'gen_table';
    protected $pk = 'table_id';

    protected $autoWriteTimestamp = false;
}
