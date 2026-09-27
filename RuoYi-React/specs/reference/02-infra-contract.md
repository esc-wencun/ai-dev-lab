# RuoYi-Vue3 前端基础设施行为契约（React + Ant Design 重写基线）

> 基于源码精读，路径均相对 `RuoYi-Vue3/`。每条标注文件与关键行号。重写工程必须逐条复刻这些行为（含文案、状态码语义、存储键名），否则与三版后端（Java/Python/Go 共用此前端）不兼容。
> 生成于 2026-09-27，由 AI 对源码全量调查产出。

---

## 1. HTTP 请求层 — `src/utils/request.js`

### 1.1 axios 实例（L14-21）

- 全局默认 `Content-Type: application/json;charset=utf-8`（L14）。
- `baseURL = import.meta.env.VITE_APP_BASE_API`（开发 `/dev-api`，生产 `/prod-api`）。
- `timeout: 10000`（10 秒，验证码接口单独改为 20000，见 `src/api/login.js` L67）。

### 1.2 请求拦截器（L24-73）

- **Authorization 注入**（L26-33）：`getToken()` 有值且 `config.headers.isToken !== false` 时，设 `Authorization: 'Bearer ' + token`。token 来自 Cookie `Admin-Token`（见 §2）。传 `headers: { isToken: false }` 可跳过（login/register/captchaImage 均这样用）。
- **GET 参数映射**（L35-40）：`method === 'get' && config.params` 时，用 `tansParams()`（`src/utils/ruoyi.js` L191-211：encodeURIComponent，嵌套对象展开为 `prop[key]=value`，跳过 null/''/undefined，结尾多一个 `&` 用 slice 去掉）把参数拼进 URL，然后清空 `config.params`。
- **防重复提交**（L28-30, L41-68）：
  - `headers.repeatSubmit === false` 时跳过（L28）。
  - 只对 `post` / `put` 生效（L41）。
  - interval 取 `headers.interval || 1000` ms（L30，自定义 interval 通过 headers 传入）。
  - 构造 `{url, data: JSON.stringify(body), time: Date.now()}`（L42-46）。
  - **序列化后大小 ≥ 5MB（5*1024*1024）跳过检查**并 console.warn「请求数据大小超出允许的5M限制…」（L47-52）。
  - 存储位置：**sessionStorage** 键 `sessionObj`（经 `cache.session.setJSON`，即 JSON 字符串，L53-55）。
  - 判定重复的条件（L60）：`上次 data === 本次 data && 本次 time - 上次 time < interval && url 相同` → console.warn 并 `Promise.reject(new Error('数据正在处理，请勿重复提交'))`（不发请求）；否则更新 sessionStorage 记录。
- 请求拦截器 error 分支仅 console.log + Promise.reject（L70-73，注意原代码 reject 未 return，为已知瑕疵）。

### 1.3 响应拦截器 — 按 **body 里的 code** 分支，不按 HTTP 状态码（L76-109）

- `code = res.data.code || 200`（未设 code 视为成功）；`msg = errorCode[code] || res.data.msg || errorCode['default']`（errorCode 表见 §1.5）。
- **二进制响应**（L82-84）：`res.request.responseType === 'blob' || 'arraybuffer'` 时直接 `return res.data`，不走 code 分支。
- **code 401**（L85-97）：
  - 用模块级标志 `isRelogin.show`（导出，permission.js 也用）防重复弹窗。
  - 弹 `ElMessageBox.confirm('登录状态已过期，您可以继续留在该页面，或者重新登录', '系统提示', { confirmButtonText: '重新登录', cancelButtonText: '取消', type: 'warning' })`。
  - 确认 → `useUserStore().logOut()` → `location.href = '/index'`；取消 → 复位 isRelogin。
  - 无论确认与否都 `Promise.reject('无效的会话，或者会话已过期，请重新登录。')`。
- **code 500**（L98-100）：`ElMessage.error(msg)`（普通错误气泡），reject Error(msg)。
- **code 601**（L101-103）：`ElMessage.warning(msg)`，reject Error(msg)。（601 是"警告级"业务码。）
- **其他非 200**（L104-106）：`ElNotification.error({ title: msg })`（右上角通知），reject 'error'。
- **code 200**：`Promise.resolve(res.data)` —— 注意返回的是整个 body（含 code/msg/data/rows/total），不是 data 字段。
- **HTTP 层错误**（L111-123）：
  - `Network Error` → 文案「后端接口连接异常」；
  - message 含 `timeout` → 「系统接口请求超时」；
  - 含 `Request failed with status code` → 「系统接口」+ 末 3 位（HTTP 状态码）+「异常」。
  - `ElMessage.error(message, duration: 5000)`，reject。

### 1.4 通用下载函数 `download(url, params, filename, config)`（L127-151）

