"""
字典与参数管理控制器（spec-06，对应java版SysDictTypeController/SysDictDataController/SysConfigController）

缓存策略（对齐java版）：启动预热(server.py) + CRUD增量维护；
key: sys_dict:{dictType} / sys_config:{configKey}
"""
from datetime import datetime
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from config.redis_cache import RedisCache
from module_admin.entity.do.entity import SysDictType, SysDictData, SysConfig
from module_admin.aspect.interface_auth import require_perm
from module_admin.annotation.log_annotation import log_decorator
from common.constant import CacheConstants
from common.enums import BusinessType
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from utils.excel_util import export_excel, ExcelColumn
from utils.log_util import logger

dictTypeController = APIRouter()
dictDataController = APIRouter()
configController = APIRouter()


# ==================== 字典类型 ====================

@dictTypeController.get('/system/dict/type/list',
                        dependencies=[Depends(require_perm('system:dict:list'))])
async def dict_type_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysDictType)
        if params.get('dictName'):
            query = query.where(SysDictType.dict_name.like(f"%{params['dictName']}%"))
        if params.get('dictType'):
            query = query.where(SysDictType.dict_type.like(f"%{params['dictType']}%"))
        if params.get('status'):
            query = query.where(SysDictType.status == params['status'])
        if params.get('beginTime'):
            query = query.where(SysDictType.create_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysDictType.create_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query.order_by(SysDictType.dict_id), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictDataController.post('/system/dict/data/export',
                         dependencies=[Depends(require_perm('system:dict:export'))])
@log_decorator(title='字典数据', business_type=BusinessType.EXPORT)
async def dict_data_export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysDictData)
        if params.get('dictType'):
            query = query.where(SysDictData.dict_type == params['dictType'])
        datas = (await query_db.execute(
            query.order_by(SysDictData.dict_type, SysDictData.dict_sort))).scalars().all()
        columns = [ExcelColumn('字典编码', 'dictCode', width=10),
                   ExcelColumn('字典排序', 'dictSort', width=10),
                   ExcelColumn('字典标签', 'dictLabel'),
                   ExcelColumn('字典键值', 'dictValue'),
                   ExcelColumn('字典类型', 'dictType'),
                   ExcelColumn('是否默认', 'isDefault', dict_type={'Y': '是', 'N': '否'}),
                   ExcelColumn('状态', 'status', dict_type={'0': '正常', '1': '停用'})]
        return export_excel('字典数据', columns, transform_result(datas), '字典数据')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.post('/system/dict/type/export',
                         dependencies=[Depends(require_perm('system:dict:export'))])
@log_decorator(title='字典类型', business_type=BusinessType.EXPORT)
async def dict_type_export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        types = (await query_db.execute(select(SysDictType))).scalars().all()
        columns = [ExcelColumn('字典主键', 'dictId', width=10),
                   ExcelColumn('字典名称', 'dictName'),
                   ExcelColumn('字典类型', 'dictType'),
                   ExcelColumn('状态', 'status', dict_type={'0': '正常', '1': '停用'})]
        return export_excel('字典类型', columns, transform_result(types), '字典类型')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


async def _check_dict_type_unique(db, dict_type: str, dict_id=None) -> bool:
    query = select(SysDictType.dict_id).where(SysDictType.dict_type == dict_type)
    if dict_id:
        query = query.where(SysDictType.dict_id != dict_id)
    return (await db.execute(query.limit(1))).first() is not None


@dictTypeController.post('/system/dict/type',
                         dependencies=[Depends(require_perm('system:dict:add'))])
