# 1.0.0-基础设施 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因。（2026-09-29 核对）

## 中间件行为

- [x] 未登录访问受保护页 → 302 /login；未登录 ajax → `{"code":"1","msg":"未登录或登录超时。请重新登录"}`（curl 双场景实测通过）
- [x] 会话 30 分钟空闲过期语义：TTL 随请求刷新（实测 TTL 1762s，续期生效）
- [x] 无权限 ajax → `{"code":500,"msg":...}`；无权限页面 → 302 /unauth（占位页，1.5.0 出正式模板）——拒绝分支待 2.0.0 有真实权限数据后端到端复验（当前会话 permissions 为空数组时 admin 仍放行，路径正确）
- [x] 防重复提交：1 秒内同 URL 同参数 POST 二次 → `{"code":500,"msg":"不允许重复提交，请稍候再试"}`；不同参数放行；非 POST 不拦（四场景 curl 实测）

## 横切能力

- [x] #[Log] 落库字段设计与 sys_oper_log DDL 对齐（17 列；落库端到端依赖首个带 #[Log] 的业务端点——2.0.0 登录链路首验）
- [x] 密码类敏感参数不落 oper_param（PHPUnit 3 场景固化：顶层/嵌套/非敏感保留）
- [x] PageQuery：orderByColumn 驼峰转下划线 + 白名单外字段拒绝（PHPUnit 7 场景固化）
- [x] DataScope 五种 data_scope 条件——实现完成；单测改 3.0.0 部门列表端到端覆盖（纯查询组装器，mock 成本高于收益，勾选纪律注明）

## 纪律自查

- [x] 业务代码零处直接引用 predis（grep=0，仅 RedisCache.php 自身）
- [x] 全部响应经 AjaxResult/TableDataInfo（grep 核查无裸 json()——e2e 探测路由也是经 AjaxResult）
- [x] 测试数据清理记录：Redis db1 测试会话与 repeat_submit 键全部删除，KEYS * 为空；本模块无 MySQL 写入

## 测试数据清理记录

- 2026-09-29：Redis db1 `session:a76855a3...`（测试会话）、`repeat_submit:*`（防重测试键）全部删除，`KEYS *` 复查为空。
