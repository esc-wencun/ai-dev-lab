# 7.0.0-通知公告 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：service 层（PHPUnit 先行）→ 控制器+路由（含 2.0.0 收编）→ /common/upload → 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。2026-09-30~10-01 动工并完成。

## Task 1 · NoticeService + NoticeReadService

- [x] `app/service/NoticeService.php`：五方法 + noticeInput 校验收敛（@Xss 正则对位）；selectNoticeList 返回 Query（orderBy 合并单 order 子句，注释注明）；固定 notice_id desc
- [x] `app/service/NoticeReadService.php`：markRead（INSERT IGNORE）/ markReadBatch（空数组短路）/ deleteByNoticeIds / selectNoticeListWithReadStatus（LEFT JOIN + isRead 布尔，原生 SQL limit 参数化）/ selectReadUsersByNoticeId（三表 JOIN + 双列 like）/ selectUnreadCount（保留无消费方）
- [x] **longblob 读写验证**：中文+HTML+✓ 写入读回字节级一致（curl 实测）
- [x] PHPUnit：NoticeTest 12 用例（@Xss 命中/放行矩阵含 `1<2>0` 误伤原样断言、noticeInput 文案、ids 解析、unreadCount 口径）——52 tests 141 assertions 全绿

## Task 2 · 控制器与路由（含 2.0.0 收编）

- [x] NoticeController 扩容 13 方法：#[Perm] 9 处（spec 写 8 漏数 readUsers/list，经典版实锤同 9）+ #[Log] 3 处
- [x] **listTop 收编**：真实查询 + unreadCount 列表内计数；响应键与主框架消费 JS 一字不差（浏览器实测徽章/列表/置灰）
- [x] markRead / markReadAll 路由补齐；幂等实测（COUNT=1）
- [x] remove 两连删（先读后公告；DB 双表同清实测）
- [x] add/edit：@Xss + @Validated 全套文案；createBy/updateBy 落库
- [x] 全局类 use 纪律（AjaxResult/PageQuery/TableDataInfo/TpConstant/Perm/Log 均 use）
- [x] 路由：2.0.0 单条 listTop 替换为 group 13 条；固定段先于 :param
- [x] RepeatSubmit 未挂（grep 0）
- [x] curl 级 13 端点自测（Temp/m7_test.py 13 项 + 补测 searchValue/空表形态）全过

## Task 3 · POST /common/upload

- [x] CommonController::upload：multipart file → runtime/upload；<uuid>_<原始名>；防 ".."；白名单逐项抄 DEFAULT_ALLOWED_EXTENSION
- [x] 返回 {code:0, url:domain+/upload/文件名, fileName, newFileName, originalFilename}
- [x] 路由 POST common/upload + GET upload/:fileName（对位 /profile 静态映射；MIME 表）；仅登录态（POST 无静态后缀走 LoginAuth）
- [x] curl 级：png 上传 → 落盘 → url GET image/png；.exe → 500「上传文件格式不允许：exe」

## Task 4 · 页面模板（5 页）

- [x] notice/index.html：三按钮 + viewUrl + 阅读用户 popupRight + 类型/状态徽章 + 搜索三字段（subagent 移植 + 引擎冒烟；浏览器联调实测）
- [x] notice/add.html：summernote 五参数 + sendFile + 下拉/radio + noticeContent 提交链（浏览器实测 172 工具栏按钮渲染）
- [x] notice/edit.html：hidden noticeId + code 回填 + 隐藏域默认转义（浏览器实测回填一字不差）
- [x] notice/view.html：类型角标三分支 + meta + `{$notice.notice_content|raw}` + 空内容兜底（浏览器实测外链图片加载）
- [x] notice/readUsers.html：search:true + showSearch:false + noticeId 注入 + readTime formatter（浏览器实测过滤）
- [x] 模板自查（subagent grep + 引擎冒烟 5 页×授权/拒权 10 用例：无 th: 残留、check_perm 输出形式、volist/if 配平、标题转义/正文 raw 断言）

## Task 5 · 端到端验收（浏览器级）

- [x] 公告列表页：3 行预置 + 徽章 + notice_id desc；标题链接 → 右侧滑出弹层（screenshot 确认）
- [x] 新增全屏：summernote 富文本（加粗+列表）→ DB 字节级一致 → 列表/详情回显
- [x] 修改全屏：编辑器回填不丢标签 → 修改保存
- [x] @Xss：标题含 script → 500 文案（curl + PHPUnit 双验）
- [x] 删除：确认框 → 公告行消失 + 已读行同清（两连删 DB 核对）
- [x] **主框架公告链路回归**：3 行 + 徽章 3 → 点开 markRead 落库 + 徽章递减 + 置灰 → 全部已读 + 徽章消失；空形态 {code:0,data:[],unreadCount:0} 回归实测（status 全关等价法，公告表全程无损）
- [x] 阅读用户弹层：表格渲染 + 搜索过滤 + admin 行三表 JOIN 数据正确
- [x] ry 表零 schema 变更；预置 3 行（含 2229B 样本）+ 已读 0 行完整复原

## Task 6 · 收尾

- [x] PHPUnit 全绿（52 tests 141 assertions，NoticeTest 12 用例新增）；spec.md 实施记录回填；checklist 勾选；README 总表更新（7.0.0 ✅）
- [x] deviations 处置回填：富文本净化/输入侧 XSS 均不登记（无行为差异）
- [x] 测试数据清理：测试公告/已读行/操作日志 166+/登录日志/上传图片全清；admin 会话正常
