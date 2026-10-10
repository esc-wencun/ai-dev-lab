-- ============================================================
-- AI 基础域表结构与内置数据（specs/2.0.1）
-- 表设计权威源：specs/AI相关表设计方案.md 第一部分（含逻辑审查修正）
-- 可重复执行：所有 insert 前置 delete；菜单/字典按主键覆盖
-- 手动执行方式（本机 Docker，宿主端口 13306）：
--   docker exec -i ruoyi-mysql mysql -uwencun -p111111 ry-vue < sql/ai_20261010.sql
-- ============================================================

SET NAMES utf8mb4;

-- ----------------------------
-- 1. AI 厂商表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_provider (
  provider_id     bigint        NOT NULL AUTO_INCREMENT COMMENT '厂商ID',
  provider_name   varchar(100)  NOT NULL                COMMENT '厂商名称',
  provider_code   varchar(50)   NOT NULL                COMMENT '厂商编码（唯一）',
  protocol        varchar(32)   NOT NULL DEFAULT 'openai' COMMENT '协议类型（openai/anthropic/ollama/gemini/dify/coze/custom）',
  api_host        varchar(500)  NOT NULL                COMMENT '默认API地址（必填，各厂商官方地址不同）',
  icon            varchar(1000) DEFAULT NULL            COMMENT '厂商图标',
  website         varchar(1000) DEFAULT NULL            COMMENT '官网地址',
  key_apply_url   varchar(1000) DEFAULT NULL            COMMENT '密钥申请地址',
  status          char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order      int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  remark          varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by       varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time     datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by       varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time     datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag        char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (provider_id),
  UNIQUE KEY uk_provider_code (provider_code, del_flag),
  KEY idx_status (status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI厂商表';

-- ----------------------------
-- 2. AI 密钥表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_api_key (
  key_id        bigint       NOT NULL AUTO_INCREMENT COMMENT '密钥ID',
  provider_id   bigint       NOT NULL                COMMENT '所属厂商ID',
  key_name      varchar(64)  NOT NULL                COMMENT '密钥名称',
  api_key       varchar(512) NOT NULL                COMMENT '密钥内容（支持${ENV_VAR}占位符）',
  base_url      varchar(500) DEFAULT NULL            COMMENT 'Key专属网关地址（NULL用厂商默认）',
  api_path      varchar(500) DEFAULT NULL            COMMENT '二级地址：拼接在网关地址后的请求路径（NULL用协议默认路径）',
  status        char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0启用 1停用）',
  expire_time   datetime     DEFAULT NULL            COMMENT '过期时间',
  create_by     varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time   datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by     varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time   datetime     DEFAULT NULL            COMMENT '更新时间',
  remark        varchar(500) DEFAULT NULL            COMMENT '备注',
  del_flag      char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (key_id),
  KEY idx_provider (provider_id, status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI密钥表';

-- ----------------------------
-- 3. AI 模型表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_model (
  model_id          bigint       NOT NULL AUTO_INCREMENT COMMENT '模型ID',
  provider_id       bigint       NOT NULL                COMMENT '所属厂商ID',
  key_id            bigint       NOT NULL                COMMENT '绑定密钥ID（须与厂商一致；同一模型可经多把Key建多条通道）',
  model_name        varchar(100) NOT NULL                COMMENT '模型名称（展示用）',
  model_code        varchar(100) NOT NULL                COMMENT '模型标识（传给API）',
  model_type        tinyint      NOT NULL DEFAULT 1      COMMENT '模型类型（1对话 2图片 3语音 4视频 5向量 6重排序）',
  support_thinking  tinyint      NOT NULL DEFAULT 0      COMMENT '支持深度思考（0否 1是）',
  support_tool      tinyint      NOT NULL DEFAULT 0      COMMENT '支持工具调用',
  support_vision    tinyint      NOT NULL DEFAULT 0      COMMENT '支持图片输入',
  support_audio     tinyint      NOT NULL DEFAULT 0      COMMENT '支持音频输入',
  support_video     tinyint      NOT NULL DEFAULT 0      COMMENT '支持视频输入',
  model_dimension   int          DEFAULT NULL            COMMENT '向量维度（仅type=5时有意义）',
  temperature       double       DEFAULT NULL            COMMENT '默认温度（NULL=供应商默认）',
  max_tokens        int          DEFAULT NULL            COMMENT '默认最大输出Token',
  max_contexts      int          DEFAULT NULL            COMMENT '默认携带历史消息条数上限',
  status            char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order        int          NOT NULL DEFAULT 0      COMMENT '显示顺序（升序）',
  remark            varchar(500) DEFAULT NULL            COMMENT '备注（选型建议）',
  create_by         varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time       datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by         varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time       datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag          char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (model_id),
  UNIQUE KEY uk_provider_key_model (provider_id, key_id, model_code, del_flag),
  KEY idx_key (key_id),
  KEY idx_type_status (model_type, status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI模型表';

-- ----------------------------
-- 内置数据：厂商（api_host 填到版本段、不带尾斜杠；2026-10-11 扩充为 21 家目录）
-- 存在性判断：按 provider_code NOT EXISTS——已存在（含用户改过配置的）不覆盖不添加
-- ----------------------------
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 100, 'DeepSeek', 'deepseek', 'openai', 'https://api.deepseek.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/deepseek-color.png', 'https://www.deepseek.com', 'https://platform.deepseek.com/api_keys', '0', 1, '深度求索', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'deepseek');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 101, '智谱开放平台', 'zhipu', 'openai', 'https://open.bigmodel.cn/api/paas', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/zhipu-color.png', 'https://open.bigmodel.cn', 'https://open.bigmodel.cn/usercenter/apikeys', '0', 2, '智谱 GLM 系列', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'zhipu');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 102, '阿里百炼', 'bailian', 'openai', 'https://dashscope.aliyuncs.com/compatible-mode', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/qwen-color.png', 'https://www.aliyun.com/product/bailian', 'https://bailian.console.aliyun.com/?apiKey=1', '0', 3, '通义千问 / DashScope 兼容模式', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'bailian');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 103, '硅基流动', 'siliconflow', 'openai', 'https://api.siliconflow.cn', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/siliconcloud-color.png', 'https://siliconflow.cn', 'https://cloud.siliconflow.cn/account/ak', '0', 4, '一站式模型云', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'siliconflow');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 104, 'Ollama（本地）', 'ollama', 'ollama', 'http://localhost:11434', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/ollama.png', 'https://ollama.com', null, '0', 5, '本地部署，无需密钥', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'ollama');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 105, '文心一言', 'yiyan', 'openai', 'https://qianfan.baidubce.com/v2', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/wenxin-color.png', null, 'https://console.bce.baidu.com/iam/#/iam/apikeylist', '0', 6, '百度千帆 v2（OpenAI 兼容）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'yiyan');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 106, '讯飞星火', 'xinghuo', 'openai', 'https://spark-api-open.xf-yun.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/spark-color.png', null, null, '0', 7, '标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'xinghuo');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 107, '豆包', 'doubao', 'openai', 'https://ark.cn-beijing.volces.com/api/v3', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/doubao-color.png', null, 'https://console.volcengine.com/ark', '0', 8, '火山方舟（OpenAI 兼容）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'doubao');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 108, '混元', 'hunyuan', 'openai', 'https://api.hunyuan.cloud.tencent.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/hunyuan-color.png', null, null, '0', 9, '腾讯；标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'hunyuan');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 109, 'MiniMax', 'minimax', 'openai', 'https://api.minimax.chat', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/minimax-color.png', null, null, '0', 10, '密钥需配二级地址 /v1/text/chatcompletion_v2', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'minimax');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 110, '月之暗面', 'moonshot', 'openai', 'https://api.moonshot.cn', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/moonshot.png', null, 'https://platform.moonshot.cn/console/api-keys', '0', 11, 'Kimi；标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'moonshot');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 111, '百川智能', 'baichuan', 'openai', 'https://api.baichuan-ai.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/baichuan-color.png', null, null, '0', 12, '标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'baichuan');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 112, '阶跃星辰', 'stepfun', 'openai', 'https://api.stepfun.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/stepfun-color.png', null, 'https://platform.stepfun.com/interface-key', '0', 13, '标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'stepfun');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 113, 'OpenAI', 'openai', 'openai', 'https://api.openai.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/openai.png', 'https://openai.com', 'https://platform.openai.com/api-keys', '0', 14, 'OpenAI 官方', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'openai');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 114, 'Azure OpenAI', 'azureopenai', 'custom', 'https://your-resource.openai.azure.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/azure-color.png', null, null, '0', 15, '微软；端点按资源域名配置，需专用适配器（本期仅登记）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'azureopenai');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 115, 'Anthropic', 'anthropic', 'anthropic', 'https://api.anthropic.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/anthropic.png', 'https://www.anthropic.com', 'https://console.anthropic.com/settings/keys', '0', 16, 'Claude；密钥需配二级地址 /v1/messages（适配器后续扩展）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'anthropic');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 116, 'Gemini', 'gemini', 'gemini', 'https://generativelanguage.googleapis.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/gemini-color.png', null, 'https://aistudio.google.com/app/apikey', '0', 17, '谷歌；密钥需配二级地址（适配器后续扩展）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'gemini');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 117, 'StableDiffusion', 'stablediffusion', 'custom', 'https://api.stability.ai', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/stability-color.png', null, null, '0', 18, 'Stability AI，图片生成', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'stablediffusion');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 118, 'Midjourney', 'midjourney', 'custom', 'https://your-proxy.example.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/midjourney.png', null, null, '0', 19, '官方无 API；常用 midjourney-proxy 反代，地址按部署填', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'midjourney');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 119, 'Suno', 'suno', 'custom', 'https://your-proxy.example.com', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/suno.png', null, null, '0', 20, 'Suno AI 音乐；常用第三方代理，地址按部署填', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'suno');
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 120, 'Grok', 'grok', 'openai', 'https://api.x.ai', 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/grok.png', null, 'https://console.x.ai', '0', 21, 'xAI；标准 /v1 兼容端点', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'grok');

-- ----------------------------
-- 内置数据：密钥（占位符引用环境变量，真实 Key 不入库）
-- ----------------------------
delete from ai_api_key where key_id between 100 and 199;
insert into ai_api_key values(100, 101, '智谱官方Key',   '${glm_api_key}',       null, null, '0', null, 'admin', sysdate(), '', null, '本机环境变量 glm_api_key', '0');
insert into ai_api_key values(101, 102, '百炼占位Key',   '${DASHSCOPE_API_KEY}', null, null, '0', null, 'admin', sysdate(), '', null, '本机环境变量 DASHSCOPE_API_KEY（未配置时调用报错）', '0');
insert into ai_api_key values(102, 104, '本地无需密钥',  'local-no-key',         null, null, '0', null, 'admin', sysdate(), '', null, 'Ollama 本地占位密钥', '0');

-- ----------------------------
-- 内置数据：模型（迁移 2.0.0 application.yml 的 GLM 配置为 DB 数据）
-- ----------------------------
delete from ai_model where model_id between 100 and 199;
-- 模型 100（智谱官方通道）：本机 glm_api_key 是内网网关 Key、配官方地址必 401，默认停用；
-- 有智谱官方 Key 时启用并把 key 换绑到智谱官方 Key 记录
insert into ai_model values(100, 101, 100, 'GLM-4.7', 'glm-4.7', 1, 0, 1, 0, 0, 0, null, null, null, null, '1', 1, '复杂编程 / 长文本；通道失配默认停用（见上）', 'admin', sysdate(), '', null, '0');

-- ----------------------------
-- 菜单（AI 管理目录 + 三页 + 按钮；官方最大 ID 1060，AI 从 2000 起百位预留段）
-- ----------------------------
delete from sys_menu where menu_id between 2000 and 2099;
insert into sys_menu values('2000', 'AI管理', '0', '50', 'ai', null, '', '', 1, 0, 'M', '0', '0', '', 'message', 'admin', sysdate(), '', null, 'AI 能力管理目录');
insert into sys_menu values('2001', '厂商管理', '2000', '1', 'provider', 'ai/provider/index', '', '', 1, 0, 'C', '0', '0', 'ai:provider:list', 'component', 'admin', sysdate(), '', null, 'AI厂商菜单');
insert into sys_menu values('2002', '密钥管理', '2000', '2', 'key',      'ai/key/index',      '', '', 1, 0, 'C', '0', '0', 'ai:key:list',      'lock',      'admin', sysdate(), '', null, 'AI密钥菜单');
insert into sys_menu values('2003', '模型管理', '2000', '3', 'model',    'ai/model/index',    '', '', 1, 0, 'C', '0', '0', 'ai:model:list',    'education', 'admin', sysdate(), '', null, 'AI模型菜单');

-- 厂商管理按钮
insert into sys_menu values('2004', '厂商查询', '2001', '1', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:provider:query',  '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2005', '厂商新增', '2001', '2', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:provider:add',    '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2006', '厂商修改', '2001', '3', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:provider:edit',   '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2007', '厂商删除', '2001', '4', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:provider:remove', '#', 'admin', sysdate(), '', null, '');
-- 密钥管理按钮
insert into sys_menu values('2008', '密钥查询', '2002', '1', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:key:query',  '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2009', '密钥新增', '2002', '2', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:key:add',    '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2010', '密钥修改', '2002', '3', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:key:edit',   '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2011', '密钥删除', '2002', '4', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:key:remove', '#', 'admin', sysdate(), '', null, '');
-- 模型管理按钮
insert into sys_menu values('2012', '模型查询', '2003', '1', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:model:query',  '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2013', '模型新增', '2003', '2', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:model:add',    '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2014', '模型修改', '2003', '3', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:model:edit',   '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2015', '模型删除', '2003', '4', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:model:remove', '#', 'admin', sysdate(), '', null, '');

-- ----------------------------
-- 字典：模型类型（状态复用 sys_normal_disable，不新增）
-- ----------------------------
delete from sys_dict_type where dict_type = 'ai_model_type';
insert into sys_dict_type values(0, 'AI模型类型', 'ai_model_type', '0', 'admin', sysdate(), '', null, 'AI 模型类型列表');
delete from sys_dict_data where dict_type = 'ai_model_type';
insert into sys_dict_data values(0, 1, '对话',   '1', 'ai_model_type', '', 'primary', 'Y', '0', 'admin', sysdate(), '', null, '对话模型');
insert into sys_dict_data values(0, 2, '图片',   '2', 'ai_model_type', '', 'success', 'N', '0', 'admin', sysdate(), '', null, '图片生成模型');
insert into sys_dict_data values(0, 3, '语音',   '3', 'ai_model_type', '', 'info',    'N', '0', 'admin', sysdate(), '', null, '语音模型');
insert into sys_dict_data values(0, 4, '视频',   '4', 'ai_model_type', '', 'warning', 'N', '0', 'admin', sysdate(), '', null, '视频生成模型');
insert into sys_dict_data values(0, 5, '向量',   '5', 'ai_model_type', '', 'danger',  'N', '0', 'admin', sysdate(), '', null, 'Embedding 模型');
insert into sys_dict_data values(0, 6, '重排序', '6', 'ai_model_type', '', '',        'N', '0', 'admin', sysdate(), '', null, 'Rerank 模型');

-- 内网模型网关（2026-10-11 端到端验证时发现：本机 glm_api_key 是企业内网网关的 Key，
-- 不是智谱官方的——内置的智谱+官方地址通道对本机环境不可用，真实可用通道是这家）
insert into ai_provider (provider_id, provider_name, provider_code, protocol, api_host, icon, website, key_apply_url, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 121, '企业内网网关', 'airouter', 'openai', 'https://airouter.sinosoft.com.cn:9000', null, null, null, '0', 0, 'sinosoft 内网模型路由（OpenAI 兼容，/v1 已含在 api_host 末段——解析约定 host 不带尾斜杠 + /v1/chat/completions 拼接）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_provider where provider_code = 'airouter');
insert into ai_api_key (key_id, provider_id, key_name, api_key, base_url, api_path, status, expire_time, create_by, create_time, update_by, update_time, remark, del_flag)
select 103, 121, '内网网关Key', '${glm_api_key}', null, null, '0', null, 'admin', sysdate(), '', null, '本机环境变量 glm_api_key', '0' from dual
where not exists (select 1 from ai_api_key where key_id = 103);
insert into ai_model (model_id, provider_id, key_id, model_name, model_code, model_type, support_thinking, support_tool, support_vision, support_audio, support_video, model_dimension, temperature, max_tokens, max_contexts, status, sort_order, remark, create_by, create_time, update_by, update_time, del_flag)
select 101, 121, 103, 'GLM-5.3-Flash', 'GLM-5.3-Flash', 1, 0, 1, 0, 0, 0, null, null, null, null, '0', 2, '内网网关 GLM（thinking 模型，本机默认）', 'admin', sysdate(), '', null, '0' from dual
where not exists (select 1 from ai_model where model_id = 101);