- 先开全屏 Loading：「正在下载数据，请稍候」，背景 `rgba(0,0,0,0.7)`。
- **POST** 请求 + `transformRequest: [(params) => tansParams(params)]` + `Content-Type: application/x-www-form-urlencoded` + `responseType: 'blob'`，`...config` 可覆盖。
- 成功后 `blobValidate(data)`（`src/utils/ruoyi.js` L226-228：`data.type !== 'application/json'`）判断是真文件还是 JSON 错误体：
  - 是文件 → `new Blob([data])` + `file-saver` 的 `saveAs(blob, filename)`；
  - 是 JSON → `await data.text()` → `JSON.parse` → `errorCode[code] || msg || default` → `ElMessage.error(errMsg)`。
- Loading 关闭在 then 内；catch 分支 console.error + `ElMessage.error('下载文件出现错误，请联系管理员！')` + 关 Loading。
- 该函数同时挂到 `app.config.globalProperties.download`（main.js L51）。

### 1.5 错误码表 — `src/utils/errorCode.js`（L1-7）

`401: '认证失败，无法访问系统资源'`；`403: '当前操作没有权限'`；`404: '访问资源不存在'`；`default: '系统未知错误，请反馈给管理员'`。

### 1.6 另一套下载 — `src/plugins/download.js`（$download）

`name(name, isDelete=true)` GET `/common/download?fileName=&delete=`；`resource(resource)` GET `/common/download/resource?resource=`；`zip(url, name)` GET（zip 类型 Blob + Loading）；文件名从响应头 `download-filename` 解码取；失败 `printErrMsg`（blob→text→JSON→errorCode）；手动带 `Authorization: Bearer`。`saveAs` 直接转发 file-saver。

---

## 2. Token 存取 — `src/utils/auth.js`（L1-16）

- Cookie 名：**`Admin-Token`**（TokenKey，L3）。
- `getToken()` = `Cookies.get('Admin-Token')`；`setToken(token)` = `Cookies.set('Admin-Token', token)`；`removeToken()` 同理。
- **未设 expires / maxAge**：会话级 Cookie（浏览器关闭即失效），非持久化。
- 登录页另有记住密码的 Cookies：`username` / `password`（password 经 RSA 加密存，`expires: 30` 天，`src/views/login.vue` L114；密钥对硬编码在 `src/utils/jsencrypt.js`）。

---

## 3. 全局路由守卫 — `src/permission.js`（全文 L1-77）

- `NProgress.configure({ showSpinner: false })`（L13）；beforeEach 开头 `NProgress.start()`，afterEach `NProgress.done()`（L74-76）。
- **白名单 `whiteList = ['/login', '/register']`**（L15），匹配用 `isPathMatch()`（`src/utils/validate.js` L7-16：支持 `*`→`[^/]*`、`**`→`.*`、`?`→`[^/]`，其余正则字符转义）。
- **有 token 分支**（L23-62），顺序严格：
  1. `to.meta.title` 存在则 `useSettingsStore().setTitle(title)`（设置页面标题，L24）。
  2. 读 `useLockStore().isLock`（L25）。
  3. 去 `/login` → 重定向 `{ path: '/' }`（L26-29）。
  4. 命中白名单 → 放行（L30-32）。
  5. `isLock && to.path !== '/lock'` → 强制跳 `/lock`（L33-36）。
  6. `!isLock && to.path === '/lock'` → 跳 `/`（L37-40）。
  7. `useUserStore().roles.length === 0`（首次进入/刷新后）（L41-61）：
     - `isRelogin.show = true` → `await getInfo()` → `isRelogin.show = false` → `await generateRoutes()`；
     - 返回的 accessRoutes 逐条 `router.addRoute(route)`（**跳过 path 是 http(s) 的外链路由**，`isHttp` 判断，L49-53）；
     - `return { ...to, replace: true }` 重新导航（L55）；
     - 异常：`await logOut()` + `ElMessage.error(err)` + 跳 `/`（L56-60）。
  8. 其余放行（L62）。
- **无 token 分支**（L63-71）：白名单放行；否则 `NProgress.done()` 并 `return '/login?redirect=' + to.fullPath`。

---

## 4. 路由表 — `src/router/index.js`

### 4.1 constantRoutes（L28-93）

| path | name | component | 备注 |
|---|---|---|---|
| `/redirect` | - | Layout | hidden；child `/redirect/:path(.*)` → `views/redirect/index.vue`（该组件用 `router.replace({'/'+path, query})` 实现刷新） |
| `/login` | - | `views/login` | hidden |
| `/register` | - | `views/register` | hidden |
| `/:pathMatch(.*)*` | - | `views/error/404` | hidden（**catch-all 404 注册在静态路由里**） |
| `/401` | - | `views/error/401` | hidden |
| `''`（根） | - | Layout，redirect `/index` | child `/index` name `Index`，meta `{title:'首页', icon:'dashboard', affix:true}`（affix = 固定页签） |
| `/lock` | - | `views/lock`（实为 `src/views/lock.vue`） | hidden，meta title「锁定屏幕」 |
| `/user` | - | Layout，hidden，redirect 'noredirect' | child `profile/:activeTab?` name `Profile`，meta `{title:'个人中心', icon:'user'}` |

