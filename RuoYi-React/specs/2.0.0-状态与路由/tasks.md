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
- [ ] 单测：filterAsyncRouter 拍平 / cachedViews 规则 / filterDynamicRoutes / settings 合并 / name 派生——**遗留未做（deriveRouteName/buildRouteObjects/filterDynamicRoutes 均为纯函数，单测随 3.0.0 批次 A 前补齐；当前以端到端链路验证覆盖）**
- [x] 静态路由表（constantRoutes + 五条 dynamicRoutes 分片写法 + 404 置底）（2026-09-27）
- [x] DynamicRoutesGate（App.tsx 按 generated 标志全量重渲染 useRoutes；未就绪渲染 constantRoutes 子集）（2026-09-27）
- [x] AuthGuard 守卫（白名单/锁屏劫持/初始化 Loading/NProgress/getInfo→generateRoutes 时序）（2026-09-27）
- [x] 验证：登录 → login/getInfo/getRouters 时序正确 → /index 首页渲染（2026-09-27，浏览器端到端，Network 面板确认；getInfo/getRouters 双次为 StrictMode 开发行为）
- [ ] 验证：三级菜单挂载访问/外链/无权限 404——遗留未做（依赖 3.0.0 侧边栏组件渲染菜单后的可视化验证）
- [ ] 验证：锁屏硬劫持——遗留未做（锁屏页已实现，劫持逻辑在守卫内；入口按钮随 4.0.0 Navbar 落地后验证）
- [ ] 验证：刷新页面权限重建——已部分验证（F5 后守卫重跑 getInfo+generateRoutes 停留 /index 成功），完整对照随 4.0.0
- [ ] KeepAlive 演示（待 6.0.0 列表页完成后执行，specs/README 遗留待办 #1）
