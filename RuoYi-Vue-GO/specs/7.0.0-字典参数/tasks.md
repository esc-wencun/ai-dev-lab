# Tasks · 07 字典参数

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：对照三 Controller + ServiceImpl + Mapper XML + DictUtils 写实 API 清单
- [x] Task 字典类型 7 端点（list/optionselect/getInfo/add/edit/remove/refreshCache；export 暂缺登记 deviations #15）
- [x] Task 字典数据 6 端点（list/type/{dictType}/getInfo/add/edit/remove；缓存优先回显）
- [x] Task 参数 7 端点（list/configKey/{key}/getInfo/add/edit/remove/refreshCache；内置删除保护）
- [x] 缓存联动：sys_dict:{type} 增改删维护、sys_config:{key} 回写、refreshCache SCAN 清理
- [x] 端到端：字典类型列表/按类型取数据/configKey 取值（2026-09-24 通过；Redis 测试键清空）
