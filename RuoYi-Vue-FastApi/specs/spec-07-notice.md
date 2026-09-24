# Spec-07 通知公告

>
> **状态：✅ 已完成（2026-09-24）**。公告CRUD(longblob内容utf-8)、listTop已读标记、markRead/markReadAll幂等、readUsers分页、删除联动清理已读。端到端验证通过。
> Java 版对应：`SysNoticeController`（注意当前 Java 版含已读功能 sys_notice_read 表，比经典 RuoYi 多）
> 前端页面：`views/system/notice/`、首页通知栏组件
> 依赖：spec-01

## 目标

公告 CRUD + 已读状态跟踪 + 首页置顶列表。

## API 清单

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/notice/list` | list（分页） | `system:notice:list` |
| 2 | GET | `/system/notice/{noticeId}` | getInfo | 登录即可 |
| 3 | POST | `/system/notice` | add | `system:notice:add` |
| 4 | PUT | `/system/notice` | edit | `system:notice:edit` |
| 5 | DELETE | `/system/notice/{noticeIds}` | remove（批量） | `system:notice:remove` |
| 6 | GET | `/system/notice/listTop` | listTop（首页置顶） | 登录即可 |
| 7 | POST | `/system/notice/markRead` | markRead（单条已读） | 登录即可 |
| 8 | POST | `/system/notice/markReadAll` | markReadAll（批量已读） | 登录即可 |
| 9 | GET | `/system/notice/readUsers/list` | readUsersList（已读人员分页） | `system:notice:list` |

## Task 1: 公告 CRUD

- [x] DO：`SysNotice`（`sys_notice` 表）+ `SysNoticeRead`（`sys_notice_read` 表，开发前对照 Java SQL 核对字段）
- [x] 分页列表（`noticeTitle / createBy / 公告类型 noticeType`）；返回含已读状态相关字段（对照 Java list 的返回结构确认）
- [x] 新增/修改/批量删除（校验权限与逻辑同其他模块）
- [x] 端到端测试：CRUD 全链路

## Task 2: 首页与已读

- [x] `listTop`：首页置顶公告（取前 N 条，排序/过滤条件对照 Java 实现——含当前用户已读状态）
- [x] `markRead`：当前用户对单条公告记录已读（写 sys_notice_read，幂等）
- [x] `markReadAll`：批量已读（body 参数 ids 对照 Java：String ids）
- [x] `readUsersList`：某公告的已读用户分页（含搜索 searchValue）
- [x] 端到端测试：用户 A 已读后 listTop/readUsersList 状态正确

## 验收清单

- [x] RuoYi-Vue3 公告页面 + 首页通知组件可用（含"未读/已读"展示）
- [x] pytest 全绿；更新 specs/README.md 状态
