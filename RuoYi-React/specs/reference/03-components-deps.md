# RuoYi-Vue3 组件库与第三方依赖调查报告

> 基线版本 RuoYi v3.9.2，Vue 3.5 + Element Plus 2.13 + Vite 6。所有路径省略前缀 `RuoYi-Vue3/`。
> 生成于 2026-09-27，由 AI 对源码全量调查产出；作为 RuoYi-React 组件对位与依赖选型基线。

---

## 1. 自研组件清单（src/components，以实际存在为准）

### 全局注册组件（main.js:61-67 注册为全局，模板直接用 kebab-case）

**Pagination**（src\components\Pagination\index.vue）

- Props：`total`(必填)、`page`(默认1)、`limit`(默认20)、`pageSizes`(默认[10,20,30,50])、`pagerCount`(视口<992px 为5否则7)、`layout`(默认'total, sizes, prev, pager, next, jumper')、`background`(默认true)、`autoScroll`(默认true)、`hidden`(默认false)
- 事件：`pagination`({page,limit})、`update:page`、`update:limit`（支持 v-model:page / v-model:limit）
- 行为：包 el-pagination；改 size 时若当前页×新size 超过 total 则重置回第1页；每次变更后 autoScroll 时 `scrollTo(0, 800)` 平滑回顶（src\utils\scroll-to.js）
- 用法示例：src\views\system\user\index.vue:89、tool\gen\index.vue:122 等几乎所有列表页

**RightToolbar**（src\components\RightToolbar\index.vue）

- Props：`showSearch`(默认true，v-model:showSearch)、`columns`(数组或对象格式，项含{label,key,visible})、`search`(默认true，是否显示检索切换钮)、`showColumnsType`('checkbox'默认 | 'transfer'穿梭框)、`gutter`(默认10，右外边距)、`storageKey`(传则用 localStorage 持久化列显隐状态，经 cache.local.getJSON/setJSON)
- 事件：`update:showSearch`、`queryTable`(刷新按钮)
- 行为：三个圆钮——显隐搜索（**自动向上查找祖先节点里的 .el-form 并做 max-height/opacity 动画折叠**，找不到表单则仅发 update:showSearch）、刷新、显隐列（checkbox 下拉含全选/半选，transfer 打开 el-transfer 弹窗）。列状态变化即 saveStorage
- 用法：几乎所有列表页右上角，如 src\views\system\user\index.vue:43（带 storageKey="xxxxxxxx"）

**DictTag**（src\components\DictTag\index.vue）

- Props：`options`(字典数组 [{label,value,elTagType,elTagClass}]，由 useDict 产出)、`value`(Number|String|Array)、`showValue`(默认true，无匹配时显示原始值)、`separator`(默认',')
- 行为：value 按分隔符拆分（数组/数字/布尔也归一化为字符串数组），匹配到的渲染 el-tag（type=elTagType、class=elTagClass）；elTagType 为 'default'/'' 且无 class 时渲染纯 span；未匹配的值收集进 unmatchArray 并在 showValue 时追加显示
- 用法：src\views\system\notice\index.vue:81、monitor\operlog\index.vue:116 等

**Editor**（富文本，详见第4节）

**FileUpload**（src\components\FileUpload\index.vue）⚠️ 目前**没有业务页面实际使用**（仅全局注册备用）

- Props：`modelValue`(String 逗号分隔url | Object | Array)、`action`(默认"/common/upload")、`data`、`limit`(默认5)、`fileSize`(MB，默认5)、`fileType`(默认["doc","docx","xls","xlsx","ppt","pptx","txt","pdf"])、`isShowTip`(默认true)、`disabled`(默认false)、`drag`(默认true 拖动排序)
- 上传：el-upload multipart 到 `VITE_APP_BASE_API + action`，header 带 Bearer token；成功取 `res.fileName`，全部传完后 modelValue = 各 url 逗号拼接字符串
- 回显：watch modelValue → 字符串按逗号拆成 {name,url}；文件名列表项链接 `baseUrl + file.url` 新窗口打开；可删除
- 校验：扩展名白名单、文件名不能含英文逗号、大小限制；上传中 $modal.loading
- 排序：Sortable（sortablejs）拖拽重排后重新 emit

