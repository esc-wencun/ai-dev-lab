# 7.0.0-通知公告 · spec

> **状态：✅ 已完成（2026-10-01 联调验收通过）**
> 对位经典若依 SysNoticeController（13 端点，含 2.0.0 已最小实现的 listTop——本模块收编为真实实现）+ SysNoticeServiceImpl / SysNoticeReadServiceImpl + SysNoticeReadMapper + templates/system/notice（端点 13 / 页面 5，逐方法核对实锤）。
> 依赖：1.0.0（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / LoginAuth）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下（按序号串行隐式满足）。
> 调研依据：reference 源码逐文件核对（Controller / ServiceImpl×2 / Mapper XML×2 / 页面 HTML×5 / SysNotice·SysNoticeRead 实体 / ry-ui.js $.operate.addFull·editFull·view·popupRight / **app/view/index/index.html 内嵌 listTop 消费 JS 实读**）+ **ry-tp 库实测**（sys_notice 10 列，**notice_content 是 longblob**；sys_notice_read 4 列 + uk_user_notice 唯一键；预置公告 **3 行**、已读 0 行；sys_menu perms：菜单 107 + 按钮 1035~1038）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、POST 分页参数、信封结构、权限两通道。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/system/NoticeController.php（**收编扩容**） | SysNoticeController | 2.0.0 最小版仅 listTop 空形态；本模块补齐 13 方法（listTop/markRead/markReadAll 转真实实现，其余新增） |
| app/service/NoticeService.php | ISysNoticeService / SysNoticeServiceImpl | 公告 CRUD（纯表操作，无缓存无级联） |
| app/service/NoticeReadService.php | ISysNoticeReadService / SysNoticeReadServiceImpl | 已读记录五方法（insert ignore 语义 / 带已读状态列表 / 未读数 / 批量 / 删除清理 / 阅读用户列表） |
| app/view/system/notice/*.html ×5 | templates/system/notice/ | notice（列表）/ add / edit（summernote）/ view（详情，**右侧弹出 iframe**）/ readUsers（已读用户） |
| POST /common/upload（**本模块落地**） | CommonController.uploadFile | summernote 图片上传依赖（onImageUpload 回调 POST common/upload）——经典版公告图片入库即 URL 引用，上传端点是公告页刚需；存 runtime/upload 目录，返回 {code:0, url, fileName, newFileName, originalFilename} |

**范围外**：头像上传（8.0.0 SysProfileController 走 avatarPath 子目录，本模块只落 uploadPath 主目录端点）、富文本存储净化策略拍板（见特殊行为 6，需人确认）。

## 页面清单（5）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 公告列表页 | view/system/notice/index.html | notice/notice.html | bootstrap-table | `prefix = ctx + "system/notice"`；options：url=prefix+/list、**viewUrl=prefix+"/view/{id}"**、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove；**无 sortName（默认不排序）**；modalName=公告；工具栏三按钮：**新增（$.operate.addFull 全屏）**/修改（$.operate.editFull 全屏）/删除(multiple)；搜索字段 noticeTitle/createBy（like）/noticeType 下拉（sys_notice_type）；noticeTitle 列链接 `$.operate.view(id)`（→ viewUrl 右侧 popupRight 弹层）；操作列三按钮：**阅读用户**（showReadUsers → $.modal.popupRight("「标题」阅读用户", prefix+"/readUsers/"+noticeId, '800px')，无权限控制）/编辑(editFlag)/删除(removeFlag)；状态徽章 datas=sys_notice_status、类型徽章 types=sys_notice_type |
| 2 | 公告新增页（**全屏 tab**） | view/system/notice/add.html | notice/add.html | **summernote 富文本** | `$.operate.addFull` → createUrl 打开（全屏 iframe 非 layer 弹层）；summernote 参数：placeholder '请输入公告内容'/height 192/lang zh-CN/followingToolbar:false/dialogsInBody:true/onImageUpload→sendFile（FormData POST common/upload，成功 `editor.insertImage(result.url, result.fileName)`）；提交：`$('.summernote').summernote('code')` 写入隐藏域 noticeContent → save(prefix+"/add")；noticeType 下拉（sys_notice_type，volist 预渲染）；status radio（sys_notice_status，默认正常）；include summernote-css/js（system/include/ 已有 43 片段） |
| 3 | 公告修改页（全屏 tab） | view/system/notice/edit.html | notice/edit.html | summernote | hidden noticeId；编辑器初始化后 `$('#editor').summernote('code', $("#noticeContent").val())` 回填（noticeContent 隐藏域回显值）；其余同新增；提交 save(prefix+"/edit") |
| 4 | 公告详情页（右侧弹层 iframe） | view/system/notice/view.html | notice/view.html | 纯展示 | $.operate.view(id) → viewUrl → popupRight 弹层（非 tab 非 layer.open 普通弹窗）；类型角标 switch（1 通知 warn 色 / 2 公告 success 色 / default 消息）；标题/meta（createBy/createTime/status 点 正常绿·关闭红）；**内容输出为富文本 HTML（不转义，对位 th:utext）**——TP 模板 `{$notice.notice_content|raw}`；空内容显示「暂无内容」；无操作按钮，footer include 保留 |
| 5 | 已读用户页（右侧弹层 iframe） | view/system/notice/readUsers.html | notice/readUsers.html | bootstrap-table（自带搜索） | options：url=prefix+"/readUsers/list"、**search:true、showSearch:false**、queryParams 覆写 search.noticeId=模板变量；列五：loginName/userName/deptName/phonenumber/readTime（sortable，formatter $.common.dateFormat(value,'yyyy-MM-dd HH:mm:ss')——**readTime 输出字符串，dateFormat 的 string 分支可解析**） |

