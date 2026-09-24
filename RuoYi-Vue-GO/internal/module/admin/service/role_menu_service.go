// 角色 + 菜单管理业务（对位 SysRoleServiceImpl/SysRoleController + SysMenuServiceImpl/SysMenuController）。
package service

import (
	"context"
	"strconv"
	"strings"

	bizerr "ruoyi-vue-go/internal/common/errors"
	"ruoyi-vue-go/internal/module/admin/dao"
)

// RoleService 角色。
type RoleService struct {
	role *dao.RoleDAO
	menu *dao.MenuDAO
}

func NewRoleService(r *dao.RoleDAO, m *dao.MenuDAO) *RoleService {
	return &RoleService{role: r, menu: m}
}

// RoleInput 角色（menuIds/deptIds 关联）。
type RoleInput struct {
	RoleID            int64   `json:"roleId"`
	RoleName          string  `json:"roleName"`
	RoleKey           string  `json:"roleKey"`
	RoleSort          int     `json:"roleSort"`
	DataScope         string  `json:"dataScope"`
	MenuCheckStrictly bool    `json:"menuCheckStrictly"`
	DeptCheckStrictly bool    `json:"deptCheckStrictly"`
	Status            string  `json:"status"`
	Remark            string  `json:"remark"`
	MenuIDs           []int64 `json:"menuIds"`
	DeptIDs           []int64 `json:"deptIds"`
}

// List 列表。
func (s *RoleService) List(ctx context.Context, roleName, roleKey, status string) ([]dao.RoleDO, error) {
	return s.role.SelectRoleList(ctx, roleName, roleKey, status)
}

// GetByID 详情（含 menuIds/deptIds）。
func (s *RoleService) GetByID(ctx context.Context, roleID int64) (*dao.RoleDO, error) {
	m, err := s.role.SelectRoleById(ctx, roleID)
	if err != nil {
		return nil, err
	}
	m.MenuIDs, _ = s.role.MenuIDsByRole(ctx, roleID, m.MenuCheckStrictly)
	m.DeptIDs, _ = s.role.DeptIDsByRole(ctx, roleID)
	return m, nil
}

// Optionselect 选择框。
func (s *RoleService) Optionselect(ctx context.Context) ([]dao.RoleDO, error) {
	return s.role.SelectRoleList(ctx, "", "", "0")
}

// Add 新增（文案逐字对位）。
func (s *RoleService) Add(ctx context.Context, in *RoleInput, operator string) error {
	if info, _ := s.role.CheckRoleNameUnique(ctx, in.RoleName); info != nil {
		return errf("新增角色'%s'失败，角色名称已存在", in.RoleName)
	}
	if info, _ := s.role.CheckRoleKeyUnique(ctx, in.RoleKey); info != nil {
		return errf("新增角色'%s'失败，角色权限已存在", in.RoleName)
	}
	m := &dao.RoleDO{RoleName: in.RoleName, RoleKey: in.RoleKey, RoleSort: in.RoleSort,
		DataScope: in.DataScope, MenuCheckStrictly: in.MenuCheckStrictly, DeptCheckStrictly: in.DeptCheckStrictly,
		Status: in.Status, Remark: in.Remark, CreateBy: operator}
	if err := s.role.InsertRole(ctx, m); err != nil {
		return err
	}
	return s.role.ReplaceRoleMenus(ctx, m.RoleID, in.MenuIDs)
}

// Edit 修改（admin 角色保护 + 关联重置）。
func (s *RoleService) Edit(ctx context.Context, in *RoleInput, operator string) error {
	if in.RoleID == 1 {
		return bizerr.New("不允许操作超级管理员角色")
	}
	if info, _ := s.role.CheckRoleNameUnique(ctx, in.RoleName); info != nil && info.RoleID != in.RoleID {
		return errf("修改角色'%s'失败，角色名称已存在", in.RoleName)
	}
	if info, _ := s.role.CheckRoleKeyUnique(ctx, in.RoleKey); info != nil && info.RoleID != in.RoleID {
		return errf("修改角色'%s'失败，角色权限已存在", in.RoleName)
	}
	m := &dao.RoleDO{RoleID: in.RoleID, RoleName: in.RoleName, RoleKey: in.RoleKey, RoleSort: in.RoleSort,
		DataScope: in.DataScope, MenuCheckStrictly: in.MenuCheckStrictly, DeptCheckStrictly: in.DeptCheckStrictly,
		Status: in.Status, Remark: in.Remark, CreateBy: operator}
	if err := s.role.UpdateRole(ctx, m); err != nil {
		return errf("修改角色'%s'失败，请联系管理员", in.RoleName)
	}
	return s.role.ReplaceRoleMenus(ctx, in.RoleID, in.MenuIDs)
}

