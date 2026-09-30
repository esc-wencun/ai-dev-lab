# 3.0.0-部门岗位 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：模型/service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。（2026-09-29 动工）

## Task 1 · 域模型与表访问层

- [x] `app/model/SysDept.php`（think-orm 模型，表 sys_dept；autoWriteTimestamp 关闭，时间由 service 显式写）
- [x] `app/model/SysPost.php`（表 sys_post）
- [x] 模型层薄（纯表映射），字段/时间逻辑并入 Task 2/3 的服务测试覆盖——独立单测无测算面，不单独设断言

## Task 2 · 部门服务 DeptService

- [x] selectDeptList(filter, user)：del_flag='0' + like/status/deptId/parentId + DataScope(admin 外) + order by parent_id, order_num
- [x] selectDeptById（parentName 子查询；deptId=100 →「无」）；selectDeptTreeData(excludeId, user)：status='0' + 排除自身及 ancestors 含 excludeId + Ztree 结构 {id,pId,name,title,checked,open,nocheck}
- [x] checkDeptNameUnique（同父同名 del_flag='0'，自身放行）
- [x] selectDeptCount / checkDeptExistUser / selectNormalChildrenDeptById（find_in_set）
- [x] insertDept：父停用 →「部门停用，不允许新增」+ ancestors=父.ancestors,parent_id + create_by/time——**踩坑：Db::getLastInsID() 需 PDO 参数，TP8 门面调用报错；改 max(dept_id)**
- [x] updateDept（事务）：ancestors 级联用 SQL 层 REPLACE(CONCAT(ancestors,','), 'old,', 'new,')-1 前缀锚定（规避 "0,10" 误伤 "0,101"）+ 启用链上级置 '0'
- [x] deleteDeptById 软删；updateDeptSort（事务，异常 → 「保存排序异常，请联系管理员」）
- [x] checkDeptDataScope：非 admin 数据范围外 → 「没有权限访问部门数据！」
- [x] PHPUnit：Ztree excludeId 过滤口径 + ancestors 边界（"0,101" 不命中 10）固化；级联的 DB 级核对在端到端完成（改父后 SELECT 核验 202 行 ancestors 0,100,101,103 → 0,100,102 实测通过）

## Task 3 · 岗位服务 PostService

- [x] selectPostList（like 过滤 + PageQuery 白名单排序）——**返回 Query 构造器**（分页/全量由调用方决定）；**踩坑：返回类型误写 array 报 TypeError**
- [x] selectPostById；checkPostNameUnique / checkPostCodeUnique（全局唯一，postId 自身放行）
- [x] insertPost / updatePost（createBy/updateBy/时间）
- [x] deletePostByIds：先逐个查 sys_user_post 占用（整批拒绝）→ 物理 delete in
- [x] PHPUnit：ids 解析；唯一口径端到端验证（董事长→false、测试岗→true）

## Task 4 · Excel 导出与通用下载

- [x] composer 引入 phpoffice/phpspreadsheet ^2.0——**踩坑：2.0 移除 getCellByColumnAndRow，改 getCell([col,row]) 坐标数组**
- [x] ExcelExportService：列定义驱动 + 列头加粗灰底（样式复刻从简，**deviations 候选定案：从简版两样式，登记 deviations**）+ 数字格式 + readConverterExp + `<uuid>_<sheet>.xlsx` 输出 runtime/download
- [x] SysPost 5 列（spec 表逐列）；**踩坑：runtime_path() 框架 helper 在 PHPUnit 不可用，服务内 downloadDir() 自解析**
- [x] GET /common/download：realFileName=timestamp+去uuid前缀；octet-stream+attachment URL编码文件名；delete=true 下载后删（实测导出→下载→目录清零）；防 ".."（`{"code":500,"msg":"非法文件名"}`）+ 扩展名白名单（`{"code":500,"msg":"不支持的文件类型"}`）——注意 BusinessException 信封是 HTTP 200 + code 500，非 HTTP 500
- [x] PHPUnit：文件名解析（uuid 剥离+timestamp 前缀）、状态转换、**真实 xlsx 生成+读回断言（PK 魔数/sheet 名/5 列列头/正常停用转换）**

## Task 5 · 字典最小只读服务 DictService

- [x] listByType：sys_dict_data（status='0'，order by dict_sort）→ Redis `dict:<type>` 缓存
- [x] 输出含 dictLabel/dictValue/listClass/cssClass（页面徽章消费）
- [x] 6.0.0 收编注释已写入类注释