> 命名沿用 3.0.0 惯例：列表页 index.html（对位 notice.html）、详情 view.html、已读用户 readUsers.html（对位 readUsers.html 原名）。

## 端点级 API 清单（notice 13 + 通用 upload 1）

### 通知公告 /system/notice（对位 SysNoticeController 13 方法）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/notice | system:notice:view | 无 | — | 渲染页面 1 |
| 2 | POST | /system/notice/list | system:notice:list | 无 | pageNum/pageSize/orderByColumn/isAsc（**白名单仅 create_time**——页面唯一 sortable 列）+ noticeTitle（like）/ noticeType / createBy（like） | TableDataInfo {code:0, msg:"查询成功", rows, total}；**固定 order by notice_id desc**（mapper 实锤，与排序参数叠加共存）；行字段驼峰 **selectVo 10 列**：noticeId/noticeTitle/noticeType/noticeContent/status/createBy/createTime/updateBy/updateTime/remark——**list 输出含 noticeContent 全文（经典版 selectVo 未排除，原样）** |
| 3 | GET | /system/notice/add | system:notice:add | 无 | — | 渲染页面 2（全屏 tab） |
| 4 | POST | /system/notice/add | system:notice:add | 通知公告, 1新增 | noticeTitle*、noticeType、noticeContent（富文本 HTML）、status、remark | @Validated（文案见特殊行为 5）；createBy=登录名；toAjax；**无缓存联动**（公告无缓存） |
| 5 | GET | /system/notice/edit/{noticeId} | system:notice:edit | 无 | 路径 noticeId | 渲染页面 3，变量 notice（noticeContent 回显，隐藏域原值 + 编辑器 code 回填） |
| 6 | POST | /system/notice/edit | system:notice:edit | 通知公告, 2修改 | noticeId、noticeTitle*、noticeType、noticeContent、status、remark | updateBy；toAjax |
| 7 | GET | /system/notice/view/{noticeId} | **无（仅登录态）** | 无 | 路径 noticeId | 渲染页面 4（popupRight iframe 内容）；**无权限注解**（普通用户读公告主入口——主框架公告弹层即走此端点） |
| 8 | GET | /system/notice/listTop | **无（仅登录态；2.0.0 已注册路由）** | 无 | — | **收编 2.0.0 空实现**：NoticeReadService::selectNoticeListWithReadStatus(userId, 5) + unreadCount = 列表中 !isRead 计数；返回 AjaxResult.success(list) + put("unreadCount")→ {code:0, msg:"操作成功", data:[…], unreadCount:n}；行含 **isRead 布尔**（@JsonProperty("isRead") 实锤——TP 输出键名 isRead）；消费者=主框架 index.html loadNoticeTop（实读：res.data / res.unreadCount / n.noticeTitle / n.noticeType / n.isRead / n.noticeId / n.createTime） |
| 9 | POST | /system/notice/markRead | **无（仅登录态）** | 无 | noticeId | NoticeReadService::markRead（insert ignore 语义）；固定 success()；消费者=主框架 previewNotice（fire and forget，2.0.0 版未实现该路由——**404 静默失败被前端容忍，收编后补齐**） |
| 10 | POST | /system/notice/markReadAll | **无（仅登录态）** | 无 | ids（逗号串；前端把当前下拉 5 条全部回传，**含已读的**） | NoticeReadService::markReadBatch（insert ignore 批量；空数组直接 return）；固定 success() |
| 11 | GET | /system/notice/readUsers/{noticeId} | system:notice:list | 无 | 路径 noticeId | 渲染页面 5，变量 notice |
| 12 | POST | /system/notice/readUsers/list | system:notice:list | 无 | pageNum/pageSize/orderByColumn/isAsc（**白名单 read_time**）+ noticeId + searchValue（bootstrap-table 自带搜索框全局词——匹配 login_name **或** user_name like） | TableDataInfo；行字段驼峰 6 列：userId/loginName/userName/deptName/phonenumber/readTime；**固定 order by read_time desc** |
| 13 | POST | /system/notice/remove | system:notice:remove | 通知公告, 3删除 | ids（逗号串） | **先 NoticeReadService::deleteByNoticeIds(ids) 清已读记录 → 再 NoticeService::deleteNoticeByIds(ids) 物理删公告**（顺序对位经典版 controller 两连调）；toAjax（以公告删除行数计；ids 全无效 → 0 行 → 「操作失败」） |
| 14 | POST | /common/upload | **无（仅登录态；summernote 图片上传，对位 CommonController.uploadFile）** | 无 | multipart 字段名 **file** | 存 runtime/upload（文件名 &lt;uuid&gt;_原始名防冲突）；成功 {code:0, msg:"操作成功", url:"&lt;scheme>://&lt;host>/upload/&lt;文件名&gt;", fileName:"/upload/&lt;文件名&gt;", newFileName, originalFilename}；失败 {code:500, msg};扩展名白名单（图片系 png/jpg/jpeg/gif/bmp + 常规文档，实施时对位 MimeTypeUtils.DEFAULT_ALLOWED_EXTENSION 抄全） |

