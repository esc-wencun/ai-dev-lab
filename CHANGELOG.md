# 更新日志（CHANGELOG）

本文件只记录**里程碑级功能更新**：新功能、既有能力的重大改造（如框架替换、架构级重构）。文档调整、叙事修改、配置说明、规范补齐等非功能性操作不记录。逐 commit 细节以 [git log](https://github.com/esc-wencun/ai-dev-lab/commits/main) 为准；模块级的过程与验收记录见各子项目 specs 目录。维护约定：每次完成一个功能里程碑，在此追加一条。

格式参照 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，日期倒序。

## 2026-09-28

- **RuoYi-React 工程落地**：前端工程整体入库，51 个基准页面对应功能全部实现或按拍板排除（build 砍除、gen 暂缓占位）；模块 00/01/05/07 完成、02/03/04/06/08 核心完成，深度验收与 KeepAlive 页签缓存等遗留项待做。
- **Go 版内嵌 tzdata**：打包产物在无 GOROOT 环境下时区（`loc=Asia/Shanghai` DSN）不再失效。

## 2026-09-27

- **RuoYi-React 前端立项**：React 19 + TypeScript + Ant Design 5 + Redux Toolkit，功能等价复刻 RuoYi-Vue3、服务三版后端；specs 0.0.0~8.0.0 任务书建立（进行中）。
- **四端一键打包**：package.bat 升级为 Java 服务端 + Vue3 客户端 + Python 源码包 + Go 三平台二进制的一键打包，按日期时分出版本号。
- **Java 版 MyBatis-Plus 收尾**：代码生成器模板适配 MP（Mapper/Service 增加 IPage 分页方法），DataScope 数据权限栈配对修复。

## 2026-09-26

- **Java 版引入 MyBatis-Plus**（[PR #3](https://github.com/esc-wencun/ai-dev-lab/pull/3)）：移除 PageHelper 切换 MP 分页，数据权限改为 MP 原生适配（specs 0.0.0/1.0.0）。
- **一键打包脚本**：新增 `/package` 命令与 package.bat（Java 服务端 + Vue3 客户端，按日期版本输出）。

## 2026-09-25

- **工作区开源发布**：初始提交——RuoYi 多语言服务端工作区（Java/Go/Python 三版复刻 + 共用前端）。
- **平台标识模块**（[PR #1](https://github.com/esc-wencun/ai-dev-lab/pull/1)、[PR #2](https://github.com/esc-wencun/ai-dev-lab/pull/2)）：新增 `GET /getPlatformInfo` 返回语言/版本/features 能力开关（含 `swaggerDocs`），前端数据监控、服务监控、系统接口页对非 Java 后端降级提示。
