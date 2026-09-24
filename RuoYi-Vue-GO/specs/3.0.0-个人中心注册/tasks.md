# Tasks · 03 个人中心注册

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。契约先行完成后写实本文件。

- [x] 执行 specs/README.md《动工检查单》，写实本文件夹三件套（API 清单参数/返回补全）
  - 注：已对照 Java SysProfileController/CommonController/SysRegisterController/SysRegisterService + FileUploadUtils/FileUtils/MimeTypeUtils 源码核实（2026-09-24），参数/返回/文案/白名单写入本 spec 与实现注释
- [x] Task 拆分（对齐 Python 版 spec-02-profile）：通用上传下载先行 → 个人信息查询/修改 → 修改密码 → 用户注册
  - 注：落地为 pkg/utils/fileutil（白名单/穿越防护/落盘规则）+ dao/profile_dao + service/profile_service + handler/profile_handler 四层
- [x] Task 通用上传下载（4 端点：upload/uploads/download/download-resource）
  - 注：白名单对位 DEFAULT_ALLOWED_EXTENSION（exe 拒绝已验证）；文件名 {基名}_{seq}.{ext}+日期目录；download 穿越拒绝（.. 与 ResolveLocal 双重防护）；delete=true 下载后删已实现
- [x] Task 个人信息查询/修改
  - 注：profile 返回 data=user JSON + roleGroup/postGroup（逗号分隔）；updateProfile 仅四字段 + 手机/邮箱唯一性（文案逐字对位）+ 会话刷新
- [x] Task 修改密码
  - 注：updatePwd 三分支（旧密码错误/新旧相同/成功）；成功后 pwd_update_date 重置 + 会话 user JSON 刷新（setLoginUser 对位）
- [x] Task 用户注册
  - 注：开关 sys.account.registerUser → 验证码（复用登录链存储）→ 校验链（顺序与文案逐字对位 Java）→ insert（nick_name=username、bcrypt、pwd_update_date）；注册成功写 sys_logininfor（REGISTER）
- [x] 单元测试：fileutil（白名单/穿越/落盘 4 个）+ 注册校验链（开关/顺序/文案/唯一性 3 个）
- [x] 端到端：curl 全链验证（profile 查询/修改、png 上传成功/exe 拒绝、resource 下载一致、穿越拒绝、注册开关关闭拒绝→开启→注册→新用户登录成功）
  - 注：**前端页面级操作（个人中心四 tab、注册页）待用户浏览器确认**——本环境浏览器预览不可用，同 02 处理

# 发现并修复的坑

- json.RawMessage（[]byte）经 gin 序列化会被 base64 编码——profile 的 data 必须用 json.RawMessage 类型显式内联（端到端发现，已修复）。
- sqlite 无 now() 函数：dao 层时间一律 Go 侧传参（types.DateTime），SQL 不写数据库方言函数（MySQL 专用 sysdate()/now() 在端到端真实库不受影响，但单元测试 sqlite 会挂）。