路由模式 `createWebHistory()`；`scrollBehavior`：有 savedPosition 用之，否则回顶（L169-178）。

### 4.2 dynamicRoutes（L96-167）—— 全部 hidden、Layout 包裹、按 `permissions` 过滤

| path | 权限 | child | name | meta |
|---|---|---|---|---|
| `/system/user-auth` | `system:user:edit` | `role/:userId(\\d+)` → authRole | AuthRole | 分配角色，activeMenu `/system/user` |
| `/system/role-auth` | `system:role:edit` | `user/:roleId(\\d+)` → authUser | AuthUser | 分配用户，activeMenu `/system/role` |
| `/system/dict-data` | `system:dict:list` | `index/:dictId(\\d+)` → dict/data | Data | 字典数据，activeMenu `/system/dict` |
| `/monitor/job-log` | `monitor:job:list` | `index/:jobId(\\d+)` → job/log | JobLog | 调度日志，activeMenu `/monitor/job` |
| `/tool/gen-edit` | `tool:gen:edit` | `index/:tableId(\\d+)` → gen/editTable | GenEdit | 修改生成配置，activeMenu `/tool/gen` |

### 4.3 filterDynamicRoutes — `src/store/modules/permission.js` L100-114

有 `permissions` 字段 → `auth.hasPermiOr(route.permissions)`（任一满足即可，见 §6.2）；否则有 `roles` 字段 → `auth.hasRoleOr(route.roles)`。两者都不带的路由**直接丢弃**。过滤后的路由在 generateRoutes 内逐条 `router.addRoute`（L46）。

---

## 5. Pinia Store — `src/store/modules/`（入口 `src/store/index.js` 仅 `createPinia()`）

### 5.1 user（user.js）

- state（L13-21）：`token: getToken()`（初始即从 Cookie 读）、`id`、`name`、`nickName`、`avatar`、`roles: []`、`permissions: []`。
- `login(userInfo)`（L24-39）：username trim，调 `POST /login`（headers: `isToken:false, repeatSubmit:false`，见 api/login.js L11-20）→ `setToken(res.token)` 写 Cookie → `this.token = res.token` → **`useLockStore().unlockScreen()`**（登录即解锁）。
- `getInfo()`（L41-77）：`GET /getInfo`。返回消费字段：
  - `res.user`：`userId→id`、`userName→name`、`nickName`、`avatar`（avatar 非空且非 http(s) 时拼 `VITE_APP_BASE_API` 前缀；空则默认图 `@/assets/images/profile.jpg`，L45-48）；
  - `res.roles`：**非空数组才存；空数组时 `roles = ['ROLE_DEFAULT']`（permissions 不同步赋值，保留旧值/空）**（L49-54）；非空时 roles、permissions 都存；
  - `res.permissions`；
  - `res.pwdChrtype` 存 sessionStorage 键 **`pwrChrtype`**（L59）；
  - `res.isDefaultModifyPwd` true → 弹「您的密码还是初始密码，请修改密码！」确认框，确认跳 `Profile` 路由 `activeTab: 'resetPwd'`（L61-65）；
  - `!isDefaultModifyPwd && res.isPasswordExpired` → 弹「您的密码已过期，请尽快修改密码！」同样跳转（L67-71）。
- `logOut()`（L79-91）：`POST /logout` → 清空 token/roles/permissions + `removeToken()`。

### 5.2 permission（permission.js）

- state（L14-20）：`routes`（constant+动态，用于面包屑/页签 affix 提取）、`addRoutes`、`defaultRoutes`、`topbarRouters`、`sidebarRouters`。
- `generateRoutes()`（L35-54）：`GET /getRouters`（api/menu.js）→ 深拷贝三份后端菜单：
  - `sidebarRoutes = filterAsyncRouter(sdata)`（侧边栏用）；
  - `rewriteRoutes = filterAsyncRouter(rdata, false, true)`（**第三参 type=true**，实际挂到 router 的版本）；
  - `defaultRoutes = filterAsyncRouter(defaultData)`（topbar 用）;
  - `asyncRoutes = filterDynamicRoutes(dynamicRoutes)` 并逐条 addRoute；
  - `setRoutes(rewriteRoutes)`（routes = constantRoutes.concat）、`setSidebarRouters(constantRoutes.concat(sidebarRoutes))`、`setDefaultRoutes(sidebarRoutes)`、`setTopbarRoutes(defaultRoutes)`；
  - resolve(rewriteRoutes)。
