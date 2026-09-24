// 用户管理数据访问（对位 SysUserMapper.selectUserList/insertUser/updateUser/deleteUserByIds 等）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ruoyi-vue-go/pkg/types"
)

// UserDAO 用户管理。
type UserDAO struct{ db *gorm.DB }

func NewUserDAO(db *gorm.DB) *UserDAO { return &UserDAO{db: db} }

// UserListRow 列表行（对位 selectUserList 联查 dept 的扁平形态）。
type UserListRow struct {
	UserID      int64          `gorm:"column:user_id" json:"userId"`
	DeptID      int64          `gorm:"column:dept_id" json:"deptId"`
	NickName    string         `gorm:"column:nick_name" json:"nickName"`
	UserName    string         `gorm:"column:user_name" json:"userName"`
	Email       string         `gorm:"column:email" json:"email"`
	Avatar      string         `gorm:"column:avatar" json:"avatar"`
	Phonenumber string         `gorm:"column:phonenumber" json:"phonenumber"`
	Sex         string         `gorm:"column:sex" json:"sex"`
	Status      string         `gorm:"column:status" json:"status"`
	DelFlag     string         `gorm:"column:del_flag" json:"delFlag"`
	LoginIP     string         `gorm:"column:login_ip" json:"loginIp"`
	LoginDate   types.DateTime `gorm:"column:login_date" json:"loginDate"`
	CreateBy    string         `gorm:"column:create_by" json:"createBy"`
	CreateTime  types.DateTime `gorm:"column:create_time" json:"createTime"`
	Remark      string         `gorm:"column:remark" json:"remark"`
	DeptName    string         `gorm:"column:dept_name" json:"dept"`
	Leader      string         `gorm:"column:leader" json:"leader"`
}

// SelectUserList 条件列表（对位 selectUserList；@DataScope 由调用点以 Scopes 注入）。
func (d *UserDAO) SelectUserList(ctx context.Context, userName, phonenumber, status, deptID string, deptAncestorsLike string) ([]UserListRow, error) {
	where := "where u.del_flag = '0'"
	var args []any
	if userName != "" {
		where += " AND u.user_name like concat('%', ?, '%')"
		args = append(args, userName)
	}
	if phonenumber != "" {
		where += " AND u.phonenumber like concat('%', ?, '%')"
		args = append(args, phonenumber)
	}
	if status != "" {
		where += " AND u.status = ?"
		args = append(args, status)
	}
	if deptID != "" && deptID != "0" {
		// 对位 deptId 条件： IN (SELECT dept_id FROM sys_dept WHERE dept_id = ? OR find_in_set(?, ancestors))
		where += " AND u.dept_id IN (SELECT dept_id FROM sys_dept WHERE dept_id = ? OR find_in_set(?, ancestors))"
		args = append(args, deptID, deptID)
	}
	var list []UserListRow
	err := d.db.WithContext(ctx).Raw(`select u.user_id, u.dept_id, u.nick_name, u.user_name, u.email, u.avatar, u.phonenumber, u.sex, u.status, u.del_flag, u.login_ip, u.login_date, u.create_by, u.create_time, u.remark, d.dept_name, d.leader
		from sys_user u left join sys_dept d on u.dept_id = d.dept_id `+where+" order by u.user_id", args...).Scan(&list).Error
	return list, err
}

// SelectUserById 详情（对位 selectUserById 三表联查 → UserDetail 复用）。
func (d *UserDAO) SelectUserById(ctx context.Context, userID int64) (*UserDetail, error) {
	var rows []userJoinRow
	if err := d.db.WithContext(ctx).Raw(userWithDeptRole+" where u.user_id = ? and u.del_flag = '0'", userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return mergeUserRows(rows), nil
}

// CheckUserNameUnique 登录账号唯一（对位 checkUserNameUnique）。
func (d *UserDAO) CheckUserNameUnique(ctx context.Context, userName string) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(*) from sys_user where user_name = ? and del_flag = '0'", userName).Scan(&n).Error
	return n == 0, err
}

// CheckPhoneUnique / CheckEmailUnique（含本人排除）。
func (d *UserDAO) CheckPhoneUnique(ctx context.Context, phone string, userID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(*) from sys_user where phonenumber = ? and user_id != ? and del_flag = '0'", phone, userID).Scan(&n).Error
	return n == 0, err
}

func (d *UserDAO) CheckEmailUnique(ctx context.Context, email string, userID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(*) from sys_user where email = ? and user_id != ? and del_flag = '0'", email, userID).Scan(&n).Error
	return n == 0, err
}

// InsertUser 新增用户主表（对位 insertUser 常用字段）。
func (d *UserDAO) InsertUser(ctx context.Context, userName, nickName, email, phone, sex, hash, status, remark, createBy string, deptID int64, pwdUpdate types.DateTime) (int64, error) {
	res := d.db.WithContext(ctx).Exec(
		"insert into sys_user (dept_id, user_name, nick_name, email, phonenumber, sex, password, status, del_flag, create_by, create_time, remark, pwd_update_date) values (?, ?, ?, ?, ?, ?, ?, ?, '0', ?, now(), ?, ?)",
		deptID, userName, nickName, email, phone, sex, hash, status, createBy, remark, pwdUpdate)
	if res.Error != nil {
		return 0, res.Error
	}
	var id int64
	d.db.WithContext(ctx).Raw("select user_id from sys_user where user_name = ? order by user_id desc limit 1", userName).Scan(&id)
	return id, nil
}

