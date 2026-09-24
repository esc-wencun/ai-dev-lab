# Tasks · 08 通知公告

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：对照 SysNoticeController（9 端点含本版定制 listTop/markRead/readUsers）写实 API 清单
- [x] Task 公告 CRUD 5 端点（list/getInfo/add/edit/remove；富文本 XSS 排除已在 1.0.0 xss.DefaultExcludes 覆盖）
- [x] Task 已读功能 4 端点（listTop 带 isRead+unreadCount/markRead/markReadAll/readUsers；sys_notice_read 表接入）
- [x] 端到端：公告创建、listTop 返回 isRead/unreadCount、删除（2026-09-24 通过；测试公告已清除）
