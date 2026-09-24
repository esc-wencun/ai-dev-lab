// 菜单管理数据访问（对位 SysMenuMapper.xml 的管理端查询）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// MenuDAO 菜单。
type MenuDAO struct{ db *gorm.DB }

func NewMenuDAO(db *gorm.DB) *MenuDAO { return &MenuDAO{db: db} }

const menuColumns = `select m.menu_id, m.menu_name, m.parent_id, m.order_num, m.path, m.component, m.query, m.route_name, m.is_frame, m.is_cache, m.menu_type, m.visible, m.status, ifnull(m.perms,'') as perms, m.icon, m.create_time from sys_menu m`

// MenuDO 菜单模型（管理端）。
type MenuDO struct {
	MenuID    int64     `gorm:"column:menu_id" json:"menuId"`
	MenuName  string    `gorm:"column:menu_name" json:"menuName"`
	ParentID  int64     `gorm:"column:parent_id" json:"parentId"`
	OrderNum  int       `gorm:"column:order_num" json:"orderNum"`
	Path      string    `gorm:"column:path" json:"path"`
	Component string    `gorm:"column:component" json:"component"`
	Query     string    `gorm:"column:query" json:"query"`
	RouteName string    `gorm:"column:route_name" json:"routeName"`
	IsFrame   string    `gorm:"column:is_frame" json:"isFrame"`
	IsCache   string    `gorm:"column:is_cache" json:"isCache"`
	MenuType  string    `gorm:"column:menu_type" json:"menuType"`
	Visible   string    `gorm:"column:visible" json:"visible"`
	Status    string    `gorm:"column:status" json:"status"`
	Perms     string    `gorm:"column:perms" json:"perms"`
	Icon      string    `gorm:"column:icon" json:"icon"`
	Children  []*MenuDO `gorm:"-" json:"children,omitempty"`
}

// SelectMenuList 条件列表（对位 selectMenuList：menuName like/status；admin 全量）。
func (d *MenuDAO) SelectMenuList(ctx context.Context, menuName, status string) ([]MenuDO, error) {
	where := "where 1=1"
	var args []any
	if menuName != "" {
		where += " AND m.menu_name like concat('%', ?, '%')"
		args = append(args, menuName)
	}
	if status != "" {
		where += " AND m.status = ?"
		args = append(args, status)
	}
	var list []MenuDO
	err := d.db.WithContext(ctx).Raw(menuColumns+" "+where+" order by m.parent_id, m.order_num", args...).Scan(&list).Error
	return list, err
}

// SelectMenuById 按 ID。
func (d *MenuDAO) SelectMenuById(ctx context.Context, menuID int64) (*MenuDO, error) {
	var m MenuDO
	err := d.db.WithContext(ctx).Raw(menuColumns+" where m.menu_id = ?", menuID).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.MenuID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

// HasChildByMenuId 有子菜单。
func (d *MenuDAO) HasChildByMenuId(ctx context.Context, menuID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_menu where parent_id = ?", menuID).Scan(&n).Error
	return n > 0, err
}

// CheckMenuExistRole 菜单已分配角色。
func (d *MenuDAO) CheckMenuExistRole(ctx context.Context, menuID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_role_menu where menu_id = ?", menuID).Scan(&n).Error
	return n > 0, err
}

// CheckMenuNameUnique 同父同名。
func (d *MenuDAO) CheckMenuNameUnique(ctx context.Context, name string, parentID int64) (*MenuDO, error) {
	var m MenuDO
	err := d.db.WithContext(ctx).Raw(menuColumns+" where m.menu_name = ? and m.parent_id = ? limit 1", name, parentID).Scan(&m).Error
	if err != nil || m.MenuID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertMenu 新增。
func (d *MenuDAO) InsertMenu(ctx context.Context, m *MenuDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_menu (menu_name, parent_id, order_num, path, component, query, route_name, is_frame, is_cache, menu_type, visible, status, perms, icon, create_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())",
		m.MenuName, m.ParentID, m.OrderNum, m.Path, m.Component, m.Query, m.RouteName, m.IsFrame, m.IsCache, m.MenuType, m.Visible, m.Status, m.Perms, m.Icon).Error
}

// UpdateMenu 修改。
func (d *MenuDAO) UpdateMenu(ctx context.Context, m *MenuDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_menu set menu_name = ?, parent_id = ?, order_num = ?, path = ?, component = ?, query = ?, route_name = ?, is_frame = ?, is_cache = ?, menu_type = ?, visible = ?, status = ?, perms = ?, icon = ? where menu_id = ?",
		m.MenuName, m.ParentID, m.OrderNum, m.Path, m.Component, m.Query, m.RouteName, m.IsFrame, m.IsCache, m.MenuType, m.Visible, m.Status, m.Perms, m.Icon, m.MenuID).Error
}

// DeleteMenuById 删除（物理）。
func (d *MenuDAO) DeleteMenuById(ctx context.Context, menuID int64) error {
	return d.db.WithContext(ctx).Exec("delete from sys_menu where menu_id = ?", menuID).Error
}

// SelectMenuPermsByRoleId 菜单权限串（角色权限刷新用，复用）。
func (d *MenuDAO) SelectMenuPermsByRoleId(ctx context.Context, roleID int64) ([]string, error) {
	var perms []string
	err := d.db.WithContext(ctx).Raw("select distinct m.perms from sys_menu m left join sys_role_menu rm on m.menu_id = rm.menu_id where m.status = '0' and rm.role_id = ?", roleID).Scan(&perms).Error
	return perms, err
}

// SelectMenuTreeAll 全部目录+菜单（getRouters 复用；此处导出给 role roleMenuTree 用）。
func (d *MenuDAO) SelectMenuTreeAll(ctx context.Context) ([]MenuDO, error) {
	var list []MenuDO
	err := d.db.WithContext(ctx).Raw(menuColumns + " where m.menu_type in ('M', 'C') and m.status = 0 order by m.parent_id, m.order_num").Scan(&list).Error
	return list, err
}

// guard（strings 供后续扩展使用）。
var _ = strings.TrimSpace

// MenuIDsByRoleDirect 角色已勾选菜单 ID（roleMenuTreeselect checkedKeys）。
func (d *MenuDAO) MenuIDsByRoleDirect(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := d.db.Raw("select menu_id from sys_role_menu where role_id = ?", roleID).Scan(&ids).Error
	return ids, err
}
