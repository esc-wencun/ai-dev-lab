from fastapi import FastAPI
from sqlalchemy import select
from contextlib import asynccontextmanager
from sub_applications.handle import handle_sub_applications
from middlewares.handle import handle_middleware
from exceptions.handle import handle_exception
from module_admin.controller.login_controller import loginController
from module_admin.controller.user_controller import userController, commonController
from module_admin.controller.dept_controller import deptController
from module_admin.controller.post_controller import postController
from module_admin.controller.user_crud_controller import userCrudController
from module_admin.controller.role_controller import roleController
from module_admin.controller.menu_controller import menuController
from module_admin.controller.dict_config_controller import dictTypeController, dictDataController, configController
from module_admin.controller.notice_controller import noticeController
from module_admin.controller.monitor_controller import (onlineController, serverController,
                                                         cacheController, operlogController,
                                                         logininforController)
from module_admin.controller.job_controller import jobController, jobLogController
from module_admin.controller.gen_controller import genController
from config.env import AppConfig
from config.get_redis import RedisUtil
from config.redis_cache import RedisCache
from utils.log_util import logger


# 生命周期事件
@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info(f"{AppConfig.app_name}开始启动")
    # 连接redis
    app.state.redis = await RedisUtil.create_redis_pool()
    # 注入RedisCache门面（业务代码统一经此访问redis，禁止直接使用app.state.redis）
    app.state.redis_cache = RedisCache(app.state.redis)
    # 启动时加载系统参数配置到redis（与java版SysConfigServiceImpl的缓存机制一致）
    await init_sys_config(app)
    # 初始化定时任务调度器并加载status=0的任务（对应java ScheduleUtils.initSystemScheduler）
    await init_scheduler_jobs(app)
    logger.info(f"{AppConfig.app_name}启动成功")
    yield
    from module_task import scheduler_util
    scheduler_util.shutdown_scheduler()
    await RedisUtil.close_redis_pool(app)
    logger.info(f"{AppConfig.app_name}关闭成功")


async def init_scheduler_jobs(app: FastAPI):
    """
    加载status='0'的定时任务到APScheduler（对应java版启动时恢复调度）
    """
    import module_task.registry  # noqa: F401 触发任务注册
    import module_task.ry_task  # noqa: F401 注册java版预置的ryTask三个方法
    from config.get_db import AsyncSessionLocal
    from module_admin.entity.do.job_do import SysJob
    from module_task import scheduler_util
    try:
        scheduler_util.init_scheduler()
        async with AsyncSessionLocal() as session:
            jobs = (await session.execute(
                select(SysJob).where(SysJob.status == '0'))).scalars().all()
            for job in jobs:
                try:
                    scheduler_util.add_job(job.job_id, job.invoke_target,
                                            job.cron_expression, job.misfire_policy)
                except Exception as je:
                    logger.error(f'任务{job.job_id}加载失败: {je}')
        logger.info(f'定时任务加载完成，共{len(jobs)}个')
    except Exception as e:
        logger.error(f'定时任务调度器初始化失败：{e}')


async def init_sys_config(app: FastAPI):
    """
    加载参数配置到redis（key前缀与java版CacheConstants.SYS_CONFIG_KEY保持一致；
    java版为@PostConstruct启动预热 + CRUD增量维护，本函数对应启动预热）
    """
    from config.get_db import AsyncSessionLocal
    from module_admin.dao.login_dao import get_config_all
    from config.redis_cache import RedisCache
    from common.constant import CacheConstants
    try:
        cache = RedisCache(app.state.redis)
        async with AsyncSessionLocal() as session:
            configs = await get_config_all(session)
            for config in configs:
                await cache.set_cache_object(
                    cache.build_key(CacheConstants.SYS_CONFIG_KEY, config.config_key),
                    config.config_value
                )
        logger.info("系统参数配置缓存加载成功")
    except Exception as e:
        logger.error(f"系统参数配置缓存加载失败：{e}")


# 初始化FastAPI对象
app = FastAPI(
    title=AppConfig.app_name,
    description=f'{AppConfig.app_name}接口文档',
    version=AppConfig.app_version,
    lifespan=lifespan
)

# 挂载子应用
handle_sub_applications(app)
# 加载中间件处理方法
handle_middleware(app)
# 加载全局异常处理方法
handle_exception(app)


# 加载路由列表
controller_list = [
    {'router': loginController, 'tags': ['登录模块']},
    {'router': userController, 'tags': ['个人中心/注册']},
    {'router': commonController, 'tags': ['通用模块']},
    {'router': deptController, 'tags': ['系统管理-部门管理']},
    {'router': postController, 'tags': ['系统管理-岗位管理']},
    {'router': userCrudController, 'tags': ['系统管理-用户管理']},
    {'router': roleController, 'tags': ['系统管理-角色管理']},
    {'router': menuController, 'tags': ['系统管理-菜单管理']},
    {'router': dictTypeController, 'tags': ['系统管理-字典类型']},
    {'router': dictDataController, 'tags': ['系统管理-字典数据']},
    {'router': configController, 'tags': ['系统管理-参数管理']},
    {'router': noticeController, 'tags': ['系统管理-通知公告']},
    {'router': onlineController, 'tags': ['系统监控-在线用户']},
    {'router': serverController, 'tags': ['系统监控-服务监控']},
    {'router': cacheController, 'tags': ['系统监控-缓存监控']},
    {'router': operlogController, 'tags': ['系统监控-操作日志']},
    {'router': logininforController, 'tags': ['系统监控-登录日志']},
    {'router': jobController, 'tags': ['系统监控-定时任务']},
    {'router': jobLogController, 'tags': ['系统监控-任务日志']},
    {'router': genController, 'tags': ['系统工具-代码生成']},
]

for controller in controller_list:
    app.include_router(router=controller.get('router'), tags=controller.get('tags'))


# API文档兼容（前端工具页依赖springdoc路径：/v3/api-docs 与 /swagger-ui.html）
@app.get('/v3/api-docs', include_in_schema=False)
async def api_docs_redirect():
    from fastapi.responses import RedirectResponse
    return RedirectResponse(url='/openapi.json')


@app.get('/swagger-ui.html', include_in_schema=False)
async def swagger_ui_redirect():
    from fastapi.responses import RedirectResponse
    return RedirectResponse(url='/docs')
