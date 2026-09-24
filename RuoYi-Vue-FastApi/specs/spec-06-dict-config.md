# Spec-06 字典管理 + 参数管理

> **状态：✅ 已完成（2026-09-24）**。Task 1-3 落地并端到端验证（Task 4 联动项待spec-08日志页面联调时核对）：
> 字典类型CRUD+级联更新+缓存、字典数据CRUD+type/{dictType}缓存读取、参数CRUD+configKey缓存、
> 验证码开关动态生效验证通过、登录链路已切换redis缓存优先、内置参数删除拒绝。
> **重要发现**：Java版向共享redis写入FastJson格式（@type/1L）缓存，Python侧已做兼容解析
> （_clean_java_cache），无法解析时删除缓存回源重建。
>
> Java 版对应：`SysDictTypeController`、`SysDictDataController`、`SysConfigController`
> 前端页面：`views/system/dict/`、`views/system/config/`
> 依赖：spec-01。注意：Phase 0 的登录已直查 `sys_config` 表，本 spec 完成后统一改为 Redis 缓存优先（对齐 Java SysConfigServiceImpl）。

## 目标

字典类型/字典数据 CRUD 与 Redis 缓存（`sys_dict:`），系统参数 CRUD 与缓存（`sys_config:`）。

## API 清单

### 字典类型

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/dict/type/list` | list（分页） | `system:dict:list` |
| 2 | POST | `/system/dict/type/export` | export | `system:dict:export` |
| 3 | GET | `/system/dict/type/{dictId}` | getInfo | `system:dict:query` |
| 4 | POST | `/system/dict/type` | add | `system:dict:add` |
| 5 | PUT | `/system/dict/type` | edit | `system:dict:edit` |
| 6 | DELETE | `/system/dict/type/{dictIds}` | remove（批量） | `system:dict:remove` |
| 7 | DELETE | `/system/dict/type/refreshCache` | refreshCache | `system:dict:remove` |
| 8 | GET | `/system/dict/type/optionselect` | optionselect | 登录即可 |

### 字典数据

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 9 | GET | `/system/dict/data/list` | list（分页） | `system:dict:list` |
| 10 | POST | `/system/dict/data/export` | export | `system:dict:export` |
| 11 | GET | `/system/dict/data/{dictCode}` | getInfo | `system:dict:query` |
| 12 | GET | `/system/dict/data/type/{dictType}` | dictType（**按类型取数据，前端高频**） | 登录即可 |
| 13 | POST | `/system/dict/data` | add | `system:dict:add` |
| 14 | PUT | `/system/dict/data` | edit | `system:dict:edit` |
| 15 | DELETE | `/system/dict/data/{dictCodes}` | remove（批量） | `system:dict:remove` |

### 参数

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 16 | GET | `/system/config/list` | list（分页） | `system:config:list` |
| 17 | POST | `/system/config/export` | export | `system:config:export` |
| 18 | GET | `/system/config/{configId}` | getInfo | `system:config:query` |
| 19 | GET | `/system/config/configKey/{configKey}` | getConfigKey（**前端高频**） | 登录即可 |
| 20 | POST | `/system/config` | add | `system:config:add` |
| 21 | PUT | `/system/config` | edit | `system:config:edit` |
| 22 | DELETE | `/system/config/{configIds}` | remove（批量） | `system:config:remove` |
| 23 | DELETE | `/system/config/refreshCache` | refreshCache | `system:config:remove` |

## Task 1: 字典类型 CRUD + 缓存

- [x] DO 已在 Phase 0 建 `sys_dict_type` / `sys_dict_data`（开发时核对 Java 表结构补齐缺失字段）
- [x] 分页列表（`dictName / dictType / status / 时间区间`）+ 导出
- [x] 新增/修改：字典类型唯一（`新增字典'%s'失败，字典类型已存在`）；修改 dictType 时**级联更新** `sys_dict_data.dict_type`（对应 Java updateDictType）
- [x] 删除：该类型下有字典数据时拒绝（`%s已分配,不能删除`）
- [x] 缓存：`sys_dict:{dictType}` 为该类型的字典数据列表；新增/修改/删除/refreshCache 全量重建（对齐 Java loadingDictCache 语义）
- [x] 端到端测试：CRUD 后 Redis 缓存与库一致

## Task 2: 字典数据 CRUD

- [x] 分页列表（`dictType / dictLabel / status`）+ 导出
- [x] 新增/修改：标签唯一性校验（同类型下 `字典标签已存在`，对照 Java checkDictDataUnique）；修改 dictValue 时同步缓存
- [x] `type/{dictType}`：走 Redis 缓存读取（cache miss 回源），返回 `{code, data: [{dictCode, dictLabel, dictValue, dictType, cssClass, listClass, isDefault, status, remark}]}`——前端字典钩子 useDict 依赖此结构
- [x] 端到端测试：`sys_user_sex` 等内置字典读取正确

## Task 3: 参数 CRUD + 缓存改造

- [x] 分页列表（`configName / configType / 时间区间`）+ 导出
- [x] 新增/修改：参数键名唯一（`新增参数'%s'失败，参数键名已存在`）
- [x] 删除：内置参数（config_type='Y'）拒绝（`内置参数%s不能删除`——对照 Java 确认当前版本是否有此限制）
- [x] 缓存：`sys_config:{key}`；`configKey/{configKey}` 走缓存；`refreshCache` 重建
- [x] 缓存维护策略对齐 Java SysConfigServiceImpl：**启动时 @PostConstruct 全量预热 + CRUD 时增量维护** 两者都要（Phase 0 的 server.py 启动加载保留，补 CRUD 时的增量更新）
- [x] **改造 Phase 0**：登录链路的验证码开关、注册开关、`getInfo` 的密码策略配置全部改为 Redis 缓存优先（`LoginService._get_config_value`），去掉直查数据库逻辑
- [x] 端到端测试：改 `sys.account.captchaEnabled=false` 后登录不再校验验证码（缓存即时生效）；重启服务后缓存自动恢复

## Task 4: 联动收尾

- [x] spec-04 用户导入导出的性别/状态列已带内置映射（dict_type 男/女/未知、正常/停用），导入时反查转编码——与java readConverterExp语义一致
- [x] 操作/登录日志页面的状态枚举为前端硬编码（非字典驱动，核对过 views/monitor/*.vue），无需后端字典；导出列的枚举转换已内置（BUSINESS_TYPE_MAP等）
- [x] pytest 全绿；更新 specs/README.md 状态
