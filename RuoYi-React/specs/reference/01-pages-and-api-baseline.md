# RuoYi-Vue3 前端功能基线调查报告（React + AntD 重写用）

> 调查根目录：`RuoYi-Vue3`（Vue 3.5.26 + Vite 6 + Element Plus 2.13.1 + Pinia + vue-router 4，版本 v3.9.2）
> 技术栈关键依赖（package.json）：element-plus、echarts 5.6、@vueup/vue-quill（富文本）、vue-cropper 1.1.1（头像裁剪）、vuedraggable 4.1 + sortablejs（拖拽）、jsencrypt 3.3.2（RSA）、js-cookie、file-saver、clipboard、fuse.js、js-beautify、nprogress。
> 生成于 2026-09-27，由 AI 对源码全量调查产出；作为 RuoYi-React 重写的功能范围基线，行为疑问以 RuoYi-Vue3 源码为准。

---

## 1. 页面清单（src/views，共 51 个 .vue）

### 1.1 顶层公共页面

| 文件路径 | 用途 | 页面类型 | 特殊控件 |
|---|---|---|---|
| `src/views/login.vue` | 登录页 | 纯表单页 | 验证码图片(base64)、jsencrypt RSA 加密 cookie 密码、记住我 cookie、路由 redirect 参数 |
| `src/views/register.vue` | 注册页 | 纯表单页 | 验证码图片、密码强度动态校验 `usePasswordRule`（`src/utils/passwordRule.js`） |
| `src/views/index.vue` | 首页 | 静态展示页 | 无 echarts；版本卡/技术选型/联系信息/更新日志(el-collapse 折叠 30+ 版本)/捐赠二维码卡 |
| `src/views/lock.vue` | 锁屏解锁页 | 特殊页(全屏) | 头像显示、密码输入、调 `unlockScreen` API + lockStore |
| `src/views/redirect/index.vue` | 页签刷新中转页 | 特殊页 | 无（router.replace 转发） |
| `src/views/error/401.vue` | 无权限页 | 特殊页 | gif 图片 |
| `src/views/error/404.vue` | 404 页 | 特殊页 | 无 |

### 1.2 系统管理 system/

| 文件路径 | 用途 | 页面类型 | 特殊控件 |
|---|---|---|---|
| `src/views/system/user/index.vue` | 用户管理 | 标准 CRUD 表格（左侧部门树分栏） | TreePanel 部门树、ExcelImportDialog 导入弹窗、用户详情 Drawer(view.vue)、重置密码 $prompt、v-hasPermi、DictTag(sys_normal_disable/sys_user_sex)、导出 `proxy.download('system/user/export')` |
| `src/views/system/user/view.vue` | 用户详情抽屉 | 详情页(el-drawer 68%) | 只读信息展示、角色/岗位分组 |
| `src/views/system/user/authRole.vue` | 用户分配角色（独立路由页 `/system/user-auth/role/:userId`） | 表格选择页 | 前端分页 el-table、reserve-selection 复选、v-hasPermi |
| `src/views/system/role/index.vue` | 角色管理 | 树表 CRUD（el-table row-key + expand） | 菜单权限 el-tree、数据权限 el-tree(deptTree)、展开/折叠切换、v-hasPermi、DictTag、changeStatus 开关 |
| `src/views/system/role/authUser.vue` | 角色分配用户（独立路由页 `/system/role-auth/user/:roleId`） | 标准 CRUD 表格 | 添加/取消授权、v-hasPermi |
| `src/views/system/role/selectUser.vue` | 选择用户弹窗 | 弹窗表格选择 | el-dialog + 多选表格 |
| `src/views/system/menu/index.vue` | 菜单管理 | 树表 CRUD（可展开/折叠） | IconSelect 图标选择器、SvgIcon 显示、批量保存排序 updateMenuSort（勾选+Sort 按钮）、v-hasPermi、DictTag |
| `src/views/system/dept/index.vue` | 部门管理 | 树表 CRUD | 展开/折叠、批量保存排序 updateDeptSort、v-hasPermi、DictTag |
| `src/views/system/post/index.vue` | 岗位管理 | 标准 CRUD 表格 | v-hasPermi、DictTag(sys_normal_disable) |
| `src/views/system/dict/index.vue` | 字典类型管理 | 标准 CRUD 表格 | DictDataDrawer 抽屉详情（detail.vue）、刷新缓存、v-hasPermi、DictTag |
| `src/views/system/dict/detail.vue` | 字典数据抽屉（字典类型行内打开） | 详情页(el-drawer 700px) | 内嵌字典数据列表 |
| `src/views/system/dict/data.vue` | 字典数据页（独立路由 `/system/dict-data/index/:dictId`） | 标准 CRUD 表格 | 字典类型下拉(optionselect)、DictTag、v-hasPermi |
| `src/views/system/config/index.vue` | 参数设置 | 标准 CRUD 表格 | 刷新缓存、v-hasPermi、DictTag(sys_yes_no) |
| `src/views/system/notice/index.vue` | 通知公告 | 标准 CRUD 表格 | 富文本 `<editor>`(Quill)、公告详情复用 `@/layout/components/HeaderNotice/DetailView`、ReadUsers 阅读用户弹窗、v-hasPermi、DictTag(sys_notice_status) |
| `src/views/system/notice/ReadUsers.vue` | 公告已读用户弹窗 | 弹窗表格 | 分页、Search 图标查询 |

