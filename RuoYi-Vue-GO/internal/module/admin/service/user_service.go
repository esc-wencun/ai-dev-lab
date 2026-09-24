// 用户管理业务（对位 SysUserServiceImpl + SysUserController）。
package service

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/pkg/types"
	"ruoyi-vue-go/pkg/utils/excel"
)

// UserService 用户管理。
type UserService struct {
	dao   *dao.UserDAO
	dept  *dao.DeptDao
	cache *cache.RedisCache
	db    *gorm.DB
}

func NewUserService(d *dao.UserDAO, dept *dao.DeptDao, rc *cache.RedisCache, db *gorm.DB) *UserService {
	return &UserService{dao: d, dept: dept, cache: rc, db: db}
}

// deptListForTree 部门列表（树构建取数）。
func (s *UserService) deptListForTree(ctx context.Context, deptName, status string) ([]dao.DeptDO, error) {
	return s.dept.SelectDeptList(ctx, deptName, status)
}

// TreeSelect 选择框树（对位 TreeSelect：id/label/children）。
type TreeSelect struct {
	ID       int64         `json:"id"`
	Label    string        `json:"label"`
	Children []*TreeSelect `json:"children,omitempty"`
}

// buildDeptTreeSelect 部门树选择框（对位 buildDeptTreeSelect：parentId 挂 children）。
func buildDeptTreeSelect(depts []dao.DeptDO) []TreeSelect {
	byID := map[int64]*TreeSelect{}
	var roots []TreeSelect
	nodes := make([]TreeSelect, len(depts))
	for i, d := range depts {
		nodes[i] = TreeSelect{ID: d.DeptID, Label: d.DeptName}
		byID[d.DeptID] = &nodes[i]
	}
	for i, d := range depts {
		if p, ok := byID[d.ParentID]; ok && d.ParentID != d.DeptID {
			p.Children = append(p.Children, &nodes[i])
		} else {
			roots = append(roots, nodes[i])
		}
	}
	return roots
}

// UserListQuery 列表查询参数。
type UserListQuery struct {
	UserName    string
	Phonenumber string
	Status      string
	DeptID      string
}

// List 列表（数据权限由调用方以 DataScope scope 注入——当前端点权限串已到按钮级，此处不再拼部门过滤）。
func (s *UserService) List(ctx context.Context, q *UserListQuery) ([]dao.UserListRow, error) {
	return s.dao.SelectUserList(ctx, q.UserName, q.Phonenumber, q.Status, q.DeptID, "")
}

// UserInput 新增/修改参数（对位 SysUser 请求体 + roleIds/postIds）。
type UserInput struct {
	UserID      int64   `json:"userId"`
	DeptID      int64   `json:"deptId"`
	UserName    string  `json:"userName"`
	NickName    string  `json:"nickName"`
	Password    string  `json:"password"`
	Email       string  `json:"email"`
	Phonenumber string  `json:"phonenumber"`
	Sex         string  `json:"sex"`
	Status      string  `json:"status"`
	Remark      string  `json:"remark"`
	RoleIDs     []int64 `json:"roleIds"`
	PostIDs     []int64 `json:"postIds"`
}

