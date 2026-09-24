// 认证中间件（对位 JwtAuthenticationTokenFilter）。
//
// 行为对齐：解析失败/会话不存在时不拦截放行链（继续 c.Next），
// 受保护端点由 RequirePerm/RequireRole 发现无会话后返回 401 信封；
// 免认证端点走白名单直接放行（对位 SecurityConfig permitAll）。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/security"
)

// CtxKeyLoginUser gin context 中登录会话的存取键
const CtxKeyLoginUser = "login_user"

// AuthWhitelist 免认证端点集中配置（对位 SecurityConfig permitAll）。
// swagger 路径对位 springdoc（Java 侧 SecurityConfig 同样放行 swagger 资源）。
// 注意：此处仅免「令牌解析」，端点自身若无权限注解语义即公开（与 Java permitAll 一致）。
var AuthWhitelist = map[string]struct{}{
	"/":             {},
	"/login":        {},
	"/register":     {},
	"/captchaImage": {},

	"/swagger-ui.html":       {},
	"/v3/api-docs":           {},
	"/swagger-ui/index.html": {},
	// swagger-ui 静态资源（/swagger-ui/*）前缀放行，见 Auth 内前缀判断
}

// authWhitelistPrefix 免认证路径前缀（swagger 静态资源无法穷举，用前缀匹配）。
var authWhitelistPrefix = []string{"/swagger-ui/"}

// Auth 认证中间件：解析令牌 → 会话写入 gin context（含 <20 分钟自动续期）。
func Auth(ts *security.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := AuthWhitelist[path]; ok {
			c.Next()
			return
		}
		for _, prefix := range authWhitelistPrefix {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}
		if lu := ts.GetLoginUser(c.Request); lu != nil {
			ts.VerifyToken(c.Request.Context(), lu)
			c.Set(CtxKeyLoginUser, lu)
		}
		c.Next()
	}
}

// GetLoginUser 从 gin context 取登录会话；未认证返回 nil。
func GetLoginUser(c *gin.Context) *model.LoginUser {
	if v, ok := c.Get(CtxKeyLoginUser); ok {
		if lu, ok := v.(*model.LoginUser); ok {
			return lu
		}
	}
	return nil
}
