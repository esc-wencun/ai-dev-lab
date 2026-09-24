package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env.test")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写临时配置文件失败: %v", err)
	}
	return path
}

const fullEnv = `
DB_HOST=db.example.com
DB_PORT=3307
DB_USERNAME=testdb
DB_PASSWORD=secret123
DB_DATABASE=ry-test
DB_ECHO=true
REDIS_HOST=redis.example.com
REDIS_PORT=10087
REDIS_PASSWORD=redispass
REDIS_DATABASE=11
APP_PORT=9090
APP_MODE=release
TOKEN_HEADER=X-Token
TOKEN_SECRET=jwt-secret
TOKEN_EXPIRE_MINUTES=60
CAPTCHA_TYPE=char
`

func TestLoadOK(t *testing.T) {
	cfg, err := Load(writeEnvFile(t, fullEnv))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != 9090 {
		t.Errorf("App.Port = %d, want 9090", cfg.App.Port)
	}
	if cfg.App.Mode != "release" {
		t.Errorf("App.Mode = %q, want release", cfg.App.Mode)
	}
	if cfg.DB.Host != "db.example.com" || cfg.DB.Port != 3307 ||
		cfg.DB.Username != "testdb" || cfg.DB.Password != "secret123" ||
		cfg.DB.Database != "ry-test" || !cfg.DB.Echo {
		t.Errorf("DB 配置解析不符: %+v", cfg.DB)
	}
	if cfg.Redis.Host != "redis.example.com" || cfg.Redis.Port != 10087 ||
		cfg.Redis.Password != "redispass" || cfg.Redis.Database != 11 {
		t.Errorf("Redis 配置解析不符: %+v", cfg.Redis)
	}
	if cfg.Token.Header != "X-Token" || cfg.Token.Secret != "jwt-secret" || cfg.Token.ExpireMinutes != 60 {
		t.Errorf("Token 配置解析不符: %+v", cfg.Token)
	}
	if cfg.Captcha.Type != "char" {
		t.Errorf("Captcha.Type = %q, want char", cfg.Captcha.Type)
	}
}

func TestLoadDefaults(t *testing.T) {
	// 只给必填项，其余走默认值
	cfg, err := Load(writeEnvFile(t, `
DB_HOST=db.example.com
DB_USERNAME=testdb
DB_PASSWORD=secret123
DB_DATABASE=ry-test
REDIS_HOST=redis.example.com
TOKEN_SECRET=jwt-secret
`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != 8080 {
		t.Errorf("App.Port 默认值 = %d, want 8080", cfg.App.Port)
	}
	if cfg.App.Mode != "debug" {
		t.Errorf("App.Mode 默认值 = %q, want debug", cfg.App.Mode)
	}
	if cfg.DB.Port != 3306 {
		t.Errorf("DB.Port 默认值 = %d, want 3306", cfg.DB.Port)
	}
	if cfg.DB.Echo {
		t.Error("DB.Echo 默认值应为 false")
	}
	if cfg.Redis.Port != 6379 || cfg.Redis.Database != 0 {
		t.Errorf("Redis 默认值不符: %+v", cfg.Redis)
	}
	if cfg.Token.Header != "Authorization" || cfg.Token.ExpireMinutes != 30 {
		t.Errorf("Token 默认值不符: %+v", cfg.Token)
	}
	if cfg.Captcha.Type != "math" {
		t.Errorf("Captcha.Type 默认值 = %q, want math", cfg.Captcha.Type)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	// 故意缺 DB_PASSWORD 与 REDIS_HOST
	_, err := Load(writeEnvFile(t, `
DB_HOST=db.example.com
DB_USERNAME=testdb
DB_DATABASE=ry-test
TOKEN_SECRET=jwt-secret
`))
	if err == nil {
		t.Fatal("缺失必填项时应返回 error")
	}
	for _, key := range []string{"DB_PASSWORD", "REDIS_HOST"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("错误信息应包含缺失项 %s，实际: %v", key, err)
		}
	}
}

func TestLoadFileNotExist(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "not-exist.env")); err == nil {
		t.Fatal("配置文件不存在时应返回 error")
	}
}