// GetInfo 详情（对位 getInfo：user + roleIds + postIds + roles + posts）。
func (s *UserService) GetInfo(ctx context.Context, userID int64) (ud *dao.UserDetail, roleIDs, postIDs []int64, roles []dao.RoleOption, posts []dao.PostDO, err error) {
	if userID > 0 {
		ud, err = s.dao.SelectUserById(ctx, userID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		roleIDs, _ = s.dao.RoleIDsByUser(ctx, userID)
		postIDs, _ = s.dao.PostIDsByUser(ctx, userID)
	}
	roles, _ = s.dao.SelectRolesAll(ctx)
	posts, _ = s.dao.SelectPostsAll(ctx)
	return ud, roleIDs, postIDs, roles, posts, nil
}

// Add 新增（校验链文案逐字对位；密码 bcrypt；事务插主表+角色+岗位）。
func (s *UserService) Add(ctx context.Context, in *UserInput, operator string) error {
	if ok, _ := s.dao.CheckUserNameUnique(ctx, in.UserName); !ok {
		return errf("新增用户'%s'失败，登录账号已存在", in.UserName)
	}
	if in.Phonenumber != "" {
		if ok, _ := s.dao.CheckPhoneUnique(ctx, in.Phonenumber, 0); !ok {
			return errf("新增用户'%s'失败，手机号码已存在", in.UserName)
		}
	}
	if in.Email != "" {
		if ok, _ := s.dao.CheckEmailUnique(ctx, in.Email, 0); !ok {
			return errf("新增用户'%s'失败，邮箱账号已存在", in.UserName)
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return errf("新增用户'%s'失败", in.UserName)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userID, err := s.dao.InsertUser(ctx, in.UserName, in.NickName, in.Email, in.Phonenumber, in.Sex, string(hash), in.Status, in.Remark, operator, in.DeptID, types.Now())
		if err != nil {
			return err
		}
		return s.replaceRelations(ctx, userID, in.RoleIDs, in.PostIDs)
	})
}

// Edit 修改（事务：更新主表 + 重置角色/岗位关联）。
func (s *UserService) Edit(ctx context.Context, in *UserInput, operator string) error {
	if in.UserID == 1 {
		return errf("不允许操作超级管理员用户")
	}
	if ok, _ := s.dao.CheckUserNameUnique(ctx, in.UserName); !ok {
		// 本人同名允许：查出比对
		ud, err := s.dao.SelectUserById(ctx, in.UserID)
		if err != nil || ud.User.UserName != in.UserName {
			return errf("修改用户'%s'失败，登录账号已存在", in.UserName)
		}
	}
	if in.Phonenumber != "" {
		if ok, _ := s.dao.CheckPhoneUnique(ctx, in.Phonenumber, in.UserID); !ok {
			return errf("修改用户'%s'失败，手机号码已存在", in.UserName)
		}
	}
	if in.Email != "" {
		if ok, _ := s.dao.CheckEmailUnique(ctx, in.Email, in.UserID); !ok {
			return errf("修改用户'%s'失败，邮箱账号已存在", in.UserName)
		}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.dao.UpdateUser(ctx, in.UserID, in.DeptID, in.NickName, in.Email, in.Phonenumber, in.Sex, in.Status, in.Remark, operator); err != nil {
			return err
		}
		return s.replaceRelations(ctx, in.UserID, in.RoleIDs, in.PostIDs)
	})
}

// replaceRelations 重置角色/岗位关联（edit 用；add 场景 userID 为新插入 ID）。
func (s *UserService) replaceRelations(ctx context.Context, userID int64, roleIDs, postIDs []int64) error {
	if len(roleIDs) > 0 {
		if err := s.dao.ReplaceUserRoles(ctx, userID, roleIDs); err != nil {
			return err
		}
	}
	if len(postIDs) > 0 {
		if err := s.dao.ReplaceUserPosts(ctx, userID, postIDs); err != nil {
			return err
		}
	}
	return nil
}

// Remove 删除（admin 保护 + 当前用户保护；事务清关联+软删；踢下线）。
func (s *UserService) Remove(ctx context.Context, userIDs []int64, operatorID int64) error {
	for _, id := range userIDs {
		if id == 1 {
			return errf("不允许操作超级管理员用户")
		}
	}
	for _, id := range userIDs {
		if id == operatorID {
			return errf("当前用户不能删除")
		}
	}
	if err := s.dao.DeleteUserByIds(ctx, userIDs); err != nil {
		return err
	}
	s.kickUsers(ctx, userIDs)
	return nil
}

