// 通知公告 + 监控（在线用户/缓存）业务与 HTTP 端点（轻 CRUD 直连 dao；对位对应 Controller）。
package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/security"
	"ruoyi-vue-go/pkg/response"
)

// NoticeHandler 通知公告。
type NoticeHandler struct{ dao *dao.NoticeDAO }

func NewNoticeHandler(d *dao.NoticeDAO) *NoticeHandler { return &NoticeHandler{dao: d} }

// List GET /system/notice/list。
func (h *NoticeHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:notice:list") == nil {
		return
	}
	list, err := h.dao.SelectNoticeList(c.Request.Context(), c.Query("noticeTitle"), c.Query("noticeType"), c.Query("createBy"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, noticePtr(list))
}

// ListTop GET /system/notice/listTop（登录即可；含 isRead + unreadCount）。
func (h *NoticeHandler) ListTop(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	list, err := h.dao.ListTopWithReadStatus(c.Request.Context(), lu.UserId, 5)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	unread := int64(0)
	for i := range list {
		if !list[i].IsRead {
			unread++
		}
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "data": noticePtr(list), "unreadCount": unread})
}

// MarkRead POST /system/notice/markRead?noticeId=。
func (h *NoticeHandler) MarkRead(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	noticeID, _ := strconv.ParseInt(c.Query("noticeId"), 10, 64)
	if err := h.dao.MarkRead(c.Request.Context(), lu.UserId, noticeID); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// MarkReadAll POST /system/notice/markReadAll?ids=。
func (h *NoticeHandler) MarkReadAll(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	for _, s := range strings.Split(c.Query("ids"), ",") {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		_ = h.dao.MarkRead(c.Request.Context(), lu.UserId, id)
	}
	response.Ok(c).JSON()
}

// ReadUsers GET /system/notice/readUsers/list。
func (h *NoticeHandler) ReadUsers(c *gin.Context) {
	if requirePermAndLogin(c, "system:notice:list") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Query("noticeId"), 10, 64)
	list, err := h.dao.ReadUsersByNotice(c.Request.Context(), id, c.Query("searchValue"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, list)
}

// GetInfo GET /system/notice/{noticeId}。
func (h *NoticeHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:notice:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("noticeId"), 10, 64)
	m, err := h.dao.SelectNoticeById(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Add POST /system/notice。
func (h *NoticeHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:notice:add")
	if lu == nil {
		return
	}
	var body dao.NoticeDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	body.CreateBy = lu.Username()
	if err := h.dao.InsertNotice(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/notice。
func (h *NoticeHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:notice:edit")
	if lu == nil {
		return
	}
	var body dao.NoticeDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	body.UpdateBy = lu.Username()
	if err := h.dao.UpdateNotice(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/notice/{noticeIds}。
func (h *NoticeHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:notice:remove") == nil {
		return
	}
	if err := h.dao.DeleteNoticeByIds(c.Request.Context(), parseInt64List(c.Param("noticeIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

func noticePtr(list []dao.NoticeDO) []*dao.NoticeDO {
	ret := make([]*dao.NoticeDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// ---------- 监控：操作日志 / 登录日志 / 在线用户 / 缓存 ----------

// OperLogHandler 操作日志。
type OperLogHandler struct{ dao *dao.OperLogDAO }

func NewOperLogHandler(d *dao.OperLogDAO) *OperLogHandler { return &OperLogHandler{dao: d} }

// List GET /monitor/operlog/list。
func (h *OperLogHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:operlog:list") == nil {
		return
	}
	list, err := h.dao.SelectOperLogList(c.Request.Context(), c.Query("title"), c.Query("operName"), c.Query("businessType"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, list)
}

// Remove DELETE /monitor/operlog/{operIds}。
func (h *OperLogHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:operlog:remove") == nil {
		return
	}
	if err := h.dao.DeleteOperLogByIds(c.Request.Context(), parseInt64List(c.Param("operIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Clean DELETE /monitor/operlog/clean。
func (h *OperLogHandler) Clean(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:operlog:remove") == nil {
		return
	}
	if err := h.dao.CleanOperLog(c.Request.Context()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// LogininforHandler 登录日志。
type LogininforHandler struct {
	dao   *dao.LogininforDAO
	cache *cache.RedisCache
}

func NewLogininforHandler(d *dao.LogininforDAO, rc *cache.RedisCache) *LogininforHandler {
	return &LogininforHandler{dao: d, cache: rc}
}

// List GET /monitor/logininfor/list。
func (h *LogininforHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:logininfor:list") == nil {
		return
	}
	list, err := h.dao.SelectLogininforList(c.Request.Context(), c.Query("userName"), c.Query("ipaddr"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, list)
}

// Remove DELETE /monitor/logininfor/{infoIds}。
func (h *LogininforHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:logininfor:remove") == nil {
		return
	}
	if err := h.dao.DeleteLogininforByIds(c.Request.Context(), parseInt64List(c.Param("infoIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Clean DELETE /monitor/logininfor/clean。
func (h *LogininforHandler) Clean(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:logininfor:remove") == nil {
		return
	}
	if err := h.dao.CleanLogininfor(c.Request.Context()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Unlock GET /monitor/logininfor/unlock/{userName}（清锁定计数）。
func (h *LogininforHandler) Unlock(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:logininfor:unlock") == nil {
		return
	}
	_, _ = h.cache.Delete(c.Request.Context(), "pwd_err_cnt:"+c.Param("userName"))
	response.Ok(c).JSON()
}

// OnlineHandler 在线用户。
type OnlineHandler struct {
	token *security.TokenService
	cache *cache.RedisCache
}

func NewOnlineHandler(t *security.TokenService, rc *cache.RedisCache) *OnlineHandler {
	return &OnlineHandler{token: t, cache: rc}
}

// onlineRow 在线用户行（对位 SysUserOnline 前端字段）。
type onlineRow struct {
	TokenID   string `json:"tokenId"`
	UserName  string `json:"userName"`
	DeptName  string `json:"deptName"`
	LoginIP   string `json:"ipaddr"`
	LoginAddr string `json:"loginLocation"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	LoginTime int64  `json:"loginTime"`
}

// List GET /monitor/online/list（SCAN login_tokens，userName/ipaddr 过滤）。
func (h *OnlineHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:online:list") == nil {
		return
	}
	keys, err := h.cache.KeysByPrefix(c.Request.Context(), "login_tokens:")
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	nameFilter, ipFilter := c.Query("userName"), c.Query("ipaddr")
	var rows []onlineRow
	for _, key := range keys {
		var lu = h.token.GetLoginUserByUUID(c.Request.Context(), strings.TrimPrefix(key, "login_tokens:"))
		if lu == nil {
			continue
		}
		if nameFilter != "" && !strings.Contains(lu.Username(), nameFilter) {
			continue
		}
		if ipFilter != "" && !strings.Contains(lu.Ipaddr, ipFilter) {
			continue
		}
		rows = append(rows, onlineRow{
			TokenID: lu.Token, UserName: lu.Username(), DeptName: lu.DeptName(),
			LoginIP: lu.Ipaddr, LoginAddr: lu.LoginLocation, Browser: lu.Browser,
			OS: lu.Os, LoginTime: lu.LoginTime,
		})
	}
	pagedJSON(c, rows)
}

// ForceLogout DELETE /monitor/online/{tokenId}。
func (h *OnlineHandler) ForceLogout(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:online:forceLogout") == nil {
		return
	}
	h.token.DelLoginUser(c.Request.Context(), c.Param("tokenId"))
	response.Ok(c).JSON()
}

// CacheHandler 缓存监控。
type CacheHandler struct{ cache *cache.RedisCache }

func NewCacheHandler(rc *cache.RedisCache) *CacheHandler { return &CacheHandler{cache: rc} }

// cacheNames 缓存监控面板的键前缀分组（对位 CacheController.cache 的静态清单）。
var cacheNames = []gin.H{
	{"cacheName": "login_tokens:", "remark": "用户信息"},
	{"cacheName": "sys_config:", "remark": "配置信息"},
	{"cacheName": "sys_dict:", "remark": "数据字典"},
	{"cacheName": "captcha_codes:", "remark": "验证码"},
	{"cacheName": "repeat_submit:", "remark": "防重提交"},
	{"cacheName": "rate_limit:", "remark": "限流"},
	{"cacheName": "pwd_err_cnt:", "remark": "密码错误次数"},
}

// Info GET /monitor/cache（面板首屏：Redis 信息）。
func (h *CacheHandler) Info(c *gin.Context) {
	info := h.cache.Info(c.Request.Context())
	response.Ok(c).
		Put("info", info).
		Put("dbSize", h.cache.DBSize(c.Request.Context())).JSON()
}

// Names GET /monitor/cache/getNames。
func (h *CacheHandler) Names(c *gin.Context) {
	response.OkData(c, cacheNames).JSON()
}

// Keys GET /monitor/cache/getKeys/{cacheName}。
func (h *CacheHandler) Keys(c *gin.Context) {
	keys, err := h.cache.KeysByPrefix(c.Request.Context(), c.Param("cacheName"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, keys).JSON()
}

// Value GET /monitor/cache/getValue/{cacheName}/{cacheKey}。
func (h *CacheHandler) Value(c *gin.Context) {
	name, key := c.Param("cacheName"), c.Param("cacheKey")
	full := name + key
	var val any
	if err := h.cache.GetObject(c.Request.Context(), full, &val); err != nil {
		response.Error(c, "缓存值不存在或解析失败").JSON()
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "data": gin.H{"cacheKey": full, "cacheValue": val}})
}

// ClearCacheKey DELETE /monitor/cache/clearCacheKey/{cacheKey}。
func (h *CacheHandler) ClearCacheKey(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:cache:list") == nil {
		return
	}
	// Java 侧 cacheKey 含完整前缀（如 login_tokens:uuid）
	_, _ = h.cache.Delete(c.Request.Context(), c.Param("cacheKey"))
	response.Ok(c).JSON()
}

// ClearCacheAll DELETE /monitor/cache/clearCacheAll。
func (h *CacheHandler) ClearCacheAll(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:cache:list") == nil {
		return
	}
	for _, n := range cacheNames {
		keys, _ := h.cache.KeysByPrefix(c.Request.Context(), n["cacheName"].(string))
		if len(keys) > 0 {
			_, _ = h.cache.Delete(c.Request.Context(), keys...)
		}
	}
	response.Ok(c).JSON()
}

// guard。
var _ = context.Background