## Task 6 · 控制器与路由

- [x] DeptController 11 方法（selectDeptTree 可选参数合并双 pattern）+ #[Perm] 10 + #[Log] 4；非 admin add 强制 parentId、deptId=100 parentName「无」、remove warn 301 两连、check 裸 bool
- [x] PostController 10 方法 + #[Perm] 8 + #[Log] 4；双唯一校验新增/修改前缀区分；export 文件名装 msg
- [x] 路由注册：dept 12 条 + post 10 条 + /common/download
- [x] OperLog 注解生效（#[Log] 中间件反射）；**未挂 RepeatSubmit**（经典版实锤）
- [x] curl 级全端点自测（UTF-8 Python 脚本）：dept/list 裸数组 ✓、post/list TableDataInfo ✓、check 裸 bool ✓、remove 103 有子 → **301**「存在下级部门,不允许删除」✓、同父重名 500 完整文案 ✓、异父同名成功 ✓、edit self-parent「上级部门不能是自己」✓、export→download 链路 ✓——**踩坑三连：全局类（PageQuery/TableDataInfo/ExcelExportService）在 namespaced 控制器必须 use；paginate()->items() 返回 array 非 Collection；BusinessException 信封 HTTP 200 非 500（测试断言曾写错）**

## Task 7 · 页面模板（7 页）（2026-09-29 完成）

- [x] 子任务交付 7 页（dept index/add/edit/tree + post index/add/edit），静态自查通过（无 th:*/@{ 残留、check_perm 全部用 {:} 输出形式、volist/if 配平、键名下划线）
- [x] **联调返工四处**（已全修）：① 控制器渲染路径——layer 命名空间（system.DeptController → 视图根 app/view/system/），模板路径改 layer 相对（'dept/index'）；② include 片段——think-template include 的 / 前缀也按 view_path 拼，全局 app/view/include/ 复制为 app/view/system/include/（layer 局部副本）；③ DictService 双键输出——index 页 JS 消费驼峰（dictValue/dictLabel，对位 Jackson）、add/edit radio 消费下划线（dict_value/is_default），并存两套键；④ dept/post list 输出驼峰键——bootstrap-table columns field: deptName/postId 等按驼峰取值（对位经典版实体 Jackson 形态），控制器组装驼峰行

## Task 8 · 端到端验收（浏览器级）（2026-09-29 完成）

- [x] 部门页：主框架「部门管理」标签 + iframe 打开；树表 10 行全渲染（若依科技→深圳总公司→研发/市场/测试/财务/运维 + 长沙分公司，层级正确）；状态徽章「正常」；工具栏四按钮（新增/修改/保存排序/展开折叠）
- [x] 部门新增：弹窗字段齐全（parentId 隐藏域+treeName 预填「若依科技」/deptName/orderNum/leader/phone/email/status 默认正常）→ 提交 → DB 落库（ancestors=0,100、create_by=admin）
- [x] 部门编辑：弹窗回显正确（deptName/orderNum/parentId/treeName）→ 改名保存 → DB update_by=admin 落库
- [x] 部门删除：确认框 → del_flag='2' 软删 DB 核对
- [x] 岗位页：iframe 打开；4 行数据渲染（董事长/项目经理/人力资源/普通员工 + 编码/排序）；状态徽章；分页条；工具栏四按钮
- [x] 岗位新增：弹窗（radio 默认正常）→ 提交 → DB 落库（create_by=admin）
- [x] 导出链路：导出确认框 → POST export 出文件名 → GET download → **浏览器实际下载文件**（1790668379岗位数据.xlsx，timestamp 前缀 + 中文解码正确）+ 服务端源文件删除
- [ ] 权限两通道（测试角色隔离验证）：**遗留至 5.0.0**（需多用户/多角色数据支撑，当前库仅 admin；CheckPerm 中间件与 check_perm 函数已在 1.0.0 单测/端到端验证）
- [x] ry 表数据零 schema 变更；预置 100~109 / 1~4 完整复原（测试行已删）

## Task 9 · 收尾（2026-09-29 完成）

- [x] PHPUnit 32 tests 97 assertions 全绿；spec.md 实施记录回填；checklist 勾选；README 总表更新
- [x] deviations #19 登记：Excel 列头样式从简（加粗+灰底两样式）
- [x] 测试数据清理：sys_dept 测试行（204）、sys_post 测试行（6）、sys_oper_log/sys_logininfor 清空、Redis session 键清空、runtime/download 清空、浏览器下载文件删除；admin 会话复原
