// RuoYi-Vue-GO 服务端入口：装配 config → mysql(GORM) → redis → router → gin engine。
// 启动：go run ./cmd/server --env=dev（监听 8080，与 Java / Python 版互斥，见 AGENTS.md）。
//
// API 文档（对位 Java springdoc）：
//
//	@title RuoYi-Vue-GO API
//	@version 1.0
//	@description RuoYi 管理系统 Go 版服务端接口（复刻 Java 版契约）
//	@BasePath /
//	@securityDefinitions.apikey Bearer
//	@in header
//	@name Authorization
//	@description 类型 "Bearer + 空格 + token"
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/config"
	"ruoyi-vue-go/internal/module/admin/scheduler"
	"ruoyi-vue-go/internal/router"
	"ruoyi-vue-go/internal/security"
	"ruoyi-vue-go/pkg/logger"
)

func main() {
	env := flag.String("env", "dev", "运行环境，对应 configs/.env.<env>")
	flag.Parse()

	cfg, err := config.Load(fmt.Sprintf("configs/.env.%s", *env))
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := newMySQL(cfg)
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}
	log.Printf("MySQL 已连接: %s:%d/%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)

	rdb := newRedis(cfg)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	log.Printf("Redis 已连接: %s:%d db%d", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Database)

	rc := cache.NewRedisCache(rdb)
	ts := security.NewTokenService(rc, cfg.Token.Header, cfg.Token.Secret, cfg.Token.ExpireMinutes)

	// 定时任务调度器：启动后加载全部启用态任务（暂停态不加载，恢复时从库补载）
	if _, err := logger.Init(logger.Options{Dir: "logs"}); err != nil {
		log.Printf("日志初始化降级为标准输出: %v", err)
	}
	scheduler.Start()
	defer scheduler.Stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router.New(router.Deps{DB: db, Cache: rc, Token: ts}),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("服务启动: http://localhost:%d (env=%s, mode=%s)", cfg.App.Port, *env, gin.Mode())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("收到退出信号，开始优雅关闭...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 关闭异常: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	if err := rdb.Close(); err != nil {
		log.Printf("Redis 关闭异常: %v", err)
	}
	log.Println("服务已退出")
}

// newMySQL 建立 GORM 连接（连接池参数对位 Java 版 druid：maxActive 20 / initialSize 5）。
func newMySQL(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)

	level := gormlogger.Warn
	if cfg.DB.Echo {
		level = gormlogger.Info
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func newRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})
}
