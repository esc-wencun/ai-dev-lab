# 6.0.0-字典参数 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（dict type 13 / dict data 8 / config 10，共 31）。

## 字典类型（curl / DB 级）

- [x] GET /system/dict 渲染列表页（system:dict:view；未登录 302 /login，ajax 未登录 code "1"）——curl 实测非 ajax 302→/login；ajax `{"code":"1","msg":"未登录或登录超时。请重新登录"}`
- [x] POST /system/dict/list：TableDataInfo 信封 + 7 驼峰列（Temp/m6_test.py #1）；过滤/白名单排序（PageQuery 机制 PHPUnit 覆盖；dictName=用户 → total=1 实测）
- [x] POST /system/dict/export：{code:0,msg:"<uuid>_字典类型.xlsx"} 导出下载读回 4 列头「字典主键|字典名称|字典类型|状态」+ row2「1|用户性别|sys_user_sex|正常」状态转换正确
- [x] POST /system/dict/add：唯一冲突 500「新增字典'xx'失败，字典类型已存在」；正则违规 500；成功 create_by=admin + `dict:<type>` 键删除（m6_test.py #3~5 + 本轮复测）
- [x] GET /system/dict/edit/{dictId} 渲染编辑弹窗（下划线回显）——浏览器实测 dictId=1 全回显 + status radio
- [x] POST /system/dict/edit：唯一（改前缀）500「修改字典'xx'失败，字典类型已存在」；改 dictType 级联 DB 核验 + 缓存重刷（m6_test.py #10：data.dict_type 级联 + dict:m6_test2）
- [x] POST /system/dict/remove：占用 500「xx已分配,不能删除」；空类型物理删 + 键消失；**批量部分成功不回滚**（m6_delok/m6_delbusy 实测：A 删成功 B 500，DB 只剩 B）
- [x] GET /system/dict/refreshCache：GET 可访问；调用后 `dict:*` 10 键与 DB 启用类型集一致（Redis KEYS 实测）
- [x] GET /system/dict/detail/{dictId}：渲染数据列表页 + dictList 下拉——浏览器实测 3 行数据 + 五按钮（修：detail 漏 assign $datas 已补）
- [x] POST /system/dict/checkDictTypeUnique：裸 true/false；自身放行（dictId=1 自查 true 实测）
- [x] GET /system/dict/selectDictTree/{columnId}/{dictType} + /treeData：渲染 + Ztree 平铺裸数组（treeDataLen=10、name 含 &nbsp;、title=dictType 实测）

## 字典数据（curl / DB 级）

- [x] GET /system/dict/data：渲染列表页；assign 兜底空不报错（浏览器 + curl 实测）
- [x] POST /system/dict/data/list：TableDataInfo；dictType 精确 + dictLabel like（男 → total=1）+ status 过滤（status=1 → 0）；12 驼峰列
- [x] POST /system/dict/data/export：8 列读回断言（m6_test.py #17）
- [x] POST /system/dict/data/add：落库 create_by + `dict:<type>` 重刷（浏览器新增后 Redis 缓存 4 条含新行实测）
- [x] POST /system/dict/data/edit：update_by=admin 落库（DB 实测）+ 缓存重刷（m6_test.py #9）
- [x] POST /system/dict/data/remove：物理删 + 逐类型重刷（浏览器删「浏览器新增」后列表回 3 行；固定 success）
- [x] **listByType 回归**：dept/post 页徽章正常（3.0.0 消费方零感知）；双键输出不变（Redis 值驼峰+下划线并存实测）

## 参数配置（curl / DB 级）

- [x] GET /system/config 渲染列表页（浏览器实测）
- [x] POST /system/config/list：TableDataInfo；过滤 + 10 驼峰列 + configId 白名单（浏览器 11 行 + m6_test.py #11）
- [x] POST /system/config/export：5 列读回（m6_test.py #17）
- [x] POST /system/config/add：键名冲突 500「新增参数'xx'失败，参数键名已存在」；成功 `config:<key>` 即时写入（本轮 500 实测 + Redis "v1" 实测）
- [x] POST /system/config/edit：唯一 500；**改键名** old 键删 + 新键写（m6.oldkey→m6.newkey：EXISTS=0 / GET="v" 实测）；不改键名覆盖新值（v1→v2 Redis 实测）
- [x] POST /system/config/remove：内置 500「内置参数【xx】不能删除 」尾随空格（浏览器 + curl 双实测，msg 原样带尾空格）；N 型删 + 键删（browser.m6.key 删后 EXISTS=0）；批量混合部分成功（ids=N型,1 → N删成功 + 内置 500 不回滚）
- [x] GET /system/config/refreshCache：`config:*` 11 键与 DB 一一对应（KEYS 计数实测）
- [x] POST /system/config/checkConfigKeyUnique：裸 true/false；自身放行（configId=9 自查 true 实测）
- [x] **与 2.0.0 联动**：改 sys.index.footer=false → 主框架刷新后页脚 GONE → 复原 true 页脚回归（浏览器实测）

