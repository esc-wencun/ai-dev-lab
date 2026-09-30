<?php
declare(strict_types=1);

/**
 * Redis 键前缀与业务常量（RuoYi-TP 自有命名，集中管理）
 *
 * 与经典版无对齐义务（自用 Redis db1），但业务代码禁止手拼键名，
 * 一律引用本常量由 RedisCache 门面拼装。
 */
final class TpConstant
{
    /** 会话键前缀：session:<uuid> */
    public const PREFIX_SESSION      = 'session:';
    /** 验证码：存自建会话内，非独立键（对位经典版 Session attribute） */
    /** 密码重试计数：pwd_retry:<loginName>，5 次锁 10 分钟 */
    public const PREFIX_PWD_RETRY    = 'pwd_retry:';
    /** 防重复提交：repeat_submit:<user>:<url+参数摘要> */
    public const PREFIX_REPEAT_SUBMIT = 'repeat_submit:';
    /** 限流：rate_limit:<key> */
    public const PREFIX_RATE_LIMIT   = 'rate_limit:';
    /** sys_config 缓存：config:<key> */
    public const PREFIX_CONFIG       = 'config:';
    /** sys_dict 缓存：dict:<type> */
    public const PREFIX_DICT         = 'dict:';
    /** 定时任务触发幂等键：jobfire:<jobId>:<YmdHis>（10.0.0；SETNX TTL 120s 防同 tick/双调度进程双跑） */
    public const PREFIX_JOB_FIRE     = 'jobfire:';
    /** 定时任务执行锁：joblock:<jobId>（10.0.0；SETNX TTL 300s 跨进程互斥 run 端点与调度触发） */
    public const PREFIX_JOB_LOCK     = 'joblock:';

    /** 密码最大错误次数（对位经典版 user.password.maxRetryCount: 5） */
    public const PWD_MAX_RETRY = 5;
    /** 密码锁定时长（秒，对位 ehcache loginRecordCache tti 10 分钟） */
    public const PWD_LOCK_SECONDS = 600;
    /** 会话空闲超时（秒，对位 shiro.session.expireTime: 30 分钟） */
    public const SESSION_IDLE_SECONDS = 1800;
    /** 续期阈值：剩余 TTL 低于此值时续期（秒，20 分钟） */
    public const SESSION_RENEW_THRESHOLD = 1200;
    /** 防重复提交间隔（秒） */
    public const REPEAT_SUBMIT_INTERVAL = 1;

    /** AjaxResult code：成功（对位经典版 AjaxResult.Type.SUCCESS） */
    public const CODE_SUCCESS = 0;
    /** AjaxResult code：警告（对位 WARN） */
    public const CODE_WARN = 301;
    /** AjaxResult code：错误（对位 ERROR） */
    public const CODE_ERROR = 500;

    private function __construct()
    {
    }
}
