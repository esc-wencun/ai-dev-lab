<?php
require __DIR__ . '/../vendor/autoload.php';
$app = new think\App();
$app->initialize();

// 模拟 isouser 会话
$user = [
    'userId' => 105, 'deptId' => 103, 'isAdmin' => false,
    'permissions' => ['system:dept:list', 'system:dept:view'],
    'roles' => [['role_id' => 102, 'role_key' => 'iso_role', 'role_name' => '隔离验证角色', 'data_scope' => '4']],
];

// 用真实控制器同款查询
$rows = app\service\DeptService::selectDeptList([], $user);
echo 'rows: ', count($rows), PHP_EOL;
foreach ($rows as $r) echo '  ', $r['dept_id'], ' ', $r['dept_name'], PHP_EOL;

// 对照：DataScope::apply 直查
$q = think\facade\Db::table('sys_dept d')->where('d.del_flag', '0')->field('d.dept_id');
$q->where('d.dept_id', 'in', [103, 205, 206, 207, 208, 209]);
echo 'raw in-query dept ids: ', json_encode($q->column('d.dept_id')), PHP_EOL;
