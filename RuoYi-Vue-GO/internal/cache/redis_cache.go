// Package cache RedisCache 门面（对位 Java ruoyi-common RedisCache / Python 版 config/redis_cache.py）。
//
// 约定：
//   - 业务代码禁止直接持有 go-redis client，读写一律经本门面；
//   - key 前缀必须来自 CacheConstants 白名单（与 Java 版共享 Redis，键前缀逐字一致）；
//   - 对共享 Redis 禁用 KEYS，统一 SCAN 增量遍历。
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"ruoyi-vue-go/internal/common/constant"
)

// ErrNotFound key 不存在（对位 Java 返回 null 的场景，调用方用 errors.Is 判断）
var ErrNotFound = errors.New("redis: key 不存在")

// keyPrefixWhitelist 允许的 key 前缀 = CacheConstants 七前缀
var keyPrefixWhitelist = map[string]struct{}{
	constant.LoginTokenKey:   {},
	constant.CaptchaCodeKey:  {},
	constant.SysConfigKey:    {},
	constant.SysDictKey:      {},
	constant.RepeatSubmitKey: {},
	constant.RateLimitKey:    {},
	constant.PwdErrCntKey:    {},
}

// RedisCache go-redis client 门面（client 由 main 装配注入）
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// BuildKey 拼装业务 key：prefix 必须在 CacheConstants 白名单内，否则 panic（编程错误，测试期暴露）。
// 例：BuildKey(constant.CaptchaCodeKey, uuid) → "captcha_codes:<uuid>"。
func BuildKey(prefix string, parts ...string) string {
	if _, ok := keyPrefixWhitelist[prefix]; !ok {
		panic(fmt.Sprintf("redis key 前缀不在 CacheConstants 白名单: %q", prefix))
	}
	return prefix + strings.Join(parts, "")
}

// SetObject 序列化为 JSON 后写入；ttl 为 0 表示不过期。
// 分钟级 TTL 对齐 Java 习惯：调用方传 N*time.Minute（如 constant.CaptchaExpiration*time.Minute）。
func (c *RedisCache) SetObject(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("redis 序列化失败 key=%s: %w", key, err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// SetRaw 原样写入字符串值（不序列化）。用于测试期模拟 Java FastJson 写入的原始 JSON
// 及后续兼容解析场景；业务代码写结构化数据一律用 SetObject。
func (c *RedisCache) SetRaw(ctx context.Context, key, val string, ttl time.Duration) error {
	return c.client.Set(ctx, key, val, ttl).Err()
}

// GetObject 读取并反序列化 JSON 到 dest；key 不存在返回 ErrNotFound。
func (c *RedisCache) GetObject(ctx context.Context, key string, dest any) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("redis 反序列化失败 key=%s: %w", key, err)
	}
	return nil
}

// Delete 删除一个或多个 key，返回实际删除数量。
func (c *RedisCache) Delete(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return c.client.Del(ctx, keys...).Result()
}

// Expire 重设 key 的过期时间，key 不存在返回 false。
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, ttl).Result()
}

// HasKey 判断 key 是否存在。
func (c *RedisCache) HasKey(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// KeysByPrefix 返回指定前缀下全部 key。
// 前缀同样走 CacheConstants 白名单；用 SCAN 增量遍历替代 KEYS（共享生产 Redis，KEYS 会阻塞）。
func (c *RedisCache) KeysByPrefix(ctx context.Context, prefix string) ([]string, error) {
	if _, ok := keyPrefixWhitelist[prefix]; !ok {
		return nil, fmt.Errorf("redis key 前缀不在 CacheConstants 白名单: %q", prefix)
	}
	var keys []string
	iter := c.client.Scan(ctx, 0, prefix+"*", 500).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

// EvalInt 执行 Lua 脚本并返回整数结果（限流计数等场景；脚本经门面统一收口，业务不持有 client）。
func (c *RedisCache) EvalInt(ctx context.Context, script string, keys []string, args ...any) (int64, error) {
	return c.client.Eval(ctx, script, keys, args...).Int64()
}

// Info Redis 服务器信息（缓存监控面板；对位 RedisInfo 子集）。
func (c *RedisCache) Info(ctx context.Context) map[string]string {
	m, err := c.client.Info(ctx).Result()
	if err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, line := range strings.Split(m, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if kv := strings.SplitN(line, ":", 2); len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}
	return out
}

// DBSize 键总数。
func (c *RedisCache) DBSize(ctx context.Context) int64 {
	n, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return 0
	}
	return n
}
