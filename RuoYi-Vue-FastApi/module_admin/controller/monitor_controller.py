"""
监控模块控制器（spec-08）：在线用户/服务监控/缓存监控/操作日志/登录日志
对应java版SysUserOnlineController/ServerController/CacheController/SysOperlogController/SysLogininforController
"""
import time
import platform
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select, func, delete
from sqlalchemy.ext.asyncio import AsyncSession
import psutil
from config.get_db import get_db
from config.redis_cache import RedisCache
from module_admin.entity.do.entity import SysOperLog, SysLogininfor
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from common.constant import CacheConstants
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, ExcelColumn
from utils.log_util import logger

onlineController = APIRouter()
serverController = APIRouter()
cacheController = APIRouter()
operlogController = APIRouter()
logininforController = APIRouter()


# ==================== 在线用户 ====================

@onlineController.get('/monitor/online/list',
                      dependencies=[Depends(require_perm('monitor:online:list'))])
async def online_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        cache: RedisCache = request.app.state.redis_cache
        result = []
        for key in await cache.keys_by_prefix(CacheConstants.LOGIN_TOKEN_KEY):
            data = await cache.get_cache_object(key)
            if not isinstance(data, dict):
                continue
            item = {
                'tokenId': data.get('token'),
                'userId': data.get('user_id'),
                'userName': data.get('user_name'),
                'deptName': '',
                'ipaddr': data.get('ipaddr', ''),
                'loginLocation': data.get('login_location', ''),
                'browser': data.get('browser', ''),
                'os': data.get('os', ''),
                'loginTime': data.get('login_time'),
            }
            if params.get('userName') and params['userName'] not in (item['userName'] or ''):
                continue
            if params.get('ipaddr') and params['ipaddr'] not in (item['ipaddr'] or ''):
                continue
            result.append(item)
        return ResponseUtil.success(msg='查询成功',
                                    dict_content={'rows': result, 'total': len(result)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@onlineController.delete('/monitor/online/{token_id}',
                         dependencies=[Depends(require_perm('monitor:online:forceLogout'))])
@log_decorator(title='在线用户', business_type=BusinessType.FORCE)
async def force_logout(request: Request, token_id: str, query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        await cache.delete_login_user(token_id)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 服务监控 ====================

@serverController.get('/monitor/server',
                      dependencies=[Depends(require_perm('monitor:server:list'))])
async def server_info(request: Request):
    try:
        mem = psutil.virtual_memory()
        disk_used_total = disk_free_total = 0
        disks = []
        for part in psutil.disk_partitions():
            try:
                usage = psutil.disk_usage(part.mountpoint)
                disks.append({
                    'dirName': part.mountpoint,
                    'sysTypeName': part.fstype,
                    'typeName': '',
                    'total': f'{usage.total / (1024**3):.1f} GB',
                    'free': f'{usage.free / (1024**3):.1f} GB',
                    'used': f'{usage.used / (1024**3):.1f} GB',
                    'usage': f'{usage.percent}%',
                })
                disk_used_total += usage.used
                disk_free_total += usage.free
            except (PermissionError, OSError):
                continue

        boot = datetime.fromtimestamp(psutil.boot_time())
        now = datetime.now()
        run_days = (now - boot).days
        run_hours = (now - boot).seconds // 3600
        run_minutes = (now - boot).seconds % 3600 // 60

        cpu_times = psutil.cpu_times()
        cpu_percent = psutil.cpu_percent(interval=0.5)

        data = {
            'cpu': {
                'cpuNum': psutil.cpu_count(logical=True),
                'total': psutil.cpu_times().user + psutil.cpu_times().system + psutil.cpu_times().idle,
                'sys': round(cpu_times.system, 1),
                'used': round(cpu_times.user, 1),
                'wait': round(getattr(cpu_times, 'iowait', 0), 1),
                'free': round(cpu_times.idle, 1),
                'usage': cpu_percent,
            },
            'mem': {
                'total': round(mem.total / (1024**3), 1),
                'used': round(mem.used / (1024**3), 1),
                'free': round(mem.available / (1024**3), 1),
                'usage': mem.percent,
            },
            'jvm': {  # python运行时信息（对位java jvm信息，字段名对齐前端）
                'name': platform.python_implementation(),
                'version': platform.python_version(),
                'startTime': boot.strftime('%Y-%m-%d %H:%M:%S'),
                'runTime': f'{run_days}天{run_hours}小时{run_minutes}分钟',
                'home': platform.system() + ' ' + platform.release(),
            },
            'sys': {
                'computerName': platform.node(),
                'osName': platform.system(),
                'computerIp': request.client.host if request.client else '',
                'osArch': platform.machine(),
            },
            'sysFiles': disks,
        }
        return ResponseUtil.success(dict_content={'data': data})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 缓存监控 ====================

CACHE_NAMES = [
    {'cacheName': 'login_tokens', 'remark': '用户信息缓存'},
    {'cacheName': 'sys_config', 'remark': '配置信息缓存'},
    {'cacheName': 'sys_dict', 'remark': '数据字典缓存'},
    {'cacheName': 'captcha_codes', 'remark': '图片验证码缓存'},
    {'cacheName': 'repeat_submit', 'remark': '防重提交缓存'},
    {'cacheName': 'rate_limit', 'remark': '限流缓存'},
    {'cacheName': 'pwd_err_cnt', 'remark': '密码错误次数缓存'},
]


@cacheController.get('/monitor/cache',
                     dependencies=[Depends(require_perm('monitor:cache:list'))])
async def cache_info(request: Request):
    try:
        redis = request.app.state.redis._redis if hasattr(request.app.state.redis, '_redis') \
            else request.app.state.redis
        info = await redis.info()
        db_size = await redis.dbsize()
        command_stats = []
        for key, val in (info.get('commandstats') or {}).items():
            command_stats.append({'name': key.replace('cmdstat_', '').split('_')[0],
                                  'value': val.get('calls', 0)})
        return ResponseUtil.success(dict_content={
            'info': {
                'dbSize': db_size,
                'uptime': info.get('uptime_in_seconds'),
                'connected_clients': info.get('connected_clients'),
                'used_memory': info.get('used_memory_human'),
                'max_memory_human': info.get('maxmemory_human'),
                'mem_fragmentation_ratio': info.get('mem_fragmentation_ratio'),
                'total_connections_received': info.get('total_connections_received'),
                'total_commands_processed': info.get('total_commands_processed'),
            },
            'dbSize': db_size,
            'commandStats': command_stats,
        })
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@cacheController.get('/monitor/cache/getNames',
                     dependencies=[Depends(require_perm('monitor:cache:list'))])
async def cache_names(request: Request):
    return ResponseUtil.success(data=CACHE_NAMES)


@cacheController.get('/monitor/cache/getKeys/{cache_name}',
                     dependencies=[Depends(require_perm('monitor:cache:list'))])
async def cache_keys(request: Request, cache_name: str):
    try:
        cache: RedisCache = request.app.state.redis_cache
        keys = await cache.keys_by_prefix(f'{cache_name}:')
        return ResponseUtil.success(data=[k for k in keys])
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@cacheController.get('/monitor/cache/getValue/{cache_name}/{cache_key}',
                     dependencies=[Depends(require_perm('monitor:cache:list'))])
async def cache_value(request: Request, cache_name: str, cache_key: str):
    try:
        cache: RedisCache = request.app.state.redis_cache
        full_key = f'{cache_name}:{cache_key}'
        value = await cache.get_cache_object(full_key)
        import json as _json
        return ResponseUtil.success(dict_content={
            'cacheKey': full_key,
            'cacheValue': _json.dumps(value, ensure_ascii=False) if not isinstance(value, str) else value,
            'remark': '',
        })
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@cacheController.delete('/monitor/cache/clearCacheName/{cache_name}',
                        dependencies=[Depends(require_perm('monitor:cache:list'))])
async def clear_cache_name(request: Request, cache_name: str):
    try:
        cache: RedisCache = request.app.state.redis_cache
        for key in await cache.keys_by_prefix(f'{cache_name}:'):
            await cache.delete_object(key)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@cacheController.delete('/monitor/cache/clearCacheKey/{cache_key}',
                        dependencies=[Depends(require_perm('monitor:cache:list'))])
async def clear_cache_key(request: Request, cache_key: str):
    try:
        cache: RedisCache = request.app.state.redis_cache
        await cache.delete_object(cache_key)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@cacheController.delete('/monitor/cache/clearCacheAll',
                        dependencies=[Depends(require_perm('monitor:cache:list'))])
async def clear_cache_all(request: Request):
    try:
        cache: RedisCache = request.app.state.redis_cache
        for item in CACHE_NAMES:
            for key in await cache.keys_by_prefix(f"{item['cacheName']}:"):
                await cache.delete_object(key)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 操作日志 ====================

@operlogController.get('/monitor/operlog/list',
                       dependencies=[Depends(require_perm('monitor:operlog:list'))])
async def operlog_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysOperLog)
        if params.get('title'):
            query = query.where(SysOperLog.title.like(f"%{params['title']}%"))
        if params.get('operName'):
            query = query.where(SysOperLog.oper_name.like(f"%{params['operName']}%"))
        if params.get('businessType') is not None and params.get('businessType') != '':
            query = query.where(SysOperLog.business_type == int(params['businessType']))
        if params.get('status') is not None and params.get('status') != '':
            query = query.where(SysOperLog.status == int(params['status']))
        if params.get('beginTime'):
            query = query.where(SysOperLog.oper_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysOperLog.oper_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query.order_by(SysOperLog.oper_id.desc()), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@operlogController.delete('/monitor/operlog/{oper_ids}',
                          dependencies=[Depends(require_perm('monitor:operlog:remove'))])
@log_decorator(title='操作日志', business_type=BusinessType.DELETE)
async def remove_operlog(request: Request, oper_ids: str,
                         query_db: AsyncSession = Depends(get_db)):
    try:
        ids = [int(x) for x in oper_ids.split(',') if x]
        await query_db.execute(SysOperLog.__table__.delete().where(SysOperLog.oper_id.in_(ids)))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@operlogController.delete('/monitor/operlog/clean',
                          dependencies=[Depends(require_perm('monitor:operlog:remove'))])
@log_decorator(title='操作日志', business_type=BusinessType.CLEAN)
async def clean_operlog(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        await query_db.execute(SysOperLog.__table__.delete())
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 登录日志 ====================

@logininforController.get('/monitor/logininfor/list',
                          dependencies=[Depends(require_perm('monitor:logininfor:list'))])
async def logininfor_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysLogininfor)
        if params.get('userName'):
            query = query.where(SysLogininfor.user_name.like(f"%{params['userName']}%"))
        if params.get('ipaddr'):
            query = query.where(SysLogininfor.ipaddr.like(f"%{params['ipaddr']}%"))
        if params.get('status'):
            query = query.where(SysLogininfor.status == params['status'])
        if params.get('beginTime'):
            query = query.where(SysLogininfor.login_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysLogininfor.login_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query.order_by(SysLogininfor.info_id.desc()), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@logininforController.delete('/monitor/logininfor/{info_ids}',
                             dependencies=[Depends(require_perm('monitor:logininfor:remove'))])
@log_decorator(title='登录日志', business_type=BusinessType.DELETE)
async def remove_logininfor(request: Request, info_ids: str,
                            query_db: AsyncSession = Depends(get_db)):
    try:
        ids = [int(x) for x in info_ids.split(',') if x]
        await query_db.execute(SysLogininfor.__table__.delete().where(SysLogininfor.info_id.in_(ids)))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@logininforController.delete('/monitor/logininfor/clean',
                             dependencies=[Depends(require_perm('monitor:logininfor:remove'))])
@log_decorator(title='登录日志', business_type=BusinessType.CLEAN)
async def clean_logininfor(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        await query_db.execute(SysLogininfor.__table__.delete())
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@logininforController.get('/monitor/logininfor/unlock/{user_name}',
                          dependencies=[Depends(require_perm('monitor:logininfor:unlock'))])
async def unlock(request: Request, user_name: str, query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        key = cache.build_key(CacheConstants.PWD_ERR_CNT_KEY, user_name)
        await cache.delete_object(key)
        return ResponseUtil.success(msg=f'{user_name}账户解锁成功')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 日志导出（spec-04挂账补齐） ====================

BUSINESS_TYPE_MAP = {'0': '其它', '1': '新增', '2': '修改', '3': '删除', '4': '授权',
                     '5': '导出', '6': '导入', '7': '强退', '8': '生成代码', '9': '清空数据'}
OPERATOR_TYPE_MAP = {'0': '其它', '1': '后台用户', '2': '手机端用户'}
STATUS_MAP = {'0': '正常', '1': '异常'}
LOGIN_STATUS_MAP = {'0': '成功', '1': '失败'}


@operlogController.post('/monitor/operlog/export',
                        dependencies=[Depends(require_perm('monitor:operlog:export'))])
@log_decorator(title='操作日志', business_type=BusinessType.EXPORT)
async def operlog_export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysOperLog)
        if params.get('title'):
            query = query.where(SysOperLog.title.like(f"%{params['title']}%"))
        if params.get('operName'):
            query = query.where(SysOperLog.oper_name.like(f"%{params['operName']}%"))
        if params.get('businessType') not in (None, ''):
            query = query.where(SysOperLog.business_type == int(params['businessType']))
        if params.get('status') not in (None, ''):
            query = query.where(SysOperLog.status == int(params['status']))
        logs = (await query_db.execute(
            query.order_by(SysOperLog.oper_id.desc()))).scalars().all()
        columns = [
            ExcelColumn('操作序号', 'operId', width=10),
            ExcelColumn('操作模块', 'title'),
            ExcelColumn('业务类型', 'businessType', dict_type=BUSINESS_TYPE_MAP),
            ExcelColumn('请求方式', 'requestMethod', width=10),
            ExcelColumn('操作类别', 'operatorType', dict_type=OPERATOR_TYPE_MAP),
            ExcelColumn('操作人员', 'operName'),
            ExcelColumn('请求地址', 'operUrl'),
            ExcelColumn('操作地点', 'operLocation'),
            ExcelColumn('状态', 'status', dict_type=STATUS_MAP),
            ExcelColumn('错误消息', 'errorMsg'),
            ExcelColumn('操作时间', 'operTime', width=30),
            ExcelColumn('消耗时间', 'costTime', dict_type=None),
        ]
        return export_excel('操作日志', columns, transform_result(logs), '操作日志')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@logininforController.post('/monitor/logininfor/export',
                           dependencies=[Depends(require_perm('monitor:logininfor:export'))])
@log_decorator(title='登录日志', business_type=BusinessType.EXPORT)
async def logininfor_export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysLogininfor)
        if params.get('userName'):
            query = query.where(SysLogininfor.user_name.like(f"%{params['userName']}%"))
        if params.get('ipaddr'):
            query = query.where(SysLogininfor.ipaddr.like(f"%{params['ipaddr']}%"))
        if params.get('status'):
            query = query.where(SysLogininfor.status == params['status'])
        logs = (await query_db.execute(
            query.order_by(SysLogininfor.info_id.desc()))).scalars().all()
        columns = [
            ExcelColumn('序号', 'infoId', width=10),
            ExcelColumn('用户账号', 'userName'),
            ExcelColumn('登录状态', 'status', dict_type=LOGIN_STATUS_MAP),
            ExcelColumn('登录地址', 'ipaddr'),
            ExcelColumn('登录地点', 'loginLocation'),
            ExcelColumn('浏览器', 'browser'),
            ExcelColumn('操作系统', 'os'),
            ExcelColumn('提示消息', 'msg'),
            ExcelColumn('访问时间', 'loginTime', width=30),
        ]
        return export_excel('登录日志', columns, transform_result(logs), '登录日志')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
