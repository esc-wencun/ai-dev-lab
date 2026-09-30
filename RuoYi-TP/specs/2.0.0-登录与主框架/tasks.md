# 2.0.0-登录与主框架 · 任务分解

> 做完即勾；没做不勾并注明原因。（2026-09-29 动工）

## Task 1 · 项目级配置 config/tp.php

- [x] captcha.enabled(true) / captcha.type('math') / rememberMe(true) / version('4.8.3')

## Task 2 · 验证码服务（GD 自绘 math/char 双型）

- [x] `app/service/CaptchaService.php`：make() math 算式 + GD 出图（jpeg 130x40 干扰线）；char 型一并实现
- [x] 匿名会话机制：cookie uuid 复用会话键空间；验证码答案挂 captcha 字段
- [x] verify 用后即删（E2E 实测同码二次提交拒绝）
- [x] **踩坑①（双 uuid）**：make() 内部原又调 anonUuid() 无 cookie 时再生成新 uuid，导致 Set-Cookie 的 uuid 与 Redis 写入的 uuid 不一致——改为 make() 纯生成、写会话由控制器统一 write($uuid)。
- [x] **踩坑②（TP varPathinfo='s' 撞名）**：TP 兼容模式路径参数名默认就是 `s`，经典版验证码 URL 的防缓存参数 `?s=<rand>` 会被 TP 解析成 pathinfo（返回 "0.123"）致路由 302——app/Request.php 覆盖 $varPathinfo='r' 解决，保留经典版 URL 形态（login.js 零改动）。

## Task 3 · 登录服务（LoginService）

- [x] `app/service/LoginService.php` 全链路：验证码 → 用户查询（del_flag）→ 锁定检查 → PasswordService.verify → 失败计数/成功清零 → 会话转正（permissions/roles 装载）→ sys_logininfor
- [x] 登录日志列名按实际 DDL（login_name/ipaddr/login_location/browser/os/status/msg/login_time；UA 简析 browser/os）
- [x] 错误文案逐条对位 messages.properties（验证码错误/用户不存在/密码错误/密码输入错误5次，帐户锁定10分钟）
- [x] E2E：错验证码 500「验证码错误」；错密码 500「用户不存在/密码错误」；连错 5 次锁定文案（pwd_retry=5）；锁定期正确密码同样拒绝；成功 code 0 + Set-Cookie + 会话含 permissions/roles/isAdmin

## Task 4 · 菜单服务（MenuService）

- [x] menusOf：admin 全量 / 普通用户角色关联 → buildTree 递归组树（内部稳定排序不依赖输入顺序）
- [x] PHPUnit 5 场景（嵌套/排序/孤儿容错/空输入/深嵌套）全绿

## Task 5 · 控制器与路由

- [x] `app/controller/LoginController.php`（index/doLogin/captchaImage/logout）+ `app/controller/IndexController.php`（index/main/switchSkin/menuStyle/unauth）
- [x] ConfigService（sys_config 读 + `config:` 缓存 + refresh 清缓存）
- [x] route/app.php 斜杠风格路由——**踩坑③：TP8 'Controller@action' 风格不拼 app\controller 前缀（报类不存在），须用 'controller/action' 斜杠风格 + controller_suffix=true**；config/app.php app_namespace 补 'app'
- [x] **修正**：/index、/system/main 从 LoginAuth 匿名表移除（对位经典版 user filter 保护），未登录 302 /login 实测通过；MenuService user_id 判空

## Task 6 · 模板接线与联调（2026-09-29 完成）

- [x] 正式模板产出（子任务）：login/index.html（82 行，表单字段/验证码/rememberme 与 login.js 一致）、index/index.html（542 行）、index/index_topnav.html、index/main.html（1890 行，由 app/view/main.html 移入 index/ 对齐控制器视图解析）、error/unauth|404|500、system/skin
- [x] **模板返工两轮**：①残留 th:* 与 `@{...}` URL 语法（index 101 处）→ 批量 sed 转换；②键名驼峰/下划线混用（menuName/isRefresh vs menu_name/is_refresh——DB 列下划线）→ 统一下划线；会话键（loginName/userName 驼峰）保留不动。**教训：子任务重写文件会覆盖手工修复，返工后须全量 grep 复核**
- [x] **模板返工发现的控制器缺口**：$isMobile 未 assign → body 误挂 canvas-menu 致侧栏 display:none；补 IndexController assign。/system/notice/listTop 404（主框架 ajax 依赖）→ NoticeController 最小实现（空列表形态，7.0.0 补全）
- [x] **里程碑浏览器级联调（preview 实测）**：登录页 1:1 渲染（蒲公英背景/logo/验证码/记住我）→ admin/admin123 登录 → 主框架（侧栏菜单：首页/系统管理/系统监控/系统工具/若依官网；用户名「若依」显示；首页 iframe + 标签页）→ 点「用户管理」→ 新标签页 + iframe 加载 /system/user（404 属预期，4.0.0 实现）→ 关标签页 → 登出 → 302 /login 全通
- [x] curl 级端到端：成功/失败/锁定/登出全过（见 Task 3）

## Task 7 · 收尾（2026-09-29 完成）

- [x] 全部勾选 + spec.md 实施记录 + checklist 核对 + README 总表标 ✅
- [x] 测试数据清理：sys_logininfor 清空、Redis session 键清空（admin 密码未动、pwd_retry 已删）
