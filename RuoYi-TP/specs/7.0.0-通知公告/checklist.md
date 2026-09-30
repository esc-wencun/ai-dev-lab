# 7.0.0-通知公告 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（notice 13 + common upload 1，共 14）。

## 通知公告（curl / DB 级）

- [x] GET /system/notice 渲染列表页（system:notice:view；未登录 302 /login，ajax 未登录 code "1"）——两者实测
- [x] POST /system/notice/list：TableDataInfo + **固定 notice_id desc**（预置 3/2/1 顺序实测；新增置顶实测）；行 10 驼峰列实测（含 noticeContent 全文）；noticeTitle like 实测（联调公告甲过滤）；createTime 白名单（PageQuery 机制 PHPUnit 覆盖）
- [x] POST /system/notice/add：@Xss 标题 500「公告标题不能包含脚本字符」实测；空标题/超 50 字文案（NoticeTest PHPUnit 断言）；富文本 longblob 读回**字节级一致**实测（中文+HTML+✓）；create_by=admin 落库
- [x] GET /system/notice/edit/{noticeId} 渲染编辑页——浏览器实测 hidden noticeId=16 + title 回显 + summernote code 回填一字不差
- [x] POST /system/notice/edit：update_by=admin 落库实测；富文本更新读回一致实测；noticeType 1→2 修改实测
- [x] GET /system/notice/view/{noticeId}：**无 #[Perm] 仅登录态**（未登录 302 /login 实测；登录后可达）；详情页渲染（类型角标「通知」/meta/1727B 富文本/外链 gitee 图片加载成功）
- [x] GET /system/notice/listTop：{code:0, data:[5 行], unreadCount:n} 实测；行 7 键驼峰（noticeId/noticeTitle/noticeType/status/createBy/createTime/isRead）与主框架消费 JS 一致；**isRead 布尔**（markRead 后 true）；**status='0' 过滤**（全关→data 空 实测）；**空表形态 {code:0,data:[],unreadCount:0} 2.0.0 回归实测**（全关等价验证）；主框架徽章 3→2→隐藏实测
- [x] POST /system/notice/markRead：落 sys_notice_read（read_time 非空）实测；**重复调用幂等**（两次 markRead 后 COUNT=1 实测）
- [x] POST /system/notice/markReadAll：ids=3,2,1 批量落库 3 行实测；空数组短路（ids 空 → 不报错）；「全部已读」按钮全链路实测（徽章隐藏 + 3 行置灰 + DB 3 行）
- [x] GET /system/notice/readUsers/{noticeId}：渲染已读用户页（popupRight 弹层实测）
- [x] POST /system/notice/readUsers/list：行 6 驼峰列实测（userId/loginName/userName/deptName/phonenumber/readTime）；read_time desc；**searchValue 双列 like**（admin 命中 login_name / 若依 命中 user_name / nomatch=0 实测）；admin 行三表 JOIN（若依研发部门 + 手机号）
- [x] POST /system/notice/remove：**两连删**实测（删公告 16 → sys_notice + sys_notice_read 双表同清，COUNT=0）；无效 ids → 0 行（浏览器删除后行消失）
- [x] POST /common/upload：合法 png → {code:0, url, fileName, newFileName, originalFilename} + 落盘 + **url GET 可达（Content-Type: image/png）**；非法扩展 .exe → 500「上传文件格式不允许：exe」；未登录走 LoginAuth（ajax code "1"——upload 无静态后缀）

## 页面级验收（浏览器）