**ImageUpload**（src\components\ImageUpload\index.vue）⚠️ 目前**没有业务页面实际使用**（仅全局注册备用）

- Props：`modelValue`、`action`(默认"/common/upload")、`data`、`limit`(默认5)、`fileSize`(MB默认5)、`fileType`(默认["png","jpg","jpeg"])、`isShowTip`、`disabled`、`drag`
- 上传：picture-card 多选；成功取 `res.fileName`；modelValue 存储时**剥掉 baseUrl 前缀**（listToString 里 `url.replace(baseUrl, "")`），逗号拼接
- 回显：字符串拆分时若不含 baseUrl 且非外链（isExternal）则补 `baseUrl + item`
- 其余同 FileUpload（loading、超限提示、Sortable 拖拽、大图 el-dialog 预览）

**ImagePreview**（src\components\ImagePreview\index.vue）⚠️ 目前**没有业务页面实际使用**

- Props：`src`(String，支持逗号分隔多图)、`width`、`height`（Number 加 px，String 原样）
- 行为：el-image fit=cover，preview-src-list 支持多图大图预览（preview-teleported）；每段路径外链原样、否则拼 baseUrl；hover 放大 1.2、错误槽显示图标

### 按需引入组件

**SvgIcon**（src\components\SvgIcon\index.vue + svgicon.js）

- Props：`iconClass`(必填)、`className`、`color`
- 行为：`<use xlink:href="#icon-{iconClass}">`；svg 精灵由 vite-plugin-svg-icons 从 src\assets\icons\svg\（约90个）生成，`main.js:22` `import 'virtual:svg-icons-register'`；全局注册名 `svg-icon`。svgicon.js 把 @element-plus/icons-vue 全量注册为独立组件（Element 图标另一套，标签直接用组件名如 `<Search/>`）
- IconSelect 的图标清单来自 src\components\IconSelect\requireIcons.js（import.meta.glob 枚举同目录的 svg 文件名）

**IconSelect**（src\components\IconSelect\index.vue）

- Props：`activeIcon`(当前选中图标名，高亮)
- 事件：`selected(name)`（选中后 document.body.click() 关闭外层弹出）
- expose：`reset()` 清空搜索
- 行为：输入框按名称过滤本地图标集，三列网格展示 svg+名称
- 用法：src\views\system\menu\index.vue:148（菜单图标选择）

**HeaderSearch**（src\components\HeaderSearch\index.vue）

- 行为：点击放大镜弹 el-dialog，输入即搜；数据源 = permission store 的 defaultRoutes 递归扁平化（path/title 数组/icon/query）；**fuse.js** 模糊（threshold 0.2, keys: title 0.7 + path 0.3）结果与 path 子串匹配合并去重；键盘 ↑↓ 切换、Enter 跳转、Esc 关闭；命中 title/path 用 v-html 高亮；http(s) 菜单 window.open 新窗口，其余 router.push（query 为 JSON 字符串则 parse）；底部快捷键说明条
- 用法：src\layout\components\Navbar.vue:13

**Screenfull**（src\components\Screenfull\index.vue）

- 无 props；@vueuse/core `useFullscreen()` 提供 isFullscreen/toggle；图标 fullscreen/exit-fullscreen 切换
- 用法：Navbar.vue:23

**SizeSelect**（src\components\SizeSelect\index.vue）

- 无 props；下拉 large/default/small（较大/默认/稍小）；选择后 appStore.setSize → 写 js-cookie 'size' → `window.location.reload()` 整页刷新；main.js:81 用 `Cookies.get('size') || 'default'` 初始化 Element Plus 全局 size
- 用法：Navbar.vue:33

**Breadcrumb**（src\components\Breadcrumb\index.vue）

- 无 props；route.path 斜杠数>2 时（多级菜单）按 permission store defaultRoutes 逐段匹配 getMatched，否则用 route.matched 过滤有 meta.title 的项；非首页前置「首页」(path /index)；meta.breadcrumb===false 的不显示；末级/redirect=noRedirect 渲染纯文本，其余可点击跳转；/redirect/ 开头的路由不刷新面包屑
- 用法：Navbar.vue:4（navType==1 时）