### 1.3 系统监控 monitor/

| 文件路径 | 用途 | 页面类型 | 特殊控件 |
|---|---|---|---|
| `src/views/monitor/online/index.vue` | 在线用户 | 标准 CRUD 表格（只读+强退） | v-hasPermi(monitor:online:forceLogout) |
| `src/views/monitor/job/index.vue` | 定时任务 | 标准 CRUD 表格 | **Crontab 组件**（Cron 表达式生成器弹窗 `@/components/Crontab`）、任务详情弹窗 detail.vue、立即执行/状态切换、v-hasPermi、DictTag(sys_job_group/sys_job_status) |
| `src/views/monitor/job/detail.vue` | 任务/调度日志详细弹窗 | 详情页(el-dialog 780px) | 卡片式详情布局、DictTag、pre 代码展示 |
| `src/views/monitor/job/log.vue` | 调度日志（独立路由 `/monitor/job-log/index/:jobId`） | 标准 CRUD 表格 | 详细弹窗(JobDetail)、导出、清空、v-hasPermi |
| `src/views/monitor/druid/index.vue` | 数据监控 | 特殊页(iframe) | `@/components/iFrame` 嵌入 `{VITE_APP_BASE_API}/druid/login.html`；`getPlatformInfo` 判断 features.druidMonitor 降级提示 |
| `src/views/monitor/server/index.vue` | 服务监控 | 详情展示页 | 纯表格展示 CPU/内存/JVM/磁盘/系统信息；`getPlatformInfo` 判断 features.serverMonitor |
| `src/views/monitor/cache/index.vue` | 缓存监控 | 详情展示页 | **echarts**：命令统计**玫瑰图(pie roseType)** + 内存**仪表盘(gauge)**，主题 "macarons" |
| `src/views/monitor/cache/list.vue` | 缓存列表 | 三栏联动页 | 左缓存名称表/中键名表/右内容表，清理名称/键/全部缓存 |
| `src/views/monitor/operlog/index.vue` | 操作日志 | 标准 CRUD 表格（只读） | 详细弹窗 detail.vue、导出、清空、v-hasPermi、DictTag(sys_oper_type/sys_common_status) |
| `src/views/monitor/operlog/detail.vue` | 操作日志详细弹窗 | 详情页(el-dialog 780px) | 卡片式分区详情（基本/操作人员/请求/返回/异常） |
| `src/views/monitor/logininfor/index.vue` | 登录日志 | 标准 CRUD 表格（只读） | 解锁账户按钮、导出、清空、v-hasPermi、DictTag(sys_common_status) |

### 1.4 系统工具 tool/

