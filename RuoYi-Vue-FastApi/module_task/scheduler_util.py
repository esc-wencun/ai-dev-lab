"""
APScheduler调度工具（spec-09，对应java版ScheduleUtils）
cron六位表达式兼容Quartz格式
"""
from datetime import datetime
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger
from utils.log_util import logger

_scheduler: AsyncIOScheduler = None

# 六位Quartz cron（秒 分 时 日 月 周）→ APScheduler CronTrigger
# java misfire_policy: 1立即执行 2执行一次 3放弃执行
_MISFIRE_MAP = {
    '1': 'next',      # misfire立即补跑
    '2': 'next',      # 简化处理：与1同
    '3': 'skip',      # 放弃错过执行
}


def convert_quartz_cron(expr: str) -> CronTrigger:
    """
    Quartz六位cron转APScheduler CronTrigger
    支持 L（周末/月末简写降级处理）、?（任意，同*）
    """
    parts = expr.split()
    if len(parts) not in (6, 7):
        raise ValueError(f'cron表达式字段数错误: {expr}')
    second, minute, hour, day, month, day_of_week = parts[:6]
    year = parts[6] if len(parts) == 7 else None
    # Quartz '?' 等价 '*'
    day = None if day == '?' else day.replace('L', 'last')
    day_of_week = None if day_of_week in ('?', '*') else day_of_week
    if day_of_week and '#' in day_of_week:
        # 周几#序号 -> APScheduler无直接等价，降级为该周几每天
        day_of_week = day_of_week.split('#')[0]
    kwargs = {'second': second, 'minute': minute, 'hour': hour,
              'day': day, 'month': month, 'day_of_week': day_of_week}
    if year:
        kwargs['year'] = year
    # 清理None
    kwargs = {k: v for k, v in kwargs.items() if v is not None}
    return CronTrigger(**kwargs)


def init_scheduler():
    """
    初始化调度器（对应java ScheduleUtils.createScheduleJob的容器初始化）
    """
    global _scheduler
    _scheduler = AsyncIOScheduler(
        job_defaults={'coalesce': True,          # 错过的多次合并为一次（防堆积）
                      'max_instances': 1},        # 对齐java @DisallowConcurrentExecution
    )
    _scheduler.start()
    logger.info('APScheduler调度器启动成功')
    return _scheduler


def shutdown_scheduler():
    global _scheduler
    if _scheduler:
        _scheduler.shutdown(wait=False)
        logger.info('APScheduler调度器已关闭')


def add_job(job_id: int, target: str, cron_expr: str, misfire_policy: str = '3'):
    """
    添加调度任务
    """
    trigger = convert_quartz_cron(cron_expr)
    _scheduler.add_job(
        run_job_safely, trigger=trigger,
        id=str(job_id),
        args=[job_id, target],
        misfire_grace_time=60 if misfire_policy != '3' else None,
        replace_existing=True,
    )


def remove_job(job_id: int):
    if _scheduler.get_job(str(job_id)):
        _scheduler.remove_job(str(job_id))


def pause_job(job_id: int):
    if _scheduler.get_job(str(job_id)):
        _scheduler.pause_job(str(job_id))


def resume_job(job_id: int):
    """
    恢复调度。注意：暂停态任务启动时不会加载进调度器（init只加载status=0），
    job不在调度器时resume是静默空操作——必须从库补载（否则"暂停的任务永远无法启用"）
    """
    if _scheduler.get_job(str(job_id)):
        _scheduler.resume_job(str(job_id))
        return
    # 从库补载（请求上下文必有运行中的事件循环）
    import asyncio
    loop = asyncio.get_running_loop()
    loop.create_task(_load_and_add(job_id))


async def _load_and_add(job_id: int):
    from sqlalchemy import select
    from config.database import AsyncSessionLocal
    from module_admin.entity.do.job_do import SysJob

    async with AsyncSessionLocal() as db:
        job = (await db.execute(
            select(SysJob).where(SysJob.job_id == job_id))).scalars().first()
    if job:
        add_job(job.job_id, job.invoke_target, job.cron_expression, job.misfire_policy)


def run_once(job_id: int, target: str):
    """
    立即执行一次（对应java ScheduleUtils.run）
    """
    import asyncio
    asyncio.create_task(run_job_safely(job_id, target))


async def run_job_safely(job_id: int, invoke_target: str):
    """
    任务执行包装：解析invoke_target（含参数）、执行、前后写sys_job_log（对应java JobExecute包装）
    """
    import asyncio
    from module_task.registry import get_task
    from module_task.target_resolver import parse_target
    from config.database import AsyncSessionLocal
    from module_admin.entity.do.job_do import SysJob, SysJobLog

    async with AsyncSessionLocal() as db:
        job = (await db.execute(
            select_job(job_id))).scalars().first()
        if not job:
            return
        start = datetime.now()
        status, message_, exc = '0', '执行成功', ''
        try:
            bean, method, params = parse_target(invoke_target)
            func = get_task(f'{bean}.{method}')
            result = func(*params)
            if asyncio.iscoroutine(result):
                await result
        except Exception as e:
            status = '1'
            message_ = '执行失败'
            exc = str(e)
            logger.error(f'任务{job.job_name}执行失败: {e}')
        cost = int((datetime.now() - start).total_seconds() * 1000)
        db.add(SysJobLog(
            job_name=job.job_name,
            job_group=job.job_group,
            invoke_target=invoke_target,
            job_message=f'{job.job_name} 总耗时{cost}毫秒',
            status=status,
            exception_info=exc,
            create_time=datetime.now(),
        ))
        await db.commit()


def select_job(job_id: int):
    from sqlalchemy import select
    from module_admin.entity.do.job_do import SysJob
    return select(SysJob).where(SysJob.job_id == job_id)
