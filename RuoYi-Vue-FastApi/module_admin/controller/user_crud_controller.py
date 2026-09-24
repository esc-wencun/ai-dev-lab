"""
用户管理控制器（spec-04，对应java版SysUserController）
权限标识：system:user:*
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends, UploadFile, File
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysUser, SysDept, SysRole, SysPost, SysUserRole, SysUserPost
from module_admin.dao import dept_dao, post_dao
from module_admin.aspect.interface_auth import require_perm
from module_admin.aspect.data_scope import build_data_scope_conditions
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from common.constant import UserConstants, SysConfig as SysConfigKey
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, export_template, parse_excel, ExcelColumn
from utils.pwd_util import PwdUtil
from utils.log_util import logger

userCrudController = APIRouter()


# ---------- 内部工具 ----------

async def _get_role_keys_and_roles(query_db: AsyncSession, user_id: int):
    roles = (await query_db.execute(
        select(SysRole).where(SysRole.del_flag == '0', SysRole.status == '0')
        .order_by(SysRole.role_sort)
    )).scalars().all()
    user_roles = (await query_db.execute(
        select(SysRole.role_id).join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
        .where(SysUserRole.user_id == user_id)
    )).scalars().all()
    return roles, list(user_roles)


async def _get_all_posts(query_db: AsyncSession):
    return (await query_db.execute(
        select(SysPost).order_by(SysPost.post_sort)
    )).scalars().all()


async def _get_user_post_ids(query_db: AsyncSession, user_id: int):
    return list((await query_db.execute(
        select(SysUserPost.post_id).where(SysUserPost.user_id == user_id)
    )).scalars().all())


async def _kick_user_sessions(request: Request, user_id: int):
    """
    停用/改密/删用户后清其所有在线会话（对齐java删除会话逻辑）
    """
    from module_admin.service.login_service import get_redis_cache
    from common.constant import CacheConstants
    cache = get_redis_cache(request)
    for key in await cache.keys_by_prefix(f'{CacheConstants.LOGIN_TOKEN_KEY}{user_id}:'):
        pass  # 会话键结构 login_tokens:{uuid}，需要扫值匹配user_id
    # SCAN所有会话键并解析user_id（会话量小场景可接受；大数据量应维护 user_id->tokens 索引）
    for key in await cache.keys_by_prefix(CacheConstants.LOGIN_TOKEN_KEY):
        data = await cache.get_cache_object(key)
        if isinstance(data, dict) and data.get('user_id') == user_id:
            await cache.delete_object(key)


# ---------- 查询 ----------

@userCrudController.get('/system/user/list', dependencies=[Depends(require_perm('system:user:list'))])
async def list_user(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        login_user = await LoginUserHelper.get_login_user_by_request(request)
        conditions = await build_data_scope_conditions(login_user, SysDept, SysUser)
        query = select(SysUser).where(SysUser.del_flag == '0', *conditions)
        if params.get('userName'):
            query = query.where(SysUser.user_name.like(f"%{params['userName']}%"))
        if params.get('phonenumber'):
            query = query.where(SysUser.phonenumber.like(f"%{params['phonenumber']}%"))
        if params.get('status'):
            query = query.where(SysUser.status == params['status'])
        if params.get('deptId'):
            # 本部门及以下（ancestors前缀匹配）
            dept = await dept_dao.get_dept_by_id(query_db, int(params['deptId']))
            if dept:
                child_ids = [d.dept_id for d in await dept_dao.get_dept_list(query_db)
                             if d.ancestors and (dept.ancestors in d.ancestors or
                                                 d.ancestors.startswith(f"{dept.ancestors},"))]
                query = query.where(SysUser.dept_id.in_(child_ids + [dept.dept_id]))
        if params.get('beginTime'):
            query = query.where(SysUser.create_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysUser.create_time <= params['endTime'] + ' 23:59:59')

        # join部门取deptName
        query = query.add_columns(SysDept)
        query = query.outerjoin(SysDept, SysUser.dept_id == SysDept.dept_id)

        domain_rows = []
        total_q = select(func.count('*')).select_from(
            query.order_by(None).subquery())
        total = (await query_db.execute(total_q)).scalar() or 0
        from utils.page_util import get_page_domain
        domain = get_page_domain(request)
        rows = (await query_db.execute(
            query.offset(domain.offset).limit(domain.page_size))).all()
        for row in rows:
            u, d = row[0], row[1]
            item = transform_result(u)
            item['dept'] = transform_result(d) if d else None
            domain_rows.append(item)
        return ResponseUtil.success(msg='查询成功',
                                    dict_content={'rows': domain_rows, 'total': int(total)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


class LoginUserHelper:
    @staticmethod
    async def get_login_user_by_request(request: Request):
        from module_admin.service.login_service import TokenService
        return await TokenService.get_login_user(request)


@userCrudController.get('/system/user/deptTree', dependencies=[Depends(require_perm('system:user:list'))])
async def dept_tree(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        depts = await dept_dao.get_dept_list(query_db)
        nodes = {d.dept_id: {'id': d.dept_id, 'label': d.dept_name, 'children': []} for d in depts}
        tree = []
        for d in depts:
            node = nodes[d.dept_id]
            parent = nodes.get(d.parent_id)
            if parent is not None and d.parent_id != d.dept_id:
                parent['children'].append(node)
            else:
                tree.append(node)
        return ResponseUtil.success(data=tree)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.get('/system/user/authRole/{user_id}',
                        dependencies=[Depends(require_perm('system:user:edit'))])
async def auth_role_info(request: Request, user_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        user = await get_user_by_id_safe(query_db, user_id)
        if not user:
            return ResponseUtil.failure(msg='用户不存在')
        all_roles, checked = await _get_role_keys_and_roles(query_db, user_id)
        from utils.common_util import transform_result
        return ResponseUtil.success(dict_content={
            'user': transform_result(user),
            'roles': transform_result([r for r in all_roles if r.role_id != 1]),
            'roleIds': checked,
        })
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


async def get_user_by_id_safe(query_db: AsyncSession, user_id: int):
    row = (await query_db.execute(
        select(SysUser).where(SysUser.user_id == user_id, SysUser.del_flag == '0'))).first()
    return row[0] if row else None


async def check_user_data_scope(request: Request, query_db: AsyncSession, user_id: int) -> str:
    """
    数据范围校验（对齐java checkUserDataScope：非admin只能查数据权限内的用户）
    :return: 错误文案；空串=通过
    """
    from module_admin.service.login_service import ADMIN_USER_ID
    login_user = await LoginUserHelper.get_login_user_by_request(request)
    if not login_user or login_user.get('user_id') == ADMIN_USER_ID:
        return ''
    conditions = await build_data_scope_conditions(login_user, SysDept, SysUser)
    visible = (await query_db.execute(
        select(SysUser.user_id).where(SysUser.user_id == user_id, *conditions).limit(1)
    )).first()
    if not visible:
        return '没有权限访问用户数据!'
    return ''


@userCrudController.get('/system/user/{user_id}', dependencies=[Depends(require_perm('system:user:query'))])
async def get_info(request: Request, user_id: int = None, query_db: AsyncSession = Depends(get_db)):
    try:
        result = {}
        if user_id:
            error = await check_user_data_scope(request, query_db, user_id)
            if error:
                return ResponseUtil.failure(msg=error)
            user = await get_user_by_id_safe(query_db, user_id)
            if user:
                result['data'] = transform_result(user)
                result['postIds'] = await _get_user_post_ids(query_db, user_id)
                _, role_ids = await _get_role_keys_and_roles(query_db, user_id)
                result['roleIds'] = role_ids
        all_roles, _ = await _get_role_keys_and_roles(query_db, 0)
        result['roles'] = transform_result([r for r in all_roles if r.role_id != 1])
        result['posts'] = transform_result(await _get_all_posts(query_db))
        return ResponseUtil.success(msg='操作成功', dict_content=result)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ---------- 增删改 ----------

def _check_admin(user_id):
    return user_id == 1


@userCrudController.post('/system/user', dependencies=[Depends(require_perm('system:user:add'))])
@log_decorator(title='用户管理', business_type=BusinessType.INSERT)
async def add_user(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.dao.login_dao import get_config_by_key, get_user_by_user_name
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        user_name = body.get('userName', '')
        if await get_user_by_user_name(query_db, user_name):
            return ResponseUtil.failure(msg=f"新增用户'{user_name}'失败，登录账号已存在")
        if body.get('phonenumber'):
            dup = (await query_db.execute(
                select(SysUser.user_id).where(SysUser.phonenumber == body['phonenumber'],
                                              SysUser.del_flag == '0').limit(1))).first()
            if dup:
                return ResponseUtil.failure(msg=f"新增用户'{user_name}'失败，手机号码已存在")
        if body.get('email'):
            dup = (await query_db.execute(
                select(SysUser.user_id).where(SysUser.email == body['email'],
                                              SysUser.del_flag == '0').limit(1))).first()
            if dup:
                return ResponseUtil.failure(msg=f"新增用户'{user_name}'失败，邮箱账号已存在")

        init_cfg = await get_config_by_key(query_db, SysConfigKey.USER_INIT_PASSWORD)
        init_pwd = init_cfg.config_value if init_cfg else '123456'
        user = SysUser(
            dept_id=body.get('deptId'),
            user_name=user_name,
            nick_name=body.get('nickName', user_name),
            email=body.get('email', ''),
            phonenumber=body.get('phonenumber', ''),
            sex=body.get('sex', '0'),
            status=body.get('status', '0'),
            remark=body.get('remark'),
            password=PwdUtil.get_password_hash(init_pwd),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(user)
        await query_db.flush()

        for rid in body.get('roleIds') or []:
            query_db.add(SysUserRole(user_id=user.user_id, role_id=rid))
        for pid in body.get('postIds') or []:
            query_db.add(SysUserPost(user_id=user.user_id, post_id=pid))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.put('/system/user', dependencies=[Depends(require_perm('system:user:edit'))])
@log_decorator(title='用户管理', business_type=BusinessType.UPDATE)
async def edit_user(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        user_id = body.get('userId')
        if _check_admin(user_id):
            return ResponseUtil.failure(msg='不允许操作超级管理员用户')
        user = await get_user_by_id_safe(query_db, user_id)
        if not user:
            return ResponseUtil.failure(msg='用户不存在')
        error = await check_user_data_scope(request, query_db, user_id)
        if error:
            return ResponseUtil.failure(msg=error)
        if body.get('phonenumber'):
            dup = (await query_db.execute(
                select(SysUser.user_id).where(SysUser.phonenumber == body['phonenumber'],
                                              SysUser.user_id != user_id,
                                              SysUser.del_flag == '0').limit(1))).first()
            if dup:
                return ResponseUtil.failure(msg=f"修改用户'{user.user_name}'失败，手机号码已存在")
        if body.get('email'):
            dup = (await query_db.execute(
                select(SysUser.user_id).where(SysUser.email == body['email'],
                                              SysUser.user_id != user_id,
                                              SysUser.del_flag == '0').limit(1))).first()
            if dup:
                return ResponseUtil.failure(msg=f"修改用户'{user.user_name}'失败，邮箱账号已存在")

        user.dept_id = body.get('deptId', user.dept_id)
        user.nick_name = body.get('nickName', user.nick_name)
        user.email = body.get('email', user.email)
        user.phonenumber = body.get('phonenumber', user.phonenumber)
        user.sex = body.get('sex', user.sex)
        user.status = body.get('status', user.status)
        user.remark = body.get('remark', user.remark)
        user.update_by = login_user.get('user_name', '')
        user.update_time = datetime.now()

        # 重建角色/岗位关联
        new_roles = body.get('roleIds')
        if new_roles is not None:
            await query_db.execute(
                SysUserRole.__table__.delete().where(SysUserRole.user_id == user_id))
            for rid in new_roles:
                query_db.add(SysUserRole(user_id=user_id, role_id=rid))
        new_posts = body.get('postIds')
        if new_posts is not None:
            await query_db.execute(
                SysUserPost.__table__.delete().where(SysUserPost.user_id == user_id))
            for pid in new_posts:
                query_db.add(SysUserPost(user_id=user_id, post_id=pid))
        await query_db.commit()

        # 停用时踢下线
        if user.status == '1':
            await _kick_user_sessions(request, user_id)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.delete('/system/user/{user_ids}',
                           dependencies=[Depends(require_perm('system:user:remove'))])
@log_decorator(title='用户管理', business_type=BusinessType.DELETE)
async def remove_users(request: Request, user_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        ids = [int(x) for x in user_ids.split(',') if x]
        for uid in ids:
            if _check_admin(uid):
                return ResponseUtil.failure(msg='不允许操作超级管理员用户')
        for uid in ids:
            user = await get_user_by_id_safe(query_db, uid)
            if not user:
                continue
            error = await check_user_data_scope(request, query_db, uid)
            if error:
                return ResponseUtil.failure(msg=error)
            user.del_flag = '2'
            await query_db.execute(SysUserRole.__table__.delete().where(SysUserRole.user_id == uid))
            await query_db.execute(SysUserPost.__table__.delete().where(SysUserPost.user_id == uid))
            await _kick_user_sessions(request, uid)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.put('/system/user/resetPwd',
                        dependencies=[Depends(require_perm('system:user:resetPwd'))])
@log_decorator(title='用户管理', business_type=BusinessType.UPDATE)
async def reset_pwd(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        user_id = body.get('userId')
        if _check_admin(user_id):
            return ResponseUtil.failure(msg='不允许操作超级管理员用户')
        user = await get_user_by_id_safe(query_db, user_id)
        if not user:
            return ResponseUtil.failure(msg='用户不存在')
        user.password = PwdUtil.get_password_hash(body.get('password', ''))
        user.pwd_update_date = datetime.now()
        await query_db.commit()
        await _kick_user_sessions(request, user_id)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.put('/system/user/changeStatus',
                        dependencies=[Depends(require_perm('system:user:edit'))])
@log_decorator(title='用户管理', business_type=BusinessType.UPDATE)
async def change_status(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        user_id = body.get('userId')
        if _check_admin(user_id):
            return ResponseUtil.failure(msg='不允许操作超级管理员用户')
        user = await get_user_by_id_safe(query_db, user_id)
        if not user:
            return ResponseUtil.failure(msg='用户不存在')
        user.status = body.get('status', user.status)
        await query_db.commit()
        if user.status == '1':
            await _kick_user_sessions(request, user_id)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ---------- Excel 导入导出（对齐java版SysUserController的export/importData/importTemplate） ----------

# 列定义对照java SysUser的@Excel注解
USER_EXPORT_COLUMNS = [
    ExcelColumn('用户序号', 'userId', width=10),
    ExcelColumn('登录名称', 'userName'),
    ExcelColumn('用户名称', 'nickName'),
    ExcelColumn('用户邮箱', 'email'),
    ExcelColumn('手机号码', 'phonenumber'),
    ExcelColumn('用户性别', 'sex', dict_type={'0': '男', '1': '女', '2': '未知'}),
    ExcelColumn('账号状态', 'status', dict_type={'0': '正常', '1': '停用'}),
    ExcelColumn('部门名称', 'deptName'),  # 嵌套dept.deptName平铺
    ExcelColumn('最后登录时间', 'loginDate', width=30),
]

USER_IMPORT_COLUMNS = [
    ExcelColumn('部门编号', 'deptId', width=10),
    ExcelColumn('登录名称', 'userName'),
    ExcelColumn('用户名称', 'nickName'),
    ExcelColumn('用户邮箱', 'email'),
    ExcelColumn('手机号码', 'phonenumber'),
    ExcelColumn('用户性别', 'sex', dict_type={'0': '男', '1': '女', '2': '未知'}),
    ExcelColumn('账号状态', 'status', dict_type={'0': '正常', '1': '停用'}),
]


async def _query_users_for_export(query_db: AsyncSession, params) -> list:
    """
    导出查询：与list同过滤（不带分页），返回驼峰dict列表（dept平铺为deptName）
    """
    query = select(SysUser).where(SysUser.del_flag == '0')
    if params.get('userName'):
        query = query.where(SysUser.user_name.like(f"%{params['userName']}%"))
    if params.get('phonenumber'):
        query = query.where(SysUser.phonenumber.like(f"%{params['phonenumber']}%"))
    if params.get('status'):
        query = query.where(SysUser.status == params['status'])
    if params.get('beginTime'):
        query = query.where(SysUser.create_time >= params['beginTime'])
    if params.get('endTime'):
        query = query.where(SysUser.create_time <= params['endTime'] + ' 23:59:59')
    users = (await query_db.execute(
        query.order_by(SysUser.user_id))).scalars().all()
    data = transform_result(users)
    # 平铺deptName（对齐java @Excels的dept.deptName联动导出）
    dept_ids = {u.get('deptId') for u in data if u.get('deptId')}
    dept_names = {}
    if dept_ids:
        depts = (await query_db.execute(
            select(SysDept).where(SysDept.dept_id.in_(dept_ids)))).scalars().all()
        dept_names = {d.dept_id: d.dept_name for d in depts}
    for u in data:
        u['deptName'] = dept_names.get(u.get('deptId'), '')
    return data


@userCrudController.post('/system/user/export',
                         dependencies=[Depends(require_perm('system:user:export'))])
@log_decorator(title='用户管理', business_type=BusinessType.EXPORT)
async def export_users(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        data = await _query_users_for_export(query_db, request.query_params)
        return export_excel('用户数据', USER_EXPORT_COLUMNS, data, '用户数据')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.post('/system/user/importTemplate')
async def import_template(request: Request):
    try:
        return export_template('用户数据', USER_IMPORT_COLUMNS, '用户数据',
                               example_rows=[{
                                   'deptId': 100, 'userName': 'ruoyi', 'nickName': '若依',
                                   'email': 'ry@163.com', 'phonenumber': '15888888888',
                                   'sex': '0', 'status': '0',
                               }])
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.post('/system/user/importData',
                         dependencies=[Depends(require_perm('system:user:import'))])
@log_decorator(title='用户管理', business_type=BusinessType.IMPORT)
async def import_data(request: Request, file: UploadFile = File(...),
                      updateSupport: bool = False,
                      query_db: AsyncSession = Depends(get_db)):
    """
    用户导入（校验链对齐java importUser；返回成功/失败明细消息）
    """
    try:
        from module_admin.service.login_service import TokenService
        from module_admin.dao.login_dao import get_config_by_key
        login_user = await TokenService.get_login_user(request) or {}
        oper_name = login_user.get('user_name', '')

        content = await file.read()
        try:
            rows = parse_excel(content, USER_IMPORT_COLUMNS)
        except ValueError as ve:
            return ResponseUtil.failure(msg=str(ve))
        if not rows:
            return ResponseUtil.failure(msg='导入用户数据不能为空！')

        success_num = failure_num = 0
        success_msgs, failure_msgs = [], []
        for row in rows:
            user_name = (row.get('userName') or '').strip()
            try:
                if not user_name:
                    raise ValueError('登录名称不能为空')
                existing = (await query_db.execute(
                    select(SysUser).where(SysUser.user_name == user_name,
                                          SysUser.del_flag == '0'))).scalars().first()
                if existing is None:
                    init_cfg = await get_config_by_key(query_db, SysConfigKey.USER_INIT_PASSWORD)
                    init_pwd = init_cfg.config_value if init_cfg else '123456'
                    user = SysUser(
                        dept_id=row.get('deptId'),
                        user_name=user_name,
                        nick_name=row.get('nickName') or user_name,
                        email=row.get('email') or '',
                        phonenumber=str(row.get('phonenumber') or ''),
                        sex=row.get('sex') or '0',
                        status=row.get('status') or '0',
                        password=PwdUtil.get_password_hash(init_pwd),
                        create_by=oper_name,
                        create_time=datetime.now(),
                    )
                    query_db.add(user)
                    await query_db.flush()
                    success_num += 1
                    success_msgs.append(f'{success_num}、账号 {user_name} 导入成功')
                elif updateSupport:
                    if existing.user_id == 1:
                        raise ValueError('不允许操作超级管理员用户')
                    existing.nick_name = row.get('nickName') or existing.nick_name
                    existing.email = row.get('email') or existing.email
                    existing.phonenumber = str(row.get('phonenumber') or existing.phonenumber or '')
                    existing.sex = row.get('sex') or existing.sex
                    existing.status = row.get('status') or existing.status
                    existing.update_by = oper_name
                    existing.update_time = datetime.now()
                    success_num += 1
                    success_msgs.append(f'{success_num}、账号 {user_name} 更新成功')
                else:
                    failure_num += 1
                    failure_msgs.append(f'{failure_num}、账号 {user_name} 已存在')
            except Exception as ie:
                failure_num += 1
                failure_msgs.append(f'{failure_num}、账号 {user_name} 导入失败：{ie}')
        await query_db.commit()

        if failure_num > 0:
            return ResponseUtil.failure(
                msg='很抱歉，导入失败！共 ' + str(failure_num) + ' 条数据格式不正确，错误如下：'
                    + '<br/>' + '<br/>'.join(failure_msgs))
        return ResponseUtil.success(
            msg='恭喜您，数据已全部导入成功！共 ' + str(success_num) + ' 条，数据如下：'
                + '<br/>' + '<br/>'.join(success_msgs))
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@userCrudController.put('/system/user/authRole',
                        dependencies=[Depends(require_perm('system:user:edit'))])
@log_decorator(title='用户管理', business_type=BusinessType.GRANT)
async def insert_auth_role(request: Request, userId: int = None, roleIds: str = None,
                           query_db: AsyncSession = Depends(get_db)):
    try:
        if not userId:
            return ResponseUtil.failure(msg='参数错误')
        await query_db.execute(
            SysUserRole.__table__.delete().where(SysUserRole.user_id == userId))
        for rid in (roleIds.split(',') if roleIds else []):
            if rid.strip():
                query_db.add(SysUserRole(user_id=userId, role_id=int(rid)))
        await query_db.commit()
        await _kick_user_sessions(request, userId)  # 刷新在线用户权限
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
