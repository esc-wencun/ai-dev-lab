"""
角色管理控制器（spec-05，对应java版SysRoleController）
权限标识：system:role:*
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysRole, SysUserRole, SysUser, SysDept, SysRoleDept
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from utils.page_util import paginate, get_page_domain
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, ExcelColumn
from utils.log_util import logger

roleController = APIRouter()

ADMIN_ROLE_ID = 1


async def _get_role(db, role_id):
    row = (await db.execute(select(SysRole).where(SysRole.role_id == role_id))).first()
    return row[0] if row else None


async def _check_name_unique(db, role_name, role_id=None):
    query = select(SysRole.role_id).where(SysRole.role_name == role_name)
    if role_id:
        query = query.where(SysRole.role_id != role_id)
    return (await db.execute(query.limit(1))).first() is not None


async def _check_key_unique(db, role_key, role_id=None):
    query = select(SysRole.role_id).where(SysRole.role_key == role_key)
    if role_id:
        query = query.where(SysRole.role_id != role_id)
    return (await db.execute(query.limit(1))).first() is not None


async def _refresh_role_sessions(request: Request, role_id: int):
    """
    角色权限变更后刷新在线用户会话权限（对应java refreshPermissionByRoleId）
    """
    from module_admin.service.login_service import get_redis_cache
    from module_admin.dao.login_dao import get_user_perms_by_role_id
    from common.constant import CacheConstants, Constants
    from config.get_db import AsyncSessionLocal
    cache = get_redis_cache(request)
    for key in await cache.keys_by_prefix(CacheConstants.LOGIN_TOKEN_KEY):
        data = await cache.get_cache_object(key)
        if not isinstance(data, dict):
            continue
        if data.get('user_id') == 1:
            continue
        # 判断该用户是否拥有此角色
        async with AsyncSessionLocal() as db:
            has_role = (await db.execute(
                select(SysUserRole.user_id).where(
                    SysUserRole.user_id == data.get('user_id'),
                    SysUserRole.role_id == role_id)
            )).first()
        if not has_role:
            continue
        # 重算权限
        async with AsyncSessionLocal() as db:
            perms = await get_user_perms_by_role_id(db, role_id)
        old = set(data.get('permissions') or [])
        data['permissions'] = list((old - {p for p in old}) | {p for p in perms if p})
        await cache.save_login_user(data.get('token'), data, 30)


@roleController.get('/system/role/list', dependencies=[Depends(require_perm('system:role:list'))])
async def list_role(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysRole).where(SysRole.del_flag == '0')
        if params.get('roleName'):
            query = query.where(SysRole.role_name.like(f"%{params['roleName']}%"))
        if params.get('roleKey'):
            query = query.where(SysRole.role_key.like(f"%{params['roleKey']}%"))
        if params.get('status'):
            query = query.where(SysRole.status == params['status'])
        if params.get('beginTime'):
            query = query.where(SysRole.create_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysRole.create_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query, request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.post('/system/role/export', dependencies=[Depends(require_perm('system:role:export'))])
@log_decorator(title='角色管理', business_type=BusinessType.EXPORT)
async def export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        roles = (await query_db.execute(
            select(SysRole).where(SysRole.del_flag == '0').order_by(SysRole.role_sort)
        )).scalars().all()
        columns = [
            ExcelColumn('角色编号', 'roleId', width=10),
            ExcelColumn('角色名称', 'roleName'),
            ExcelColumn('权限字符', 'roleKey'),
            ExcelColumn('显示顺序', 'roleSort', width=10),
        ]
        return export_excel('角色数据', columns, transform_result(roles), '角色数据')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.get('/system/role/optionselect',
                    dependencies=[Depends(require_perm('system:role:query'))])
async def optionselect(query_db: AsyncSession = Depends(get_db)):
    try:
        roles = (await query_db.execute(
            select(SysRole).where(SysRole.del_flag == '0').order_by(SysRole.role_sort)
        )).scalars().all()
        return ResponseUtil.success(data=transform_result(roles))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.get('/system/role/{role_id}', dependencies=[Depends(require_perm('system:role:query'))])
async def get_info(request: Request, role_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        role = await _get_role(query_db, role_id)
        return ResponseUtil.success(dict_content={'data': transform_result(role) if role else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.post('/system/role', dependencies=[Depends(require_perm('system:role:add'))])
@log_decorator(title='角色管理', business_type=BusinessType.INSERT)
async def add(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        role_name = body.get('roleName', '')
        if await _check_name_unique(query_db, role_name):
            return ResponseUtil.failure(msg=f"新增角色'{role_name}'失败，角色名称已存在")
        if await _check_key_unique(query_db, body.get('roleKey', '')):
            return ResponseUtil.failure(msg=f"新增角色'{role_name}'失败，角色权限已存在")
        role = SysRole(
            role_name=role_name,
            role_key=body.get('roleKey', ''),
            role_sort=body.get('roleSort', 0),
            data_scope=body.get('dataScope', '1'),
            menu_check_strictly=1 if body.get('menuCheckStrictly', True) else 0,
            dept_check_strictly=1 if body.get('deptCheckStrictly', True) else 0,
            status=body.get('status', '0'),
            remark=body.get('remark'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(role)
        await query_db.flush()
        for mid in body.get('menuIds') or []:
            query_db.add(SysRoleMenu(role_id=role.role_id, menu_id=mid))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


from module_admin.entity.do.entity import SysRoleMenu  # noqa: E402


@roleController.put('/system/role', dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.UPDATE)
async def edit(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        role_id = body.get('roleId')
        if role_id == ADMIN_ROLE_ID:
            return ResponseUtil.failure(msg='不允许操作超级管理员角色')
        role = await _get_role(query_db, role_id)
        if not role:
            return ResponseUtil.failure(msg='角色不存在')
        if await _check_name_unique(query_db, body.get('roleName', ''), role_id):
            return ResponseUtil.failure(msg=f"修改角色'{body.get('roleName')}'失败，角色名称已存在")
        if await _check_key_unique(query_db, body.get('roleKey', ''), role_id):
            return ResponseUtil.failure(msg=f"修改角色'{body.get('roleName')}'失败，角色权限已存在")

        role.role_name = body.get('roleName', role.role_name)
        role.role_key = body.get('roleKey', role.role_key)
        role.role_sort = body.get('roleSort', role.role_sort)
        role.status = body.get('status', role.status)
        role.remark = body.get('remark', role.remark)
        role.update_by = login_user.get('user_name', '')
        role.update_time = datetime.now()

        menu_ids = body.get('menuIds')
        if menu_ids is not None:
            await query_db.execute(
                SysRoleMenu.__table__.delete().where(SysRoleMenu.role_id == role_id))
            for mid in menu_ids:
                query_db.add(SysRoleMenu(role_id=role_id, menu_id=mid))
        await query_db.commit()
        await _refresh_role_sessions(request, role_id)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.put('/system/role/dataScope', dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.UPDATE)
async def data_scope(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        role_id = body.get('roleId')
        role = await _get_role(query_db, role_id)
        if not role:
            return ResponseUtil.failure(msg='角色不存在')
        role.data_scope = body.get('dataScope', role.data_scope)
        if role_id != ADMIN_ROLE_ID:
            await query_db.execute(
                SysRoleDept.__table__.delete().where(SysRoleDept.role_id == role_id))
            for did in body.get('deptIds') or []:
                query_db.add(SysRoleDept(role_id=role_id, dept_id=did))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.put('/system/role/changeStatus', dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.UPDATE)
async def change_status(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        role_id = body.get('roleId')
        if role_id == ADMIN_ROLE_ID:
            return ResponseUtil.failure(msg='不允许操作超级管理员角色')
        role = await _get_role(query_db, role_id)
        if not role:
            return ResponseUtil.failure(msg='角色不存在')
        role.status = body.get('status', role.status)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.delete('/system/role/{role_ids}', dependencies=[Depends(require_perm('system:role:remove'))])
@log_decorator(title='角色管理', business_type=BusinessType.DELETE)
async def remove(request: Request, role_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        for rid in [int(x) for x in role_ids.split(',') if x]:
            if rid == ADMIN_ROLE_ID:
                return ResponseUtil.failure(msg='不允许操作超级管理员角色')
            role = await _get_role(query_db, rid)
            if not role:
                continue
            role.del_flag = '2'  # 逻辑删除
            await query_db.execute(
                SysRoleMenu.__table__.delete().where(SysRoleMenu.role_id == rid))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ---------- 分配用户 ----------

@roleController.get('/system/role/authUser/allocatedList',
                    dependencies=[Depends(require_perm('system:role:list'))])
async def allocated_list(request: Request, role_id: int = 0,
                         query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysUser).outerjoin(
            SysUserRole, SysUser.user_id == SysUserRole.user_id
        ).where(
            SysUser.del_flag == '0',
            SysUserRole.role_id == role_id
        )
        if params.get('userName'):
            query = query.where(SysUser.user_name.like(f"%{params['userName']}%"))
        if params.get('phonenumber'):
            query = query.where(SysUser.phonenumber.like(f"%{params['phonenumber']}%"))
        return await paginate(query_db, query, request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.get('/system/role/authUser/unallocatedList',
                    dependencies=[Depends(require_perm('system:role:list'))])
async def unallocated_list(request: Request, role_id: int = 0,
                           query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        sub = select(SysUserRole.user_id).where(SysUserRole.role_id == role_id)
        query = select(SysUser).where(
            SysUser.del_flag == '0',
            ~SysUser.user_id.in_(sub)
        )
        if params.get('userName'):
            query = query.where(SysUser.user_name.like(f"%{params['userName']}%"))
        if params.get('phonenumber'):
            query = query.where(SysUser.phonenumber.like(f"%{params['phonenumber']}%"))
        return await paginate(query_db, query, request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.put('/system/role/authUser/cancel',
                    dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.GRANT)
async def cancel_auth_user(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        await query_db.execute(SysUserRole.__table__.delete().where(
            SysUserRole.user_id == body.get('userId'),
            SysUserRole.role_id == body.get('roleId')))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.put('/system/role/authUser/cancelAll',
                    dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.GRANT)
async def cancel_auth_all(request: Request, roleId: int = None, userIds: str = None,
                          query_db: AsyncSession = Depends(get_db)):
    try:
        if not roleId or not userIds:
            return ResponseUtil.failure(msg='参数错误')
        for uid in [int(x) for x in userIds.split(',') if x.strip()]:
            await query_db.execute(SysUserRole.__table__.delete().where(
                SysUserRole.user_id == uid, SysUserRole.role_id == roleId))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.put('/system/role/authUser/selectAll',
                    dependencies=[Depends(require_perm('system:role:edit'))])
@log_decorator(title='角色管理', business_type=BusinessType.GRANT)
async def select_auth_all(request: Request, roleId: int = None, userIds: str = None,
                          query_db: AsyncSession = Depends(get_db)):
    try:
        if not roleId or not userIds:
            return ResponseUtil.failure(msg='参数错误')
        for uid in [int(x) for x in userIds.split(',') if x.strip()]:
            exists = (await query_db.execute(
                select(SysUserRole).where(SysUserRole.user_id == uid,
                                          SysUserRole.role_id == roleId))).first()
            if not exists:
                query_db.add(SysUserRole(user_id=uid, role_id=roleId))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@roleController.get('/system/role/deptTree/{role_id}',
                    dependencies=[Depends(require_perm('system:role:query'))])
async def dept_tree(request: Request, role_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        depts = (await query_db.execute(
            select(SysDept).where(SysDept.del_flag == '0')
            .order_by(SysDept.parent_id, SysDept.order_num)
        )).scalars().all()
        checked = list((await query_db.execute(
            select(SysRoleDept.dept_id).where(SysRoleDept.role_id == role_id)
        )).scalars().all())
        nodes = {d.dept_id: {'id': d.dept_id, 'label': d.dept_name, 'children': []} for d in depts}
        tree = []
        for d in depts:
            node = nodes[d.dept_id]
            parent = nodes.get(d.parent_id)
            if parent is not None and d.parent_id != d.dept_id:
                parent['children'].append(node)
            else:
                tree.append(node)
        return ResponseUtil.success(dict_content={
            'checkedKeys': checked, 'depts': tree})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
