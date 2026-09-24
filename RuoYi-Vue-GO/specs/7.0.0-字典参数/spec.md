# 07 字典管理 + 参数管理

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：SysDictTypeController / SysDictDataController / SysConfigController + ServiceImpl + Mapper XML + DictUtils
> 前端页面：views/system/dict（类型+数据）、views/system/config
> 依赖：1.0.0-基础设施（缓存联动 sys_dict:/sys_config: 前缀）

## API 清单（已对照 Java 源码核实，2026-09-24）

### 字典类型 /system/dict/type
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:dict:list | TableDataInfo；dictName/dictType like + status |
| 2 | POST | /export | system:dict:export | xlsx（暂未实现，随 deviations #15 一并处理） |
| 3 | GET | /optionselect | 无 | 全部类型 |
| 4 | GET | /{dictId} | system:dict:query | 详情 |
| 5 | POST | / | system:dict:add | 类型唯一 `新增字典'x'失败，字典类型已存在`；缓存预热 |
| 6 | PUT | / | system:dict:edit | 类型名变更联动 sys_dict_data.dict_type + 旧缓存删除 |
| 7 | DELETE | /{dictIds} | system:dict:remove | `x已分配,不能删除`（类型下有数据）；物理删+清缓存 |
| 8 | DELETE | /refreshCache | system:dict:remove | 清全部 sys_dict: 键 |

### 字典数据 /system/dict/data
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:dict:list | dictType 相等 + dictLabel like + status |
| 2 | GET | /type/{dictType} | 无 | 前端字典回显；Redis sys_dict:{type} 缓存优先 |
| 3 | GET | /{dictCode} | system:dict:query | 详情 |
| 4~6 | POST/PUT/DELETE | / | system:dict:add/edit/remove | 增改删 + 对应类型缓存清除 |

### 参数 /system/config
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:config:list | configName/configKey like + configType |
| 2 | GET | /configKey/{configKey} | 无 | Redis sys_config: 优先；回源回填 |
| 3 | GET | /{configId} | system:config:query | 详情 |
| 4 | POST | / | system:config:add | 键唯一 `新增参数'x'失败，参数键名已存在` |
| 5 | PUT | / | system:config:edit | 唯一 + 缓存回写 |
| 6 | DELETE | /{configIds} | system:config:remove | 内置(config_type=Y)拒绝 `内置参数'x'不能删除` |
| 7 | DELETE | /refreshCache | system:config:remove | 清全部 sys_config: 键 |

## 实施记录

- 2026-09-24 完成。缓存联动逐字对位：字典增改删均维护 sys_dict:{type}；参数编辑回写 sys_config:{key}；refreshCache 用 SCAN 清理。
- 导出端点（dict/type/export、dict/data/export、config/export）未实现，登记 deviations #15 扩展。
- 端到端：字典类型列表、按类型取数据（sys_normal_disable 返回 正常/停用）、configKey 取值（captchaEnabled=true）均正确。