- `filterAsyncRouter(asyncRouterMap, lastRouter, type)`（L59-84）——component 字符串 → 组件映射规则：
  - `'Layout'` → `@/layout`（布局壳）；`'ParentView'` → `@/components/ParentView`（其模板仅 `<router-view/>`，二级父占位）；`'InnerLink'` → `@/layout/components/InnerLink`（iframe 容器）；**其余当作 views 相对路径 → `loadView()`**。
  - 无 children 时删除 `children`、`redirect` 键（L78-81）。
  - `type=true` 时先对顶层 route.children 调 `filterChildren()`（L86-97）：把 ParentView 中间层拍平——子路径拼父路径（`lastRouter.path + '/' + el.path`），component 为 ParentView 且还有 children 的继续递归拍平。作用：让实际挂载的路由是扁平的，避免多层 ParentView 嵌套。
- `loadView(view)`（L116-125）：`modules = import.meta.glob('./../../views/**/*.vue')`（L9），遍历所有 key，取 `path.split('views/')[1].split('.vue')[0]` 与 view 串全等匹配，命中返回 `() => modules[path]()`（懒加载）。**找不到匹配返回 undefined**（对应菜单 component 配错的场景）。
- **没有 topNav 专用拆分逻辑在 store 里**：`topbarRouters` 就是 filterAsyncRouter 普通版结果，TopNav/TopBar 组件自行从 `sidebarRouters`/`topbarRouters` 派生（见 §9.4）。

### 5.3 dict（dict.js L1-58）

- state：`dict: []`（数组，元素 `{key, value}`；**纯内存，不持久化，刷新即失**）。
- `getDict(_key)` 线性查找返回 value；`setDict(_key, value)` push；`removeDict` splice；`cleanDict` 清空；`initDict` 空实现。
- 消费方 `src/utils/dict.js` `useDict(...types)`（L7-24）：先查 store 缓存，没有则 `GET /system/dict/data/type/{dictType}`，映射为 `{label: dictLabel, value: dictValue, elTagType: listClass, elTagClass: cssClass}` 后写缓存；返回 `toRefs`。挂为全局方法 + auto-import。

### 5.4 app（app.js L1-47）

- state：`sidebar: {opened（初始读 Cookie 'sidebarStatus'，无则 true）, withoutAnimation, hide: false}`、`device: 'desktop'`、`size: Cookies.get('size') || 'default'`。
- actions：`toggleSideBar`（翻转 opened，写 Cookie sidebarStatus=1/0）、`closeSideBar`（sidebarStatus=0）、`toggleDevice`、`setSize`（写 Cookie `size`）、`toggleSideBarHide`。

### 5.5 settings（settings.js L1-57）

- 默认值来自 `src/settings.js`，可被 localStorage 键 **`layout-setting`** 的 JSON 覆盖（L11）。
- state 字段与默认（settings.js L1-66）：
  - `title: ''`（运行时页面标题）
  - `theme: '#409EFF'`
  - `sideTheme: 'theme-dark'`
  - `showSettings: true`
  - `navType: 1`（1=纯左侧，2=混合，3=纯顶部）
  - `tagsView: true`、`tagsViewPersist: false`、`tagsIcon: false`、`tagsViewStyle: 'card'`（另一值 'chrome'）
  - `fixedHeader: true`、`sidebarLogo: true`、`dynamicTitle: false`
  - `footerVisible: false`、`footerContent: 'Copyright © 2018-2026 RuoYi. All Rights Reserved.'`
  - `isDark`（由 `useDark()`（vueuse）驱动，写 html 的 `dark` class；L6-7, L31）
- actions：`changeSetting({key,value})`；`setTitle(title)`（存 title + `useDynamicTitle()`：dynamicTitle 开时 `document.title = title + ' - ' + VITE_APP_TITLE`，否则 = VITE_APP_TITLE，`src/utils/dynamicTitle.js`）；`toggleTheme()`（翻转 isDark + toggleDark + nextTick 后 `handleThemeStyle(theme)`）。

### 5.6 tagsView（tagsView.js，store id 'tags-view'，L24-226）

- state：`visitedViews[]`、`cachedViews[]`（存 **路由 name 字符串**）、`iframeViews[]`。
- **cachedViews 机制**：`addCachedView` 只在 `!view.meta.noCache` 时 push `view.name`（L62-67）；AppMain 的 `<keep-alive :include="cachedViews">` 按组件名缓存（AppMain.vue L5）——因此页面组件必须通过 `unplugin-vue-setup-extend-plus` 支持 `<script setup name="X">` 声明名字。
- **iframeViews**：`meta.link` 存在的页面（内嵌外链）。`addIframeView` 按 path 去重，title 取 `meta.title || 'no-name'`（L37-44）。
- **visitedViews 持久化**：条件性持久化——`settingsStore.tagsViewPersist` 为 true 时，每次 visitedViews 变化（add/del/delLeft/delRight/delOthers）调用 `saveVisitedViews`：**过滤掉 `meta.affix` 的固定页签**后存 localStorage 键 **`tags-view-visited`**（`cache.local.setJSON`，字段裁剪为 path/fullPath/name/title/query/meta，L4-22）。`loadPersistedViews()` 在 TagsView initTags 时恢复（L217-222）。`delAllVisitedViews` 清除该键（L145-153）。
- 其余动作：`addView`（visited+cached）、`addAffixView`（unshift 固定页签）、`delView/delVisitedView/delCachedView`、`delOthersViews`（保留 affix + 当前，iframeViews 只留当前）、`delAllViews`（visited 只留 affix，iframeViews 清空）、`updateVisitedView`、`delRightTags/delLeftTags`（保留 affix，同步清理 cachedViews 与 iframeViews 中的 meta.link 项）。
- 判重均按 `path`；title 缺省 'no-name'。

