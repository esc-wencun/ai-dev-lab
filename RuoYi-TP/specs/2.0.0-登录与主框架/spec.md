# 2.0.0-登录与主框架 · spec

> **状态：✅ 已完成（2026-09-29，P1 里程碑达成：浏览器级 admin/admin123 登录进经典若依同款主界面）**
> 复刻经典若依登录链路 + 主框架页。**P1 里程碑：admin/admin123 登录后看到经典版同款主界面，菜单按权限渲染，标签页/iframe 交互正常。**
> 依赖：1.0.0（✅）+ 1.5.0（模板由子任务并行移植中，本模块负责接线）。
> 契约依据：tech-stack.md 第四节实锤 + SysLoginController / SysIndexController / SysCaptchaController / ShiroConfig（filter chain）/ login.js / index.js。

## 页面与端点清单

| 方法 | 路径 | 对位经典版 | 要点 |
|---|---|---|---|
| GET | /login | SysLoginController.login | 渲染 login/index.html；变量 isRemembered（config `tp.rememberMe`，默认 true 对位 yml）/ isAllowRegister（sys_config `sys.account.registerUser`）/ captchaEnabled / captchaType（config，对位 yml `shiro.user.captchaEnabled/captchaType`——经典版 sys_config 无 captcha 键）；ajax 访问返回 `{"code":"1",...}`（LoginAuth 已实现） |
| POST | /login | SysLoginController.ajaxLogin | 表单参数 `username`/`password`/`validateCode`/`rememberMe`；成功 code 0（JS 跳 /index）+ 写 sys_logininfor + 建会话；失败 code 500 文案对位经典版 messages.properties；验证码校验失败同 500；密码 5 次锁 10 分钟（`pwd_retry:` 键） |
| GET | /captcha/captchaImage?type=math&s=<rand> | SysCaptchaController | **匿名**；GD 出图 math 型（算式 + 答案存会话——见设计说明 2）；no-store 响应头 |
| GET | /logout | LogoutFilter | 销毁会话 + 删 cookie → 302 /login |
| GET | /index | SysIndexController.index | 主框架：menus（递归树）/ user / sideTheme / skinName / footer / tagsView / menuStyle（全部 sys_config `sys.index.*`）；menuStyle=topnav → index_topnav 模板；移动 UA → index |
| GET | /system/main | SysIndexController.main | iframe 内容页 main.html（version + 密码策略提醒逻辑 initPasswordIsModify/passwordIsExpiration） |
| GET | /unauth | error/unauth | 正式模板（替换 1.0.0 占位） |
| GET | /system/switchSkin | SysIndexController | 皮肤切换页 skin.html |
| GET | /system/menuStyle/{style} | SysIndexController | 写 nav-style cookie（无页面） |

> 锁屏（/lockscreen /unlockscreen）与注册（/register）不在本模块（8.0.0）。

## 关键设计说明

1. **验证码开关联动**：config 文件 `config/tp.php`（新增项目级配置）`captcha.enabled=true`、`captcha.type='math'`、`rememberMe=true`——对位经典版 yml；登录页渲染与 POST 校验同源。
2. **验证码答案存储**：经典版存 Shiro Session（登录前无会话）。TP 版登录前无自建会话，答案存 **Redis 临时键**：`captcha:<randomid>`（2 分钟过期），图片响应带 `captchaId` cookie（或直接复用即将成为会话 uuid 的预生成 id——采用方案：**登录页 GET 时预生成 session uuid 写 cookie + 建"匿名会话"**，验证码答案放该会话的 captcha 字段，POST /login 校验后转正为完整会话——语义最贴近经典版 Session attribute，且登录成功 cookie 不变）。char 型一并支持。
3. **登录成功会话内容**：userId / loginName / userName / deptId / permissions（PermissionService 计算）/ roles（role_id+roleKey+dataScope 数组）/ isAdmin。密码验证：PasswordService.verify(loginName, password, salt, hash)；**用户不存在与密码错误统一文案「用户不存在/密码错误」**（对位经典版 user.notExists 与 user.notPasswordError 的合并提示，动工时按 messages.properties 逐条抄）。
4. **密码错误锁定**：`pwd_retry:<loginName>` incrWithExpire(600s)，达 5 次锁——锁定期登录直接拒绝「密码输入错误5次，帐户锁定10分钟」。
5. **登录日志**：成功/失败写 sys_logininfor（info/login_name/ipaddr/login_time/status 0成功 1失败/msg；browser/os 字段用 UA 简析或留空——实施时看列定义决定，可空列留空）。
6. **菜单树**：service 读 sys_menu（admin 全量 `menu_type in ('M','C') and visible='0'`，普通用户经角色关联）→ getChildPerms 式递归组树 → 模板 volist 渲染（href=url 原值，target/isRefresh 原样）。
7. **rememberMe**：config 默认 true；勾选时 cookie 过期 30 天，会话 TTL 语义不变（deviations #17 降级版）。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录

（联调进行中——curl 级链路已全部验证，浏览器级联调待模板产出）

### 2026-09-29 curl 级验证结论（全过）

- 验证码：math 型出图 + 答案入匿名会话 + 用后即删；char 型分支实测正常（JPFQ 形态 4 位）
- 登录成功/失败/锁定三态文案与 code 全对位；sys_logininfor 三行实测（成功 0/验证码错误 1/密码错 1）；测试行已清理
- 锁定：连错 5 次「密码输入错误5次，帐户锁定10分钟」；锁定期正确密码拒绝；测试后 pwd_retry 键已删（admin 解锁复原）
- rememberMe=true 时 cookie Max-Age=2592000（30 天）实测正确
- menuStyle 302 / logout 会话销毁（Redis 键 1→0）/ 登出后 /index 302 /login 全过
- PHPUnit 25 tests 79 assertions 全绿（含 MenuService 树构建 5 场景）

### 踩坑记录（详见 tasks.md）

1. **TP varPathinfo 默认 's' 与经典版验证码 URL `?s=<rand>` 撞名**——TP 把 s 参数当 pathinfo，路由 302。app/Request.php 覆盖 $varPathinfo='r'。
2. **TP8 'Controller@action' 路由风格不拼 app\controller 前缀**——用 'controller/action' 斜杠风格 + controller_suffix=true。
3. **captcha 双 uuid**——make() 内部再调 anonUuid() 导致 cookie 与 Redis 键不一致，验证码必失败。写会话统一由控制器收口。
4. **/index 误入 LoginAuth 匿名表**（1.0.0 遗留）——主框架拿不到会话；已收紧，对齐经典版 user filter 保护范围。
5. **模板子任务两轮覆盖手工修复**——子任务重写模板文件把驼峰键名（menuName/isRefresh/menuId）带回来，且收尾时误删 index/main.html（误判为并行残留，实际是 TP 视图解析所需的正确位置）。最终修复后与子任务交付报告对齐：**后续业务页移植禁止再用「整文件重写」交付已联调文件**；键名规范定死为「DB 列下划线、自建会话键驼峰」。
6. **子任务带回的重要发现（后续业务页适用）**：think-template 的 `{if condition="check_perm('system:user:add')"}` 语法不可用（parseCondition 在冒号处截断字符串字面量），按钮权限须用输出形式 `{:check_perm('...') ? '1' : '0'}`（实测可用）；模板注释里不能出现 `{...}` 模板序列（会被当真标签执行）。
7. **include 片段机制**：子任务交付 43 个 include/*.html 插件片段（ztree/summernote/bootstrap-table 全套扩展等，对位经典版 include :: 片段名），3.0.0+ 业务页按需引用。
