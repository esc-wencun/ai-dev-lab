# 全量完成总验收（Final Acceptance）

> 所有模块（00~11，见 [README.md](README.md) 总表）完成后的整体回归清单。单个模块的验收在各自文件夹 checklist.md 中，本清单只管"整体是否真正对齐 Java 版"。
> Python 版同款清单已跑过一轮，本清单为 GO 版执行版（结构对齐 Python 版 final-acceptance.md，新增差异核对节）。

## A. 接口完整性

- [ ] 前端 `src/api/**/*.js` 全部 URL 与 GO 版路由逐一对照，无 404（写脚本扫描比对，输出对照表）
- [ ] Java 版全部 `@RequestMapping` 端点与 GO 版路由对照，差异均登记在 [deviations.md](deviations.md)
- [ ] RuoYi-Vue3 全站人工过一遍：登录 → 首页 → 系统管理五页 → 系统监控六页 → 系统工具，无报错弹窗、无 undefined 渲染

## B. 行为一致性抽查

- [ ] 错误码约定全站一致：业务失败 HTTP 200 + code 500；警告 code 601；无权限 code 403；未登录 code 401（body）
- [ ] 分页返回 `{code, msg, rows, total}`；单对象返回 `{code, msg, data}`；自由字段（token/uuid/img 等）与 Java 命名一致
- [ ] **日期字段全站抽查**：所有列表/详情的时间显示为 `yyyy-MM-dd HH:mm:ss`（Go time.Time 默认 RFC3339 的坑已全局规避——0.0.0-工程基础驼峰时间 Task）
- [ ] 所有面向用户的错误文案与 Java messages.properties / ServiceImpl 文案一致（抽查 20 条）
- [ ] 密码策略、锁定策略、验证码开关等行为参数化项与 Java 配置行为一致
- [ ] 密码错误锁定、权限变更踢下线/刷权限、停用踢下线等跨模块联动与 Java 行为一致

## C. 工程质量

- [ ] `gofmt -l .` 无输出；`go vet ./...` 零告警；`go test ./...` 全绿（各 spec 单元测试累计 ≥ 60 个）
- [ ] 分层规范执行抽查：handler 无 db/redis 直接访问、dao 无业务判断、业务代码无裸 Redis key 前缀、vo json tag 全显式驼峰
- [ ] context 传播抽查：dao 层无 `context.Background()` 逃逸（全部来自请求 context）
- [ ] 端到端测试数据零残留（共用库无测试用户/角色/公告遗留；admin 密码 admin123；无测试会话键）
- [ ] README.md 与 specs/README.md 状态同步；tech-stack.md 选型对位表更新实际落地情况

## D. 差异核对

- [ ] [deviations.md](deviations.md) 逐条核对：每条差异仍为有意差异、未引入新的未登记差异、对前端影响面为零

## E. 文档收尾

- [ ] specs/README.md 总表全部标 ✅ + 实际完成日期
- [ ] 遗留决策（golangci-lint 是否启用、asynq 是否引入、测试库 schema 是否建立）在总表中给出结论或再次延期理由
- [ ] 各 spec"实施记录"与"测试数据清理记录"完整（无空占位）