统计：**控制器方法 14 个**（notice 13 + common upload 1，URL pattern 14 条——listTop/markRead/markReadAll/view/readUsers 等在 2.0.0 后新增注册）；**#[Perm] 共 8 处**（view/list/add/edit×2/readUsers×2/remove；**listTop/markRead/markReadAll/view/{id} 四端点无权限注解仅登录态**——普通用户读公告链路，经典版实锤）；**#[Log] 共 3 处**（INSERT/UPDATE/DELETE；**公告无 EXPORT**——controller 无 export 端点实锤）。

## 特殊行为清单（XSS / listTop·markRead 收编 / 公告类型 / 已读体系）

1. **summernote XSS 处理——经典版的三层防线与 TP 对位**（本模块最关键议题）：
   - 经典版防线①：全局 XssFilter（yml enabled:true）对 `/system/*` POST 生效，但 **excludes `/system/notice/*` 显式排除**——公告正文富文本 HTML 免于转义（转义后编辑器无法回显）；
   - 防线②：SysNotice.noticeTitle 标注 `@Xss`（XssValidator 正则检测 HTML 标签）——**标题禁含 HTML**「公告标题不能包含脚本字符」，正文无此约束；
   - 防线③：view.html 用 `th:utext` 原样输出正文（信任 summernote 产物），标题 `th:text` 转义输出；
   - TP 对位：**标题**——后端复刻 @Xss 校验（同款正则 `/<(\S*?)[^>]*>.*?|<.*? />`，命中 → 500「公告标题不能包含脚本字符」）+ 页面 JS escapeHtml 已有（index.html 主框架消费侧实测在用）；**正文**——入库不转义（对位 excludes 语义），输出侧 view.html `|raw` 原样（对位 utext）+ **列表页消费侧 title 属性转义已有**。**正文存储即富文本 HTML 是经典版原样行为，不复刻更严策略**；若要上净化（如 HTMLPurifier）属行为差异，须人拍板并登记 deviations（见拟登记表）。
