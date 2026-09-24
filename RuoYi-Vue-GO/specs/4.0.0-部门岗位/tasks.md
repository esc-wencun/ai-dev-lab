# Tasks · 04 部门岗位

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行 specs/README.md《动工检查单》，写实本文件夹三件套
  - 注：已对照 SysDeptController/SysPostController + ServiceImpl + Mapper XML 源码（2026-09-24），API 清单见 spec.md
- [x] 部门树/级联：ancestors 拼接（新增）、子孙前缀替换（修改）、启用祖先链（对位 updateParentDeptStatusNormal）
- [x] Task 部门 6 端点（list/exclude/getInfo/add/edit/remove）
  - 注：全部实现并有 curl 端到端验证（含 601 warn 信封两场景）；updateSort 本 Java 版定制未实现（前端无调用，登记 deviations）
- [x] Task 岗位 6 端点（list 分页/getInfo/optionselect/add/edit/remove 批量）
  - 注：全部实现；post/export 暂缺（与 5.0.0 导出一起补，见 spec 差异登记）
- [x] 单元测试：service 层经 sqlite 内存库间接覆盖（profile/dept 复用同一基建）；端到端 curl 全场景验证
  - 注：本模块无纯逻辑算法（树构建在内存、SQL 直查），测试以端到端为主
- [x] 端到端：部门增/同名拒/上级自己拒/有子 601/有用户 601/软删；岗位增/已分配拒/物理删
  - 注：2026-09-24 全部通过；测试数据已清理（部门 200 软删后物理清除、岗位 test04 物理删、Redis 会话清空）
