# 6.0.0-字典参数 · spec

> **状态：✅ 已完成（2026-09-30 联调验收通过）**
> 对位经典若依 SysDictTypeController（13 端点）/ SysDictDataController（8 端点）/ SysConfigController（10 端点）+ SysDictTypeServiceImpl / SysDictDataServiceImpl / SysConfigServiceImpl + DictUtils / CacheUtils + templates/system/dict|config（端点 31 / 页面 10，逐方法核对实锤）。
> 依赖：1.0.0（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / LoginAuth）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下（按序号串行隐式满足）。
> 调研依据：reference 源码逐文件核对（Controller×3 / ServiceImpl×3 / Mapper XML×3 / 页面 HTML×10 / SysDictType·SysDictData·SysConfig 实体 / ry-ui.js $.table.exportExcel·$.operate.get + type.html 内嵌字典抽屉 JS）+ **ry-tp 库实测**（sys_dict_type 9 列 / sys_dict_data 14 列 / sys_config 10 列；预置 dict_type **10 行**、dict_data **29 行**、config **11 行全内置**；sys_menu perms 11 行：菜单 105/106 + 按钮 1025~1034）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、POST 分页参数、信封结构、权限两通道（#[Perm] 注解 + check_perm 模板函数）。
> **本模块核心命题 = 收编**：3.0.0 落的 DictService 最小只读版（listByType + `dict:` 缓存）扩为完整字典服务（类型/数据管理 + 缓存失效联动）；2.0.0 落的 ConfigService（get/getBool/refresh，`config:` 键）补齐管理端联动。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/system/DictTypeController.php | SysDictTypeController | 13 个方法（含 check/selectDictTree/treeData 三个无权限注解端点） |
| app/controller/system/DictDataController.php | SysDictDataController | 8 个方法 |
| app/controller/system/ConfigController.php | SysConfigController | 10 个方法 |
| app/service/DictService.php（**收编扩容**） | SysDictTypeServiceImpl + SysDictDataServiceImpl + DictUtils | 三者合并为单类：类型/数据两域的查询维护 + `dict:<type>` 缓存门面（set/remove/clear/loading/reset）；理由=缓存维护共享且域极小，3.0.0 已定收编主体；listByType 签名与双键输出**不变**（3.0.0/1.5.0 已消费） |
| app/service/ConfigService.php（**收编补齐**） | SysConfigServiceImpl + CacheUtils | 2.0.0 已有 get/getBool/refresh；本模块补 set（对位 CacheUtils.put 单键写值）/ loadAll（全量预热）/ reset（清+预热，对位 resetConfigCache）；管理端增删改联动在本模块接线 |
| app/view/system/dict/type/*.html ×4、app/view/system/dict/data/*.html ×3、app/view/system/config/*.html ×3 | templates/system/dict/type|data/、templates/system/config/ | 10 页；静态 JS（ry-ui.js / bootstrap-table / select2 / zTree）零改动对接；type 页内嵌「字典数据抽屉」为纯前端 JS（POST data/list 复用现有端点），零后端增量 |

**范围外**（后续模块，勿在本模块实现）：sys_job 相关字典消费（10.0.0 任务状态/分组，只读经 listByType 天然可用）、代码生成器字典树消费（11.0.0）。

## 页面清单（10）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 字典类型列表页 | view/system/dict/type/index.html | dict/type/type.html | bootstrap-table（POST + server 分页） | `prefix = ctx + "system/dict"`；options：url=prefix+/list、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove（**无 {id}**，ids 走 body）、exportUrl=prefix+/export；sortName=dictId（**初始请求恒发 orderByColumn=dictId&isAsc=asc，白名单必须含 dict_id**）；sortable 列 dictName/dictType/createTime；modalName=类型；工具栏五按钮：新增/修改(single)/删除(multiple)/导出/刷新缓存（refreshCache 按钮权限串=system:dict:remove，原样 quirk）；搜索字段 dictName/dictType/status/params[beginTime]/params[endTime]；**dictType 列链接点击打开右侧抽屉**（内嵌 drawer JS：POST system/dict/data/list，data {dictType, pageNum:1, pageSize:100}，前端渲卡片 + 状态徽章，ESC/遮罩关闭）——纯前端，零后端增量 |
| 2 | 字典类型新增弹窗 | view/system/dict/type/add.html | dict/type/add.html | 表单 | validate remote：POST prefix+/checkDictTypeUnique（data 冗余发 `name` 字段，jquery validate 自动附带 dictType 字段值，后端以 dictType 取参）；文案「该字典类型已经存在」；dictType 另有 minlength:5 前端校验；radio sys_normal_disable（默认正常）；提交 save(prefix+"/add") |
| 3 | 字典类型修改弹窗 | view/system/dict/type/edit.html | dict/type/edit.html | 表单 | remote data 带 dictId+dictType；同上文案；提交 save(prefix+"/edit")；hidden dictId |
| 4 | 字典树选择弹窗 | view/system/dict/type/tree.html | dict/type/tree.html | **zTree（平铺无层级）** | `url = ctx + "system/dict/treeData"`（GET，返回 Ztree 裸数组，无 pId）；zOnClick 回填 treeNode.title（=dictType 值）；搜索/展开折叠纯前端 |
| 5 | 字典数据列表页（tab 页） | view/system/dict/data/index.html | dict/data/data.html | bootstrap-table + **select2**（类型下拉） | `prefix = ctx + "system/dict/data"`；options：url=prefix+/list、createUrl=prefix+"/add/{id}"（**{id} 由类型下拉当前值替换**，add() 函数取 option value 传入）、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove、exportUrl=prefix+/export；sortName=dictSort（白名单须含 dict_sort）；sortable 列 createTime；queryParams 覆写 search.dictType = 下拉当前值；modalName=数据；工具栏五按钮：新增/修改(single)/删除(multiple)/导出/**关闭**（closeItem 关 tab，无权限控制）；include select2-css/js；重置按钮 resetPre() 重置后 trigger("change")（select2 联动） |
| 6 | 字典数据新增弹窗 | view/system/dict/data/add.html | dict/data/add.html | 表单 | 字段：dictLabel*/dictValue*/dictType（readonly 预填）/cssClass/dictSort*（digits）/listClass 下拉（空/default/primary/success/info/warning/danger）/isDefault radio（sys_yes_no，默认是）/status radio（sys_normal_disable，默认正常）/remark；无 remote 校验；提交 save(prefix+"/add") |
| 7 | 字典数据修改弹窗 | view/system/dict/data/edit.html | dict/data/edit.html | 表单 | hidden dictCode；字段同上回显（下划线键）；提交 save(prefix+"/edit") |
| 8 | 参数列表页 | view/system/config/index.html | config/config.html | bootstrap-table | `prefix = ctx + "system/config"`；options：url/createUrl/add/edit 同套路、removeUrl=prefix+/remove、exportUrl=prefix+/export；sortName=configId（白名单须含 config_id）；**无 sortable 列**；modalName=参数；configValue/remark 列 tooltip(value,10,"open)；configType 列徽章 datas=sys_yes_no；工具栏五按钮：新增/修改/删除/导出/刷新缓存（权限串 system:config:remove）；搜索字段 configName/configKey/configType/时间范围 |
| 9 | 参数新增弹窗 | view/system/config/add.html | config/add.html | 表单 | configName*/configKey*/configValue*（textarea rows=4）/configType radio（sys_yes_no，默认是）/remark；remote：POST prefix+/checkConfigKeyUnique，文案「参数键名已经存在」；提交 save(prefix+"/add") |
| 10 | 参数修改弹窗 | view/system/config/edit.html | config/edit.html | 表单 | hidden configId；remote data 带 configId+configKey；同上文案；提交 save(prefix+"/edit") |

> 模板命名沿用 3.0.0 惯例：列表页统一 index.html（对位 type.html/data.html/config.html），控制器显式指定渲染文件。**本模块无「新增/修改全屏弹窗」**（均普通 layer 弹层，对位 3.0.0）。

## 端点级 API 清单（dict type 13 + dict data 8 + config 10）

### 字典类型 /system/dict（对位 SysDictTypeController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/dict | system:dict:view | 无 | — | 渲染页面 1 |
| 2 | POST | /system/dict/list | system:dict:list | 无 | pageNum/pageSize/orderByColumn/isAsc（**白名单 dict_id/dict_name/dict_type/create_time**——sortName=dictId 默认发 dictId）+ dictName（like）/ dictType（like）/ status / params[beginTime]/[endTime]（按日粒度过滤 create_time） | TableDataInfo {code:0, msg:"查询成功", rows, total}；行字段驼峰且**仅 selectVo 7 列**：dictId/dictName/dictType/status/createBy/createTime/remark（updateBy/updateTime 经典版 SQL 未查，不输出） |
| 3 | POST | /system/dict/export | system:dict:export | 字典类型, 5导出 | 同搜索参数；无分页（导全量） | {code:0, msg:"&lt;uuid&gt;_字典类型.xlsx"}；Excel 4 列（见导出列定义节） |
| 4 | GET | /system/dict/add | system:dict:add | 无 | — | 渲染页面 2 |
| 5 | POST | /system/dict/add | system:dict:add | 字典类型, 1新增 | dictName*、dictType*、status、remark | 校验顺序：@Validated（文案见特殊行为 8）→ checkDictTypeUnique 失败 → error(500)「新增字典'{dictName}'失败，字典类型已存在」；成功 toAjax + **缓存联动：DictService::removeCache(dictType)**（对位 setDictCache(type, null)，下次读回填）；createBy=登录名 |
| 6 | GET | /system/dict/edit/{dictId} | system:dict:edit | 无 | 路径 dictId | 渲染页面 3，变量 dict（下划线键回显） |
| 7 | POST | /system/dict/edit | system:dict:edit | 字典类型, 2修改 | dictId、dictName*、dictType*、status、remark | 唯一「修改字典'{dictName}'失败，字典类型已存在」(500)；**事务**：update sys_dict_data.dict_type（旧→新级联）→ update 类型行 → 成功后 DictService 按**新类型**全量重刷缓存（setCache(newType, 查库结果)）；updateBy=登录名；toAjax |
| 8 | POST | /system/dict/remove | system:dict:remove | 字典类型, 3删除 | ids（逗号串） | 逐个循环：countDictDataByType(type)>0 → **error(500)**「{dictName}已分配,不能删除」（ServiceException；**循环中途抛异常不回滚，此前已删的保留——经典版无事务，原样**）→ 物理删类型行 → DictService::removeCache(type)；**固定返回 success()**（非 toAjax，全删光也返回 {code:0,msg:操作成功}） |
| 9 | GET | /system/dict/refreshCache | system:dict:remove | 字典类型, 8清空 | — | **GET 方法 + remove 权限（原样 quirk）**；DictService::resetCache()（清全部 dict:* + 全量预热）；固定 success() |
| 10 | GET | /system/dict/detail/{dictId} | system:dict:list | 无 | 路径 dictId | 渲染**页面 5**（复用 data/index.html），变量 dict（当前类型）+ dictList（全类型列表 selectDictTypeAll，供类型下拉） |
| 11 | POST | /system/dict/checkDictTypeUnique | **无（仅登录态）** | 无 | dictType（+编辑时 dictId；add 页前端冗余发的 name 参数忽略） | **裸 boolean**：口径 = dict_type **全局唯一**（DB 层有 UNIQUE 键）limit 1，dictId 相同视为自身放行 |
| 12 | GET | /system/dict/selectDictTree/{columnId}/{dictType} | **无（仅登录态）** | 无 | 路径 columnId + dictType | 渲染页面 4，变量 columnId / dict（按 dictType 查的类型行）；辅助页（当前经典版模板无调用方，原样保留） |
| 13 | GET | /system/dict/treeData | **无（仅登录态）** | 无 | — | **Ztree 裸数组（平铺，无 pId）**：[{id:dictId, name:"(dictName)&amp;nbsp;&amp;nbsp;&amp;nbsp;{dictType}", title:dictType}]——**仅 status='0'** 的类型；name 串含 HTML 空格实体原样输出 |

### 字典数据 /system/dict/data（对位 SysDictDataController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 14 | GET | /system/dict/data | system:dict:view | 无 | — | 渲染页面 5；**经典版不传模板变量**——TP 版 assign dictList=[] / dict=null 兜底（类型下拉空，直接 URL 访问的边缘场景，原样） |
| 15 | POST | /system/dict/data/list | system:dict:list | 无 | pageNum/pageSize/orderByColumn/isAsc（**白名单 dict_sort/create_time**）+ dictType（**精确 =**）/ dictLabel（like）/ status；抽屉调用固定发 {dictType, pageNum:1, pageSize:100} | TableDataInfo；行字段驼峰且**仅 selectVo 12 列**：dictCode/dictSort/dictLabel/dictValue/dictType/cssClass/listClass/isDefault/status/createBy/createTime/remark（无 update 列） |
| 16 | POST | /system/dict/data/export | system:dict:export | 字典数据, 5导出 | 同搜索参数；导全量 | {code:0, msg:"&lt;uuid&gt;_字典数据.xlsx"}；Excel 8 列（见导出列定义节） |
| 17 | GET | /system/dict/data/add/{dictType} | system:dict:add | 无 | 路径 dictType（来自页面 5 类型下拉） | 渲染页面 6，变量 dictType（readonly 预填） |
| 18 | POST | /system/dict/data/add | system:dict:add | 字典数据, 1新增 | dictLabel*、dictValue*、dictType、cssClass、dictSort、listClass、isDefault、status、remark | 无唯一校验；createBy；toAjax；成功 → **DictService::setCache(dictType, 按该类型全量重查)**（status='0'，order by dict_sort） |
| 19 | GET | /system/dict/data/edit/{dictCode} | system:dict:edit | 无 | 路径 dictCode | 渲染页面 7，变量 dict（下划线键回显） |
| 20 | POST | /system/dict/data/edit | system:dict:edit | 字典数据, 2修改 | dictCode、dictLabel*、dictValue*、dictType（readonly 但随表单提交）、cssClass、dictSort、listClass、isDefault、status、remark | updateBy；toAjax；成功 → setCache(dictType, 全量重查)（与 add 同联动） |
| 21 | POST | /system/dict/data/remove | system:dict:remove | 字典数据, 3删除 | ids（逗号串） | 逐个循环：查行 → 物理删 → setCache(该行 dictType, 重查)；**固定 success()**（非 toAjax） |

### 参数配置 /system/config（对位 SysConfigController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 22 | GET | /system/config | system:config:view | 无 | — | 渲染页面 8 |
| 23 | POST | /system/config/list | system:config:list | 无 | pageNum/pageSize/orderByColumn/isAsc（**白名单 config_id**）+ configName（like）/ configKey（like）/ configType / params[beginTime]/[endTime] | TableDataInfo；行字段驼峰 **selectVo 10 列**：configId/configName/configKey/configValue/configType/createBy/createTime/updateBy/updateTime/remark |
| 24 | POST | /system/config/export | system:config:export | 参数管理, 5导出 | 同搜索参数；导全量 | {code:0, msg:"&lt;uuid&gt;_参数数据.xlsx"}；Excel 5 列（见导出列定义节）；sheet 名「参数数据」 |
| 25 | GET | /system/config/add | system:config:add | 无 | — | 渲染页面 9 |
| 26 | POST | /system/config/add | system:config:add | 参数管理, 1新增 | configName*、configKey*、configValue*、configType、remark | @Validated → checkConfigKeyUnique 失败 → error(500)「新增参数'{configName}'失败，参数键名已存在」；createBy；toAjax；成功 → **ConfigService::set(configKey, configValue)**（对位 CacheUtils.put） |
| 27 | GET | /system/config/edit/{configId} | system:config:edit | 无 | 路径 configId | 渲染页面 10，变量 config（下划线键回显） |
| 28 | POST | /system/config/edit | system:config:edit | 参数管理, 2修改 | configId、configName*、configKey*、configValue*、configType、remark | 唯一「修改参数'{configName}'失败，参数键名已存在」(500)；**改键名联动**：先查旧行，configKey 变更 → ConfigService::refresh(oldKey)；update 后 ConfigService::set(newKey, configValue)；updateBy；toAjax |
| 29 | POST | /system/config/remove | system:config:remove | 参数管理, 3删除 | ids（逗号串） | 逐个循环：configType='Y'（内置）→ **error(500)**「内置参数【{configKey}】不能删除 」（**句尾含一个空格，原样复刻**；ServiceException；循环中途异常不回滚）→ 物理删 → ConfigService::refresh(configKey)；**固定 success()** |
| 30 | GET | /system/config/refreshCache | system:config:remove | 参数管理, 8清空 | — | **GET + remove 权限（原样 quirk）**；ConfigService::reset()（清全部 + 全量预热 loadAll）；固定 success() |
| 31 | POST | /system/config/checkConfigKeyUnique | **无（仅登录态）** | 无 | configKey（+编辑时 configId） | 裸 boolean；config_key 全局唯一 limit 1，自身放行 |

统计：**控制器方法 31 个**（dict type 13 + dict data 8 + config 10），URL pattern 31 条（1:1）；**#[Perm] 共 27 处**（type 10 + data 8 + config 9；#11/#12/#13/#31 四个端点无权限注解，仅登录态）；**#[Log] 共 14 处**（type 5：EXPORT/INSERT/UPDATE/DELETE/CLEAN；data 4：EXPORT/INSERT/UPDATE/DELETE；config 5：EXPORT/INSERT/UPDATE/DELETE/CLEAN）。

## 特殊行为清单（缓存联动 / 删除校验 / 唯一性 / quirk）

1. **字典缓存联动全景**（DictService 收编后的职责；缓存键 `dict:<type>`，TP 自定前缀 tech-stack 已授权）：
   - listByType(type)：缓存 miss 查库回填（3.0.0 已落，签名不变）；
   - 类型新增 → removeCache(type)（对位经典 setDictCache(type, null)——ehcache 放 null 元素等效失效，TP 直接删键，下次读回填，行为等价）；
   - 类型修改（事务内）→ sys_dict_data.dict_type 旧值级联 UPDATE 为新值 → setCache(newType, 全量重查)；**旧类型缓存键不删（经典版同样残留，原样复刻勿"修复"）**；
   - 类型删除 → 物理删 + removeCache(type)；
   - 数据增/改/删 → setCache(该数据 dictType, 按 status='0' + order by dict_sort 全量重查)；
   - refreshCache 端点 → resetCache() = clearCache()（SCAN `dict:*` 全删）+ loadingCache()（全表 status='0' 数据 group by dict_type、dict_sort 排序后逐类型回填——对位经典 @PostConstruct loadingDictCache 同款口径）。
2. **ConfigService 联动全景**（`config:<key>`；2.0.0 已有 get/getBool/refresh）：
   - 参数新增 → set(key, value)（补的方法，对位 CacheUtils.put）；参数修改 → 键名变更先 refresh(oldKey)，保存后 set(newKey, value)；参数删除 → refresh(key)；refreshCache 端点 → reset() = refresh()（SCAN 全清）+ **loadAll()（全表 sys_config 逐行回填，对位 resetConfigCache 语义——不复用只清不填的 refresh()，保证端点返回后 `config:*` 键与 DB 11 行一一对应可断言）**；
   - **与 2.0.0 已消费键的天然联动测试场景**：改 `sys.index.footer` 为 false → 主框架页脚消失（2.0.0 读 ConfigService::get 同一缓存键）；测试后复原 true。
3. **字典类型删除校验 dict_data 占用**：删除前 countDictDataByType>0 → 「{dictName}已分配,不能删除」error(500)。**注意批量删除中途抛异常不回滚**（经典 deleteDictTypeByIds 无 @Transactional，循环删到谁算谁）——原样复刻，不自作主张包事务。
4. **内置参数不可删**：config_type='Y' → 「内置参数【{configKey}】不能删除 」（**尾随空格是经典版源码字面量，原样**）。ry-tp 预置 11 条**全部内置**——即预置参数一条都删不掉，校验必现；测试需自建 N 型参数再删。
5. **唯一性校验文案全集**（前后端双保险，文案不同属经典版原样）：
   - 后端 add/edit：`新增字典'{dictName}'失败，字典类型已存在` / `修改字典'{dictName}'失败，字典类型已存在`；`新增参数'{configName}'失败，参数键名已存在` / `修改参数'{configName}'失败，参数键名已存在`
   - 前端 remote 提示：字典「该字典类型已经存在」（另 dictType minlength:5）；参数「参数键名已经存在」
   - 口径：dict_type / config_key 均**全局唯一**（非同名同父那类维度）；check 端点裸 boolean。
6. **后端参数校验文案**（对位 @Validated + BindException → error 首条）：
   - 字典类型：字典名称「字典名称不能为空」「字典类型名称长度不能超过100个字符」；字典类型「字典类型不能为空」「字典类型类型长度不能超过100个字符」（"类型类型"是原文案）＋**正则 `^[a-z][a-z0-9_]*$`**「字典类型必须以字母开头，且只能为（小写字母，数字，下滑线）」
   - 字典数据：dictLabel「字典标签不能为空/长度不能超过100个字符」；dictValue「字典键值不能为空/长度不能超过100个字符」；dictType「字典类型不能为空/长度不能超过100个字符」；cssClass「样式属性长度不能超过100个字符」；dictSort 无后端校验（前端 digits）
   - 参数：configName「参数名称不能为空/参数名称不能超过100个字符」；configKey「参数键名长度不能为空/参数键名长度不能超过100个字符」；configValue「参数键值不能为空/参数键值长度不能超过500个字符」（DB varchar(500) 同口径）
7. **删除返回值差异**：dict type / dict data / config 三个 remove 均固定 success()（{code:0,msg:操作成功}），**非 toAjax**——删 0 行也报成功（原样；3.0.0 dept 的「重复删报操作失败」行为在此**不**适用）。
8. **refreshCache 双 quirk（原样）**：① HTTP 方法是 **GET**；② 权限串复用 `system:{dict,config}:remove`（不是独立 refresh 权限）；③ #[Log] businessType=8 清空（经典 CLEAN 枚举序号 8，1.0.0 Log 注解已定义）。
9. **bootstrap-table 默认排序必须进白名单**（前序踩坑预防）：type 页 sortName=dictId、data 页 sortName=dictSort、config 页 sortName=configId——初始请求恒发 orderByColumn，PageQuery 白名单缺了对应下划线列（dict_id/dict_sort/config_id）默认排序即静默丢失。
10. **前端 quirk（原样保留，勿修）**：① type/add.html remote 冗余发 `name` 参数（jquery validate 自动附带 dictType 字段值才是生效参数）；② data 页「关闭」按钮无权限控制；③ selectDictTree/treeData 辅助端点当前无页面调用方（历史遗留，原样保留）；④ GET /system/dict/data 无模板变量（TP 兜底空数组）。
11. **防重复提交**：三个控制器零 @RepeatSubmit（grep 实锤）→ 不挂 RepeatSubmit 中间件。
12. **视图 layer / include**：控制器在 system 层 → 模板路径 'dict/type/index'、'dict/data/index'、'config/index'（layer 相对）；include 片段用 app/view/system/include/ 局部副本（43 个已有：select2-css/js 页面 5 用、ztree-css/js 页面 4 用）。**全局类（RedisCache/TpConstant/PageQuery/TableDataInfo/AjaxResult/ExcelExportService）在 namespaced 文件必须 use（三次踩坑，写码前检查）**。

## 数据权限声明

- **本模块无数据权限**：SysDictTypeServiceImpl / SysDictDataServiceImpl / SysConfigServiceImpl 均无 @DataScope（实锤）；所有查询 admin 与普通用户同视野。
- 无 admin 保护逻辑（无 checkXxxDataScope 类端点）。
- checkDictTypeUnique / selectDictTree / treeData / checkConfigKeyUnique 无 #[Perm] 注解仅登录态（经典版实锤，原样）。

## Excel 导出列定义（@Excel 逐字段抄录）

### SysDictType（4 列，sheet「字典类型」，文件名 &lt;uuid&gt;_字典类型.xlsx）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | dictId | name="字典主键", cellType=NUMERIC | 数字格式 |
| 2 | dictName | name="字典名称" | 文本 |
| 3 | dictType | name="字典类型" | 文本 |
| 4 | status | name="状态", readConverterExp="0=正常,1=停用" | 转换后文本 |

### SysDictData（8 列，sheet「字典数据」，文件名 &lt;uuid&gt;_字典数据.xlsx）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | dictCode | name="字典编码", cellType=NUMERIC | 数字格式 |
| 2 | dictSort | name="字典排序", cellType=NUMERIC | 数字格式 |
| 3 | dictLabel | name="字典标签" | 文本 |
| 4 | dictValue | name="字典键值" | 文本 |
| 5 | dictType | name="字典类型" | 文本 |
| 6 | cssClass | name="字典样式" | 文本 |
| 7 | isDefault | name="是否默认", readConverterExp="Y=是,N=否" | 转换后文本 |
| 8 | status | name="状态", readConverterExp="0=正常,1=停用" | 转换后文本 |

> listClass **无 @Excel 注解不导出**；createBy/createTime/remark（BaseEntity）无 @Excel 不导出。

### SysConfig（5 列，sheet「参数数据」，文件名 &lt;uuid&gt;_参数数据.xlsx）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | configId | name="参数主键", cellType=NUMERIC | 数字格式 |
| 2 | configName | name="参数名称" | 文本 |
| 3 | configKey | name="参数键名" | 文本 |
| 4 | configValue | name="参数键值" | 文本 |
| 5 | configType | name="系统内置", readConverterExp="Y=是,N=否" | 转换后文本 |

- 三个 export 均复用 3.0.0 ExcelExportService（列定义数组驱动 + readConverterExp + NUMERIC 格式）；响应文件名装 msg，前端 GET /common/download（已落）。

## 关键设计说明

1. **DictService 收编形态**（单类扩容，`W:\wencun\ai\java_ai\ruoyi\RuoYi-TP\app\service\DictService.php`）：
   - 保留 `listByType(string $type): array`（签名/双键输出/缓存语义三不变——3.0.0 消费方 dept/post 页与 1.5.0 主框架零感知）；
   - 新增缓存门面：`setCache(type, ?rows)`（null 等价删键）/ `removeCache(type)` / `clearCache()`（RedisCache::keysScan('dict:*') 逐删）/ `loadingCache()`（全量 group 回填）/ `resetCache()`；
   - 新增类型域：selectDictTypeList（like/status/时间过滤 + Query 构造器返回，分页/全量由调用方定——3.0.0 PostService 先例）/ selectDictTypeById / selectDictTypeByType / selectDictTypeAll / checkDictTypeUnique / insertDictType / updateDictType（事务 + 级联）/ deleteDictTypeByIds（循环删 + 占用校验）/ selectDictTree（Ztree 平铺数组）；
   - 新增数据域：selectDictDataList / selectDictDataById / insertDictData / updateDictData / deleteDictDataByIds / countDictDataByType。
   - PHPUnit 先行：缓存门面六方法键断言（真实 Redis db1）、checkDictTypeUnique 自身放行、Ztree name/title 组装、@Validated 文案映射表。
2. **ConfigService 收编形态**：补 `set(key, value)` / `loadAll()` / `reset()` 三个方法；get/getBool/refresh 不动（2.0.0 消费方 login/index 零感知）。
3. **输出键名双轨**（前序踩坑#1）：列表/export 输出**驼峰**（bootstrap-table columns field 消费）；add/edit 回显 assign **下划线**（模板 $dict.dict_type 直取）。时间统一 'Y-m-d H:i:s'。
4. **页面 JS 权限变量接线**（check_perm 输出形式，2.0.0 踩坑#6）：`var editFlag = '{:check_perm('system:dict:edit') ? '' : 'hidden'}'`；按钮 `{if condition=...}` 语法**不可用**，一律 `{:check_perm(...) ? '' : 'hidden'}` 或标签对包裹输出形式。搜索下拉字典预渲染：`{volist name="datas" id="d"}<option value="{$d.dict_value}">{$d.dict_label}</option>{/volist}`（datas = DictService::listByType 预 assign，下划线键）。
5. **select2 页面 5**：include system/include/select2-css + select2-js（已有）；重置后 `$("#dictType").trigger("change")` 联动。
6. **字典抽屉**（页面 1 内嵌）：view(dictId) 从当前页数据取 dictName/dictType → POST /system/dict/data/list（pageSize=100）→ 前端渲卡片；status 徽章/listClass badge/cssClass span 逻辑全在前端 JS（照抄 type.html 内嵌脚本），后端零增量。
7. **事务边界**：仅 updateDictType（数据行 dict_type 级联 + 类型行更新）包显式事务；三个 remove 循环**不包事务**（经典版无事务，见特殊行为 3）。
8. **预置数据基线**（动工检查单#2/#5 实测）：sys_dict_type 10 行（用户性别/菜单状态/系统开关/任务状态/任务分组/系统是否/通知类型/通知状态/操作类型/系统状态）；sys_dict_data 29 行（sys_oper_type 10 行、sys_user_sex 3 行、其余 8 类各 2 行）；sys_config 11 行**全部内置 Y**。本模块 CRUD 全物理删除（无 del_flag 列），测试数据自建自删。
9. **路由注册**：route/app.php 追加三个 group（system/dict、system/dict/data、system/config）；**固定段路由（refreshCache/detail/checkDictTypeUnique/treeData/selectDictTree/checkConfigKeyUnique）注册在含 :param 的路由之前**，防 `:dictId` 类单段通配吞并固定路径（AGENTS.md 坑 7 的 TP 版对应——listTop 案例 2.0.0 已踩）。
10. **status 字段枚举**：三表 status/configType/isDefault 均沿用字典语义（sys_normal_disable / sys_yes_no），页面 radio/下拉全部经 DictService::listByType 渲染，不硬编码。

## 拟登记 deviations

经逐点核对（响应格式 / 权限串 / 校验文案 / 缓存联动语义 / 删除语义 / 导出列），**无行为级差异需登记**——复刻点均可与经典版行为一致。候补两条，实施时定：

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| 全局 XssFilter | yml `xss.enabled: true`，urlPatterns `/system/*,/monitor/*,/tool/*`（excludes `/system/notice/*`）——入参 HTML 字符转义后入库 | 本模块 dict/config 字段入参**不做全局转义**（PHP 无 servlet wrapper 生态，3.0.0 起未挂）；输出侧由前端 JS escapeHtml 承担 | 实施时定；维持不复刻则登记 deviations（「输入侧全局 XSS 转义未复刻」），7.0.0 notice 模块同议题合并定 |
| Excel 列头样式 | POI ExcelUtil 固定样式 | 沿用 3.0.0 deviations #19 从简版（加粗+灰底） | 已有登记覆盖，无需新增 |

另注（非 deviation，防误改说明）：refreshCache 的 GET 方法与 remove 权限、内置参数不可删、「字典类型类型长度」重复措辞、remove 固定 success 非 toAjax、批量删除无事务，均为经典版原样行为，**照抄不修正**。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-09-29~30）

- DictService 收编完整版（类型域 + 数据域 + 缓存门面六方法）、ConfigService 补 set/loadAll/reset；listByType 双键输出三不变，消费方零回归。
- 三控制器 31 方法 + 路由 31 条；#[Perm] 27 / #[Log] 14（grep 实测）；CheckPerm/OperLog 经 config/route.php route 管线挂载。
- 10 页模板由 subagent 移植 + 引擎冒烟 20 用例；主线浏览器联调全链路（抽屉/弹窗/detail 页/缓存联动/页脚场景）。
- 联调期修复 2 处：① 路由 `add/:parentId` → `add`（spec 笔误，经典版 GET /add 无路径参数）；② detail() 漏 assign `$datas`（data/index 模板状态下拉需要）。
- 测试数据全清理复原（10/29/11）；PHPUnit 40 tests 118 assertions 全绿；quirk 清单逐项确认未"修复"。
