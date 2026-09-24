# 03 个人中心 + 注册 + 通用文件上传下载

> **状态：✅ 已完成（2026-09-24；curl 端到端通过，浏览器页面级验收与 02 一并由用户确认）**
>
> 2026-09-24 重梳任务清单时补入的模块（原清单遗漏；Python 版 spec-02-profile 已覆盖，deviations.md #8 也预留了核对项）。
> Java 版对应（起点，动工时重读）：`SysProfileController`、`CommonController`、`SysRegisterService`（共 9 端点，API 清单范本见 Python 版 `RuoYi-Vue-FastApi/specs/spec-02-profile.md`）。
> 前端页面：`views/system/user/profile`（个人中心四 tab）、`views/register.vue`、上传组件 `components/ImageUpload` / `FileUpload`。
> 依赖：1.0.0-基础设施（认证中间件）；改密/头像与 2.0.0-登录闭环 的 getInfo（pwdChrtype 等）呼应。

## 端点范围（待契约先行动作写实）

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/user/profile` | SysProfileController.profile | 登录即可 |
| 2 | PUT | `/system/user/profile` | updateProfile | 登录即可 |
| 3 | PUT | `/system/user/profile/updatePwd` | updatePwd | 登录即可 |
| 4 | POST | `/system/user/profile/avatar` | avatar（参数名 `avatarfile`） | 登录即可 |
| 5 | POST | `/register` | SysRegisterService.register | 匿名（开关控制） |
| 6 | POST | `/common/upload` | CommonController.uploadFile | 登录即可 |
| 7 | POST | `/common/uploads` | uploadFiles（多文件） | 登录即可 |
| 8 | GET | `/common/download` | fileDownload（fileName + delete） | 登录即可 |
| 9 | GET | `/common/download/resource` | resourceDownload | 登录即可 |

- [ ] 契约先行：上表补全参数与返回 JSON 要点（含唯一性校验文案、注册开关、上传白名单），写实本 spec

## Python 版同名 spec 踩坑复用

- **multipart 上传与 body 预读冲突**（Python 版 log_decorator 预读 body 导致 Stream consumed；Go 版中间件/日志包装读 body 时同样注意 `c.Request.Body` 只能读一次，需 GetBody 重放）
- 通用上传：扩展名白名单（拒绝 exe）、文件名长度 ≤100、按 `{profile}/{yyyy/MM/dd}/` 落盘、返回 url/fileName/newFileName/originalFilename
- 下载：路径穿越防护（`..` 拒绝）、`delete=true` 下载后删、resource 前缀校验
- 改密：旧密码不匹配/新旧相同文案、chrtype 复杂度策略、成功后刷新 Redis 会话 + pwd_update_date
- 注册：开关 `sys.account.registerUser`、验证码复用登录逻辑、校验链顺序与文案对照 Java SysRegisterService
- profile 查询返回 roleGroup/postGroup 逗号分隔格式；修改仅四字段（nickName/email/phonenumber/sex）+ 手机号/邮箱唯一性校验 + 会话刷新

任务分解见 [tasks.md](tasks.md)，验收清单见 [checklist.md](checklist.md)。

## 实施记录

- 2026-09-24 完成。契约对照源码：SysProfileController（profile 四字段修改/roleGroup-postGroup 逗号分隔/updatePwd 三文案/avatar 参数名 avatarfile）、CommonController（upload 白名单/uploads 逗号拼接/download 时间戳重命名+delete/download-resource stripPrefix）、SysRegisterService（开关→验证码→校验链顺序）、FileUploadUtils（50MB/100 字符/日期目录/基名_seq.ext）、MimeTypeUtils 白名单逐项移植。
- 踩坑：① json.RawMessage 经 gin 序列化变 base64 字符串——user JSON 必须 RawMessage 类型内联输出（端到端发现修复）；② sqlite 无 now()——dao 层时间全部 Go 侧传参。
- profile 根目录 Go 侧默认 ./uploads（Java D:/ruoyi/uploadPath），URL 前缀 /profile 一致；部署路径差异对前端透明（前端只消费返回的 url/fileName）。
