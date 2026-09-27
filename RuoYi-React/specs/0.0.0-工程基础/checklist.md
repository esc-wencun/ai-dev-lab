# Checklist · 0 工程基础

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] `npm run dev` 在 8090 起服务，浏览器显示占位页（2026-09-27，预览截图确认）
- [x] 起任一后端后，`/dev-api/captchaImage` 经代理 200（2026-09-27，fetch 验证 status 200，前缀剥离正确）
- [x] `npm run build:prod` 构建成功产出 dist/（2026-09-27，tsc -b + vite build 11.7s）
- [x] `npm test` 通过（2026-09-27，Vitest v5 冒烟 1/1）
- [x] svg 精灵注册成功：控制台无报错，`<use href="#icon-user">` / `#icon-dashboard` 截图可见（2026-09-27）
- [x] 根 AGENTS.md / readme.md / RuoYi-Vue3 AGENTS.md 同步无漂移；RuoYi-React/AGENTS.md 存在且指向正确（2026-09-27）

## 测试数据清理记录

- （本模块无后端数据写入；如启动后端验证代理，登录会话键测完即删）