### 5.7 lock（lock.js，L1-28）

- state：`isLock`（localStorage 键 **`screen-lock`**，'true'/'false' 字符串 JSON.parse）、`lockPath`（localStorage 键 **`screen-lock-path`**，默认 '/index'）。
- `lockScreen(currentPath)`：记录当前路径到 lockPath，isLock=true。`unlockScreen()`：isLock=false、lockPath 复位 '/index'。
- 锁屏页 `src/views/lock.vue`：粒子背景 + 时钟；解锁调 `POST /unlockscreen`（api/login.js L43-49，body `{password}`），成功后 `router.replace(lockStore.lockPath)` 回原页面；失败显示后端 msg + 抖动动画；「退出重新登录」= unlockScreen + logOut + push '/login'。

---

## 6. 权限指令与工具

### 6.1 v-hasPermi — `src/directive/permission/hasPermi.js`（L7-27）

- `mounted` 钩子里判断（**只在 mounted 执行一次，不做响应式更新**）。
- 通配权限 **`*:*:*`**：用户 permissions 数组里含 `*:*:*` 即全部通过；否则要求 binding.value 数组与用户 permissions 有交集（`permissionFlag.includes(permission)`）。
- 不通过 → `el.parentNode.removeChild(el)`（**DOM 移除，非 v-if 隐藏**）。
- value 非非空数组 → throw `请设置操作权限标签值`。

### 6.2 v-hasRole — `src/directive/permission/hasRole.js`（L7-27）

- 超管角色字符串 **`admin`**：用户 roles 含 'admin' 即全部通过；否则与 value 数组求交集。删除方式、异常文案同上（`请设置角色权限标签值`）。
- 注意 `ROLE_DEFAULT` 本身**没有**特殊逻辑——它只是空角色用户的一个普通角色值，指令/脚本层不会通配它。

### 6.3 脚本判断 — `src/plugins/auth.js`（挂 `$auth`，L27-60）

`authPermission`（`*:*:*` 或相等）、`authRole`（'admin' 或相等）为底，导出：`hasPermi(p)`、`hasPermiOr(arr)`（some）、`hasPermiAnd(arr)`（every）、`hasRole(r)`、`hasRoleOr(arr)`、`hasRoleAnd(arr)`。`filterDynamicRoutes` 用的就是 hasPermiOr/hasRoleOr。

### 6.4 非指令版 — `src/utils/permission.js`：`checkPermi(value)` / `checkRole(value)`，逻辑同指令，value 非法时 console.error 并返回 false。

### 6.5 v-copyText — `src/directive/common/copyText.js`

`v-copyText="text"` 点击复制（textarea + execCommand('copy') 方案）；`v-copyText:callback="fn"` 注册回调。指令注册入口 `src/directive/index.js`：`v-hasRole`、`v-hasPermi`、`v-copyText`。

---

## 7. 插件 — `src/plugins/`（index.js 挂到 `app.config.globalProperties`：`$tab` `$auth` `$cache` `$modal` `$download`）

### 7.1 $modal（modal.js，基于 Element Plus）

- `msg/content→ElMessage.info`；`msgError/msgSuccess/msgWarning` → ElMessage error/success/warning。
- `alert(content)` / `alertError/alertSuccess/alertWarning` → `ElMessageBox.alert(content, "系统提示", {type})`。
- `notify*` 四个 → ElNotification info/error/success/warning。
- `confirm(content)` → `ElMessageBox.confirm(content, "系统提示", {confirmButtonText:'确定', cancelButtonText:'取消', type:'warning'})` 返回 Promise。
- `prompt(content)` → ElMessageBox.prompt 同款按钮/type。
- `loading(content)` / `closeLoading()` → ElLoading.service（lock:true，黑底 0.7）。

### 7.2 $tab（tab.js，依赖 tagsView store + router）

