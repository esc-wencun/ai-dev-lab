# 2.0.0-登录与主框架 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因。（2026-09-29 浏览器级联调通过）

## 登录链路（curl 级）

- [x] GET /captcha/captchaImage 出图（jpeg 头 + no-store 响应头）且答案已入匿名会话
- [x] POST /login admin/admin123 + 正确验证码 → `{"code":0,"msg":"登录成功"}` + Set-Cookie tp_session
- [x] 错误密码 → code 500「用户不存在/密码错误」；sys_logininfor 新增失败行
- [x] 验证码错误 → code 500「验证码错误」；用后即删（同码二次登录拒绝）
- [x] 密码连错 5 次 → 「密码输入错误5次，帐户锁定10分钟」；10 分钟内正确密码也拒绝；测试后 pwd_retry 键已删（admin 解锁）
- [x] 成功登录 sys_logininfor 新增成功行；会话内容含 permissions/roles/isAdmin（admin 权限集 80+ 项全装载）
- [x] GET /logout → 会话删除（Redis 键 1→0 实测）+ 302 /login
- [x] rememberMe=true → cookie Max-Age=2592000（30 天）；false → 会话 cookie

## 主框架（浏览器级，里程碑 ✅）

- [x] 登录页 1:1 渲染（蒲公英背景、若依 logo、技术栈列表、验证码行、记住我复选框——preview 截图确认）
- [x] 浏览器表单登录 → /index 渲染经典版同款布局（顶部导航 + 侧栏 + iframe 容器 + 标签页）
- [x] 菜单按 ry-tp 库数据渲染（首页/系统管理/系统监控/系统工具/若依官网——admin 全量，href 为 DB url 原值）
- [x] 首页 iframe 加载 /system/main（200，Hello Guest + 官网介绍 + 版本 v4.8.3）
- [x] 标签页开合正常（点用户管理 → 新标签 + iframe /system/user；关标签恢复——业务页 404 属预期，4.0.0 实现）
- [x] 主框架「服务器错误」弹窗修复（/system/notice/listTop 404 → NoticeController 最小实现）
- [x] 侧栏显示修复（isMobile 未 assign 致 canvas-menu 误挂）
- [x] 登出 → 302 /login

## 横切接线

- [x] 登录日志走 sys_logininfor（不走 sys_oper_log，对位经典版——登录链路不挂 #[Log]）
- [x] check_perm 模板函数在登录页（未登录）返回 false 不抛异常（login 页正常渲染验证）
- [x] 验证码开关联动：config/tp.php captcha.type 切 char 实测出 4 位字母数字形态，后还原 math

## 纪律自查

- [x] 业务代码零处直接引用 predis（grep=0）
- [x] 测试数据清理记录：sys_logininfor 11 行测试数据清空；Redis session/captcha/pwd_retry 测试键全删（KEYS 仅剩 config:* 业务缓存）；admin 密码未改动

## 测试数据清理记录

- 2026-09-29：sys_logininfor 测试行（登录成功/失败/锁定 11 行）全部删除；Redis db1 session:* 全删；pwd_retry:admin 已删（账号解锁）。
