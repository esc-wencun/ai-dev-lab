// 字典类型/字典数据/参数管理 HTTP 端点（对位 SysDictTypeController/SysDictDataController/SysConfigController）。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/pkg/response"
)

// DictTypeHandler 字典类型。
type DictTypeHandler struct{ svc *service.DictTypeService }

func NewDictTypeHandler(s *service.DictTypeService) *DictTypeHandler { return &DictTypeHandler{svc: s} }

// List GET /system/dict/type/list。
func (h *DictTypeHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:list") == nil {
		return
	}
	list, err := h.svc.List(c.Request.Context(), c.Query("dictName"), c.Query("dictType"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, dictTypePtr(list))
}

// GetInfo GET /system/dict/type/{dictId}。
func (h *DictTypeHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("dictId"), 10, 64)
	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Add POST /system/dict/type。
func (h *DictTypeHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dict:add")
	if lu == nil {
		return
	}
	var body service.DictTypeInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/dict/type。
func (h *DictTypeHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dict:edit")
	if lu == nil {
		return
	}
	var body service.DictTypeInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/dict/type/{dictIds}。
func (h *DictTypeHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:remove") == nil {
		return
	}
	if err := h.svc.Remove(c.Request.Context(), parseInt64List(c.Param("dictIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// RefreshCache DELETE /system/dict/type/refreshCache。
func (h *DictTypeHandler) RefreshCache(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:remove") == nil {
		return
	}
	if err := h.svc.RefreshCache(c.Request.Context()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Optionselect GET /system/dict/type/optionselect。
func (h *DictTypeHandler) Optionselect(c *gin.Context) {
	list, err := h.svc.Optionselect(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, dictTypePtr(list)).JSON()
}

// DictDataHandler 字典数据。
type DictDataHandler struct{ svc *service.DictDataService }

func NewDictDataHandler(s *service.DictDataService) *DictDataHandler { return &DictDataHandler{svc: s} }

// List GET /system/dict/data/list。
func (h *DictDataHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:list") == nil {
		return
	}
	list, err := h.svc.List(c.Request.Context(), c.Query("dictType"), c.Query("dictLabel"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, dictDataPtr(list))
}

// GetInfo GET /system/dict/data/{dictCode}。
func (h *DictDataHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("dictCode"), 10, 64)
	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// TypeByDictType GET /system/dict/data/type/{dictType}（前端字典回显；免权限串）。
func (h *DictDataHandler) TypeByDictType(c *gin.Context) {
	list, err := h.svc.OptionType(c.Request.Context(), c.Param("dictType"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, dictDataPtr(list)).JSON()
}

// Add POST /system/dict/data。
func (h *DictDataHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dict:add")
	if lu == nil {
		return
	}
	var body service.DictDataInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/dict/data。
func (h *DictDataHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dict:edit")
	if lu == nil {
		return
	}
	var body service.DictDataInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/dict/data/{dictCodes}。
func (h *DictDataHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:dict:remove") == nil {
		return
	}
	if err := h.svc.Remove(c.Request.Context(), parseInt64List(c.Param("dictCodes"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ConfigHandler 参数管理。
type ConfigHandler struct{ svc *service.ConfigService }

func NewConfigHandler(s *service.ConfigService) *ConfigHandler { return &ConfigHandler{svc: s} }

// List GET /system/config/list。
func (h *ConfigHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:config:list") == nil {
		return
	}
	list, err := h.svc.List(c.Request.Context(), c.Query("configName"), c.Query("configKey"), c.Query("configType"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, configPtr(list))
}

// GetInfo GET /system/config/{configId}。
func (h *ConfigHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:config:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("configId"), 10, 64)
	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// ConfigKey GET /system/config/configKey/{configKey}。
func (h *ConfigHandler) ConfigKey(c *gin.Context) {
	v, err := h.svc.ConfigKeyByKey(c.Request.Context(), c.Param("configKey"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, v).JSON()
}

// Add POST /system/config。
func (h *ConfigHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:config:add")
	if lu == nil {
		return
	}
	var body service.ConfigInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/config。
func (h *ConfigHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:config:edit")
	if lu == nil {
		return
	}
	var body service.ConfigInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.svc.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/config/{configIds}。
func (h *ConfigHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:config:remove") == nil {
		return
	}
	if err := h.svc.Remove(c.Request.Context(), parseInt64List(c.Param("configIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// RefreshCache DELETE /system/config/refreshCache。
func (h *ConfigHandler) RefreshCache(c *gin.Context) {
	if requirePermAndLogin(c, "system:config:remove") == nil {
		return
	}
	if err := h.svc.RefreshCache(c.Request.Context()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

func dictTypePtr(list []dao.DictTypeDO) []*dao.DictTypeDO {
	ret := make([]*dao.DictTypeDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

func dictDataPtr(list []dao.DictDataDO) []*dao.DictDataDO {
	ret := make([]*dao.DictDataDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

func configPtr(list []dao.ConfigDO) []*dao.ConfigDO {
	ret := make([]*dao.ConfigDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// guard：登录上下文引用（免权限串端点用 Auth 中间件即可）。
var _ = middleware.GetLoginUser
