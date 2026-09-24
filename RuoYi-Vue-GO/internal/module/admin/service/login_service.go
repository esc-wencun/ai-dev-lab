// Package service 业务逻辑层：事务边界与缓存维护，禁止拼 SQL。
// 本文件对位 SysLoginService + SysPasswordService + SysPermissionService + CaptchaController 的业务部分。
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/message"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/pkg/types"
)

// AdminUserID 超级管理员（对位 SysUser.isAdmin）。
const AdminUserID = 1

// LoginService 登录校验（对位 SysLoginService + SysPasswordService 合并——Go 无 Spring Security
// AuthenticationManager，authenticate 拆为直查用户 + bcrypt 校验 + 重试计数）。
type LoginService struct {
	dao           *dao.LoginDao
	cache         *cache.RedisCache
	maxRetryCount int // 对位 user.password.maxRetryCount，默认 5
	lockTimeMin   int // 对位 user.password.lockTime，默认 10 分钟
}

func NewLoginService(d *dao.LoginDao, rc *cache.RedisCache) *LoginService {
	return &LoginService{dao: d, cache: rc, maxRetryCount: 5, lockTimeMin: 10}
}

// LoginInput 对位 LoginBody。
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
	UUID     string `json:"uuid"`
	IP       string `json:"-"` // controller 层从请求提取后填入
}

// Login 登录校验（对位 SysLoginService.login 四步链，文案与顺序逐条对齐）。
// 返回 user 聚合（供上层组装会话）。
func (s *LoginService) Login(ctx context.Context, in *LoginInput) (*dao.UserDetail, error) {
	// 1. 验证码校验
	if err := s.ValidateCaptcha(ctx, in.Username, in.Code, in.UUID); err != nil {
		return nil, err
	}
	// 2. 前置校验
	if err := s.PreCheck(ctx, in.Username, in.Password, in.IP); err != nil {
		return nil, err
	}
	// 3. 用户加载 + 密码校验（对位 authenticationManager.authenticate → validate）
	ud, err := s.dao.LoadUserByName(ctx, in.Username)
	if err != nil {
		// 用户不存在与密码错误同文案（防用户名枚举，对位 UserNotExistsException 文案）
		s.recordLoginFail(ctx, in.Username, in.IP, message.UserNotExists)
		return nil, fmt.Errorf("%s", message.UserNotExists)
	}
	if err := s.ValidatePassword(ctx, in.Username, in.Password, ud.User.Password); err != nil {
		s.recordLoginFail(ctx, in.Username, in.IP, err.Error())
		return nil, err
	}
	// 4. 成功路径
	s.recordLoginSuccess(ctx, in.Username, in.IP)
	s.UpdateLoginInfo(ctx, ud.User.UserID, in.IP)
	return ud, nil
}

// ValidateCaptcha 验证码校验（对位 validateCaptcha：取到即删、忽略大小写）。
func (s *LoginService) ValidateCaptcha(ctx context.Context, username, code, uuid string) error {
	if !s.CaptchaEnabled(ctx) {
		return nil
	}
	key := cache.BuildKey(constant.CaptchaCodeKey, nvl(uuid))
	var stored string
	if err := s.cache.GetObject(ctx, key, &stored); err != nil {
		s.recordLoginFail(ctx, username, "", message.UserJcaptchaExpire)
		return fmt.Errorf("%s", message.UserJcaptchaExpire)
	}
	_, _ = s.cache.Delete(ctx, key)
	if !strings.EqualFold(code, stored) {
		s.recordLoginFail(ctx, username, "", message.UserJcaptchaError)
		return fmt.Errorf("%s", message.UserJcaptchaError)
	}
	return nil
}

