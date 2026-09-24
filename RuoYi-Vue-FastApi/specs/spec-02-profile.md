# Spec-02 个人中心 + 注册 + 通用文件上传下载

> **状态：✅ 已完成（2026-09-24）**。Task 1-4 + 3.5 全部落地并端到端验证：
> profile查询(roleGroup/postGroup)、四字段修改(唯一性校验)、改密(旧密码/重复/chrtype策略)、
> 头像上传(Stream consumed已修：multipart跳过body预读)、注册全链路(开关/验证码/长度/唯一性，文案对齐java)、
> 通用上传(白名单拒绝exe)、下载(路径穿越防护)。测试数据已清理（admin昵称还原、测试用户已删、开关还原false）。
>
> Java 版对应：`SysProfileController`、`CommonController`、`SysRegisterService`
> 前端页面：`views/system/user/profile/index.vue`、`views/register.vue`、上传组件 `components/ImageUpload` / `FileUpload`

## 目标

完成个人信息维护、密码修改、头像上传，以及配套的通用文件上传下载（后续模块的富文本/附件都依赖它）。

## API 清单（对照 Java 版逐一确认后开发）

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/user/profile` | SysProfileController.profile | 登录即可 |
| 2 | PUT | `/system/user/profile` | updateProfile | 登录即可 |
| 3 | PUT | `/system/user/profile/updatePwd` | updatePwd | 登录即可 |
| 4 | POST | `/system/user/profile/avatar` | avatar（参数名 `avatarfile`） | 登录即可 |
| 9 | POST | `/register` | SysRegisterController.register | 匿名（开关控制） |
| 5 | POST | `/common/upload` | CommonController.uploadFile | 登录即可 |
| 6 | POST | `/common/uploads` | uploadFiles（多文件） | 登录即可 |
| 7 | GET | `/common/download` | fileDownload（fileName + delete） | 登录即可 |
| 8 | GET | `/common/download/resource` | resourceDownload | 登录即可 |

## Task 1: 通用文件上传下载（先做，头像依赖它）

- [x] `utils/upload_util.py`：
  - [x] 上传路径：`{UPLOAD_PATH}/{yyyy/MM/dd}/{编码文件名.ext}`（对应 Java FileUploadUtils，编码防中文/路径穿越）
  - [x] 文件扩展名白名单校验（`DEFAULT_ALLOWED_EXTENSION`：图片/office/压缩/视频/pdf），非法返回 `上传文件扩展名是不允许的扩展名`
  - [x] 文件名长度校验（>100 返回 `上传文件名是最长100个字符`）
  - [x] 返回 `url`（/profile 前缀）+ `fileName` + `newFileName` + `originalFilename`
- [x] `GET /common/download`：按文件名下载 upload_path 下文件，`delete=true` 时下载后删除；文件名非法（含 `..`）返回错误；本地日志记录下载动作（对应 Java downloadFile 日志）
- [x] `GET /common/download/resource`：下载本地资源（profile 前缀校验，防目录穿越）
- [x] 端到端测试：上传 png/docx 成功、上传 exe 被拒、下载一致、delete 生效

## Task 2: 个人信息查询与修改

- [x] `GET /profile` 返回：`data`（用户信息，含 dept）、`roleGroup`（角色名组，逗号分隔）、`postGroup`（岗位名组）——对照 Java selectUserRoleGroup/selectUserPostGroup 的格式
- [x] `PUT /profile`：仅允许改 `nickName / email / phonenumber / sex`（对应 Java 只 set 这四个字段）；手机号唯一性校验（`修改用户'%s'失败，手机号码已存在`）、邮箱唯一性校验；成功后刷新 Redis 会话中的用户信息（对应 tokenService.setLoginUser）
- [x] VO 注意：返回给前端不序列化 password
- [x] 端到端测试：改昵称后 getInfo 中生效

## Task 3: 修改密码

- [x] `PUT /profile/updatePwd`：body `{oldPassword, newPassword}`
- [x] 校验顺序与 Java 一致：旧密码不匹配 → `修改密码失败，旧密码错误`；新旧相同 → `新密码不能与旧密码相同`
- [x] 成功：更新 `password`（bcrypt）+ `pwd_update_date`，刷新会话；返回 `操作成功`
- [x] 密码复杂度策略：按 `sys.account.chrtype` 配置校验（0 任意 / 1 纯数字 / 2 纯字母 / 3 字母+数字 / 4 字母数字特殊字符，允许 `~!@#$%^&*()-=_+`），与 getInfo 返回的 pwdChrtype 呼应；失败提示对照 Java 前端提示文案
- [x] 端到端测试：改回原密码链路（测试后恢复 admin/admin123，避免污染共用库）

## Task 3.5: 用户注册（Phase 0 遗留 stub，此轮补齐）

- [x] `POST /register`：开关 `sys.account.registerUser` 非 true → `当前系统没有开启注册功能！`
- [x] 校验链完全对照 Java SysRegisterService.register 顺序与文案：
  - [x] 验证码（开关开启时，逻辑与登录一致）
  - [x] 用户名非空（`用户名不能为空`）、密码非空（`用户密码不能为空`）
  - [x] 用户名长度 2-20（`账户长度必须在2到20个字符之间`）
  - [x] 密码长度 5-20（`密码长度必须在5到20个字符之间`）
  - [x] 用户名唯一（`保存用户'%s'失败，注册账号已存在`）
- [x] 注册成功：nickName=userName、bcrypt 密码、pwd_update_date=now、登录日志记一条 Register 类型；返回 `注册成功`
- [x] 注册失败的 HTTP/code 行为与登录一致（HTTP 200 + code 500）
- [x] 端到端测试：开关关闭被拒 → 开启后注册 → 新用户可登录 → 清理测试用户

## Task 4: 头像上传

- [x] `POST /profile/avatar`：multipart 字段名 `avatarfile`；校验文件非空（`上传图片异常，请联系管理员`）
- [x] 保存后更新 `avatar` 字段并刷新会话；返回 `{imgUrl: url}`（与 Java AjaxResult.put("imgUrl", url) 一致）
- [x] 端到端测试：上传后 profile 接口 avatar 有值，前端 ImageUpload 可显示

## 验收清单

- [x] RuoYi-Vue3 个人中心页面：基本信息修改、改密、头像上传全部可用
- [x] pytest 全绿；更新 specs/README.md 状态