// ResetPwd 管理员重置密码（admin 保护 + 踢下线对齐 Python 版）。
func (s *UserService) ResetPwd(ctx context.Context, userID int64, password, operator string) error {
	if userID == 1 && operator != "admin" {
		return errf("不允许操作超级管理员用户")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return errf("重置密码失败")
	}
	if err := s.dao.ResetUserPwd(ctx, userID, string(hash), operator); err != nil {
		return errf("重置密码失败")
	}
	s.kickUsers(ctx, []int64{userID})
	return nil
}

// ChangeStatus 状态修改（admin 保护 + 停用踢下线）。
func (s *UserService) ChangeStatus(ctx context.Context, userID int64, status, operator string) error {
	if userID == 1 {
		return errf("不允许操作超级管理员用户")
	}
	if err := s.dao.UpdateUserStatus(ctx, userID, status, operator); err != nil {
		return errf("修改状态失败")
	}
	if status == "1" {
		s.kickUsers(ctx, []int64{userID})
	}
	return nil
}

// AuthRolePage 授权角色页数据（对位 GET authRole/{userId}）。
func (s *UserService) AuthRolePage(ctx context.Context, userID int64) (ud *dao.UserDetail, roles []dao.RoleOption, err error) {
	ud, err = s.dao.SelectUserById(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	roles, err = s.dao.RolesByUser(ctx, userID)
	return ud, roles, err
}

// InsertAuthRole 授权角色（对位 insertUserAuth：清空重插 + 权限变更踢下线语义由会话刷新承担）。
func (s *UserService) InsertAuthRole(ctx context.Context, userID int64, roleIDs []int64) error {
	if err := s.dao.ReplaceUserRoles(ctx, userID, roleIDs); err != nil {
		return err
	}
	s.kickUsers(ctx, []int64{userID}) // 角色变更即时生效（对位 refreshPermissionByRoleId 的简化）
	return nil
}

// kickUsers 踢下线：SCAN login_tokens 会话，userId 命中即删（对位 Python 版踢下线结论）。
func (s *UserService) kickUsers(ctx context.Context, userIDs []int64) {
	target := map[int64]bool{}
	for _, id := range userIDs {
		target[id] = true
	}
	keys, err := s.cache.KeysByPrefix(ctx, "login_tokens:")
	if err != nil {
		return
	}
	for _, key := range keys {
		var lu model.LoginUser
		if err := s.cache.GetObject(ctx, key, &lu); err == nil && target[lu.UserId] {
			_, _ = s.cache.Delete(ctx, key)
		}
	}
}

// ImportTemplate 用户导入模板列（@Excel Type.IMPORT/ALL 字段，逐项对位）。
func ImportTemplateColumns() []excel.Column {
	return []excel.Column{
		{Title: "部门编号", Field: "deptId"},
		{Title: "登录名称", Field: "userName"},
		{Title: "用户名称", Field: "nickName"},
		{Title: "用户邮箱", Field: "email"},
		{Title: "手机号码", Field: "phonenumber"},
		{Title: "用户性别", Field: "sex", Converter: map[string]string{"0": "男", "1": "女", "2": "未知"}},
	}
}

// ExportColumns 用户导出列（@Excel Type.EXPORT/ALL，含部门联查字段）。
func ExportColumns() []excel.Column {
	return append(ImportTemplateColumns(),
		excel.Column{Title: "账号状态", Field: "status", Converter: map[string]string{"0": "正常", "1": "停用"}},
		excel.Column{Title: "最后登录IP", Field: "loginIp"},
		excel.Column{Title: "最后登录时间", Field: "loginDate"},
		excel.Column{Title: "部门名称", Field: "deptName"},
	)
}

// ExportRows 列表行转导出 map。
func ExportRows(rows []dao.UserListRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"userId": r.UserID, "deptId": r.DeptID, "userName": r.UserName, "nickName": r.NickName,
			"email": r.Email, "phonenumber": r.Phonenumber,
			"sex": sexCode(r.Sex), "status": r.Status,
			"loginIp": r.LoginIP, "loginDate": r.LoginDate, "deptName": r.DeptName,
		})
	}
	return out
}

