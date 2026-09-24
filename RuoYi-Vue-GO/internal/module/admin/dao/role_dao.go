// 角色管理数据访问（对位 SysRoleMapper.xml）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ruoyi-vue-go/pkg/types"
)

// RoleDAO 角色。
type RoleDAO struct{ db *gorm.DB }

func NewRoleDAO(db *gorm.DB) *RoleDAO { return &RoleDAO{db: db} }

const roleColumns = `select r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly, r.status, r.del_flag, r.create_by, r.create_time, r.remark from sys_role r`

// RoleDO 角色模型。
type RoleDO struct {
	RoleID            int64          `gorm:"column:role_id" json:"roleId"`
	RoleName          string         `gorm:"column:role_name" json:"roleName"`
	RoleKey           string         `gorm:"column:role_key" json:"roleKey"`
	RoleSort          int            `gorm:"column:role_sort" json:"roleSort"`
	DataScope         string         `gorm:"column:data_scope" json:"dataScope"`
	MenuCheckStrictly bool           `gorm:"column:menu_check_strictly" json:"menuCheckStrictly"`
	DeptCheckStrictly bool           `gorm:"column:dept_check_strictly" json:"deptCheckStrictly"`
	Status            string         `gorm:"column:status" json:"status"`
	DelFlag           string         `gorm:"column:del_flag" json:"delFlag"`
	CreateBy          string         `gorm:"column:create_by" json:"createBy"`
	CreateTime        types.DateTime `gorm:"column:create_time" json:"createTime"`
	Remark            string         `gorm:"column:remark" json:"remark"`
	Flag              bool           `gorm:"-" json:"flag"`
	MenuIDs           []int64        `gorm:"-" json:"menuIds,omitempty"`
	DeptIDs           []int64        `gorm:"-" json:"deptIds,omitempty"`
}

// SelectRoleList 条件列表（对位 selectRoleList：roleName/roleKey like、status；del_flag='0'，order by role_sort）。
func (d *RoleDAO) SelectRoleList(ctx context.Context, roleName, roleKey, status string) ([]RoleDO, error) {
	where := "where r.del_flag = '0'"
	var args []any
	if roleName != "" {
		where += " AND r.role_name like concat('%', ?, '%')"
		args = append(args, roleName)
	}
	if roleKey != "" {
		where += " AND r.role_key like concat('%', ?, '%')"
		args = append(args, roleKey)
	}
	if status != "" {
		where += " AND r.status = ?"
		args = append(args, status)
	}
	var list []RoleDO
	err := d.db.WithContext(ctx).Raw(roleColumns+" "+where+" order by r.role_sort", args...).Scan(&list).Error
	return list, err
}

// SelectRoleById 按 ID。
func (d *RoleDAO) SelectRoleById(ctx context.Context, roleID int64) (*RoleDO, error) {
	var m RoleDO
	err := d.db.WithContext(ctx).Raw(roleColumns+" where r.role_id = ?", roleID).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.RoleID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

// CheckRoleNameUnique / CheckRoleKeyUnique（limit 1）。
func (d *RoleDAO) CheckRoleNameUnique(ctx context.Context, name string) (*RoleDO, error) {
	var m RoleDO
	err := d.db.WithContext(ctx).Raw(roleColumns+" where r.role_name = ? and r.del_flag = '0' limit 1", name).Scan(&m).Error
	if err != nil || m.RoleID == 0 {
		return nil, err
	}
	return &m, nil
}

func (d *RoleDAO) CheckRoleKeyUnique(ctx context.Context, key string) (*RoleDO, error) {
	var m RoleDO
	err := d.db.WithContext(ctx).Raw(roleColumns+" where r.role_key = ? and r.del_flag = '0' limit 1", key).Scan(&m).Error
	if err != nil || m.RoleID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertRole 新增。
func (d *RoleDAO) InsertRole(ctx context.Context, m *RoleDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_role (role_name, role_key, role_sort, data_scope, menu_check_strictly, dept_check_strictly, status, del_flag, create_by, create_time, remark) values (?, ?, ?, ?, ?, ?, ?, '0', ?, now(), ?)",
		m.RoleName, m.RoleKey, m.RoleSort, m.DataScope, m.MenuCheckStrictly, m.DeptCheckStrictly, m.Status, m.CreateBy, m.Remark).Error
}

// UpdateRole 修改。
func (d *RoleDAO) UpdateRole(ctx context.Context, m *RoleDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_role set role_name = ?, role_key = ?, role_sort = ?, data_scope = ?, menu_check_strictly = ?, dept_check_strictly = ?, status = ?, update_by = ?, update_time = now(), remark = ? where role_id = ?",
		m.RoleName, m.RoleKey, m.RoleSort, m.DataScope, m.MenuCheckStrictly, m.DeptCheckStrictly, m.Status, m.CreateBy, m.Remark, m.RoleID).Error
}

// CountUserRoleByRoleId 角色使用数（对位 countUserRoleByRoleId）。
func (d *RoleDAO) CountUserRoleByRoleId(ctx context.Context, roleID int64) (int64, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_user_role where role_id = ?", roleID).Scan(&n).Error
	return n, err
}

// DeleteRoleByIds 批量删（对位 deleteRoleByIds 事务：角色+菜单关联+部门关联；物理 delete）。
func (d *RoleDAO) DeleteRoleByIds(ctx context.Context, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(roleIDs)), ",")
	args := make([]any, len(roleIDs))
	for i, id := range roleIDs {
		args[i] = id
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_role_menu where role_id in ("+ph+")", args...).Error; err != nil {
			return err
		}
		if err := tx.Exec("delete from sys_role_dept where role_id in ("+ph+")", args...).Error; err != nil {
			return err
		}
		return tx.Exec("delete from sys_role where role_id in ("+ph+")", args...).Error
	})
}