- `refreshPage(obj?)`：当前路径以 `/redirect/` 开头直接 resolve；obj 缺省时从 matched 找出组件名（排除 Layout/ParentView）构造 `{name, path, query}`；`delCachedView` 后 `router.replace('/redirect' + path + query)` 借 redirect 页强制刷新。
- `closeOpenPage(obj?)`：删当前 view 后可 push 新页。
- `closePage(obj?)`：obj 缺省 = 关当前并跳 visitedViews 最后一个（无则 `/`）；带 obj 则只删指定。
- `closeAllPage()` / `closeLeftPage(obj?)` / `closeRightPage(obj?)` / `closeOtherPage(obj?)`（obj 缺省用当前路由）。
- `openPage(title, url, params)`：addView({path:url, meta:{title}}) + `router.push({path, query:params})`。
- `updatePage(obj)` → updateVisitedView。

### 7.3 $auth：见 §6.3。

### 7.4 $cache（cache.js L70-79）：`session` / `local` 两个对象，各有 `set/get/setJSON/getJSON/remove`（JSON 版即 stringify/parse 包装；键值 null 检查）。防重复提交用的 `sessionObj`、tagsView 持久化用的 `tags-view-visited` 都经它。

### 7.5 $download：见 §1.6。

---

## 8. 应用入口 — `src/main.js`（L1-85）

- Element Plus 全量引入：`import ElementPlus` + `element-plus/dist/index.css` + **`element-plus/theme-chalk/dark/css-vars.css`**（暗色变量，L6-7）+ 中文 locale（zh-cn）；实例化时 `size: Cookies.get('size') || 'default'`（L78-82）。
- 全局样式 `@/assets/styles/index.scss`；`virtual:svg-icons-register`（svg 雪碧图，见 §10）。
- **全局方法** `app.config.globalProperties`（L50-58）：`useDict`、`download`（utils/request 的下载函数）、`parseTime`、`resetForm`、`handleTree`、`addDateRange`、`getConfigKey`（api/system/config）、`selectDictLabel`、`selectDictLabels`（均出自 `src/utils/ruoyi.js`）。
- **全局组件**（L61-67）：`DictTag`、`Pagination`、`FileUpload`、`ImageUpload`、`ImagePreview`、`RightToolbar`、`Editor`；另有 `svg-icon`（SvgIcon，L73）。
- `app.use(router)`、`app.use(store)`（pinia）、`app.use(plugins)`（五个 $ 插件）、`app.use(elementIcons)`（`@element-plus/icons-vue` 全部图标注册为全局组件，`src/components/SvgIcon/svgicon.js`）、`directive(app)` 注册三个指令。
- `import './permission'` 副作用式挂路由守卫（L26）。
- `src/App.vue`：模板仅 `<router-view/>`；onMounted 时 `handleThemeStyle(useSettingsStore().theme)` 初始化主题色。

---

## 9. 布局与设置 — `src/settings.js` + `src/layout/`

### 9.1 默认配置清单（settings.js，详见 §5.5）——title(VITE_APP_TITLE) / sideTheme 'theme-dark' / showSettings true / navType 1 / tagsView true / tagsViewPersist false / tagsIcon false / tagsViewStyle 'card' / fixedHeader true / sidebarLogo true / dynamicTitle false / footerVisible false / footerContent 'Copyright © 2018-2026 RuoYi. All Rights Reserved.'

### 9.2 布局骨架 — `src/layout/index.vue`（L1-63）

- 结构：`.app-wrapper`（CSS 变量 `--current-color` 等由 theme 派生）→ `sidebar`（`!sidebar.hide` 时渲染）+ `.main-container`（fixed-header 包 navbar + tags-view；`app-main` 内容区；`settings` 抽屉）。
- **响应式**：`useWindowSize`，宽 < 992px 切 `device='mobile'` 并收起侧边栏（L37-53）；移动端 opened 时点遮罩关闭。
- `fixedHeader` class 控制头部 fixed；`hideSidebar/openSidebar/mobile/withoutAnimation` class 由 app store 派生（L30-35）。
- 侧边栏宽 200px（`variables.module.scss` `$base-sidebar-width`），收起后 54px。

### 9.3 三种导航模式（navType）— Navbar.vue（L1-65）

- **1 纯左侧**：hamburger + breadcrumb；侧边栏显示完整 `sidebarRouters`。
- **2 混合**：hamburger + TopNav（顶部一级菜单）；点击一级菜单 → `activeRoutes(key)` 把该一级的 children 设为 `permissionStore.setSidebarRouters(...)` 联动左侧（TopNav L142-158）；无 children 的项内部打开且 `toggleSideBarHide(true)`。`hideList = ['/index', '/user/profile']` 不参与一级高亮拆分。
- **3 纯顶部**：隐藏 hamburger，Navbar 里渲染 Logo + TopBar（TopBar 也从 `sidebarRouters` 取前 N 项渲染水平菜单，溢出进「更多菜单」，N = max(1, 窗口宽/3/85)）。
- Settings 抽屉切 navType 时（Settings/index.vue L170-186）：1 → opened=true, hide=false；2 → opened=true；3 → opened=false, hide=true；1/3 时把 sidebarRouters 复位为 `permissionStore.defaultRoutes`。