2. **listTop / markRead / markReadAll 与 2.0.0 最小实现的关系（收编契约）**：
   - 2.0.0 现状（`app/controller/system/NoticeController.php`）：listTop 硬编码 `{code:0, msg, data:[], unreadCount:0}` 空形态（主框架「暂无公告」）；**markRead/markReadAll 路由不存在**（主框架点击公告 POST 404，fire-and-forget 静默）；
   - 本模块收编：listTop 改调 NoticeReadService 真实查询（status='0' 最新 5 条 + isRead 标记 + unreadCount）；补 markRead/markReadAll 两路由；**响应形态不变**（键名 data/unreadCount/isRead/noticeId/noticeTitle/noticeType/createTime 与主框架消费 JS 一字不差——实读消费代码固化于 spec 上表）；2.0.0 的 checklist「服务器错误弹窗」条目视为本模块验收起点；
   - 兼容回归断言：公告表清空时 listTop 仍返回 `{code:0, data:[], unreadCount:0}`（与 2.0.0 形态一致）。
3. **公告类型 1通知/2公告**（sys_notice_type 预置：1=通知 warning 默认 / 2=公告 success；sys_notice_status：0=正常 primary 默认 / 1=关闭 danger）：noticeType 下拉/radio 全部经 DictService::listByType 渲染不硬编码；主框架消费 JS 硬编码 `typeClass = n.noticeType === '1' ? 'warning' : 'success'`（静态资源原样）；view 页类型角标 1/2/default 三分支。
4. **已读体系细节**：
   - `sys_notice_read` 有 **uk_user_notice 唯一键（user_id+notice_id）**——insert ignore 防重复已读（TP 用 `INSERT IGNORE` 或 duplicate 容错，重复 markRead 不报错不重复插）；
   - listTop 的 isRead = LEFT JOIN sys_notice_read on (notice_id, user_id) 非空（SQL 层 case when）；unreadCount = 列表内 !isRead 计数（**Java 流式计数非独立 SQL**——同款在 PHP 侧对 5 条计数即可，selectUnreadCount 独立方法经典版存在但无人调用，保留服务方法不接端点）；
   - markReadAll 前端把 5 条全传（含已读）→ insert ignore 幂等；
   - 删除公告**两连删**（先清已读记录再删公告——controller 顺序实锤，防孤儿已读行）；
   - readUsers 列表 JOIN sys_user（del_flag='0'）+ LEFT JOIN sys_dept；searchValue 全局词双列 like。
5. **后端参数校验文案**（@Validated + BindException → error 首条）：
   - 标题「公告标题不能为空」「公告标题不能超过50个字符」（DB varchar(50) 同口径）+ @Xss「公告标题不能包含脚本字符」；
   - 类型/状态/内容无后端校验注解（noticeType/status 由前端字典约束，content 无 @Size——longblob 容量天然上限）。
