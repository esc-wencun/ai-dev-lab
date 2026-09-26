# Checklist · 0.0.0 MyBatis-Plus 功能增加

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。
> 共用库纪律：本项目连本地 Docker MySQL（ry-vue）验证；端到端测试造的数据（测试岗位/公告/日志等）测完必须清理。

## 验收清单

- [x] tasks.md 的 9 个 Task 全部勾选（含各 Task 内单元测试与端到端验证）
- [x] `mvn clean package -Dmaven.test.skip=true` 全模块编译通过（2026-09-26 最终版）
- [x] **PageHelper 零残留**：全局 grep `com.github.pagehelper`（含 .java/.xml/.vm/pom）0 命中（2026-09-26 复核）
- [x] **MP 生效判据**：生产代码 BaseMapper 路径跑通——post 模块单表 CRUD 走 MP 注入模板（SQL 日志大写 INSERT INTO/UPDATE...WHERE 确认），分页走 PaginationInnerInterceptor
- [x] **契约零变化自查**：分页接口请求参数（pageNum/pageSize/orderByColumn/isAsc/reasonable）与响应 JSON（TableDataInfo{total,rows,code,msg}）结构一致；Task 4~8 各端到端均含响应字段断言（total/rows/code/msg）
- [x] 前端接口全量回归（2026-09-26 Task 8 冷启动）：登录 + post/user/role/menu/dept/dict-type/dict-data/config/notice（含 readUsers）/operlog/logininfor/online/job/joblog/gen（含 db/list）/allocated/unallocated 共 19 项接口 + post CRUD 往复，全绿（浏览器页面级人工操作留待用户日常使用确认，接口层 curl 已全过）
- [x] BaseService 生效：4 个 ServiceImpl extends BaseService 且 checkUnique 走公用方法（Post/Config/DictType/DictData；日志模块按坑 6 不适用，User/Role/Menu/Dept 本期范围外）
- [x] **数据权限回归**：造 deptId=100 + 角色2（dataScope=2）测试用户实测——其 user list 仅见本部门 2 人，admin（研发部）不可见，`${user.params.dataScope}` 在 Page 版语句正确生效；测试用户已删
- [x] 验证测试数据清理（测试岗位/字典/参数/公告/用户全部删除；DB 实查 0 残留；admin 会话已清，密码未动过保持 admin/admin123）
- [x] 文档同步完成：Python specs/README 对位表、GO tech-stack ORM 行、工作区 readme（2 处）、根 AGENTS.md 工作区结构行
- [x] `RuoYi-Vue/specs/README.md` 模块总表标 ✅ + 日期（2026-09-26）

## 遗留与决策记录

- pageSize 无上限：现状行为，本次保持（不悄悄改契约）；是否加上限留待后续决策。（2026-09-25 立项时登记）
- reasonable（pageNum 超界回正）无 MP 等价物（坑 2）：本期不启用 overflow（超界返回空列表），前端 total 驱动翻页正常触不到；如需保护后续全局开 `PaginationInnerInterceptor.setOverflow`。（2026-09-26 开发实查修正）
- IService/ServiceImpl MP 体系不引入，用自建 BaseService——决策见 spec.md 第一节"不做"。（2026-09-25 用户决策）
- 复杂实体（SysUser/SysRole/SysMenu/SysDept）的 BaseMapper 迁移、生成器模板全套 MP 化（mapper.java.vm/mapper.xml.vm）：本期范围外，后续模块再议。
- 日志/无审计列表（sys_oper_log/sys_logininfor/sys_job_log）走 IPage 首参 + XML 模式（坑 6），不迁 BaseMapper——后续同型表沿用。
- Java 版单测基建已建（surefire 3.5.3 + junit-jupiter 5.12.2 test scope，坑 3），后续模块单测直接用。
- 浏览器页面级人工验收：接口层 curl 已全过；前端页面翻页/排序/筛选的人工点验随日常使用确认即可（非阻塞项）。
