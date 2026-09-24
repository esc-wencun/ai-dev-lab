// Package constant 通用常量（对位 Java com.ruoyi.common.constant）。
//
// 本文件对位 CacheConstants.java：缓存 key 前缀，与 Java 版逐字一致，
// 前缀拼装一律经 internal/cache RedisCache.BuildKey（白名单校验）。
package constant

// 缓存 key 前缀（Redis 键 = 前缀 + 业务标识，共享 Redis 与 Java 版互通，禁止改动字面值）
const (
	LoginTokenKey   = "login_tokens:"  // 登录用户 redis key
	CaptchaCodeKey  = "captcha_codes:" // 验证码 redis key
	SysConfigKey    = "sys_config:"    // 参数管理 cache key
	SysDictKey      = "sys_dict:"      // 字典管理 cache key
	RepeatSubmitKey = "repeat_submit:" // 防重提交 redis key
	RateLimitKey    = "rate_limit:"    // 限流 redis key
	PwdErrCntKey    = "pwd_err_cnt:"   // 登录账户密码错误次数 redis key
)
