"""
部门管理 DAO（spec-03，对应java版SysDeptMapper.xml）
分层职责：仅查询与写库语句，无业务判断
"""
from typing import List, Optional
from sqlalchemy import select, and_, or_
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.entity.do.entity import SysDept, SysUser


async def get_dept_by_id(db: AsyncSession, dept_id: int) -> Optional[SysDept]:
    dept = (await db.execute(
        select(SysDept).where(SysDept.dept_id == dept_id, SysDept.del_flag == '0')
    )).first()
    return dept[0] if dept else None


async def get_dept_list(db: AsyncSession, dept_name: str = '', status: str = '') -> List[SysDept]:
    """
    全量部门列表（父级排序在前，供服务端组树）
    """
    query = select(SysDept).where(SysDept.del_flag == '0')
    if dept_name:
        query = query.where(SysDept.dept_name.like(f'%{dept_name}%'))
    if status:
        query = query.where(SysDept.status == status)
    return (await db.execute(
        query.order_by(SysDept.parent_id, SysDept.order_num)
    )).scalars().all()


async def get_dept_children(db: AsyncSession, dept_id: int) -> List[SysDept]:
    """
    子部门（含各层级，供exclude校验）
    """
    return (await db.execute(
        select(SysDept).where(SysDept.del_flag == '0')
    )).scalars().all()


async def has_child(db: AsyncSession, dept_id: int) -> bool:
    """
    是否存在子部门（正常状态）
    """
    found = (await db.execute(
        select(SysDept.dept_id).where(SysDept.parent_id == dept_id, SysDept.del_flag == '0').limit(1)
    )).first()
    return found is not None


async def check_dept_exist_user(db: AsyncSession, dept_id: int) -> bool:
    """
    部门是否存在用户
    """
    from sqlalchemy import func
    count = (await db.execute(
        select(func.count('*')).select_from(SysUser)
        .where(SysUser.dept_id == dept_id, SysUser.del_flag == '0')
    )).scalar()
    return (count or 0) > 0


async def check_dept_name_unique(db: AsyncSession, dept_name: str, parent_id: int,
                                  dept_id: Optional[int] = None) -> bool:
    """
    同级部门名称是否唯一（java语义：同级下同名即不唯一）
    """
    query = select(SysDept.dept_id).where(
        SysDept.dept_name == dept_name,
        SysDept.parent_id == parent_id,
        SysDept.del_flag == '0'
    )
    if dept_id:
        query = query.where(SysDept.dept_id != dept_id)
    # 存在同名=不唯一(True)；不存在=唯一(False)。函数名是check_unique，
    # 但语义返回的是"已存在"（java NOT_UNIQUE风格）。修正：返回True表示已存在（不唯一）
    found = (await db.execute(query.limit(1))).first()
    return found is not None


async def insert_dept(db: AsyncSession, dept: SysDept) -> SysDept:
    db.add(dept)
    await db.flush()
    return dept


async def update_dept(db: AsyncSession, dept: SysDept):
    await db.flush()


async def update_dept_children_ancestors(db: AsyncSession, dept_id: int,
                                          new_ancestors: str, old_ancestors: str):
    """
    级联更新子孙ancestors（对应java updateDeptChildren的replaceFirst语义）
    """
    children = (await db.execute(
        select(SysDept).where(SysDept.del_flag == '0')
    )).scalars().all()
    updated = 0
    for child in children:
        if child.ancestors and old_ancestors in child.ancestors:
            # 找到以 old_ancestors 开头链路上的子孙（含孙子）：
            # java replaceFirst(oldAncestors, newAncestors)——只替换前缀部分
            if child.ancestors == old_ancestors or child.ancestors.startswith(old_ancestors + ','):
                child.ancestors = child.ancestors.replace(old_ancestors, new_ancestors, 1)
                updated += 1
    return updated
