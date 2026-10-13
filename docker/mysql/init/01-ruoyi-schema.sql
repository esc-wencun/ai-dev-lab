SET NAMES utf8mb4;

DROP TABLE IF EXISTS `ai_agent`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_agent` (
  `agent_id` bigint NOT NULL AUTO_INCREMENT COMMENT '智能体ID',
  `agent_name` varchar(100) NOT NULL COMMENT '智能体名称',
  `avatar` varchar(1000) DEFAULT NULL COMMENT '头像URL',
  `category` varchar(64) DEFAULT NULL COMMENT '分类（市场分组）',
  `description` varchar(500) DEFAULT NULL COMMENT '一句话描述',
  `system_prompt` text NOT NULL COMMENT '系统提示词',
  `model_id` bigint DEFAULT NULL COMMENT '默认模型ID（NULL=会话时另选）',
  `knowledge_ids` varchar(1024) DEFAULT NULL COMMENT '知识库ID数组（JSON）',
  `tool_ids` varchar(1024) DEFAULT NULL COMMENT '工具ID数组（JSON）',
  `mcp_ids` varchar(1024) DEFAULT NULL COMMENT 'MCP服务ID数组（JSON，关联ai_mcp）',
  `enable_thinking` tinyint NOT NULL DEFAULT '0' COMMENT '默认深度思考开关（0关 1开）',
  `welcome_message` varchar(1000) DEFAULT NULL COMMENT '开场白',
  `suggested_questions` varchar(1024) DEFAULT NULL COMMENT '建议问题数组（JSON，开场白chip）',
  `visibility` tinyint NOT NULL DEFAULT '0' COMMENT '可见性（0私有 1公开）',
  `user_id` bigint DEFAULT NULL COMMENT '归属用户（作者）',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '显示顺序',
  `status` char(1) NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (`agent_id`),
  KEY `idx_user` (`user_id`,`status`),
  KEY `idx_category` (`category`)
) ENGINE=InnoDB AUTO_INCREMENT=105 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI智能体表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_agent` VALUES (100,'平台使用助手','/profile/upload/2026/10/13/platform_20261013203404A001.png','平台','解答本平台的使用问题','你是本平台的操作助手，只回答与本平台功能使用相关的问题；与平台无关的问题礼貌拒答。\n',101,NULL,NULL,NULL,1,'你好！我是平台使用助手，关于系统功能的问题都可以问我。\n（//TODO：暂未接入 平台功能RAG,待开发）','[\"怎么重置密码？\",\"用户管理在哪？\",\"怎么新增一个角色？\"]',1,1,1,'0','平台内置助手','admin','2026-10-10 10:12:32','admin',NULL,'0');
INSERT INTO `ai_agent` VALUES (101,'写作助手','/profile/upload/2026/10/13/writing_20261013203404A002.png','效率','文案润色与起草','你是专业的中文写作助手，帮助用户润色和起草各类文案。',101,NULL,NULL,NULL,0,'把需要润色或起草的内容发给我，我会给出改进版本。','[\"帮我润色一段工作总结\",\"写一封会议通知邮件\"]',1,1,2,'0','写作润色助手','admin','2026-10-11 02:12:32','',NULL,'0');
DROP TABLE IF EXISTS `ai_api_key`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_api_key` (
  `key_id` bigint NOT NULL AUTO_INCREMENT COMMENT '密钥ID',
  `provider_id` bigint NOT NULL COMMENT '所属厂商ID',
  `key_name` varchar(64) NOT NULL COMMENT '密钥名称',
  `api_key` varchar(512) NOT NULL COMMENT '密钥内容（支持${ENV_VAR}占位符）',
  `base_url` varchar(500) DEFAULT NULL COMMENT 'Key专属网关地址（NULL用厂商默认）',
  `api_path` varchar(500) DEFAULT NULL COMMENT '二级地址：拼接在网关地址后的请求路径（NULL用协议默认路径）',
  `status` char(1) NOT NULL DEFAULT '0' COMMENT '状态（0启用 1停用）',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (`key_id`),
  KEY `idx_provider` (`provider_id`,`status`)
) ENGINE=InnoDB AUTO_INCREMENT=136 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI密钥表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_api_key` VALUES (100,101,'智谱官方Key','${glm_api_key}',NULL,NULL,'0',NULL,'admin','2026-10-11 00:56:36','',NULL,'本机环境变量 glm_api_key','0');
INSERT INTO `ai_api_key` VALUES (102,104,'本地无需密钥','local-no-key',NULL,NULL,'0',NULL,'admin','2026-10-11 00:56:36','',NULL,'Ollama 本地占位密钥','0');
INSERT INTO `ai_api_key` VALUES (103,121,'内网网关Key','${glm_api_key}',NULL,NULL,'0',NULL,'admin','2026-10-11 00:56:36','',NULL,'本机环境变量 glm_api_key','0');
INSERT INTO `ai_api_key` VALUES (131,100,'deepseek环境变量','${deepseek_api_key}',NULL,NULL,'0',NULL,'admin',NULL,'',NULL,NULL,'0');
DROP TABLE IF EXISTS `ai_conversation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_conversation` (
  `conversation_id` bigint NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` bigint NOT NULL COMMENT '归属用户',
  `title` varchar(256) NOT NULL COMMENT '会话标题',
  `agent_id` bigint DEFAULT NULL COMMENT '绑定的智能体（NULL=自由对话）',
  `model_id` bigint NOT NULL COMMENT '会话使用模型ID',
  `model_code` varchar(100) DEFAULT NULL COMMENT '模型标识（冗余，免联表）',
  `system_prompt` text COMMENT '会话级系统提示词（创建时从智能体复制）',
  `knowledge_ids` varchar(1024) DEFAULT NULL COMMENT '会话级知识库ID数组（JSON快照）',
  `tool_ids` varchar(1024) DEFAULT NULL COMMENT '会话级工具ID数组（JSON快照，创建时从智能体复制）',
  `mcp_ids` varchar(1024) DEFAULT NULL COMMENT '会话级MCP服务ID数组（JSON快照，创建时从智能体复制）',
  `enable_thinking` tinyint NOT NULL DEFAULT '0' COMMENT '深度思考开关快照（0关 1开）',
  `summary_upto_message_id` bigint DEFAULT NULL COMMENT '压缩水位线（context_summary覆盖到的最后一条消息ID；NULL=未压缩）',
  `temperature` double DEFAULT NULL COMMENT '温度覆盖（NULL=模型默认）',
  `max_tokens` int DEFAULT NULL COMMENT '最大输出Token覆盖（NULL=模型默认）',
  `max_contexts` int DEFAULT NULL COMMENT '历史消息条数上限覆盖（NULL=模型默认）',
  `pinned` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶（0否 1是）',
  `pinned_time` datetime DEFAULT NULL COMMENT '置顶时间',
  `last_message_time` datetime DEFAULT NULL COMMENT '最后消息时间（列表排序；建会话时初始化为创建时间防NULL沉底）',
  `message_count` int NOT NULL DEFAULT '0' COMMENT '消息总条数（冗余计数）',
  `context_msg_count` int NOT NULL DEFAULT '0' COMMENT '上下文活跃消息数（use_context=1条数，压缩触发依据）',
  `context_summary` text COMMENT '上下文压缩摘要（长对话优化）',
  `status` char(1) NOT NULL DEFAULT '0' COMMENT '状态（0正常 1归档）',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`conversation_id`),
  KEY `idx_user_pinned` (`user_id`,`pinned`,`last_message_time`),
  KEY `idx_agent` (`agent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=168 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI会话表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_conversation` VALUES (131,1,'你是哪个模型？',NULL,200,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 00:05:30',8,8,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (136,1,'你是哪个模型?能做什么',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-13 20:50:52',6,6,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (137,1,'怎么重置密码？',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-13 20:24:55',6,6,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (138,1,'平台使用助手',100,101,NULL,'你是本平台的操作助手，只回答与本平台功能使用相关的问题；与平台无关的问题礼貌拒答。\n',NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,0,NULL,'2026-10-13 20:34:13',2,2,NULL,'0','',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (143,1,'现在几点了',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 00:19:15',6,6,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (144,1,'今天天气怎么样？',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 00:19:48',2,2,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (165,1,'有 3 个传教士和 3 个食人族在河左岸',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 01:30:25',0,0,NULL,'0','',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (166,1,'新会话',NULL,201,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 01:30:29',0,0,NULL,'0','',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_conversation` VALUES (167,1,'三个逻辑学家走进酒吧，酒保问：\"你们三个',NULL,101,'GLM-5.3-Flash',NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,0,NULL,'2026-10-14 01:31:28',2,2,NULL,'0','',NULL,'admin',NULL,'0',NULL);
DROP TABLE IF EXISTS `ai_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_message` (
  `message_id` bigint NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `conversation_id` bigint NOT NULL COMMENT '所属会话ID',
  `reply_id` bigint DEFAULT NULL COMMENT '回复的消息ID（user/assistant配对；user消息为NULL）',
  `msg_type` varchar(16) NOT NULL COMMENT '消息类型（user/assistant/system）',
  `msg_status` tinyint NOT NULL DEFAULT '1' COMMENT '消息状态（0生成中 1完成 2出错 3已取消；启动时残留0置3）',
  `user_id` bigint NOT NULL COMMENT '触发用户',
  `agent_id` bigint DEFAULT NULL COMMENT '智能体ID（冗余，统计用）',
  `model_id` bigint DEFAULT NULL COMMENT '实际生成模型ID（中途切模型时与会话当前值不同）',
  `model_code` varchar(100) DEFAULT NULL COMMENT '模型标识（冗余，免联表）',
  `content` mediumtext COMMENT '消息内容（流式聚合后完整值）',
  `use_context` tinyint NOT NULL DEFAULT '1' COMMENT '是否纳入上下文拼装（0剔除但不删）',
  `segment_ids` varchar(1024) DEFAULT NULL COMMENT '引用的知识库分段ID数组（JSON，RAG溯源）',
  `web_search_pages` text COMMENT '联网来源（JSON：标题/URL/摘要）',
  `attachment_urls` text COMMENT '附件地址数组（JSON，多模态）',
  `prompt_tokens` int NOT NULL DEFAULT '0' COMMENT '输入Token用量',
  `completion_tokens` int NOT NULL DEFAULT '0' COMMENT '输出Token用量',
  `duration_ms` int DEFAULT NULL COMMENT '生成耗时（毫秒）',
  `error_message` varchar(1000) DEFAULT NULL COMMENT '失败原因（流式出错落库）',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`message_id`),
  KEY `idx_conversation` (`conversation_id`,`message_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=356 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI消息表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_message` VALUES (252,131,NULL,'user',1,1,NULL,200,'qwen3:4b','你是哪个模型？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (253,131,252,'assistant',3,1,NULL,200,'qwen3:4b','',1,NULL,NULL,NULL,0,0,NULL,'生成中断：推理内容超出 reasoning_content 列容量导致落库失败（列已扩容修复）','admin',NULL,'system',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (260,137,NULL,'user',1,1,NULL,201,'deepseek-flash','怎么重置密码？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (261,137,260,'assistant',1,1,NULL,201,'deepseek-flash','重置密码的步骤取决于你要重置的是哪个平台。通用方法是：\n\n1. 打开对应网站/APP的登录页  \n2. 点击 **“忘记密码 / 重置密码”**  \n3. 输入账号，通常是手机号、邮箱或用户名  \n4. 通过 **短信验证码 / 邮箱链接 / 安全问题** 验证身份  \n5. 设置新密码并确认  \n6. 用新密码重新登录\n\n如果你已经登录了：  \n进入 **设置 → 账号与安全 → 修改密码**。\n\n如果是这些情况，方法不同：\n- **微信/QQ**：登录页点“忘记密码”，按引导用手机号/好友辅助验证\n- **Apple ID**：在 Apple 官网或设备设置里选“忘记密码”\n- **Windows**：设置 → 账户 → 登录选项 → 密码 → 更改；忘了可用微软账户在线重置\n- **路由器**：长按 Reset 孔恢复出厂，再用默认密码登录\n- **公司/学校系统**：通常要找管理员或 IT 重置\n\n你具体是要重置哪个密码？告诉我平台，我可以给你对应步骤。',1,NULL,NULL,NULL,0,0,5921,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (262,137,NULL,'user',1,1,NULL,201,'deepseek-flash','你是哪个模型？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (263,137,262,'assistant',1,1,NULL,201,'deepseek-flash','我是 **DeepSeek 最新版模型**，由深度求索公司开发。  \n如果你想确认具体版本号，可以在当前产品的“关于”或版本信息里查看。',1,NULL,NULL,NULL,0,0,3644,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (264,131,NULL,'user',1,1,NULL,200,'qwen3:4b','你是哪个模型？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (265,131,264,'assistant',2,1,NULL,200,'qwen3:4b','',1,NULL,NULL,NULL,0,0,17,'com.openai.errors.OpenAIIoException: Request failed','admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (266,137,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','你是哪个模型》',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (267,137,266,'assistant',1,1,NULL,101,'GLM-5.3-Flash','抱歉，我需要更正之前的回答——那是不准确的。\n\n我是 **GLM**，由 **Z.ai（智谱）** 训练的大语言模型，并不是 DeepSeek。\n\n如果你还有其他问题，比如之前提到的密码重置，欢迎继续问我。',1,NULL,NULL,NULL,0,0,4527,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (268,131,NULL,'user',1,1,NULL,200,'qwen3:4b','你是哪个模型？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (269,131,268,'assistant',1,1,NULL,200,'qwen3:4b','你好！我是**Qwen3**，是通义千问系列的最新一代大语言模型。我能够回答问题、创作文字、编程、表达观点等，还支持多语言交流。如果你有任何问题或需要帮助，欢迎随时告诉我！ 😊',1,NULL,NULL,NULL,0,0,4134,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (270,138,NULL,'user',1,1,100,101,'GLM-5.3-Flash','1+1等于几？只回答数字',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (271,138,270,'assistant',1,1,100,101,'GLM-5.3-Flash','抱歉，我是平台操作助手，只能解答与本平台功能使用相关的问题。1+1属于数学问题，不在我的服务范围内，请您谅解。\n\n如果您在使用平台过程中遇到任何操作问题，欢迎随时咨询！',1,NULL,NULL,NULL,0,0,8592,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (274,136,NULL,'user',1,1,NULL,201,'deepseek-flash','你是哪个模型?能做什么',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (275,136,274,'assistant',2,1,NULL,201,'deepseek-flash','',1,NULL,NULL,NULL,0,0,1467,'com.openai.errors.UnexpectedStatusCodeException: 402: Insufficient Balance (request_id: 5217d304-bf85-49b0-a91a-e3de461d6d4e)','admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (276,136,NULL,'user',1,1,NULL,100,'glm-4.7','你是哪个模型?能做什么',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (277,136,276,'assistant',2,1,NULL,100,'glm-4.7','',1,NULL,NULL,NULL,0,0,205,'模型无响应：该模型可能不存在、已下线，或服务网关异常（可切换其他模型重试）','admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (278,136,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','你是哪个模型?能做什么',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (279,136,278,'assistant',1,1,NULL,101,'GLM-5.3-Flash','我是Z.ai开发的GLM大语言模型，很高兴与您交流。\n\n我可以回答问题、提供信息、参与创意写作、解释概念等各类语言任务。我的知识来源于大规模文本训练，能够进行多领域的对话交流，但我不会存储您的私人数据。虽然我很强大，但也有知识边界和能力限制。\n\n您有什么特定领域的问题想了解吗？或者有什么我可以帮助您的？',1,NULL,NULL,NULL,0,0,15639,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (294,131,NULL,'user',1,1,NULL,200,'qwen3:4b','今天是什么时间？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (295,131,294,'assistant',1,1,NULL,200,'qwen3:4b','我无法直接获取当前实时时间（因为我的响应是基于服务器时间，不是实时系统时间）。不过，你可以查看手机、电脑或网络时间！如果需要其他帮助，随时告诉我哦 😊',1,NULL,NULL,NULL,0,0,9700,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (298,143,NULL,'user',1,1,NULL,201,'deepseek-flash','现在几点了',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (299,143,298,'assistant',2,1,NULL,201,'deepseek-flash','',1,NULL,NULL,NULL,0,0,1629,'com.openai.errors.UnexpectedStatusCodeException: 402: Insufficient Balance (request_id: d04c3f56-9d22-478c-bb9f-3c5a02982068)','admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (300,143,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','现在几点了',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (301,143,300,'assistant',1,1,NULL,101,'GLM-5.3-Flash','抱歉，我无法获取实时信息，所以不知道现在的具体时间。\n\n您可以通过以下方式查看当前时间：\n- 📱 手机或电脑的时钟\n- ⌚ 手表\n- 🔍 在搜索引擎中搜索“现在时间”\n\n如果您有其他问题，比如需要计算时间段、设置日程安排等，我很乐意帮忙！',1,NULL,NULL,NULL,0,0,4433,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (302,143,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','你是不是傻？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (303,143,302,'assistant',1,1,NULL,101,'GLM-5.3-Flash','抱歉之前的回复可能让您感到困扰。我理解连续收到重复的回复确实会让人不耐烦。\n\n关于时间的问题，我确实无法获取实时信息，因为我是AI语言模型，没有时钟功能，也不连接互联网。这不是借口，只是说明这个限制是客观存在的。\n\n如果您愿意，我可以帮您做这些事情：\n- ✍️ 写作、翻译、总结\n- 🧮 数学计算、逻辑推理\n- 💡 回答知识性问题\n- 💬 陪聊解闷\n\n请问有什么我可以帮您的吗？',1,NULL,NULL,NULL,0,0,5539,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (304,144,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','今天天气怎么样？',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (305,144,304,'assistant',1,1,NULL,101,'GLM-5.3-Flash','抱歉，我无法获取实时天气信息，因为我不能联网查询数据，也不知道您所在的位置。\n\n如果您想了解今天的天气，可以：\n\n- **查看天气应用**：手机自带的天气App或墨迹天气等\n- **搜索引擎**：在百度、必应等搜索“天气”即可看到当地天气\n- **中国天气网**：www.weather.com.cn（中国气象局官方网站）\n- **语音助手**：询问手机上的语音助手\n\n不过我可以帮您解答其他问题，比如天气相关的知识、不同季节的穿衣建议等。有什么我可以帮您的吗？',1,NULL,NULL,NULL,0,0,22713,NULL,'admin',NULL,'admin',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (352,165,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','有 3 个传教士和 3 个食人族在河左岸，要全部过到右岸。\n船最多载 2 人，且船上必须有人才能移动。\n约束：任何一岸，只要该岸有传教士，传教士人数就不能少于食人族人数。\n\n请给出步数最少的完整渡河方案（每一步写清谁去、谁回），并证明为什么不可能更少。',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (353,165,352,'assistant',0,1,NULL,101,'GLM-5.3-Flash','',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (354,167,NULL,'user',1,1,NULL,101,'GLM-5.3-Flash','三个逻辑学家走进酒吧，酒保问：\"你们三个都要啤酒吗？\"\n第一个说：\"我不知道。\" 第二个说：\"我不知道。\" 第三个说：\"是的。\"\n请问第三个是如何确定的？逐步说明每个人的推理过程。',1,NULL,NULL,NULL,0,0,NULL,NULL,'admin',NULL,'',NULL,'0',NULL);
INSERT INTO `ai_message` VALUES (355,167,354,'assistant',1,1,NULL,101,'GLM-5.3-Flash','# 谜题的关键\n\n酒保问的不是三个人各自的问题，而是一个**复合判断**：“你们三个**都**要啤酒吗？”\n\n- 只有三个人**全部**要啤酒，正确答案才是“是”；\n- 只要**任何一个人**不要，正确答案就是“否”。\n\n每个人都知道自己想不想喝，但不知道别人的想法。下面逐步分析。\n\n---\n\n## 第一个逻辑学家\n\n- 他知道自己要啤酒，但不知道另外两人的想法。\n- **如果他自己不要啤酒**，那么“三人都不要”的前提就已被打破，他可以立刻确定地回答：“不是。”\n- 但他回答的是“**我不知道**”——说明他无法否定“三人都同意”的可能。\n\n**推论：第一个人要啤酒。** 他的“不知道”实际上暴露了这个信息。\n\n---\n\n## 第二个逻辑学家\n\n- 他听到了第一个人的回答，做出了与上面相同的推理，得知**第一个人要啤酒**。\n- 现在看自己：**如果他不要啤酒**，那么无论第三个人怎样，“三人都同意”必然为假，他可以直接回答：“不是。”\n- 但他也回答“**我不知道**”——说明他也无法否定这种可能。\n\n**推论：第二个人也要啤酒。**\n\n---\n\n## 第三个逻辑学家\n\n- 他听到了前两人的回答，推出：**前两人都要啤酒**。\n- 他知道自己**要啤酒**。\n- 三人都要 → 整体答案确定为“**是**”。\n\n所以他可以斩钉截铁地回答：“是的。”\n\n---\n\n## 核心洞察\n\n在这个问题里，“我不知道”并不是没有信息量的回答，而是一种**间接的肯定**：\n\n> 只要有人不要啤酒，他就能立刻断定答案是“否”。因此，说“不知道”的唯一可能，就是这个人心里的答案是“我要”。\n\n前两句“我不知道”传递了两个确定的事实，加上第三个人对自己的了解，拼出了完整的答案。\n\n（顺带一提：第三个人回答“是的”也隐含了他自己要啤酒——如果他不想要，他就必须回答“不是”了。）',1,NULL,NULL,NULL,0,0,26868,NULL,'admin',NULL,'admin',NULL,'0',NULL);
DROP TABLE IF EXISTS `ai_model`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_model` (
  `model_id` bigint NOT NULL AUTO_INCREMENT COMMENT '模型ID',
  `provider_id` bigint NOT NULL COMMENT '所属厂商ID',
  `key_id` bigint NOT NULL COMMENT '绑定密钥ID（须与厂商一致；同一模型可经多把Key建多条通道）',
  `model_name` varchar(100) NOT NULL COMMENT '模型名称（展示用）',
  `model_code` varchar(100) NOT NULL COMMENT '模型标识（传给API）',
  `model_type` tinyint NOT NULL DEFAULT '1' COMMENT '模型类型（1对话 2图片 3语音 4视频 5向量 6重排序）',
  `support_thinking` tinyint NOT NULL DEFAULT '0' COMMENT '支持深度思考（0否 1是）',
  `support_tool` tinyint NOT NULL DEFAULT '0' COMMENT '支持工具调用',
  `support_vision` tinyint NOT NULL DEFAULT '0' COMMENT '支持图片输入',
  `support_audio` tinyint NOT NULL DEFAULT '0' COMMENT '支持音频输入',
  `support_video` tinyint NOT NULL DEFAULT '0' COMMENT '支持视频输入',
  `model_dimension` int DEFAULT NULL COMMENT '向量维度（仅type=5时有意义）',
  `temperature` double DEFAULT NULL COMMENT '默认温度（NULL=供应商默认）',
  `max_tokens` int DEFAULT NULL COMMENT '默认最大输出Token',
  `max_contexts` int DEFAULT NULL COMMENT '默认携带历史消息条数上限',
  `status` char(1) NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '显示顺序（升序）',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注（选型建议）',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (`model_id`),
  UNIQUE KEY `uk_provider_key_model` (`provider_id`,`key_id`,`model_code`,`del_flag`),
  KEY `idx_key` (`key_id`),
  KEY `idx_type_status` (`model_type`,`status`)
) ENGINE=InnoDB AUTO_INCREMENT=202 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI模型表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_model` VALUES (101,121,103,'GLM-5.3-Flash','GLM-5.3-Flash',1,1,1,1,0,0,NULL,NULL,NULL,NULL,'0',2,'内网网关 GLM（thinking 模型，本机默认）','admin','2026-10-11 00:56:36','admin',NULL,'0');
INSERT INTO `ai_model` VALUES (200,104,102,'ollama-qwen3:4b','qwen3:4b',1,1,0,0,0,0,NULL,NULL,NULL,NULL,'0',0,NULL,'admin',NULL,'admin',NULL,'0');
INSERT INTO `ai_model` VALUES (201,100,131,'deepseek-flash','deepseek-flash',1,1,1,1,0,0,NULL,NULL,NULL,NULL,'0',0,NULL,'admin',NULL,'admin',NULL,'0');
DROP TABLE IF EXISTS `ai_provider`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_provider` (
  `provider_id` bigint NOT NULL AUTO_INCREMENT COMMENT '厂商ID',
  `provider_name` varchar(100) NOT NULL COMMENT '厂商名称',
  `provider_code` varchar(50) NOT NULL COMMENT '厂商编码（唯一）',
  `protocol` varchar(32) NOT NULL DEFAULT 'openai' COMMENT '协议类型（openai/anthropic/ollama/gemini/dify/coze/custom）',
  `api_host` varchar(500) NOT NULL COMMENT '默认API地址（必填，各厂商官方地址不同）',
  `icon` varchar(1000) DEFAULT NULL COMMENT '厂商图标',
  `website` varchar(1000) DEFAULT NULL COMMENT '官网地址',
  `key_apply_url` varchar(1000) DEFAULT NULL COMMENT '密钥申请地址',
  `status` char(1) NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '显示顺序',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_by` varchar(64) DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `del_flag` char(1) NOT NULL DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (`provider_id`),
  UNIQUE KEY `uk_provider_code` (`provider_code`,`del_flag`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=200 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='AI厂商表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `ai_provider` VALUES (100,'DeepSeek','deepseek','openai','https://api.deepseek.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/deepseek-color.png','https://www.deepseek.com','https://platform.deepseek.com/api_keys','0',1,'深度求索','admin','2026-10-11 00:00:21','',NULL,'0');
INSERT INTO `ai_provider` VALUES (101,'智谱开放平台','zhipu','openai','https://open.bigmodel.cn/api/paas','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/zhipu-color.png','https://open.bigmodel.cn','https://open.bigmodel.cn/usercenter/apikeys','0',2,'智谱 GLM 系列','admin','2026-10-11 00:00:21','',NULL,'0');
INSERT INTO `ai_provider` VALUES (102,'阿里百炼','bailian','openai','https://dashscope.aliyuncs.com/compatible-mode','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/qwen-color.png','https://www.aliyun.com/product/bailian','https://bailian.console.aliyun.com/?apiKey=1','0',3,'通义千问 / DashScope 兼容模式','admin','2026-10-11 00:00:21','',NULL,'0');
INSERT INTO `ai_provider` VALUES (103,'硅基流动','siliconflow','openai','https://api.siliconflow.cn','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/siliconcloud-color.png','https://siliconflow.cn','https://cloud.siliconflow.cn/account/ak','0',4,'一站式模型云','admin','2026-10-11 00:00:21','',NULL,'0');
INSERT INTO `ai_provider` VALUES (104,'Ollama（本地）','ollama','ollama','http://localhost:11434','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/ollama.png','https://ollama.com',NULL,'0',5,'本地部署，无需密钥','admin','2026-10-11 00:00:21','',NULL,'0');
INSERT INTO `ai_provider` VALUES (105,'文心一言','yiyan','openai','https://qianfan.baidubce.com/v2','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/wenxin-color.png',NULL,'https://console.bce.baidu.com/iam/#/iam/apikeylist','0',6,'百度千帆 v2（OpenAI 兼容）','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (106,'讯飞星火','xinghuo','openai','https://spark-api-open.xf-yun.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/spark-color.png',NULL,NULL,'0',7,'标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (107,'豆包','doubao','openai','https://ark.cn-beijing.volces.com/api/v3','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/doubao-color.png',NULL,'https://console.volcengine.com/ark','0',8,'火山方舟（OpenAI 兼容）','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (108,'混元','hunyuan','openai','https://api.hunyuan.cloud.tencent.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/hunyuan-color.png',NULL,NULL,'0',9,'腾讯；标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (109,'MiniMax','minimax','openai','https://api.minimax.chat','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/minimax-color.png',NULL,NULL,'0',10,'密钥需配二级地址 /v1/text/chatcompletion_v2','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (110,'月之暗面','moonshot','openai','https://api.moonshot.cn','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/moonshot.png',NULL,'https://platform.moonshot.cn/console/api-keys','0',11,'Kimi；标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (111,'百川智能','baichuan','openai','https://api.baichuan-ai.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/baichuan-color.png',NULL,NULL,'0',12,'标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (112,'阶跃星辰','stepfun','openai','https://api.stepfun.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/stepfun-color.png',NULL,'https://platform.stepfun.com/interface-key','0',13,'标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (113,'OpenAI','openai','openai','https://api.openai.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/openai.png','https://openai.com','https://platform.openai.com/api-keys','0',14,'OpenAI 官方','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (114,'Azure OpenAI','azureopenai','custom','https://your-resource.openai.azure.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/azure-color.png',NULL,NULL,'0',15,'微软；端点按资源域名配置，需专用适配器（本期仅登记）','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (115,'Anthropic','anthropic','anthropic','https://api.anthropic.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/anthropic.png','https://www.anthropic.com','https://console.anthropic.com/settings/keys','0',16,'Claude；密钥需配二级地址 /v1/messages（适配器后续扩展）','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (116,'Gemini','gemini','gemini','https://generativelanguage.googleapis.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/gemini-color.png',NULL,'https://aistudio.google.com/app/apikey','0',17,'谷歌；密钥需配二级地址（适配器后续扩展）','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (117,'StableDiffusion','stablediffusion','custom','https://api.stability.ai','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/stability-color.png',NULL,NULL,'0',18,'Stability AI，图片生成','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (118,'Midjourney','midjourney','custom','https://your-proxy.example.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/midjourney.png',NULL,NULL,'0',19,'官方无 API；常用 midjourney-proxy 反代，地址按部署填','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (119,'Suno','suno','custom','https://your-proxy.example.com','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/suno.png',NULL,NULL,'0',20,'Suno AI 音乐；常用第三方代理，地址按部署填','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (120,'Grok','grok','openai','https://api.x.ai','https://cdn.jsdelivr.net/npm/@lobehub/icons-static-png@latest/light/grok.png',NULL,'https://console.x.ai','0',21,'xAI；标准 /v1 兼容端点','admin','2026-10-11 00:43:10','',NULL,'0');
INSERT INTO `ai_provider` VALUES (121,'企业内网网关','airouter','openai','https://airouter.sinosoft.com.cn:9000',NULL,NULL,NULL,'0',0,'sinosoft 内网模型路由（OpenAI 兼容，/v1 已含在 api_host 末段——解析约定 host 不带尾斜杠 + /v1/chat/completions 拼接）','admin','2026-10-11 00:53:13','',NULL,'0');
DROP TABLE IF EXISTS `sys_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dept` (
  `dept_id` bigint NOT NULL AUTO_INCREMENT COMMENT '部门id',
  `parent_id` bigint DEFAULT '0' COMMENT '父部门id',
  `ancestors` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '祖级列表',
  `dept_name` varchar(30) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '部门名称',
  `order_num` int DEFAULT '0' COMMENT '显示顺序',
  `leader` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '负责人',
  `phone` varchar(11) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '联系电话',
  `email` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '邮箱',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '部门状态（0正常 1停用）',
  `del_flag` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`dept_id`)
) ENGINE=InnoDB AUTO_INCREMENT=901 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_dept` VALUES (100,0,'0','若依科技',0,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (101,100,'0,100','深圳总公司',1,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (102,100,'0,100','长沙分公司',2,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (103,101,'0,100,101','研发部门',1,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (104,101,'0,100,101','市场部门',2,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (105,101,'0,100,101','测试部门',3,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (106,101,'0,100,101','财务部门',4,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (107,101,'0,100,101','运维部门',5,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (108,102,'0,100,102','市场部门',1,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
INSERT INTO `sys_dept` VALUES (109,102,'0,100,102','财务部门',2,'若依','15888888888','ry@qq.com','0','0','admin','2026-09-24 20:07:05','',NULL);
DROP TABLE IF EXISTS `sys_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role` (
  `role_id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `role_name` varchar(30) COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色名称',
  `role_key` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色权限字符串',
  `role_sort` int NOT NULL COMMENT '显示顺序',
  `data_scope` char(1) COLLATE utf8mb4_general_ci DEFAULT '1' COMMENT '数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限）',
  `menu_check_strictly` tinyint(1) DEFAULT '1' COMMENT '菜单树选择项是否关联显示',
  `dept_check_strictly` tinyint(1) DEFAULT '1' COMMENT '部门树选择项是否关联显示',
  `status` char(1) COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色状态（0正常 1停用）',
  `del_flag` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=901 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色信息表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_role` VALUES (1,'超级管理员','admin',1,'1',1,1,'0','0','admin','2026-09-24 20:07:05','',NULL,'超级管理员');
INSERT INTO `sys_role` VALUES (2,'普通角色','common',2,'5',0,0,'0','0','admin','2026-09-24 20:07:05','admin','2026-09-28 16:33:43','普通角色');
DROP TABLE IF EXISTS `sys_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user` (
  `user_id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `dept_id` bigint DEFAULT NULL COMMENT '部门ID',
  `user_name` varchar(30) COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户账号',
  `nick_name` varchar(30) COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户昵称',
  `user_type` varchar(2) COLLATE utf8mb4_general_ci DEFAULT '00' COMMENT '用户类型（00系统用户）',
  `email` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '用户邮箱',
  `phonenumber` varchar(11) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '手机号码',
  `sex` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '用户性别（0男 1女 2未知）',
  `avatar` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '头像地址',
  `password` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '密码',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '账号状态（0正常 1停用）',
  `del_flag` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '删除标志（0代表存在 2代表删除）',
  `login_ip` varchar(128) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '最后登录IP',
  `login_date` datetime DEFAULT NULL COMMENT '最后登录时间',
  `pwd_update_date` datetime DEFAULT NULL COMMENT '密码最后更新时间',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=903 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户信息表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_user` VALUES (1,103,'admin','admin','00','ry@163.com','15888888888','0',NULL,'$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2','0','0','127.0.0.1','2026-10-14 00:36:27','2026-09-24 20:07:05','admin','2026-09-24 20:07:05','','2026-09-28 16:14:34','管理员');
INSERT INTO `sys_user` VALUES (2,105,'ry','若依','00','ry@qq.com','15666666666','0','','$2a$10$XM0bCPvDNmYeCVgBt8XyCOTtfYBClam.xQv9qvtdE5EOL0KSz2sEq','0','0','127.0.0.1','2026-09-26 18:52:25','2026-09-26 10:56:24','admin','2026-09-24 20:07:05','admin','2026-09-29 01:19:56','测试员');
INSERT INTO `sys_user` VALUES (103,104,'wencun','wencun','00','','13067229333','0','','$2a$10$zhBFR4MJlo67wwKybd5y/uOWnQHfjUbEPHJUGWExhD7ggbS3rWkEC','0','0','',NULL,NULL,'admin','2026-09-26 11:47:45','admin','2026-09-29 00:41:55',NULL);
DROP TABLE IF EXISTS `gen_table`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `gen_table` (
  `table_id` bigint NOT NULL AUTO_INCREMENT COMMENT '编号',
  `table_name` varchar(200) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '表名称',
  `table_comment` varchar(500) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '表描述',
  `sub_table_name` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '关联子表的表名',
  `sub_table_fk_name` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '子表关联的外键名',
  `class_name` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '实体类名称',
  `tpl_category` varchar(200) COLLATE utf8mb4_general_ci DEFAULT 'crud' COMMENT '使用的模板（crud单表操作 tree树表操作）',
  `tpl_web_type` varchar(30) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '前端模板类型（element-ui模版 element-plus模版）',
  `package_name` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '生成包路径',
  `module_name` varchar(30) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '生成模块名',
  `business_name` varchar(30) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '生成业务名',
  `function_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '生成功能名',
  `function_author` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '生成功能作者',
  `form_col_num` int DEFAULT '1' COMMENT '表单布局（单列 双列 三列）',
  `gen_type` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '生成代码方式（0zip压缩包 1自定义路径）',
  `gen_path` varchar(200) COLLATE utf8mb4_general_ci DEFAULT '/' COMMENT '生成路径（不填默认项目路径）',
  `options` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '其它生成选项',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`table_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='代码生成业务表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `gen_table_column`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `gen_table_column` (
  `column_id` bigint NOT NULL AUTO_INCREMENT COMMENT '编号',
  `table_id` bigint DEFAULT NULL COMMENT '归属表编号',
  `column_name` varchar(200) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '列名称',
  `column_comment` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '列描述',
  `column_type` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '列类型',
  `java_type` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT 'JAVA类型',
  `java_field` varchar(200) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT 'JAVA字段名',
  `is_pk` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否主键（1是）',
  `is_increment` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否自增（1是）',
  `is_required` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否必填（1是）',
  `is_insert` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否为插入字段（1是）',
  `is_edit` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否编辑字段（1是）',
  `is_list` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否列表字段（1是）',
  `is_query` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '是否查询字段（1是）',
  `query_type` varchar(200) COLLATE utf8mb4_general_ci DEFAULT 'EQ' COMMENT '查询方式（等于、不等于、大于、小于、范围）',
  `html_type` varchar(200) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '显示类型（文本框、文本域、下拉框、复选框、单选框、日期控件）',
  `dict_type` varchar(200) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典类型',
  `sort` int DEFAULT NULL COMMENT '排序',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`column_id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='代码生成业务表字段';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `sys_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_config` (
  `config_id` int NOT NULL AUTO_INCREMENT COMMENT '参数主键',
  `config_name` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '参数名称',
  `config_key` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '参数键名',
  `config_value` varchar(500) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '参数键值',
  `config_type` char(1) COLLATE utf8mb4_general_ci DEFAULT 'N' COMMENT '系统内置（Y是 N否）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`config_id`)
) ENGINE=InnoDB AUTO_INCREMENT=102 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='参数配置表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_config` VALUES (1,'主框架页-默认皮肤样式名称','sys.index.skinName','skin-blue','Y','admin','2026-09-24 20:07:08','',NULL,'蓝色 skin-blue、绿色 skin-green、紫色 skin-purple、红色 skin-red、黄色 skin-yellow');
INSERT INTO `sys_config` VALUES (2,'用户管理-账号初始密码','sys.user.initPassword','123456','Y','admin','2026-09-24 20:07:08','',NULL,'初始化密码 123456');
INSERT INTO `sys_config` VALUES (3,'主框架页-侧边栏主题','sys.index.sideTheme','theme-dark','Y','admin','2026-09-24 20:07:08','',NULL,'深色主题theme-dark，浅色主题theme-light');
INSERT INTO `sys_config` VALUES (4,'�˺���-��֤�뿪��','sys.account.captchaEnabled','false','Y','admin','2026-09-24 20:07:08','admin','2026-10-13 20:05:22','是否开启验证码功能（true开启，false关闭）');
INSERT INTO `sys_config` VALUES (5,'账号自助-是否开启用户注册功能','sys.account.registerUser','false','Y','admin','2026-09-24 20:07:08','',NULL,'是否开启注册用户功能（true开启，false关闭）');
INSERT INTO `sys_config` VALUES (6,'用户登录-黑名单列表','sys.login.blackIPList','','Y','admin','2026-09-24 20:07:08','',NULL,'设置登录IP黑名单限制，多个匹配项以;分隔，支持匹配（*通配、网段）');
INSERT INTO `sys_config` VALUES (7,'用户管理-初始密码修改策略','sys.account.initPasswordModify','1','Y','admin','2026-09-24 20:07:08','',NULL,'0：初始密码修改策略关闭，没有任何提示，1：提醒用户，如果未修改初始密码，则在登录时就会提醒修改密码对话框');
INSERT INTO `sys_config` VALUES (8,'用户管理-账号密码更新周期','sys.account.passwordValidateDays','0','Y','admin','2026-09-24 20:07:08','',NULL,'密码更新周期（填写数字，数据初始化值为0不限制，若修改必须为大于0小于365的正整数），如果超过这个周期登录系统时，则在登录时就会提醒修改密码对话框');
INSERT INTO `sys_config` VALUES (9,'用户管理-密码字符范围','sys.account.chrtype','0','Y','admin','2026-09-24 20:07:08','',NULL,'默认任意字符范围，0任意（密码可以输入任意字符），1数字（密码只能为0-9数字），2英文字母（密码只能为a-z和A-Z字母），3字母和数字（密码必须包含字母，数字）,4字母数字和特殊字符（目前支持的特殊字符包括：~!@#$%^&*()-=_+）');
DROP TABLE IF EXISTS `sys_dict_data`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict_data` (
  `dict_code` bigint NOT NULL AUTO_INCREMENT COMMENT '字典编码',
  `dict_sort` int DEFAULT '0' COMMENT '字典排序',
  `dict_label` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典标签',
  `dict_value` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典键值',
  `dict_type` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典类型',
  `css_class` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '样式属性（其他样式扩展）',
  `list_class` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '表格回显样式',
  `is_default` char(1) COLLATE utf8mb4_general_ci DEFAULT 'N' COMMENT '是否默认（Y是 N否）',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`dict_code`)
) ENGINE=InnoDB AUTO_INCREMENT=151 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='字典数据表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_dict_data` VALUES (1,1,'男','0','sys_user_sex','','','Y','0','admin','2026-09-24 20:07:08','',NULL,'性别男');
INSERT INTO `sys_dict_data` VALUES (2,2,'女','1','sys_user_sex','','','N','0','admin','2026-09-24 20:07:08','',NULL,'性别女');
INSERT INTO `sys_dict_data` VALUES (3,3,'未知','2','sys_user_sex','','','N','0','admin','2026-09-24 20:07:08','',NULL,'性别未知');
INSERT INTO `sys_dict_data` VALUES (4,1,'显示','0','sys_show_hide','','primary','Y','0','admin','2026-09-24 20:07:08','',NULL,'显示菜单');
INSERT INTO `sys_dict_data` VALUES (5,2,'隐藏','1','sys_show_hide','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'隐藏菜单');
INSERT INTO `sys_dict_data` VALUES (6,1,'正常','0','sys_normal_disable','','primary','Y','0','admin','2026-09-24 20:07:08','',NULL,'正常状态');
INSERT INTO `sys_dict_data` VALUES (7,2,'停用','1','sys_normal_disable','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'停用状态');
INSERT INTO `sys_dict_data` VALUES (8,1,'正常','0','sys_job_status','','primary','Y','0','admin','2026-09-24 20:07:08','',NULL,'正常状态');
INSERT INTO `sys_dict_data` VALUES (9,2,'暂停','1','sys_job_status','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'停用状态');
INSERT INTO `sys_dict_data` VALUES (10,1,'默认','DEFAULT','sys_job_group','','','Y','0','admin','2026-09-24 20:07:08','',NULL,'默认分组');
INSERT INTO `sys_dict_data` VALUES (11,2,'系统','SYSTEM','sys_job_group','','','N','0','admin','2026-09-24 20:07:08','',NULL,'系统分组');
INSERT INTO `sys_dict_data` VALUES (12,1,'是','Y','sys_yes_no','','primary','Y','0','admin','2026-09-24 20:07:08','',NULL,'系统默认是');
INSERT INTO `sys_dict_data` VALUES (13,2,'否','N','sys_yes_no','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'系统默认否');
INSERT INTO `sys_dict_data` VALUES (14,1,'通知','1','sys_notice_type','','warning','Y','0','admin','2026-09-24 20:07:08','',NULL,'通知');
INSERT INTO `sys_dict_data` VALUES (15,2,'公告','2','sys_notice_type','','success','N','0','admin','2026-09-24 20:07:08','',NULL,'公告');
INSERT INTO `sys_dict_data` VALUES (16,1,'正常','0','sys_notice_status','','primary','Y','0','admin','2026-09-24 20:07:08','',NULL,'正常状态');
INSERT INTO `sys_dict_data` VALUES (17,2,'关闭','1','sys_notice_status','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'关闭状态');
INSERT INTO `sys_dict_data` VALUES (18,99,'其他','0','sys_oper_type','','info','N','0','admin','2026-09-24 20:07:08','',NULL,'其他操作');
INSERT INTO `sys_dict_data` VALUES (19,1,'新增','1','sys_oper_type','','info','N','0','admin','2026-09-24 20:07:08','',NULL,'新增操作');
INSERT INTO `sys_dict_data` VALUES (20,2,'修改','2','sys_oper_type','','info','N','0','admin','2026-09-24 20:07:08','',NULL,'修改操作');
INSERT INTO `sys_dict_data` VALUES (21,3,'删除','3','sys_oper_type','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'删除操作');
INSERT INTO `sys_dict_data` VALUES (22,4,'授权','4','sys_oper_type','','primary','N','0','admin','2026-09-24 20:07:08','',NULL,'授权操作');
INSERT INTO `sys_dict_data` VALUES (23,5,'导出','5','sys_oper_type','','warning','N','0','admin','2026-09-24 20:07:08','',NULL,'导出操作');
INSERT INTO `sys_dict_data` VALUES (24,6,'导入','6','sys_oper_type','','warning','N','0','admin','2026-09-24 20:07:08','',NULL,'导入操作');
INSERT INTO `sys_dict_data` VALUES (25,7,'强退','7','sys_oper_type','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'强退操作');
INSERT INTO `sys_dict_data` VALUES (26,8,'生成代码','8','sys_oper_type','','warning','N','0','admin','2026-09-24 20:07:08','',NULL,'生成操作');
INSERT INTO `sys_dict_data` VALUES (27,9,'清空数据','9','sys_oper_type','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'清空操作');
INSERT INTO `sys_dict_data` VALUES (28,1,'成功','0','sys_common_status','','primary','N','0','admin','2026-09-24 20:07:08','',NULL,'正常状态');
INSERT INTO `sys_dict_data` VALUES (29,2,'失败','1','sys_common_status','','danger','N','0','admin','2026-09-24 20:07:08','',NULL,'停用状态');
INSERT INTO `sys_dict_data` VALUES (145,1,'对话','1','ai_model_type','','primary','Y','0','admin','2026-10-11 00:56:36','',NULL,'对话模型');
INSERT INTO `sys_dict_data` VALUES (146,2,'图片','2','ai_model_type','','success','N','0','admin','2026-10-11 00:56:36','',NULL,'图片生成模型');
INSERT INTO `sys_dict_data` VALUES (147,3,'语音','3','ai_model_type','','info','N','0','admin','2026-10-11 00:56:36','',NULL,'语音模型');
INSERT INTO `sys_dict_data` VALUES (148,4,'视频','4','ai_model_type','','warning','N','0','admin','2026-10-11 00:56:36','',NULL,'视频生成模型');
INSERT INTO `sys_dict_data` VALUES (149,5,'向量','5','ai_model_type','','danger','N','0','admin','2026-10-11 00:56:36','',NULL,'Embedding 模型');
INSERT INTO `sys_dict_data` VALUES (150,6,'重排序','6','ai_model_type','','','N','0','admin','2026-10-11 00:56:36','',NULL,'Rerank 模型');
DROP TABLE IF EXISTS `sys_dict_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict_type` (
  `dict_id` bigint NOT NULL AUTO_INCREMENT COMMENT '字典主键',
  `dict_name` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典名称',
  `dict_type` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '字典类型',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`dict_id`),
  UNIQUE KEY `dict_type` (`dict_type`)
) ENGINE=InnoDB AUTO_INCREMENT=110 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='字典类型表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_dict_type` VALUES (1,'用户性别','sys_user_sex','0','admin','2026-09-24 20:07:08','',NULL,'用户性别列表');
INSERT INTO `sys_dict_type` VALUES (2,'菜单状态','sys_show_hide','0','admin','2026-09-24 20:07:08','',NULL,'菜单状态列表');
INSERT INTO `sys_dict_type` VALUES (3,'系统开关','sys_normal_disable','0','admin','2026-09-24 20:07:08','',NULL,'系统开关列表');
INSERT INTO `sys_dict_type` VALUES (4,'任务状态','sys_job_status','0','admin','2026-09-24 20:07:08','',NULL,'任务状态列表');
INSERT INTO `sys_dict_type` VALUES (5,'任务分组','sys_job_group','0','admin','2026-09-24 20:07:08','',NULL,'任务分组列表');
INSERT INTO `sys_dict_type` VALUES (6,'系统是否','sys_yes_no','0','admin','2026-09-24 20:07:08','',NULL,'系统是否列表');
INSERT INTO `sys_dict_type` VALUES (7,'通知类型','sys_notice_type','0','admin','2026-09-24 20:07:08','',NULL,'通知类型列表');
INSERT INTO `sys_dict_type` VALUES (8,'通知状态','sys_notice_status','0','admin','2026-09-24 20:07:08','',NULL,'通知状态列表');
INSERT INTO `sys_dict_type` VALUES (9,'操作类型','sys_oper_type','0','admin','2026-09-24 20:07:08','',NULL,'操作类型列表');
INSERT INTO `sys_dict_type` VALUES (10,'系统状态','sys_common_status','0','admin','2026-09-24 20:07:08','',NULL,'登录状态列表');
INSERT INTO `sys_dict_type` VALUES (109,'AI模型类型','ai_model_type','0','admin','2026-10-11 00:56:36','',NULL,'AI 模型类型列表');
DROP TABLE IF EXISTS `sys_job`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_job` (
  `job_id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `job_name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '任务名称',
  `job_group` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'DEFAULT' COMMENT '任务组名',
  `invoke_target` varchar(500) COLLATE utf8mb4_general_ci NOT NULL COMMENT '调用目标字符串',
  `cron_expression` varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT 'cron执行表达式',
  `misfire_policy` varchar(20) COLLATE utf8mb4_general_ci DEFAULT '3' COMMENT '计划执行错误策略（1立即执行 2执行一次 3放弃执行）',
  `concurrent` char(1) COLLATE utf8mb4_general_ci DEFAULT '1' COMMENT '是否并发执行（0允许 1禁止）',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '状态（0正常 1暂停）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '备注信息',
  PRIMARY KEY (`job_id`,`job_name`,`job_group`)
) ENGINE=InnoDB AUTO_INCREMENT=101 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='定时任务调度表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_job` VALUES (1,'系统默认（无参）','DEFAULT','ryTask.ryNoParams','0/10 * * * * ?','3','1','1','admin','2026-09-24 20:07:08','',NULL,'');
INSERT INTO `sys_job` VALUES (2,'系统默认（有参）','DEFAULT','ryTask.ryParams(\'ry\')','0/15 * * * * ?','3','1','1','admin','2026-09-24 20:07:08','',NULL,'');
INSERT INTO `sys_job` VALUES (3,'系统默认（多参）','DEFAULT','ryTask.ryMultipleParams(\'ry\', true, 2000L, 316.50D, 100)','0 * * * * ?','3','1','1','admin','2026-09-24 20:07:08','admin','2026-09-24 23:36:03','');
DROP TABLE IF EXISTS `sys_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_menu` (
  `menu_id` bigint NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  `menu_name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL COMMENT '菜单名称',
  `parent_id` bigint DEFAULT '0' COMMENT '父菜单ID',
  `order_num` int DEFAULT '0' COMMENT '显示顺序',
  `path` varchar(200) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '路由地址',
  `component` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '组件路径',
  `query` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '路由参数',
  `route_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '路由名称',
  `is_frame` int DEFAULT '1' COMMENT '是否为外链（0是 1否）',
  `is_cache` int DEFAULT '0' COMMENT '是否缓存（0缓存 1不缓存）',
  `menu_type` char(1) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '菜单类型（M目录 C菜单 F按钮）',
  `visible` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '菜单状态（0显示 1隐藏）',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '菜单状态（0正常 1停用）',
  `perms` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '权限标识',
  `icon` varchar(100) COLLATE utf8mb4_general_ci DEFAULT '#' COMMENT '菜单图标',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2035 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='菜单权限表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_menu` VALUES (1,'系统管理',0,1,'system',NULL,'','',1,0,'M','0','0','','system','admin','2026-09-24 20:07:05','',NULL,'系统管理目录');
INSERT INTO `sys_menu` VALUES (2,'系统监控',0,2,'monitor',NULL,'','',1,0,'M','0','0','','monitor','admin','2026-09-24 20:07:05','',NULL,'系统监控目录');
INSERT INTO `sys_menu` VALUES (3,'系统工具',0,3,'tool',NULL,'','',1,0,'M','0','0','','tool','admin','2026-09-24 20:07:05','',NULL,'系统工具目录');
INSERT INTO `sys_menu` VALUES (4,'若依官网',0,4,'http://ruoyi.vip',NULL,'','',0,0,'M','0','0','','guide','admin','2026-09-24 20:07:05','',NULL,'若依官网地址');
INSERT INTO `sys_menu` VALUES (100,'用户管理',1,1,'user','system/user/index','','',1,0,'C','0','0','system:user:list','user','admin','2026-09-24 20:07:05','',NULL,'用户管理菜单');
INSERT INTO `sys_menu` VALUES (101,'角色管理',1,2,'role','system/role/index','','',1,0,'C','0','0','system:role:list','peoples','admin','2026-09-24 20:07:05','',NULL,'角色管理菜单');
INSERT INTO `sys_menu` VALUES (102,'菜单管理',1,3,'menu','system/menu/index','','',1,0,'C','0','0','system:menu:list','tree-table','admin','2026-09-24 20:07:05','',NULL,'菜单管理菜单');
INSERT INTO `sys_menu` VALUES (103,'部门管理',1,4,'dept','system/dept/index','','',1,0,'C','0','0','system:dept:list','tree','admin','2026-09-24 20:07:05','',NULL,'部门管理菜单');
INSERT INTO `sys_menu` VALUES (104,'岗位管理',1,5,'post','system/post/index','','',1,0,'C','0','0','system:post:list','post','admin','2026-09-24 20:07:05','',NULL,'岗位管理菜单');
INSERT INTO `sys_menu` VALUES (105,'字典管理',1,6,'dict','system/dict/index','','',1,0,'C','0','0','system:dict:list','dict','admin','2026-09-24 20:07:05','',NULL,'字典管理菜单');
INSERT INTO `sys_menu` VALUES (106,'参数设置',1,7,'config','system/config/index','','',1,0,'C','0','0','system:config:list','edit','admin','2026-09-24 20:07:05','',NULL,'参数设置菜单');
INSERT INTO `sys_menu` VALUES (107,'通知公告',1,8,'notice','system/notice/index','','',1,0,'C','0','0','system:notice:list','message','admin','2026-09-24 20:07:05','',NULL,'通知公告菜单');
INSERT INTO `sys_menu` VALUES (108,'日志管理',1,9,'log','','','',1,0,'M','0','0','','log','admin','2026-09-24 20:07:05','',NULL,'日志管理菜单');
INSERT INTO `sys_menu` VALUES (109,'在线用户',2,1,'online','monitor/online/index','','',1,0,'C','0','0','monitor:online:list','online','admin','2026-09-24 20:07:05','',NULL,'在线用户菜单');
INSERT INTO `sys_menu` VALUES (110,'定时任务',2,2,'job','monitor/job/index','','',1,0,'C','0','0','monitor:job:list','job','admin','2026-09-24 20:07:05','',NULL,'定时任务菜单');
INSERT INTO `sys_menu` VALUES (111,'数据监控',2,3,'druid','monitor/druid/index','','',1,0,'C','0','0','monitor:druid:list','druid','admin','2026-09-24 20:07:05','',NULL,'数据监控菜单');
INSERT INTO `sys_menu` VALUES (112,'服务监控',2,4,'server','monitor/server/index','','',1,0,'C','0','0','monitor:server:list','server','admin','2026-09-24 20:07:05','',NULL,'服务监控菜单');
INSERT INTO `sys_menu` VALUES (113,'缓存监控',2,5,'cache','monitor/cache/index','','',1,0,'C','0','0','monitor:cache:list','redis','admin','2026-09-24 20:07:05','',NULL,'缓存监控菜单');
INSERT INTO `sys_menu` VALUES (114,'缓存列表',2,6,'cacheList','monitor/cache/list','','',1,0,'C','0','0','monitor:cache:list','redis-list','admin','2026-09-24 20:07:05','',NULL,'缓存列表菜单');
INSERT INTO `sys_menu` VALUES (115,'表单构建',3,1,'build','tool/build/index','','',1,0,'C','0','0','tool:build:list','build','admin','2026-09-24 20:07:05','',NULL,'表单构建菜单');
INSERT INTO `sys_menu` VALUES (116,'代码生成',3,2,'gen','tool/gen/index','','',1,0,'C','0','0','tool:gen:list','code','admin','2026-09-24 20:07:05','',NULL,'代码生成菜单');
INSERT INTO `sys_menu` VALUES (117,'系统接口',3,3,'swagger','tool/swagger/index','','',1,0,'C','0','0','tool:swagger:list','swagger','admin','2026-09-24 20:07:05','',NULL,'系统接口菜单');
INSERT INTO `sys_menu` VALUES (500,'操作日志',108,1,'operlog','monitor/operlog/index','','',1,0,'C','0','0','monitor:operlog:list','form','admin','2026-09-24 20:07:05','',NULL,'操作日志菜单');
INSERT INTO `sys_menu` VALUES (501,'登录日志',108,2,'logininfor','monitor/logininfor/index','','',1,0,'C','0','0','monitor:logininfor:list','logininfor','admin','2026-09-24 20:07:05','',NULL,'登录日志菜单');
INSERT INTO `sys_menu` VALUES (1000,'用户查询',100,1,'','','','',1,0,'F','0','0','system:user:query','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1001,'用户新增',100,2,'','','','',1,0,'F','0','0','system:user:add','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1002,'用户修改',100,3,'','','','',1,0,'F','0','0','system:user:edit','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1003,'用户删除',100,4,'','','','',1,0,'F','0','0','system:user:remove','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1004,'用户导出',100,5,'','','','',1,0,'F','0','0','system:user:export','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1005,'用户导入',100,6,'','','','',1,0,'F','0','0','system:user:import','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1006,'重置密码',100,7,'','','','',1,0,'F','0','0','system:user:resetPwd','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1007,'角色查询',101,1,'','','','',1,0,'F','0','0','system:role:query','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1008,'角色新增',101,2,'','','','',1,0,'F','0','0','system:role:add','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1009,'角色修改',101,3,'','','','',1,0,'F','0','0','system:role:edit','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1010,'角色删除',101,4,'','','','',1,0,'F','0','0','system:role:remove','#','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_menu` VALUES (1011,'角色导出',101,5,'','','','',1,0,'F','0','0','system:role:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1012,'菜单查询',102,1,'','','','',1,0,'F','0','0','system:menu:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1013,'菜单新增',102,2,'','','','',1,0,'F','0','0','system:menu:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1014,'菜单修改',102,3,'','','','',1,0,'F','0','0','system:menu:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1015,'菜单删除',102,4,'','','','',1,0,'F','0','0','system:menu:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1016,'部门查询',103,1,'','','','',1,0,'F','0','0','system:dept:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1017,'部门新增',103,2,'','','','',1,0,'F','0','0','system:dept:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1018,'部门修改',103,3,'','','','',1,0,'F','0','0','system:dept:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1019,'部门删除',103,4,'','','','',1,0,'F','0','0','system:dept:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1020,'岗位查询',104,1,'','','','',1,0,'F','0','0','system:post:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1021,'岗位新增',104,2,'','','','',1,0,'F','0','0','system:post:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1022,'岗位修改',104,3,'','','','',1,0,'F','0','0','system:post:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1023,'岗位删除',104,4,'','','','',1,0,'F','0','0','system:post:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1024,'岗位导出',104,5,'','','','',1,0,'F','0','0','system:post:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1025,'字典查询',105,1,'#','','','',1,0,'F','0','0','system:dict:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1026,'字典新增',105,2,'#','','','',1,0,'F','0','0','system:dict:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1027,'字典修改',105,3,'#','','','',1,0,'F','0','0','system:dict:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1028,'字典删除',105,4,'#','','','',1,0,'F','0','0','system:dict:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1029,'字典导出',105,5,'#','','','',1,0,'F','0','0','system:dict:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1030,'参数查询',106,1,'#','','','',1,0,'F','0','0','system:config:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1031,'参数新增',106,2,'#','','','',1,0,'F','0','0','system:config:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1032,'参数修改',106,3,'#','','','',1,0,'F','0','0','system:config:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1033,'参数删除',106,4,'#','','','',1,0,'F','0','0','system:config:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1034,'参数导出',106,5,'#','','','',1,0,'F','0','0','system:config:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1035,'公告查询',107,1,'#','','','',1,0,'F','0','0','system:notice:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1036,'公告新增',107,2,'#','','','',1,0,'F','0','0','system:notice:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1037,'公告修改',107,3,'#','','','',1,0,'F','0','0','system:notice:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1038,'公告删除',107,4,'#','','','',1,0,'F','0','0','system:notice:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1039,'操作查询',500,1,'#','','','',1,0,'F','0','0','monitor:operlog:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1040,'操作删除',500,2,'#','','','',1,0,'F','0','0','monitor:operlog:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1041,'日志导出',500,3,'#','','','',1,0,'F','0','0','monitor:operlog:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1042,'登录查询',501,1,'#','','','',1,0,'F','0','0','monitor:logininfor:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1043,'登录删除',501,2,'#','','','',1,0,'F','0','0','monitor:logininfor:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1044,'日志导出',501,3,'#','','','',1,0,'F','0','0','monitor:logininfor:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1045,'账户解锁',501,4,'#','','','',1,0,'F','0','0','monitor:logininfor:unlock','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1046,'在线查询',109,1,'#','','','',1,0,'F','0','0','monitor:online:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1047,'批量强退',109,2,'#','','','',1,0,'F','0','0','monitor:online:batchLogout','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1048,'单条强退',109,3,'#','','','',1,0,'F','0','0','monitor:online:forceLogout','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1049,'任务查询',110,1,'#','','','',1,0,'F','0','0','monitor:job:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1050,'任务新增',110,2,'#','','','',1,0,'F','0','0','monitor:job:add','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1051,'任务修改',110,3,'#','','','',1,0,'F','0','0','monitor:job:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1052,'任务删除',110,4,'#','','','',1,0,'F','0','0','monitor:job:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1053,'状态修改',110,5,'#','','','',1,0,'F','0','0','monitor:job:changeStatus','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1054,'任务导出',110,6,'#','','','',1,0,'F','0','0','monitor:job:export','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1055,'生成查询',116,1,'#','','','',1,0,'F','0','0','tool:gen:query','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1056,'生成修改',116,2,'#','','','',1,0,'F','0','0','tool:gen:edit','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1057,'生成删除',116,3,'#','','','',1,0,'F','0','0','tool:gen:remove','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1058,'导入代码',116,4,'#','','','',1,0,'F','0','0','tool:gen:import','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1059,'预览代码',116,5,'#','','','',1,0,'F','0','0','tool:gen:preview','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (1060,'生成代码',116,6,'#','','','',1,0,'F','0','0','tool:gen:code','#','admin','2026-09-24 20:07:06','',NULL,'');
INSERT INTO `sys_menu` VALUES (2000,'AI管理',0,50,'ai',NULL,'','',1,0,'M','0','0','','skill','admin','2026-10-11 00:56:36','',NULL,'AI 能力管理目录');
INSERT INTO `sys_menu` VALUES (2001,'厂商管理',2000,1,'provider','ai/provider/index','','',1,0,'C','0','0','ai:provider:list','component','admin','2026-10-11 00:56:36','',NULL,'AI厂商菜单');
INSERT INTO `sys_menu` VALUES (2002,'密钥管理',2000,2,'key','ai/key/index','','',1,0,'C','0','0','ai:key:list','lock','admin','2026-10-11 00:56:36','',NULL,'AI密钥菜单');
INSERT INTO `sys_menu` VALUES (2003,'模型管理',2000,3,'model','ai/model/index','','',1,0,'C','0','0','ai:model:list','education','admin','2026-10-11 00:56:36','',NULL,'AI模型菜单');
INSERT INTO `sys_menu` VALUES (2004,'厂商查询',2001,1,'','','','',1,0,'F','0','0','ai:provider:query','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2005,'厂商新增',2001,2,'','','','',1,0,'F','0','0','ai:provider:add','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2006,'厂商修改',2001,3,'','','','',1,0,'F','0','0','ai:provider:edit','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2007,'厂商删除',2001,4,'','','','',1,0,'F','0','0','ai:provider:remove','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2008,'密钥查询',2002,1,'','','','',1,0,'F','0','0','ai:key:query','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2009,'密钥新增',2002,2,'','','','',1,0,'F','0','0','ai:key:add','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2010,'密钥修改',2002,3,'','','','',1,0,'F','0','0','ai:key:edit','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2011,'密钥删除',2002,4,'','','','',1,0,'F','0','0','ai:key:remove','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2012,'模型查询',2003,1,'','','','',1,0,'F','0','0','ai:model:query','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2013,'模型新增',2003,2,'','','','',1,0,'F','0','0','ai:model:add','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2014,'模型修改',2003,3,'','','','',1,0,'F','0','0','ai:model:edit','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2015,'模型删除',2003,4,'','','','',1,0,'F','0','0','ai:model:remove','#','admin','2026-10-11 00:56:36','',NULL,'');
INSERT INTO `sys_menu` VALUES (2020,'AI对话',0,10,'chat','ai/chat/index','','',1,0,'C','0','0','ai:chat:chat','message','admin','2026-10-11 02:12:32','',NULL,'AI 对话页（顶级菜单）');
INSERT INTO `sys_menu` VALUES (2030,'智能体管理',2000,5,'agent','ai/agent/index','','',1,0,'C','0','0','ai:agent:list','education','admin','2026-10-11 23:48:42','',NULL,'AI 智能体管理（管理页 2.0.4）');
INSERT INTO `sys_menu` VALUES (2031,'智能体查询',2030,1,'','','','',1,0,'F','0','0','ai:agent:query','#','admin','2026-10-11 23:48:42','',NULL,'');
INSERT INTO `sys_menu` VALUES (2032,'智能体新增',2030,2,'','','','',1,0,'F','0','0','ai:agent:add','#','admin','2026-10-11 23:48:42','',NULL,'');
INSERT INTO `sys_menu` VALUES (2033,'智能体修改',2030,3,'','','','',1,0,'F','0','0','ai:agent:edit','#','admin','2026-10-11 23:48:42','',NULL,'');
INSERT INTO `sys_menu` VALUES (2034,'智能体删除',2030,4,'','','','',1,0,'F','0','0','ai:agent:remove','#','admin','2026-10-11 23:48:42','',NULL,'');
DROP TABLE IF EXISTS `sys_notice`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_notice` (
  `notice_id` int NOT NULL AUTO_INCREMENT COMMENT '公告ID',
  `notice_title` varchar(50) COLLATE utf8mb4_general_ci NOT NULL COMMENT '公告标题',
  `notice_type` char(1) COLLATE utf8mb4_general_ci NOT NULL COMMENT '公告类型（1通知 2公告）',
  `notice_content` longblob COMMENT '公告内容',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '公告状态（0正常 1关闭）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`notice_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知公告表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_notice` VALUES (1,'温馨提醒：2018-07-01 若依新版本发布啦','2',0xE696B0E78988E69CACE58685E5AEB9,'0','admin','2026-09-24 20:07:09','',NULL,'管理员');
INSERT INTO `sys_notice` VALUES (2,'维护通知：2018-07-01 若依系统凌晨维护','1',0xE7BBB4E68AA4E58685E5AEB9,'0','admin','2026-09-24 20:07:09','',NULL,'管理员');
INSERT INTO `sys_notice` VALUES (3,'若依开源框架介绍','1',0x3C703E3C7370616E207374796C653D22636F6C6F723A20726762283233302C20302C2030293B223EE9A1B9E79BAEE4BB8BE7BB8D3C2F7370616E3E3C2F703E3C703E3C666F6E7420636F6C6F723D2223333333333333223E52756F5969E5BC80E6BA90E9A1B9E79BAEE698AFE4B8BAE4BC81E4B89AE794A8E688B7E5AE9AE588B6E79A84E5908EE58FB0E8849AE6898BE69EB6E6A186E69EB6EFBC8CE4B8BAE4BC81E4B89AE68993E980A0E79A84E4B880E7AB99E5BC8FE8A7A3E586B3E696B9E6A188EFBC8CE9998DE4BD8EE4BC81E4B89AE5BC80E58F91E68890E69CACEFBC8CE68F90E58D87E5BC80E58F91E69588E78E87E38082E4B8BBE8A681E58C85E68BACE794A8E688B7E7AEA1E79086E38081E8A792E889B2E7AEA1E79086E38081E983A8E997A8E7AEA1E79086E38081E88F9CE58D95E7AEA1E79086E38081E58F82E695B0E7AEA1E79086E38081E5AD97E585B8E7AEA1E79086E380813C2F666F6E743E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE5B297E4BD8DE7AEA1E790863C2F7370616E3E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE38081E5AE9AE697B6E4BBBBE58AA13C2F7370616E3E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE380813C2F7370616E3E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE69C8DE58AA1E79B91E68EA7E38081E799BBE5BD95E697A5E5BF97E38081E6938DE4BD9CE697A5E5BF97E38081E4BBA3E7A081E7949FE68890E7AD89E58A9FE883BDE38082E585B6E4B8ADEFBC8CE8BF98E694AFE68C81E5A49AE695B0E68DAEE6BA90E38081E695B0E68DAEE69D83E99990E38081E59BBDE99985E58C96E380815265646973E7BC93E5AD98E38081446F636B6572E983A8E7BDB2E38081E6BB91E58AA8E9AA8CE8AF81E7A081E38081E7ACACE4B889E696B9E8AEA4E8AF81E799BBE5BD95E38081E58886E5B883E5BC8FE4BA8BE58AA1E380813C2F7370616E3E3C666F6E7420636F6C6F723D2223333333333333223EE58886E5B883E5BC8FE69687E4BBB6E5AD98E582A83C2F666F6E743E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE38081E58886E5BA93E58886E8A1A8E5A484E79086E7AD89E68A80E69CAFE789B9E782B9E380823C2F7370616E3E3C2F703E3C703E3C696D67207372633D2268747470733A2F2F666F727564612E67697465652E636F6D2F696D616765732F313737333933313834383334323433393033322F61346432323331335F313831353039352E706E6722207374796C653D2277696474683A20363470783B223E3C62723E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A20726762283233302C20302C2030293B223EE5AE98E7BD91E58F8AE6BC94E7A4BA3C2F7370616E3E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE88BA5E4BE9DE5AE98E7BD91E59CB0E59D80EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F72756F79692E7669703C2F613E3C6120687265663D22687474703A2F2F72756F79692E76697022207461726765743D225F626C616E6B223E3C2F613E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE88BA5E4BE9DE69687E6A1A3E59CB0E59D80EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F646F632E72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F646F632E72756F79692E7669703C2F613E3C62723E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE6BC94E7A4BAE59CB0E59D80E38090E4B88DE58886E7A6BBE78988E38091EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F64656D6F2E72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F64656D6F2E72756F79692E7669703C2F613E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE6BC94E7A4BAE59CB0E59D80E38090E58886E7A6BBE78988E69CACE38091EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F7675652E72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F7675652E72756F79692E7669703C2F613E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE6BC94E7A4BAE59CB0E59D80E38090E5BEAEE69C8DE58AA1E78988E38091EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F636C6F75642E72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F636C6F75642E72756F79692E7669703C2F613E3C2F703E3C703E3C7370616E207374796C653D22636F6C6F723A207267622835312C2035312C203531293B223EE6BC94E7A4BAE59CB0E59D80E38090E7A7BBE58AA8E7ABAFE78988E38091EFBC9A266E6273703B3C2F7370616E3E3C6120687265663D22687474703A2F2F68352E72756F79692E76697022207461726765743D225F626C616E6B223E687474703A2F2F68352E72756F79692E7669703C2F613E3C2F703E3C703E3C6272207374796C653D22636F6C6F723A207267622834382C2034392C203531293B20666F6E742D66616D696C793A202671756F743B48656C766574696361204E6575652671756F743B2C2048656C7665746963612C20417269616C2C2073616E732D73657269663B20666F6E742D73697A653A20313270783B223E3C2F703E,'0','admin','2026-09-24 20:07:09','',NULL,'管理员');
DROP TABLE IF EXISTS `sys_notice_read`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_notice_read` (
  `read_id` bigint NOT NULL AUTO_INCREMENT COMMENT '已读主键',
  `notice_id` int NOT NULL COMMENT '公告id',
  `user_id` bigint NOT NULL COMMENT '用户id',
  `read_time` datetime NOT NULL COMMENT '阅读时间',
  PRIMARY KEY (`read_id`),
  UNIQUE KEY `uk_user_notice` (`user_id`,`notice_id`) COMMENT '同一用户同一公告只记录一次'
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='公告已读记录表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_notice_read` VALUES (1,3,2,'2026-09-26 18:52:37');
INSERT INTO `sys_notice_read` VALUES (5,3,1,'2026-09-29 01:28:42');
DROP TABLE IF EXISTS `sys_post`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_post` (
  `post_id` bigint NOT NULL AUTO_INCREMENT COMMENT '岗位ID',
  `post_code` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '岗位编码',
  `post_name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL COMMENT '岗位名称',
  `post_sort` int NOT NULL COMMENT '显示顺序',
  `status` char(1) COLLATE utf8mb4_general_ci NOT NULL COMMENT '状态（0正常 1停用）',
  `create_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`post_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='岗位信息表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_post` VALUES (1,'ceo','董事长',1,'0','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_post` VALUES (2,'se','项目经理',2,'0','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_post` VALUES (3,'hr','人力资源',3,'0','admin','2026-09-24 20:07:05','',NULL,'');
INSERT INTO `sys_post` VALUES (4,'user','普通员工',4,'0','admin','2026-09-24 20:07:05','',NULL,'');
DROP TABLE IF EXISTS `sys_role_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role_dept` (
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `dept_id` bigint NOT NULL COMMENT '部门ID',
  PRIMARY KEY (`role_id`,`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色和部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `sys_role_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role_menu` (
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `menu_id` bigint NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`role_id`,`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色和菜单关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_role_menu` VALUES (0,1);
INSERT INTO `sys_role_menu` VALUES (0,2);
INSERT INTO `sys_role_menu` VALUES (2,1);
INSERT INTO `sys_role_menu` VALUES (2,2);
INSERT INTO `sys_role_menu` VALUES (2,3);
INSERT INTO `sys_role_menu` VALUES (2,4);
INSERT INTO `sys_role_menu` VALUES (2,100);
INSERT INTO `sys_role_menu` VALUES (2,101);
INSERT INTO `sys_role_menu` VALUES (2,102);
INSERT INTO `sys_role_menu` VALUES (2,103);
INSERT INTO `sys_role_menu` VALUES (2,104);
INSERT INTO `sys_role_menu` VALUES (2,105);
INSERT INTO `sys_role_menu` VALUES (2,106);
INSERT INTO `sys_role_menu` VALUES (2,107);
INSERT INTO `sys_role_menu` VALUES (2,108);
INSERT INTO `sys_role_menu` VALUES (2,109);
INSERT INTO `sys_role_menu` VALUES (2,110);
INSERT INTO `sys_role_menu` VALUES (2,111);
INSERT INTO `sys_role_menu` VALUES (2,112);
INSERT INTO `sys_role_menu` VALUES (2,113);
INSERT INTO `sys_role_menu` VALUES (2,114);
INSERT INTO `sys_role_menu` VALUES (2,115);
INSERT INTO `sys_role_menu` VALUES (2,116);
INSERT INTO `sys_role_menu` VALUES (2,117);
INSERT INTO `sys_role_menu` VALUES (2,500);
INSERT INTO `sys_role_menu` VALUES (2,501);
INSERT INTO `sys_role_menu` VALUES (2,1000);
INSERT INTO `sys_role_menu` VALUES (2,1001);
INSERT INTO `sys_role_menu` VALUES (2,1002);
INSERT INTO `sys_role_menu` VALUES (2,1003);
INSERT INTO `sys_role_menu` VALUES (2,1004);
INSERT INTO `sys_role_menu` VALUES (2,1005);
INSERT INTO `sys_role_menu` VALUES (2,1006);
INSERT INTO `sys_role_menu` VALUES (2,1007);
INSERT INTO `sys_role_menu` VALUES (2,1008);
INSERT INTO `sys_role_menu` VALUES (2,1009);
INSERT INTO `sys_role_menu` VALUES (2,1010);
INSERT INTO `sys_role_menu` VALUES (2,1011);
INSERT INTO `sys_role_menu` VALUES (2,1012);
INSERT INTO `sys_role_menu` VALUES (2,1013);
INSERT INTO `sys_role_menu` VALUES (2,1014);
INSERT INTO `sys_role_menu` VALUES (2,1015);
INSERT INTO `sys_role_menu` VALUES (2,1016);
INSERT INTO `sys_role_menu` VALUES (2,1017);
INSERT INTO `sys_role_menu` VALUES (2,1018);
INSERT INTO `sys_role_menu` VALUES (2,1019);
INSERT INTO `sys_role_menu` VALUES (2,1020);
INSERT INTO `sys_role_menu` VALUES (2,1021);
INSERT INTO `sys_role_menu` VALUES (2,1022);
INSERT INTO `sys_role_menu` VALUES (2,1023);
INSERT INTO `sys_role_menu` VALUES (2,1024);
INSERT INTO `sys_role_menu` VALUES (2,1025);
INSERT INTO `sys_role_menu` VALUES (2,1026);
INSERT INTO `sys_role_menu` VALUES (2,1027);
INSERT INTO `sys_role_menu` VALUES (2,1028);
INSERT INTO `sys_role_menu` VALUES (2,1029);
INSERT INTO `sys_role_menu` VALUES (2,1030);
INSERT INTO `sys_role_menu` VALUES (2,1031);
INSERT INTO `sys_role_menu` VALUES (2,1032);
INSERT INTO `sys_role_menu` VALUES (2,1033);
INSERT INTO `sys_role_menu` VALUES (2,1034);
INSERT INTO `sys_role_menu` VALUES (2,1035);
INSERT INTO `sys_role_menu` VALUES (2,1036);
INSERT INTO `sys_role_menu` VALUES (2,1037);
INSERT INTO `sys_role_menu` VALUES (2,1038);
INSERT INTO `sys_role_menu` VALUES (2,1039);
INSERT INTO `sys_role_menu` VALUES (2,1040);
INSERT INTO `sys_role_menu` VALUES (2,1041);
INSERT INTO `sys_role_menu` VALUES (2,1042);
INSERT INTO `sys_role_menu` VALUES (2,1043);
INSERT INTO `sys_role_menu` VALUES (2,1044);
INSERT INTO `sys_role_menu` VALUES (2,1045);
INSERT INTO `sys_role_menu` VALUES (2,1046);
INSERT INTO `sys_role_menu` VALUES (2,1047);
INSERT INTO `sys_role_menu` VALUES (2,1048);
INSERT INTO `sys_role_menu` VALUES (2,1049);
INSERT INTO `sys_role_menu` VALUES (2,1050);
INSERT INTO `sys_role_menu` VALUES (2,1051);
INSERT INTO `sys_role_menu` VALUES (2,1052);
INSERT INTO `sys_role_menu` VALUES (2,1053);
INSERT INTO `sys_role_menu` VALUES (2,1054);
INSERT INTO `sys_role_menu` VALUES (2,1055);
INSERT INTO `sys_role_menu` VALUES (2,1056);
INSERT INTO `sys_role_menu` VALUES (2,1057);
INSERT INTO `sys_role_menu` VALUES (2,1058);
INSERT INTO `sys_role_menu` VALUES (2,1059);
INSERT INTO `sys_role_menu` VALUES (2,1060);
DROP TABLE IF EXISTS `sys_user_post`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_post` (
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `post_id` bigint NOT NULL COMMENT '岗位ID',
  PRIMARY KEY (`user_id`,`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与岗位关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_user_post` VALUES (1,1);
DROP TABLE IF EXISTS `sys_user_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_role` (
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户和角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `sys_user_role` VALUES (1,1);
DROP TABLE IF EXISTS `sys_job_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_job_log` (
  `job_log_id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务日志ID',
  `job_name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务名称',
  `job_group` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务组名',
  `invoke_target` varchar(500) COLLATE utf8mb4_general_ci NOT NULL COMMENT '调用目标字符串',
  `job_message` varchar(500) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '日志信息',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '执行状态（0正常 1失败）',
  `exception_info` varchar(2000) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '异常信息',
  `start_time` datetime DEFAULT NULL COMMENT '执行开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '执行结束时间',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`job_log_id`)
) ENGINE=InnoDB AUTO_INCREMENT=120 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='定时任务调度日志表';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `sys_logininfor`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_logininfor` (
  `info_id` bigint NOT NULL AUTO_INCREMENT COMMENT '访问ID',
  `user_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '用户账号',
  `ipaddr` varchar(128) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '登录IP地址',
  `login_location` varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '登录地点',
  `browser` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '浏览器类型',
  `os` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '操作系统',
  `status` char(1) COLLATE utf8mb4_general_ci DEFAULT '0' COMMENT '登录状态（0成功 1失败）',
  `msg` varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '提示消息',
  `login_time` datetime DEFAULT NULL COMMENT '访问时间',
  PRIMARY KEY (`info_id`),
  KEY `idx_sys_logininfor_s` (`status`),
  KEY `idx_sys_logininfor_lt` (`login_time`)
) ENGINE=InnoDB AUTO_INCREMENT=360 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='系统访问记录';
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `sys_oper_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_oper_log` (
  `oper_id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志主键',
  `title` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '模块标题',
  `business_type` int DEFAULT '0' COMMENT '业务类型（0其它 1新增 2修改 3删除）',
  `method` varchar(200) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '方法名称',
  `request_method` varchar(10) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '请求方式',
  `operator_type` int DEFAULT '0' COMMENT '操作类别（0其它 1后台用户 2手机端用户）',
  `oper_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '操作人员',
  `dept_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '部门名称',
  `oper_url` varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '请求URL',
  `oper_ip` varchar(128) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '主机地址',
  `oper_location` varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '操作地点',
  `oper_param` varchar(2000) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '请求参数',
  `json_result` varchar(2000) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '返回参数',
  `status` int DEFAULT '0' COMMENT '操作状态（0正常 1异常）',
  `error_msg` varchar(2000) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '错误消息',
  `oper_time` datetime DEFAULT NULL COMMENT '操作时间',
  `cost_time` bigint DEFAULT '0' COMMENT '消耗时间',
  PRIMARY KEY (`oper_id`),
  KEY `idx_sys_oper_log_bt` (`business_type`),
  KEY `idx_sys_oper_log_s` (`status`),
  KEY `idx_sys_oper_log_ot` (`oper_time`)
) ENGINE=InnoDB AUTO_INCREMENT=177 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='操作日志记录';
/*!40101 SET character_set_client = @saved_cs_client */;
