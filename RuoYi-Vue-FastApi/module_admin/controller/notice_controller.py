"""
通知公告控制器（spec-07，对应java版SysNoticeController，含已读功能sys_notice_read）
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select, func, text
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.entity import SysNotice, SysNoticeRead, SysUser
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.log_util import logger

noticeController = APIRouter()


@noticeController.get('/system/notice/list', dependencies=[Depends(require_perm('system:notice:list'))])
async def list_notice(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysNotice)
        if params.get('noticeTitle'):
            query = query.where(SysNotice.notice_title.like(f"%{params['noticeTitle']}%"))
        if params.get('noticeType'):
            query = query.where(SysNotice.notice_type == params['noticeType'])
        if params.get('createBy'):
            query = query.where(SysNotice.create_by.like(f"%{params['createBy']}%"))
        return await paginate(query_db, query.order_by(SysNotice.notice_id.desc()), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.get('/system/notice/listTop')
async def list_top(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    首页置顶公告（正常状态前10条倒序，附带当前用户已读标记）
    """
    try:
        login_user = await TokenService.get_login_user(request)
        user_id = login_user.get('user_id') if login_user else 0
        notices = (await query_db.execute(
            select(SysNotice).where(SysNotice.status == '0')
            .order_by(SysNotice.notice_id.desc()).limit(10))).scalars().all()
        read_ids = set()
        if user_id and notices:
            rows = (await query_db.execute(
                select(SysNoticeRead.notice_id).where(
                    SysNoticeRead.user_id == user_id,
                    SysNoticeRead.notice_id.in_([n.notice_id for n in notices])))).scalars().all()
            read_ids = set(rows)
        data = transform_result(notices)
        for item in data:
            item['read'] = item['noticeId'] in read_ids
        return ResponseUtil.success(data=data)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.get('/system/notice/{notice_id}')
async def get_info(request: Request, notice_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        notice = (await query_db.execute(
            select(SysNotice).where(SysNotice.notice_id == notice_id))).scalars().first()
        return ResponseUtil.success(dict_content={'data': transform_result(notice) if notice else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.post('/system/notice', dependencies=[Depends(require_perm('system:notice:add'))])
@log_decorator(title='通知公告', business_type=BusinessType.INSERT)
async def add(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        notice = SysNotice(
            notice_title=body.get('noticeTitle', ''),
            notice_type=body.get('noticeType', '1'),
            notice_content=(body.get('noticeContent') or '').encode('utf-8'),
            status=body.get('status', '0'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
            remark=body.get('remark'),
        )
        query_db.add(notice)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.put('/system/notice', dependencies=[Depends(require_perm('system:notice:edit'))])
@log_decorator(title='通知公告', business_type=BusinessType.UPDATE)
async def edit(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        notice_id = body.get('noticeId')
        notice = (await query_db.execute(
            select(SysNotice).where(SysNotice.notice_id == notice_id))).scalars().first()
        if not notice:
            return ResponseUtil.failure(msg='公告不存在')
        notice.notice_title = body.get('noticeTitle', notice.notice_title)
        notice.notice_type = body.get('noticeType', notice.notice_type)
        if body.get('noticeContent') is not None:
            notice.notice_content = body['noticeContent'].encode('utf-8')
        notice.status = body.get('status', notice.status)
        notice.remark = body.get('remark', notice.remark)
        notice.update_by = login_user.get('user_name', '')
        notice.update_time = datetime.now()
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.delete('/system/notice/{notice_ids}',
                         dependencies=[Depends(require_perm('system:notice:remove'))])
@log_decorator(title='通知公告', business_type=BusinessType.DELETE)
async def remove(request: Request, notice_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        for nid in [int(x) for x in notice_ids.split(',') if x]:
            notice = (await query_db.execute(
                select(SysNotice).where(SysNotice.notice_id == nid))).scalars().first()
            if notice:
                await query_db.delete(notice)
            await query_db.execute(
                text('delete from sys_notice_read where notice_id = :nid'),
                {'nid': nid})
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.post('/system/notice/markRead')
async def mark_read(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        login_user = await TokenService.get_login_user(request)
        user_id = login_user.get('user_id')
        notice_id = body.get('noticeId')
        exists = (await query_db.execute(
            select(SysNoticeRead.read_id).where(
                SysNoticeRead.user_id == user_id,
                SysNoticeRead.notice_id == notice_id).limit(1))).first()
        if not exists:
            query_db.add(SysNoticeRead(notice_id=notice_id, user_id=user_id,
                                        read_time=datetime.now()))
            await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.post('/system/notice/markReadAll')
async def mark_read_all(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        login_user = await TokenService.get_login_user(request)
        user_id = login_user.get('user_id')
        ids = [int(x) for x in (body.get('ids') or '').split(',') if x.strip()]
        for nid in ids:
            exists = (await query_db.execute(
                select(SysNoticeRead.read_id).where(
                    SysNoticeRead.user_id == user_id,
                    SysNoticeRead.notice_id == nid).limit(1))).first()
            if not exists:
                query_db.add(SysNoticeRead(notice_id=nid, user_id=user_id,
                                            read_time=datetime.now()))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@noticeController.get('/system/notice/readUsers/list',
                      dependencies=[Depends(require_perm('system:notice:list'))])
async def read_users_list(request: Request, noticeId: int = 0, searchValue: str = '',
                          query_db: AsyncSession = Depends(get_db)):
    try:
        query = select(SysUser, SysNoticeRead.read_time).join(
            SysNoticeRead, SysUser.user_id == SysNoticeRead.user_id
        ).where(
            SysNoticeRead.notice_id == noticeId
        )
        if searchValue:
            query = query.where(SysUser.user_name.like(f"%{searchValue}%"))
        return await paginate(query_db, query, request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


from module_admin.service.login_service import TokenService  # noqa: E402
