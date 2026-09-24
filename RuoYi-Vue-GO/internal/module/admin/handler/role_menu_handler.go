// 角色 + 菜单管理 HTTP 端点（对位 SysRoleController 13 端点 + SysMenuController 9 端点）。
package handler

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/pkg/response"
	"ruoyi-vue-go/pkg/utils/excel"
)

// RoleHandler 角色管理。
type RoleHandler struct{ role *service.RoleService }

func NewRoleHandler(r *service.RoleService) *RoleHandler { return &RoleHandler{role: r} }

// List GET /system/role/list（TableDataInfo）。
func (h *RoleHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:list") == nil {
		return
	}
	list, err := h.role.List(c.Request.Context(), c.Query("roleName"), c.Query("roleKey"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, roleListPtr(list))
}

// Export POST /system/role/export。
func (h *RoleHandler) Export(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:export") == nil {
		return
	}
	list, err := h.role.List(c.Request.Context(), c.Query("roleName"), c.Query("roleKey"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	cols := []excel.Column{
		{Title: "角色序号", Field: "roleId"},
		{Title: "角色名称", Field: "roleName"},
		{Title: "角色权限", Field: "roleKey"},
		{Title: "显示顺序", Field: "roleSort"},
		{Title: "状态", Field: "status", Converter: map[string]string{"0": "正常", "1": "停用"}},
		{Title: "创建时间", Field: "createTime"},
	}
	rows := make([]map[string]any, 0, len(list))
	for _, r := range list {
		rows = append(rows, map[string]any{"roleId": r.RoleID, "roleName": r.RoleName, "roleKey": r.RoleKey,
			"roleSort": r.RoleSort, "status": r.Status, "createTime": r.CreateTime})
	}
	excel.Export(c.Writer, "角色数据", cols, rows)
}

// GetInfo GET /system/role/{roleId}。
func (h *RoleHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("roleId"), 10, 64)
	m, err := h.role.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Optionselect GET /system/role/optionselect。
func (h *RoleHandler) Optionselect(c *gin.Context) {
	list, err := h.role.Optionselect(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, roleListPtr(list)).JSON()
}

// Add POST /system/role。
func (h *RoleHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:role:add")
	if lu == nil {
		return
	}
	var body service.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.role.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/role。
func (h *RoleHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:role:edit")
	if lu == nil {
		return
	}
	var body service.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.role.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// DataScope PUT /system/role/dataScope。
func (h *RoleHandler) DataScope(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:edit") == nil {
		return
	}
	var body service.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.role.DataScope(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ChangeStatus PUT /system/role/changeStatus。
func (h *RoleHandler) ChangeStatus(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:role:edit")
	if lu == nil {
		return
	}
	var body service.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.role.ChangeStatus(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/role/{roleIds}。
func (h *RoleHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:remove") == nil {
		return
	}
	ids := parseInt64List(c.Param("roleIds"))
	if err := h.role.Remove(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// AllocatedList GET /system/role/authUser/allocatedList。
func (h *RoleHandler) AllocatedList(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:list") == nil {
		return
	}
	roleID, _ := strconv.ParseInt(c.Query("roleId"), 10, 64)
	list, err := h.role.AllocatedUsers(c.Request.Context(), roleID, c.Query("userName"), c.Query("phonenumber"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, userListPtr(list))
}

// UnallocatedList GET /system/role/authUser/unallocatedList。
func (h *RoleHandler) UnallocatedList(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:list") == nil {
		return
	}
	roleID, _ := strconv.ParseInt(c.Query("roleId"), 10, 64)
	list, err := h.role.UnallocatedUsers(c.Request.Context(), roleID, c.Query("userName"), c.Query("phonenumber"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, userListPtr(list))
}

// CancelAuthUser PUT /system/role/authUser/cancel（body JSON {userId,roleId}）。
func (h *RoleHandler) CancelAuthUser(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:edit") == nil {
		return
	}
	var body struct {
		UserID int64 `json:"userId"`
		RoleID int64 `json:"roleId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.role.DeleteAuthUser(c.Request.Context(), body.UserID, body.RoleID); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// CancelAuthUserAll PUT /system/role/authUser/cancelAll?roleId=&userIds=。
func (h *RoleHandler) CancelAuthUserAll(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:edit") == nil {
		return
	}
	roleID, _ := strconv.ParseInt(c.Query("roleId"), 10, 64)
	if err := h.role.DeleteAuthUsers(c.Request.Context(), roleID, parseInt64List(c.Query("userIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// SelectAuthUserAll PUT /system/role/authUser/selectAll?roleId=&userIds=。
func (h *RoleHandler) SelectAuthUserAll(c *gin.Context) {
	if requirePermAndLogin(c, "system:role:edit") == nil {
		return
	}
	roleID, _ := strconv.ParseInt(c.Query("roleId"), 10, 64)
	if err := h.role.InsertAuthUsers(c.Request.Context(), roleID, parseInt64List(c.Query("userIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ---------- 菜单管理 ----------

// MenuHandler 菜单管理。
type MenuHandler struct{ menu *service.MenuServiceMgmt }

func NewMenuHandler(m *service.MenuServiceMgmt) *MenuHandler { return &MenuHandler{menu: m} }

// List GET /system/menu/list（扁平，不分页）。
func (h *MenuHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:list") == nil {
		return
	}
	list, err := h.menu.List(c.Request.Context(), c.Query("menuName"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	ptrs := make([]*dao.MenuDO, len(list))
	for i := range list {
		ptrs[i] = &list[i]
	}
	response.OkData(c, ptrs).JSON()
}

// GetInfo GET /system/menu/{menuId}。
func (h *MenuHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("menuId"), 10, 64)
	m, err := h.menu.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// TreeSelect GET /system/menu/treeselect。
func (h *MenuHandler) TreeSelect(c *gin.Context) {
	list, err := h.menu.TreeSelect(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, list).JSON()
}

// RoleMenuTreeSelect GET /system/menu/roleMenuTreeselect/{roleId}。
func (h *MenuHandler) RoleMenuTreeSelect(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("roleId"), 10, 64)
	checked, menus, err := h.menu.RoleMenuTreeSelect(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).Put("checkedKeys", checked).Put("menus", menus).JSON()
}

// Add POST /system/menu。
func (h *MenuHandler) Add(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:add") == nil {
		return
	}
	var body dao.MenuDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.menu.Add(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/menu。
func (h *MenuHandler) Edit(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:edit") == nil {
		return
	}
	var body dao.MenuDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.menu.Edit(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/menu/{menuId}。
func (h *MenuHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:menu:remove") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("menuId"), 10, 64)
	if err := h.menu.Remove(c.Request.Context(), id); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ---------- 公共辅助 ----------

// pagedJSON TableDataInfo 形态输出（内存分页）。
func pagedJSON(c *gin.Context, all any) {
	// 列表多为全量返回（角色/分配用户量级小），total 取切片长度（反射统一处理类型）
	rv := reflect.ValueOf(all)
	total := 0
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		total = rv.Len()
	}
	c.JSON(200, gin.H{"code": 200, "msg": "查询成功", "rows": all, "total": total})
}

func roleListPtr(list []dao.RoleDO) []*dao.RoleDO {
	ret := make([]*dao.RoleDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

func userListPtr(list []dao.UserListRow) []*dao.UserListRow {
	ret := make([]*dao.UserListRow, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// parseInt64List 逗号分隔 ID。
func parseInt64List(s string) []int64 {
	var out []int64
	for _, p := range strings.Split(s, ",") {
		if v, e := strconv.ParseInt(p, 10, 64); e == nil {
			out = append(out, v)
		}
	}
	return out
}