| 文件路径 | 用途 | 页面类型 | 特殊控件 |
|---|---|---|---|
| `src/views/tool/gen/index.vue` | 代码生成主列表 | 标准 CRUD 表格 | 预览弹窗（el-tabs 按生成文件分组 + `<pre>` 展示 + v-copyText 复制指令，**无语法高亮库**）、导入表弹窗、创建表弹窗、生成代码(zip 下载)、同步库、v-hasPermi |
| `src/views/tool/gen/importTable.vue` | 导入数据库表弹窗 | 弹窗表格选择 | 查询 db 表、多选导入 |
| `src/views/tool/gen/createTable.vue` | 创建表弹窗 | 弹窗表单 | 手工建表 SQL |
| `src/views/tool/gen/editTable.vue` | 修改生成配置（独立路由 `/tool/gen-edit/index/:tableId`） | Tab 表单页 | el-tabs(基本信息/字段信息/生成信息)；**Sortable 行拖拽排序**(字段表)；字典 optionselect |
| `src/views/tool/gen/basicInfoForm.vue` | 生成配置-基本信息表单 | 子表单 | 表名/描述/实体/作者 |
| `src/views/tool/gen/genInfoForm.vue` | 生成配置-生成信息表单 | 子表单 | 模板类型 select（单表/树表/主子表）、前端类型、生成方式(zip/自定义路径)、上级菜单 **el-tree-select**（加载 menu 列表）、树表配置、子表关联配置 |
| `src/views/tool/build/index.vue` | 表单构建（拖拽设计器） | 特殊页 | **vuedraggable** 拖入/画布排序、IconsDialog、RightPanel、CodeTypeDialog 导出(Vue/HTML + js-beautify 格式化)、ClipboardJS 复制 |
| `src/views/tool/build/DraggableItem.vue` | 设计器组件递归渲染 | 子组件 | vuedraggable 嵌套 + `@/utils/generator/render` 动态渲染 |
| `src/views/tool/build/RightPanel.vue` | 设计器右侧属性面板 | 子组件 | 选项拖拽排序、el-tree draggable（树选项/布局预览） |
| `src/views/tool/build/IconsDialog.vue` | Element Plus 图标选择弹窗 | 子组件 | 遍历 @element-plus/icons-vue 全量图标 |
| `src/views/tool/build/CodeTypeDialog.vue` | 导出类型选择弹窗 | 子组件 | radio-button(file/vue) + 文件名 |
| `src/views/tool/build/TreeNodeDialog.vue` | 树选项添加弹窗 | 子组件 | id/label 树字段映射 |
| `src/views/tool/swagger/index.vue` | 接口文档 | 特殊页(iframe) | iFrame 组件；`getPlatformInfo` 判断 springdoc 功能开关降级 |

### 1.5 个人中心 profile/

| 文件路径 | 用途 | 页面类型 | 特殊控件 |
|---|---|---|---|
| `src/views/system/user/profile/index.vue` | 个人中心（路由 `/user/profile/:activeTab?`，name: Profile） | 左卡片+右 Tab 页 | 左：头像+用户信息列表；右 el-tabs：基本资料/修改密码；支持 `params.activeTab` 直达 resetPwd（登录后初始密码/过期密码提示会跳转此 tab） |
| `src/views/system/user/profile/userInfo.vue` | 基本资料表单 | 纯表单页 | 昵称/手机/邮箱/性别单选，校验手机号正则+邮箱格式 |
| `src/views/system/user/profile/resetPwd.vue` | 修改密码 | 纯表单页 | 旧/新/确认密码，新密码走 `infoPwdValidator`（pwdChrtype 动态规则 0-4） |
| `src/views/system/user/profile/userAvatar.vue` | 头像上传裁剪 | 弹窗组件 | **vue-cropper**（固定 200x200 截图框、缩放/左右旋转、实时预览）、el-upload 自定义 http-request、getCropBlob |

### 全局注册组件（main.js line 61-73，重写需对应实现）

`DictTag`、`Pagination`、`FileUpload`、`ImageUpload`、`ImagePreview`、`RightToolbar`、`Editor`(Quill 富文本)、`svg-icon`；指令：`v-hasPermi`(`src/directive/permission/hasPermi.js`)、`v-hasRole`(`hasRole.js`)、`v-copyText`(`src/directive/common/copyText.js`)。

