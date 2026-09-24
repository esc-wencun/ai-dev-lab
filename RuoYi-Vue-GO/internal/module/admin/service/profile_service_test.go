// 注册校验链单元测试（miniredis + sqlite 内存库）。
package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/message"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/pkg/types"
)

// newProfileEnv 测试环境。
func newProfileEnv(t *testing.T) (*ProfileService, *cache.RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rc := cache.NewRedisCache(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	db, err := gorm.Open(sqlite.Open("file:profile_"+strings.NewReplacer("-", "", ":", "").Replace(t.Name())+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite 打开失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	db.Exec(`create table sys_user (user_id integer primary key autoincrement, dept_id int, user_name text unique,
		nick_name text, email text, phonenumber text, password text, sex text, status text, del_flag text,
		pwd_update_date text, create_time text)`)
	pd := dao.NewProfileDao(db)
	ld := dao.NewLoginDao(db)
	ls := NewLoginService(ld, rc)
	return NewProfileService(pd, ld, rc, ls), rc, mr
}

// TestRegisterClosedSwitch 注册开关关闭（默认无配置 → 非 true）。
func TestRegisterClosedSwitch(t *testing.T) {
	svc, _, _ := newProfileEnv(t)
	msg := svc.Register(context.Background(), &RegisterInput{Username: "newuser", Password: "pass1234"})
	if msg != "当前系统没有开启注册功能！" {
		t.Errorf("msg = %q", msg)
	}
}

// TestRegisterValidationChain 开关开启后的校验链顺序与文案（对位 SysRegisterService.register）。
func TestRegisterValidationChain(t *testing.T) {
	svc, rc, mr := newProfileEnv(t)
	ctx := context.Background()
	// 开注册开关
	_ = rc.SetObject(ctx, cache.BuildKey(constant.SysConfigKey, "sys.account.registerUser"), "true", 0)

	// 校验链顺序：开关 → 验证码 → 用户名空 → 密码空 → 用户名长度 → 密码长度 → 唯一性（对位 Java）
	// 验证码开关默认开 → 缺验证码报失效（先于用户名/密码校验）
	if msg := svc.Register(ctx, &RegisterInput{Username: "abc", Password: "pass1234"}); msg != message.UserJcaptchaExpire {
		t.Errorf("msg = %q", msg)
	}
	// 预置验证码后：用户名空
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "rv0"), "0", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Password: "pass1234", Code: "0", UUID: "rv0"}); msg != "用户名不能为空" {
		t.Errorf("msg = %q", msg)
	}
	// 密码空
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "rv1"), "1", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Username: "abc", Code: "1", UUID: "rv1"}); msg != "用户密码不能为空" {
		t.Errorf("msg = %q", msg)
	}
	// 用户名过短
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "rv2"), "2", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Username: "a", Password: "pass1234", Code: "2", UUID: "rv2"}); msg != "账户长度必须在2到20个字符之间" {
		t.Errorf("msg = %q", msg)
	}
	// 密码过短
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "rv3"), "3", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Username: "abc", Password: "p", Code: "3", UUID: "rv3"}); msg != "密码长度必须在5到20个字符之间" {
		t.Errorf("msg = %q", msg)
	}
	// 预置验证码后注册成功
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "ru1"), "7", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Username: "newuser", Password: "pass1234", Code: "7", UUID: "ru1"}); msg != "" {
		t.Errorf("注册应成功, msg = %q", msg)
	}
	// 重复注册 → 账号已存在
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "ru2"), "8", 2*time.Minute)
	if msg := svc.Register(ctx, &RegisterInput{Username: "newuser", Password: "pass1234", Code: "8", UUID: "ru2"}); msg != "保存用户'newuser'失败，注册账号已存在" {
		t.Errorf("msg = %q", msg)
	}
	_ = mr
}

// TestCheckUnique 唯一性校验（本人排除语义）。
func TestCheckUnique(t *testing.T) {
	svc, _, _ := newProfileEnv(t)
	ctx := context.Background()
	// 空库里 phone/email 唯一
	if ok, _ := svc.profileDao.CheckPhoneUnique(ctx, "13800000001", 1); !ok {
		t.Error("无冲突应唯一")
	}
	svc.profileDao.InsertUser(ctx, "u1", "u1", "hash", types.DateTime{})
	// 用户名查重
	if ok, _ := svc.profileDao.CheckUserNameUnique(ctx, "u1"); ok {
		t.Error("同名应不唯一")
	}
	if ok, _ := svc.profileDao.CheckUserNameUnique(ctx, "u2"); !ok {
		t.Error("不同名应唯一")
	}
}
