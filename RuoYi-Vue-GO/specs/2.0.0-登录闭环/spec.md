# 02 登录闭环：验证码 / 登录 / getInfo / getRouters / logout / 解锁屏幕

> **状态：✅ 已完成（2026-09-24）**（4 Task 全部实现；curl 端到端 + RuoYi-Vue3 浏览器页面级验收均通过；验收中发现退出/注册日志状态未归一化的 bug，已修复并复验，见实施记录）
>
> 对位 Python 版 Phase 0；**本模块的 API 清单已对照 Java 版源码逐端点核实**（含该版本特有的密码策略字段与 unlockscreen），**作为其他模块 API 清单的格式范本**。
> Java 版对应：`CaptchaController`、`SysLoginController`、`SysIndexController`、`SysLoginService`、`TokenService`、`SysPasswordService`、`LogoutSuccessHandlerImpl`。
> 前端页面：登录页、首页、顶部用户菜单（个人中心入口/退出/解锁）。
> 依赖：1.0.0-基础设施（认证中间件/白名单机制/操作日志）。

## API 清单（已核实，2026-09-24 对照 Java 源码）

| # | 方法 | 路径 | 认证 | 参数 | 返回要点 |
|---|------|------|------|------|---------|
| 1 | GET | `/captchaImage` | 免认证 | 无 | `{code,msg,captchaEnabled,uuid,img}`；img 为 jpg 的 base64（无 data: 前缀）；math 型题目形如 `1+2=@`；验证码值存 `captcha_codes:{uuid}` TTL 2 分钟；开关 `sys.account.captchaEnabled` 关闭时只返回 captchaEnabled=false |
| 2 | POST | `/login` | 免认证 | body JSON `{username,password,code,uuid}` | `{code,msg,token}`；全链失败均 HTTP 200 |
| 3 | GET | `/getInfo` | Bearer | 无 | `{code,msg,user,roles,permissions,pwdChrtype,isDefaultModifyPwd,isPasswordExpired}`（后三字段为**本 Java 版定制**，标准 RuoYi 无） |
| 4 | GET | `/getRouters` | Bearer | 无 | `{code,msg,data:[菜单路由树]}` |
| 5 | POST | `/logout` | Bearer | 无 | `{code:200,msg:退出成功}`；删 Redis 会话 + 记登录日志 |
| 6 | POST | `/unlockscreen` | Bearer | body JSON `{password}` | 成功 `{code:200,msg:解锁成功}`；密码空 `密码不能为空`；错误 `密码错误，请重新输入`；用户不存在 `服务器超时，请重新登录`（本 Java 版定制端点） |
| 7 | ALL | `/` | 免认证 | 无 | 纯文本欢迎语（对位 SysIndexController.index） |

### /login 行为链（对位 SysLoginService.login，顺序敏感）

1. **验证码校验**（validateCaptcha）：开关开启时——`captcha_codes:{uuid}` 不存在→`验证码已失效`；取到即删（一次性）；忽略大小写比对失败→`验证码错误`。
2. **前置校验**（loginPreCheck）：用户名或密码空→`用户不存在/密码错误`；密码长度 5-20、用户名长度 2-20 越界→报密码不匹配；IP 黑名单（`sys.login.blackIPList`）命中→`很遗憾，访问IP已被列入系统黑名单`。
3. **密码校验**（SysPasswordService.validate）：先查 `pwd_err_cnt:{username}`，≥ maxRetryCount（`user.password.maxRetryCount`，默认 5）→`密码输入错误5次，帐户锁定10分钟`；bcrypt 不匹配→计数+1 写回（TTL lockTime 分钟）→`用户不存在/密码错误`；匹配→清计数。
4. **成功路径**：记登录日志（`Success`）→ `updateLoginInfo`（login_date/ipaddr）→ createToken：随机 UUID 作为会话键 → Redis `login_tokens:{uuid}` 存 LoginUser JSON（TTL 30 分钟，`token.expireTime`）→ JWT（HS512，secret 同 Java `token.secret`）claims 只放两个：`login_user_key`=uuid、subject=用户名 → 返回 `{code:200,msg:操作成功,token}`。

### getInfo 细节

- `user` 为用户完整 JSON（**不含 password**），roles 为角色权限字串集合（admin 为 `admin`），permissions 为菜单权限集合（admin 为 `*:*:*`）。
- 权限与 Redis 会话中不一致时刷新会话（对位 tokenService.refreshToken）。
- `pwdChrtype`=配置 `sys.account.chrtype`（默认 "0"）；`isDefaultModifyPwd`=配置 `sys.account.initPasswordModify`==1 且 pwd_update_date 为空；`isPasswordExpired`=配置 `sys.account.passwordValidateDays`>0 且（未改过密 或 距上次改密超 N 天）。

任务分解见 [tasks.md](tasks.md)，验收清单见 [checklist.md](checklist.md)。

## 实施记录

- 2026-09-24 开发完成。分层：dao/login_dao.go（用户三表联查/菜单树 SQL 逐字对位）、service/login_service.go（四步登录链/验证码/重试计数/黑名单）+ menu_service.go（buildMenus 三分支）、handler/login_handler.go（7 端点）、security/token_service.go（CreateToken 补齐）。
- 差异登记（同步 deviations.md）：① IP 黑名单分隔符 `,`（Java `;`+三形态匹配，按 Python 版语义简化）；② getInfo 权限变更即时回写会话未做（6.0.0 补）；③ `/` 欢迎语版本号写死；④ uuid 32 位 hex（Java simpleUUID 同长度同用途）；⑤ sys-user.log 文件日志未接线（6.0.0 前补）。
- 关键验证：Java 存量 bcrypt 哈希可直接登录（同库互验 ✓）；验证码答案 JSON 形态与 FastJson 一致；菜单路由树 admin/ry 双用户结构正确。
- 2026-09-24 浏览器页面验收通过（RuoYi-Vue3 @80）：验证码出图、admin 登录进首页、菜单树 22 项与 Java 版 admin 一致、顶部退出回登录页。**发现并修复 1 个 bug**：logout/register 日志状态串（constant.Logout/Register，照抄 Java Constants 的业务语义值）未落库前归一化，`sys_logininfor.status` 是 char(1)，插入报 `Data too long` 被异步 goroutine 吞掉——退出日志静默丢失（登录日志传的是 constant.Success="0" 恰好没暴露）。修复：login_service.go 增加 `logininforStatus()` 映射（对位 Java AsyncFactory.recordLogininfor 的 SUCCESS/FAIL 归一化）+ 单测 TestLogininforStatus；复验退出日志正常落库（status=0/msg=退出成功，浏览器退出时 browser/os 解析为 Edge/Windows）、unlockscreen 三分支 curl 端到端补验通过、全量 go test 绿、进程无异常退出。
- 复验注意：退出日志写库失败此前的表象是"后端日志一行 GORM Error 1406、HTTP 响应仍 200"，页面无感知——同类"异步日志静默丢失"问题（登录日志/操作日志）排查时先查落库行，别只看接口响应。
