<?php
require __DIR__ . '/../vendor/autoload.php';
$app = new think\App();
$app->initialize();

// 最小复现：闭包内 whereIn 是否生效
$q = think\facade\Db::table('sys_dept d')->where('d.del_flag', '0')->field('d.dept_id');
$q->where(function ($sub) {
    $sub->whereIn('d.dept_id', [103, 205]);
});
echo 'closure whereIn: ', json_encode($q->column('d.dept_id')), PHP_EOL;

// 不带闭包
$q2 = think\facade\Db::table('sys_dept d')->where('d.del_flag', '0')->field('d.dept_id');
$q2->whereIn('d.dept_id', [103, 205]);
echo 'direct whereIn: ', json_encode($q2->column('d.dept_id')), PHP_EOL;

// 闭包 + whereOr in
$q3 = think\facade\Db::table('sys_dept d')->where('d.del_flag', '0')->field('d.dept_id');
$q3->where(function ($sub) {
    $sub->where('d.dept_id', 'in', [103])->whereOr('d.dept_id', 'in', [205]);
});
echo 'closure where+whereOr in: ', json_encode($q3->column('d.dept_id')), PHP_EOL;

// 条件数组形态 in 手工循环
$q4 = think\facade\Db::table('sys_dept d')->where('d.del_flag', '0')->field('d.dept_id');
$q4->where(function ($sub) {
    $conds = [['d.dept_id', 'in', [103, 205, 206]]];
    foreach ($conds as $i => $cond) {
        [$col, $op, $val] = $cond;
        if ($i === 0) { $sub->whereIn($col, $val); } else { $sub->whereOr($col, 'in', $val); }
    }
});
echo 'loop whereIn: ', json_encode($q4->column('d.dept_id')), PHP_EOL;