// PreCheck 登录前置校验（对位 loginPreCheck：非空/长度/IP 黑名单）。
func (s *LoginService) PreCheck(ctx context.Context, username, password, ip string) error {
	if username == "" || password == "" {
		s.recordLoginFail(ctx, username, ip, message.NotNull)
		return fmt.Errorf("%s", message.UserNotExists)
	}
	if l := len(password); l < constant.PasswordMinLength || l > constant.PasswordMaxLength {
		s.recordLoginFail(ctx, username, ip, message.UserPasswordNotMatch)
		return fmt.Errorf("%s", message.UserPasswordNotMatch)
	}
	if l := len([]rune(username)); l < constant.UsernameMinLength || l > constant.UsernameMaxLength {
		s.recordLoginFail(ctx, username, ip, message.UserPasswordNotMatch)
		return fmt.Errorf("%s", message.UserPasswordNotMatch)
	}
	blackStr, _ := s.ConfigValue(ctx, "sys.login.blackIPList")
	if IPOutlineMatched(blackStr, ip) {
		s.recordLoginFail(ctx, username, ip, message.LoginBlocked)
		return fmt.Errorf("%s", message.LoginBlocked)
	}
	return nil
}

// ValidatePassword 密码校验（对位 SysPasswordService.validate：重试计数/锁定/bcrypt）。
// Java 存量哈希 $2a$ 由 golang.org/x/crypto/bcrypt 原生兼容。
func (s *LoginService) ValidatePassword(ctx context.Context, username, rawPassword, hash string) error {
	key := cache.BuildKey(constant.PwdErrCntKey, username)
	retryCount := 0
	if err := s.cache.GetObject(ctx, key, &retryCount); err == nil && retryCount >= s.maxRetryCount {
		return fmt.Errorf(message.UserPasswordRetryLimitExceed, s.maxRetryCount, s.lockTimeMin)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawPassword)) != nil {
		retryCount++
		_ = s.cache.SetObject(ctx, key, retryCount, time.Duration(s.lockTimeMin)*time.Minute)
		return fmt.Errorf("%s", message.UserPasswordNotMatch)
	}
	// 匹配：清计数
	_, _ = s.cache.Delete(ctx, key)
	return nil
}

// CaptchaEnabled 验证码开关（对位 selectCaptchaEnabled：配置空视为开启）。
func (s *LoginService) CaptchaEnabled(ctx context.Context) bool {
	v, err := s.ConfigValue(ctx, "sys.account.captchaEnabled")
	if err != nil || v == "" {
		return true
	}
	return v == "true" || v == "1"
}

// ConfigValue 参数查询（对位 selectConfigByKey：sys_config: 缓存未命中回源并回填）。
func (s *LoginService) ConfigValue(ctx context.Context, key string) (string, error) {
	ck := cache.BuildKey(constant.SysConfigKey, key)
	var val string
	if err := s.cache.GetObject(ctx, ck, &val); err == nil && val != "" {
		return val, nil
	}
	v, err := s.dao.GetConfigByKey(ctx, key)
	if err != nil {
		return "", err
	}
	_ = s.cache.SetObject(ctx, ck, v, 0)
	return v, nil
}

// UpdateLoginInfo 登录成功更新 login_ip/login_date（对位 recordLoginInfo）。
func (s *LoginService) UpdateLoginInfo(ctx context.Context, userID int64, ip string) {
	_ = s.dao.UpdateLoginInfo(ctx, userID, ip, types.Now())
}

// recordLoginSuccess / recordLoginFail 登录日志（对位 AsyncFactory.recordLogininfor 异步写）。
func (s *LoginService) recordLoginSuccess(ctx context.Context, username, ip string) {
	s.recordLogininfor(ctx, username, constant.Success, message.UserLoginSuccess, ip)
}

func (s *LoginService) recordLoginFail(ctx context.Context, username, ip, msg string) {
	s.recordLogininfor(ctx, username, constant.Fail, msg, ip)
}

// recordLogininfor 异步落库（goroutine + recover，对位 AsyncManager；请求返回后仍可写库，
// 用脱离请求的独立 ctx）。browser/os 由 User-Agent 解析，controller 层传入。
func (s *LoginService) recordLogininfor(ctx context.Context, username, status, msg, ip string) {
	s.recordLogininforUA(ctx, username, status, msg, ip, "", "")
}

// RecordLogininforUA 带浏览器/OS 的登录日志（controller 层解析 User-Agent 后调用）。
func (s *LoginService) RecordLogininforUA(ctx context.Context, username, status, msg, ip, browser, os string) {
	s.recordLogininforUA(ctx, username, status, msg, ip, browser, os)
}

