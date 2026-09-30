<?php
require __DIR__ . '/../vendor/autoload.php';

$app = new think\App();
$app->initialize();

$user = [
    'userId' => 105, 'deptId' => 103, 'isAdmin' => false,
    'roles' => [['role_id' => 102, 'data_scope' => '4']],
];

// 复制 selectDeptList 主体但捕获 SQL
$query = think\facade\Db::table('sys_dept d')
    ->leftJoin('sys_dept p', 'p.dept_id = d.parent_id')
    ->where('d.del_flag', '0')
    ->field('d.' . str_replace(',', ',d.', 'dept_id,parent_id,ancestors,dept_name,order_num,leader,phone,email,status,del_flag,create_by,create_time,update_by,update_time'));
\DataScope::apply($query, $user, 'd');
$rows = $query->order('d.parent_id,d.order_num')->select()->toArray();
echo 'rows: ', count($rows), PHP_EOL;
// 打印 SQL
echo think\facade\Db::getLastSql(), PHP_EOL;