// DataScope 数据权限（对位 authDataScope：更新角色 + 重置自定义部门）。
func (s *RoleService) DataScope(ctx context.Context, in *RoleInput) error {
	if in.RoleID == 1 {
		return bizerr.New("不允许操作超级管理员角色")
	}
	m := &dao.RoleDO{RoleID: in.RoleID, RoleName: in.RoleName, RoleKey: in.RoleKey, RoleSort: in.RoleSort,
		DataScope: in.DataScope, MenuCheckStrictly: in.MenuCheckStrictly, DeptCheckStrictly: in.DeptCheckStrictly,
		Status: in.Status, Remark: in.Remark}
	if err := s.role.UpdateRole(ctx, m); err != nil {
		return err
	}
	if in.DataScope == "2" { // 自定数据权限
		return s.role.ReplaceRoleDepts(ctx, in.RoleID, in.DeptIDs)
	}
	return nil
}

// ChangeStatus 状态修改。
func (s *RoleService) ChangeStatus(ctx context.Context, in *RoleInput, operator string) error {
	if in.RoleID == 1 {
		return bizerr.New("不允许操作超级管理员角色")
	}
	return s.role.UpdateRole(ctx, &dao.RoleDO{RoleID: in.RoleID, Status: in.Status, CreateBy: operator})
}

// Remove 批量删（admin 保护 + 已分配保护）。
func (s *RoleService) Remove(ctx context.Context, roleIDs []int64) error {
	for _, id := range roleIDs {
		if id == 1 {
			return bizerr.New("不允许操作超级管理员角色")
		}
		role, err := s.role.SelectRoleById(ctx, id)
		if err != nil {
			continue
		}
		if n, _ := s.role.CountUserRoleByRoleId(ctx, id); n > 0 {
			return errf("%s已分配,不能删除", role.RoleName)
		}
	}
	return s.role.DeleteRoleByIds(ctx, roleIDs)
}

// AllocatedUsers / UnallocatedUsers 分配用户列表。
func (s *RoleService) AllocatedUsers(ctx context.Context, roleID int64, userName, phonenumber string) ([]dao.UserListRow, error) {
	return s.role.AllocatedUsers(ctx, roleID, userName, phonenumber)
}

func (s *RoleService) UnallocatedUsers(ctx context.Context, roleID int64, userName, phonenumber string) ([]dao.UserListRow, error) {
	return s.role.UnallocatedUsers(ctx, roleID, userName, phonenumber)
}

// InsertAuthUsers 批量授权。
func (s *RoleService) InsertAuthUsers(ctx context.Context, roleID int64, userIDs []int64) error {
	return s.role.InsertAuthUsers(ctx, roleID, userIDs)
}

// DeleteAuthUser / DeleteAuthUsers 取消授权。
func (s *RoleService) DeleteAuthUser(ctx context.Context, userID, roleID int64) error {
	return s.role.DeleteAuthUser(ctx, userID, roleID)
}

func (s *RoleService) DeleteAuthUsers(ctx context.Context, roleID int64, userIDs []int64) error {
	return s.role.DeleteAuthUsers(ctx, roleID, userIDs)
}

// MenuService 菜单管理。
type MenuServiceMgmt struct {
	menu *dao.MenuDAO
}

func NewMenuServiceMgmt(m *dao.MenuDAO) *MenuServiceMgmt { return &MenuServiceMgmt{menu: m} }

// List 列表（扁平）。
func (s *MenuServiceMgmt) List(ctx context.Context, menuName, status string) ([]dao.MenuDO, error) {
	return s.menu.SelectMenuList(ctx, menuName, status)
}

