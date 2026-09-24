// Package dao 数据访问层：一个业务域一个文件，仅查询与写库，禁止业务判断与缓存操作。
// 本文件对位登录闭环相关查询（SysUserMapper.selectUserByUserName / SysConfigMapper /
// SysLogininforMapper / SysUserMapper.updateLoginInfo / SysMenuMapper.selectMenuTree*）。
package dao

import (
	"context"

	"gorm.io/gorm"

	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/pkg/types"
)

// typesTime 别名简化签名。
type typesTime = types.DateTime

// LoginDao 登录闭环数据访问。
type LoginDao struct {
	db *gorm.DB
}

func NewLoginDao(db *gorm.DB) *LoginDao { return &LoginDao{db: db} }

// userWithDeptRole 列集（对位 SysUserMapper.selectUserVo：用户+部门+角色三表联查）。
const userWithDeptRole = `
	select u.user_id, u.dept_id, u.user_name, u.nick_name, u.email, u.avatar, u.phonenumber,
	       u.password, u.sex, u.status, u.del_flag, u.login_ip, u.login_date, u.pwd_update_date,
	       u.create_by, u.create_time, u.update_by, u.update_time, u.remark,
	       d.dept_id as dept_dept_id, d.parent_id as dept_parent_id, d.ancestors, d.dept_name,
	       d.order_num as dept_order_num, d.leader, d.status as dept_status,
	       r.role_id as role_role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.status as role_status
	from sys_user u
	left join sys_dept d on u.dept_id = d.dept_id
	left join sys_user_role ur on u.user_id = ur.user_id
	left join sys_role r on r.role_id = ur.role_id
`

// userJoinRow 扫描行（三表联查的扁平形态；多角色用户返回多行，LoadUserByName 合并）。
type userJoinRow struct {
	do.SysUser
	DeptDeptID   *int64  `gorm:"column:dept_dept_id"`
	DeptParentID *int64  `gorm:"column:dept_parent_id"`
	Ancestors    *string `gorm:"column:ancestors"`
	DeptName     *string `gorm:"column:dept_name"`
	DeptOrderNum *int    `gorm:"column:dept_order_num"`
	Leader       *string `gorm:"column:leader"`
	DeptStatus   *string `gorm:"column:dept_status"`
	RoleRoleID   *int64  `gorm:"column:role_role_id"`
	RoleName     *string `gorm:"column:role_name"`
	RoleKey      *string `gorm:"column:role_key"`
	RoleSort     *int    `gorm:"column:role_sort"`
	DataScope    *string `gorm:"column:data_scope"`
	RoleStatus   *string `gorm:"column:role_status"`
}

// UserDetail 用户+部门+角色聚合（对位 Java SysUser 内嵌 dept/roles 的形态）。
type UserDetail struct {
	User  do.SysUser
	Dept  *do.SysDept
	Roles []do.SysRole
}

// LoadUserByName 按用户名加载用户（del_flag='0'），带部门与角色（对位 selectUserByUserName）。
func (d *LoginDao) LoadUserByName(ctx context.Context, userName string) (*UserDetail, error) {
	var rows []userJoinRow
	if err := d.db.WithContext(ctx).Raw(userWithDeptRole+" where u.user_name = ? and u.del_flag = '0'", userName).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return mergeUserRows(rows), nil
}

// mergeUserRows 扁平行合并为一个聚合（Java resultMap 的 dept/roles 嵌套收集语义）。
func mergeUserRows(rows []userJoinRow) *UserDetail {
	first := rows[0]
	ud := &UserDetail{User: first.SysUser}
	if first.DeptDeptID != nil {
		ud.Dept = &do.SysDept{
			DeptID:    *first.DeptDeptID,
			ParentID:  deref(first.DeptParentID),
			Ancestors: derefStr(first.Ancestors),
			DeptName:  derefStr(first.DeptName),
			OrderNum:  derefInt(first.DeptOrderNum),
			Leader:    derefStr(first.Leader),
			Status:    derefStr(first.DeptStatus),
		}
	}
	seen := map[int64]bool{}
	for _, r := range rows {
		if r.RoleRoleID != nil && !seen[*r.RoleRoleID] {
			seen[*r.RoleRoleID] = true
			ud.Roles = append(ud.Roles, do.SysRole{
				RoleID:    *r.RoleRoleID,
				RoleName:  derefStr(r.RoleName),
				RoleKey:   derefStr(r.RoleKey),
				RoleSort:  derefInt(r.RoleSort),
				DataScope: derefStr(r.DataScope),
				Status:    derefStr(r.RoleStatus),
			})
		}
	}
	return ud
}

