from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.entity.do.entity import SysUser, SysDept, SysRole, SysUserRole, SysMenu, SysRoleMenu, SysConfig


async def get_user_by_user_name(db: AsyncSession, user_name: str):
    """
    根据用户名查询用户信息（对应java版UserDetailsServiceImpl.loadUserByUsername）
    :param db: orm对象
    :param user_name: 用户名
    :return: 用户对象
    """
    user = (await db.execute(
        select(SysUser)
            .where(SysUser.user_name == user_name, SysUser.del_flag == '0')
            .distinct()
    )).first()

    return user[0] if user else None


async def get_user_by_id(db: AsyncSession, user_id: int):
    """
    根据用户id查询用户信息
    :param db: orm对象
    :param user_id: 用户id
    :return: 用户对象
    """
    user = (await db.execute(
        select(SysUser)
            .where(SysUser.user_id == user_id, SysUser.del_flag == '0')
            .distinct()
    )).first()

    return user[0] if user else None


async def get_user_roles(db: AsyncSession, user_id: int):
    """
    根据用户id查询用户角色
    :param db: orm对象
    :param user_id: 用户id
    :return: 角色列表
    """
    roles = (await db.execute(
        select(SysRole)
            .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
            .where(
                SysUserRole.user_id == user_id,
                SysRole.del_flag == '0',
                SysRole.status == '0'
            )
            .distinct()
    )).scalars().all()

    return roles


async def get_user_role_keys(db: AsyncSession, user_id: int):
    """
    根据用户id查询角色权限字符串（对应java版selectRolePermissionByUserId）
    """
    role_keys = (await db.execute(
        select(SysRole.role_key)
            .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
            .where(
                SysUserRole.user_id == user_id,
                SysRole.del_flag == '0',
                SysRole.status == '0'
            )
            .distinct()
    )).scalars().all()

    return role_keys


async def get_user_perms_by_user_id(db: AsyncSession, user_id: int):
    """
    根据用户id查询菜单权限标识（对应java版selectMenuPermsByUserId）
    """
    perms = (await db.execute(
        select(SysMenu.perms)
            .join(SysRoleMenu, SysMenu.menu_id == SysRoleMenu.menu_id)
            .join(SysUserRole, SysRoleMenu.role_id == SysUserRole.role_id)
            .join(SysRole, SysUserRole.role_id == SysRole.role_id)
            .where(
                SysUserRole.user_id == user_id,
                SysMenu.status == '0',
                SysRole.status == '0'
            )
            .distinct()
    )).scalars().all()

    return perms


async def get_user_perms_by_role_id(db: AsyncSession, role_id: int):
    """
    根据角色id查询菜单权限标识（对应java版selectMenuPermsByRoleId）
    """
    perms = (await db.execute(
        select(SysMenu.perms)
            .join(SysRoleMenu, SysMenu.menu_id == SysRoleMenu.menu_id)
            .where(SysRoleMenu.role_id == role_id, SysMenu.status == '0')
            .distinct()
    )).scalars().all()

    return perms


async def get_menu_tree_by_user_id(db: AsyncSession, user_id: int):
    """
    根据用户id查询菜单树列表（对应java版selectMenuTreeByUserId）
    """
    menus = (await db.execute(
        select(SysMenu)
            .join(SysRoleMenu, SysMenu.menu_id == SysRoleMenu.menu_id)
            .join(SysUserRole, SysRoleMenu.role_id == SysUserRole.role_id)
            .join(SysRole, SysUserRole.role_id == SysRole.role_id)
            .where(
                SysUserRole.user_id == user_id,
                SysMenu.menu_type.in_(['M', 'C']),
                SysMenu.status == '0',
                SysRole.status == '0'
            )
            .order_by(SysMenu.parent_id, SysMenu.order_num)
            .distinct()
    )).scalars().all()

    return menus


async def get_menu_tree_all(db: AsyncSession):
    """
    查询所有菜单树列表（管理员，对应java版selectMenuTreeAll）
    """
    menus = (await db.execute(
        select(SysMenu)
            .where(SysMenu.menu_type.in_(['M', 'C']), SysMenu.status == '0')
            .order_by(SysMenu.parent_id, SysMenu.order_num)
            .distinct()
    )).scalars().all()

    return menus


async def get_dept_by_id(db: AsyncSession, dept_id: int):
    """
    根据部门id查询部门信息
    """
    dept = (await db.execute(
        select(SysDept)
            .where(SysDept.dept_id == dept_id, SysDept.del_flag == '0')
    )).first()

    return dept[0] if dept else None


async def get_config_by_key(db: AsyncSession, config_key: str):
    """
    根据键名查询参数配置
    """
    config = (await db.execute(
        select(SysConfig)
            .where(SysConfig.config_key == config_key)
    )).first()

    return config[0] if config else None


async def get_config_all(db: AsyncSession):
    """
    查询所有参数配置（应用启动时缓存到redis）
    """
    configs = (await db.execute(
        select(SysConfig)
    )).scalars().all()

    return configs
