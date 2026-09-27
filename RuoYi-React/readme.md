# RuoYi-React

若依管理系统的 **React + Ant Design** 前端实现：与 `RuoYi-Vue3`（Vue 3 + Element Plus）功能等价、界面不强求一致，服务三个后端（Java / Python / Go，同监听 8080 互斥运行）。

> 规范入口 [AGENTS.md](AGENTS.md)；任务台账与模块三件套在 [specs/README.md](specs/README.md)；与基准前端的有意差异登记在 [specs/deviations.md](specs/deviations.md)。

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
- 有意差异（dev 端口 8090、antd 视觉、KeepAlive 暂缓、tool/build 砍除等）见 [specs/deviations.md](specs/deviations.md)。
