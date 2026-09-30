<?php
require __DIR__ . '/../vendor/autoload.php';
$app = new think\App();
$app->initialize();

// 精确复现 selectDeptList 的 where 链 + DataScope 条件（单条件 dept_id in [103,...]）
$q = think\facade\Db::table('sys_dept d')
    ->leftJoin('sys_dept p', 'p.dept_id = d.parent_id')
    ->where('d.del_flag', '0')
    ->field('d.dept_id,d.dept_name');

// 模拟 DataScope 条件（不带别名验证）
$q->where(function ($sub) {
    $sub->where('d.dept_id', 'in', [103, 205, 206, 207, 208, 209]);
});
echo 'with closure: ', json_encode($q->column('d.dept_id')), PHP_EOL;

// 再试带 d. 前缀的列名在闭包里
$q2 = think\facade\Db::table('sys_dept d')
    ->leftJoin('sys_dept p', 'p.dept_id = d.parent_id')
    ->where('d.del_flag', '0')
    ->field('d.dept_id');
$q2->where(function ($sub) {
    $sub->whereOr('d.dept_id', 'in', [103, 205]);
});
echo 'with d-prefix in closure: ', json_encode($q2->column('d.dept_id')), PHP_EOL;
