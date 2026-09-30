# 6.0.0-字典参数 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：DictService 收编（PHPUnit 先行）→ ConfigService 补齐 → 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。2026-09-29~30 动工并完成。

## Task 1 · DictService 收编（类型域）

- [x] `selectDictTypeList(filter)`：返回 Query 构造器（List 返回 Query；getLastInsID 坑已避开用 max）
- [x] `selectDictTypeAll()` / `selectDictTypeById` / `selectDictTypeByType`
- [x] `checkDictTypeUnique(dictType, ?dictId)`：全局唯一 limit 1，自身放行
- [x] `insertDictType`：create_by/create_time 显式写
- [x] `updateDictType`（显式事务）：级联 UPDATE sys_dict_data.dict_type + setCache 新类型重刷（m6_test.py #10 级联实测）
- [x] `deleteDictTypeByIds`：逐个循环占用校验 + 物理删 + removeCache；不包事务（批量部分成功实测）
- [x] `selectDictTree()`：平铺无 pId；name 串含 &nbsp;、title=dictType（treeDataLen=10 实测）
- [x] PHPUnit：Module6Test 覆盖 Ztree 组装/唯一口径/删除占用文案（全绿）

## Task 2 · DictService 收编（数据域 + 缓存门面）

- [x] 缓存门面六方法（set/remove/clear/loading/reset）真实 Redis db1 断言
- [x] `selectDictDataList(filter)` Query 构造器；`selectDictDataById` / `countDictDataByType`
- [x] `insertDictData` / `updateDictData` / `deleteDictDataByIds`：三联动 setCache 重刷（浏览器/Redis 双实测）
- [x] **listByType 兼容回归**：签名/双键/缓存语义三不变（Redis 值驼峰+下划线并存实测）；dept/post 徽章消费方零回归
- [x] PHPUnit 缓存门面 + 分组排序口径（全绿）

## Task 3 · ConfigService 收编（补齐管理端联动）

- [x] `set(configKey, value)`：写 `config:<key>`（浏览器新增 browser.m6.key 后 Redis "v1" 实测）
- [x] `loadAll()`：全表回填（reset 后 11 键实测）
- [x] `reset()`：全清 + loadAll
- [x] PHPUnit：set 后 get 命中缓存；loadAll 后 11 键齐全（全绿）

## Task 4 · 控制器与路由

- [x] DictTypeController 13 方法：#[Perm] 10 + #[Log] 5（grep 实测）；refreshCache=GET+remove+CLEAN；remove 固定 success；detail 渲染 data/index（补 assign $datas）；三免注解端点裸登录态
- [x] DictDataController 8 方法：#[Perm] 8 + #[Log] 4；index 兜底空；remove 固定 success
- [x] ConfigController 10 方法：#[Perm] 9 + #[Log] 5；内置尾随空格文案；edit 改键名 refresh(old)+set(new)（m6.oldkey→newkey 实测）；check 裸 bool
- [x] 全局类 use 纪律执行（本轮零 use 缺失）
- [x] 路由 31 条注册；固定段先于 :param（dict add 路由修正：去掉误加的 :parentId，GET /add 无路径参数——经典版实锤）
- [x] RepeatSubmit 未挂（grep 0）
- [x] curl 级 31 端点自测：m6_test.py 22 项 + 本轮补测 14 项断言全过（含缓存联动 DB+Redis 双核验）

## Task 5 · 页面模板（10 页）

- [x] dict/type/index.html：五按钮 + 字典抽屉 JS + 搜索四字段（浏览器实测抽屉「共计条目 3」+ 数据卡片）
- [x] dict/type/add.html + edit.html：remote + minlength:5 + radio 默认（浏览器实测）
- [x] dict/type/tree.html：zTree 平铺（引擎冒烟已验；无调用方，联调未触发——原样保留端点）
- [x] dict/data/index.html：select2 下拉 + queryParams 覆写 + add() 取下拉值 + 关闭按钮（detail 页浏览器实测）
- [x] dict/data/add.html + edit.html：dictType readonly 预填、listClass 七选、radio 默认（浏览器实测 add 弹窗）
- [x] config/index.html：五按钮 + 徽章 + tooltip（浏览器实测 11 行）
- [x] config/add.html + edit.html：remote + textarea + radio 默认是（浏览器实测回显）
- [x] 模板自查（subagent 引擎冒烟 20 用例全过：无 th: 残留、check_perm 输出形式、配平、双轨键名）

## Task 6 · 端到端验收（浏览器级）

- [x] 字典类型页：10 行 + 徽章 + 五按钮；dictType 链接开抽屉（3 条数据统计卡实测；「29 条」未逐一验——抽屉数据源 data/list 实测正确）
- [x] 类型新增/编辑/删除全链路（m6_browser 入库→编辑回显→删除实测）；删除含数据类型 500；空类型删除成功
- [x] 类型编辑改 dictType：级联 + 缓存重刷（m6_test.py #10）
- [x] 数据页：detail tab 打开、CRUD 全链路（「浏览器新增」dictCode 101 入库后删）、增删后 listByType 即时反映（Redis 4 条实测）
- [x] 刷新缓存按钮：双模块各点一次 → dict:* 10 / config:* 11 键实测
- [x] 参数页：内置删除 500 尾随空格（浏览器实测「内置参数【sys.index.skinName】不能删除 」）；N 型 CRUD + 改键名联动实测
- [x] 导出链路 ×3：API 导出+下载读回 4/8/5 列内容正确（浏览器点击下载未逐一点验——下载由 common/download 标准链路承担，API 侧已断言）
- [x] ry-tp 零 schema 变更；预置 10/29/11 完整复原

## Task 7 · 收尾

- [x] PHPUnit 全绿（40 tests 118 assertions，含 Module6Test 新增用例）；spec.md 实施记录回填；checklist 勾选；README 总表更新（6.0.0 ✅）
- [x] deviations 处置回填：输入侧 XSS 转义维持不转义（无行为差异，不登记 deviations；7.0.0 summernote 再评估）
- [x] 测试数据清理：字典/参数测试行全删（10/29/11 复原）、Redis 测试键清（10/11）、sys_oper_log 保留（审计数据，与 3.0.0/5.0.0 处置一致）、runtime/download 已清、admin 会话正常
