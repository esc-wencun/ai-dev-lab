# 8.0.0-个人中心 · spec

> **状态：✅ 已完成（2026-10-01 联调验收通过）**
> 对位经典若依 SysProfileController（资料/改密/头像，8 方法实核）+ SysIndexController 锁屏端点（GET /lockscreen + POST /unlockscreen）+ SysRegisterController + SysRegisterService（注册评估，本文给出「做」的定稿建议）+ templates 根 lock.html / register.html 与 templates/system/user/profile 三页（页面共 5 个，逐方法核对实锤）。
> 依赖：1.0.0（LoginAuth 全局登录态 / #[Log]+OperLog）、2.0.0（SessionService 会话设施 / 匿名会话验证码模式 / PasswordService md5+salt）；页面渲染于 1.5.0 主框架内、运行在 2.0.0 登录态下。**无 #[Perm] 端点**——本模块三个经典控制器均无 @RequiresPermissions（grep 实锤），全部仅登录态，register 匿名。
> 调研依据：reference 源码逐文件核对（SysProfileController / SysIndexController.lockscreen·unlockscreen / SysRegisterController / SysRegisterService / SysPasswordService.matches·encryptPassword / ShiroUtils.randomSalt / FileUploadUtils.upload·uuidFilename / MimeTypeUtils.IMAGE_EXTENSION / ResourcesConfig /profile 静态映射 / CaptchaValidateFilter / ehcache-shiro.xml / 模板 profile·avatar·resetPwd·lock·register / ruoyi/register.js / index.js #lockScreen 绑定）+ **ry-tp 库实测**（sys_user 21 列含 avatar varchar(100)；sys_config 11 行：`sys.account.registerUser=false`、`sys.account.chrtype=0`、`sys.account.initPasswordModify=1`、`sys.account.passwordValidateDays=0`；admin/ry 两用户 avatar 均空）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、信封结构、敏感字段排除（password/oldPassword/newPassword/confirmPassword——OperLog 已实现）。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/system/ProfileController.php | SysProfileController | **7 条路由**（8 方法中 `GET /edit` 为死端点不复刻——返回模板 `system/user/profile/edit` 在经典 templates 中不存在，grep 实锤，页面上也无人调用） |
| app/service/UserService.php（**最小集**） | ISysUserService 的 6 个方法 | selectUserById / updateUserInfo / resetUserPwd / updateUserAvatar / checkPhoneUnique / checkEmailUnique（+ registerUser 归 Task 注册）。**4.0.0 用户管理落地时收编扩展**（同 3.0.0 DictService 最小只读先例） |
| IndexController 扩展 lockscreen/unlockscreen 两方法 | SysIndexController 同名方法 | 锁屏会话保持语义见特殊行为 4 |
| app/controller/RegisterController.php + RegisterService | SysRegisterController + SysRegisterService | **建议做**（定稿建议见特殊行为 5） |
| app/view/system/user/profile/{profile,avatar,resetPwd}.html、app/view/lock/index.html、app/view/register/index.html | templates/system/user/profile/×3、templates/lock.html、templates/register.html | 5 页；cropper 静态资源 include 片段已在 1.5.0 移植（app/view/system/include/cropper-*.html），零新增静态资源 |
| config/profile.php（新增项目配置） | RuoYiConfig.profile / getAvatarPath | 头像上传根目录（deviations #14 落地细化：指向 `public/profile`，URL `/profile/**` 由 web server 直接服务，**无需新增读取端点**） |

**范围外**（后续模块，勿在本模块实现）：用户管理列表/增删改/导入导出（4.0.0）；sys_user_online 表读写（9.0.0 在线用户，且 TP 版数据源为 Redis 不落该表）；`sys.account.initPasswordModify/passwordValidateDays` 弹窗提醒已在 2.0.0 main 页实现，本模块只保证改密后 pwd_update_date 正确落库使其闭环。

## 页面清单（5）

