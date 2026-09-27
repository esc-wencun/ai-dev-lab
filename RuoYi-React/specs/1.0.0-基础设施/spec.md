# Spec 1.0.0 基础设施：请求层契约 + 工具函数 + API 层

>
> **状态：✅ 已完成（2026-09-27）**。请求层契约全部落地并浏览器端到端验证：验证码接口、401 确认框（文案逐字一致、可留页、防重入复位）、防重复提交拦截、tansParams 嵌套编码；25 项单测全绿；API 层 20 文件对照 reference/01 §2 表核对无缺漏。两项遗留（500 形态人工验证、download 真实导出）分别待 5.0.0/6.0.0 随页覆盖。实施要点：antd 自动升 6.x 已钉回 5.x；React 19 告警出现后引入官方 v5-patch（1.0.3）。
> **背景**：请求层是三版后端兼容的生命线，行为契约**逐条复刻** RuoYi-Vue3 `src/utils/request.js`（含文案、存储键名、分支顺序），禁止顺手优化。
> **契约侦察**：[../reference/02-infra-contract.md](../reference/02-infra-contract.md) §1~§8（唯一权威行为描述）；features 开关见 reference/03 §3；API 全端点表见 reference/01 §2。
> **依赖**：00。

## 范围

- utils 纯函数全集（单测覆盖）：errorCode / ruoyi（parseTime/handleTree/tansParams 等）/ auth / jsencrypt / dict / cache / passwordRule。
- axios 请求层 `utils/request.ts`：Bearer 注入、GET 参数映射、防重复提交、body code 四分支、blob 下载、HTTP 层错误文案。
- 下载插件 `plugins/download.ts`（$download 三件）。
- API 层 20 个文件全端点（reference/01 §2 对照表逐一实现）。

## 行为契约要点（实现时逐条对照 reference/02，此处只列易错）

1. 401 分支：模块级 `isRelogin.show` 防重入 + **确认框**（可留页）文案「登录状态已过期，您可以继续留在该页面，或者重新登录」→ 确认走 user store logOut + `location.href='/index'`；无论确认与否都 reject。
2. code 500 → message.error；601 → message.warning；其他非 200 → notification.error；文案优先级 `errorCode[code] > msg > errorCode.default`。
3. code 200 resolve **整个 body**（页面直接取 res.rows/res.total/res.data）。
4. 防重：仅 post/put；`headers.repeatSubmit===false` 跳过；`headers.interval||1000`；sessionStorage 键 `sessionObj`；≥5MB 跳过 + console.warn。
5. `download()`：**POST + tansParams transformRequest + x-www-form-urlencoded + blob**；loading「正在下载数据，请稍候」；blobValidate（`type!=='application/json'`）失败走 blob.text→JSON→errorCode。
6. HTTP 层错误三文案（Network Error/timeout/status code）+ duration 5000。
7. token Cookie `Admin-Token` 会话级；验证码接口 timeout 20000。
8. uploadAvatar 按 multipart 声明（基准历史瑕疵修正，见 deviations #12）。

## 设计决策

1. 401 确认框用 antd Modal.confirm；message/notification 用 antd 静态方法——**已按官方方案引入 `@ant-design/v5-patch-for-react-19`**（0.0.0 预留项兑现），静态方法可用；「Static function can not consume context」告警存在但不影响功能（动态主题穿透不支持，视觉等价即可）。
2. `tansParams`/`handleTree` 等从 RuoYi-Vue3 `src/utils/ruoyi.js` 逐行改写 TS（不重写逻辑），每个函数单测对照基准行为。
3. API 函数签名与基准同构（listXxx/getXxx/addXxx/updateXxx/delXxx），TS 化参数与返回类型标注 body 信封。
4. request.ts 的 401 onOk 通过动态 `import('@/store')` dispatch 全局 store，避免 request ↔ store 循环依赖。
5. **版本钉扎**（2026-09-27 实施发现）：npm 默认装 antd 6.6.5（超 spec 大版本），已钉回 `antd@^5`（5.29.3）、`@ant-design/icons@^5`（5.6.1）、`echarts@^5`（5.6.0）——v6 破坏性变更未知，严格按 spec 选型。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
