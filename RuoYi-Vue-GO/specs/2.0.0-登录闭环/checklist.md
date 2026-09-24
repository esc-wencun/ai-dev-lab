# Checklist · 02 登录闭环

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。

## 验收清单

- [x] tasks.md 的 4 个 Task 全部勾选
- [x] RuoYi-Vue3 登录页正常出验证码、登录进入首页、菜单渲染正确、退出回登录页
  - 注：2026-09-24 浏览器页面级验收通过（Tabbit + RuoYi-Vue3 @localhost，Go 版后端）——登录页算术验证码出图（截图题目 3-3=? 与 Redis 存储答案一致）→ admin/admin123 登录成功跳 /index → 左侧菜单树完整（首页/系统管理/系统监控/系统工具/若依官网，子菜单 22 项与 Java 版 admin 一致）→ 顶部用户菜单退出、确认框后回登录页。**验收中发现并修复 1 个 bug**：退出/注册日志状态串（"Logout"/"Register"）未按 Java AsyncFactory 归一化为 char(1)，落库报 Data too long 致退出日志静默丢失——修复后 curl 复验退出日志正常落库（status=0/msg=退出成功）、unlockscreen 三分支端到端补验通过，详见 spec.md 实施记录
- [x] 密码连错 5 次锁定、锁定文案正确；锁定期间正确密码也拒登
  - 注：单元测试 TestValidatePasswordRetry 全覆盖（阈值/文案/锁定期间拒登/清零）；Redis pwd_err_cnt: 计数在真实环境验证（错 3 次=3）
- [x] 与 Java 版互验：Go 签发的 token 能被 Go 自己续期，且 Java 库存的 bcrypt 密码可直接登录（同库同 Redis 冒烟）
  - 注：已验证——admin（Java 存量 $2a$ 哈希）+ admin123 登录成功拿 token；token 多次请求 getInfo/getRouters 有效（自动续期链路工作）；会话写入共享 Redis login_tokens:（JSON 形态，重启后 Go 自身可解析）
- [x] `go test ./...` 全绿；更新 specs/README.md 状态
  - 注：2026-09-24 复验全绿（含新增 TestLogininforStatus），gofmt/vet 干净
- [x] 端到端测试数据清理完毕（sys_logininfor 测试记录、captcha_codes/pwd_err_cnt/login_tokens 测试键、admin 会话复原）
  - 注：2026-09-24 清理——sys_logininfor 删除 info_id>=103（保留历史 100~102）；Redis 清空全部 captcha_codes:/login_tokens:/pwd_err_cnt:admin；admin 密码未改动（登录用的是存量哈希）；admin 无 Java 侧活跃会话需复原

## 测试数据清理记录

- 2026-09-24（第一轮，curl 端到端后）：端到端产生 sys_logininfor 记录 6 条（info_id 103~108：2 成功 + 4 失败）已删除；Redis captcha_codes:*、login_tokens:*（Go 会话 2 个）、pwd_err_cnt:admin 已清空；sys_user.admin 的 login_ip/login_date 为运行时字段被更新（非污染，Java 登录同样更新），未复原。
- 2026-09-24（第二轮，浏览器页面验收 + bug 修复复验后）：删除 sys_logininfor info_id 113~135（23 条：页面/复验产生的登录成功 + 退出成功 + 验证码失败），保留历史 100~102；Redis login_tokens:*（1 个）清空，captcha_codes:*/pwd_err_cnt:* 无残留（TTL 自然过期/未产生）；admin 密码哈希未改动（$2a$10$ 存量），login_ip/login_date 运行时字段同前未复原。
