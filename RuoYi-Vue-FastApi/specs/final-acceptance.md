# 全量完成总验收（Final Acceptance）

> 所有 spec（00~10）完成后的整体回归清单。单个 spec 的验收在各自文件中，本清单只管"整体是否真正对齐 Java 版"。

## A. 接口完整性

- [ ] 前端 `src/api/**/*.js` 全部 URL 与 Python 版路由逐一对照，无 404（写脚本扫描比对，输出对照表）
- [ ] Java 版全部 `@RequestMapping` 端点与 Python 版路由对照，确认差异均为有意裁剪（如 swagger 反代方式不同）并记录在案
- [ ] RuoYi-Vue3 全站人工过一遍：登录 → 首页（公告/统计）→ 系统管理五页 → 系统监控六页 → 系统工具，无报错弹窗、无 undefined 渲染

## B. 行为一致性抽查

- [ ] 错误码约定全站一致：业务失败 HTTP 200 + code 500；警告 code 601；无权限 code 403；未登录 code 401（body）
- [ ] 分页返回 `{code, msg, rows, total}`；单对象返回 `{code, msg, data}`；自由字段（token/uuid/img 等）与 Java 命名一致
- [ ] 所有面向用户的错误文案与 Java messages.properties / ServiceImpl 中的文案一致（抽查 20 条）
- [ ] 密码策略、锁定策略、验证码开关等行为参数化项与 Java 配置行为一致

## C. 工程质量

- [ ] `ruff check` 零 error；`pytest tests/` 全绿（各 spec 的单元测试累计 ≥ 60 个）
- [ ] 端到端测试数据零残留（共用库中无测试用户/角色/公告遗留；admin 密码为 admin123）
- [ ] 分层规范执行情况抽查：controller 无 db/redis 直接访问、dao 无业务判断、业务代码无裸 Redis key 前缀
- [ ] README.md 与 specs/README.md 状态同步为全部完成；技术选型对位表更新实际落地情况

## D. 文档收尾

- [ ] specs/README.md 总表全部标 ✅，记录实际完成日期
- [ ] 遗留决策（sqlalchemy-crud-plus 是否引入、mypy 是否启用、测试库 schema 是否建立）在总表中给出结论或再次延期理由