**Hamburger**（src\components\Hamburger\index.vue）

- Props：`isActive`(Boolean)；事件 `toggleClick`；点击旋转 180°；Navbar.vue:3 用来开合侧栏

**TopNav**（src\layout\components\TopNav\index.vue，注意在 layout 不在 src/components）

- 顶部横向菜单（navType==2 混合模式）：数据源 permissionStore.topbarRouters；按窗口宽/3/85px 计算可视数量，超出的折叠进「更多菜单」子菜单；点无子路由项直接跳转并隐藏侧栏，点有子路由项则 activeRoutes 把 children 写入 sidebarRouters 联动左侧；http(s) 新窗口
- 用法：Navbar.vue:5

**TagsView**（src\layout\components\TagsView\index.vue，同在 layout）

- 页签栏：visitedViews 来自 tagsView store；affix（meta.affix）页签不可关；支持卡片(card，激活底色=主题色)/chrome 两种样式（settings.tagsViewStyle）、页签图标（tagsIcon）、持久化（tagsViewPersist 时 loadPersistedViews）；左右滚动箭头+ScrollPane；中键点击关闭；右键上下文菜单（刷新/关闭当前/关闭其他/关闭左侧/关闭右侧/全部关闭）；下拉菜单同款+全屏显示（隐藏 navbar/sidebar，Esc 退出）；刷新走 proxy.$tab.refreshPage（keep-alive 移除重载，src\plugins\tab.js）；iframe 页 delIframeView
- 缓存机制：页面组件须 `<script setup name="X">`（unplugin-vue-setup-extend-plus），keep-alive 按 tagsView.cachedViews 匹配（AppMain.vue）
- 用法：src\layout\index.vue:8

**ThemePicker**：**不存在**。本版无 ThemePicker 组件；主题=顶部明暗切换钮（Navbar.vue:26 toggleTheme，@vueuse/core useDark + document.startViewTransition 圆形扩散动画）+ settings store theme 色（src\utils\theme.js handleThemeStyle），布局设置面板在 src\layout\components\Settings\index.vue

**RightPanel**：src\components 下**不存在**；仅有表单构建器的属性面板 src\views\tool\build\RightPanel.vue（局部组件，非通用）

**Crontab**（src\components\Crontab\index.vue + second/min/hour/day/month/week/year/result 8个子文件）

- Props：`expression`(String，传入则反解析填充)、`hideComponent`(Array，要隐藏的域)
- 事件：`hide`(取消)、`fill`(确定回填表达式)
- 行为：el-tabs 七个页签（秒/分/时/日/月/周/年）各自单选规则组；底部表格实时拼接 cron 串（默认 `* * * * * ?`）；CrontabResult 根据表达式解析显示下次执行时间说明；重置/取消/确定
- 用法：src\views\monitor\job\index.vue:233（定时任务 cron 生成器弹窗）

**TreePanel**（src\components\TreePanel\index.vue）

- Props：`treeData`、`title`(默认'树形结构')、`titleIcon`、`showSearch`(默认true)、`searchPlaceholder`、`defaultCollapsed`、`treeProps`(默认{children,label})、`nodeKey`(默认'id')、`expandOnClickNode`(默认false)、`showCheckbox`、`checkStrictly`、`defaultExpandAll`、`defaultExpandedKeys`、`defaultWidth`(220)/`collapsedWidth`(20)/`minWidth`(180)/`maxWidth`(400)、`storageKey`(默认'tree-sidebar-width'，localStorage 存宽度)、`enableStorage`、`filterMethod`
- 事件：`collapsed-change`、`expanded-all-change`、`refresh`、`node-click`、`check`、`node-expand`、`node-collapse`、`search`
- 行为：可折叠可拖拽调宽的树侧栏（rAF 节流、宽度持久化）；头部展开全部/收起全部、刷新、自定义 actions 插槽；搜索走 el-tree filter；节点默认插槽可覆写；expose 一整套 tree 方法（setCurrentKey/getCheckedKeys/setWidth 等）
- 用法：src\views\system\user\index.vue:3（部门树侧栏）

