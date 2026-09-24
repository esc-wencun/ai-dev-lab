// Package security 认证安全（对位 Java ruoyi-framework 的 TokenService）。
//
// 链路：Authorization 头剥 "Bearer " → JWT HS512 解析 → claim login_user_key（会话 uuid）
// → Redis login_tokens:{uuid} → LoginUser。
//
// 密钥语义：与 Python 版（已验收基准）一致，secret 原文 UTF-8 字节直接作 HMAC key，
// 不做 base64 解码（JWT 跨版互通无意义——AGENTS.md 明确会话不互通，切换后端需重新登录）。
package security

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/model"
)

// refreshThresholdMs 对位 TokenService.MILLIS_MINUTE_TWENTY：剩余有效期不足 20 分钟自动续期
const refreshThresholdMs = 20 * 60 * 1000

// TokenService 令牌验证处理
type TokenService struct {
	cache         *cache.RedisCache
	header        string // 请求头名称，Java 版为 Authorization
	secret        []byte
	expireMinutes int // 会话有效期（分钟），对位 token.expireTime
}

func NewTokenService(cache *cache.RedisCache, header, secret string, expireMinutes int) *TokenService {
	return &TokenService{
		cache:         cache,
		header:        header,
		secret:        []byte(secret),
		expireMinutes: expireMinutes,
	}
}

// GetLoginUser 从请求解析令牌并加载会话；任何环节失败返回 nil（调用方按未登录处理，
// 对位 Java getLoginUser 捕获异常后返回 null，不拦截放行链）。
func (s *TokenService) GetLoginUser(r *http.Request) *model.LoginUser {
	token := s.stripPrefix(r.Header.Get(s.header))
	if token == "" {
		return nil
	}
	uuid, err := s.parseUUID(token)
	if err != nil {
		return nil
	}
	var lu model.LoginUser
	key := cache.BuildKey(constant.LoginTokenKey, uuid)
	if err := s.cache.GetObject(r.Context(), key, &lu); err != nil {
		return nil
	}
	return &lu
}

// VerifyToken 剩余有效期不足 20 分钟自动续期（对位 verifyToken + JwtAuthenticationTokenFilter 调用点）。
func (s *TokenService) VerifyToken(ctx context.Context, lu *model.LoginUser) {
	if lu.ExpireTime-time.Now().UnixMilli() <= refreshThresholdMs {
		_ = s.RefreshToken(ctx, lu)
	}
}

// RefreshToken 刷新会话：重置 loginTime/expireTime 并按有效期回写 Redis（对位 refreshToken）。
func (s *TokenService) RefreshToken(ctx context.Context, lu *model.LoginUser) error {
	if lu.Token == "" {
		return fmt.Errorf("会话 uuid 为空，无法续期")
	}
	now := time.Now().UnixMilli()
	lu.LoginTime = now
	lu.ExpireTime = now + int64(s.expireMinutes)*60*1000
	key := cache.BuildKey(constant.LoginTokenKey, lu.Token)
	return s.cache.SetObject(ctx, key, lu, time.Duration(s.expireMinutes)*time.Minute)
}

// stripPrefix 剥离令牌前缀（对位 getToken：replace 首个 "Bearer "）。
func (s *TokenService) stripPrefix(headerVal string) string {
	if headerVal == "" {
		return ""
	}
	if strings.HasPrefix(headerVal, constant.TokenPrefix) {
		return strings.Replace(headerVal, constant.TokenPrefix, "", 1)
	}
	return headerVal
}

// parseUUID 解析 JWT 并取 claim login_user_key（对位 parseToken + Constants.LOGIN_USER_KEY）。
func (s *TokenService) parseUUID(token string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{"HS512"}))
	if err != nil {
		return "", err
	}
	uuid, ok := claims[constant.LoginUserKey].(string)
	if !ok || uuid == "" {
		return "", fmt.Errorf("claim %s 缺失或为空", constant.LoginUserKey)
	}
	return uuid, nil
}

// CreateToken 登录成功创建令牌（对位 createToken）：随机 UUID 会话键 → LoginUser 写 Redis
// （TTL=expireTime 分钟）→ JWT（HS512，claims 仅 login_user_key + subject=用户名）。
// ip/browser/os 由调用方从请求解析后传入（对位 setUserAgent）。
func (s *TokenService) CreateToken(ctx context.Context, lu *model.LoginUser, ip, browser, osName string) (string, error) {
	token := uuid()
	lu.Token = token
	lu.Ipaddr = ip
	lu.Browser = browser
	lu.Os = osName
	if err := s.RefreshToken(ctx, lu); err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		constant.LoginUserKey: token,
		constant.JwtUsername:  lu.Username(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(s.secret)
}

// DelLoginUser 删除会话（对位 delLoginUser：logout 用）。
func (s *TokenService) DelLoginUser(ctx context.Context, token string) {
	if token == "" {
		return
	}
	_, _ = s.cache.Delete(ctx, cache.BuildKey(constant.LoginTokenKey, token))
}

// uuid 32 位无连字符随机串（对位 IdUtils.fastUUID 的会话键用途；格式差异不影响契约——
// uuid 仅作 Redis 键后缀与 JWT claim，前端原样回传）。
func uuid() string {
	b := make([]byte, 16)
	_, _ = cryptorand.Read(b)
	return hex.EncodeToString(b)
}

// GetLoginUserByUUID 按会话 uuid 直取 LoginUser（在线用户监控用，跳过 JWT 解析）。
func (s *TokenService) GetLoginUserByUUID(ctx context.Context, uuid string) *model.LoginUser {
	var lu model.LoginUser
	if err := s.cache.GetObject(ctx, cache.BuildKey(constant.LoginTokenKey, uuid), &lu); err != nil {
		return nil
	}
	lu.Token = uuid
	return &lu
}
