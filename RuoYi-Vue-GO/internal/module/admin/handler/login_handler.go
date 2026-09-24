// Package handler HTTP 层：参数接收与校验、调 service、组响应；禁止直接操作 db/redis。
// 本文件对位 CaptchaController + SysLoginController + SysIndexController（登录闭环 7 端点）。
package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/message"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/internal/security"
	"ruoyi-vue-go/pkg/response"
)

// LoginHandler 登录闭环端点集合。
type LoginHandler struct {
	login *service.LoginService
	menu  *service.MenuService
	token *security.TokenService
	cache *cache.RedisCache
	dao   *dao.LoginDao
}

func NewLoginHandler(l *service.LoginService, m *service.MenuService, t *security.TokenService, rc *cache.RedisCache, d *dao.LoginDao) *LoginHandler {
	return &LoginHandler{login: l, menu: m, token: t, cache: rc, dao: d}
}

// CaptchaImage GET /captchaImage（对位 CaptchaController.getCode）。
// 返回 {code,msg,captchaEnabled,uuid,img}；img 为 jpg base64（无 data: 前缀）。
func (h *LoginHandler) CaptchaImage(c *gin.Context) {
	ajax := response.Ok(c)
	enabled := h.login.CaptchaEnabled(c.Request.Context())
	ajax.Put("captchaEnabled", enabled)
	if !enabled {
		ajax.JSON()
		return
	}
	cap := service.GenMathCaptcha()
	img, err := service.RenderTextJPEG(cap.Text)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	uuid := newUUID()
	_ = h.cache.SetObject(c.Request.Context(), cache.BuildKey(constant.CaptchaCodeKey, uuid), cap.Answer, constant.CaptchaExpiration*time.Minute)
	ajax.Put("uuid", uuid).Put("img", service.EncodeBase64(img)).JSON()
}

// Login POST /login（对位 SysLoginController.login）。
// 全链失败均 HTTP 200 + body code 500（service 已按 Java 顺序校验并记日志）。
func (h *LoginHandler) Login(c *gin.Context) {
	var body service.LoginInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, message.NotNull).JSON()
		return
	}
	body.IP = c.ClientIP()

	ud, err := h.login.Login(c.Request.Context(), &body)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}

	// 组装会话（对位 UserDetailsServiceImpl.createLoginUser + getMenuPermission）：
	// user 转 JSON（不含 password，dao 已不选出——此处直接序列化聚合）
	roles, permissions := h.userPermissions(c.Request.Context(), ud)
	lu, err := buildLoginUser(ud, roles, permissions)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	token, err := h.token.CreateToken(c.Request.Context(), lu, c.ClientIP(), uaBrowser(c), uaOS(c))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).Put(constant.Token, token).JSON()
}

// GetInfo GET /getInfo（对位 SysLoginController.getInfo + 定制三字段）。
func (h *LoginHandler) GetInfo(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	ctx := c.Request.Context()
	// 会话内嵌用户聚合（从 Redis 会话 JSON 还原；user JSON 原样回传）
	roles, permissions := h.sessionPermissions(ctx, lu)

	ajax := response.Ok(c)
	ajax.Put("user", jsonRaw(lu.User)).
		Put("roles", roles).
		Put("permissions", permissions)
	// 定制三字段（本 Java 版特有）
	pwdChrtype, _ := h.login.ConfigValue(ctx, "sys.account.chrtype")
	if pwdChrtype == "" {
		pwdChrtype = "0"
	}
	ajax.Put("pwdChrtype", pwdChrtype)
	ajax.Put("isDefaultModifyPwd", h.isDefaultModifyPwd(ctx, lu))
	ajax.Put("isPasswordExpired", h.isPasswordExpired(ctx, lu))
	ajax.JSON()
}

// GetRouters GET /getRouters（对位 SysLoginController.getRouters）。
func (h *LoginHandler) GetRouters(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	menus, err := h.menu.GetMenuTreeByUser(c.Request.Context(), lu.UserId)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, h.menu.BuildRouters(menus)).JSON()
}

// Logout POST /logout（对位 LogoutSuccessHandlerImpl：删会话 + Logout 日志 + 退出成功）。
func (h *LoginHandler) Logout(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu != nil {
		h.token.DelLoginUser(c.Request.Context(), lu.Token)
		h.login.RecordLogininforUA(c.Request.Context(), lu.Username(), constant.Logout, message.UserLogoutSuccess, c.ClientIP(), uaBrowser(c), uaOS(c))
	}
	response.Ok(c).Msg(message.UserLogoutSuccess).JSON()
}

// Unlockscreen POST /unlockscreen（对位 SysIndexController.unlockScreen 三分支）。
func (h *LoginHandler) Unlockscreen(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Password) == "" {
		response.Error(c, "密码不能为空").JSON()
		return
	}
	ud, err := h.dao.LoadUserByName(c.Request.Context(), lu.Username())
	if err != nil {
		response.Error(c, "服务器超时，请重新登录").JSON()
		return
	}
	if err := h.login.ValidateUnlockPassword(c.Request.Context(), lu.Username(), body.Password, ud.User.Password); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).Msg("解锁成功").JSON()
}