**ExcelImportDialog**（src\components\ExcelImportDialog\index.vue）

- Props：`title`(默认'数据导入')、`width`('400px')、`action`(**必传**，上传接口)、`templateAction`(传则显示"下载模板"链接)、`templateFileName`(默认'template')、`updateSupportLabel`(默认'是否更新已经存在的数据')
- 事件：`success`
- expose：`open()`
- 行为：拖拽上传 el-upload（limit 1、accept .xlsx/.xls、auto-upload=false，点确定才 submit）；上传 URL = `VITE_APP_BASE_API + action + '?updateSupport=0|1'`，Bearer 头；成功后 $alert 弹 response.msg（dangerouslyUseHTMLString）并 emit success；模板下载走全局 proxy.download（POST blob）
- 用法：src\views\system\user\index.vue:183（action="/system/user/importData" template-action="/system/user/importTemplate"）

**iFrame**（src\components\iFrame\index.vue）

- Props：`src`(必填)；全高 iframe（clientHeight-94.5），300ms 后取消 loading，窗口 resize 跟随
- 用法：monitor\druid\index.vue:9（/druid/login.html）、tool\swagger\index.vue:8（/swagger-ui/index.html）、layout IframeToggle

**ParentView**（src\components\ParentView\index.vue）

- 仅 `<router-view />`；不是页面级组件而是**路由占位**——permission store（src\store\modules\permission.js:68）把后端菜单 component 字符串 'ParentView' 映射为该组件，用于多级菜单中间层

**RuoYi/Git、RuoYi/Doc**（src\components\RuoYi\Git\index.vue、Doc\index.vue）

- 无 props；点击 svg-icon 新窗口打开 gitee 仓库 / doc.ruoyi.vip；Navbar.vue:16-21 使用

---

## 2. 第三方库使用映射（package.json 全部运行时依赖）

