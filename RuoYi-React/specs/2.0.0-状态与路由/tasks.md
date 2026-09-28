# Tasks · 2 状态与路由

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。
> 实施说明：登录页（5.0.0 核心页）提前随本模块落地作为链路验证入口；login/register/index/lock/profile 骨架页一并就位。

- [x] store/index.ts（configureStore 七 slice 注册 + 类型导出）（2026-09-27）
- [x] user slice（login/getInfo/logOut 三 thunk，avatar 三分支/ROLE_DEFAULT/pwrChrtype 契约键）（2026-09-27）
- [x] permission slice（generateRoutes 三份深拷贝/filterDynamicRoutes/setBackendMenus，挂载版路由对象存模块级 Map 避免非序列化入 Redux）（2026-09-27）
- [x] dict slice（内存数组，刷新即失）（2026-09-27，1.0.0 已建，本模块注册）
- [x] app slice（sidebarStatus/size Cookie + device；Cookie 写入在 reducer 内为刻意对齐基准 Pinia 行为）（2026-09-27）
- [x] settings slice（AppSettings 14 字段 + isDark + layout-setting 覆盖 + setTitle/dynamicTitle）（2026-09-27）
- [x] tagsView slice（visited/cached/iframe 三视图 + noCache/affix 规则 + del 族语义；持久化 UI 联动随 4.0.0）（2026-09-27）
- [x] lock slice（screen-lock/screen-lock-path 两键 + lockScreen/unlockScreen）（2026-09-27）
- [x] 单测：filterAsyncRouter 拍平 / cachedViews 规则 / filterDynamicRoutes / settings 合并 / name 派生（2026-09-28，gates.test.ts 6 用例 + permission.test.ts 8 用例 + tagsView.test.ts 16 用例 + settings 两文件 8 用例，全绿）
- [x] 静态路由表（constantRoutes + 五条 dynamicRoutes 分片写法 + 404 置底）（2026-09-27）
- [x] DynamicRoutesGate（App.tsx 按 generated 标志全量重渲染 useRoutes；未就绪渲染 constantRoutes 子集）（2026-09-27）
- [x] AuthGuard 守卫（白名单/锁屏劫持/初始化 Loading/NProgress/getInfo→generateRoutes 时序）（2026-09-27）
- [x] 验证：登录 → login/getInfo/getRouters 时序正确 → /index 首页渲染（2026-09-27，浏览器端到端，Network 面板确认；getInfo/getRouters 双次为 StrictMode 开发行为）
- [x] 验证：三级菜单挂载访问/外链/无权限 404（2026-09-28：/system/user-auth/role/1 分配角色子页挂载成功——期间修复 buildRouteObjects hidden 过滤吞掉五条 dynamicRoutes 的 bug；外链若依官网侧栏渲染 `<a>`；404 兜底在 /redirect 修复中验证）+ 路由语法 v6→v7 迁移（`:path(.*)`→`*`、`:id(\d+)`→`:id`）
- [x] 验证：锁屏硬劫持（2026-09-29 实操：UI 锁定 → isLock=true + 跳 /lock；锁屏态访问他页被守卫弹回；错误密码拒绝仍锁；正确密码解锁回锁屏前路径。期间修复解锁回跳竞态：守卫「!isLock 且在 /lock → '/'」分支会把解锁导航顶掉，移出守卫改为锁屏页自检 + 渲染期 Navigate（unlockFlow 标志），isLock 硬劫持分支保留）
- [x] 验证：刷新页面权限重建——F5 后守卫重跑 getInfo+generateRoutes 停留原页面成功（2026-09-28 页签持久化验证时再次确认：刷新后 /system/user 直接可用）
- [ ] KeepAlive 演示（2026-09-28 已执行：user 页查询条件切页签后丢失，符合 deviations #1 预期；决策待用户）
