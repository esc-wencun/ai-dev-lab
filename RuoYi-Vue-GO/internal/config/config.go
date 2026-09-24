// Package config 配置加载（对位 Python 版 config/env_config + .env.dev / Java 版 application.yml）。
//
// 字段命名继承 Python 版 .env.dev 风格（DB_HOST / REDIS_HOST / APP_PORT ...）；
// 数据库 / Redis / Token / 验证码的值以 Java 版 RuoYi-Vue 的
// application.yml 与 application-druid.yml 为准手动同步，Java 端改动后需同步本目录。
package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构（main.go 装配后显式注入各依赖，不用全局变量持有句柄）。
type Config struct {
	App     App
	DB      DB
	Redis   Redis
	Token   Token
	Captcha Captcha
}

// App 应用运行配置。
type App struct {
	Port int    // 监听端口，默认 8080（三版互斥，见 AGENTS.md 环境约束）
	Mode string // gin 模式：debug / release
}

// DB MySQL 连接配置（与 Java 版共用同一个阿里云 RDS 库）。
type DB struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Echo     bool // 是否输出 SQL 日志（对位 Python DB_ECHO）
}

// Redis 连接配置（与 Java 版共用，db11）。
type Redis struct {
	Host     string
	Port     int
	Username string
	Password string
	Database int
}

// Token JWT 配置（对位 Java application.yml 的 token 段，HS512 见 2.0.0-登录闭环）。
type Token struct {
	Header        string // 请求头名称，Java 版为 Authorization
	Secret        string // HS512 密钥，必须与 Java 版一致
	ExpireMinutes int    // 过期分钟数，Java 版 expireTime: 30
}

// Captcha 验证码配置（对位 Java application.yml 的 captchaType）。
type Captcha struct {
	Type string // math / char
}

// Load 从 dotenv 文件加载配置并校验必填项。path 形如 configs/.env.dev。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("env")
	v.AutomaticEnv() // 允许真实环境变量覆盖文件值

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{
		App: App{
			Port: getInt(v, "APP_PORT", 8080),
			Mode: getString(v, "APP_MODE", "debug"),
		},
		DB: DB{
			Host:     v.GetString("DB_HOST"),
			Port:     getInt(v, "DB_PORT", 3306),
			Username: v.GetString("DB_USERNAME"),
			Password: v.GetString("DB_PASSWORD"),
			Database: v.GetString("DB_DATABASE"),
			Echo:     v.GetBool("DB_ECHO"),
		},
		Redis: Redis{
			Host:     v.GetString("REDIS_HOST"),
			Port:     getInt(v, "REDIS_PORT", 6379),
			Username: v.GetString("REDIS_USERNAME"),
			Password: v.GetString("REDIS_PASSWORD"),
			Database: getInt(v, "REDIS_DATABASE", 0),
		},
		Token: Token{
			Header:        getString(v, "TOKEN_HEADER", "Authorization"),
			Secret:        v.GetString("TOKEN_SECRET"),
			ExpireMinutes: getInt(v, "TOKEN_EXPIRE_MINUTES", 30),
		},
		Captcha: Captcha{
			Type: getString(v, "CAPTCHA_TYPE", "math"),
		},
	}

	required := map[string]string{
		"DB_HOST":      cfg.DB.Host,
		"DB_USERNAME":  cfg.DB.Username,
		"DB_PASSWORD":  cfg.DB.Password,
		"DB_DATABASE":  cfg.DB.Database,
		"REDIS_HOST":   cfg.Redis.Host,
		"TOKEN_SECRET": cfg.Token.Secret,
	}
	var missing []string
	for key, val := range required {
		if val == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("缺少必填配置项: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getInt(v *viper.Viper, key string, def int) int {
	if !v.IsSet(key) {
		return def
	}
	return v.GetInt(key)
}

func getString(v *viper.Viper, key, def string) string {
	if !v.IsSet(key) {
		return def
	}
	return v.GetString(key)
}