| 依赖 | 版本 | 使用文件 | 用途 |
|---|---|---|---|
| **vue** | 3.5.26 | 全部；入口 src\main.js | 框架本体。注意 vite\plugins\auto-import.js 配了 unplugin-auto-import（vue/vue-router/pinia 全 API + useDict/selectDictLabel），代码里 ref/computed 等不写 import |
| **vue-router** | 4.6.4 | src\router\index.js（createRouter+createWebHistory）；大量视图 useRouter/useRoute；src\permission.js 路由守卫 | 路由；动态路由来自 GET /getRouters → permission store filterAsyncRouter |
| **pinia** | 3.0.4 | src\store\index.js（createPinia）；7 个 store：user/permission/dict/app/settings/tagsView/lock（src\store\modules\*.js） | 状态管理 |
| **element-plus** | 2.13.1 | src\main.js:5-8（全量引入 + index.css + dark css-vars + zh-cn locale + 全局 size cookie） | UI 组件库，全站 |
| **@element-plus/icons-vue** | 2.3.2 | src\components\SvgIcon\svgicon.js:1（全量 app.component 注册）；src\views\system\notice\ReadUsers.vue:49（Search）；src\views\tool\build\IconsDialog.vue:24 | Element 图标第二套 |
| **axios** | 1.13.2 | src\utils\request.js:1（全站唯一实例：baseURL=VITE_APP_BASE_API、Bearer 头、按 body code 分支 401/500/601、防重复提交 sessionStorage、blob 下载）；src\plugins\download.js:1（下载工具）；src\components\Editor\index.vue:30（粘贴图片直传） | HTTP |
| **js-cookie** | 3.0.5 | src\utils\auth.js（Admin-Token 存取删）；src\store\modules\app.js（sidebarStatus、size）；src\main.js:81（读 size）；src\views\login.vue:113-120（记住我：username/password(RSA加密)/rememberMe 三 cookie，30天） | Cookie |
| **nprogress** | 0.2.0 | src\permission.js:3-4,13（showSpinner:false，路由切换 start/done） | 顶部进度条，仅此一处 |
| **jsencrypt** | 3.3.2 | src\utils\jsencrypt.js（**硬编码公私钥对**，导出 encrypt/decrypt）；src\views\login.vue:70,114,159（记住密码时 cookie 存 RSA 密文，回显解密） | RSA 加密，仅登录记住我用 |
| **echarts** | 5.6.0 | **仅 src\views\monitor\cache\index.vue:69-127**（Redis 缓存监控：命令统计 roseType 饼图 + 内存峰值仪表盘，`echarts.init(el,"macarons")`——注意 macarons 主题未注册实际回退默认主题；window resize 时 resize 两个实例） | 图表。首页 index.vue 里的 echarts 字样只是更新日志文本 |
| **@vueup/vue-quill** | 1.2.0（overrides 锁 quill 2.0.2） | **仅 src\components\Editor\index.vue:31-32**（QuillEditor + snow css） | 富文本，详见第4节 |
| **vue-cropper** | 1.1.1 | **仅 src\views\system\user\profile\userAvatar.vue:7,62-63** | 头像裁剪，详见第5节 |
| **vuedraggable** | 4.1.0 | **仅 src\views\tool\build\**（表单构建器）：index.vue:99,14-82（左侧组件面板拖入画布/画布内排序）、RightPanel.vue:467,218（选项列表拖拽排序）、DraggableItem.vue:28,9（嵌套布局子项拖拽）；均 import "vuedraggable/dist/vuedraggable.common" | **不用于代码生成预览、不用于 gen 页**，详见第6节 |
| **fuse.js** | 7.1.0 | **仅 src\components\HeaderSearch\index.vue:80,139-153** | 菜单模糊搜索 |
| **js-beautify** | 1.15.4 | **仅 src\views\tool\build\index.vue:101,276-282**（generateCode：html+script+css 拼接后 `beautifier.html(..., beautifierConf.html)` 美化，配置在 src\utils\index.js:336） | **不用于 gen 预览页**（gen 预览是 `<pre>{{value}}</pre>` 原样显示，src\views\tool\gen\index.vue:139） |
| **clipboard** | 2.0.11 | **仅 src\views\tool\build\index.vue:100,298-313**（new ClipboardJS('#copyNode',{text:返回生成代码}) 复制生成的表单代码）。注意：代码生成预览页的复制用的是**自定义指令 v-copyText**（src\directive\common\copyText.js，textarea+execCommand 实现），不依赖 clipboard 包 | 剪贴板 |
| **file-saver** | 2.0.5 | src\plugins\download.js:3,70（$download 的 name/resource/zip 三种下载 + saveAs 包装）；src\utils\request.js:7,138（通用 download() 失败分支外成功保存）；src\views\tool\build\index.vue:267（生成代码下载） | 浏览器端保存 blob |
| **@vueuse/core** | 14.1.0 | src\layout\index.vue:17（useWindowSize 响应式移动端断点992）；src\components\Screenfull\index.vue:8（useFullscreen）；src\store\modules\settings.js:2,6-7（useDark/useToggle 暗色模式） | 组合式工具 |

**未被使用的依赖**：全部 17 个运行时依赖都有实际用途，没有完全未用的。但有两点值得 React 重写时注意：

1. **sortablejs（1.14.0）不是直接依赖却被直接 import**：src\components\FileUpload\index.vue:45、src\components\ImageUpload\index.vue:53、src\views\tool\gen\editTable.vue:129 三处 `import Sortable from 'sortablejs'`，靠 vuedraggable 的传递依赖存活。gen\editTable.vue 用它做**字段列表行拖拽排序**（handle=".allowDrag"，onEnd 重排 columns 并重写 sort 序号，editTable.vue:198-211）。
2. Editor/FileUpload/ImageUpload/ImagePreview 虽全局注册，但 FileUpload/ImageUpload/ImagePreview 当前无任何业务页面消费（Editor 仅通知公告页用）。

---

## 3. 平台能力开关

### src\api\platform.js 完整内容（共9行）

```js
import request from '@/utils/request'

// 获取平台信息（语言/框架版本与功能开关，数据监控等服务端特有页面据此降级提示）
export function getPlatformInfo() {
  return request({
    url: '/getPlatformInfo',
    method: 'get'
  })
}
```