// Index GET /（对位 SysIndexController.index 纯文本欢迎语；具体文案与 Java 差异已登记 deviations）。
func (h *LoginHandler) Index(c *gin.Context) {
	c.String(http.StatusOK, "欢迎使用RuoYi后台管理框架，当前版本：v1.0，请通过前端地址访问。")
}

// GetPlatformInfo GET /getPlatformInfo（对位 SysPlatformController，登录即可）。
// 返回语言/框架版本与功能开关，前端据此对 Druid 数据监控等 Java 特有功能做降级提示。
func (h *LoginHandler) GetPlatformInfo(c *gin.Context) {
	if middleware.GetLoginUser(c) == nil {
		response.Unauthorized(c).JSON()
		return
	}
	goVer := strings.TrimPrefix(runtime.Version(), "go")
	response.Ok(c).
		Put("framework", "RuoYi-Vue-GO").
		Put("version", "1.0.0").
		Put("language", "go").
		Put("languageVersion", goVer).
		Put("features", gin.H{"druidMonitor": false, "serverMonitor": false, "swaggerDocs": false}).
		JSON()
}

// userPermissions 登录时计算角色集合与菜单权限集合（对位 getRolePermission + getMenuPermission）。
func (h *LoginHandler) userPermissions(ctx context.Context, ud *dao.UserDetail) ([]string, []string) {
	ctx = context0(ctx)
	var roles, permissions []string
	if ud.User.UserID == service.AdminUserID {
		roles = []string{constant.SuperAdmin}
		permissions = []string{constant.AllPermission}
		return roles, permissions
	}
	keys, _ := h.dao.RoleKeysByUser(ctx, ud.User.UserID)
	roles = keys
	// 对位 getMenuPermission：正常状态非 admin 角色逐个取菜单权限并合并（同时存入角色对象语义由会话 JSON 承载）
	seen := map[string]bool{}
	for _, r := range ud.Roles {
		if r.Status != constant.RoleNormal {
			continue
		}
		perms, _ := h.dao.MenuPermsByRole(ctx, r.RoleID)
		for _, p := range perms {
			if p == "" {
				continue
			}
			if !seen[p] {
				seen[p] = true
				permissions = append(permissions, p)
			}
		}
	}
	if len(permissions) == 0 {
		perms, _ := h.dao.MenuPermsByUser(ctx, ud.User.UserID)
		permissions = perms
	}
	return roles, permissions
}

// sessionPermissions 已登录会话的权限刷新（对位 getInfo 中 getMenuPermission 比对 + refreshToken；
// Go 版简化为重新计算并回写会话，权限变更后下次 getInfo 生效）。
func (h *LoginHandler) sessionPermissions(ctx context.Context, lu *model.LoginUser) ([]string, []string) {
	if lu.UserId == service.AdminUserID {
		return []string{constant.SuperAdmin}, []string{constant.AllPermission}
	}
	// 会话的 permissions 即登录时存的集合；刷新走 refreshPermissionByRoleId（6.0.0 接入）
	return nil, lu.Permissions
}

// isDefaultModifyPwd 对位 initPasswordIsModify：initPasswordModify==1 且从未改过密码。
func (h *LoginHandler) isDefaultModifyPwd(ctx context.Context, lu *model.LoginUser) bool {
	v, err := h.login.ConfigValue(ctx, "sys.account.initPasswordModify")
	if err != nil || v != "1" {
		return false
	}
	return userPwdUpdateDate(lu) == nil
}

// isPasswordExpired 对位 passwordIsExpiration：passwordValidateDays>0 且（未改密 或 超 N 天）。
func (h *LoginHandler) isPasswordExpired(ctx context.Context, lu *model.LoginUser) bool {
	v, err := h.login.ConfigValue(ctx, "sys.account.passwordValidateDays")
	if err != nil || v == "" {
		return false
	}
	days := parsePositiveInt(v)
	if days <= 0 {
		return false
	}
	t := userPwdUpdateDate(lu)
	if t == nil {
		return true
	}
	return elapsedDays(*t) > days
}

// newUUID 32 位 hex 随机串（对位 IdUtils.simpleUUID）。
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// uaBrowser / uaOS User-Agent 解析（轻量规则，对位 UserAgentUtils 的常见形态）。
func uaBrowser(c *gin.Context) string {
	ua := c.GetHeader("User-Agent")
	switch {
	case strings.Contains(ua, "Edg"):
		return "Edge"
	case strings.Contains(ua, "Chrome"):
		return "Chrome"
	case strings.Contains(ua, "Firefox"):
		return "Firefox"
	case strings.Contains(ua, "Safari"):
		return "Safari"
	default:
		return "Unknown"
	}
}

func uaOS(c *gin.Context) string {
	ua := c.GetHeader("User-Agent")
	switch {
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "Mac OS"):
		return "Mac OS X"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		return "iOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}
