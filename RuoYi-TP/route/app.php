<?php
// +----------------------------------------------------------------------
// | RuoYi-TP 路由（URL 基准 = 经典若依）
// +----------------------------------------------------------------------
// 路由地址用「controller/action」斜杠风格（TP 会自动拼 app\controller\ 前缀）；
// 'Controller@action' @ 风格在 TP8 中不拼前缀，会报类不存在。
use think\facade\Route;

// ===== 登录链路（匿名；LoginAuth 中间件已放行这些路径） =====
Route::get('login', 'login/index');
Route::post('login', 'login/doLogin');
Route::get('logout', 'login/logout');
Route::get('captcha/captchaImage', 'login/captchaImage');

// ===== 注册（匿名；对位 SysRegisterController；开关只拦 POST） =====
Route::get('register', 'register/index');
Route::post('register', 'register/doRegister');

// ===== 锁屏（对位 SysIndexController.lockscreen/unlockscreen；软锁仅会话字段） =====
Route::get('lockscreen', 'index/lockscreen');
Route::post('unlockscreen', 'index/unlockscreen');

// ===== 监控日志（对位 SysUserOnline/Server/Cache/SysOperlog/SysLogininfor + 自研 data）——固定段先于 :param =====
Route::group('monitor/online', function () {
    Route::get('/', 'monitor.online/index');
    Route::post('list', 'monitor.online/list');
    Route::post('batchForceLogout', 'monitor.online/batchForceLogout');
});
Route::get('monitor/server', 'monitor.server/index');
Route::group('monitor/cache', function () {
    Route::get('/', 'monitor.cache/index');
    Route::post('getNames', 'monitor.cache/getNames');
    Route::post('getKeys', 'monitor.cache/getKeys');
    Route::post('getValue', 'monitor.cache/getValue');
    Route::post('clearCacheName', 'monitor.cache/clearCacheName');
    Route::post('clearCacheKey', 'monitor.cache/clearCacheKey');
    Route::get('clearAll', 'monitor.cache/clearAll');
});
Route::group('monitor/operlog', function () {
    Route::get('/', 'monitor.operlog/index');
    Route::post('list', 'monitor.operlog/list');
    Route::post('export', 'monitor.operlog/export');
    Route::post('remove', 'monitor.operlog/remove');
    Route::get('detail/:operId', 'monitor.operlog/detail');
    Route::post('clean', 'monitor.operlog/clean');
});
Route::group('monitor/logininfor', function () {
    Route::get('/', 'monitor.logininfor/index');
    Route::post('list', 'monitor.logininfor/list');
    Route::post('export', 'monitor.logininfor/export');
    Route::post('remove', 'monitor.logininfor/remove');
    Route::post('clean', 'monitor.logininfor/clean');
    Route::post('unlock', 'monitor.logininfor/unlock');
});
Route::get('monitor/data', 'monitor.data/index');

// ===== 定时任务（对位 SysJobController 14 路由 + SysJobLogController 6 路由）——固定段先于 :param =====
Route::group('monitor/job', function () {
    Route::get('/', 'monitor.job/index');
    Route::post('list', 'monitor.job/list');
    Route::post('export', 'monitor.job/export');
    Route::post('remove', 'monitor.job/remove');
    Route::post('changeStatus', 'monitor.job/changeStatus');
    Route::post('run', 'monitor.job/run');
    Route::get('add', 'monitor.job/add');
    Route::post('add', 'monitor.job/addSave');
    Route::get('edit/:jobId', 'monitor.job/edit');
    Route::post('edit', 'monitor.job/editSave');
    Route::post('checkCronExpressionIsValid', 'monitor.job/checkCronExpressionIsValid');
    Route::get('cron', 'monitor.job/cron');
    Route::get('queryCronExpression', 'monitor.job/queryCronExpression');
    Route::get('detail/:jobId', 'monitor.job/detail');
});
Route::group('monitor/jobLog', function () {
    Route::get('/', 'monitor.jobLog/index');
    Route::post('list', 'monitor.jobLog/list');
    Route::post('export', 'monitor.jobLog/export');
    Route::post('remove', 'monitor.jobLog/remove');
    Route::post('clean', 'monitor.jobLog/clean');
    Route::get('detail/:jobLogId', 'monitor.jobLog/detail');
});

// ===== 个人中心（对位 SysProfileController 7 路由；无 #[Perm] 仅登录态） =====
Route::group('system/user/profile', function () {
    Route::get('/', 'system.profile/index');
    Route::get('checkPassword', 'system.profile/checkPassword');
    Route::get('resetPwd', 'system.profile/resetPwd');
    Route::post('resetPwd', 'system.profile/resetPwdSave');
    Route::post('update', 'system.profile/update');
    Route::get('avatar', 'system.profile/avatar');
    Route::post('updateAvatar', 'system.profile/updateAvatar');
});

// ===== 主框架 =====
Route::get('index', 'index/index');
Route::get('system/main', 'index/main');
Route::get('system/switchSkin', 'index/switchSkin');
Route::get('system/menuStyle/:style', 'index/menuStyle');
Route::get('unauth', 'index/unauth');