// logininforStatus 对位 AsyncFactory.recordLogininfor 的状态归一化：status 列是 char(1)，
// LoginSuccess/Logout/Register 落 0（成功），LoginFail 落 1（失败），其余原样（预期已是 0/1）。
func logininforStatus(status string) string {
	switch status {
	case constant.LoginSuccess, constant.Logout, constant.Register:
		return constant.Success
	case constant.LoginFail:
		return constant.Fail
	default:
		return status
	}
}

func (s *LoginService) recordLogininforUA(ctx context.Context, username, status, msg, ip, browser, osName string) {
	m := &do.SysLogininfor{
		UserName: username,
		IPAddr:   ip,
		Status:   logininforStatus(status),
		Msg:      msg,
		Browser:  browser,
		OS:       osName,
	}
	go func() {
		defer func() { _ = recover() }()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		m.LoginTime = types.Now()
		_ = s.dao.InsertLogininfor(c, m)
	}()
}

// IPOutlineMatched IP 黑名单匹配（对位 IpUtils.isMatchedIp：filter 段含 * 时按前缀匹配，逗号分隔）。
func IPOutlineMatched(blackStr, ip string) bool {
	if blackStr == "" || ip == "" {
		return false
	}
	for _, seg := range strings.Split(blackStr, ",") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if seg == "*" || strings.HasPrefix(ip, strings.TrimSuffix(seg, "*")) {
			return true
		}
	}
	return false
}

// ValidateUnlockPassword 解锁屏密码校验（对位 unlockScreen：bcrypt 校验，无重试计数联动）。
func (s *LoginService) ValidateUnlockPassword(ctx context.Context, username, rawPassword, hash string) error {
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawPassword)) != nil {
		return fmt.Errorf("%s", "密码错误，请重新输入")
	}
	return nil
}

// nvl 空串转 ""（对位 StringUtils.nvl 的本场景语义）。
func nvl(s string) string {
	if s == "" {
		return ""
	}
	return s
}

// MathCaptcha math 型验证码题目（对位 KaptchaTextCreator.getText：0-9 取数、乘/除或加/减、
// 形如 "1+2=?@3"——? 前为题目、@ 后为答案；Java kaptcha 渲染形态 "1+2=@3"，? 由前端展示约束等价）。
type MathCaptcha struct {
	Text   string // 题目（如 "1+2=?"）
	Answer string // 答案
}

// GenMathCaptcha 生成算术验证码（分支与取值范围逐条对位 KaptchaTextCreator）。
func GenMathCaptcha() *MathCaptcha {
	cnum := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	x := randIntn(10)
	y := randIntn(10)
	var result int
	var sb strings.Builder
	switch randIntn(3) {
	case 0:
		result = x * y
		sb.WriteString(cnum[x])
		sb.WriteString("*")
		sb.WriteString(cnum[y])
	case 1:
		if x != 0 && y%x == 0 {
			result = y / x
			sb.WriteString(cnum[y])
			sb.WriteString("/")
			sb.WriteString(cnum[x])
		} else {
			result = x + y
			sb.WriteString(cnum[x])
			sb.WriteString("+")
			sb.WriteString(cnum[y])
		}
	default:
		if x >= y {
			result = x - y
			sb.WriteString(cnum[x])
			sb.WriteString("-")
			sb.WriteString(cnum[y])
		} else {
			result = y - x
			sb.WriteString(cnum[y])
			sb.WriteString("-")
			sb.WriteString(cnum[x])
		}
	}
	sb.WriteString("=?@")
	sb.WriteString(fmt.Sprintf("%d", result))
	full := sb.String()
	// 对位 CaptchaController 拆分：capText="8+3=?@11"，capStr（题目）= "8+3=?"，code（答案）= "11"
	idx := strings.LastIndex(full, "@")
	return &MathCaptcha{
		Text:   full[:idx],
		Answer: full[idx+1:],
	}
}

// randIntn 密码学安全随机 [0,n)。
func randIntn(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// EncodeBase64 标准 base64（对位 Java Base64.encode，无 data: 前缀）。
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
