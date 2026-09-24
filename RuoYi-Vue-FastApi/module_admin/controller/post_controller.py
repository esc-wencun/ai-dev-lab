"""
岗位管理控制器（spec-03，对应java版SysPostController）
"""
from fastapi import APIRouter, Request, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysPost
from module_admin.dao import post_dao
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, ExcelColumn
from utils.log_util import logger

postController = APIRouter()


@postController.get('/system/post/list', dependencies=[Depends(require_perm('system:post:list'))])
async def list_post(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        posts = await post_dao.get_post_list(
            query_db, params.get('postCode', ''), params.get('postName', ''),
            params.get('status', ''))
        rows = transform_result(posts)
        # paginate需要query对象，这里数据已全取，手动分页包装
        from utils.page_util import get_page_domain
        domain = get_page_domain(request)
        start = domain.offset
        page_rows = rows[start:start + domain.page_size]
        return ResponseUtil.success(msg='查询成功',
                                    dict_content={'rows': page_rows, 'total': len(rows)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.post('/system/post/export', dependencies=[Depends(require_perm('system:post:export'))])
@log_decorator(title='岗位管理', business_type=BusinessType.EXPORT)
async def export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        posts = await post_dao.get_post_list(query_db)
        columns = [
            ExcelColumn('岗位编号', 'postId', width=10),
            ExcelColumn('岗位编码', 'postCode'),
            ExcelColumn('岗位名称', 'postName'),
            ExcelColumn('岗位排序', 'postSort', width=10),
        ]
        return export_excel('岗位数据', columns, transform_result(posts), '岗位数据')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.get('/system/post/optionselect')
async def optionselect(query_db: AsyncSession = Depends(get_db)):
    try:
        posts = await post_dao.get_post_list(query_db)
        return ResponseUtil.success(data=transform_result(posts))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.get('/system/post/{post_id}', dependencies=[Depends(require_perm('system:post:query'))])
async def get_info(request: Request, post_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        post = await post_dao.get_post_by_id(query_db, post_id)
        return ResponseUtil.success(dict_content={'data': transform_result(post) if post else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.post('/system/post', dependencies=[Depends(require_perm('system:post:add'))])
@log_decorator(title='岗位管理', business_type=BusinessType.INSERT)
async def add(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        if await post_dao.check_post_name_unique(query_db, body.get('postName', '')):
            return ResponseUtil.failure(msg=f"新增岗位'{body.get('postName')}'失败，岗位名称已存在")
        if await post_dao.check_post_code_unique(query_db, body.get('postCode', '')):
            return ResponseUtil.failure(msg=f"新增岗位'{body.get('postName')}'失败，岗位编码已存在")
        post = SysPost(
            post_code=body.get('postCode', ''),
            post_name=body.get('postName', ''),
            post_sort=body.get('postSort', 0),
            status=body.get('status', '0'),
            create_by=login_user.get('user_name', ''),
            remark=body.get('remark'),
        )
        import datetime as _dt
        post.create_time = _dt.datetime.now()
        await post_dao.insert_post(query_db, post)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.put('/system/post', dependencies=[Depends(require_perm('system:post:edit'))])
@log_decorator(title='岗位管理', business_type=BusinessType.UPDATE)
async def edit(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        post_id = body.get('postId')
        post = await post_dao.get_post_by_id(query_db, post_id)
        if not post:
            return ResponseUtil.failure(msg='岗位不存在')
        if await post_dao.check_post_name_unique(query_db, body.get('postName', ''), post_id):
            return ResponseUtil.failure(msg=f"修改岗位'{body.get('postName')}'失败，岗位名称已存在")
        if await post_dao.check_post_code_unique(query_db, body.get('postCode', ''), post_id):
            return ResponseUtil.failure(msg=f"修改岗位'{body.get('postName')}'失败，岗位编码已存在")
        post.post_code = body.get('postCode', post.post_code)
        post.post_name = body.get('postName', post.post_name)
        post.post_sort = body.get('postSort', post.post_sort)
        post.status = body.get('status', post.status)
        post.remark = body.get('remark', post.remark)
        post.update_by = login_user.get('user_name', '')
        import datetime as _dt
        post.update_time = _dt.datetime.now()
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@postController.delete('/system/post/{post_ids}', dependencies=[Depends(require_perm('system:post:remove'))])
@log_decorator(title='岗位管理', business_type=BusinessType.DELETE)
async def remove(request: Request, post_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        failed = []
        for pid in [int(x) for x in post_ids.split(',') if x]:
            if await post_dao.count_post_users(query_db, pid) > 0:
                post = await post_dao.get_post_by_id(query_db, pid)
                failed.append(post.post_name if post else str(pid))
                continue
            post = await post_dao.get_post_by_id(query_db, pid)
            if post:
                await post_dao.delete_post(query_db, post)
        await query_db.commit()
        if failed:
            return ResponseUtil.failure(msg=f"{''.join(f'{name}已分配,不能删除' for name in failed)}")
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
