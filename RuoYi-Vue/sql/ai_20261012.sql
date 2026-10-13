-- ============================================================
-- AI 对话域表结构与内置数据（specs/2.0.2 阶段 B）
-- 表设计权威源：specs/AI相关表设计方案.md 第二部分（含逻辑审查修正）
-- 可重复执行：insert 前置 delete；菜单按主键覆盖
-- 手动执行（本机 Docker，宿主端口 13306）：
--   docker exec -i ruoyi-mysql mysql -uwencun -p111111 ry-vue < sql/ai_20261012.sql
-- ============================================================

SET NAMES utf8mb4;

-- ----------------------------
-- 4. AI 智能体表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_agent (
  agent_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '智能体ID',
  agent_name       varchar(100)  NOT NULL                COMMENT '智能体名称',
  avatar           varchar(1000) DEFAULT NULL            COMMENT '头像URL',
  category         varchar(64)   DEFAULT NULL            COMMENT '分类（市场分组）',
  description      varchar(500)  DEFAULT NULL            COMMENT '一句话描述',
  system_prompt    text          NOT NULL                COMMENT '系统提示词',
  model_id         bigint        DEFAULT NULL            COMMENT '默认模型ID（NULL=会话时另选）',
  knowledge_ids    varchar(1024) DEFAULT NULL            COMMENT '知识库ID数组（JSON）',
  tool_ids         varchar(1024) DEFAULT NULL            COMMENT '工具ID数组（JSON）',
  mcp_ids          varchar(1024) DEFAULT NULL            COMMENT 'MCP服务ID数组（JSON，关联ai_mcp）',
  enable_thinking  tinyint       NOT NULL DEFAULT 0      COMMENT '默认深度思考开关（0关 1开）',
  welcome_message  varchar(1000) DEFAULT NULL            COMMENT '开场白',
  suggested_questions varchar(1024) DEFAULT NULL        COMMENT '建议问题数组（JSON，开场白chip）',
  visibility       tinyint       NOT NULL DEFAULT 0      COMMENT '可见性（0私有 1公开）',
  user_id          bigint        DEFAULT NULL            COMMENT '归属用户（作者）',
  sort_order       int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  status           char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  remark           varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by        varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time      datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by        varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time      datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag         char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (agent_id),
  KEY idx_user (user_id, status),
  KEY idx_category (category)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI智能体表';

-- ----------------------------
-- 5. AI 会话表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_conversation (
  conversation_id   bigint       NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  user_id           bigint       NOT NULL                COMMENT '归属用户',
  title             varchar(256) NOT NULL                COMMENT '会话标题',
  agent_id          bigint       DEFAULT NULL            COMMENT '绑定的智能体（NULL=自由对话）',
  model_id          bigint       NOT NULL                COMMENT '会话使用模型ID',
  model_code        varchar(100) DEFAULT NULL            COMMENT '模型标识（冗余，免联表）',
  system_prompt     text                                 COMMENT '会话级系统提示词（创建时从智能体复制）',
  knowledge_ids     varchar(1024) DEFAULT NULL           COMMENT '会话级知识库ID数组（JSON快照）',
  tool_ids          varchar(1024) DEFAULT NULL           COMMENT '会话级工具ID数组（JSON快照，创建时从智能体复制）',
  mcp_ids           varchar(1024) DEFAULT NULL           COMMENT '会话级MCP服务ID数组（JSON快照，创建时从智能体复制）',
  enable_thinking   tinyint      NOT NULL DEFAULT 0      COMMENT '深度思考开关快照（0关 1开）',
  summary_upto_message_id bigint  DEFAULT NULL           COMMENT '压缩水位线（context_summary覆盖到的最后一条消息ID；NULL=未压缩）',
  temperature       double       DEFAULT NULL            COMMENT '温度覆盖（NULL=模型默认）',
  max_tokens        int          DEFAULT NULL            COMMENT '最大输出Token覆盖（NULL=模型默认）',
  max_contexts      int          DEFAULT NULL            COMMENT '历史消息条数上限覆盖（NULL=模型默认）',
  pinned            tinyint      NOT NULL DEFAULT 0      COMMENT '是否置顶（0否 1是）',
  pinned_time       datetime     DEFAULT NULL            COMMENT '置顶时间',
  last_message_time datetime     DEFAULT NULL            COMMENT '最后消息时间（列表排序；建会话时初始化为创建时间防NULL沉底）',
  message_count     int          NOT NULL DEFAULT 0      COMMENT '消息总条数（冗余计数）',
  context_msg_count int          NOT NULL DEFAULT 0      COMMENT '上下文活跃消息数（use_context=1条数，压缩触发依据）',
  context_summary   text                                 COMMENT '上下文压缩摘要（长对话优化）',
  status            char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1归档）',
  remark            varchar(500) DEFAULT NULL            COMMENT '备注',
  create_by         varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time       datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by         varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time       datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag          char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (conversation_id),
  KEY idx_user_pinned (user_id, pinned, last_message_time),
  KEY idx_agent (agent_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI会话表';

-- ----------------------------
-- 6. AI 消息表
-- ----------------------------
CREATE TABLE IF NOT EXISTS ai_message (
  message_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  conversation_id    bigint        NOT NULL                COMMENT '所属会话ID',
  reply_id           bigint        DEFAULT NULL            COMMENT '回复的消息ID（user/assistant配对；user消息为NULL）',
  msg_type           varchar(16)   NOT NULL                COMMENT '消息类型（user/assistant/system）',
  msg_status         tinyint       NOT NULL DEFAULT 1      COMMENT '消息状态（0生成中 1完成 2出错 3已取消；启动时残留0置3）',
  user_id            bigint        NOT NULL                COMMENT '触发用户',
  agent_id           bigint        DEFAULT NULL            COMMENT '智能体ID（冗余，统计用）',
  model_id           bigint        DEFAULT NULL            COMMENT '实际生成模型ID（中途切模型时与会话当前值不同）',
  model_code         varchar(100)  DEFAULT NULL            COMMENT '模型标识（冗余，免联表）',
  content            mediumtext                            COMMENT '消息内容（流式聚合后完整值）',
  use_context        tinyint       NOT NULL DEFAULT 1      COMMENT '是否纳入上下文拼装（0剔除但不删）',
  segment_ids        varchar(1024) DEFAULT NULL            COMMENT '引用的知识库分段ID数组（JSON，RAG溯源）',
  web_search_pages   text                                  COMMENT '联网来源（JSON：标题/URL/摘要）',
  attachment_urls    text                                  COMMENT '附件地址数组（JSON，多模态）',
  prompt_tokens      int           NOT NULL DEFAULT 0      COMMENT '输入Token用量',
  completion_tokens  int           NOT NULL DEFAULT 0      COMMENT '输出Token用量',
  duration_ms        int           DEFAULT NULL            COMMENT '生成耗时（毫秒）',
  error_message      varchar(1000) DEFAULT NULL            COMMENT '失败原因（流式出错落库）',
  remark             varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by          varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time        datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by          varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time        datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag           char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (message_id),
  KEY idx_conversation (conversation_id, message_id),
  KEY idx_user (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI消息表';

-- ----------------------------
-- 内置数据：预置智能体（模型指向 2.0.1 内置 model 101/100）
-- ----------------------------
delete from ai_agent where agent_id between 100 and 199;
insert into ai_agent values(100, '平台使用助手', null, '平台', '解答本平台的使用问题',
  '`你是本平台的操作助手，只回答与本平台功能使用相关的问题；与平台无关的问题礼貌拒答。',
  101, null, null, null, 0,
  '你好！我是平台使用助手，关于系统功能的问题都可以问我。',
  '["怎么重置密码？","用户管理在哪？","怎么新增一个角色？"]',
  1, 1, 1, '0', '平台内置助手', 'admin', sysdate(), '', null, '0');
insert into ai_agent values(101, '写作助手', null, '效率', '文案润色与起草',
  '你是专业的中文写作助手，帮助用户润色和起草各类文案。',
  101, null, null, null, 0,
  '把需要润色或起草的内容发给我，我会给出改进版本。',
  '["帮我润色一段工作总结","写一封会议通知邮件"]',
  1, 1, 2, '0', '写作润色助手', 'admin', sysdate(), '', null, '0');

-- ----------------------------
-- 菜单：「AI对话」为顶级菜单（parent 0，与 AI管理并列——对话是用户侧功能，不属于管理目录）；
-- 「AI管理」图标 skill，避免与对话的 message 重复。2020 号段。
-- ----------------------------
delete from sys_menu where menu_id between 2020 and 2029;
insert into sys_menu values('2020', 'AI对话', '0', '10', 'chat', 'ai/chat/index', '', '', 1, 0, 'C', '0', '0', 'ai:chat:chat', 'message', 'admin', sysdate(), '', null, 'AI 对话页（顶级菜单）');
update sys_menu set icon = 'skill' where menu_id = 2000;
