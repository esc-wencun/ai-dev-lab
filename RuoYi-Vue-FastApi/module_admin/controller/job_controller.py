"""
定时任务管理控制器（spec-09，对应java版SysJobController/SysJobLogController）
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.job_do import SysJob, SysJobLog
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, ExcelColumn
from module_task import scheduler_util, registry
from utils.log_util import logger

jobController = APIRouter()
jobLogController = APIRouter()


@jobController.get('/monitor/job/list', dependencies=[Depends(require_perm('monitor:job:list'))])
async def list_job(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysJob)
        if params.get('jobName'):
            query = query.where(SysJob.job_name.like(f"%{params['jobName']}%"))
        if params.get('jobGroup'):
            query = query.where(SysJob.job_group == params['jobGroup'])
        if params.get('status'):
            query = query.where(SysJob.status == params['status'])
        return await paginate(query_db, query.order_by(SysJob.job_id), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.get('/monitor/job/{job_id}', dependencies=[Depends(require_perm('monitor:job:query'))])
async def get_info(request: Request, job_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        job = (await query_db.execute(
            select(SysJob).where(SysJob.job_id == job_id))).scalars().first()
        return ResponseUtil.success(dict_content={'data': transform_result(job) if job else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.post('/monitor/job', dependencies=[Depends(require_perm('monitor:job:add'))])
@log_decorator(title='定时任务', business_type=BusinessType.INSERT)
async def add_job(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        from module_task.target_resolver import validate_target
        login_user = await TokenService.get_login_user(request) or {}
        target = body.get('invokeTarget', '')
        cron = body.get('cronExpression', '')
        # 完整校验链（文案对齐java SysJobController：cron/rmi/ldap/http/违规/白名单）
        error = validate_target(target, lambda name: registry.is_registered(name))
        if error:
            return ResponseUtil.failure(msg=f'新增任务{body.get("jobName")}失败，{error}')
        try:
            scheduler_util.convert_quartz_cron(cron)
        except Exception:
            return ResponseUtil.failure(msg=f'新增任务{body.get("jobName")}失败，Cron表达式不正确')
        job = SysJob(
            job_name=body.get('jobName', ''),
            job_group=body.get('jobGroup', 'DEFAULT'),
            invoke_target=target,
            cron_expression=cron,
            misfire_policy=body.get('misfirePolicy', '3'),
            concurrent=body.get('concurrent', '1'),
            status=body.get('status', '0'),
            remark=body.get('remark'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(job)
        await query_db.commit()
        await query_db.refresh(job)
        if job.status == '0':
            scheduler_util.add_job(job.job_id, target, cron, job.misfire_policy)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.put('/monitor/job', dependencies=[Depends(require_perm('monitor:job:edit'))])
@log_decorator(title='定时任务', business_type=BusinessType.UPDATE)
async def edit_job(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        from module_task.target_resolver import validate_target
        login_user = await TokenService.get_login_user(request) or {}
        job_id = body.get('jobId')
        job = (await query_db.execute(
            select(SysJob).where(SysJob.job_id == job_id))).scalars().first()
        if not job:
            return ResponseUtil.failure(msg='任务不存在')
        target = body.get('invokeTarget', job.invoke_target)
        cron = body.get('cronExpression', job.cron_expression)
        error = validate_target(target, lambda name: registry.is_registered(name))
        if error:
            return ResponseUtil.failure(msg=f'修改任务{body.get("jobName")}失败，{error}')
        try:
            scheduler_util.convert_quartz_cron(cron)
        except Exception:
            return ResponseUtil.failure(msg=f'修改任务{body.get("jobName")}失败，Cron表达式不正确')
        job.job_name = body.get('jobName', job.job_name)
        job.job_group = body.get('jobGroup', job.job_group)
        job.invoke_target = target
        job.cron_expression = cron
        job.misfire_policy = body.get('misfirePolicy', job.misfire_policy)
        job.concurrent = body.get('concurrent', job.concurrent)
        job.status = body.get('status', job.status)
        job.remark = body.get('remark', job.remark)
        job.update_by = login_user.get('user_name', '')
        job.update_time = datetime.now()
        await query_db.commit()
        scheduler_util.remove_job(job_id)
        if job.status == '0':
            scheduler_util.add_job(job_id, target, cron, job.misfire_policy)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.put('/monitor/job/changeStatus',
                   dependencies=[Depends(require_perm('monitor:job:changeStatus'))])
@log_decorator(title='定时任务', business_type=BusinessType.UPDATE)
async def change_status(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        job_id = body.get('jobId')
        status = body.get('status')
        job = (await query_db.execute(
            select(SysJob).where(SysJob.job_id == job_id))).scalars().first()
        if not job:
            return ResponseUtil.failure(msg='任务不存在')
        job.status = status
        await query_db.commit()
        if status == '0':
            scheduler_util.resume_job(job_id)
        else:
            scheduler_util.pause_job(job_id)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.put('/monitor/job/run',
                   dependencies=[Depends(require_perm('monitor:job:changeStatus'))])
@log_decorator(title='定时任务', business_type=BusinessType.RUN if hasattr(BusinessType, 'RUN') else BusinessType.UPDATE)
async def run_job(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        job_id = body.get('jobId')
        job = (await query_db.execute(
            select(SysJob).where(SysJob.job_id == job_id))).scalars().first()
        if not job:
            return ResponseUtil.failure(msg='任务不存在')
        scheduler_util.run_once(job_id, job.invoke_target)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobController.delete('/monitor/job/{job_ids}',
                      dependencies=[Depends(require_perm('monitor:job:remove'))])
@log_decorator(title='定时任务', business_type=BusinessType.DELETE)
async def remove_jobs(request: Request, job_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        for jid in [int(x) for x in job_ids.split(',') if x]:
            job = (await query_db.execute(
                select(SysJob).where(SysJob.job_id == jid))).scalars().first()
            if job:
                await query_db.delete(job)
                scheduler_util.remove_job(jid)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobLogController.get('/monitor/jobLog/list',
                      dependencies=[Depends(require_perm('monitor:joblog:list'))])
async def list_job_log(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysJobLog)
        if params.get('jobName'):
            query = query.where(SysJobLog.job_name.like(f"%{params['jobName']}%"))
        if params.get('jobGroup'):
            query = query.where(SysJobLog.job_group == params['jobGroup'])
        if params.get('status'):
            query = query.where(SysJobLog.status == params['status'])
        if params.get('beginTime'):
            query = query.where(SysJobLog.create_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysJobLog.create_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query.order_by(SysJobLog.job_log_id.desc()), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobLogController.delete('/monitor/jobLog/{job_log_ids}',
                         dependencies=[Depends(require_perm('monitor:joblog:remove'))])
@log_decorator(title='任务日志', business_type=BusinessType.DELETE)
async def remove_job_log(request: Request, job_log_ids: str,
                         query_db: AsyncSession = Depends(get_db)):
    """
    批量删除任务日志（clean为保留字路由，与Java一致：/clean定义在前优先匹配）
    """
    try:
        if job_log_ids == 'clean':
            await query_db.execute(SysJobLog.__table__.delete())
            await query_db.commit()
            return ResponseUtil.success()
        ids = [int(x) for x in job_log_ids.split(',') if x]
        await query_db.execute(SysJobLog.__table__.delete().where(SysJobLog.job_log_id.in_(ids)))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobLogController.delete('/monitor/jobLog/clean',
                         dependencies=[Depends(require_perm('monitor:joblog:remove'))])
@log_decorator(title='任务日志', business_type=BusinessType.CLEAN)
async def clean_job_log(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    清空任务日志（FastAPI路由匹配：/clean会被/{job_log_ids}优先捕获，
    因此remove_job_log内对'clean'做了转发处理，本路由作为API文档呈现）
    """
    try:
        await query_db.execute(SysJobLog.__table__.delete())
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 导出（spec-04挂账补齐，列对齐java SysJob/SysJobLog @Excel） ====================