### /getPlatformInfo 响应结构

AjaxResult 信封（HTTP 200 + body code）。契约主文档：`RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md`；Java 实现基准：`RuoYi-Vue/ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysPlatformController.java`

```json
{
  "code": 200,
  "msg": "操作成功",
  "framework": "RuoYi",          // Java=RuoYiConfig.getName()；GO=RuoYi-Vue-GO；Python=AppConfig.app_name
  "version": "3.9.2",            // Java=RuoYiConfig.getVersion()；GO/Python=常量1.0.0
  "language": "java",            // java | go | python
  "languageVersion": "17.0.x",   // java.version / runtime.Version() / platform.python_version()
  "features": {
    "druidMonitor": true,        // ← features 字段名完整清单就是这3个
    "serverMonitor": true,
    "swaggerDocs": true
  }
}
```

features 字段名全集（**只有3个**，前端不得自行发明）：`druidMonitor`、`serverMonitor`、`swaggerDocs`。各版取值：Java 全 true；GO 全 false；Python serverMonitor=true、druidMonitor/swaggerDocs=false。鉴权：登录即可（无 @PreAuthorize）。

### 三个消费页面（模式完全一致）

| 页面 | features 初值 | 开关字段 | 降级渲染 | 重新检测 |
|---|---|---|---|---|
| src\views\monitor\server\index.vue | `ref({ serverMonitor: true })` (L183) | serverMonitor | `v-if="!features.serverMonitor"` → el-result icon="info" title="该功能仅 Java 版提供" sub-title="服务监控在当前后端运行时未实现，暂不可用。"(L2-7)；否则渲染 CPU/内存/JVM/磁盘卡片 | extra 插槽按钮 `@click="getPlatform"` |
| src\views\monitor\druid\index.vue | `ref({ druidMonitor: true })` (L20) | druidMonitor | 同上，sub-title="数据监控基于 Druid 连接池控制台…"(L3-8)；正常时渲染 `<i-frame :src="VITE_APP_BASE_API + '/druid/login.html'">` | 同上 |
| src\views\tool\swagger\index.vue | `ref({ swaggerDocs: true })` (L18) | swaggerDocs | 同上，sub-title="系统接口页基于 springdoc swagger-ui…"(L2-7)；正常时 `<i-frame :src="VITE_APP_BASE_API + '/swagger-ui/index.html'">` | 同上 |

关键逻辑细节：

- **初值全 true 的理由**（RuoYi-Vue3\AGENTS.md:70）：保持 Java 版原行为，且接口慢/失败时不误显降级页。
- `getPlatform()`：`getPlatformInfo().then(response => { features.value = response.features || {} })`；server 页额外 `if (features.value.serverMonitor !== false) getList()`（即只有明确 false 才不拉 /monitor/server 数据，L197）。
- 「重新检测」按钮 = 再调一次 getPlatform() 覆盖 features，切换后端后用户手动点即可恢复。
- 判断一律 `!== false`/`!features.x` 宽松判断，不判后端语言。

---

## 4. 富文本 Editor

- **存储格式：HTML 字符串**。src\components\Editor\index.vue:20-21 `<quill-editor v-model:content="content" contentType="html">`，内容经 `update:modelValue` 直接是 HTML。存储字段如 notice.noticeContent；展示端用 `v-html`（src\layout\components\HeaderNotice\DetailView.vue:45）。
- **使用页面：仅通知公告** src\views\system\notice\index.vue:146 `<editor v-model="form.noticeContent" :min-height="192"/>`。
- **props**（Editor\index.vue:43-73）：`modelValue`(String，HTML)、`height`(Number px)、`minHeight`(Number)、`readOnly`(默认false)、`fileSize`(图片上限 MB，默认5)、`type`(默认"url"，图片以 URL 插入；可选 "base64")。
- 工具栏：加粗/斜体/下划线/删除线、引用/代码块、有序/无序列表、缩进、字号、标题H1-6、颜色/背景、对齐、清除格式、链接/图片/视频（options L75-96）。
- **图片上传**：type=url 时隐藏 el-upload 直传 `VITE_APP_BASE_API + "/common/upload"`（name="file"，Bearer 头）；成功取 `res.fileName`，在光标处 `quill.insertEmbed(length, "image", VITE_APP_BASE_API + res.fileName)`（L153-167）；校验 jpeg/jpg/png/svg、≤fileSize。**粘贴图片**也走同一接口（handlePasteCapture→insertImage，裸 axios multipart，L175-195）。type=base64 时走 quill 内建（不推荐）。
- 回显：watch modelValue immediate，空值兜底 `"<p></p>"`（L110-114）。
- quill 版本被 package.json overrides/resolutions 锁定为 2.0.2。

