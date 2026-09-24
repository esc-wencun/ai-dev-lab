# AGENTS.local.md（本机环境配置）

本文件记录**当前开发机器（Windows）的本机环境**：命令实际使用的解释器路径、数据库 / Redis 的当前指向等因机器而异的内容。文件名虽带 `.local`（惯例上常被 gitignore），本文件**刻意提交入库**，作为本机配置范例供参考；换机器部署时按实际环境改写本文件即可。[AGENTS.md](AGENTS.md) 保持通用写法，在本机执行其中命令时以本文件的值为准。

[CLAUDE.md](CLAUDE.md) 已通过 `@` 导入本文件；其他 AI 编码工具请把本文件与 AGENTS.md 一并加入上下文。

## Python 解释器（RuoYi-Vue-FastApi）

- 必须用全路径 `"C:\Program Files\Python310\python.exe"`（3.10）。
- 本机 PATH 默认的 Python 是 Anaconda 3.8，版本过旧不能用于本项目（FastAPI + SQLAlchemy 2.0 async 需要 3.10+）。

## Java 运行时（RuoYi-Vue）

- 运行 `ruoyi-admin.jar` 必须显式指定 JDK 17：
  `"C:\Program Files\Java\jdk-17.0.12\bin\java.exe" -jar ruoyi-admin/target/ruoyi-admin.jar`
- 本机 PATH 上的 java.exe 是 JDK 1.6，直接 `java -jar` 会失败。

## MySQL / Redis 当前指向（2026-09-25 起）

三版后端共用的数据库与缓存指向 `docker/` 提供的本地 Docker 环境（详见 [docker/README.md](docker/README.md)），各版 `.env.dev` 与 Java 版 `application.yml` / `application-druid.yml` 已同步为此配置：

- MySQL：`127.0.0.1:3306`，库 `ry-vue`，账号 `wencun` / `111111`
- Redis：`127.0.0.1:6379`，db `0`，密码 `111111`

> 历史备注：早期开发用的是阿里云 RDS（库 `ry-vue-26-09-24`）+ Redis db11，2026-09-25 前后迁移到本地 Docker。环境再变动时，以 Java 版配置文件为准并同步各版 `.env.dev`，同时更新本节。
