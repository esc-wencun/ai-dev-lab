// 用户管理 HTTP 端点（对位 SysUserController 13 端点）。
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"encoding/json"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/pkg/response"
	"ruoyi-vue-go/pkg/utils/excel"
)

// UserHandler 用户管理。
type UserHandler struct {
	user *service.UserService
}

func NewUserHandler(u *service.UserService) *UserHandler { return &UserHandler{user: u} }

// List GET /system/user/list（TableDataInfo 分页 + 条件）。
func (h *UserHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:list") == nil {
		return
	}
	rows, err := h.user.List(c.Request.Context(), &service.UserListQuery{
		UserName: c.Query("userName"), Phonenumber: c.Query("phonenumber"),
		Status: c.Query("status"), DeptID: c.Query("deptId"),
	})
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	// 内存分页（对位 PageHelper 形态；列表已按 user_id 排序）
	pageNum := queryIntDef(c, "pageNum", 1)
	pageSize := queryIntDef(c, "pageSize", 10)
	total := len(rows)
	start := (pageNum - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	ptrs := make([]*dao.UserListRow, 0, end-start)
	for i := start; i < end; i++ {
		ptrs = append(ptrs, &rows[i])
	}
	c.JSON(200, gin.H{"code": 200, "msg": "查询成功", "rows": ptrs, "total": total})
}

// Export POST /system/user/export（xlsx 流，Excel 列对位 @Excel 注解）。
func (h *UserHandler) Export(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:export") == nil {
		return
	}
	rows, err := h.user.List(c.Request.Context(), &service.UserListQuery{
		UserName: c.Query("userName"), Phonenumber: c.Query("phonenumber"),
		Status: c.Query("status"), DeptID: c.Query("deptId"),
	})
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	excel.Export(c.Writer, "用户数据", service.ExportColumns(), service.ExportRows(rows))
}

// ImportTemplate POST /system/user/importTemplate（模板下载：表头+示例行）。
func (h *UserHandler) ImportTemplate(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:import") == nil {
		return
	}
	example := []map[string]any{{
		"deptId": int64(100), "userName": "ruoyi", "nickName": "若依",
		"email": "ry@163.com", "phonenumber": "15888888888", "sex": "男",
	}}
	excel.Export(c.Writer, "用户数据", service.ImportTemplateColumns(), example)
}

// ImportData POST /system/user/importData（multipart file + updateSupport）。
func (h *UserHandler) ImportData(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:import")
	if lu == nil {
		return
	}
	file, err := c.FormFile("file")
	if err != nil || file == nil {
		response.Error(c, "导入用户数据不能为空！").JSON()
		return
	}
	src, err := file.Open()
	if err != nil {
		response.Error(c, "读取文件失败").JSON()
		return
	}
	defer src.Close()
	buf := make([]byte, file.Size)
	_, _ = src.Read(buf)
	updateSupport := c.PostForm("updateSupport") == "true" || c.PostForm("updateSupport") == "1"
	msg, err := h.user.ImportData(c.Request.Context(), buf, updateSupport, lu.Username())
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).Msg(msg).JSON()
}

// GetInfo GET /system/user/ 与 /system/user/{userId}。
func (h *UserHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:query") == nil {
		return
	}
	var userID int64
	if p := c.Param("userId"); p != "" {
		userID, _ = strconv.ParseInt(p, 10, 64)
	}
	ud, roleIDs, postIDs, roles, posts, err := h.user.GetInfo(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	ajax := response.Ok(c)
	if ud != nil {
		ajax.Data(json.RawMessage(mustJSON(service.UserSessionView(ud)))).
			Put("roleIds", roleIDs).
			Put("postIds", postIDs)
		// roles 过滤 admin 角色（对位 !r.isAdmin()——role_id=1）
		filtered := make([]dao.RoleOption, 0, len(roles))
		for _, r := range roles {
			if r.RoleID != 1 {
				filtered = append(filtered, r)
			}
		}
		ajax.Put("roles", filtered)
	} else {
		ajax.Put("roles", roles)
	}
	ajax.Put("posts", posts).JSON()
}

// Add POST /system/user。
func (h *UserHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:add")
	if lu == nil {
		return
	}
	var body service.UserInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.user.Add(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Edit PUT /system/user。
func (h *UserHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:edit")
	if lu == nil {
		return
	}
	var body service.UserInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.user.Edit(c.Request.Context(), &body, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Remove DELETE /system/user/{userIds}。
func (h *UserHandler) Remove(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:remove")
	if lu == nil {
		return
	}
	var ids []int64
	for _, s := range strings.Split(c.Param("userIds"), ",") {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	if err := h.user.Remove(c.Request.Context(), ids, lu.UserId); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ResetPwd PUT /system/user/resetPwd。
func (h *UserHandler) ResetPwd(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:resetPwd")
	if lu == nil {
		return
	}
	var body struct {
		UserID   int64  `json:"userId"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.user.ResetPwd(c.Request.Context(), body.UserID, body.Password, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// ChangeStatus PUT /system/user/changeStatus。
func (h *UserHandler) ChangeStatus(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:edit")
	if lu == nil {
		return
	}
	var body struct {
		UserID int64  `json:"userId"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.user.ChangeStatus(c.Request.Context(), body.UserID, body.Status, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// AuthRolePage GET /system/user/authRole/{userId}。
func (h *UserHandler) AuthRolePage(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	ud, roles, err := h.user.AuthRolePage(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	ajax := response.Ok(c)
	if ud != nil {
		ajax.Put("user", json.RawMessage(mustJSON(service.UserSessionView(ud))))
	}
	// admin 用户不过滤角色列表（对位 isAdmin ? roles : filter）
	if ud != nil && ud.User.UserID == 1 {
		ajax.Put("roles", roles)
	} else {
		filtered := make([]dao.RoleOption, 0, len(roles))
		for _, r := range roles {
			if r.RoleID != 1 {
				filtered = append(filtered, r)
			}
		}
		ajax.Put("roles", filtered)
	}
	ajax.JSON()
}

// InsertAuthRole PUT /system/user/authRole?userId=&roleIds=（query 参数，对位 Java 无注解参数绑定）。
func (h *UserHandler) InsertAuthRole(c *gin.Context) {
	lu := requirePermAndLogin(c, "system:user:edit")
	if lu == nil {
		return
	}
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	var roleIDs []int64
	if rs := c.Query("roleIds"); rs != "" {
		for _, s := range strings.Split(rs, ",") {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				roleIDs = append(roleIDs, id)
			}
		}
	}
	if err := h.user.InsertAuthRole(c.Request.Context(), userID, roleIDs); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// DeptTree GET /system/user/deptTree（部门树 TreeSelect 形态：id/label/children）。
func (h *UserHandler) DeptTree(c *gin.Context) {
	if requirePermAndLogin(c, "system:user:list") == nil {
		return
	}
	list, err := h.user.ListDeptTree(c.Request.Context(), c.Query("deptName"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, list).JSON()
}

// guard：LoginUser 引用保留（会话上下文语义）。
var _ = model.LoginUser{}