// UpdateUser 修改（对位 updateUser 常用字段；不改密码）。
func (d *UserDAO) UpdateUser(ctx context.Context, userID int64, deptID int64, nickName, email, phone, sex, status, remark, updateBy string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set dept_id = ?, nick_name = ?, email = ?, phonenumber = ?, sex = ?, status = ?, update_by = ?, update_time = now(), remark = ? where user_id = ?",
		deptID, nickName, email, phone, sex, status, updateBy, remark, userID).Error
}

// ResetUserPwd 管理员重置密码（对位 resetPwd：pwd_update_date 同步更新）。
func (d *UserDAO) ResetUserPwd(ctx context.Context, userID int64, hash, updateBy string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set pwd_update_date = now(), password = ?, update_by = ?, update_time = now() where user_id = ?",
		hash, updateBy, userID).Error
}

// UpdateUserStatus 状态修改（对位 updateUserStatus）。
func (d *UserDAO) UpdateUserStatus(ctx context.Context, userID int64, status, updateBy string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set status = ?, update_by = ?, update_time = now() where user_id = ?", status, updateBy, userID).Error
}

// DeleteUserByIds 批量软删 + 清关联（对位 deleteUserByIds 事务三连）。
func (d *UserDAO) DeleteUserByIds(ctx context.Context, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",")
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_user_role where user_id in ("+ph+")", args...).Error; err != nil {
			return err
		}
		if err := tx.Exec("delete from sys_user_post where user_id in ("+ph+")", args...).Error; err != nil {
			return err
		}
		return tx.Exec("update sys_user set del_flag = '2' where user_id in ("+ph+")", args...).Error
	})
}

// RoleIDsByUser 用户角色 ID 列表。
func (d *UserDAO) RoleIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := d.db.WithContext(ctx).Raw("select role_id from sys_user_role where user_id = ?", userID).Scan(&ids).Error
	return ids, err
}

// PostIDsByUser 用户岗位 ID 列表（对位 selectPostListByUserId）。
func (d *UserDAO) PostIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := d.db.WithContext(ctx).Raw("select post_id from sys_user_post where user_id = ?", userID).Scan(&ids).Error
	return ids, err
}

// ReplaceUserRoles 重置用户角色关联（对位 deleteUserRoleByUserId + insertUserRole）。
func (d *UserDAO) ReplaceUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_user_role where user_id = ?", userID).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Exec("insert into sys_user_role (user_id, role_id) values (?, ?)", userID, rid).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceUserPosts 重置用户岗位关联。
func (d *UserDAO) ReplaceUserPosts(ctx context.Context, userID int64, postIDs []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from sys_user_post where user_id = ?", userID).Error; err != nil {
			return err
		}
		for _, pid := range postIDs {
			if err := tx.Exec("insert into sys_user_post (user_id, post_id) values (?, ?)", userID, pid).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SelectRolesAll 全部角色（对位 selectRoleAll；getInfo 组 roles 用）。
func (d *UserDAO) SelectRolesAll(ctx context.Context) ([]RoleOption, error) {
	var list []RoleOption
	err := d.db.WithContext(ctx).Raw("select role_id, role_name, role_key, role_sort, data_scope, status from sys_role where del_flag = '0' order by role_sort").Scan(&list).Error
	return list, err
}

// RoleOption 角色选项（getInfo/授权角色）。
type RoleOption struct {
	RoleID    int64  `gorm:"column:role_id" json:"roleId"`
	RoleName  string `gorm:"column:role_name" json:"roleName"`
	RoleKey   string `gorm:"column:role_key" json:"roleKey"`
	RoleSort  int    `gorm:"column:role_sort" json:"roleSort"`
	DataScope string `gorm:"column:data_scope" json:"dataScope"`
	Status    string `gorm:"column:status" json:"status"`
	Flag      bool   `gorm:"-" json:"flag"`
}

// SelectPostsAll 全部岗位（对位 selectPostAll）。
func (d *UserDAO) SelectPostsAll(ctx context.Context) ([]PostDO, error) {
	var list []PostDO
	err := d.db.WithContext(ctx).Raw("select post_id, post_code, post_name, post_sort, status from sys_post order by post_sort").Scan(&list).Error
	return list, err
}

// RolesByUserIDs 用户已授权角色（对位 selectRolesByUserId，含 flag 标记）。
func (d *UserDAO) RolesByUser(ctx context.Context, userID int64) ([]RoleOption, error) {
	var list []RoleOption
	err := d.db.WithContext(ctx).Raw(`select r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.status,
		case when ur.user_id is null then 0 else 1 end as flag
		from sys_role r
		left join sys_user_role ur on ur.role_id = r.role_id and ur.user_id = ?
		where r.del_flag = '0' order by r.role_sort`, userID).Scan(&list).Error
	return list, err
}

// SelectUserByName 按用户名查（导入判存；对位 selectUserByUserName 的轻量版）。
func (d *UserDAO) SelectUserByName(ctx context.Context, userName string) (*UserDetail, error) {
	var rows []userJoinRow
	if err := d.db.WithContext(ctx).Raw(userWithDeptRole+" where u.user_name = ? and u.del_flag = '0'", userName).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return mergeUserRows(rows), nil
}

// UpdateUserImport 导入更新（不改密码/状态；对位 importUser 更新分支的 updateUser 调用）。
func (d *UserDAO) UpdateUserImport(ctx context.Context, userName, nickName, email, phone, sex, updateBy string, deptID int64) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set dept_id = ?, nick_name = ?, email = ?, phonenumber = ?, sex = ?, update_by = ?, update_time = now() where user_name = ?",
		deptID, nickName, email, phone, sex, updateBy, userName).Error
}
