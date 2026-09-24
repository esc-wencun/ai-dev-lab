// 防重复提交 + 限流中间件（对位 Java @RepeatSubmit + RepeatSubmitInterceptor/SameUrlDataInterceptor、
// @RateLimiter + RateLimiterAspect + RedisConfig.limitScriptText）。
package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/pkg/response"
)

// RepeatSubmitOptions 对位 @RepeatSubmit 注解属性。
type RepeatSubmitOptions struct {
	Interval time.Duration // 间隔（毫秒），默认 5s（对位 interval() default 5000）
	Message  string        // 提示文案，默认"不允许重复提交，请稍候再试"
}

// PreventRepeatSubmit 防重复提交：同一 token 在 interval 内、同 URL、同参数的请求判为重复，
// 返回 annotation.message() 信封（HTTP 200 + body code 500，对位 AjaxResult.error(message)）。
// Redis key：repeat_submit:{url}{token}，值 {url: {params, time}}，TTL = interval（对位 SameUrlDataInterceptor）。
func PreventRepeatSubmit(rc *cache.RedisCache, opt RepeatSubmitOptions) gin.HandlerFunc {
	if opt.Interval <= 0 {
		opt.Interval = 5 * time.Second
	}
	if opt.Message == "" {
		opt.Message = "不允许重复提交，请稍候再试"
	}
	return func(c *gin.Context) {
		body, err := snapshotRequestBody(c)
		if err != nil {
			body = nil
		}
		nowParams := string(body)
		if strings.TrimSpace(nowParams) == "" {
			nowParams = c.Request.URL.RawQuery // body 为空时取 query（对位 getParameterMap 序列化）
		}

		url := c.Request.URL.RequestURI()
		submitKey := strings.TrimSpace(c.GetHeader(constant.Token)) // 对位 request.getHeader(header)
		cacheKey := cache.BuildKey(constant.RepeatSubmitKey, url, submitKey)

		type repeatEntry struct {
			Params string `json:"repeatParams"`
			Time   int64  `json:"repeatTime"`
		}
		var pre map[string]repeatEntry
		if err := rc.GetObject(c.Request.Context(), cacheKey, &pre); err == nil {
			if entry, ok := pre[url]; ok {
				if entry.Params == nowParams && time.Now().UnixMilli()-entry.Time < opt.Interval.Milliseconds() {
					response.Error(c, opt.Message).JSON()
					c.Abort()
					return
				}
			}
		}
		// 非重复（或 key 失效）：刷新缓存
		_ = rc.SetObject(c.Request.Context(), cacheKey, map[string]repeatEntry{
			url: {Params: nowParams, Time: time.Now().UnixMilli()},
		}, opt.Interval)
		c.Next()
	}
}

// limitScriptText 限流 Lua 脚本（对位 RedisConfig.limitScriptText 逐字移植）：
// 当前计数已超 count 时直接返回现值（不再 incr）；
// 否则自增，首个请求设置过期时间；返回自增后计数。
const limitScriptText = `local key = KEYS[1]
local count = tonumber(ARGV[1])
local time = tonumber(ARGV[2])
local current = redis.call('get', key);
if current and tonumber(current) > count then
    return tonumber(current);
end
current = redis.call('incr', key)
if tonumber(current) == 1 then
    redis.call('expire', key, time)
end
return tonumber(current);`

// LimitType 限流维度（对位 Java LimitType 枚举）。
type LimitType int

const (
	LimitDefault LimitType = iota // 全局维度：key = 前缀 + handler 标识
	LimitIP                       // IP 维度：key = 前缀 + ip-handler 标识
)

// RateLimiterOptions 对位 @RateLimiter 注解属性。
type RateLimiterOptions struct {
	Key       string    // 缓存 key 前缀之外的定制段（对位 key()，默认空）
	Time      int       // 时间窗口（秒），默认 60
	Count     int       // 窗口内最大请求数，默认 100
	LimitType LimitType // 限流维度，默认全局
}

// RateLimiter 限流：Lua 原子计数超限返回"访问过于频繁，请稍候再试"
// （HTTP 200 + body code 500，对位 ServiceException 经 GlobalExceptionHandler）。
// key：rate_limit:{Key}{ip-}{handlerIdent}，handlerIdent 由路由注册时的处理函数名提供（对位 类名-方法名）。
func RateLimiter(rc *cache.RedisCache, handlerIdent string, opt RateLimiterOptions) gin.HandlerFunc {
	if opt.Time <= 0 {
		opt.Time = 60
	}
	if opt.Count <= 0 {
		opt.Count = 100
	}
	return func(c *gin.Context) {
		var sb strings.Builder
		sb.WriteString(constant.RateLimitKey)
		if opt.Key != "" {
			sb.WriteString(opt.Key)
		}
		if opt.LimitType == LimitIP {
			sb.WriteString(c.ClientIP())
			sb.WriteString("-")
		}
		sb.WriteString(handlerIdent)
		combineKey := sb.String()

		// 业务代码不直接碰 client：Lua 经 RedisCache 门面执行
		number, err := rc.EvalInt(c.Request.Context(), limitScriptText, []string{combineKey}, opt.Count, opt.Time)
		if err != nil {
			// Redis 异常：对位 Java RuntimeException("服务器限流异常") 但不阻断业务——
			// 共享 Redis 抖动不应打挂接口，记录后放行（偏差：Java 会抛 500）
			response.Error(c, "服务器限流异常，请稍候再试").JSON()
			c.Abort()
			return
		}
		if number > int64(opt.Count) {
			response.Error(c, "访问过于频繁，请稍候再试").JSON()
			c.Abort()
			return
		}
		c.Next()
	}
}

// snapshotRequestBody 读取 body 并回放（防重中间件先读，handler 仍可正常 bind）。
func snapshotRequestBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return nil, nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