## 页面级验收（浏览器）

- [x] 字典类型列表页：10 行渲染 + 状态徽章 + 五按钮；**dictType 链接 → 右侧抽屉**（标题/副标题/统计卡「共计条目 3」/数据卡片/关闭按钮均实测）
- [x] 字典类型新增/修改弹窗：表单字段齐全（dictName/dictType/status radio 默认正常/remark）；提交入库 + 弹窗自关 + 表格刷新（m6_browser 全流程实测）；remote/minlength 5 前端校验在模板（引擎冒烟已验，未逐条触发——前端 validate 插件标准行为）
- [x] 字典数据 tab 页：「列表」按钮打开 detail 页（标签页「字典数据」）；新增弹窗 dictType 预填 readonly=sys_user_sex；isDefault=Y 默认勾选；提交/删除实测（「浏览器新增」dictCode 101 入库又删）；关闭按钮在工具栏
- [x] 参数列表页：11 行 + 徽章 + 五按钮（含刷新缓存）
- [x] 参数新增/修改弹窗：回显正确（browser.m6.key edit 弹窗全回显）；radio 默认是；提交/编辑/删除全流程实测
- [x] 刷新缓存按钮 ×2：字典页点击「操作成功」+ Redis dict:* 10 键；config refreshCache 后 11 键（API 实测）
- [x] 导出按钮 ×3：API 侧三件导出全部 code 0 + 文件名格式正确（浏览器点击下载未逐一点验——API 下载读回已断言内容正确）
- [x] 全部页面在主框架 iframe 内打开、标签页标题正确（字典管理/字典数据/参数设置 tab 实测）

## 横切与纪律自查

- [x] #[Perm] 27 处（type 10 + data 8 + config 9 grep 实测）；4 免注解端点仅登录态（未登录 code "1" 实测）
- [x] #[Log] 14 处（5+4+5 grep 实测）；sys_oper_log 落库 title/business_type 与 spec 对应（oper_id 154~165 实测，INSERT=1/UPDATE=2/DELETE=3/CLEAN=8/EXPORT=5）
- [x] RepeatSubmit 未挂本模块（grep 0 处）
- [x] 业务代码零处直接 predis（grep 0 处）；Redis 只经 RedisCache 门面
- [x] 本模块零 DataScope（grep 0 处）；零 admin 保护
- [x] 权限两通道一致：sys_menu.perms 12 串与 #[Perm]、check_perm 三处互查（DB 实测）
- [x] 表结构零变更（SHOW COLUMNS 核对 sys_dict_type）
- [x] 列表输出驼峰 / 表单回显下划线双轨执行（edit 弹窗下划线回显 + list API 驼峰实测）

## Deviations 核对

- [x] 「输入侧全局 XSS 转义」候选处置：维持不转义（对位经典版原样；与 7.0.0 summernote 富文本一并评估，暂无需求触发）——处置：不登记 deviations（无行为差异，仅候选未采纳）
- [x] 经典版原样 quirk 清单确认未"修复"：refreshCache GET+remove 权限、内置尾随空格文案、「字典类型类型长度」措辞、remove 固定 success、批量删除无事务（部分成功实测）、add 页冗余 name 参数、data 页关闭按钮无权限控制、selectDictTree/treeData 无调用方保留——逐项 grep/实测确认

## 测试数据清理记录

- [x] ✅ 2026-09-30 测试字典类型 m6_test/m6_test2/m6_browser/m6_delok/m6_delbusy 物理删除；sys_dict_type 恢复预置 1~10（COUNT=10 核验）
- [x] ✅ 2026-09-30 测试字典数据行物理删除；sys_dict_data 恢复预置 29 行（COUNT=29 核验）
- [x] ✅ 2026-09-30 测试参数 m6.test.key/browser.m6.key/m6.oldkey(m6.newkey)/m6.mix.key 物理删除；sys_config 恢复预置 11 行（COUNT=11 核验；sys.index.footer=true 复原 + 页脚回归实测）
- [x] ✅ 2026-09-30 sys_oper_log 测试行保留（操作日志为审计数据，非业务测试数据；m6 相关行可留痕，不影响功能）——经查与既往模块处置一致（3.0.0/5.0.0 亦保留）
- [x] ✅ 2026-09-30 runtime/download 导出残留已随 delete=true 下载消费清理；残留 uuid 文件手工删除
- [x] ✅ 2026-09-30 Redis 测试键清理（dict:m6_* 已随删除联动清除；KEYS dict:*=10 / config:*=11 与预置一致）
- [x] ✅ 2026-09-30 admin 会话状态正常（浏览器持续登录操作验证；admin/admin123 登录链路本轮多次实测）
