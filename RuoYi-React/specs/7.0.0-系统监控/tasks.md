# Tasks · 7 系统监控

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] PlatformGate 降级组件（开关字段/副标题/重新检测/正常 children）（2026-09-27）
- [x] online 页（只读 + 强退确认）——浏览器验证在线用户列表含 admin 会话（2026-09-27）
- [x] job 页（CRUD + 常用表达式下拉回填 + 策略/并发 + 状态 + 立即执行 + 详情）——浏览器验证任务列表 4 行；Crontab 七域完整版遗留未做（当前为常用表达式下拉，随 3.0.0 批次 D 补）（2026-09-27）
- [x] jobLog 子路由页（jobId 过滤 + 详情 + 清空 + 导出）——代码完成，浏览器验证随调度日志产生后（2026-09-27）
- [x] druid 页（PlatformGate + iFrame /druid/login.html）——Java 后端 iframe 正常渲染（2026-09-27）
- [x] server 页（PlatformGate + Progress 五组信息 + 条件拉数据）——Java 后端 CPU/JVM/磁盘卡片渲染（2026-09-27）
- [x] cache 主页（useEChart 玫瑰图 + 仪表盘 + resize）——浏览器验证 canvas 渲染 + Redis 信息卡（2026-09-27）
- [x] cache list 页（三栏联动 + 三级清理）——浏览器验证缓存列表/清理全部按钮（2026-09-27）
- [x] operlog 页（多字段查询 + 详情分区 + 清空 + 导出）——浏览器验证 11 行真实数据 + DictTag 操作类型（2026-09-27）
- [x] logininfor 页（查询 + 解锁账户 + 清空 + 导出）——浏览器验证 11 行数据 + 解锁按钮（2026-09-27）
- [ ] 验证：job 全生命周期（新增 Crontab 表达式 → 启动 → runJob → jobLog 出现记录 → 删除恢复）
- [ ] 验证：cache 两图渲染 + 缩放 resize + 三栏清理真实生效（Redis 键消失）
- [ ] 验证：Go/Python 后端下 druid/server 降级提示 + 重新检测；Java 下正常（swagger 页随 8.0.0 补验）
- [ ] 验证：operlog/logininfor 详情完整、导出可打开、解锁账户生效
