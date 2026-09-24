// 部门 + 岗位 HTTP 端点（对位 SysDeptController + SysPostController）。
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/pkg/response"
	"ruoyi-vue-go/pkg/utils/page"
)

// DeptHandler 部门管理。
type DeptHandler struct {
	dept *service.DeptService
}

func NewDeptHandler(d *service.DeptService) *DeptHandler { return &DeptHandler{dept: d} }

// requirePermAndLogin 组合校验：未登录 401；无权限 403（复用 middleware.HasPerm）。
func requirePermAndLogin(c *gin.Context, perm string) *model.LoginUser {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return nil
	}
	if !middleware.HasPerm(lu, perm) {
		response.Forbidden(c).JSON()
		return nil
	}
	return lu
}

// List GET /system/dept/list。
func (h *DeptHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:dept:list") == nil {
		return
	}
	list, err := h.dept.List(c.Request.Context(), c.Query("deptName"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, deptListPtr(list)).JSON()
}

// ListExclude GET /system/dept/list/exclude/{deptId}。
func (h *DeptHandler) ListExclude(c *gin.Context) {
	if requirePermAndLogin(c, "system:dept:list") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("deptId"), 10, 64)
	list, err := h.dept.ListExclude(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, deptListPtr(list)).JSON()
}

// GetInfo GET /system/dept/{deptId}。
func (h *DeptHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:dept:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("deptId"), 10, 64)
	m, err := h.dept.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Add POST /system/dept。
func (h *DeptHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dept:add")
	if lu == nil {
		return
	}
	var body struct {
		ParentID int64  `json:"parentId"`
		DeptName string `json:"deptName"`
		OrderNum int    `json:"orderNum"`
		Leader   string `json:"leader"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.dept.Add(c.Request.Context(), &service.DeptInput{
		ParentID: body.ParentID, DeptName: body.DeptName, OrderNum: body.OrderNum,
		Leader: body.Leader, Phone: body.Phone, Email: body.Email, Status: body.Status,
	}, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/dept。
func (h *DeptHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:dept:edit")
	if lu == nil {
		return
	}
	var body struct {
		DeptID   int64  `json:"deptId"`
		ParentID int64  `json:"parentId"`
		DeptName string `json:"deptName"`
		OrderNum int    `json:"orderNum"`
		Leader   string `json:"leader"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.dept.Edit(c.Request.Context(), &service.DeptInput{
		DeptID: body.DeptID, ParentID: body.ParentID, DeptName: body.DeptName, OrderNum: body.OrderNum,
		Leader: body.Leader, Phone: body.Phone, Email: body.Email, Status: body.Status,
	}, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/dept/{deptId}。
func (h *DeptHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:dept:remove") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("deptId"), 10, 64)
	warnMsg, err := h.dept.Remove(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	if warnMsg != "" {
		response.Warn(c, warnMsg).JSON()
		return
	}
	response.Ok(c).JSON()
}

// deptListPtr []dao.DeptDO → 指针切片（统一 JSON 形态）。
func deptListPtr(list []dao.DeptDO) []*dao.DeptDO {
	ret := make([]*dao.DeptDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// PostHandler 岗位管理。
type PostHandler struct {
	post *service.PostService
	db   interface{ ExecContextOK() bool } // 占位
}

func NewPostHandler(p *service.PostService) *PostHandler { return &PostHandler{post: p} }

// List GET /system/post/list（分页 TableDataInfo）。
func (h *PostHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:post:list") == nil {
		return
	}
	list, err := h.post.List(c.Request.Context(), c.Query("postCode"), c.Query("postName"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	var total int64 = int64(len(list))
	rows := postListPtr(list)
	// 内存分页（列表数据量小；对位 PageHelper 分页形态）
	pageNum := queryIntDef(c, "pageNum", 1)
	pageSize := queryIntDef(c, "pageSize", 10)
	start := (pageNum - 1) * pageSize
	if start > len(rows) {
		start = len(rows)
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	rows = rows[start:end]
	c.JSON(200, gin.H{"code": 200, "msg": "查询成功", "rows": rows, "total": total})
	_ = total
}

// GetInfo GET /system/post/{postId}。
func (h *PostHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:post:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("postId"), 10, 64)
	m, err := h.post.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Optionselect GET /system/post/optionselect。
func (h *PostHandler) Optionselect(c *gin.Context) {
	list, err := h.post.Optionselect(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, postListPtr(list)).JSON()
}

// Add POST /system/post。
func (h *PostHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:post:add")
	if lu == nil {
		return
	}
	var body service.PostInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	body.PostID = 0
	if err := h.post.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/post。
func (h *PostHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:post:edit")
	if lu == nil {
		return
	}
	var body service.PostInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.post.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/post/{postIds}（逗号分隔多值）。
func (h *PostHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "system:post:remove") == nil {
		return
	}
	var ids []int64
	for _, s := range strings.Split(c.Param("postIds"), ",") {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	if err := h.post.Remove(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

func postListPtr(list []dao.PostDO) []*dao.PostDO {
	ret := make([]*dao.PostDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

func queryIntDef(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// guard 编译期引用（分页/事务包在后续模块接入时复用）。
var (
	_ = page.Paginate
	_ = do.SysDept{}
)

// guard 编译期引用。
var _ = page.Paginate