// GetByID 详情。
func (s *MenuServiceMgmt) GetByID(ctx context.Context, menuID int64) (*dao.MenuDO, error) {
	return s.menu.SelectMenuById(ctx, menuID)
}

// TreeSelect 部门树式选择框（对位 /treeselect）。
func (s *MenuServiceMgmt) TreeSelect(ctx context.Context) ([]TreeSelect, error) {
	list, err := s.menu.SelectMenuTreeAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTreeSelect(list), nil
}

// RoleMenuTreeSelect 角色菜单树（对位 /roleMenuTreeselect/{roleId}：checkedKeys + menus）。
func (s *MenuServiceMgmt) RoleMenuTreeSelect(ctx context.Context, roleID int64) (checkedKeys []int64, menus []TreeSelect, err error) {
	var role dao.RoleDO
	role.MenuCheckStrictly = true
	m := &role
	_ = m
	checked, err := s.role_MenuIDs(ctx, roleID)
	if err != nil {
		return nil, nil, err
	}
	list, err := s.menu.SelectMenuTreeAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	return checked, buildMenuTreeSelect(list), nil
}

func (s *MenuServiceMgmt) role_MenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.menu.MenuIDsByRoleDirect(ctx, roleID)
}

// Add 新增（同父同名拒绝）。
func (s *MenuServiceMgmt) Add(ctx context.Context, in *dao.MenuDO) error {
	if info, _ := s.menu.CheckMenuNameUnique(ctx, in.MenuName, in.ParentID); info != nil {
		return errf("新增菜单'%s'失败，菜单名称已存在", in.MenuName)
	}
	if !validMenuParentID(in.ParentID) {
		return errf("新增菜单'%s'失败，上级菜单不能为自己", in.MenuName)
	}
	return s.menu.InsertMenu(ctx, in)
}

// Edit 修改（admin 菜单保护 + 上级不能自己 + 同名）。
func (s *MenuServiceMgmt) Edit(ctx context.Context, in *dao.MenuDO) error {
	if in.MenuID == 1 {
		return bizerr.New("不允许操作系统默认菜单")
	}
	if in.ParentID == in.MenuID {
		return errf("修改菜单'%s'失败，上级菜单不能选择自己", in.MenuName)
	}
	if info, _ := s.menu.CheckMenuNameUnique(ctx, in.MenuName, in.ParentID); info != nil && info.MenuID != in.MenuID {
		return errf("修改菜单'%s'失败，菜单名称已存在", in.MenuName)
	}
	return s.menu.UpdateMenu(ctx, in)
}

// Remove 删除（子菜单/已分配保护）。
func (s *MenuServiceMgmt) Remove(ctx context.Context, menuID int64) error {
	if has, _ := s.menu.HasChildByMenuId(ctx, menuID); has {
		return errf("存在子菜单,不允许删除")
	}
	if hasRole, _ := s.menu.CheckMenuExistRole(ctx, menuID); hasRole {
		return errf("菜单已分配,不允许删除")
	}
	return s.menu.DeleteMenuById(ctx, menuID)
}

// validMenuParentID 防御 parent=自身（新增时 ID 未定，恒 true；预留）。
func validMenuParentID(parentID int64) bool { return parentID >= 0 }

// parseInt64s 逗号分隔 ID 列表（排序端点预留）。
func parseInt64s(s string) []int64 {
	var out []int64
	for _, p := range strings.Split(s, ",") {
		if v, e := strconv.ParseInt(p, 10, 64); e == nil {
			out = append(out, v)
		}
	}
	return out
}

// buildMenuTreeSelect 菜单树选择框（对位 buildMenuTreeSelect）。
func buildMenuTreeSelect(menus []dao.MenuDO) []TreeSelect {
	byID := map[int64]*TreeSelect{}
	var roots []TreeSelect
	nodes := make([]TreeSelect, len(menus))
	for i, m := range menus {
		nodes[i] = TreeSelect{ID: m.MenuID, Label: m.MenuName}
		byID[m.MenuID] = &nodes[i]
	}
	for i, m := range menus {
		if p, ok := byID[m.ParentID]; ok && m.ParentID != m.MenuID {
			p.Children = append(p.Children, &nodes[i])
		} else {
			roots = append(roots, nodes[i])
		}
	}
	return roots
}