// ===== 通知公告（对位 SysNoticeController 13 方法）——固定段路由先于 :param 通配（防遮蔽） =====
Route::group('system/notice', function () {
    Route::get('/', 'system.notice/index');
    Route::post('list', 'system.notice/list');
    Route::get('add', 'system.notice/add');
    Route::post('add', 'system.notice/addSave');
    Route::get('edit/:noticeId', 'system.notice/edit');
    Route::post('edit', 'system.notice/editSave');
    Route::post('remove', 'system.notice/remove');
    // listTop/markRead/markReadAll/view 无 #[Perm]（仅登录态，普通用户读公告链路——经典版实锤）
    Route::get('listTop', 'system.notice/listTop');
    Route::post('markRead', 'system.notice/markRead');
    Route::post('markReadAll', 'system.notice/markReadAll');
    Route::get('view/:noticeId', 'system.notice/view');
    Route::get('readUsers/:noticeId', 'system.notice/readUsers');
    Route::post('readUsers/list', 'system.notice/readUsersList');
});

// ===== 通用上传（summernote 图片上传；对位 CommonController.uploadFile + /profile 静态映射） =====
Route::post('common/upload', 'common/upload');
Route::get('upload/:fileName', 'common/serveUpload');

// ===== 字典类型（对位 SysDictTypeController 13 方法）——固定段路由先于 :param 通配（防遮蔽） =====
Route::group('system/dict', function () {
    Route::get('/', 'system.dictType/index');
    Route::post('list', 'system.dictType/list');
    Route::post('export', 'system.dictType/export');
    Route::post('add', 'system.dictType/addSave');
    Route::post('edit', 'system.dictType/editSave');
    Route::post('remove', 'system.dictType/remove');
    Route::post('checkDictTypeUnique', 'system.dictType/checkDictTypeUnique');
    Route::get('refreshCache', 'system.dictType/refreshCache');
    Route::get('detail/:dictId', 'system.dictType/detail');
    Route::get('treeData', 'system.dictType/treeData');
    Route::get('selectDictTree/:columnId/:dictType', 'system.dictType/selectDictTree');
    Route::get('add', 'system.dictType/add');
    Route::get('edit/:dictId', 'system.dictType/edit');
});

// ===== 字典数据（对位 SysDictDataController 8 方法） =====
Route::group('system/dict/data', function () {
    Route::get('/', 'system.dictData/index');
    Route::post('list', 'system.dictData/list');
    Route::post('export', 'system.dictData/export');
    Route::post('add', 'system.dictData/addSave');
    Route::post('edit', 'system.dictData/editSave');
    Route::post('remove', 'system.dictData/remove');
    Route::get('add/:dictType', 'system.dictData/add');
    Route::get('edit/:dictCode', 'system.dictData/edit');
});

// ===== 参数配置（对位 SysConfigController 10 方法） =====
Route::group('system/config', function () {
    Route::get('/', 'system.config/index');
    Route::post('list', 'system.config/list');
    Route::post('export', 'system.config/export');
    Route::get('add', 'system.config/add');
    Route::post('add', 'system.config/addSave');
    Route::get('edit/:configId', 'system.config/edit');
    Route::post('edit', 'system.config/editSave');
    Route::post('remove', 'system.config/remove');
    Route::post('checkConfigKeyUnique', 'system.config/checkConfigKeyUnique');
    Route::get('refreshCache', 'system.config/refreshCache');
});

// ===== 部门管理（对位 SysDeptController 11 方法）——CheckPerm/OperLog 在 config/route.php route 管线统一挂载 =====
Route::group('system/dept', function () {
    Route::get('/', 'system.dept/index');
    Route::post('list', 'system.dept/list');
    Route::get('add/:parentId', 'system.dept/add');
    Route::post('add', 'system.dept/addSave');
    Route::get('edit/:deptId', 'system.dept/edit');
    Route::post('edit', 'system.dept/editSave');
    Route::post('updateSort', 'system.dept/updateSort');
    Route::post('remove/:deptId', 'system.dept/remove');
    Route::post('checkDeptNameUnique', 'system.dept/checkDeptNameUnique');
    Route::get('selectDeptTree/:deptId', 'system.dept/selectDeptTree');
    Route::get('selectDeptTree/:deptId/:excludeId', 'system.dept/selectDeptTree');
    Route::get('treeData/:excludeId', 'system.dept/treeData');
});

// ===== 岗位管理（对位 SysPostController 10 方法） =====
Route::group('system/post', function () {
    Route::get('/', 'system.post/index');
    Route::post('list', 'system.post/list');
    Route::post('export', 'system.post/export');
    Route::get('add', 'system.post/add');
    Route::post('add', 'system.post/addSave');
    Route::get('edit/:postId', 'system.post/edit');
    Route::post('edit', 'system.post/editSave');
    Route::post('remove', 'system.post/remove');
    Route::post('checkPostNameUnique', 'system.post/checkPostNameUnique');
    Route::post('checkPostCodeUnique', 'system.post/checkPostCodeUnique');
});

