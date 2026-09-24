// 权限校验中间件（对位 @PreAuthorize("@ss.hasPermi(...)") / @ss.hasRole(...)）。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/pkg/response"
)

// RequirePerm 权限校验：无会话 401；校验不过 403（信封 HTTP 恒 200，对齐 Java 行为）。
func RequirePerm(perm string) gin.HandlerFunc {
	return require(func(lu *model.LoginUser) bool { return HasPerm(lu, perm) })
}

// RequireRole 角色校验：对位 @ss.hasRole。
func RequireRole(role string) gin.HandlerFunc {
	return require(func(lu *model.LoginUser) bool { return HasRole(lu, role) })
}

func require(check func(*model.LoginUser) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		lu := GetLoginUser(c)
		if lu == nil {
			response.Unauthorized(c).JSON()
			c.Abort()
			return
		}
		if check(lu) {
			c.Next()
			return
		}
		response.Forbidden(c).JSON()
		c.Abort()
	}
}

// HasPerm 权限匹配（对位 PermissionService.hasPermi）：admin（userId=1）放行；
// 含 *:*:* 全权标识放行；否则精确匹配（trim 后）。
func HasPerm(lu *model.LoginUser, perm string) bool {
	if lu == nil {
		return false
	}
	if lu.UserId == 1 {
		return true
	}
	for _, p := range lu.Permissions {
		if p == constant.AllPermission || p == strings.TrimSpace(perm) {
			return true
		}
	}
	return false
}

// HasRole 角色匹配（对位 PermissionService.hasRole）：admin（userId=1）放行；
// 任一角色 roleKey 为超级管理员标识（admin）放行；否则 roleKey 精确匹配（trim 后）。
func HasRole(lu *model.LoginUser, role string) bool {
	if lu == nil {
		return false
	}
	if lu.UserId == 1 {
		return true
	}
	for _, rk := range lu.RoleKeys() {
		if rk == constant.SuperAdmin || rk == strings.TrimSpace(role) {
			return true
		}
	}
	return false
}