// GetConfigValue 参数值：先 Redis sys_config: 缓存、未命中回源 sys_config 并回填
// （对位 SysConfigServiceImpl.selectConfigByKey；缓存操作在 service 层，dao 只查库）。
// 规范注记：本方法供 ConfigService 调用，Redis 读写由 service 承担，dao 仅提供回源查询。
func (d *LoginDao) GetConfigByKey(ctx context.Context, configKey string) (string, error) {
	var m do.SysConfig
	err := d.db.WithContext(ctx).Where("config_key = ?", configKey).First(&m).Error
	if err != nil {
		return "", err
	}
	return m.ConfigValue, nil
}

// UpdateLoginInfo 登录成功后更新 login_ip/login_date（对位 SysUserMapper.updateLoginInfo）。
func (d *LoginDao) UpdateLoginInfo(ctx context.Context, userID int64, ip string, loginTime typesTime) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_user set login_ip = ?, login_date = ? where user_id = ?", ip, loginTime, userID).Error
}

// MenuTree 菜单树查询（admin 全量 / 非 admin 按角色，SQL 逐字对位 SysMenuMapper）。
func (d *LoginDao) MenuTreeAll(ctx context.Context) ([]do.SysMenu, error) {
	var menus []do.SysMenu
	err := d.db.WithContext(ctx).Raw(`
		select distinct m.menu_id, m.parent_id, m.menu_name, m.path, m.component, m.query, m.route_name,
		       m.visible, m.status, ifnull(m.perms,'') as perms, m.is_frame, m.is_cache, m.menu_type, m.icon, m.order_num, m.create_time
		from sys_menu m where m.menu_type in ('M', 'C') and m.status = 0
		order by m.parent_id, m.order_num`).Scan(&menus).Error
	return menus, err
}

// MenuTreeByUser 非 admin 用户菜单（角色正常状态，SQL 对位 selectMenuTreeByUserId）。
func (d *LoginDao) MenuTreeByUser(ctx context.Context, userID int64) ([]do.SysMenu, error) {
	var menus []do.SysMenu
	err := d.db.WithContext(ctx).Raw(`
		select distinct m.menu_id, m.parent_id, m.menu_name, m.path, m.component, m.query, m.route_name,
		       m.visible, m.status, ifnull(m.perms,'') as perms, m.is_frame, m.is_cache, m.menu_type, m.icon, m.order_num, m.create_time
		from sys_menu m
			 left join sys_role_menu rm on m.menu_id = rm.menu_id
			 left join sys_user_role ur on rm.role_id = ur.role_id
			 left join sys_role ro on ur.role_id = ro.role_id
			 left join sys_user u on ur.user_id = u.user_id
		where u.user_id = ? and m.menu_type in ('M', 'C') and m.status = 0 AND ro.status = 0
		order by m.parent_id, m.order_num`, userID).Scan(&menus).Error
	return menus, err
}

// RoleKeysByUser 角色权限字串集合（对位 selectRolePermissionByUserId 的 role_key 列）。
func (d *LoginDao) RoleKeysByUser(ctx context.Context, userID int64) ([]string, error) {
	var keys []string
	err := d.db.WithContext(ctx).Raw(`
		select distinct r.role_key from sys_role r
		left join sys_user_role ur on ur.role_id = r.role_id
		left join sys_user u on u.user_id = ur.user_id
		WHERE r.del_flag = '0' and ur.user_id = ?`, userID).Scan(&keys).Error
	return keys, err
}

// MenuPermsByUser 菜单权限集合（对位 selectMenuPermsByUserId）。
func (d *LoginDao) MenuPermsByUser(ctx context.Context, userID int64) ([]string, error) {
	var perms []string
	err := d.db.WithContext(ctx).Raw(`
		select distinct m.perms from sys_menu m
			 left join sys_role_menu rm on m.menu_id = rm.menu_id
			 left join sys_user_role ur on rm.role_id = ur.role_id
			 left join sys_role r on r.role_id = ur.role_id
		where m.status = '0' and r.status = '0' and ur.user_id = ?`, userID).Scan(&perms).Error
	return perms, err
}

// MenuPermsByRole 按角色查菜单权限（对位 selectMenuPermsByRoleId；getMenuPermission 填充会话角色 permissions 用）。
func (d *LoginDao) MenuPermsByRole(ctx context.Context, roleID int64) ([]string, error) {
	var perms []string
	err := d.db.WithContext(ctx).Raw(`
		select distinct m.perms from sys_menu m
			 left join sys_role_menu rm on m.menu_id = rm.menu_id
		where m.status = '0' and rm.role_id = ?`, roleID).Scan(&perms).Error
	return perms, err
}

// InsertLogininfor 写登录日志（对位 SysLogininforMapper.insertLogininfor）。
func (d *LoginDao) InsertLogininfor(ctx context.Context, m *do.SysLogininfor) error {
	return d.db.WithContext(ctx).Create(m).Error
}

func deref(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