MISFIRE_MAP = {'0': '默认', '1': '立即触发执行', '2': '触发一次执行', '3': '不触发立即执行'}
CONCURRENT_MAP = {'0': '允许', '1': '禁止'}
JOB_STATUS_MAP = {'0': '正常', '1': '暂停'}


@jobController.post('/monitor/job/export',
                    dependencies=[Depends(require_perm('monitor:job:export'))])
@log_decorator(title='定时任务', business_type=BusinessType.EXPORT)
async def export_jobs(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        jobs = (await query_db.execute(select(SysJob).order_by(SysJob.job_id))).scalars().all()
        columns = [
            ExcelColumn('任务序号', 'jobId', width=10),
            ExcelColumn('任务名称', 'jobName'),
            ExcelColumn('任务组名', 'jobGroup'),
            ExcelColumn('调用目标字符串', 'invokeTarget'),
            ExcelColumn('执行表达式', 'cronExpression'),
            ExcelColumn('计划策略', 'misfirePolicy', dict_type=MISFIRE_MAP),
            ExcelColumn('并发执行', 'concurrent', dict_type=CONCURRENT_MAP),
            ExcelColumn('任务状态', 'status', dict_type=JOB_STATUS_MAP),
        ]
        from utils.excel_util import export_excel
        return export_excel('定时任务', columns, transform_result(jobs), '定时任务')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@jobLogController.post('/monitor/jobLog/export',
                       dependencies=[Depends(require_perm('monitor:joblog:export'))])
@log_decorator(title='任务日志', business_type=BusinessType.EXPORT)
async def export_job_logs(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysJobLog)
        if params.get('jobName'):
            query = query.where(SysJobLog.job_name.like(f"%{params['jobName']}%"))
        if params.get('status'):
            query = query.where(SysJobLog.status == params['status'])
        logs = (await query_db.execute(
            query.order_by(SysJobLog.job_log_id.desc()))).scalars().all()
        columns = [
            ExcelColumn('日志序号', 'jobLogId', width=10),
            ExcelColumn('任务名称', 'jobName'),
            ExcelColumn('任务组名', 'jobGroup'),
            ExcelColumn('调用目标字符串', 'invokeTarget'),
            ExcelColumn('日志信息', 'jobMessage'),
            ExcelColumn('执行状态', 'status', dict_type={'0': '正常', '1': '失败'}),
            ExcelColumn('异常信息', 'exceptionInfo'),
        ]
        from utils.excel_util import export_excel
        return export_excel('任务日志', columns, transform_result(logs), '任务日志')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
