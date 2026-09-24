from sqlalchemy import Column, BigInteger, String, DateTime, Integer
from config.database import Base


class SysUser(Base):
    """
    用户信息表（对应java版ry-vue库sys_user表）
    """
    __tablename__ = 'sys_user'

    user_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='用户ID')
    dept_id = Column(BigInteger, comment='部门ID')
    user_name = Column(String(30), nullable=False, comment='用户账号')
    nick_name = Column(String(30), nullable=False, comment='用户昵称')
    user_type = Column(String(2), default='00', comment='用户类型（00系统用户）')
    email = Column(String(50), default='', comment='用户邮箱')
    phonenumber = Column(String(11), default='', comment='手机号码')
    sex = Column(String(1), default='0', comment='用户性别（0男 1女 2未知）')
    avatar = Column(String(100), default='', comment='头像地址')
    password = Column(String(100), default='', comment='密码')
    status = Column(String(1), default='0', comment='账号状态（0正常 1停用）')
    del_flag = Column(String(1), default='0', comment='删除标志（0代表存在 2代表删除）')
    login_ip = Column(String(128), default='', comment='最后登录IP')
    login_date = Column(DateTime, comment='最后登录时间')
    pwd_update_date = Column(DateTime, comment='密码最后更新时间')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysDept(Base):
    """
    部门表
    """
    __tablename__ = 'sys_dept'

    dept_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='部门id')
    parent_id = Column(BigInteger, default=0, comment='父部门id')
    ancestors = Column(String(50), default='', comment='祖级列表')
    dept_name = Column(String(30), default='', comment='部门名称')
    order_num = Column(Integer, default=0, comment='显示顺序')
    leader = Column(String(20), default=None, comment='负责人')
    phone = Column(String(11), default=None, comment='联系电话')
    email = Column(String(50), default=None, comment='邮箱')
    status = Column(String(1), default='0', comment='部门状态（0正常 1停用）')
    del_flag = Column(String(1), default='0', comment='删除标志（0代表存在 2代表删除）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')


class SysPost(Base):
    """
    岗位信息表（对应java版sys_post）
    """
    __tablename__ = 'sys_post'

    post_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='岗位ID')
    post_code = Column(String(64), nullable=False, comment='岗位编码')
    post_name = Column(String(50), nullable=False, comment='岗位名称')
    post_sort = Column(Integer, nullable=False, comment='显示顺序')
    status = Column(String(1), nullable=False, comment='状态（0正常 1停用）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysRole(Base):
    """
    角色信息表
    """
    __tablename__ = 'sys_role'

    role_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='角色ID')
    role_name = Column(String(30), nullable=False, comment='角色名称')
    role_key = Column(String(100), nullable=False, comment='角色权限字符串')
    role_sort = Column(Integer, nullable=False, comment='显示顺序')
    data_scope = Column(String(1), default='1', comment='数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限 5：仅本人数据权限）')
    menu_check_strictly = Column(Integer, default=1, comment='菜单树选择项是否关联显示')
    dept_check_strictly = Column(Integer, default=1, comment='部门树选择项是否关联显示')
    status = Column(String(1), nullable=False, comment='角色状态（0正常 1停用）')
    del_flag = Column(String(1), default='0', comment='删除标志（0代表存在 2代表删除）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysMenu(Base):
    """
    菜单权限表
    """
    __tablename__ = 'sys_menu'

    menu_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='菜单ID')
    menu_name = Column(String(50), nullable=False, comment='菜单名称')
    parent_id = Column(BigInteger, default=0, comment='父菜单ID')
    order_num = Column(Integer, default=0, comment='显示顺序')
    path = Column(String(200), default='', comment='路由地址')
    component = Column(String(255), comment='组件路径')
    query = Column(String(255), comment='路由参数')
    route_name = Column(String(50), default='', comment='路由名称')
    is_frame = Column(Integer, default=1, comment='是否为外链（0是 1否）')
    is_cache = Column(Integer, default=0, comment='是否缓存（0缓存 1不缓存）')
    menu_type = Column(String(1), default='', comment='菜单类型（M目录 C菜单 F按钮）')
    visible = Column(String(1), default='0', comment='菜单状态（0显示 1隐藏）')
    status = Column(String(1), default='0', comment='菜单状态（0正常 1停用）')
    perms = Column(String(100), comment='权限标识')
    icon = Column(String(100), default='#', comment='菜单图标')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), default='', comment='备注')


class SysUserRole(Base):
    """
    用户和角色关联表
    """
    __tablename__ = 'sys_user_role'

    user_id = Column(BigInteger, primary_key=True, nullable=False, comment='用户ID')
    role_id = Column(BigInteger, primary_key=True, nullable=False, comment='角色ID')


class SysUserPost(Base):
    """
    用户与岗位关联表
    """
    __tablename__ = 'sys_user_post'

    user_id = Column(BigInteger, primary_key=True, nullable=False, comment='用户ID')
    post_id = Column(BigInteger, primary_key=True, nullable=False, comment='岗位ID')


class SysRoleMenu(Base):
    """
    角色和菜单关联表
    """
    __tablename__ = 'sys_role_menu'

    role_id = Column(BigInteger, primary_key=True, nullable=False, comment='角色ID')
    menu_id = Column(BigInteger, primary_key=True, nullable=False, comment='菜单ID')


