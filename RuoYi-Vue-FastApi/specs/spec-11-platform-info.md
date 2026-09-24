# Spec-11 平台标识：`GET /getPlatformInfo` + 前端监控页降级提示

>
> **状态：进行中（2026-09-25）**
> **契约主文档**：本模块为跨端增量（Java/Go/Python/前端四端），契约以 Go 版任务书为准 → [`../../RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md`](../../../RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md)。此处仅记 Python 侧实现与验收。
> Java 版对应：新建 `SysPlatformController`（基准新增，非复刻既有端点）

## API 清单

| # | 方法 | 路径 | 鉴权 | 权限串 |
|---|------|------|------|--------|
| 1 | GET | `/getPlatformInfo` | 登录即可（`get_current_user` 校验，AuthException → 401 信封） | 无 |

返回（ResponseUtil.success，驼峰）：

```json
{
  "code": 200, "msg": "操作成功",
  "framework": "RuoYi-Vue-FastApi",
  "version": "1.0.0",
  "language": "python",
  "languageVersion": "3.10.x",
  "features": { "druidMonitor": false, "serverMonitor": true }
}
```

- `druidMonitor: false`——Druid 是 Java 连接池，Python 版无对应物，前端数据监控页据此降级提示
- `serverMonitor: true`——服务监控已实现（spec-08 psutil 方案），前端正常使用

## Task 1: 端点实现

- [ ] `login_controller.py` 新增 `GET /getPlatformInfo`：鉴权模式照抄 unlockscreen（`get_current_user` + AuthException 分支），返回 AppConfig.app_name/app_version + platform.python_version + features 常量
- [ ] 路由随 loginController 在 server.py controller_list 自动注册，确认无需改 server.py

## Task 2: 验证

- [ ] 端到端：冷启动 → 登录拿 token → `GET /getPlatformInfo`（Bearer 头）返回上述结构；无 token 返回 401 信封（附日期）
- [ ] Python 侧不涉数据库写入，无测试数据清理项（登录会话测完清理）

## 实施记录

（实现过程中的坑与修正记录在此）
