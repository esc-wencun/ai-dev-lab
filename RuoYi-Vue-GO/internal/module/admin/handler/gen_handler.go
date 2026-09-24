// 代码生成器 HTTP 端点（降级范围：数据层端点；模板生成类端点有意排除，见 spec/deviations）。
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/pkg/response"
)

// GenHandler 代码生成。
type GenHandler struct{ dao *dao.GenTableDAO }

func NewGenHandler(d *dao.GenTableDAO) *GenHandler { return &GenHandler{dao: d} }

// List GET /tool/gen/list。
func (h *GenHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "tool:gen:list") == nil {
		return
	}
	list, err := h.dao.SelectGenList(c.Request.Context(), c.Query("tableName"), c.Query("tableComment"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, genPtr(list))
}

// DbList GET /tool/gen/db/list。
func (h *GenHandler) DbList(c *gin.Context) {
	if requirePermAndLogin(c, "tool:gen:list") == nil {
		return
	}
	list, err := h.dao.SelectDbTables(c.Request.Context(), c.Query("tableName"), c.Query("tableComment"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, genPtr(list))
}

// GetInfo GET /tool/gen/{tableId}。
func (h *GenHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "tool:gen:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("tableId"), 10, 64)
	m, err := h.dao.SelectGenTableById(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// ImportTable POST /tool/gen/importTable?tables=a,b（读 information_schema 装配表+列）。
func (h *GenHandler) ImportTable(c *gin.Context) {
	lu := requirePermAndLogin(c, "tool:gen:import")
	if lu == nil {
		return
	}
	tables := strings.Split(c.Query("tables"), ",")
	for _, t := range tables {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		tableID, err := h.dao.InsertGenTable(c.Request.Context(), &dao.GenTableDO{
			TableName: t, ClassName: toCamel(t), FunctionAuthor: lu.Username(),
		})
		if err != nil {
			response.Error(c, err.Error()).JSON()
			return
		}
		if err := h.dao.InsertGenColumns(c.Request.Context(), tableID, t); err != nil {
			response.Error(c, err.Error()).JSON()
			return
		}
	}
	response.Ok(c).JSON()
}

// Remove DELETE /tool/gen/{tableIds}。
func (h *GenHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "tool:gen:remove") == nil {
		return
	}
	if err := h.dao.DeleteGenTableByIds(c.Request.Context(), parseInt64List(c.Param("tableIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// toCamel 下划线转大驼峰（class_name 推断）。
func toCamel(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if p[0] >= 'a' && p[0] <= 'z' {
			b.WriteString(string(p[0]-32) + p[1:])
		} else {
			b.WriteString(p)
		}
	}
	return b.String()
}

func genPtr(list []dao.GenTableDO) []*dao.GenTableDO {
	ret := make([]*dao.GenTableDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}
