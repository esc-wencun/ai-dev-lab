package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ruoyi-vue-go/internal/common/constant"
)

func newTestCache(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisCache(client), mr
}

func TestBuildKeyWhitelist(t *testing.T) {
	cases := []struct {
		prefix string
		parts  []string
		want   string
	}{
		{constant.LoginTokenKey, []string{"uuid-1"}, "login_tokens:uuid-1"},
		{constant.PwdErrCntKey, []string{"admin"}, "pwd_err_cnt:admin"},
		{constant.SysConfigKey, []string{"sys.account", ".captchaEnabled"}, "sys_config:sys.account.captchaEnabled"},
		{constant.RateLimitKey, nil, "rate_limit:"},
	}
	for _, c := range cases {
		if got := BuildKey(c.prefix, c.parts...); got != c.want {
			t.Errorf("BuildKey(%q, %v) = %q, want %q", c.prefix, c.parts, got, c.want)
		}
	}
}

func TestBuildKeyRejectsForeignPrefix(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("非 CacheConstants 前缀应 panic")
		}
	}()
	BuildKey("foreign:foo:", "x")
}

func TestSetGetObjectRoundTrip(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	type inner struct {
		City string `json:"city"`
	}
	type payload struct {
		Name    string    `json:"name"`
		Age     int       `json:"age"`
		LoginAt time.Time `json:"loginAt"`
		Inner   inner     `json:"inner"`
	}
	want := payload{
		Name:    "admin",
		Age:     18,
		LoginAt: time.Date(2026, 9, 24, 10, 30, 0, 0, time.Local),
		Inner:   inner{City: "上海"},
	}
	key := BuildKey(constant.LoginTokenKey, "uuid-1")
	if err := c.SetObject(ctx, key, want, time.Minute); err != nil {
		t.Fatalf("SetObject() error = %v", err)
	}

	var got payload
	if err := c.GetObject(ctx, key, &got); err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	if got != want || !got.LoginAt.Equal(want.LoginAt) {
		t.Errorf("往返不一致: got %+v, want %+v", got, want)
	}
}

func TestGetObjectEmptyString(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()
	key := BuildKey(constant.SysConfigKey, "empty.value")

	if err := c.SetObject(ctx, key, "", time.Minute); err != nil {
		t.Fatalf("SetObject(\"\") error = %v", err)
	}
	var got string
	if err := c.GetObject(ctx, key, &got); err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	if got != "" {
		t.Errorf("空串往返后 = %q, want \"\"", got)
	}
}

func TestGetObjectMissing(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	var got map[string]any
	err := c.GetObject(ctx, BuildKey(constant.LoginTokenKey, "no-such"), &got)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("key 不存在应返回 ErrNotFound, got %v", err)
	}
}

func TestSetObjectTTLMinutes(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()
	key := BuildKey(constant.CaptchaCodeKey, "uuid-2")

	if err := c.SetObject(ctx, key, "1234", constant.CaptchaExpiration*time.Minute); err != nil {
		t.Fatalf("SetObject() error = %v", err)
	}
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl <= 0 || ttl > constant.CaptchaExpiration*time.Minute {
		t.Errorf("TTL = %v, want (0, %v]", ttl, constant.CaptchaExpiration*time.Minute)
	}
}

func TestKeysByPrefixScan(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	// 种入目标前缀与干扰前缀
	for _, k := range []string{"sys_config:a", "sys_config:b", "sys_config:c"} {
		if err := mr.Set(k, "1"); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []string{"login_tokens:x", "sys_dict:y"} {
		if err := mr.Set(k, "1"); err != nil {
			t.Fatal(err)
		}
	}

	keys, err := c.KeysByPrefix(ctx, constant.SysConfigKey)
	if err != nil {
		t.Fatalf("KeysByPrefix() error = %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("KeysByPrefix 返回 %d 个 key, want 3: %v", len(keys), keys)
	}
	for _, k := range keys {
		if k[:len(constant.SysDictKey)] == constant.SysDictKey {
			t.Errorf("不应混入其他前缀 key: %q", k)
		}
	}
}

func TestKeysByPrefixRejectsForeignPrefix(t *testing.T) {
	c, _ := newTestCache(t)
	if _, err := c.KeysByPrefix(context.Background(), "job:run:"); err == nil {
		t.Error("非 CacheConstants 前缀应返回 error")
	}
}

func TestDeleteExpireHasKey(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()
	key := BuildKey(constant.RepeatSubmitKey, "k1")

	if err := c.SetObject(ctx, key, map[string]int{"n": 1}, 0); err != nil {
		t.Fatalf("SetObject() error = %v", err)
	}

	ok, err := c.HasKey(ctx, key)
	if err != nil || !ok {
		t.Errorf("HasKey = %v, %v; want true, nil", ok, err)
	}

	set, err := c.Expire(ctx, key, time.Minute)
	if err != nil || !set {
		t.Errorf("Expire = %v, %v; want true, nil", set, err)
	}

	n, err := c.Delete(ctx, key)
	if err != nil || n != 1 {
		t.Errorf("Delete = %d, %v; want 1, nil", n, err)
	}

	ok, err = c.HasKey(ctx, key)
	if err != nil || ok {
		t.Errorf("删除后 HasKey = %v, %v; want false, nil", ok, err)
	}

	if n, err := c.Delete(ctx); err != nil || n != 0 {
		t.Errorf("Delete(空参数) = %d, %v; want 0, nil", n, err)
	}
}
