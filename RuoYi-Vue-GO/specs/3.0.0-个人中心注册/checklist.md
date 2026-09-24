# Checklist · 03 个人中心注册

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。

## 验收清单

- [x] tasks.md 全部勾选
- [x] `go test ./...` 全绿；更新 specs/README.md 状态
  - 注：17 个包全部 ok（新增 fileutil 4 测试、注册链 3 测试）；gofmt/vet 干净
- [x] 端到端：curl 全链（profile 查询返回 user+roleGroup/postGroup、updateProfile 成功、png 上传成功/exe 拒绝、下载一致、穿越 0 字节拒绝、注册开关关闭拒绝→开启→注册落库→新用户登录成功）
- [ ] 端到端：个人中心四 tab 页面操作、注册页全链路（浏览器）
  - 注：待用户浏览器确认（同 02 的页面验收任务）；curl 层已全部覆盖

## 测试数据清理记录

- 2026-09-24：sys_user 删除 testreg01（注册测试）；sys_logininfor 清理 info_id>102；Redis 清空 login_tokens:/captcha_codes:/pwd_err_cnt:/sys_config:sys.account.registerUser（临时开关）；uploads 目录（测试上传文件）整目录删除；admin 会话未污染（token 均为测试后即弃）。
