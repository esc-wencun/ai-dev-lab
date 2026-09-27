# Tasks · 1 基础设施

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] utils：errorCode.ts（四条文案逐字一致）（2026-09-27）
- [x] utils：ruoyi.ts 全集（parseTime/addDateRange/selectDictLabel(s)/handleTree/tansParams/parseStrEmpty/getNormalPath/blobValidate；isHttp/isPathMatch 落在 validate.ts；resetForm 不移植——React 无 $refs 对应物，页面用 antd Form resetFields）（2026-09-27，21 项单测对照基准）
- [x] utils：auth.ts（Cookie `Admin-Token`，会话级）（2026-09-27）
- [x] utils：jsencrypt.ts（密钥对从基准复制，文件头注明来源与"仅记住我"用途）（2026-09-27）
- [x] utils：cache.ts → 落位 src/plugins/cache.ts（对位基准 $cache 路径）（2026-09-27）
- [x] utils：dict.ts（useDict hook：store 缓存 → GET /system/dict/data/type/{type} → 映射 label/value/elTagType/elTagClass；并发不去重与基准一致，spec 原文"只发一次"系误记已修正）（2026-09-27，真实后端 getDicts 链路待字典页联调）
- [x] utils：passwordRule.ts（0-4 动态规则 + register 固定规则'0' + 特殊字符集，键 `pwrChrtype`）（2026-09-27）
- [x] 单测：tansParams（嵌套/空值/编码/尾 &）、handleTree、passwordRule 0-4、blobValidate、parseTime、addDateRange、selectDictLabel、getNormalPath、parseStrEmpty——25 项全绿（2026-09-27）
- [x] request.ts：实例（baseURL/timeout 10000/默认 JSON 头）（2026-09-27）
- [x] request.ts：请求拦截器（Bearer / isToken 跳过 / GET tansParams 拼 URL / 防重四条）（2026-09-27）
- [x] request.ts：响应拦截器（blob 直通 / 401 确认框防重入 / 500/601/其他 / 200 全 body / HTTP 层三文案）（2026-09-27）
- [x] request.ts：download()（POST form-urlencoded + loading + blobValidate 分支）（2026-09-27；blob 错误分支待 6.0.0 导出链路实测）
- [x] plugins/download.ts（name/resource/zip + download-filename 头 + printErrMsg）（2026-09-27；$download 三方法随 gen/资源页联调）
- [x] API 层 20 文件（login/platform/menu + system 9 + monitor 7 + tool/gen）——对照 reference/01 §2 表逐函数核对（2026-09-27）
- [x] 验证：getCodeImg 拿验证码图（img/uuid/captchaEnabled 齐全）；401 确认框「系统提示/登录状态已过期…/重新登录/取消」逐字渲染、取消可留页 + 标志复位、防重第二次拦截「数据正在处理，请勿重复提交」（2026-09-27，浏览器端到端）
- [x] 依赖修正：antd 6 自动升版 → 钉回 antd@5.29.3 / @ant-design/icons@5.6.1 / echarts@5.6.0（对齐 spec 选型大版本）；React 19 兼容告警出现 → 引入 @ant-design/v5-patch-for-react-19@1.0.3（main.tsx 引入，0.0.0 预留项兑现）
