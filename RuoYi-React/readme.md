# RuoYi-React

若依管理系统的 **React + Ant Design** 前端实现：与 `RuoYi-Vue3`（Vue 3 + Element Plus）功能等价、界面不强求一致，服务三个后端（Java / Python / Go，同监听 8080 互斥运行）。

> 规范入口 [AGENTS.md](AGENTS.md)；任务台账与模块三件套在 [specs/README.md](specs/README.md)；与基准前端的有意差异登记在 [specs/deviations.md](specs/deviations.md)。

## ⚠️ 测试状态声明

本项目**尚未进行系统性的详细测试**：核心链路（登录 / 系统管理 / 系统监控 / 布局与页签 / 导入导出 / 定时任务等）已按 spec checklist 逐项做过浏览器实操验收（记录见 specs 各模块），纯逻辑单测 127 项全绿；但**未做**完整回归测试、跨浏览器兼容测试、三版后端全矩阵验证（Python/Go 降级场景）与性能/安全测试。用于生产前请自行充分验证，问题反馈欢迎提 issue。

## 技术栈

React 19 · TypeScript（strict） · Ant Design 5 · Redux Toolkit · Vite · Vitest

## 启动

```bash
npm install
npm run dev          # http://localhost:8090，/dev-api 代理到 localhost:8080
npm run build:prod   # 生产构建
npm test             # 单元测试
```

默认账号 admin / admin123（需先启动任一版后端，三版互斥）。

## 与 RuoYi-Vue3 的关系

- RuoYi-Vue3 是**基准前端**：行为契约以其源码为准，51 页面 / 20 接口文件的调查基线固化在 [specs/reference/](specs/reference/)。
- 有意差异（dev 端口 8090、antd 视觉、KeepAlive 不实现、tool/build 砍除、tool/gen 暂缓等）见 [specs/deviations.md](specs/deviations.md)。

## 当前进度（2026-09-29）

- 51 个基准页面对应功能全部实现或按拍板排除（tool/build 砍除、tool/gen 暂缓占位）。
- 模块 0.0.0 ~ 7.0.0 完成（登录闭环 / 系统管理八页 / 系统监控八页 / 布局三模式 / TagsView 页签 / 设置抽屉 / 暗色主题 / 锁屏 / Crontab 七域生成器）；8.0.0 除暂缓的 gen 外完成。
- 核心链路已浏览器实操验收：CRUD 全链路、树表排序、导入导出（含 updateSupport 双分支）、角色菜单半选提交、字典联动、job 调度全生命周期、锁屏硬劫持与解锁、页签持久化。