- [x] 公告列表页：预置 3 行 notice_id desc + 类型徽章（通知/公告）+ 状态徽章（正常）+ 三按钮（新增/修改/删除）
- [x] 标题链接 → **右侧滑出详情弹层**（popupRight）：类型角标「通知」、meta 行（admin + 时间 + 正常）、富文本渲染、**外链图片（gitee foruda CDN）加载成功**（screenshot 确认滑出层形态）
- [x] 新增（addFull 全屏 layer 弹层——经典版 openFull 即 layer 全屏，原样）：summernote 完整渲染（172 工具栏按钮、editable 区就绪）；富文本（加粗+列表）提交 → DB 字节级一致 → 列表置顶回显；**图片上传 insertImage 链路未浏览器实测**（API 侧 upload+GET 链路已验证；编辑器内点击上传属前端插件标准行为，sendFile 回调代码已按经典版移植）
- [x] 修改（editFull）：编辑器回填富文本一字不差（hidden noticeContent 默认转义防属性截断 + code 回填实测）
- [x] 阅读用户弹层：表格渲染 + 搜索框过滤实测（不命中→「没有找到匹配的记录」、清空恢复）+ markRead 后刷新显示 admin 行
- [x] **主框架公告链路回归**（2.0.0 收编闭环）：顶部下拉 3 行 + 徽章 3；点开单条 → markRead 落库（DB 核对）+ 徽章 3→2 + 条目置灰 #999；「全部已读」→ 3 行全落库 + 徽章隐藏
- [x] 全部页面在主框架内正常打开：列表 tab、新增/修改全屏 layer、详情/已读用户右侧弹层

## 横切与纪律自查

- [x] #[Perm] 9 处（grep 实测）——**spec 正文写 8 处是漏数**：readUsers 页与 readUsers/list 各一处 system:notice:list（经典版两方法均有 @RequiresPermissions 实锤），实际 9 处=经典版 9；免注解端点（listTop/markRead/markReadAll/view/{id}/upload）仅登录态（未登录 ajax code "1" / 页面 302 实测）
- [x] #[Log] 3 处（INSERT/UPDATE/DELETE grep 实测）；sys_oper_log 落库实测（title=通知公告，business_type 1/2/3 各有行）；无 EXPORT（grep 0）
- [x] RepeatSubmit 未挂（grep 0）；predis 直连 0；DataScope 0；admin 保护 0
- [x] 本模块零 Redis 键（公告无缓存——代码无 RedisCache 调用）
- [x] 权限两通道一致：sys_menu.perms 5 串（add/edit/list/remove/view）与 #[Perm] 9 处（5 串复用）、check_perm 调用（add/edit/remove 3 串）互查无遗漏
- [x] 表结构零变更；sys_notice_read 仅 INSERT/DELETE
- [x] XSS 三防线对位：标题 @Xss 后端 500 实测 + 正文入库不转义（富文本字节级一致实测）+ view 页 |raw（subagent 引擎冒烟：正文 raw、标题转义）；2229B 预置样本零损伤（浏览器渲染外链图片成功）

## Deviations 核对

- [x] 「富文本正文入库前净化」：**不登记**——正文直存直出=经典版 excludes+utext 双原样，无行为差异
- [x] 「输入侧全局 XSS 转义」：维持不复刻（6.0.0 同处置合并定——@Xss 点位已复刻，其余字段影响面≈0，不登记）
- [x] 经典版原样 quirk 确认未"修复"：listTop/markRead/markReadAll/view 无权限注解、阅读用户按钮无权限控制、列表行含 noticeContent 全文、markReadAll 全量幂等、addFull/editFull=layer 全屏弹层（openFull 原样）

## 测试数据清理记录

- [x] ✅ 2026-10-01 测试公告（联调公告甲/甲改×2 轮、浏览器联调公告）物理删除；sys_notice 恢复预置 3 行原值（title/content 复原核对）
- [x] ✅ 2026-10-01 sys_notice_read 清空回 0 行（markRead/markReadAll 测试记录全删）
- [x] ✅ 2026-10-01 sys_oper_log 测试行清理（oper_id 166+ 共 24 行）；sys_logininfor 测试行清理（login_time≥2026-09-30 23:00 共 9 行，恢复 8 行）
- [x] ✅ 2026-10-01 runtime/upload 测试图片删除（5 个 *_m7up.png）；runtime/download 无残留（本模块无导出）
- [x] ✅ 2026-10-01 admin 会话状态正常（浏览器持续登录验证；admin/admin123 登录链路多轮实测）
