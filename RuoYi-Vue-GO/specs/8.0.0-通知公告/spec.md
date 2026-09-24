# 08 通知公告

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：SysNoticeController + SysNoticeReadServiceImpl（本 Java 版定制：公告已读功能）
> 前端页面：views/system/notice + 首页公告组件（listTop/markRead）
> 依赖：1.0.0-基础设施；富文本内容 XSS 排除（/system/notice 在 xss.DefaultExcludes）

## API 清单（已对照 Java 源码核实，2026-09-24）

| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:notice:list | noticeTitle/createBy like + noticeType 相等 |
| 2 | GET | /listTop | 无（登录即可） | 前 5 条 + isRead 标记 + unreadCount（sys_notice_read 联查） |
| 3 | POST | /markRead | 无 | noticeId 参数；写 sys_notice_read（唯一键防重） |
| 4 | POST | /markReadAll | 无 | ids 逗号分隔批量 |
| 5 | GET | /readUsers/list | system:notice:list | 已读用户列表 |
| 6 | GET | /{noticeId} | system:notice:query | 详情 |
| 7 | POST | / | system:notice:add | noticeContent 富文本（XSS 排除路径） |
| 8 | PUT | / | system:notice:edit | 修改 |
| 9 | DELETE | /{noticeIds} | system:notice:remove | 物理删 |

## 实施记录

- 2026-09-24 完成。sys_notice_read 表（本 Java 版定制）按 ry_20260417.sql 结构接入（unique key uk_user_notice 防重）。
- 端到端：公告创建/listTop（isRead+unreadCount）/删除全部通过。
