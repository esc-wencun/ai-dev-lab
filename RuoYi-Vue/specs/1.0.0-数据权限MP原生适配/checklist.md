# Checklist · 1.0.0 数据权限 MP 原生适配

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。
> 共用库纪律：连本地 Docker MySQL（ry-vue）验证；数据权限专测必须造受限角色 + 测试用户（admin 全权限测不出失效），测完清理。

## 验收清单

- [x] tasks.md 的 6 个 Task 全部勾选（含各 Task 内单元测试与端到端验证）
- [x] `mvn clean package -Dmaven.test.skip=true` 全模块编译通过（按编译纪律 A 类执行：动 mvn 前停 IDEA 服务并明说，验证完停掉后台进程）
- [x] **语义完整性**：5 种 dataScope 类型 × DTO→Expression 单测断言全绿（IN/=/(A OR B)/null=null 兜底）
- [x] **拦截器路径端到端**：临时白名单注册 biz_order 测试表后，受限角色（本部门及以下 dept105 + self）经纯 BaseMapper 查询只见 ry-order-1/2 共 2 条（103/106 的数据不可见），total=2 与明细一致；admin 全量 4 条；executeIgnore 豁免 4 条；无注解放行 4 条；无登录态被 Security 拦截。其余 scope 类型（仅本人/自定义/全部）由同一条 DTO→Expression 链路覆盖并有单测断言（受限角色多 scope 实测因需多角色用户矩阵，以单测+单一受限角色实测组合覆盖——如实登记，非全实测）
- [x] **XML 路径回归**：现有 5 处 `${params.dataScope}` 查询（用户×3/角色/部门）在受限角色下行为与改造前一致（双路径互不干扰的核心证据）
- [x] **ThreadLocal 安全**：单测覆盖嵌套与异常路径清理；handler 层上下文未激活时对全项目其余查询零影响（抽 3 个无关接口回归）
- [x] **性能**：JSqlParser 解析缓存生效；核心列表接口响应时间对比数据登记在 spec 实施记录
- [x] 测试数据清理：测试表/测试角色/测试用户删除，**白名单恢复空表**，admin 会话复原 admin/admin123
- [x] 文档同步 3 处完成（项目 AGENTS.md MP 约定、0.0.0 红线更新、specs/README.md 台账）
- [x] `RuoYi-Vue/specs/README.md` 模块总表标 ✅ + 日期

## 遗留与决策记录

- 白名单本期空表起步：现有 4 张走 BaseMapper 的表（post/config/dict_type/dict_data）均无范围语义列，注册随首个带范围语义的纯 MP 业务查询出现时进行（注册动作本身走 checklist 实测）。（2026-09-26 方案定稿）
- XML 5 处 `${params.dataScope}` 不迁移拦截器：双路径并存是既定方案（spec 第四节方案 A 否决记录），全量统一留待后续模块评估。（2026-09-26）
- `selectById/updateById/deleteById` 内置方法不短路（dcoa 有此性能宽口子，RuoYi 安全优先不做）：单条查询同样受数据权限约束。（2026-09-26 决策）
- 非 Web 上下文（定时任务等）默认放行不过滤：与 dcoa 语义一致；若未来定时任务需按某用户范围过滤，另行设计上下文构造器。（2026-09-26 决策）
- "本部门及以下"改用 `selectChildrenDeptById` 预展开 ID 集 + IN（dcoa 路线，索引友好），弃用 find_in_set——与现有 XML 路径的 find_in_set 行为存在实现差异但结果语义一致，端到端专测覆盖。（2026-09-26 决策）
- **角色数据源用登录快照**（`LoginUser.user.getRoles()`，与切面同源）：不查实时 sys_user_role——避免"切面用快照、拦截器用实时库"分裂；代价是角色变更需重新登录生效（与现状一致）。（2026-09-26 方案自查修正）
- **显式注解才过滤**（区别于 dcoa 默认开启）：不标 `@DataScope` 的查询零影响，保住"零行为变化"红线；安全取舍是"新代码忘写注解会漏过滤"，以 code review + checklist 数据权限专条兜底。（2026-09-26 方案自查修正）
- resolver 与切面为并存的两份判断实现（切面本体不动的务实取舍）：语义对齐靠单测对照锁定（同一输入两路产出必须语义一致）；全量统一（方案 A）留待后续评估。（2026-09-26 方案自查修正）
