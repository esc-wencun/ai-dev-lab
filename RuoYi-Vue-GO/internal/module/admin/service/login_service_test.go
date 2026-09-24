// 登录闭环单元测试：验证码、密码互验、重试计数、黑名单（miniredis + sqlite 内存库）。
package service

import (
	"context"
	"encoding/base64"
	"image/jpeg"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/message"
	"ruoyi-vue-go/internal/module/admin/dao"
)

// newLoginEnv 测试环境：miniredis + sqlite 内存库 + 种子用户（bcrypt 哈希 $2a$ 形态）。
func newLoginEnv(t *testing.T) (*LoginService, *cache.RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rc := cache.NewRedisCache(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	db, err := gorm.Open(sqlite.Open("file:login_svc_"+strings.NewReplacer("-", "", ":", "").Replace(t.Name())+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite 打开失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	db.Exec("create table sys_user (user_id int primary key, dept_id int, user_name text, password text, status text, del_flag text)")
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	db.Exec("insert into sys_user values (1, 100, 'admin', ?, '0', '0')", string(hash))
	return NewLoginService(dao.NewLoginDao(db), rc), rc, mr
}

// TestGenMathCaptcha 题目形态与答案正确性（对位 KaptchaTextCreator 语义）。
func TestGenMathCaptcha(t *testing.T) {
	for i := 0; i < 200; i++ {
		c := GenMathCaptcha()
		// 形态：数字op数字=?@答案（LastIndex("@") 拆分）
		idx := strings.LastIndex(c.Text, "@")
		if idx >= 0 {
			t.Fatalf("Text 不应含 @: %q", c.Text)
		}
		if !strings.HasSuffix(c.Text, "=?") {
			t.Errorf("题目应以 =? 结尾: %q", c.Text)
		}
		expr := strings.TrimSuffix(c.Text, "=?")
		ans := calcExpr(expr)
		if ans != c.Answer {
			t.Errorf("答案错误: %s = %s, got %s", expr, c.Answer, ans)
		}
	}
}

// calcExpr 解析 "xopy" 形态算式（测试辅助；除法语义=y/x 对位 KaptchaTextCreator）。
func calcExpr(expr string) string {
	ops := []string{"*", "/", "+", "-"}
	for _, op := range ops {
		if idx := strings.Index(expr, op); idx > 0 {
			x := atoi(expr[:idx])
			y := atoi(expr[idx+1:])
			switch op {
			case "*":
				return itoa(x * y)
			case "/":
				if y != 0 {
					return itoa(x / y) // 表达式 "y/x" 按书写顺序求值（KaptchaTextCreator 已保证整除）
				}
			case "+":
				return itoa(x + y)
			case "-":
				return itoa(x - y)
			}
		}
	}
	return "?"
}

func atoi(s string) int {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return -1
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

// TestCaptchaImageJPEG 验证码图可编码为合法 jpg 并 base64。
func TestCaptchaImageJPEG(t *testing.T) {
	c := GenMathCaptcha()
	data, err := RenderTextJPEG(c.Text)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	img, err := jpeg.Decode(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("非合法 jpg: %v", err)
	}
	if img.Bounds().Dx() == 0 {
		t.Error("图片尺寸异常")
	}
	b64 := EncodeBase64(data)
	if strings.Contains(b64, "data:") {
		t.Error("base64 不应带 data: 前缀（对齐 Java Base64.encode）")
	}
	// 可回解码
	if _, err := base64.StdEncoding.DecodeString(b64); err != nil {
		t.Errorf("base64 回解码失败: %v", err)
	}
}

// TestCaptchaValidateFlow 验证码校验链：正确通过、错误拒绝、过期拒绝、取到即删。
func TestCaptchaValidateFlow(t *testing.T) {
	svc, rc, mr := newLoginEnv(t)
	ctx := context.Background()

	// 开关默认开（无配置）
	if !svc.CaptchaEnabled(ctx) {
		t.Fatal("无配置应默认开启")
	}
	// 预置验证码
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "u1"), "5", 2*time.Minute)

	// 错误 → 验证码错误
	if err := svc.ValidateCaptcha(ctx, "admin", "6", "u1"); err == nil || err.Error() != message.UserJcaptchaError {
		t.Errorf("错误码应拒绝: %v", err)
	}
	// 取到即删：再次校验 → 已失效
	if err := svc.ValidateCaptcha(ctx, "admin", "5", "u1"); err == nil || err.Error() != message.UserJcaptchaExpire {
		t.Errorf("一次性应失效: %v", err)
	}
	// 重新预置 → 忽略大小写通过
	_ = rc.SetObject(ctx, cache.BuildKey(constant.CaptchaCodeKey, "u2"), "Ab", 2*time.Minute)
	if err := svc.ValidateCaptcha(ctx, "admin", "aB", "u2"); err != nil {
		t.Errorf("忽略大小写应通过: %v", err)
	}
	// 过期
	mr.FastForward(3 * time.Minute)
	if err := svc.ValidateCaptcha(ctx, "admin", "Ab", "u2"); err == nil {
		t.Error("过期后应失效")
	}
}

// TestValidatePasswordRetry 重试计数：错 4 次后第 5 次报错文案，第 6 次锁定；正确密码清零。
func TestValidatePasswordRetry(t *testing.T) {
	svc, rc, _ := newLoginEnv(t)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 4) // 低 cost 提速
	// 错 5 次（maxRetryCount=5）
	for i := 1; i <= 5; i++ {
		err := svc.ValidatePassword(ctx, "admin", "wrong-pw", string(hash))
		if i < 5 && (err == nil || err.Error() != message.UserPasswordNotMatch) {
			t.Fatalf("第 %d 次应报不匹配: %v", i, err)
		}
	}
	// 第 6 次：达到阈值 → 锁定文案
	err := svc.ValidatePassword(ctx, "admin", "wrong-pw", string(hash))
	want := "密码输入错误5次，帐户锁定10分钟"
	if err == nil || err.Error() != want {
		t.Errorf("锁定文案 = %v, want %q", err, want)
	}
	// 正确密码在锁定期间同样拒登（validate 先查计数）
	err = svc.ValidatePassword(ctx, "admin", "admin123", string(hash))
	if err == nil || err.Error() != want {
		t.Errorf("锁定期间正确密码也应拒登: %v", err)
	}
	// 清空计数后正确密码通过
	_, _ = rc.Delete(ctx, cache.BuildKey(constant.PwdErrCntKey, "admin"))
	if err := svc.ValidatePassword(ctx, "admin", "admin123", string(hash)); err != nil {
		t.Errorf("正确密码应通过: %v", err)
	}
}

// TestBcryptJavaCompat Java 存量哈希互验冒烟：$2a$ 前缀哈希可被 Go bcrypt 校验。
// （Java BCryptPasswordEncoder 生成 $2a$10$...；真实库数据互验在端到端阶段执行）
func TestBcryptJavaCompat(t *testing.T) {
	// Python bcrypt 生成的 $2b$ 同样兼容（x/crypto/bcrypt 支持 $2a/$2b/$2y）
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	if !strings.HasPrefix(string(hash), "$2") {
		t.Fatalf("哈希前缀异常: %s", hash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte("admin123")) != nil {
		t.Error("哈希应可校验")
	}
	if bcrypt.CompareHashAndPassword(hash, []byte("wrong")) == nil {
		t.Error("错误密码不应通过")
	}
}

// TestIPOutlineMatched 黑名单匹配：通配 * 与多段前缀（对位 IpUtils.isMatchedIp）。
func TestIPOutlineMatched(t *testing.T) {
	cases := []struct {
		black, ip string
		want      bool
	}{
		{"", "1.2.3.4", false},
		{"192.168.1.100", "192.168.1.100", true},
		{"192.168.1.100", "192.168.1.101", false},
		{"10.20.*", "10.20.3.4", true},
		{"10.20.*", "10.21.3.4", false},
		{"192.168.1.100,10.20.*", "10.20.5.6", true},
		{"*", "8.8.8.8", true},
	}
	for _, tc := range cases {
		if got := IPOutlineMatched(tc.black, tc.ip); got != tc.want {
			t.Errorf("IPOutlineMatched(%q, %q) = %v, want %v", tc.black, tc.ip, got, tc.want)
		}
	}
}

// TestLogininforStatus 退出/注册等文案状态归一化为 char(1)（对位 AsyncFactory.recordLogininfor 映射）。
func TestLogininforStatus(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{constant.LoginSuccess, "0"},
		{constant.Logout, "0"},
		{constant.Register, "0"},
		{constant.LoginFail, "1"},
		{"0", "0"},
		{"1", "1"},
	}
	for _, tc := range cases {
		if got := logininforStatus(tc.in); got != tc.want {
			t.Errorf("logininforStatus(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