// ===== 通用下载 =====
Route::get('common/download', 'common/download');

// ===== 用户管理（对位 SysUserController 21 方法） =====
Route::group('system/user', function () {
    Route::get('/', 'system.user/index');
    Route::post('list', 'system.user/list');
    Route::post('export', 'system.user/export');
    Route::post('importData', 'system.user/importData');
    Route::post('importTemplate', 'system.user/importTemplate');
    Route::get('add', 'system.user/add');
    Route::post('add', 'system.user/addSave');
    Route::get('edit/:userId', 'system.user/edit');
    Route::post('edit', 'system.user/editSave');
    Route::get('view/:userId', 'system.user/view');
    Route::get('resetPwd/:userId', 'system.user/resetPwd');
    Route::post('resetPwd', 'system.user/resetPwdSave');
    Route::get('authRole/:userId', 'system.user/authRole');
    Route::post('authRole/insertAuthRole', 'system.user/insertAuthRole');
    Route::post('remove', 'system.user/remove');
    Route::post('checkLoginNameUnique', 'system.user/checkLoginNameUnique');
    Route::post('checkPhoneUnique', 'system.user/checkPhoneUnique');
    Route::post('checkEmailUnique', 'system.user/checkEmailUnique');
    Route::post('changeStatus', 'system.user/changeStatus');
    Route::get('deptTreeData', 'system.user/deptTreeData');
    Route::get('selectDeptTree/:deptId', 'system.user/selectDeptTree');
});

// ===== 角色管理（对位 SysRoleController 23 方法；selectMenuTree 死端点不注册） =====
Route::group('system/role', function () {
    Route::get('/', 'system.role/index');
    Route::post('list', 'system.role/list');
    Route::post('export', 'system.role/export');
    Route::get('add', 'system.role/add');
    Route::post('add', 'system.role/addSave');
    Route::get('edit/:roleId', 'system.role/edit');
    Route::post('edit', 'system.role/editSave');
    Route::get('authDataScope/:roleId', 'system.role/authDataScope');
    Route::post('authDataScope', 'system.role/authDataScopeSave');
    Route::post('remove', 'system.role/remove');
    Route::post('checkRoleNameUnique', 'system.role/checkRoleNameUnique');
    Route::post('checkRoleKeyUnique', 'system.role/checkRoleKeyUnique');
    Route::post('changeStatus', 'system.role/changeStatus');
    // 注意顺序：authUser 固定段（selectUser/allocatedList 等）必须先于 authUser/:roleId 通配（防路由遮蔽）
    Route::get('authUser/selectUser/:roleId', 'system.role/selectUser');
    Route::post('authUser/allocatedList', 'system.role/allocatedList');
    Route::post('authUser/unallocatedList', 'system.role/unallocatedList');
    Route::post('authUser/cancel', 'system.role/cancel');
    Route::post('authUser/cancelAll', 'system.role/cancelAll');
    Route::post('authUser/selectAll', 'system.role/selectAll');
    Route::get('authUser/:roleId', 'system.role/authUser');
    Route::get('deptTreeData', 'system.role/deptTreeData');
    Route::get('view/:roleId', 'system.role/view');
});

// ===== 菜单管理（对位 SysMenuController 13 方法） =====
Route::group('system/menu', function () {
    Route::get('/', 'system.menu/index');
    Route::post('list', 'system.menu/list');
    Route::post('remove/:menuId', 'system.menu/remove');
    Route::get('add/:parentId', 'system.menu/add');
    Route::post('add', 'system.menu/addSave');
    Route::get('edit/:menuId', 'system.menu/edit');
    Route::post('edit', 'system.menu/editSave');
    Route::post('updateSort', 'system.menu/updateSort');
    Route::get('icon', 'system.menu/icon');
    Route::post('checkMenuNameUnique', 'system.menu/checkMenuNameUnique');
    Route::get('roleMenuTreeData', 'system.menu/roleMenuTreeData');
    Route::get('menuTreeData', 'system.menu/menuTreeData');
    Route::get('selectMenuTree/:menuId', 'system.menu/selectMenuTree');
});

// ===== 代码生成（对位 GenController 16 路由中的 8 条数据层端点）=====
// 模板生成链（GET edit 渲染 / preview / download / genCode / batchGenCode / createTable /
// synchDb）按 11.0.0 范围拍板 A 有意排除：路由不注册即 404（deviations #21）。
Route::group('tool/gen', function () {
    Route::get('/', 'tool.gen/index');
    Route::post('list', 'tool.gen/list');
    Route::post('db/list', 'tool.gen/dbList');
    Route::post('column/list', 'tool.gen/columnList');
    Route::get('importTable', 'tool.gen/importTable');
    Route::post('importTable', 'tool.gen/importTableSave');
    Route::post('edit', 'tool.gen/editSave');
    Route::post('remove', 'tool.gen/remove');
});
