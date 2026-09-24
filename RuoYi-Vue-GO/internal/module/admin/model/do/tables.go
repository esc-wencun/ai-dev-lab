package do

import "ruoyi-vue-go/pkg/types"

// SysUser 用户信息表（对位 Java SysUser / sys_user 表）。
type SysUser struct {
	UserID        int64          `gorm:"column:user_id;primaryKey" json:"userId"`
	DeptID        int64          `gorm:"column:dept_id" json:"deptId"`
	UserName      string         `gorm:"column:user_name" json:"userName"`
	NickName      string         `gorm:"column:nick_name" json:"nickName"`
	Email         string         `gorm:"column:email" json:"email"`
	Avatar        string         `gorm:"column:avatar" json:"avatar"`
	Phonenumber   string         `gorm:"column:phonenumber" json:"phonenumber"`
	Password      string         `gorm:"column:password" json:"-"` // 任何响应不得带出（Java 用 @JsonIgnore）
	Sex           string         `gorm:"column:sex" json:"sex"`
	Status        string         `gorm:"column:status" json:"status"`
	DelFlag       string         `gorm:"column:del_flag" json:"delFlag"`
	LoginIP       string         `gorm:"column:login_ip" json:"loginIp"`
	LoginDate     types.DateTime `gorm:"column:login_date" json:"loginDate"`
	PwdUpdateDate types.DateTime `gorm:"column:pwd_update_date" json:"pwdUpdateDate"`
	CreateBy      string         `gorm:"column:create_by" json:"createBy"`
	CreateTime    types.DateTime `gorm:"column:create_time" json:"createTime"`
	UpdateBy      string         `gorm:"column:update_by" json:"updateBy"`
	UpdateTime    types.DateTime `gorm:"column:update_time" json:"updateTime"`
	Remark        string         `gorm:"column:remark" json:"remark"`
}

func (SysUser) TableName() string { return "sys_user" }

// SysDept 部门表。
type SysDept struct {
	DeptID    int64  `gorm:"column:dept_id;primaryKey" json:"deptId"`
	ParentID  int64  `gorm:"column:parent_id" json:"parentId"`
	Ancestors string `gorm:"column:ancestors" json:"ancestors"`
	DeptName  string `gorm:"column:dept_name" json:"deptName"`
	OrderNum  int    `gorm:"column:order_num" json:"orderNum"`
	Leader    string `gorm:"column:leader" json:"leader"`
	Status    string `gorm:"column:status" json:"status"`
	DelFlag   string `gorm:"column:del_flag" json:"delFlag"`
}

func (SysDept) TableName() string { return "sys_dept" }

// SysRole 角色表（数据权限/角色状态承载）。
type SysRole struct {
	RoleID    int64  `gorm:"column:role_id;primaryKey" json:"roleId"`
	RoleName  string `gorm:"column:role_name" json:"roleName"`
	RoleKey   string `gorm:"column:role_key" json:"roleKey"`
	RoleSort  int    `gorm:"column:role_sort" json:"roleSort"`
	DataScope string `gorm:"column:data_scope" json:"dataScope"`
	Status    string `gorm:"column:status" json:"status"`
	DelFlag   string `gorm:"column:del_flag" json:"delFlag"`
}

func (SysRole) TableName() string { return "sys_role" }

// SysUserRole 用户-角色关联。
type SysUserRole struct {
	UserID int64 `gorm:"column:user_id;primaryKey"`
	RoleID int64 `gorm:"column:role_id;primaryKey"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

// SysRoleMenu 角色-菜单关联。
type SysRoleMenu struct {
	RoleID int64 `gorm:"column:role_id;primaryKey"`
	MenuID int64 `gorm:"column:menu_id;primaryKey"`
}

func (SysRoleMenu) TableName() string { return "sys_role_menu" }

// SysMenu 菜单权限表（对位 SysMenu）。
type SysMenu struct {
	MenuID     int64          `gorm:"column:menu_id;primaryKey" json:"menuId"`
	MenuName   string         `gorm:"column:menu_name" json:"menuName"`
	ParentID   int64          `gorm:"column:parent_id" json:"parentId"`
	OrderNum   int            `gorm:"column:order_num" json:"orderNum"`
	Path       string         `gorm:"column:path" json:"path"`
	Component  string         `gorm:"column:component" json:"component"`
	Query      string         `gorm:"column:query" json:"query"`
	RouteName  string         `gorm:"column:route_name" json:"routeName"`
	IsFrame    string         `gorm:"column:is_frame" json:"isFrame"`
	IsCache    string         `gorm:"column:is_cache" json:"isCache"`
	MenuType   string         `gorm:"column:menu_type" json:"menuType"`
	Visible    string         `gorm:"column:visible" json:"visible"`
	Status     string         `gorm:"column:status" json:"status"`
	Perms      string         `gorm:"column:perms" json:"perms"`
	Icon       string         `gorm:"column:icon" json:"icon"`
	CreateTime types.DateTime `gorm:"column:create_time" json:"createTime"`
	Children   []*SysMenu     `gorm:"-" json:"children,omitempty"`
}

func (SysMenu) TableName() string { return "sys_menu" }

// SysConfig 参数配置表（selectConfigByKey 回源查询用）。
type SysConfig struct {
	ConfigID    int64          `gorm:"column:config_id;primaryKey" json:"configId"`
	ConfigName  string         `gorm:"column:config_name" json:"configName"`
	ConfigKey   string         `gorm:"column:config_key" json:"configKey"`
	ConfigValue string         `gorm:"column:config_value" json:"configValue"`
	ConfigType  string         `gorm:"column:config_type" json:"configType"`
	CreateTime  types.DateTime `gorm:"column:create_time" json:"createTime"`
	UpdateBy    string         `gorm:"column:update_by" json:"updateBy"`
	UpdateTime  types.DateTime `gorm:"column:update_time" json:"updateTime"`
	Remark      string         `gorm:"column:remark" json:"remark"`
}

func (SysConfig) TableName() string { return "sys_config" }

// SysLogininfor 系统访问记录（对位 sys_logininfor 表）。
type SysLogininfor struct {
	InfoID        int64          `gorm:"column:info_id;primaryKey;autoIncrement" json:"infoId"`
	UserName      string         `gorm:"column:user_name" json:"userName"`
	IPAddr        string         `gorm:"column:ipaddr" json:"ipaddr"`
	LoginLocation string         `gorm:"column:login_location" json:"loginLocation"`
	Browser       string         `gorm:"column:browser" json:"browser"`
	OS            string         `gorm:"column:os" json:"os"`
	Status        string         `gorm:"column:status" json:"status"`
	Msg           string         `gorm:"column:msg" json:"msg"`
	LoginTime     types.DateTime `gorm:"column:login_time" json:"loginTime"`
}

func (SysLogininfor) TableName() string { return "sys_logininfor" }