（注：FileUpload/ImageUpload/ImagePreview 在 views 中仅 notice 富文本内嵌使用，但属全局组件基线。）

---

## 2. 接口清单（src/api，20 个文件）

### 认证与会话 `src/api/login.js`

| 函数 | 方法 | 端点 |
|---|---|---|
| login(username,password,code,uuid) | POST | `/login`（headers: isToken:false, repeatSubmit:false） |
| register(data) | POST | `/register`（isToken:false） |
| getInfo() | GET | `/getInfo` |
| unlockScreen(password) | POST | `/unlockscreen`（body {password}） |
| logout() | POST | `/logout` |
| getCodeImg() | GET | `/captchaImage`（isToken:false, timeout 20s） |

### 平台信息 `src/api/platform.js`

| getPlatformInfo() | GET | `/getPlatformInfo`（返回语言/框架版本与 features 开关，druid/swagger/服务监控据此降级） |
|---|---|---|

### 动态路由 `src/api/menu.js`

| getRouters() | GET | `/getRouters` |
|---|---|---|

### 用户 `src/api/system/user.js`

| 函数 | 方法 | 端点 |
|---|---|---|
| listUser(query) | GET | `/system/user/list` |
| getUser(userId) | GET | `/system/user/{userId}`（空则 `/system/user/`） |
| addUser(data) | POST | `/system/user` |
| updateUser(data) | PUT | `/system/user` |
| delUser(userId) | DELETE | `/system/user/{userId}` |
| resetUserPwd(userId,password) | PUT | `/system/user/resetPwd` |
| changeUserStatus(userId,status) | PUT | `/system/user/changeStatus` |
| getUserProfile() | GET | `/system/user/profile` |
| updateUserProfile(data) | PUT | `/system/user/profile` |
| updateUserPwd(oldPassword,newPassword) | PUT | `/system/user/profile/updatePwd` |
| uploadAvatar(data) | POST | `/system/user/profile/avatar`（**Content-Type: application/x-www-form-urlencoded**，实际传 FormData avatarfile） |
| getAuthRole(userId) | GET | `/system/user/authRole/{userId}` |
| updateAuthRole(data) | PUT | `/system/user/authRole`（params） |
| deptTreeSelect() | GET | `/system/user/deptTree` |

页面另有（非此文件）：导出 `GET /system/user/export`（proxy.download）、导入 `POST /system/user/importData?updateSupport=`、模板 `GET /system/user/importTemplate`（ExcelImportDialog）。

### 角色 `src/api/system/role.js`

listRole GET `/system/role/list`；getRole GET `/system/role/{roleId}`；addRole POST `/system/role`；updateRole PUT `/system/role`；dataScope PUT `/system/role/dataScope`；changeRoleStatus PUT `/system/role/changeStatus`；delRole DELETE `/system/role/{roleId}`；allocatedUserList GET `/system/role/authUser/allocatedList`；unallocatedUserList GET `/system/role/authUser/unallocatedList`；authUserCancel PUT `/system/role/authUser/cancel`；authUserCancelAll PUT `/system/role/authUser/cancelAll`（params）；authUserSelectAll PUT `/system/role/authUser/selectAll`（params）；deptTreeSelect GET `/system/role/deptTree/{roleId}`

### 菜单 `src/api/system/menu.js`

listMenu GET `/system/menu/list`；getMenu GET `/system/menu/{menuId}`；treeselect GET `/system/menu/treeselect`；roleMenuTreeselect GET `/system/menu/roleMenuTreeselect/{roleId}`；addMenu POST `/system/menu`；updateMenu PUT `/system/menu`；**updateMenuSort PUT `/system/menu/updateSort`**（批量排序，menuIds+orderNums）；delMenu DELETE `/system/menu/{menuId}`

### 部门 `src/api/system/dept.js`

listDept GET `/system/dept/list`；listDeptExcludeChild GET `/system/dept/list/exclude/{deptId}`；getDept GET `/system/dept/{deptId}`；addDept POST `/system/dept`；updateDept PUT `/system/dept`；**updateDeptSort PUT `/system/dept/updateSort`**（批量排序）；delDept DELETE `/system/dept/{deptId}`

### 岗位 `src/api/system/post.js`

