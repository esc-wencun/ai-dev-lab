# 全量完成总验收（Final Acceptance）

> 全部模块（00~08，见 [README.md](README.md) 总表）完成后的整体回归清单。单个模块的验收在各自文件夹 checklist.md 中，本清单只管「整体是否真正对齐 RuoYi-Vue3 基准」。
> 「对齐」的含义：功能等价（每页可操作、行为契约逐条复现），界面允许差异（deviations 已登记的除外）。

## A. 功能完整性

- [ ] reference/01 的 51 页面清单逐一对照（tool/build 按 deviations #2 排除）：每页有对应实现或已登记差异
- [ ] reference/01 §2 的 20 个接口文件全部端点在 api 层有对应函数，URL/方法/参数结构与基准逐一比对
- [ ] RuoYi-React 全站人工过一遍（Java 后端）：登录 → 首页 → 系统管理八页 → 系统监控八页 → 系统工具两页，无报错弹窗、无 undefined 渲染
- [ ] 与 RuoYi-Vue3（80 端口）同屏并行对照：任取五页，查询参数/请求体/响应处理行为一致

## B. 行为一致性抽查

- [ ] 网络层契约：HTTP 恒 200 + body code 分支（401 确认框 / 500 message / 601 warning / 其他 notification）；与基准逐形态对照
- [ ] 防重复提交：sessionStorage `sessionObj`、1000ms、≥5MB 跳过、repeatSubmit=false 关闭——四条逐项验证
- [ ] 下载导出：POST + form-urlencoded + blob + 错误 JSON 二次解析；导出文件可打开
- [ ] token 会话级 Cookie `Admin-Token`；Bearer 前缀；isToken:false 跳过
- [ ] 权限：`*:*:*` / 角色 admin 通配；`<Auth>` 无权限不渲染；dynamicRoutes 五页按权限显隐；ROLE_DEFAULT 语义
- [ ] 动态路由：后端菜单任意三级结构（含 ParentView）正确挂载；外链菜单新窗口；InnerLink iframe 常驻切换
- [ ] 锁屏：守卫硬劫持、解锁回原路径、退出重登
- [ ] 提示文案抽查 20 条与基准一致（「系统提示」「登录状态已过期…」「数据正在处理，请勿重复提交」等）

## C. 三版后端通用性

- [ ] Java 后端全量走查（A/B 两节即覆盖）
- [ ] Python 后端：登录链路 + 系统管理核心页 CRUD + druid/swagger 降级 + serverMonitor 正常 + 导出一条链路
- [ ] Go 后端：登录链路 + 核心 CRUD 抽查 + druid/server/swagger 三页全部降级提示
- [ ] features 三字段消费正确（初值全 true、`response.features || {}` 覆盖、重新检测按钮），不判断后端语言

## D. 工程质量

- [ ] `npm run build:prod` 零错误；`npm test` 全绿（纯逻辑单测：tansParams/parseTime/handleTree/passwordRule/filterAsyncRouter 拍平/cachedViews 规则/settings 合并）
- [ ] 请求层/路由守卫无「顺手优化」：与 reference/02 契约逐条对照，diff 可追溯
- [ ] 端到端测试数据零残留（共用库无测试用户/角色/公告遗留；admin 密码 admin123；无测试会话键）
- [ ] specs/README.md 总表与实际状态同步

## E. 差异核对与文档收尾

- [ ] [deviations.md](deviations.md) 逐条核对：每条仍为有意差异、未引入新的未登记差异、接口契约零差异
- [ ] KeepAlive 演示待办闭环（演示后：实现 → 移除 #1，或放弃 → #1 定稿）
- [ ] RuoYi-React 自身 readme.md 与根 AGENTS.md/readme.md 状态同步（无漂移）