// MenuIDsByRole 角色菜单 ID（对位 selectMenuListByRoleId；menuCheckStrictly 排除父节点）。
func (d *RoleDAO) MenuIDsByRole(ctx context.Context, roleID int64, strictly bool) ([]int64, error) {
	q := `select m.menu_id from sys_menu m left join sys_role_menu rm on m.menu_id = rm.menu_id where rm.role_id = ?`
	if strictly {
		q += " and m.menu_id not in (select m2.parent_id from sys_menu m2 inner join sys_role_menu rm2 on m2.menu_id = rm2.menu_id and rm2.role_id = ?)"
	}
	q += " order by m.parent_id, m.order_num"
	args := []any{roleID}
	if strictly {
		args = append(args, roleID)
	}
	var ids []int64
	err := d.db.WithContext(ctx).Raw(q, args...).Scan(&ids).Error
	return ids, err
}

// DeptIDsByRole 角色部门 ID（自定数据权限）。
func (d *RoleDAO) DeptIDsByRole(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := d.db.WithContext(ctx).Raw("select dept_id from sys_role_dept where role_id = ?", roleID).Scan(&ids).Error
	return ids, err
}

// ReplaceRoleMenus 重置角色菜单关联。
func (d *RoleDAO) ReplaceRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_role_menu where role_id = ?", roleID).Error; err != nil {
			return err
		}
		for _, id := range menuIDs {
			if err := tx.Exec("insert into sys_role_menu (role_id, menu_id) values (?, ?)", roleID, id).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceRoleDepts 重置角色部门关联。
func (d *RoleDAO) ReplaceRoleDepts(ctx context.Context, roleID int64, deptIDs []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_role_dept where role_id = ?", roleID).Error; err != nil {
			return err
		}
		for _, id := range deptIDs {
			if err := tx.Exec("insert into sys_role_dept (role_id, dept_id) values (?, ?)", roleID, id).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AllocatedUsers 已分配用户（对位 selectAllocatedList：命中 userName/phonenumber 过滤在外层）。
func (d *RoleDAO) AllocatedUsers(ctx context.Context, roleID int64, userName, phonenumber string) ([]UserListRow, error) {
	where := `select u.user_id, u.dept_id, u.nick_name, u.user_name, u.email, u.avatar, u.phonenumber, u.status, d.dept_name
		from sys_user u left join sys_dept d on u.dept_id = d.dept_id
		join sys_user_role ur on ur.user_id = u.user_id
		where u.del_flag = '0' and ur.role_id = ?`
	var args []any
	args = append(args, roleID)
	if userName != "" {
		where += " AND u.user_name like concat('%', ?, '%')"
		args = append(args, userName)
	}
	if phonenumber != "" {
		where += " AND u.phonenumber like concat('%', ?, '%')"
		args = append(args, phonenumber)
	}
	var list []UserListRow
	err := d.db.WithContext(ctx).Raw(where, args...).Scan(&list).Error
	return list, err
}

// UnallocatedUsers 未分配用户。
func (d *RoleDAO) UnallocatedUsers(ctx context.Context, roleID int64, userName, phonenumber string) ([]UserListRow, error) {
	where := `select u.user_id, u.dept_id, u.nick_name, u.user_name, u.email, u.avatar, u.phonenumber, u.status, d.dept_name
		from sys_user u left join sys_dept d on u.dept_id = d.dept_id
		where u.del_flag = '0' and u.user_id not in (select user_id from sys_user_role where role_id = ?)`
	var args []any
	args = append(args, roleID)
	if userName != "" {
		where += " AND u.user_name like concat('%', ?, '%')"
		args = append(args, userName)
	}
	if phonenumber != "" {
		where += " AND u.phonenumber like concat('%', ?, '%')"
		args = append(args, phonenumber)
	}
	var list []UserListRow
	err := d.db.WithContext(ctx).Raw(where, args...).Scan(&list).Error
	return list, err
}

// InsertAuthUsers 批量授权（对位 insertAuthUsers）。
func (d *RoleDAO) InsertAuthUsers(ctx context.Context, roleID int64, userIDs []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, uid := range userIDs {
			// 忽略重复（对位 insert ignore 语义，MySQL 主键冲突报错则跳过）
			var n int64
			if err := tx.Raw("select count(1) from sys_user_role where user_id = ? and role_id = ?", uid, roleID).Scan(&n).Error; err != nil {
				return err
			}
			if n == 0 {
				if err := tx.Exec("insert into sys_user_role (user_id, role_id) values (?, ?)", uid, roleID).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// DeleteAuthUser 取消授权（单/批）。
func (d *RoleDAO) DeleteAuthUser(ctx context.Context, userID, roleID int64) error {
	return d.db.WithContext(ctx).Exec("delete from sys_user_role where user_id = ? and role_id = ?", userID, roleID).Error
}

func (d *RoleDAO) DeleteAuthUsers(ctx context.Context, roleID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",")
	args := append([]any{roleID}, toAnySlice(userIDs)...)
	return d.db.WithContext(ctx).Exec("delete from sys_user_role where role_id = ? and user_id in ("+ph+")", args...).Error
}

func toAnySlice(ids []int64) []any {
	out := make([]any, len(ids))
	for i, v := range ids {
		out[i] = v
	}
	return out
}