listPost GET `/system/post/list`；getPost GET `/system/post/{postId}`；addPost POST `/system/post`；updatePost PUT `/system/post`；delPost DELETE `/system/post/{postId}`

### 参数配置 `src/api/system/config.js`

listConfig GET `/system/config/list`；getConfig GET `/system/config/{configId}`；getConfigKey GET `/system/config/configKey/{configKey}`；addConfig POST `/system/config`；updateConfig PUT `/system/config`；delConfig DELETE `/system/config/{configId}`；refreshCache DELETE `/system/config/refreshCache`

### 字典 `src/api/system/dict/type.js` + `src/api/system/dict/data.js`

- type：listType GET `/system/dict/type/list`；getType GET `/system/dict/type/{dictId}`；addType POST；updateType PUT；delType DELETE `/system/dict/type/{dictId}`；refreshCache DELETE `/system/dict/type/refreshCache`；optionselect GET `/system/dict/type/optionselect`
- data：listData GET `/system/dict/data/list`；getData GET `/system/dict/data/{dictCode}`；**getDicts GET `/system/dict/data/type/{dictType}`**（字典核心接口）；addData POST；updateData PUT；delData DELETE `/system/dict/data/{dictCode}`

### 通知公告 `src/api/system/notice.js`

listNotice GET `/system/notice/list`；getNotice GET `/system/notice/{noticeId}`；addNotice POST；updateNotice PUT；delNotice DELETE `/system/notice/{noticeId}`；**listNoticeTop GET `/system/notice/listTop`**（首页顶部公告+已读状态）；**markNoticeRead POST `/system/notice/markRead`（params noticeId）**；**markNoticeReadAll POST `/system/notice/markReadAll`（params ids）**；**listNoticeReadUsers GET `/system/notice/readUsers/list`**

### 在线用户 `src/api/monitor/online.js`

list GET `/monitor/online/list`；forceLogout DELETE `/monitor/online/{tokenId}`

### 定时任务 `src/api/monitor/job.js` + `src/api/monitor/jobLog.js`

- job：listJob GET `/monitor/job/list`；getJob GET `/monitor/job/{jobId}`；addJob POST；updateJob PUT；delJob DELETE `/monitor/job/{jobId}`；changeJobStatus PUT `/monitor/job/changeStatus`；runJob PUT `/monitor/job/run`
- jobLog：listJobLog GET `/monitor/jobLog/list`；delJobLog DELETE `/monitor/jobLog/{jobLogId}`；cleanJobLog DELETE `/monitor/jobLog/clean`

### 操作日志 `src/api/monitor/operlog.js`

list GET `/monitor/operlog/list`；delOperlog DELETE `/monitor/operlog/{operId}`；cleanOperlog DELETE `/monitor/operlog/clean`

### 登录日志 `src/api/monitor/logininfor.js`

list GET `/monitor/logininfor/list`；delLogininfor DELETE `/monitor/logininfor/{infoId}`；**unlockLogininfor GET `/monitor/logininfor/unlock/{userName}`**；cleanLogininfor DELETE `/monitor/logininfor/clean`

### 缓存监控 `src/api/monitor/cache.js`

getCache GET `/monitor/cache`；listCacheName GET `/monitor/cache/getNames`；listCacheKey GET `/monitor/cache/getKeys/{cacheName}`；getCacheValue GET `/monitor/cache/getValue/{cacheName}/{cacheKey}`；clearCacheName DELETE `/monitor/cache/clearCacheName/{cacheName}`；clearCacheKey DELETE `/monitor/cache/clearCacheKey/{cacheKey}`；clearCacheAll DELETE `/monitor/cache/clearCacheAll`

### 服务监控 `src/api/monitor/server.js`

getServer GET `/monitor/server`

### 代码生成 `src/api/tool/gen.js`

listTable GET `/tool/gen/list`；listDbTable GET `/tool/gen/db/list`；getGenTable GET `/tool/gen/{tableId}`；updateGenTable PUT `/tool/gen`；importTable POST `/tool/gen/importTable`（params）；**createTable POST `/tool/gen/createTable`**（params）；previewTable GET `/tool/gen/preview/{tableId}`；delTable DELETE `/tool/gen/{tableId}`；genCode GET `/tool/gen/genCode/{tableName}`；synchDb GET `/tool/gen/synchDb/{tableName}`