@log_decorator(title='字典类型', business_type=BusinessType.INSERT)
async def add_dict_type(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        if await _check_dict_type_unique(query_db, body.get('dictType', '')):
            return ResponseUtil.failure(msg=f"新增字典'{body.get('dictName')}'失败，字典类型已存在")
        dt = SysDictType(
            dict_name=body.get('dictName', ''),
            dict_type=body.get('dictType', ''),
            status=body.get('status', '0'),
            remark=body.get('remark'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(dt)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.put('/system/dict/type',
                        dependencies=[Depends(require_perm('system:dict:edit'))])
@log_decorator(title='字典类型', business_type=BusinessType.UPDATE)
async def edit_dict_type(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        dict_id = body.get('dictId')
        dt = (await query_db.execute(
            select(SysDictType).where(SysDictType.dict_id == dict_id))).scalars().first()
        if not dt:
            return ResponseUtil.failure(msg='字典类型不存在')
        if await _check_dict_type_unique(query_db, body.get('dictType', ''), dict_id):
            return ResponseUtil.failure(msg=f"修改字典'{body.get('dictName')}'失败，字典类型已存在")
        old_type = dt.dict_type
        dt.dict_name = body.get('dictName', dt.dict_name)
        dt.dict_type = body.get('dictType', dt.dict_type)
        dt.status = body.get('status', dt.status)
        dt.remark = body.get('remark', dt.remark)
        dt.update_by = login_user.get('user_name', '')
        dt.update_time = datetime.now()
        # 级联更新字典数据的dict_type
        if old_type != dt.dict_type:
            await query_db.execute(
                SysDictData.__table__.update()
                .where(SysDictData.dict_type == old_type)
                .values(dict_type=dt.dict_type))
            # 清旧缓存
            cache: RedisCache = request.app.state.redis_cache
            await cache.delete_object(
                cache.build_key(CacheConstants.SYS_DICT_KEY, old_type))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.delete('/system/dict/type/{dict_ids}',
                           dependencies=[Depends(require_perm('system:dict:remove'))])
@log_decorator(title='字典类型', business_type=BusinessType.DELETE)
async def remove_dict_type(request: Request, dict_ids: str,
                           query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        failed = []
        for did in [int(x) for x in dict_ids.split(',') if x]:
            dt = (await query_db.execute(
                select(SysDictType).where(SysDictType.dict_id == did))).scalars().first()
            if not dt:
                continue
            used = (await query_db.execute(
                select(SysDictData.dict_code).where(SysDictData.dict_type == dt.dict_type)
                .limit(1))).first()
            if used:
                failed.append(dt.dict_name)
                continue
            await query_db.delete(dt)
            await cache.delete_object(
                cache.build_key(CacheConstants.SYS_DICT_KEY, dt.dict_type))
        await query_db.commit()
        if failed:
            return ResponseUtil.failure(msg=f"{','.join(failed)}已分配,不能删除")
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.delete('/system/dict/type/refreshCache',
                           dependencies=[Depends(require_perm('system:dict:remove'))])
@log_decorator(title='字典类型', business_type=BusinessType.CLEAN)
async def refresh_dict_cache(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        # 全量重建：清空所有 sys_dict:* 再重载
        for key in await cache.keys_by_prefix(f'{CacheConstants.SYS_DICT_KEY}*'):
            await cache.delete_object(key)
        datas = (await query_db.execute(
            select(SysDictData).where(SysDictData.status == '0')
            .order_by(SysDictData.dict_sort))).scalars().all()
        grouped = {}
        for d in datas:
            grouped.setdefault(d.dict_type, []).append(d)
        for dtype, items in grouped.items():
            await cache.set_cache_object(
                cache.build_key(CacheConstants.SYS_DICT_KEY, dtype),
                transform_result(items))
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.get('/system/dict/type/optionselect')
async def dict_type_optionselect(query_db: AsyncSession = Depends(get_db)):
    try:
        types = (await query_db.execute(select(SysDictType))).scalars().all()
        return ResponseUtil.success(data=transform_result(types))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictTypeController.get('/system/dict/type/{dict_id}',
                        dependencies=[Depends(require_perm('system:dict:query'))])
async def get_dict_type(request: Request, dict_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        dt = (await query_db.execute(
            select(SysDictType).where(SysDictType.dict_id == dict_id))).scalars().first()
        return ResponseUtil.success(dict_content={'data': transform_result(dt) if dt else None})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 字典数据 ====================

@dictDataController.get('/system/dict/data/list',
                        dependencies=[Depends(require_perm('system:dict:list'))])
async def dict_data_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysDictData)
        if params.get('dictType'):
            query = query.where(SysDictData.dict_type == params['dictType'])
        if params.get('dictLabel'):
            query = query.where(SysDictData.dict_label.like(f"%{params['dictLabel']}%"))
        if params.get('status'):
            query = query.where(SysDictData.status == params['status'])
        return await paginate(query_db, query.order_by(SysDictData.dict_sort), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictDataController.get('/system/dict/data/type/{dict_type}')
async def dict_data_by_type(request: Request, dict_type: str,
                            query_db: AsyncSession = Depends(get_db)):
    """
    按类型取字典数据（前端useDict高频接口；redis缓存优先，miss回源并写缓存）
    """
    try:
        cache: RedisCache = request.app.state.redis_cache
        cache_key = cache.build_key(CacheConstants.SYS_DICT_KEY, dict_type)
        cached = await cache.get_cache_object(cache_key)
        if cached:
            return ResponseUtil.success(data=cached)
        datas = (await query_db.execute(
            select(SysDictData).where(SysDictData.dict_type == dict_type,
                                      SysDictData.status == '0')
            .order_by(SysDictData.dict_sort))).scalars().all()
        data = transform_result(datas)
        await cache.set_cache_object(cache_key, data)
        return ResponseUtil.success(data=data)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictDataController.post('/system/dict/data',
                         dependencies=[Depends(require_perm('system:dict:add'))])
@log_decorator(title='字典数据', business_type=BusinessType.INSERT)
async def add_dict_data(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        dd = SysDictData(
            dict_sort=body.get('dictSort', 0),
            dict_label=body.get('dictLabel', ''),
            dict_value=body.get('dictValue', ''),
            dict_type=body.get('dictType', ''),
            css_class=body.get('cssClass'),
            list_class=body.get('listClass'),
            is_default=body.get('isDefault', 'N'),
            status=body.get('status', '0'),
            remark=body.get('remark'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(dd)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictDataController.put('/system/dict/data',
                        dependencies=[Depends(require_perm('system:dict:edit'))])
@log_decorator(title='字典数据', business_type=BusinessType.UPDATE)
async def edit_dict_data(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        code = body.get('dictCode')
        dd = (await query_db.execute(
            select(SysDictData).where(SysDictData.dict_code == code))).scalars().first()
        if not dd:
            return ResponseUtil.failure(msg='字典数据不存在')
        dd.dict_sort = body.get('dictSort', dd.dict_sort)
        dd.dict_label = body.get('dictLabel', dd.dict_label)
        dd.dict_value = body.get('dictValue', dd.dict_value)
        dd.dict_type = body.get('dictType', dd.dict_type)
        dd.css_class = body.get('cssClass', dd.css_class)
        dd.list_class = body.get('listClass', dd.list_class)
        dd.is_default = body.get('isDefault', dd.is_default)
        dd.status = body.get('status', dd.status)
        dd.remark = body.get('remark', dd.remark)
        dd.update_by = login_user.get('user_name', '')
        dd.update_time = datetime.now()
        await query_db.commit()
        # 刷新该类型缓存
        cache: RedisCache = request.app.state.redis_cache
        datas = (await query_db.execute(
            select(SysDictData).where(SysDictData.dict_type == dd.dict_type,
                                      SysDictData.status == '0')
            .order_by(SysDictData.dict_sort))).scalars().all()
        await cache.set_cache_object(
            cache.build_key(CacheConstants.SYS_DICT_KEY, dd.dict_type),
            transform_result(datas))
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@dictDataController.delete('/system/dict/data/{dict_codes}',
                           dependencies=[Depends(require_perm('system:dict:remove'))])
@log_decorator(title='字典数据', business_type=BusinessType.DELETE)
async def remove_dict_data(request: Request, dict_codes: str,
                           query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        touched_types = set()
        for code in [int(x) for x in dict_codes.split(',') if x]:
            dd = (await query_db.execute(
                select(SysDictData).where(SysDictData.dict_code == code))).scalars().first()
            if dd:
                touched_types.add(dd.dict_type)
                await query_db.delete(dd)
        await query_db.commit()
        for dtype in touched_types:
            datas = (await query_db.execute(
                select(SysDictData).where(SysDictData.dict_type == dtype,
                                          SysDictData.status == '0')
                .order_by(SysDictData.dict_sort))).scalars().all()
            await cache.set_cache_object(
                cache.build_key(CacheConstants.SYS_DICT_KEY, dtype),
                transform_result(datas))
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


# ==================== 参数配置 ====================

@configController.post('/system/config/export',
                       dependencies=[Depends(require_perm('system:config:export'))])
@log_decorator(title='参数管理', business_type=BusinessType.EXPORT)
async def config_export(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        configs = (await query_db.execute(
            select(SysConfig).order_by(SysConfig.config_id))).scalars().all()
        columns = [ExcelColumn('参数主键', 'configId', width=10),
                   ExcelColumn('参数名称', 'configName'),
                   ExcelColumn('参数键名', 'configKey'),
                   ExcelColumn('参数键值', 'configValue'),
                   ExcelColumn('系统内置', 'configType', dict_type={'Y': '是', 'N': '否'})]
        return export_excel('参数数据', columns, transform_result(configs), '参数数据')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.get('/system/config/list',
                      dependencies=[Depends(require_perm('system:config:list'))])
async def config_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(SysConfig)
        if params.get('configName'):
            query = query.where(SysConfig.config_name.like(f"%{params['configName']}%"))
        if params.get('configKey'):
            query = query.where(SysConfig.config_key.like(f"%{params['configKey']}%"))
        if params.get('configType'):
            query = query.where(SysConfig.config_type == params['configType'])
        if params.get('beginTime'):
            query = query.where(SysConfig.create_time >= params['beginTime'])
        if params.get('endTime'):
            query = query.where(SysConfig.create_time <= params['endTime'] + ' 23:59:59')
        return await paginate(query_db, query.order_by(SysConfig.config_id), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.get('/system/config/configKey/{config_key}')
async def get_config_key(request: Request, config_key: str,
                         query_db: AsyncSession = Depends(get_db)):
    """
    按键名取参数值（redis缓存优先，miss回源并写缓存）
    """
    try:
        cache: RedisCache = request.app.state.redis_cache
        cache_key = cache.build_key(CacheConstants.SYS_CONFIG_KEY, config_key)
        cached = await cache.get_cache_object(cache_key)
        if cached is not None:
            return ResponseUtil.success(dict_content={'msg': '操作成功', 'configValue': cached})
        config = (await query_db.execute(
            select(SysConfig).where(SysConfig.config_key == config_key))).scalars().first()
        if config:
            await cache.set_cache_object(cache_key, config.config_value)
            return ResponseUtil.success(dict_content={'configValue': config.config_value})
        return ResponseUtil.failure(msg='参数不存在')
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.post('/system/config', dependencies=[Depends(require_perm('system:config:add'))])
@log_decorator(title='参数管理', business_type=BusinessType.INSERT)
async def add_config(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        key = body.get('configKey', '')
        dup = (await query_db.execute(
            select(SysConfig.config_id).where(SysConfig.config_key == key).limit(1))).first()
        if dup:
            return ResponseUtil.failure(msg=f"新增参数'{body.get('configName')}'失败，参数键名已存在")
        config = SysConfig(
            config_name=body.get('configName', ''),
            config_key=key,
            config_value=body.get('configValue', ''),
            config_type=body.get('configType', 'N'),
            remark=body.get('remark'),
            create_by=login_user.get('user_name', ''),
            create_time=datetime.now(),
        )
        query_db.add(config)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.put('/system/config', dependencies=[Depends(require_perm('system:config:edit'))])
@log_decorator(title='参数管理', business_type=BusinessType.UPDATE)
async def edit_config(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        config_id = body.get('configId')
        config = (await query_db.execute(
            select(SysConfig).where(SysConfig.config_id == config_id))).scalars().first()
        if not config:
            return ResponseUtil.failure(msg='参数不存在')
        dup = (await query_db.execute(
            select(SysConfig.config_id).where(SysConfig.config_key == body.get('configKey', ''),
                                              SysConfig.config_id != config_id).limit(1))).first()
        if dup:
            return ResponseUtil.failure(msg=f"修改参数'{body.get('configName')}'失败，参数键名已存在")
        config.config_name = body.get('configName', config.config_name)
        config.config_key = body.get('configKey', config.config_key)
        config.config_value = body.get('configValue', config.config_value)
        config.config_type = body.get('configType', config.config_type)
        config.remark = body.get('remark', config.remark)
        config.update_by = login_user.get('user_name', '')
        config.update_time = datetime.now()
        await query_db.commit()
        if config.config_type == 'Y':
            cache: RedisCache = request.app.state.redis_cache
            await cache.set_cache_object(
                cache.build_key(CacheConstants.SYS_CONFIG_KEY, config.config_key),
                config.config_value)
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.delete('/system/config/{config_ids}',
                         dependencies=[Depends(require_perm('system:config:remove'))])
@log_decorator(title='参数管理', business_type=BusinessType.DELETE)
async def remove_config(request: Request, config_ids: str,
                        query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        for cid in [int(x) for x in config_ids.split(',') if x]:
            config = (await query_db.execute(
                select(SysConfig).where(SysConfig.config_id == cid))).scalars().first()
            if not config:
                continue
            if config.config_type == 'Y':
                return ResponseUtil.failure(msg=f'内置参数{config.config_key}不能删除')
            await query_db.delete(config)
            await cache.delete_object(
                cache.build_key(CacheConstants.SYS_CONFIG_KEY, config.config_key))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@configController.delete('/system/config/refreshCache',
                         dependencies=[Depends(require_perm('system:config:remove'))])
@log_decorator(title='参数管理', business_type=BusinessType.CLEAN)
async def refresh_config_cache(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        cache: RedisCache = request.app.state.redis_cache
        for key in await cache.keys_by_prefix(f'{CacheConstants.SYS_CONFIG_KEY}*'):
            await cache.delete_object(key)
        configs = (await query_db.execute(select(SysConfig))).scalars().all()
        for c in configs:
            await cache.set_cache_object(
                cache.build_key(CacheConstants.SYS_CONFIG_KEY, c.config_key),
                c.config_value)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