## 5. 头像裁剪（vue-cropper）

- 唯一使用页：src\views\system\user\profile\userAvatar.vue（个人中心头像）。
- 配置（L75-84）：`img`=userStore.avatar、autoCrop=true、autoCropWidth/Height=200、fixedBox=true（固定裁剪框）、outputType="png"。左右两栏：左裁剪右实时预览（@realTime）。按钮：选择（el-upload 覆盖 http-request 为空函数，FileReader 读 dataURL 填入 options.img）/放大/缩小/左旋/右旋/提交。
- **上传方式：multipart 二进制，不是 base64**（L130-142）：`cropper.getCropBlob(blob)` → `FormData.append("avatarfile", blob, filename)` → `uploadAvatar(formData)`（src\api\system\user.js:104-110，`POST /system/user/profile/avatar`，该 api 函数声明了 `Content-Type: application/x-www-form-urlencoded` 但 FormData 实际按 multipart/form-data+boundary 发送；React 重写时直接用 multipart 即可）。
- 成功后 `options.img = VITE_APP_BASE_API + response.imgUrl`，同步 `userStore.avatar`，响应字段为 `{ imgUrl }`。关闭弹窗时 options.img 还原为 userStore.avatar。

## 6. vuedraggable / js-beautify

- **vuedraggable：只用于「表单构建器」页 src\views\tool\build\**，与代码生成（gen）无关：
  - build\index.vue:14-82：左侧三类组件面板（input/select/layout 各一个 draggable，`:sort="false"` 只拖出不重排）拖入中间 drawing-board（画布本身也是 draggable，支持组内排序）；
  - build\DraggableItem.vue:9：嵌套布局（rowFormItem 子项）递归拖拽容器；
  - build\RightPanel.vue:218：右侧属性面板中 checkbox/radio/select 的选项列表拖拽排序。
- **gen 页面（src\views\tool\gen\）没有用 vuedraggable**：
  - 表字段拖拽排序用的是 sortablejs（editTable.vue:129,198-211，非 package.json 直接依赖）；
  - 代码预览（gen\index.vue:130-142）是 el-tabs + `<pre>` 原样展示，无美化、无拖拽；复制走 v-copyText 指令（textarea+execCommand，src\directive\common\copyText.js）。
- **js-beautify：只用于 build 页**生成代码的格式化：build\index.vue:276-282 `generateCode()` 把 makeUpHtml/vueScript/makeUpCss 产物拼接后 `beautifier.html(..., beautifierConf.html)`（配置 src\utils\index.js:336，2空格缩进、wrap_line_length 110 等），结果用于 ClipboardJS 复制或 file-saver 下载。gen 预览不经 beautify。

---

### 补充：全局插件（组件行为依赖的上下文）

- src\plugins\index.js 挂载 `$tab`（页签刷新/关闭）、`$auth`（hasPermi/hasRole）、`$cache`（session/local JSON 封装，RightToolbar storageKey 用它）、`$modal`（msgError/msgSuccess/loading/confirm/alert，各上传组件用它）、`$download`（file-saver 下载，ExcelImportDialog 模板下载用它）。
- main.js 全局方法：useDict、parseTime、resetForm、handleTree、addDateRange、getConfigKey、selectDictLabel(s)、download（POST 表单编码 blob 下载）。
- 环境变量 `VITE_APP_BASE_API`（.env.development 中 /dev-api，代理到 8080）是所有上传/iframe/图片回显的前缀。
