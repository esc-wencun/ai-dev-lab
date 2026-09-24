"""
菜单管理控制器（spec-05，对应java版SysMenuController）
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysMenu, SysRoleMenu, SysRole
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from common.constant import UserConstants
from module_admin.service.login_service import build_menus, get_child_perms, ADMIN_USER_ID
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.log_util import logger

menuController = APIRouter()


async def _get_menu(db, menu_id):
    row = (await db.execute(select(SysMenu).where(SysMenu.menu_id == menu_id))).first()
    return row[0] if row else None


async def _check_menu_name_unique(db, menu_name: str, parent_id: int, menu_id=None) -> bool:
    """
    同级菜单名称唯一（对齐java checkMenuNameUnique）
    :return: True=已存在（不唯一）
    """
    query = select(SysMenu.menu_id).where(
        SysMenu.menu_name == menu_name, SysMenu.parent_id == parent_id)
    if menu_id:
        query = query.where(SysMenu.menu_id != menu_id)
    return (await db.execute(query.limit(1))).first() is not None


async def _get_all_menus(db, menu_name='', status=''):
    query = select(SysMenu)
    if menu_name:
        query = query.where(SysMenu.menu_name.like(f'%{menu_name}%'))
    if status:
        query = query.where(SysMenu.status == status)
    return (await db.execute(query.order_by(SysMenu.parent_id, SysMenu.order_num))).scalars().all()


def _has_child_by_id(db_menus, menu_id):
    return any(m.parent_id == menu_id for m in db_menus)


@menuController.get('/system/menu/list', dependencies=[Depends(require_perm('system:menu:list'))])
async def list_menu(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        menus = await _get_all_menus(query_db, params.get('menuName', ''), params.get('status', ''))
        return ResponseUtil.success(data=transform_result(menus))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.get('/system/menu/treeselect')
async def treeselect(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        menus = await _get_all_menus(query_db)

        def build(pid):
            return [{'id': m.menu_id, 'label': m.menu_name,
                     'children': build(m.menu_id)} for m in menus if m.parent_id == pid]
        return ResponseUtil.success(data=build(0))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.get('/system/menu/roleMenuTreeselect/{role_id}')
async def role_menu_tree(request: Request, role_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        menus = await _get_all_menus(query_db)

        def build(pid):
            return [{'id': m.menu_id, 'label': m.menu_name,
                     'children': build(m.menu_id)} for m in menus if m.parent_id == pid]

        # 选中keys（menuCheckStrictly=1时剔除有子节点被选中的父节点，对齐java selectMenuListByRoleId）
        checked = list((await query_db.execute(
            select(SysRoleMenu.menu_id).where(SysRoleMenu.role_id == role_id)
        )).scalars().all())
        role = (await query_db.execute(
            select(SysRole).where(SysRole.role_id == role_id))).first()
        strictly = bool(role[0].menu_check_strictly) if role else False
        if strictly:
            menu_ids = {m.menu_id for m in menus}
            checked = [mid for mid in checked
                       if mid not in menu_ids or
                       not any(m.parent_id == mid for m in menus if m.menu_id in checked)]
        return ResponseUtil.success(dict_content={'checkedKeys': checked, 'menus': build(0)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.get('/system/menu/{menu_id}', dependencies=[Depends(require_perm('system:menu:query'))])
async def get_info(request: Request, menu_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        menu = await _get_menu(query_db, menu_id)
        return ResponseUtil.success(dict_content={'data': transform_result(menu) if menu else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.post('/system/menu', dependencies=[Depends(require_perm('system:menu:add'))])
@log_decorator(title='菜单管理', business_type=BusinessType.INSERT)
async def add(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        # 同级菜单名称唯一（对齐java checkMenuNameUnique，add/edit都会走）
        if await _check_menu_name_unique(query_db, body.get('menuName', ''),
                                          body.get('parentId', 0)):
            return ResponseUtil.failure(msg=f"新增菜单'{body.get('menuName')}'失败，菜单名称已存在")
        menu = SysMenu(
            menu_name=body.get('menuName', ''),
            parent_id=body.get('parentId', 0),
            order_num=body.get('orderNum', 0),
            path=body.get('path', ''),
            component=body.get('component'),
            query=body.get('query'),
            route_name=body.get('routeName', ''),
            is_frame=body.get('isFrame', 1),
            is_cache=body.get('isCache', 0),
            menu_type=body.get('menuType', ''),
            visible=body.get('visible', '0'),
            status=body.get('status', '0'),
            perms=body.get('perms'),
            icon=body.get('icon', '#'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(menu)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.put('/system/menu', dependencies=[Depends(require_perm('system:menu:edit'))])
@log_decorator(title='菜单管理', business_type=BusinessType.UPDATE)
async def edit(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        menu_id = body.get('menuId')
        menu = await _get_menu(query_db, menu_id)
        if not menu:
            return ResponseUtil.failure(msg='菜单不存在')
        parent_id = body.get('parentId', menu.parent_id)
        if parent_id == menu_id:
            return ResponseUtil.failure(msg=f"修改菜单{body.get('menuName')}失败，上级菜单不能选择自己")
        # 同级菜单名称唯一（对齐java checkMenuNameUnique）
        if await _check_menu_name_unique(db=query_db, menu_name=body.get('menuName', menu.menu_name),
                                          parent_id=parent_id, menu_id=menu_id):
            return ResponseUtil.failure(msg=f"修改菜单'{body.get('menuName')}'失败，菜单名称已存在")
        menu.menu_name = body.get('menuName', menu.menu_name)
        menu.parent_id = parent_id
        menu.order_num = body.get('orderNum', menu.order_num)
        menu.path = body.get('path', menu.path)
        menu.component = body.get('component', menu.component)
        menu.query = body.get('query', menu.query)
        menu.route_name = body.get('routeName', menu.route_name)
        menu.is_frame = body.get('isFrame', menu.is_frame)
        menu.is_cache = body.get('isCache', menu.is_cache)
        menu.menu_type = body.get('menuType', menu.menu_type)
        menu.visible = body.get('visible', menu.visible)
        menu.status = body.get('status', menu.status)
        menu.perms = body.get('perms', menu.perms)
        menu.icon = body.get('icon', menu.icon)
        menu.update_by = login_user.get('user_name', '')
        menu.update_time = datetime.now()
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.put('/system/menu/updateSort', dependencies=[Depends(require_perm('system:menu:edit'))])
@log_decorator(title='菜单管理', business_type=BusinessType.UPDATE)
async def update_sort(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        menu_ids = (body.get('menuIds') or '').split(',')
        order_nums = (body.get('orderNums') or '').split(',')
        for mid, onum in zip(menu_ids, order_nums):
            menu = await _get_menu(query_db, int(mid))
            if menu:
                menu.order_num = int(onum)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@menuController.delete('/system/menu/{menu_id}', dependencies=[Depends(require_perm('system:menu:remove'))])
@log_decorator(title='菜单管理', business_type=BusinessType.DELETE)
async def remove(request: Request, menu_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        if _has_child_by_id(await _get_all_menus(query_db), menu_id):
            return ResponseUtil.failure(msg='存在子菜单,不允许删除')
        assigned = (await query_db.execute(
            select(SysRoleMenu.menu_id).where(SysRoleMenu.menu_id == menu_id).limit(1))).first()
        if assigned:
            return ResponseUtil.failure(msg='菜单已分配,不允许删除')
        menu = await _get_menu(query_db, menu_id)
        if not menu:
            return ResponseUtil.failure(msg='菜单不存在')
        await query_db.delete(menu)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
