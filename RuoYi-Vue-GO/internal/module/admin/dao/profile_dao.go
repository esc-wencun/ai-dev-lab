// 个人中心 + 通用文件（对位 SysProfileController/CommonController/SysRegisterService 的数据访问）。
package dao

import (
	"context"

	"gorm.io/gorm"

	"ruoyi-vue-go/pkg/types"
)

// ProfileDao 个人中心数据访问。
type ProfileDao struct {
	db *gorm.DB
}

func NewProfileDao(db *gorm.DB) *ProfileDao { return &ProfileDao{db: db} }

// CheckPhoneUnique 手机号唯一（对位 checkPhoneUnique：非本人且未删除的存在即不唯一）。
func (d *ProfileDao) CheckPhoneUnique(ctx context.Context, phone string, userID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw(
		"select count(*) from sys_user where phonenumber = ? and user_id != ? and del_flag = '0'", phone, userID).Scan(&n).Error
	return n == 0, err
}

// CheckEmailUnique 邮箱唯一（对位 checkEmailUnique）。
func (d *ProfileDao) CheckEmailUnique(ctx context.Context, email string, userID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw(
		"select count(*) from sys_user where email = ? and user_id != ? and del_flag = '0'", email, userID).Scan(&n).Error
	return n == 0, err
}

// UpdateProfile 修改个人信息四字段（对位 updateUserProfile → updateUser 动态 set；
// 本端点仅 nickName/email/phonenumber/sex 四字段被赋值）。
func (d *ProfileDao) UpdateProfile(ctx context.Context, userID int64, nickName, email, phone, sex string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set nick_name = ?, email = ?, phonenumber = ?, sex = ?, update_time = now() where user_id = ?",
		nickName, email, phone, sex, userID).Error
}

// ResetUserPwd 重置密码（对位 resetUserPwd：pwd_update_date + password + update_time）。
func (d *ProfileDao) ResetUserPwd(ctx context.Context, userID int64, hash string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set pwd_update_date = now(), password = ?, update_time = now() where user_id = ?",
		hash, userID).Error
}

// UpdateAvatar 更新头像（对位 updateUserAvatar）。
func (d *ProfileDao) UpdateAvatar(ctx context.Context, userID int64, avatar string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set avatar = ?, update_time = now() where user_id = ?", avatar, userID).Error
}

// RoleNamesByUser 角色名集合（对位 selectRolesByUserName + joining(",") 的取数部分）。
func (d *ProfileDao) RoleNamesByUser(ctx context.Context, userName string) ([]string, error) {
	var names []string
	err := d.db.WithContext(ctx).Raw(`
		select distinct r.role_name from sys_role r
			 left join sys_user_role ur on ur.role_id = r.role_id
			 left join sys_user u on u.user_id = ur.user_id
		 WHERE r.del_flag = '0' and u.user_name = ?`, userName).Scan(&names).Error
	return names, err
}

// PostNamesByUser 岗位名集合（对位 selectPostsByUserName）。
func (d *ProfileDao) PostNamesByUser(ctx context.Context, userName string) ([]string, error) {
	var names []string
	err := d.db.WithContext(ctx).Raw(`
		select distinct p.post_name from sys_post p
			 left join sys_user_post up on up.post_id = p.post_id
			 left join sys_user u on u.user_id = up.user_id
		 where u.user_name = ?`, userName).Scan(&names).Error
	return names, err
}

// CheckUserNameUnique 用户名唯一（对位 checkUserNameUnique：存在未删除的同名即不唯一）。
func (d *ProfileDao) CheckUserNameUnique(ctx context.Context, userName string) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw(
		"select count(*) from sys_user where user_name = ? and del_flag = '0'", userName).Scan(&n).Error
	return n == 0, err
}

// InsertUser 注册落库（对位 registerUser → insertUser：nickName=username、加密密码、pwd_update_date；
// Java registerUser 仅插 sys_user 主表，不挂角色）。时间由 Go 侧传参（MySQL now() 与 sqlite 兼容性隔离在 dao 之外）。
func (d *ProfileDao) InsertUser(ctx context.Context, userName, nickName, hash string, pwdUpdateDate types.DateTime) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_user (user_name, nick_name, password, status, del_flag, create_time, pwd_update_date) values (?, ?, ?, '0', '0', ?, ?)",
		userName, nickName, hash, pwdUpdateDate, pwdUpdateDate).Error
}
