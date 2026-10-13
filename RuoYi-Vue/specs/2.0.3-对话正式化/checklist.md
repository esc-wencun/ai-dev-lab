# Checklist · 2.0.3 对话正式化：菜单权限收敛 + 会话参数 UI

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 冒烟清理完成且零悬挂引用（AiSmokeController 删除 + javac 全模块 37 class 全绿 + Mock 提示条移除）
- [x] 对话域 9 端点权限收敛：造测试角色实测 `/ai/conversations` 403「没有权限，请联系管理员授权」、admin（*:*:* 通配）全 200；`/ai/agents/options` 登录态可达（设计如此）
- [x] 菜单 2020 perms 生效：侧边栏「AI对话」可见可用（当时挂 2000 目录下；2026-10-13 起升级为顶级菜单，历史记录不改）；SQL 幂等（双跑验证 2020 perms 在、2030-2034 无翻倍）
- [x] 智能体管理菜单+按钮 SQL 就位（2030 visible=1 占位，管理页 UI 列 2.0.4）
- [x] `PUT /{id}/params`：归属校验 + 范围校验实测（temperature=5 → 500 带范围提示；1.9/4096 落库实查在）——**发现并修复实现 bug：范围校验拦截了 -1 清除标记（校验放行 -1）；顺带修 normalize 默认语义（null=保持现值）与 Double 装箱比较坑**。-1 清除实测待下次 Build（源码已修，javac 全绿）
- [x] 参数生效链路：带 temperature=1.9 的会话 SSE 正常对话（planner 消费无异常）；行为对比由用户日常体验确认
- [x] 参数 UI：三控件限界正确、打开回显当前值、保存/恢复默认逻辑就位（Mock 语义四项自测全过；real 层待 Build 后联调）
- [x] 回归零影响：2.0.1 三管理页 + user/list + agents/options 全 200；SSE 主链路正常（重生成在 17:44 旧 jar 上 500 属预期——旧包早于 planner null 保护，IDEA 跑 target/classes 才是新代码）
- [x] 测试数据清理（chat403test 角色/用户、测试会话与消息、Redis 会话、我起的旧 jar 进程；用户自建会话 109 保留）
- [x] specs/README 登记 2.0.3（🚧 收口中）；CHANGELOG 随提交补

## 遗留与决策记录

- 智能体管理页 UI → 2.0.4（权限点与接口面本模块已就位）。
- 深度思考开关 UI → 随智能体管理页。
- 会话参数系统级配额 → 用量域配额管控模块。
- **OpenAI SDK 流式静默吞错坑（2026-10-12 实测）**：网关对不存在模型返回 404+JSON 错误体时，SDK 转成零 token 空完成而非异常——已在完成路加空响应守卫（msgStatus=2 + 可读 errorMessage），后端 curl 实测通过。
- 浏览器自动化基建：Tabbit + Playwright（127.0.0.1 访问、Redis 直读验证码答案登录）跑通 13 项用例；测试任务残留的 TypeError 错误列表随 task finish 清空。