（生成代码下载 zip 走 `$download.zip('/tool/gen/download/{tableName}', ...)`，`src/plugins/download.js`）

### 通用端点（非 src/api，但被全局使用）

`GET /common/download?fileName=&delete=`、`GET /common/download/resource?resource=`（`src/plugins/download.js`）；各业务导出 `GET /xxx/export`（`src/utils/request.js` 的 download 函数）。

---

## 3. 登录/注册页细节

### 登录 `src/views/login.vue`

1. **验证码**：进入页面即调 `getCodeImg()` → GET `/captchaImage`（isToken:false，20s 超时）。响应 `{ captchaEnabled, img, uuid }`；`captchaEnabled===undefined` 视为开启；开启时 `codeUrl = "data:image/gif;base64," + res.img` 直接展示 base64 图片（点击刷新），`loginForm.uuid = res.uuid`。登录失败后若验证码开启会自动刷新验证码。
2. **RSA 加密**：**登录请求本身不加密**——`login(username, password, code, uuid)` 明文 POST `/login`。jsencrypt（`src/utils/jsencrypt.js`，硬编码 512 位公私钥对）仅用于**"记住我"密码存 Cookie**：勾选 rememberMe 时 `Cookies.set("password", encrypt(password), { expires: 30 })`，回显时 `decrypt(password)`。即 RSA 的触发条件 = 勾选"记住密码"，30 天有效期（cookies: username / password(密文) / rememberMe）。
3. **记住我**：`el-checkbox`，逻辑如上；未勾选则移除三个 cookie。默认账号密码 admin/admin123 预填。
4. **注册开关**：`register ref` 默认 false，**代码中从未赋值**（`src/views/login.vue` 中"立即注册"链接 `v-if="register"` 恒不显示；若需保持行为一致，React 版同样不显示注册入口）。
5. 登录成功：`userStore.login()`（`src/store/modules/user.js`）→ setToken(res.token) → `useLockStore().unlockScreen()` → 按 route.query.redirect 跳转（保留其他 query 参数）。
6. 登录后 `getInfo()` 附加逻辑（user store）：头像处理（http(s) 直用 / 空→默认 profile.jpg / 否则加 VITE_APP_BASE_API 前缀）；`pwdChrtype` 存 session（驱动密码强度规则）；`isDefaultModifyPwd` 弹"初始密码"提示 → 跳 `/user/profile/resetPwd`；`isPasswordExpired` 弹"密码过期"提示 → 同上。

### 注册 `src/views/register.vue`

- 表单：username（2-20 长度）、password（`registerPwdValidator`：6-20 位 + 非法字符 `<> "'\|` 校验，`src/utils/passwordRule.js`，固定用规则'0'）、confirmPassword（一致性校验）、验证码+uuid（与登录同一 `/captchaImage` 接口）。
- 调 `register(registerForm)` POST `/register`（isToken:false）；成功弹 ElMessageBox HTML 提示"注册成功"→ 跳 `/login`；失败刷新验证码。
- 底部"使用已有账户登录"链接到 /login。页面标题取 `import.meta.env.VITE_APP_TITLE`，页脚取 `settings.js footerContent`。

---

## 4. 首页 `src/views/index.vue`

- **无 echarts、无待办**，纯静态展示页（`<script setup name="Index">` 仅一个 version ref='3.9.2' 和 goTarget window.open）。
- 区块：
  1. 左上：若依后台管理框架介绍文案 + 当前版本 v3.9.2 + "免费开源" el-tag + 访问码云/主页按钮
  2. 右上：技术选型两列表（后端：SpringBoot/Security/JWT/MyBatis/Druid/Fastjson；前端：Vue/Vuex/Element-ui/Axios/Sass/Quill）
  3. el-divider 后三卡片行（xs24/md12/lg8 响应式）：
     - 联系信息卡（官网/QQ群列表/微信/支付宝）
     - 更新日志卡（el-collapse accordion，v3.9.2 至 v1.0.0 共 30+ 个版本折叠项，内容为 <ol> 更新条目）
     - 捐赠支持卡（pay.png 收款码图片）

