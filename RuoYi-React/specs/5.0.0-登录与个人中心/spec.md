# Spec 5.0.0 登录与个人中心

>
> **状态：✅ 完成（2026-09-29 收口）**。登录页/register/首页/锁屏/错误页/redirect 中转随 2.0.0 提前落地并端到端验证。个人中心完整版（AvatarCropper 裁剪上传 multipart avatarfile / 资料表单校验 / 改密 infoPwdValidator + 改密后强制重登）已实现。2026-09-28~29 补齐实操：头像上传 API 链路（multipart avatarfile → code 200 + imgUrl + 落盘 + DB，测试数据已复原）、记住我三 Cookie（password 为 RSA 密文）、**初始密码/过期密码提醒弹窗（发现实现缺口——store 存标志但无 UI 消费，补齐于 getInfo thunk 内与基准同位，弹窗文案与 resetPwd 跳转实操逐字验证）**。遗留仅：改资料/改密表单提交实操（避免动 admin 密码）、头像裁剪 UI 交互（依赖原生文件对话框）。
> **背景**：打通「登录 → getInfo → getRouters → 菜单」全链路的页面层；含记住我 RSA、验证码、头像裁剪上传等基准细节。
> **契约侦察**：[../reference/01-pages-and-api-baseline.md](../reference/01-pages-and-api-baseline.md) §3~§5（登录/注册/首页/个人中心基线）。
> **依赖**：02（守卫/store）、03 批次 B。

## 范围

- 登录页 / 注册页 / 401 / 404 / redirect 中转页。
- 首页（静态介绍页等价，文案归属本工程——deviations #10）。
- 个人中心（资料/改密/头像裁剪）。
- utils/passwordRule 已在 1.0.0 完成，本模块消费。

## 关键行为契约

1. **登录页**：进页 getCodeImg（`captchaEnabled===undefined` 视为开启；base64 图点击刷新；uuid 入表单；失败自动刷新）；**登录明文 POST**（jsencrypt 仅用于记住我）；记住我三 cookie（username/password 密文/rememberMe，expires 30 天，回显 decrypt）；**注册入口保持不显示**（基准 register ref 恒 false，是行为不是遗漏）；成功 → unlockScreen → redirect 参数跳转。
2. **注册页**：username 2-20 / password 规则'0'（6-20 + 非法字符 `<> "'\|`）/ 确认一致 / 验证码；成功弹「注册成功」→ /login；失败刷验证码。
3. **初始密码/过期密码**：getInfo 弹窗文案「您的密码还是初始密码，请修改密码！」「您的密码已过期，请尽快修改密码！」→ 跳 `/user/profile/resetPwd`。
4. **个人中心**：左卡片（头像 + 部门 postGroup/角色 roleGroup/创建日期）+ 右 Tabs（基本资料 / 修改密码），`:activeTab?` 直达；资料校验（昵称 30/手机 11 正则/邮箱 50）；改密 infoPwdValidator（pwdChrtype 0-4）。
5. **头像裁剪**：react-cropper 固定 200×200（aspectRatio 1:1）+ 缩放/左右旋 + 实时预览；`getCroppedCanvas().toBlob` → FormData(`avatarfile`) → POST avatar（multipart，deviations #12）；成功回显 `VITE_APP_BASE_API + imgUrl` 写 store。
6. redirect 中转页：`navigate('/' + path, {query})` 等价 replace 转发。

## 设计决策

1. 首页静态内容重写为本工程叙事（React/TS/antd/RTK 真实栈 + 本项目里程碑折叠面板），捐赠卡去除——deviations #10 已登记。
2. 裁剪框「固定 200×200」语义：cropperjs 无 fixedBox，用 viewMode + aspectRatio + 容器尺寸约束等价实现（裁剪结果恒 200×200 输出即可）。
3. 错误页 gif 素材不复用（视觉差异允许），保留语义（401 无权限/404 不存在 + 返回按钮）。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
