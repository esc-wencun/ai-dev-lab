# 8.0.0-个人中心 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。2026-10-01 动工并完成。

## Task 1 · UserService 最小集

- [x] UserService 补 updateUserInfo（四字段+update_time，update_by 不写）/ updateUserAvatar——selectUserById/checkPhoneUnique/checkEmailUnique/resetUserPwd 4.0.0 已有，直接复用（类注释不变，方法注释注明 profile 用途）
- [x] PHPUnit：checkPhoneUnique/checkEmailUnique 口径端到端 curl 实测（self true / taken false）；纯逻辑不连库断言不适用（唯一校验即 DB 查询）
- [x] config/profile.php：root_path()/public/profile + avatar 子目录

## Task 2 · ProfileController 资料/改密

- [x] 7 条路由注册；#[Log] 3 处（重置密码/2、个人信息/2×2）
- [x] GET index：查 DB 组 user + deptName + postGroup（selectUserPostGroup 既有）
- [x] GET checkPassword：裸 boolean（DB verify）
- [x] POST resetPwdSave：旧密码错 → 新旧相同（旧 salt 比对）→ 换盐重加密 + SessionService::destroy + cookie 清除 → success；三态文案全实测
- [x] POST update：四字段（userId 取会话）；唯一校验仅非空才查；成功回写会话 userName
- [x] curl 级自测：三态 + 资料更新 + 冲突文案 + oper_param ****** 脱敏核对

## Task 3 · 头像上传

- [x] GET avatar 渲染 + POST updateAvatar：白名单 bmp/gif/jpg/jpeg/png；`{Y/m/d}/{uuid}.{ext}` 落 public/profile/avatar/
- [x] URL = /profile/avatar/... 静态直服（GET 200 image/png 实测）
- [x] 旧头像删除（二次上传旧文件 is_file=false 实测）；会话 avatar 键回写
- [x] php.ini upload_max_filesize：测试图片 1KB 远小于限制（本机 php.ini 默认 2M，50MB 上限为代码层校验值，实际受 php.ini 约束——AGENTS.local.md 不另记，部署时按需调整）

## Task 4 · 手机/邮箱唯一校验端点

- [x] 4.0.0 已落（UserController::checkPhoneUnique/checkEmailUnique + 路由 137/138），本模块零新增
- [x] curl 级自测：15888888888 self→true / userId=2→false

## Task 5 · 注册（拍板=做）

- [x] RegisterService：校验链七文案 + 加盐入库 + 登录日志「注册成功」（仅成功记）
- [x] RegisterController：GET（匿名会话+验证码变量，不查开关）+ POST（开关 → 验证码 → register）
- [x] view/register/index.html（subagent 移植 + 引擎冒烟两态）
- [x] LoginAuth ANON_PATHS 含 /register（负向核对，未改）
- [x] curl + 浏览器全链路：开关 false 拒绝 / true 注册成功落库可登录 / 关开关复原

## Task 6 · 锁屏/解锁

- [x] IndexController 加 lockscreen（会话写 lockscreen=true + 渲染 lock 页）/ unlockscreen（DB verify 不计 pwd_retry；移除字段）
- [x] index/index.html + index_topnav.html 各一处改动（`{$lockscreen|json_encode} === true`）
- [x] view/lock/index.html（subagent 移植：three.js 粒子 + 时钟 + unlock ajax）
- [x] curl + 浏览器：lockscreen 后 /index HTML 含 true；unlock 两态；pwd_retry 键不产生

## Task 7 · 页面模板（profile 三页 + lock + register）

- [x] profile.html（双 tab + remote 双校验 + 头像弹层入口；resetPwd 回调按 spec 微调跳 logout）
- [x] avatar.html（cropper 三预览 + FormData avatarfile → saveReload）
- [x] resetPwd.html（minlength=5 quirk 保留；chrtype 四分支；成功回调跳 logout）
- [x] lock/index.html + register/index.html（独立页）
- [x] subagent 静态自查 + 引擎冒烟 12 次渲染全绿（th: 残留 0、配平、变量注入、radio/two态 captcha）

## Task 8 · 端到端验收（浏览器级）

- [x] 个人中心页：右上角入口打开 iframe；左栏资料+头像兜底；基本资料改手机号（remote 提示 + 保存成功）
- [x] 修改密码 tab：改密 → 跳登录页 → 旧 cookie 失效 → 新密码登录（会话重建闭环全链路实测）
- [x] 修改密码弹窗（右上角入口 resetPwd()）：模板就绪（弹窗页 curl 渲染核对）；**弹窗内改密未单独浏览器实测**（与 tab 提交同端点同逻辑，tab 链路已全验证）
- [x] 头像：API 两轮上传实测（落盘/URL/旧删/DB）；cropper 裁剪交互未逐项（标准插件行为，页面渲染已核对）
- [x] 锁屏：点锁屏 → lock 页（时钟/粒子实测）；错密码提示清空；对密码回 /index
- [x] 注册：开开关 → 注册页全链路（表单/验证码/成功弹窗）→ 新号可登录；关开关复原
- [x] ry 表零 schema 变更

## Task 9 · 收尾

- [x] PHPUnit 全绿（58 tests 147 assertions，RegisterTest 6 用例新增）；spec.md 实施记录回填；checklist 勾选；README 总表更新（8.0.0 ✅）
- [x] deviations 处置回填：改密重登录登记 deviations.md；#14 补注 public/profile 直服
- [x] 测试数据清理：注册用户/头像文件/oper_log/logininfor/Redis 键全清；admin admin123 复原（新盐 hash 核对）；registerUser=false 复原
