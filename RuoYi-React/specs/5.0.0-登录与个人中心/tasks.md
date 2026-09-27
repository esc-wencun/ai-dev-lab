# Tasks · 5 登录与个人中心

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [ ] 登录页（验证码全行为/记住我 RSA 三 cookie/预填 admin/admin123/redirect 跳转/失败刷验证码/注册入口不显示）
- [ ] 注册页（规则'0' 校验/验证码/成功弹窗跳转/失败刷新）
- [ ] 401 / 404 错误页 + redirect 中转页
- [ ] 首页（本工程版本卡/真实技术栈列表/里程碑折叠/去掉捐赠卡）
- [ ] 个人中心布局（左卡片 + 右 Tabs + activeTab 直达）
- [ ] 基本资料表单（校验 + updateUserProfile）
- [ ] 修改密码表单（infoPwdValidator + updateUserPwd）
- [ ] 头像裁剪弹窗（react-cropper 200×200/缩放旋转/预览/toBlob multipart 上传/回显写 store）
- [ ] 验证：全链路登录（验证码/记住我回显/redirect）→ 首页 → 个人中心三件事
- [ ] 验证：初始密码/密码过期弹窗跳转 resetPwd tab（造一个过期密码用户，测完清理）
- [ ] 验证：记住我 30 天 Cookie 为 RSA 密文（DevTools 确认无明文）
- [ ] 验证：头像裁剪上传成功且导航栏头像即时更新
