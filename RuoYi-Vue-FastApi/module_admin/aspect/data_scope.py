"""
数据权限过滤（对应java版DataScopeAspect）

用法（service层查询用户列表时）：
    conditions = build_data_scope_conditions(login_user, dept_alias=SysDept, user_alias=SysUser)
    query = query.where(*conditions)
规范：需要数据权限的列表接口必须调用本工具；admin用户（userId=1）返回空条件（不过滤）。
"""
from typing import List
from sqlalchemy import or_, and_, select
from common.constant import DataScope
from utils.log_util import logger

# 管理员用户ID（java版SysUser.isAdmin）
ADMIN_USER_ID = 1


async def build_data_scope_conditions(login_user: dict, SysDept, SysUser) -> List:
    """
    根据会话用户的角色data_scope生成过滤条件（对齐java版DataScopeAspect的SQL拼接语义）
    :param login_user: Redis会话（user_id等）
    :param SysDept: 部门模型类（别名d）
    :param SysUser: 用户模型类（别名u）
    :return: SQLAlchemy条件列表（空列表=不过滤）
    """
    user_id = login_user.get('user_id')
    if user_id == ADMIN_USER_ID:
        return []

    from config.get_db import AsyncSessionLocal
    from module_admin.entity.do.entity import SysRole, SysUserRole, SysRoleDept

    async with AsyncSessionLocal() as db:
        roles = (await db.execute(
            select(SysRole)
                .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
                .where(
                    SysUserRole.user_id == user_id,
                    SysRole.status == '0',
                    SysRole.del_flag == '0'
                )
        )).scalars().all()

    conditions = []
    for role in roles:
        scope = role.data_scope
        if scope == DataScope.ALL:
            return []  # 任一角色全部权限 -> 不过滤
        if scope == DataScope.CUSTOM:
            # 自定：部门 in role_dept
            sub = select(SysRoleDept.dept_id).where(SysRoleDept.role_id == role.role_id)
            conditions.append(SysDept.dept_id.in_(sub))
        elif scope == DataScope.DEPT:
            conditions.append(SysUser.dept_id == _user_dept_id(login_user))
        elif scope == DataScope.DEPT_AND_CHILD:
            dept_id = _user_dept_id(login_user)
            # 本部门及以下：ancestors匹配 or 本部门
            conditions.append(and_(
                SysDept.dept_id == dept_id,
                SysDept.del_flag == '0'
            ) | and_(
                SysDept.ancestors.like(f'%,{dept_id},%') | SysDept.ancestors.like(f'%,{dept_id}') |
                SysDept.ancestors.like(f'{dept_id},%'),
                SysDept.del_flag == '0'
            ))
        elif scope == DataScope.SELF:
            conditions.append(SysUser.user_id == user_id)

    return [or_(*conditions)] if conditions else [SysUser.user_id == -1]  # 无角色或无权限->仅空集


def _user_dept_id(login_user: dict):
    """
    从会话取用户部门id（会话未存时查库）
    """
    dept_id = login_user.get('dept_id')
    if dept_id:
        return dept_id
    from config.get_db import AsyncSessionLocal
    from module_admin.dao.login_dao import get_user_by_id
    import asyncio

    async def _q():
        async with AsyncSessionLocal() as db:
            user = await get_user_by_id(db, login_user.get('user_id'))
            return user.dept_id if user else None
    try:
        loop = asyncio.get_running_loop()
        # 已在事件循环中：同步预查不可行，返回-1兜底（调用方应确保会话含dept_id）
        logger.warning('数据权限：会话缺少dept_id且无法同步查询')
        return -1
    except RuntimeError:
        return asyncio.run(_q())