6. **notice_content 列是 longblob**（实测 SHOW COLUMNS）：写入侧 PDO 绑定二进制安全（think-orm 参数绑定默认即安全，**注意 utf8mb4 中文经 longblob 存取需连接 charset 一致**——端到端写入中文+HTML 后读回比对字节级一致）；读取侧经典版 mapper `cast(notice_content as char)`，TP 侧读出即字符串（PDO blob 转字符串），无需显式 cast。
7. **列表默认排序 notice_id desc**（mapper 实锤）+ 用户排序参数 orderByColumn=create_time 可叠加（PageHelper OrderBy 语义：**追加 order by，两排序共存**——TP 侧 `order('notice_id desc')` 后再按 PageQuery 追加，或合并为一条 order 子句：`create_time asc/desc, notice_id desc`，实施时取实现简单且行为等价的）；readUsers/list 同理 `read_time desc` 固定。
8. **防重复提交**：SysNoticeController 零 @RepeatSubmit（grep 实锤）→ 不挂 RepeatSubmit 中间件。
9. **视图 layer / include**：控制器在 system 层 → 模板路径 'notice/index'、'notice/add'、'notice/edit'、'notice/view'、'notice/readUsers'（layer 相对）；include 片段用 app/view/system/include/ 局部副本：**summernote-css / summernote-js**（已有）；全局类（AjaxResult/TpConstant/PageQuery/TableDataInfo/Perm/Log）在 namespaced 文件必须 use。
10. **前端 quirk（原样保留，勿修）**：① 新增/修改走 addFull/editFull 全屏 tab（非弹层，与其他模块风格不同是经典版原样）；② 阅读用户按钮无权限控制；③ readUsers 页 search:true+showSearch:false（有搜索框但无检索区切换钮）；④ view 弹层无关闭按钮自定义（popupRight 自带 shadeClose）。

## 数据权限声明

- **本模块无数据权限**：SysNoticeServiceImpl / SysNoticeReadServiceImpl 均无 @DataScope（实锤）；公告全员可见（普通用户经主框架公告弹层/view 端点可读，无权限注解）。
- 无 admin 保护逻辑。
- 无 #[Perm] 端点四个：listTop / markRead / markReadAll / view/{noticeId}（+ upload）——仅登录态（LoginAuth 主链覆盖）。

## 关键设计说明

1. **NoticeService 薄**：selectNoticeById / selectNoticeList（like/eq 过滤 + 固定 notice_id desc + Query 构造器）/ insertNotice / updateNotice / deleteNoticeByIds（物理删 in）——纯表操作无缓存无级联（对位经典版零业务逻辑的 ServiceImpl）。
2. **NoticeReadService 五方法**：markRead（insert ignore）/ markReadBatch（空数组短路 + 批量 insert ignore）/ deleteByNoticeIds / selectNoticeListWithReadStatus(userId, limit)（LEFT JOIN + case when isRead，**行输出驼峰 isRead 布尔**）/ selectReadUsersByNoticeId(noticeId, searchValue)（三表 JOIN，行 6 驼峰列）；selectUnreadCount 保留（服务层完整性，无端点消费）。
3. **PHPUnit 先行**：selectNoticeListWithReadStatus 的 SQL 组装/isRead 口径（真实库三行公告断言）、@Xss 正则命中/放行矩阵（`<script>`/`<img onerror>`/纯文本/空串）、markReadBatch 空数组短路、ids 解析。
4. **/common/upload 端点设计**：runtime/upload 目录（与 download 目录同级的 runtime 约定）；url 拼 scheme://host 前缀（对位 serverConfig.getUrl()——TP `$request->domain()`）；**上传文件名 &lt;uuid&gt;_&lt;原始名&gt;**（对位 FileUploadUtils 编码策略，防中文/重名）；白名单对位 MimeTypeUtils.DEFAULT_ALLOWED_EXTENSION 逐项抄；checkAllowDownload 同款防 ".."。此端点为 8.0.0 头像上传预留复用（avatarPath 子目录届时扩展）。
5. **输出键名双轨**（前序踩坑#1）：list/readUsers/listTop 输出**驼峰**；add/edit 回显 assign **下划线**（$notice.notice_title 直取 + noticeContent 隐藏域回显 `{$notice.notice_content}`——**富文本回显勿加 |raw 转义反转问题：隐藏域 value 输出需 HTML 实体转义防属性截断，TP 默认 {$var} 即转义，恰好正确，勿加 raw**）。
6. **页面 JS 权限变量接线**（check_perm 输出形式）：editFlag/removeFlag 同 3.0.0 模式；addFull/editFull 按钮用 `{if check_perm(...)}` **不可用**——用输出形式或按钮 hidden class 拼接（与 3.0.0 一致）。
7. **主框架公告弹层回归**：7.0.0 落地后主框架「暂无公告」→ 3 行真实公告；点开 previewNotice → layer 右滑层渲染 view 页 + markRead 落库（sys_notice_read 出行）；未读徽章 unreadCount 正确递减；「全部已读」→ 5 键 insert；此链路是本模块端到端主场景。
8. **预置数据基线**（动工检查单#2/#5 实测）：sys_notice 3 行（id 1 公告型「温馨提醒…新版本发布啦」内容 15B 短文本；id 2 通知型「维护通知」12B；id 3 通知型「若依开源框架介绍」**2229B 富文本 HTML 含外链图片/表格样式 span/font**——summernote 产物兼容性天然验证样本）；sys_notice_read 0 行。测试数据自建自删（物理删）。
9. **路由注册**：route/app.php 的 `system/notice/listTop` 一条路由（2.0.0）**替换**为完整 group（12 条）；group 内固定段路由（listTop/markRead/markReadAll/view/readUsers/readUsers/list）注册在 :param 路由之前；/common/upload 追加 POST 路由。
10. **无导出**：公告无 @Excel 注解、无 export 端点（动工检查单#4 核对结论——本模块不动用 ExcelExportService）。