| # | 页面 | TP 模板路径 | 对位经典版 | 组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 个人中心页 | view/system/user/profile/profile.html | profile/profile.html | 双 tab（基本资料/修改密码）+ 头像入口 | 左栏：头像 img（空→/img/profile.jpg，onerror 兜底）+ 登录名称/手机号码/所属部门（deptName / postGroup，postGroup 空→「无岗位」）/邮箱（abbreviate 16 截断）/创建时间（yyyy-MM-dd）；头像点击 → `top.layer.open` 打开 system/user/profile/avatar（btn 确定/关闭，yes 回调 iframe 内 submitHandler）；右栏基本资料 tab：userName 必填、email 必填+email 格式+**remote POST ctx+"system/user/checkEmailUnique"**（data: userId+email，文案「Email已经存在」）、phonenumber 必填+isPhone+**remote POST checkPhoneUnique**（文案「手机号码已经存在」）→ `$.operate.saveModal(ctx+"system/user/profile/update", serialize)`；修改密码 tab：oldPassword 必填+**remote GET ctx+"system/user/profile/checkPassword"**（文案「原密码错误」）、newPassword 必填 **minlength=6** maxlength=20 specialSign、confirmPassword equalTo → submitChangPassword（chrtype=`sys.account.chrtype` 传 checkpwd 前端规则校验）→ `$.operate.saveModal(ctx+"system/user/profile/resetPwd", serialize)` |
| 2 | 修改头像弹窗 | view/system/user/profile/avatar.html | profile/avatar.html | **Cropper**（aspectRatio 1 / viewMode 1 / autoCropArea 0.9 / 三档圆形预览） | 上传图像 input[type=file] accept=image/* → FileReader → cropper.replace；缩放/旋转/翻转/重置按钮 data-method；submitHandler：`cropper.getCroppedCanvas().toBlob` → FormData **avatarfile** → `$.ajax POST ctx+"system/user/profile/updateAvatar"`（processData/contentType false）→ `$.operate.saveReload(result)` |
| 3 | 修改密码弹窗 | view/system/user/profile/resetPwd.html | profile/resetPwd.html | 表单（主框架右上角下拉「修改密码」`resetPwd()` 打开，770×380） | userId 隐藏域 + loginName readonly；oldPassword remote checkPassword（同上）；newPassword 必填 **minlength=5** maxlength=20（**与页面 1 的 6 不同，经典版原样 quirk，保留勿统一**）；confirmPassword equalTo；chrtype help-block 文案四分支；提交 `$.operate.save(ctx+"system/user/profile/resetPwd", serialize)` |
| 4 | 锁屏页 | view/lock/index.html | lock.html | 时钟 + **three.min.js 粒子背景**（静态资源已移植） | 显示 loginName / userName（空→'-'）/头像；input[name=password] 回车或箭头按钮 → `$.ajax POST ctx+"unlockscreen"` data {password}（beforeSend 保留 X-CSRF-Token 头发送——csrf 关闭无影响，模板保留 meta 空值）；成功 code 0 → `location.href = ctx+'index'`；失败 msg 提示 + 清空输入框；「退出重新登录」链接 → /logout |
| 5 | 注册页 | view/register/index.html | register.html | 登录页同款布局 | username(maxlength 20)/password/confirmPassword(maxlength 20) + 验证码行（captchaEnabled 才渲染，imgcode 点击换图 `captcha/captchaImage?type=captchaType&s=Math.random()`）+ acceptTerm 勾选；validate：username required minlength=2、password required minlength=5 specialSign、confirmPassword equalTo；提交 `$.ajax POST ctx+"register"` data **{loginName, password, validateCode}**（confirmPassword **不提交**）；成功 layer.alert「恭喜你，您的账号 xxx 注册成功！」→ 跳 /login；失败 msg + 刷新验证码 |

> 页面 3 入口在主框架 index.html 右上角下拉（`resetPwd()` 已移植），页面 1 入口在右上角头像下拉「个人中心」（`/system/user/profile` 链接已移植）。**动工时需改 app/view/index/index.html 一处**：`var lockscreen = false;` 硬编码改为读会话 `lockscreen` 字段（对位经典版 `[[${session.lockscreen}]]`）——本模块唯一主框架模板改动点。

## 端点级 API 清单（profile 7 + 前置 check 2 + 锁屏 2 + 注册 2 = 13 条路由）

### 个人中心 /system/user/profile（对位 SysProfileController，无 @RequiresPermissions 实锤）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/user/profile | 无（仅登录态） | 无 | — | 渲染页面 1；用户数据**查 DB**（selectUserById，含 phonenumber/email/sex/avatar/createTime——经典版用会话快照 getSysUser()，TP 会话无 phone/email/sex，改查 DB 为实现简化，行为差异仅「他人代改资料时显示实时值」，无害）；模板变量 user / deptName（按会话 deptId 查 sys_dept）/ postGroup（sys_user_post join sys_post 按 "," join，空→''，模板显「无岗位」）；roleGroup 为经典版死变量（传而模板未用）不传 |
| 2 | GET | /system/user/profile/checkPassword | 无 | 无 | password（GET 参数） | **裸 boolean**：查 DB 取 hash/salt → PasswordService::verify(loginName, password, salt, hash)；无防爆破限制（经典版原样，安全注记见特殊行为 7） |
| 3 | GET | /system/user/profile/resetPwd | 无 | 无 | — | 渲染页面 3，变量 user = selectUserById（DB 最新） |
| 4 | POST | /system/user/profile/resetPwd | 无 | 重置密码, 2修改 | oldPassword、newPassword、confirmPassword（后端只消费前两个；敏感字段已被 OperLog 排除） | 校验顺序：oldPassword 不匹配 → error(500)「修改密码失败，旧密码错误」；newPassword 与旧密码相同（用**旧 salt** 比对）→ error(500)「新密码不能与旧密码相同」；成功：**salt=PasswordService::randomSalt()（6 位 hex）重加密** md5(loginName+newPassword+newSalt) → update sys_user SET password, salt, pwd_update_date=now(), update_time=now()（**update_by 不写**，mapper 实锤）→ **销毁当前会话并强制重新登录**（任务指明方案，与经典版差异见特殊行为 1）→ success()「操作成功」；update 0 行 → error「修改密码异常，请联系管理员」 |
| 5 | POST | /system/user/profile/update | 无 | 个人信息, 2修改 | userName、email、phonenumber、sex（表单序列化多出的空 id 参数忽略——经典版同） | 仅四字段可改（userId 取**会话**，防横向越权）；phonenumber 非空才查 checkPhoneUnique（口径：sys_user where phonenumber=? and del_flag='0' limit 1，userId 不同即不唯一）→ error(500)「修改用户'{loginName}'失败，手机号码已存在」；email 非空才查 checkEmailUnique → 「修改用户'{loginName}'失败，邮箱账号已存在」；update sys_user SET user_name/email/phonenumber/sex/update_time=now()（update_by 不写）→ 成功写回会话 userName → success()「操作成功」；0 行 → error()「操作失败」 |
| 6 | GET | /system/user/profile/avatar | 无 | 无 | — | 渲染页面 2，变量 user = selectUserById |
| 7 | POST | /system/user/profile/updateAvatar | 无 | 个人信息, 2修改 | multipart **avatarfile**（cropper 裁剪输出 blob，png） | 空文件 → error()「操作失败」；扩展名白名单 bmp/gif/jpg/jpeg/png（IMAGE_EXTENSION 实锤）+ 大小 ≤50MB；uuid 命名 `{Y/m/d}/{uuid}.{ext}` 落 `public/profile/avatar/`（getAvatarPath=profile+"/avatar" 对位）→ update sys_user SET avatar='/profile/avatar/…' → **删旧头像文件**（avatar 非空时按旧 URL 映射删，stripPrefix 去掉 /profile 前缀对位）→ 写回会话 avatar → success()「操作成功」；异常 → error(e.getMessage()) |

### 手机号/邮箱唯一校验（对位 SysUserController 两端点，本模块提前落最小版，4.0.0 收编）

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 8 | POST | /system/user/checkPhoneUnique | 无（仅登录态） | 无 | phonenumber（+编辑时 userId） | **裸 boolean**（jquery validate remote 直接消费）；口径 = phonenumber 全局唯一（del_flag='0'）limit 1；userId 相同视为自身放行 |
| 9 | POST | /system/user/checkEmailUnique | 无 | 无 | email（+userId） | 裸 boolean；email 全局唯一（del_flag='0'）；同上 |

### 锁屏（对位 SysIndexController.lockscreen/unlockscreen）

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 10 | GET | /lockscreen | 无（仅登录态） | 无 | — | 渲染页面 4；**会话写入 lockscreen=true 字段**（SessionService::write 更新 JSON，会话不销毁）；主框架 index 模板读该字段 JS 跳转（见页面清单注） |
| 11 | POST | /unlockscreen | 无 | 无 | password | 会话 user 空 → error(500)「服务器超时，请重新登录」（TP 版 LoginAuth 先拦，此分支实际不可达，保留防御）；密码匹配（查 DB verify，**不计入 pwd_retry 失败计数**——经典版 matches 直比不走 validate，实锤）→ 会话移除 lockscreen 字段 → success()「操作成功」；不匹配 → error(500)「密码不正确，请重新输入。」 |

### 注册（对位 SysRegisterController + SysRegisterService；GET/POST /register 均匿名，LoginAuth ANON_PATHS 已含 /register）

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 12 | GET | /register | 无（匿名） | 无 | — | 渲染页面 5；变量 captchaEnabled/captchaType（config('tp.captcha.*)，对位 CaptchaValidateFilter 塞 request attribute）；**GET 不检查注册开关**（经典版原样：开关只拦 POST，页面始终可渲染）；匿名会话模式与登录页相同（预生成/复用 anonUuid + cookie，验证码答案进匿名会话） |
| 13 | POST | /register | 无（匿名） | 无 | loginName、password、validateCode（confirmPassword 前端校验不提交） | 开关 `sys.account.registerUser` ≠ 'true'（ConfigService::get，DB 实测默认 false）→ error(500)「当前系统没有开启注册功能！」（**开关检查在验证码之前**——对位经典版 controller 先查开关、filter 后校验码的顺序）；验证码校验（enabled 时，答案用后即删，失败文案「验证码错误」）；loginName 空 → 「用户名不能为空」；password 空 → 「用户密码不能为空」；密码长度 ∉[5,20] → 「密码长度必须在5到20个字符之间」；用户名长度 ∉[2,20] → 「账户长度必须在2到20个字符之间」；checkLoginNameUnique（login_name 全局唯一 del_flag='0' limit 1）→ 「保存用户'{loginName}'失败，注册账号已存在」；成功：salt=randomSalt + md5 重加密 → insert sys_user（login_name/user_name=loginName/salt/password/pwd_update_date=now，dept_id NULL、sex/status/user_type 走列默认）→ 失败「注册失败,请联系系统管理人员」；成功记 sys_logininfor（status '0' 成功、msg「注册成功」——对位 Constants.REGISTER 计入成功组，**仅成功记日志，失败不记**，经典版实锤）→ success() |

统计：**13 条路由**；**#[Perm] 0 处**（#10/#11/#1~#9 仅登录态，#12/#13 匿名）；**#[Log] 3 处**（#4 重置密码/2、#5 个人信息/2、#7 个人信息/2）。

## 特殊行为清单（改密会话重建 / 头像路径 / 锁屏语义 / 注册拍板 / quirk）

1. **改密后强制重登录（会话重建）——任务指明要求，与经典版存在行为差异，明示如下**：经典版实况 = resetPwd 成功后 `setSysUser(selectUserById(...))` **刷新会话用户数据、登录态保持、不踢出**（SysProfileController 源码实锤）。本 spec 按任务要求定稿为：**销毁当前 Redis 会话 + 删除 cookie → 返回 success**，前端（页面 1/3 的提交回调）由 TP 模板定制：alert「密码修改成功，请重新登录」→ `location.href = ctx + 'logout'`（经典版回调是纯 saveModal alertSuccess，此处 TP 页面 JS 允许微调——页面本就是自家移植物）。动工时若用户拍板改回经典版行为（刷新会话不踢出），则撤销本条并回填实施记录。**登记 deviations（新条目，见文末）**。
2. **「新密码不能与旧密码相同」口径**：用**旧 salt + 旧 hash** 验证 newPassword（即 newPassword==oldPassword 明文等价），先于换盐重加密执行；oldPassword 错误优先返回。
3. **头像链路闭环**：上传目录 = `public/profile/avatar/`（config/profile.php 返回绝对路径，deviations #14 落地）→ URL `/profile/avatar/Y/m/d/{uuid}.{ext}` 由 web server 对 public/ 的静态服务直接命中，**零新增读取端点**（对位经典版 ResourcesConfig 的 /profile/** 资源映射）；旧头像删除 = 旧 URL 去 `/profile` 前缀映射回磁盘删文件（对位 stripPrefix）；会话 avatar 键同步回写（主框架右上角头像立即生效）；上传校验扩展名白名单 + 50MB（对位 FileUploadUtils.assertAllowed；php.ini upload_max_filesize 需 ≥50M，动工时核对）。
4. **锁屏会话保持语义（软锁）**：GET /lockscreen 仅在会话 JSON 写 `lockscreen=true`，**不销毁会话**，30 分钟空闲超时照常生效；主链（LoginAuth）**不拦锁屏态请求**——锁屏只是 /index 渲染时的 JS 跳转（对位经典版 session attribute + index.html 内联判断，无过滤器强制）；解锁成功移除字段；锁屏期间会话过期 → unlockscreen 得 code "1" JSON，lock 页显示「未登录或登录超时。请重新登录」，用户走「退出重新登录」或重新登录，行为自洽；解锁密码错误**不计失败次数**（经典版 matches 直比实锤）；锁屏前的活动页签由 index.js 的 storage('lockPath') 在解锁回来后恢复（静态 JS 已移植，零改动）。
5. **注册拍板建议：做**（本 spec 已按「做」写实端点清单）。理由：① 复刻成本小——1 页面 + 2 端点，验证码/密码加密/登录日志/唯一校验全是既有设施复用（LoginService/CaptchaService/PasswordService 模式照抄）；② 开关 `sys.account.registerUser` 默认 false（DB 实测 config_id=4），入口隐藏 + POST 拒绝，上线风险为零，且给 6.0.0 参数管理提供了一个「改开关立即生效」的天然验证场景（ConfigService::refresh 清 config: 键）；③ Python/Go 版均有注册能力，功能对齐。**若拍板不做**：删除 Task 5，登录页 `{if $isAllowRegister}` 入口本就因 false 不渲染，无残留影响；tech-stack 第四节第 14 条回填「不复刻」结论。
6. **前端双保险文案**（remote 提前提示 + 后端兜底，文案不同属经典版原样）：remote「Email已经存在」「手机号码已经存在」「原密码错误」；后端「修改用户'{loginName}'失败，手机号码已存在」「…邮箱账号已存在」「修改密码失败，旧密码错误」「新密码不能与旧密码相同」。
7. **quirk（原样保留，勿"修复"）**：① 页面 1 修改密码 tab newPassword minlength=6，页面 3 弹窗 minlength=5（两处不一致为经典版原样）；② GET /register 无开关检查，关闭注册时注册页仍可渲染（只拦 POST）；③ checkPassword 端点无防爆破限制（可无限次试探旧密码，经典版原样——安全加固留待后期统一评估，不在本模块擅自加）；④ GET /system/user/profile/edit 死端点（模板缺失）不复刻；⑤ roleGroup 为死变量不传；⑥ register.js 不提交 confirmPassword，仅前端校验。
8. **防重复提交**：经典版三个控制器零处 @RepeatSubmit（grep 实锤）→ 本模块全部路由不挂 RepeatSubmit 中间件。
9. **chrtype 密码策略**：DB 实测 `sys.account.chrtype=0`（不限制）；页面 1/3 的 help-block 四分支文案与 checkpwd 前端校验照模板移植，chrtype 从 ConfigService 取（6.0.0 改参后生效）。

## 关键设计说明

1. **会话结构向后兼容扩展**：本模块新增会话键 `lockscreen`（bool）；改写会话一律走 `SessionService::write($uuid, $data)`（读改写全量，勿局部 set）。update/updateAvatar 成功后用**当前会话数组**改键回写（对位经典版 currentUser.set… + setSysUser，不重查 DB）。
2. **ProfileController 挂中间件**：仅 LoginAuth（全局）+ OperLog（路由级，#4/#5/#7 三处生效）；不挂 CheckPerm（无注解）/RepeatSubmit。
3. **UserService 最小集签名**：`selectUserById(int $userId): ?array`（全列）/ `updateUserInfo(array $data, int $userId): int`（四字段+update_time）/ `resetUserPwd(string $hash, string $salt, int $userId): int`（password+salt+pwd_update_date+update_time）/ `updateUserAvatar(string $avatar, int $userId): int` / `checkPhoneUnique(string $phone, int $selfUserId): bool` / `checkEmailUnique(string $email, int $selfUserId): bool`（selfUserId<=0 视为新增场景，4.0.0 收编）。类注释注明「4.0.0 收编扩展」。
4. **密码验证走 DB 不走会话**：经典版 checkPassword/resetPwd 用会话里的 SysUser（Shiro Principal 存全量含 hash）；TP 会话刻意不存 hash/salt（2.0.0 设计），两处均查 DB——行为等价（改密后旧会话hash失效问题因会话重建而不存在）。
5. **注册匿名会话复用登录页机制**：GET /register 的 anonUuid 生成/复用、验证码出入会话、cookie 落地，照抄 LoginController::index / captchaImage 现有模式（captcha 双 uuid 坑已修复的收口方式：会话写统一由控制器做）。
6. **模板 layer 归属**：ProfileController 在 `app\controller\system` layer（视图根 app/view/system/，模板路径写 'user/profile/profile' 相对路径）；RegisterController 在根 layer（'register/index'）；lock 页由 IndexController 渲染（'lock/index'）。include 片段按 3.0.0 踩坑先例需在对应 layer 内有副本（avatar 页需 system/include/cropper-*.html——已存在）。
7. **lock/register 页是无布局独立页**（对位经典版直接引 bootstrap+jquery+ry-ui，不套 include :: header 的 iframe 页框架），模板自引 css/js，`var ctx = '/';` 照抄内联。
8. **时间字段**：profile 页创建时间显示 yyyy-MM-dd（模板格式化）；pwd_update_date 落库 Y-m-d H:i:s（sys_user 列 datetime，2.0.0 main 页过期判断消费）。

## 拟登记 deviations

| # | 差异点 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|---|
| 1 | 改密后登录态 | setSysUser 刷新会话，**不踢出、不强制重登录** | **销毁会话强制重新登录**（会话重建），前端提示「密码修改成功，请重新登录」后跳 /logout | 任务指明的安全加固要求（2026-09-29 写实任务明示）；与经典版不一致故登记；动工时若用户拍板改回经典版行为则撤销本条 |
| 2 | profile 页用户数据源 | 会话快照（getSysUser） | 查 DB（TP 会话不含 phone/email/sex） | 实现简化，行为差异为「显示实时值」，无害加固不单独登记（本表留痕备查） |

另：头像路径 deviations **#14 已登记**（本模块落地为 public/profile 直服方案，动工时在 #14 行补注实现形态）；锁屏/注册无行为级差异需登记（锁屏不计失败次数、GET /register 无开关检查均为经典版原样照抄）。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-10-01）

- ProfileController（7 路由：双 tab 页/checkPassword 裸 bool/resetPwd 三态+会话销毁/update 四字段+会话回写/avatar/updateAvatar 旧删新传）+ IndexController 锁屏两方法 + RegisterController/RegisterService（开关→验证码→校验链→加盐入库→成功日志）。
- UserService 补 updateUserInfo/updateUserAvatar 两方法（其余四方法 4.0.0 已有直接复用）；config/profile.php 落 public/profile 直服。
- 5 页模板 subagent 移植 + 引擎冒烟 12 次渲染；主框架两模板各改一处 lockscreen 读会话。
- 浏览器全链路：改密会话重建闭环（改密→跳登录→旧 cookie 失效→新密码登录）、锁屏（粒子/时钟/两态解锁/不计 pwd_retry）、注册（开关两态+成功弹窗+可登录）、资料 remote 校验与保存、头像 API 两轮（旧删新传/URL 直服）。
- 修复 1 处：RegisterService 缺 `use PasswordService;`（全局类 use 纪律第 7 次踩坑——已在纪律记录）。
- PHPUnit 58 tests 147 assertions 全绿（RegisterTest 6 用例新增）；admin 密码/资料/开关/日志/头像文件全复原。