### 9.4 暗色主题实现

- 开关：settings store `isDark` 经 vueuse `useDark()`（写/读 **html 元素的 `dark` class**，默认跟随系统偏好）。
- Element Plus 侧：`element-plus/theme-chalk/dark/css-vars.css` 已在 main.js 引入，`html.dark` 自动生效。
- 项目自定义侧：`src/assets/styles/variables.module.scss` 末尾定义 `html.dark` 下大量 CSS 变量（`--el-bg-color:#141414`、`--sidebar-bg`、`--navbar-bg`、`--tags-*` 等）及组件覆盖（侧边栏/标签栏/表格/树/下拉等）。
- 主题色：`src/utils/theme.js` `handleThemeStyle(theme)`（L1-12）设置 `--el-color-primary` 并生成 light-1..9 / dark-1..9 十八档色阶（hex 混白/混黑）；暗色下主色先与 `#2d3036` 按 0.34 混合柔化（`softenPrimaryForDark`）。
- Navbar 的主题切换按钮（Navbar.vue L135-173）：用 `document.startViewTransition` 做以点击点为圆心的扩散动画（650ms，circle clip-path），reduced-motion 或不支持时直接 toggle。

### 9.5 设置抽屉 — `src/layout/components/Settings/index.vue`（L1-224）

可配置项：导航模式（1/2/3 三张示意图）、侧边栏主题（dark/light 两张图）、主题颜色（el-color-picker，预置色板 `["#409EFF","#ff4500","#ff8c00","#ffd700","#90ee90","#00ced1","#1e90ff","#c71585"]`）、开启页签、持久化标签页（页签关闭时禁用）、显示页签图标、标签页样式（卡片/谷歌 radio）、固定 Header、显示 Logo、动态标题、底部版权。
- 「保存配置」→ `$modal.loading` + 把 11 个字段 JSON 存 localStorage `layout-setting`（L188-208）；tagsViewPersist 关闭时顺带删 `tags-view-visited`。
- 「重置配置」→ 删 `layout-setting` + 删 `tags-view-visited` + `location.reload()`（L210-215）。
- 抽屉经 `defineExpose({openSetting})` 暴露，入口：Navbar 头像下拉「布局设置」→ emit setLayout → layout index 的 `settingRef.openSetting()`。

### 9.6 锁屏入口

Navbar 头像下拉「锁定屏幕」→ `lockStore.lockScreen(route.fullPath)` + `router.push('/lock')`（Navbar.vue L129-133）。`/lock` 是静态路由（hidden），守卫强制：isLock 时非 /lock 一律跳 /lock（§3）。

### 9.7 iframe 页面渲染

- 菜单 `meta.link`（http/https 外链或内网地址）→ 路由 component 为 InnerLink 的项由 IframeToggle 呈现：AppMain 中 `<router-view>` 对 `route.meta.link` 的组件渲染 `v-if="!route.meta.link"` 跳过，改由 `<iframe-toggle>` 按 `tagsViewStore.iframeViews` 渲染常驻 iframe，`v-show="route.path === item.path"` 切换（AppMain.vue L6, L10；IframeToggle/index.vue）。
- InnerLink（layout/components/InnerLink/index.vue）：全宽 iframe，高 = `document.documentElement.clientHeight - 94.5` px，带 loading（「正在加载页面，请稍候！」），onload 后关。
- iframe src = `meta.link` + query 串拼接（IframeToggle `iframeUrl()`）。
- AppMain 还包含 `<copyright>`（footer，`footerVisible` 控制的固定底条，36px 高）。

### 9.8 其他布局细节

- TagsView（TagsView/index.vue）：card / chrome 两种样式；滚动箭头 + 下拉操作菜单（关闭当前/其他/左侧/右侧/全部、全屏显示）+ 右键菜单（刷新/关闭当前/关闭其他/关闭左/右/全部）+ 刷新按钮；中键点击关页签；affix 页签不可关；Esc 退出全屏；全屏模式隐藏 navbar 与 sidebar。
- SidebarItem 递归渲染；只有一个可见子路由时直接渲染子级；`meta.link` 外链经 Link.vue 渲染为 `<a target="_blank" rel="noopener">`；`resolvePath` 处理外链、query（meta.query 为 JSON 字符串需 parse）、`getNormalPath` 双斜杠归一。
- HeaderNotice：通知公告 popover（`listNoticeTop`/`markNoticeRead`/`markNoticeReadAll`，未读徽标）。
- 滚动条样式全局 6px 自定义（AppMain.vue L110-123）。

---

## 10. Vite 构建 — `vite.config.js` + `vite/plugins/`

### 10.1 vite.config.js