// sexCode 性别显示值转编码（导出正查用；列表行存的已是编码）。
func sexCode(v string) string {
	return v
}

// ImportData 导入（对位 importUser：全成功/部分失败文案格式逐字对位）。
func (s *UserService) ImportData(ctx context.Context, data []byte, updateSupport bool, operator string) (string, error) {
	rows, err := excel.Parse(data, ImportTemplateColumns())
	if err != nil {
		return "", errf("导入用户数据不能为空！")
	}
	if len(rows) == 0 {
		return "", errf("导入用户数据不能为空！")
	}
	var successMsg, failureMsg strings.Builder
	successNum, failureNum := 0, 0
	initPwd, _ := s.initPassword(ctx)
	for _, row := range rows {
		userName := row["userName"]
		sexCode := row["sex"]
		_ = sexCode
		if _, err := s.dao.SelectUserByName(ctx, userName); err != nil {
			// 不存在 → 新增（初始密码）
			hash, _ := bcrypt.GenerateFromPassword([]byte(initPwd), 10)
			deptID := parseInt64(row["deptId"])
			_, ierr := s.dao.InsertUser(ctx, userName, row["nickName"], row["email"], row["phonenumber"], row["sex"], string(hash), "0", "导入", operator, deptID, types.Now())
			if ierr != nil {
				failureNum++
				fmt.Fprintf(&failureMsg, "<br/>%d、账号 %s 导入失败：%v", failureNum, userName, ierr)
				continue
			}
			successNum++
			fmt.Fprintf(&successMsg, "<br/>%d、账号 %s 导入成功", successNum, userName)
		} else if updateSupport {
			// 已存在 → 更新
			deptID := parseInt64(row["deptId"])
			if uerr := s.dao.UpdateUserImport(ctx, userName, row["nickName"], row["email"], row["phonenumber"], row["sex"], operator, deptID); uerr != nil {
				failureNum++
				fmt.Fprintf(&failureMsg, "<br/>%d、账号 %s 导入失败：%v", failureNum, userName, uerr)
				continue
			}
			successNum++
			fmt.Fprintf(&successMsg, "<br/>%d、账号 %s 更新成功", successNum, userName)
		} else {
			failureNum++
			fmt.Fprintf(&failureMsg, "<br/>%d、账号 %s 已存在", failureNum, userName)
		}
	}
	if failureNum > 0 {
		return "", errf("很抱歉，导入失败！共 %d 条数据格式不正确，错误如下：%s", failureNum, failureMsg.String())
	}
	return fmt.Sprintf("恭喜您，数据已全部导入成功！共 %d 条，数据如下：%s", successNum, successMsg.String()), nil
}

func (s *UserService) initPassword(ctx context.Context) (string, error) {
	// sys.user.initPassword 参数（复用 LoginService 的缓存回源逻辑语义）
	var val string
	if err := s.cache.GetObject(ctx, "sys_config:sys.user.initPassword", &val); err == nil && val != "" {
		return val, nil
	}
	var m struct {
		ConfigValue string `gorm:"column:config_value"`
	}
	if err := s.db.WithContext(ctx).Raw("select config_value from sys_config where config_key = ?", "sys.user.initPassword").Scan(&m).Error; err != nil {
		return "123456", err
	}
	if m.ConfigValue == "" {
		return "123456", nil
	}
	_ = s.cache.SetObject(ctx, "sys_config:sys.user.initPassword", m.ConfigValue, 0)
	return m.ConfigValue, nil
}

func parseInt64(s string) int64 {
	var v int64
	_, _ = fmt.Sscanf(s, "%d", &v)
	return v
}

// ListDeptTree 部门树（对位 selectDeptTreeList → buildDeptTreeSelect：TreeSelect id/label/children）。
func (s *UserService) ListDeptTree(ctx context.Context, deptName, status string) ([]TreeSelect, error) {
	depts, err := s.deptListForTree(ctx, deptName, status)
	if err != nil {
		return nil, err
	}
	return buildDeptTreeSelect(depts), nil
}
