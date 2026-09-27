# Spec 2.0.0 状态与路由：RTK 七 store + 动态路由 + 守卫

>
> **状态：🔶 核心完成（2026-09-27）**。七 store 全部落地；动态路由 Gate 方案 + 守卫 + 登录闭环打通并浏览器端到端验证（login → getInfo → getRouters → /index，时序正确）。登录页/register/index/lock/profile 骨架随本模块提前落地。**未完结项**（保持🔶，随后续模块收口）：拍平/cachedViews/name 派生纯函数单测（3.0.0 前）、三级菜单与锁屏劫持可视化验证（3.0.0/4.0.0 后）。
> **实施记录（2026-09-27）**：① 写入过程发生输出损坏事故（app.ts/gates.tsx 两次写坏），恢复后改用「短写纪律」——长文件拆 Write 短桩 + heredoc/小段 Edit 追加，tsc 交错验证；② React element 不可序列化，挂载版路由对象存模块级 Map（setBackendMenus/getResolvedRoutes），不进 Redux；③ StrictMode 下 getInfo/getRouters 双次调用为开发模式行为，生产构建不受影响；④ antd Form.Item 内不能包 div 包裹 Input（值绑定失效导致登录验证码丢失），验证码图改用 Input suffix。
> **背景**：React 无 addRoute、无 keep-alive，本模块解决前端复刻的两个结构性差异；store 行为逐条对齐基准 Pinia 实现。
> **契约侦察**：[../reference/02-infra-contract.md](../reference/02-infra-contract.md) §3~§5（守卫/路由表/store 逐条行为）+ §12 易错清单。
> **依赖**：01。

## 范围

- RTK 七 store（user / permission / dict / app / settings / tagsView / lock）。
- 静态路由表（constantRoutes + dynamicRoutes 五条）。
- 动态路由 Gate 挂载方案 + 全局守卫组件。
- ~~KeepAlive 自研~~（**用户拍板暂不实现**，见 deviations #1 与遗留草稿 §5）。

## 关键行为契约（逐条对照 reference/02）

1. **user store**：login（isToken:false+repeatSubmit:false → setToken Cookie → lock 解锁）；getInfo（avatar 三分支；**roles 空数组 → ['ROLE_DEFAULT']**；pwdChrtype 存 sessionStorage 键 `pwrChrtype`；isDefaultModifyPwd/isPasswordExpired 弹窗跳 Profile resetPwd）；logOut。
2. **permission store**：generateRoutes 三份深拷贝（sidebar 版/拍平版/topbar 版）+ filterDynamicRoutes（hasPermiOr/hasRoleOr，都不带丢弃）；filterAsyncRouter 三占位映射（Layout/ParentView/InnerLink/loadView glob 懒加载）；type=true 拍平 ParentView 链。
3. **守卫顺序**（严格）：setTitle → 读 isLock → 去 /login 重定向 / → 白名单放行 → isLock 强跳 /lock → !isLock 挡 /lock → roles 空则 getInfo+generateRoutes 后重渲染 → 放行。白名单 `['/login','/register']`（isPathMatch 通配）。
4. **tagsView store**：visitedViews/cachedViews(name)/iframeViews；noCache 规则；affix；持久化开关（`tagsViewPersist` + localStorage 键 `tags-view-visited` 过滤 affix）；del 族语义（delOthers 保留 affix+当前、delAll visited 只留 affix）。
5. **settings store**：11 字段默认值 + localStorage `layout-setting` 覆盖；setTitle/dynamicTitle。
6. **app store**：sidebarStatus/size Cookie；**lock store**：`screen-lock`/`screen-lock-path` localStorage 键。
7. NProgress start/done；meta.title → document.title。

## 设计决策

1. **动态路由 Gate 方案**（deviations 无差异，效果等价 `{...to, replace:true}`）：
   - `<DynamicRoutesGate>`：权限数据未就绪（roles 空且非白名单）渲染全屏 Spin；就绪后 `useRoutes(全量表)` 重渲染。
   - 404 catch-all 放路由数组**最后**；外链路由不注册 Route（侧边栏渲染 `<a target=_blank>`）；InnerLink 注册为占位路由由 IframeToggle 渲染。
2. **KeepAlive 暂缓**（2026-09-27 用户决定）：现阶段路由切换即重挂载，页面状态不保留；tagsView.cachedViews 数据照常维护（noCache 规则 + 单测），将来补缓存层时直接可用。演示后决策（specs/README 遗留待办 #1）。设计草稿保留在 §5。
3. RTK slice 划分一对一对应基准 Pinia store，命名同构；thunk 内不直接 fetch，一律走 1.0.0 API 层。
4. **路由 name 是契约**：后端菜单无 name，前端按 path 派生稳定 name（对齐基准 loadView 后的行为），cachedViews/tagsView/KeepAlive 将来都依赖它——单测覆盖派生规则。

### §5 KeepAlive 设计草稿（暂不实现，留档）

- `<KeepAlive include={cachedViews}>`：`Map<name, {element}>` 缓存组件实例，非激活 `display:none`；iframe 页签独立常驻（IframeToggle display 切换）；刷新 = 删缓存重挂载（对齐 $tab.refreshPage）。
- 触发条件：specs/README 遗留待办 #1 演示后用户拍板。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