- `base: '/'`（两种环境同）；alias：`@` → `./src`，`~` → 项目根（L20-24）；resolve.extensions 带 `.vue`（L26）。
- **dev server（L44-61）：`port: 80`，`host: true`，`open: true`；代理：`'/dev-api'` → `http://localhost:8080`（`baseUrl` 常量 L5），`changeOrigin: true`，rewrite 去掉 `/dev-api` 前缀**；另有 `^/v3/api-docs/(.*)` → 同目标（springdoc，不 rewrite）。
- build（L29-41）：sourcemap 构建时 false / dev 'inline'；outDir `dist`；chunkSizeWarningLimit 2000；产物路径 `static/js/[name]-[hash].js` 与 `static/[ext]/[name]-[hash].[ext]`。
- postcss 内联插件：移除 `@charset` at-rule（L62-77）。

### 10.2 插件装配 — `vite/plugins/index.js`（L8-15）

`@vitejs/plugin-vue` → auto-import → setup-extend → svg-icon → **仅 build 时** compression。

- **svg-icon**（svg-icon.js L4-10）：`vite-plugin-svg-icons`，`iconDirs: [src/assets/icons/svg]`（约 90 个 svg），`symbolId: 'icon-[dir]-[name]'`，`svgoOptions: isBuild`。运行时 `virtual:svg-icons-register` 注入；`<svg-icon icon-class="user">` 组件内 `<use xlink:href="#icon-user">`——icon-class 直接拼 `#icon-` 前缀（SvgIcon/index.vue L33）。React 重写可用等价雪碧图方案或图标组件库。
- **auto-import**（auto-import.js L3-16）：`imports: ['vue', 'vue-router', 'pinia', {'@/utils/dict': ['useDict'], '@/utils/ruoyi': ['selectDictLabel']}]`，`dts: false`。即所有 vue/vue-router/pinia API 与 useDict、selectDictLabel 全站免 import——重写 React 时需保证等价的自动注入或显式导入约定。
- **setup-extend**（setup-extend.js）：`unplugin-vue-setup-extend-plus`——支持 `<script setup name="X">`，keep-alive 缓存依赖它。
- **compression**（compression.js）：`VITE_BUILD_COMPRESS` 含 'gzip' → `.gz`（deleteOriginFile:false）；含 'brotli' → `.br`。

---

## 11. 环境变量

### `.env.development`

| 变量 | 值 | 含义 |
|---|---|---|
| `VITE_APP_TITLE` | 若依管理系统 | 页面标题（document.title / Logo 文案） |
| `VITE_APP_ENV` | 'development' | 环境标识（vite.config 用于 base 判断） |
| `VITE_APP_BASE_API` | `/dev-api` | axios baseURL，走 dev server 代理到 localhost:8080 |

### `.env.production`

| 变量 | 值 |
|---|---|
| `VITE_APP_TITLE` | 若依管理系统 |
| `VITE_APP_ENV` | 'production' |
| `VITE_APP_BASE_API` | `/prod-api`（生产由网关/nginx 转发，前端不再代理） |
| `VITE_BUILD_COMPRESS` | gzip |

### `.env.staging`

`VITE_APP_ENV='staging'`、`VITE_APP_BASE_API='/stage-api'`、`VITE_BUILD_COMPRESS=gzip`，标题同上。

变量消费点：request.js baseURL、plugins/download.js、user.js avatar 前缀、Logo 标题、dynamicTitle。

---

## 12. 重写时的关键行为差异点提示（易错清单）

1. 后端契约是「HTTP 恒 200 + body.code 分支」；401 弹**确认框**（可留在页面），500/601 用 Message，其他非 200 用 Notification——三种 UI 形态不同。
2. `download()` 是 **POST + form-urlencoded**，不是 GET；错误藏在 JSON blob 里需二次解析。
3. 防重复提交状态在 **sessionStorage**（tab 级隔离），键 `sessionObj`，默认 1000ms，body≥5MB 跳过。
4. token 是**会话级 Cookie**（非 localStorage、非过期 cookie），键名 `Admin-Token` 必须保持（后端可能依赖）。
5. `getInfo` 的 roles 为空 → `['ROLE_DEFAULT']`；`*:*:*` 与角色 `admin` 是通配；权限判断指令是**移除 DOM**。
6. 动态路由 component 三占位（Layout/ParentView/InnerLink）+ views 路径映射；type=true 分支会拍平 ParentView 链。
7. keep-alive（React 中为状态保持方案）按**组件 name** 匹配 cachedViews，`meta.noCache` 为 true 不缓存。
8. visitedViews 持久化受 `tagsViewPersist` 开关控制且排除 affix 页签，键 `tags-view-visited`。
9. 锁屏由路由守卫硬性劫持（非仅 UI 覆盖），解锁接口 `POST /unlockscreen`。
10. dev 端口 80、代理目标 localhost:8080、`/dev-api` 前缀剥离——代理契约必须一致。