---

## 5. 个人中心 `src/views/system/user/profile/`

- **index.vue**：布局 = 左 el-col span6 个人信息卡（userAvatar 头像组件 + 用户名称/手机/邮箱/所属部门(含 postGroup)/所属角色(roleGroup)/创建日期）+ 右 el-col span18 基本资料卡内 el-tabs（tab "userinfo"=基本资料 userInfo、tab "resetPwd"=修改密码 resetPwd）。路由 `/user/profile/:activeTab?`（name Profile），onMounted 读取 `route.params.activeTab` 定位 tab。数据源 `getUserProfile()` GET `/system/user/profile`，响应含 `data`（用户）、`roleGroup`、`postGroup`。
- **userInfo.vue**：表单字段 nickName(必填,30)/phonenumber(必填,11,正则 `^1[3|4|5|6|7|8|9][0-9]\d{8}$`)/email(必填,50,格式)/sex(radio 0男1女)；提交 `updateUserProfile` PUT `/system/user/profile`。
- **resetPwd.vue**：oldPassword/newPassword/confirmPassword；新密码校验 `infoPwdValidator`（按 session 的 pwdChrtype 动态：0 任意合法字符 / 1 纯数字 / 2 纯字母 / 3 字母+数字 / 4 字母+数字+特殊字符 `~!@#$%^&*()-=_+`，长度均 6-20，见 `src/utils/passwordRule.js`）；提交 `updateUserPwd` PUT `/system/user/profile/updatePwd`（body {oldPassword,newPassword}）。
- **userAvatar.vue 头像上传**：
  - 点击头像开 el-dialog(800px)，内嵌 `vue-cropper`（VueCropper）：`img=userStore.avatar`、autoCrop 200×200、fixedBox 固定框、outputType png；右侧实时预览区；操作按钮：选择(el-upload 自定义 http-request 空实现，beforeUpload 校验 image/* 后 FileReader readAsDataURL)、放大/缩小(changeScale ±1)、左旋/右旋。
  - 提交：`cropper.getCropBlob(blob)` → `new FormData()` → `formData.append("avatarfile", blob, options.filename)`（filename 为原文件名，默认 'avatar'）→ **POST `/system/user/profile/avatar`**，请求头 `'Content-Type': 'application/x-www-form-urlencoded'`（实际 multipart 由 FormData 生效，axios 代码如此书写）。
  - 成功后 `options.img = VITE_APP_BASE_API + response.imgUrl`，同步 `userStore.avatar`，响应字段为 `{ imgUrl }`。

---

## 补充：重写时需注意的全局机制

- **请求层** `src/utils/request.js`：Bearer token 头；headers.isToken 跳过 token；headers.repeatSubmit:false 跳过防重（默认开启防重复提交）；`download(url, params, filename)` 通用导出（blob + Loading 遮罩）；返回码 401 → 重新登录提示。
- **下载插件** `src/plugins/download.js`：`/common/download`、`/common/download/resource`、`zip(url, name)` 三种。
- **字典机制**：`useDict(...types)`（`src/utils/dict.js`）+ dictStore 缓存 + `<dict-tag>` 组件，等价 React 需一套 DictTag + useDict hook。
- **权限**：`v-hasPermi` / `v-hasRole` 指令（`src/directive/permission/`），路由级用 `permissions`/`roles` meta 拦截。
- **动态路由**：登录后 `/getRouters` 动态挂载（`src/permission.js` + `src/store/modules/permission.js`）；静态路由见 `src/router/index.js`（redirect/login/register/404/401/index/lock/profile + 5 条带权限的动态路由页：AuthRole、AuthUser、DictData、JobLog、GenEdit）。
- **布局** `src/settings.js` 可配：navType（纯左/混合/纯顶部）、tagsView（含持久化、图标、card/chrome 样式）、fixedHeader、dynamicTitle、footerVisible/footerContent 等；另有 HeaderNotice 顶部公告组件（layout/components/HeaderNotice，listNoticeTop/markRead）。
