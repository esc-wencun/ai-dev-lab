package com.ruoyi.ai.common;

/**
 * AI 模块业务常量（specs/2.0.3 收口）。
 *
 * 范围界定：只收【跨处复用或语义易错】的业务魔法值——状态机值、消息类型、协议名、
 * SSE 事件名、删除标志。以下几类刻意不抽：
 * - 注解字符串（@TableName / @PreAuthorize / @RequestMapping 路径）：注解要求编译期常量，
 *   且路径/权限串语义就地自明，抽走反而要跳转查阅；
 * - 日志与异常文案：就地可读性优先，抽成常量只会让 message 拼接更难读。
 *
 * @author ruoyi
 */
public final class AiConstants
{
    private AiConstants() {}

    // ---------- 通用状态（char(1) 列；语义对齐 RuoYi sys_normal_disable） ----------

    /** 状态：正常/启用 */
    public static final String STATUS_NORMAL = "0";

    /** 状态：停用 */
    public static final String STATUS_DISABLED = "1";

    /** 删除标志：存在 */
    public static final String DEL_FLAG_EXISTS = "0";

    /** 删除标志：已删除（软删） */
    public static final String DEL_FLAG_REMOVED = "2";

    // ---------- 对话域 ----------

    /** 消息类型：用户 */
    public static final String MSG_TYPE_USER = "user";

    /** 消息类型：助手 */
    public static final String MSG_TYPE_ASSISTANT = "assistant";

    /** 消息状态：生成中（占位行） */
    public static final int MSG_STATUS_GENERATING = 0;

    /** 消息状态：已完成 */
    public static final int MSG_STATUS_DONE = 1;

    /** 消息状态：出错 */
    public static final int MSG_STATUS_ERROR = 2;

    /** 消息状态：已取消（保留半截） */
    public static final int MSG_STATUS_CANCELLED = 3;

    /** SSE 事件信封字段：事件名 */
    public static final String SSE_FIELD_EVENT = "event";

    /** SSE 事件信封字段：载荷 */
    public static final String SSE_FIELD_PAYLOAD = "payload";

    /** SSE 事件：开始（assistant 占位消息） */
    public static final String SSE_EVENT_BEGIN = "begin";

    /** SSE 事件：增量 token */
    public static final String SSE_EVENT_TOKEN = "token";

    /** SSE 事件：结束（最终消息，含完成/出错/取消三态） */
    public static final String SSE_EVENT_DONE = "done";

    /** 思考过程增量（深度思考模型）：payload = { delta }，前端折叠区追加 */
    public static final String SSE_EVENT_REASONING = "reasoning";

    /** SSE token 载荷字段：增量文本 */
    public static final String SSE_FIELD_DELTA = "delta";

    // ---------- 协议 ----------

    /** 协议：OpenAI 兼容 */
    public static final String PROTOCOL_OPENAI = "openai";

    /** 协议：Ollama 本地 */
    public static final String PROTOCOL_OLLAMA = "ollama";

    /** 密钥环境变量占位符模式：${ENV_VAR}（解析见 AiModelResolver#resolveApiKey） */
    public static final String ENV_PLACEHOLDER_PATTERN = "^\\$\\{(\\w+)}$";

    /** 模型类型：对话（ai_model.model_type） */
    public static final int MODEL_TYPE_CHAT = 1;

    /** 模型类型：向量 embedding */
    public static final int MODEL_TYPE_EMBEDDING = 5;

    /** 模型类型：重排序 rerank */
    public static final int MODEL_TYPE_RERANK = 6;
}