## 拟登记 deviations

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| 富文本正文入库前净化 | XssFilter 对 `/system/*` 生效但 excludes `/system/notice/*`——正文实际**未净化**直存（三层防线里正文只有「输出信任」没有「输入过滤」） | 同经典版：正文不转义直存 + 输出 raw；**不加 HTMLPurifier**（加了属行为增强非复刻，且可能破坏 2229B 预置样本的 font/span 标签链） | 建议**不登记**（TP 行为=经典行为，无差异）；若实施时决定上净化则当场登记 |
| 输入侧全局 XSS 转义（标题外的普通字段） | XssFilter 对 /system/* POST 参数 EscapeUtil.clean（HTMLFilter 白名单滤标签）+ trim | 6.0.0 已提候选；本模块标题走 @Xss 复刻，**其余字段（noticeTitle 之外无文本输入面）实际影响面≈0** | 与 6.0.0 同议题合并定：维持不复刻则两模块合并登记一条「输入侧全局 XSS 转义未复刻（@Xss 点位已复刻）」 |
| longblob 列存储形态 | MyBatis 强制 cast(notice_content as char) 读出 | PDO longblob 读出即字符串（连接 charset utf8mb4 一致时字节等价） | 无需登记（行为等价，防误改说明） |

另注（非 deviation，防误改说明）：正文富文本直存直出（excludes+utext 双原样）、markReadAll 全量幂等、阅读用户按钮无权限、view 端点无权限注解，均为经典版原样行为，**照抄不修正**。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-09-30~10-01）

- NoticeService（五方法 + noticeInput/@Xss）+ NoticeReadService（六方法，INSERT IGNORE 幂等）；NoticeController 13 方法收编（listTop 转真实查询，markRead/markReadAll 补齐）；/common/upload + /upload/:fileName 静态服务落地。
- 5 页模板 subagent 移植 + 引擎冒烟 10 用例；#[Perm] 9 处（spec 正文写 8 为漏数 readUsers/list，经典版同 9）；#[Log] 3 处。
- 浏览器联调全链路：主框架公告链路（徽章 3→2→隐藏、markRead 落库、全部已读）、popupRight 详情（外链图片加载）、summernote 富文本（提交字节级一致、编辑回填）、readUsers 搜索过滤、两连删。
- PHPUnit 52 tests 141 assertions 全绿（NoticeTest 12 用例新增）；空表形态 2.0.0 回归实测（status 全关等价法）。
- 修正 2 处：会话 userId 键实为 user_id（SessionService 注释已勘误）；listTop 输出 status 键额外存在（主框架消费 JS 不读，无影响）。
- 测试数据全清理（公告 3 行/已读 0 行/日志/upload 复原）。
