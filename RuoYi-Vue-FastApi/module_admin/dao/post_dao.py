"""
岗位管理 DAO（spec-03，对应java版SysPostMapper.xml）
"""
from typing import List, Optional
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.entity.do.entity import SysPost, SysUserPost


async def get_post_by_id(db: AsyncSession, post_id: int) -> Optional[SysPost]:
    row = (await db.execute(select(SysPost).where(SysPost.post_id == post_id))).first()
    return row[0] if row else None


async def get_post_list(db: AsyncSession, post_code: str = '', post_name: str = '',
                         status: str = '') -> List[SysPost]:
    query = select(SysPost)
    if post_code:
        query = query.where(SysPost.post_code.like(f'%{post_code}%'))
    if post_name:
        query = query.where(SysPost.post_name.like(f'%{post_name}%'))
    if status:
        query = query.where(SysPost.status == status)
    return (await db.execute(query.order_by(SysPost.post_sort))).scalars().all()


async def check_post_name_unique(db: AsyncSession, post_name: str,
                                  post_id: Optional[int] = None) -> bool:
    query = select(SysPost.post_id).where(SysPost.post_name == post_name)
    if post_id:
        query = query.where(SysPost.post_id != post_id)
    return (await db.execute(query.limit(1))).first() is not None


async def check_post_code_unique(db: AsyncSession, post_code: str,
                                  post_id: Optional[int] = None) -> bool:
    query = select(SysPost.post_id).where(SysPost.post_code == post_code)
    if post_id:
        query = query.where(SysPost.post_id != post_id)
    return (await db.execute(query.limit(1))).first() is not None


async def count_post_users(db: AsyncSession, post_id: int) -> int:
    """
    岗位被用户引用数（java版countPostUsers）
    """
    count = (await db.execute(
        select(func.count('*')).select_from(SysUserPost).where(SysUserPost.post_id == post_id)
    )).scalar()
    return count or 0


async def insert_post(db: AsyncSession, post: SysPost) -> SysPost:
    db.add(post)
    await db.flush()
    return post


async def delete_post(db: AsyncSession, post: SysPost):
    await db.delete(post)