class SysRoleDept(Base):
    """
    角色和部门关联表
    """
    __tablename__ = 'sys_role_dept'

    role_id = Column(BigInteger, primary_key=True, nullable=False, comment='角色ID')
    dept_id = Column(BigInteger, primary_key=True, nullable=False, comment='部门ID')


class SysConfig(Base):
    """
    参数配置表
    """
    __tablename__ = 'sys_config'

    config_id = Column(Integer, primary_key=True, autoincrement=True, comment='参数主键')
    config_name = Column(String(100), default='', comment='参数名称')
    config_key = Column(String(100), default='', comment='参数键名')
    config_value = Column(String(500), default='', comment='参数键值')
    config_type = Column(String(1), default='N', comment='系统内置（Y是 N否）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysNotice(Base):
    """
    通知公告表（对应java版sys_notice）
    """
    __tablename__ = 'sys_notice'

    notice_id = Column(Integer, primary_key=True, autoincrement=True, comment='公告ID')
    notice_title = Column(String(50), nullable=False, comment='公告标题')
    notice_type = Column(String(1), nullable=False, comment='公告类型（1通知 2公告）')
    notice_content = Column(String(65535), comment='公告内容（longblob，utf-8文本）')
    status = Column(String(1), default='0', comment='公告状态（0正常 1关闭）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(255), comment='备注')


class SysNoticeRead(Base):
    """
    公告已读记录表（对应java版sys_notice_read）
    """
    __tablename__ = 'sys_notice_read'

    read_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='已读主键')
    notice_id = Column(Integer, nullable=False, comment='公告id')
    user_id = Column(BigInteger, nullable=False, comment='用户id')
    read_time = Column(DateTime, nullable=False, comment='阅读时间')


class SysDictType(Base):
    """
    字典类型表（对应java版sys_dict_type）
    """
    __tablename__ = 'sys_dict_type'

    dict_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='字典主键')
    dict_name = Column(String(100), default='', comment='字典名称')
    dict_type = Column(String(100), default='', comment='字典类型')
    status = Column(String(1), default='0', comment='状态（0正常 1停用）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysDictData(Base):
    """
    字典数据表（对应java版sys_dict_data）
    """
    __tablename__ = 'sys_dict_data'

    dict_code = Column(BigInteger, primary_key=True, autoincrement=True, comment='字典编码')
    dict_sort = Column(Integer, default=0, comment='字典排序')
    dict_label = Column(String(100), default='', comment='字典标签')
    dict_value = Column(String(100), default='', comment='字典键值')
    dict_type = Column(String(100), default='', comment='字典类型')
    css_class = Column(String(100), comment='样式属性')
    list_class = Column(String(100), comment='表格回显样式')
    is_default = Column(String(1), default='N', comment='是否默认（Y是 N否）')
    status = Column(String(1), default='0', comment='状态（0正常 1停用）')
    create_by = Column(String(64), default='', comment='创建者')
    create_time = Column(DateTime, comment='创建时间')
    update_by = Column(String(64), default='', comment='更新者')
    update_time = Column(DateTime, comment='更新时间')
    remark = Column(String(500), comment='备注')


class SysOperLog(Base):
    """
    操作日志记录表（对应java版sys_oper_log）
    """
    __tablename__ = 'sys_oper_log'

    oper_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='日志主键')
    title = Column(String(50), default='', comment='模块标题')
    business_type = Column(Integer, default=0, comment='业务类型（0其它 1新增 2修改 3删除）')
    method = Column(String(200), default='', comment='方法名称')
    request_method = Column(String(10), default='', comment='请求方式')
    operator_type = Column(Integer, default=0, comment='操作类别（0其它 1后台用户 2手机端用户）')
    oper_name = Column(String(50), default='', comment='操作人员')
    dept_name = Column(String(50), default='', comment='部门名称')
    oper_url = Column(String(255), default='', comment='请求URL')
    oper_ip = Column(String(128), default='', comment='主机地址')
    oper_location = Column(String(255), default='', comment='操作地点')
    oper_param = Column(String(2000), default='', comment='请求参数')
    json_result = Column(String(2000), default='', comment='返回参数')
    status = Column(Integer, default=0, comment='操作状态（0正常 1异常）')
    error_msg = Column(String(2000), default='', comment='错误消息')
    oper_time = Column(DateTime, comment='操作时间')
    cost_time = Column(BigInteger, default=0, comment='消耗时间')


class SysLogininfor(Base):
    """
    系统访问记录表
    """
    __tablename__ = 'sys_logininfor'

    info_id = Column(BigInteger, primary_key=True, autoincrement=True, comment='访问ID')
    user_name = Column(String(50), default='', comment='用户账号')
    ipaddr = Column(String(128), default='', comment='登录IP地址')
    login_location = Column(String(255), default='', comment='登录地点')
    browser = Column(String(50), default='', comment='浏览器类型')
    os = Column(String(50), default='', comment='操作系统')
    status = Column(String(1), default='0', comment='登录状态（0成功 1失败）')
    msg = Column(String(255), default='', comment='提示消息')
    login_time = Column(DateTime, comment='访问时间')
