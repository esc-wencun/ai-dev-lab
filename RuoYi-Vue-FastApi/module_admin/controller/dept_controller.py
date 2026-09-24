"""
部门管理控制器（spec-03，对应java版SysDeptController）
权限标识：system:dept:*
"""
from fastapi import APIRouter, Request, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysDept
from module_admin.dao import dept_dao
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType, UserStatus
from common.constant import UserConstants
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.log_util import logger

deptController = APIRouter(dependencies=[Depends(require_perm('system:dept:list'))])

TYPE_DIR = 'M'


def _build_dept_tree(depts) -> list:
    """
    部门组树（children嵌套，前端el-table树形展示）
    """
    nodes = {d.dept_id: transform_result(d) for d in depts}
    tree = []
    for d in depts:
        node = nodes[d.dept_id]
        parent = nodes.get(d.parent_id)
        if parent is not None and d.parent_id != d.dept_id:
            parent.setdefault('children', []).append(node)
        else:
            tree.append(node)
    return tree


@deptController.get('/system/dept/list')
async def list_dept(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    部门列表（服务端组树，过滤deptName/status）
    """
    try:
        params = request.query_params
        depts = await dept_dao.get_dept_list(
            query_db, params.get('deptName', ''), params.get('status', ''))
        return ResponseUtil.success(msg='操作成功', dict_content={'data': _build_dept_tree(depts)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.get('/system/dept/list/exclude/{dept_id}')
async def exclude_child(request: Request, dept_id: int, query_db: AsyncSession = Depends(get_db)):
    """
    部门列表（排除自己及子孙——编辑上级选择）
    """
    try:
        depts = await dept_dao.get_dept_list(query_db)
        # 收集自己及所有子孙id（基于ancestors前缀匹配）
        target = next((d for d in depts if d.dept_id == dept_id), None)
        excluded = {dept_id}
        if target:
            for d in depts:
                if d.ancestors and (d.ancestors == target.ancestors or
                                     d.ancestors.startswith(target.ancestors + ',') or
                                     f',{dept_id},' in f',{d.ancestors},{d.dept_id},' or
                                     d.dept_id == dept_id):
                    if d.dept_id == dept_id or (d.ancestors.startswith(f'{target.ancestors},{dept_id}')):
                        excluded.add(d.dept_id)
        tree = _build_dept_tree([d for d in depts if d.dept_id not in excluded])
        return ResponseUtil.success(msg='操作成功', dict_content={'data': tree})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.get('/system/dept/{dept_id}')
async def get_info(request: Request, dept_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        dept = await dept_dao.get_dept_by_id(query_db, dept_id)
        return ResponseUtil.success(dict_content={'data': transform_result(dept) if dept else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.post('/system/dept')
@log_decorator(title='部门管理', business_type=BusinessType.INSERT)
async def add(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        parent_id = body.get('parentId', 0)
        parent = await dept_dao.get_dept_by_id(query_db, parent_id)
        if not parent:
            return ResponseUtil.failure(msg='上级部门不存在')
        if parent.status != UserConstants.NORMAL:
            return ResponseUtil.failure(msg='部门停用，不允许新增')
        if await dept_dao.check_dept_name_unique(query_db, body.get('deptName', ''), parent_id):
            return ResponseUtil.failure(msg=f"新增部门'{body.get('deptName')}'失败，部门名称已存在")
        dept = SysDept(
            parent_id=parent_id,
            ancestors=f'{parent.ancestors},{parent_id}',
            dept_name=body.get('deptName', ''),
            order_num=body.get('orderNum', 0),
            leader=body.get('leader'),
            phone=body.get('phone'),
            email=body.get('email'),
            status=body.get('status', '0'),
        )
        # 操作人从会话取
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request)
        dept.create_by = login_user.get('user_name') if login_user else ''
        await dept_dao.insert_dept(query_db, dept)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.put('/system/dept')
@log_decorator(title='部门管理', business_type=BusinessType.UPDATE)
async def edit(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        dept_id = body.get('deptId')
        dept = await dept_dao.get_dept_by_id(query_db, dept_id)
        if not dept:
            return ResponseUtil.failure(msg='部门不存在')
        new_parent_id = body.get('parentId', dept.parent_id)
        if new_parent_id == dept.dept_id:
            return ResponseUtil.failure(msg=f"修改部门{body.get('deptName')}失败，上级部门不能是自己")
        if await dept_dao.check_dept_name_unique(query_db, body.get('deptName', ''),
                                                  new_parent_id, dept_id):
            return ResponseUtil.failure(msg=f"修改部门'{body.get('deptName')}'失败，部门名称已存在")

        new_parent = await dept_dao.get_dept_by_id(query_db, new_parent_id)
        if new_parent:
            new_ancestors = f'{new_parent.ancestors},{new_parent_id}'
            old_ancestors = dept.ancestors
            if old_ancestors and new_ancestors != old_ancestors:
                # 级联更新子孙
                from module_admin.service.login_service import TokenService
                await dept_dao.update_dept_children_ancestors(query_db, dept_id,
                                                               new_ancestors, old_ancestors)
            dept.ancestors = new_ancestors

        dept.parent_id = new_parent_id
        dept.dept_name = body.get('deptName', dept.dept_name)
        dept.order_num = body.get('orderNum', dept.order_num)
        dept.leader = body.get('leader', dept.leader)
        dept.phone = body.get('phone', dept.phone)
        dept.email = body.get('email', dept.email)
        dept.status = body.get('status', dept.status)
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request)
        dept.update_by = login_user.get('user_name') if login_user else ''
        dept.update_time = __import__('datetime').datetime.now()
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.put('/system/dept/updateSort')
@log_decorator(title='部门管理', business_type=BusinessType.UPDATE)
async def update_sort(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        dept_ids = (body.get('deptIds') or '').split(',')
        order_nums = (body.get('orderNums') or '').split(',')
        for did, onum in zip(dept_ids, order_nums):
            dept = await dept_dao.get_dept_by_id(query_db, int(did))
            if dept:
                dept.order_num = int(onum)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@deptController.delete('/system/dept/{dept_id}')
@log_decorator(title='部门管理', business_type=BusinessType.DELETE)
async def remove(request: Request, dept_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        if await dept_dao.has_child(query_db, dept_id):
            return ResponseUtil.failure(msg='存在下级部门,不允许删除')
        if await dept_dao.check_dept_exist_user(query_db, dept_id):
            return ResponseUtil.failure(msg='部门存在用户,不允许删除')
        dept = await dept_dao.get_dept_by_id(query_db, dept_id)
        if not dept:
            return ResponseUtil.failure(msg='部门不存在')
        dept.del_flag = '2'  # 逻辑删除
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
