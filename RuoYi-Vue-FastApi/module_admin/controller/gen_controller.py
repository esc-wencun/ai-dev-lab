"""
代码生成控制器（spec-10，对齐java版GenController）
"""
import re
from fastapi import APIRouter, Request, Depends
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from module_admin.entity.do.gen_do import GenTable, GenTableColumn
from module_admin.aspect.interface_auth import require_perm, validate_role
from module_admin.annotation.log_annotation import log_decorator
from common.enums import BusinessType
from utils.page_util import paginate
from utils.response_util import ResponseUtil
from utils.common_util import transform_result
from module_admin.service import gen_service, gen_render
from utils.log_util import logger

genController = APIRouter()


@genController.get('/tool/gen/list', dependencies=[Depends(require_perm('tool:gen:list'))])
async def gen_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        query = select(GenTable)
        if params.get('tableName'):
            query = query.where(GenTable.table_name.like(f"%{params['tableName']}%"))
        if params.get('tableComment'):
            query = query.where(GenTable.table_comment.like(f"%{params['tableComment']}%"))
        return await paginate(query_db, query.order_by(GenTable.table_id.desc()), request)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/db/list', dependencies=[Depends(require_perm('tool:gen:list'))])
async def db_list(request: Request, query_db: AsyncSession = Depends(get_db)):
    try:
        params = request.query_params
        tables = await gen_service.get_db_tables(
            query_db, params.get('tableName', ''), params.get('tableComment', ''))
        # 前端分页在本地做（java是PageHelper，这里数据量小手动分页保持rows/total结构）
        from utils.page_util import get_page_domain
        domain = get_page_domain(request)
        start = domain.offset
        page_rows = tables[start:start + domain.page_size]
        return ResponseUtil.success(msg='查询成功',
                                    dict_content={'rows': transform_result(page_rows),
                                                  'total': len(tables)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/preview/{table_id}',
                   dependencies=[Depends(require_perm('tool:gen:preview'))])
async def preview(request: Request, table_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        table = await gen_service.get_gen_table(query_db, table_id)
        if not table:
            return ResponseUtil.failure(msg='生成数据不存在')
        columns = await gen_service.get_gen_columns(query_db, table_id)
        rendered = gen_render.render_table(table, columns)
        # 前端tabs的key用文件名（java用模板名，这里用生成物文件名更直观）
        result = {}
        for template, content in rendered.items():
            result[gen_render.get_file_name(template, table)] = content
        return ResponseUtil.success(data=result)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/{table_id}', dependencies=[Depends(require_perm('tool:gen:query'))])
async def get_info(request: Request, table_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        table = await gen_service.get_gen_table(query_db, table_id)
        tables = await gen_service.get_all_gen_tables(query_db)
        columns = await gen_service.get_gen_columns(query_db, table_id)
        return ResponseUtil.success(dict_content={
            'info': transform_result(table) if table else None,
            'rows': transform_result(columns),
            'tables': transform_result(tables),
        })
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/column/{table_id}',
                   dependencies=[Depends(require_perm('tool:gen:list'))])
async def column_list(request: Request, table_id: int, query_db: AsyncSession = Depends(get_db)):
    try:
        columns = await gen_service.get_gen_columns(query_db, table_id)
        return ResponseUtil.success(msg='查询成功',
                                    dict_content={'rows': transform_result(columns),
                                                  'total': len(columns)})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.post('/tool/gen/importTable', dependencies=[Depends(require_perm('tool:gen:import'))])
@log_decorator(title='代码生成', business_type=BusinessType.IMPORT)
async def import_table(request: Request, tables: str = '', tplWebType: str = 'element-plus',
                       query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        oper_name = login_user.get('user_name', '')
        imported = 0
        for table_name in [t.strip() for t in tables.split(',') if t.strip()]:
            db_tables = await gen_service.get_db_tables(query_db, table_name=table_name,
                                                         include_imported=True)
            db_table = next((t for t in db_tables if t['tableName'] == table_name), None)
            if not db_table:
                continue
            table_meta = gen_service.init_table_meta(db_table, oper_name)
            table_meta.tpl_web_type = tplWebType
            query_db.add(table_meta)
            await query_db.flush()
            # 读取并初始化列
            col_metas = await gen_service.get_db_columns(query_db, table_name)
            for sort, col_meta in enumerate(col_metas, start=1):
                col_meta['_sort'] = sort
                col = gen_service.init_column_field(col_meta, table_meta.table_id, oper_name)
                query_db.add(col)
            imported += 1
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.post('/tool/gen/createTable')
@log_decorator(title='创建表', business_type=BusinessType.OTHER)
async def create_table(request: Request, sql: str = '', tplWebType: str = 'element-plus',
                       query_db: AsyncSession = Depends(get_db)):
    """
    根据建表SQL创建表并导入（仅admin；SQL安全过滤后逐条执行CREATE TABLE）
    """
    try:
        await validate_role(request, 'admin')
        from utils.sql_util import filter_keyword
        filter_keyword(sql)
        # 仅允许CREATE TABLE语句（对齐java只处理MySqlCreateTableStatement）
        if not re.search(r'\bcreate\s+table\b', sql, re.IGNORECASE):
            return ResponseUtil.failure(msg='创建表结构异常')
        # 提取表名（防注入已过滤关键字；表名用白名单字符校验）
        table_names = re.findall(r'create\s+table\s+(?:if\s+not\s+exists\s+)?`?([\w]+)`?',
                                  sql, re.IGNORECASE)
        if not table_names:
            return ResponseUtil.failure(msg='创建表结构异常')
        statements = [s.strip() for s in re.split(r';\s*(?=\n|$)', sql.strip()) if s.strip()]
        for stmt in statements:
            if re.search(r'\bcreate\s+table\b', stmt, re.IGNORECASE):
                await query_db.execute(text(stmt))
        await query_db.commit()
        # 导入到gen_table
        for table_name in table_names:
            db_tables = await gen_service.get_db_tables(query_db, table_name=table_name,
                                                         include_imported=True)
            db_table = next((t for t in db_tables if t['tableName'] == table_name), None)
            if not db_table:
                continue
            from module_admin.service.login_service import TokenService
            login_user = await TokenService.get_login_user(request) or {}
            table_meta = gen_service.init_table_meta(db_table, login_user.get('user_name', ''))
            table_meta.tpl_web_type = tplWebType
            query_db.add(table_meta)
            await query_db.flush()
            col_metas = await gen_service.get_db_columns(query_db, table_name)
            for sort, col_meta in enumerate(col_metas, start=1):
                col_meta['_sort'] = sort
                query_db.add(gen_service.init_column_field(col_meta, table_meta.table_id,
                                                            table_meta.create_by))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.failure(msg='创建表结构异常')


@genController.put('/tool/gen', dependencies=[Depends(require_perm('tool:gen:edit'))])
@log_decorator(title='代码生成', business_type=BusinessType.UPDATE)
async def edit_save(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        table_id = body.get('tableId')
        table = await gen_service.get_gen_table(query_db, table_id)
        if not table:
            return ResponseUtil.failure(msg='生成数据不存在')
        table.table_name = body.get('tableName', table.table_name)
        table.table_comment = body.get('tableComment', table.table_comment)
        table.class_name = body.get('className', table.class_name)
        table.function_name = body.get('functionName', table.function_name)
        table.function_author = body.get('functionAuthor', table.function_author)
        table.module_name = body.get('moduleName', table.module_name)
        table.business_name = body.get('businessName', table.business_name)
        table.tpl_category = body.get('tplCategory', table.tpl_category)
        table.tpl_web_type = body.get('tplWebType', table.tpl_web_type)
        table.form_col_num = body.get('formColNum', table.form_col_num)
        table.gen_type = body.get('genType', table.gen_type)
        table.gen_path = body.get('genPath', table.gen_path)
        options = body.get('options')
        if isinstance(options, dict):
            import json as _json
            table.options = _json.dumps(options, ensure_ascii=False)
        elif isinstance(options, str):
            table.options = options
        table.remark = body.get('remark', table.remark)
        table.update_by = login_user.get('user_name', '')
        table.update_time = datetime.now()
        # 字段列信息更新
        for col_body in body.get('columns') or []:
            col_id = col_body.get('columnId')
            col = (await query_db.execute(
                select(GenTableColumn).where(GenTableColumn.column_id == col_id))).scalars().first()
            if not col:
                continue
            for field in ('javaType', 'javaField', 'queryType', 'htmlType', 'dictType',
                          'sort', 'isRequired', 'isInsert', 'isEdit', 'isList', 'isQuery'):
                camel = field
                snake = re.sub(r'(?<!^)(?=[A-Z])', '_', field).lower()
                if camel in col_body:
                    setattr(col, snake, col_body[camel])
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.delete('/tool/gen/{table_ids}',
                      dependencies=[Depends(require_perm('tool:gen:remove'))])
@log_decorator(title='代码生成', business_type=BusinessType.DELETE)
async def remove(request: Request, table_ids: str, query_db: AsyncSession = Depends(get_db)):
    try:
        for tid in [int(x) for x in table_ids.split(',') if x]:
            table = await gen_service.get_gen_table(query_db, tid)
            if table:
                await query_db.delete(table)
            await query_db.execute(
                GenTableColumn.__table__.delete().where(GenTableColumn.table_id == tid))
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/genCode/{table_name}',
                   dependencies=[Depends(require_perm('tool:gen:code'))])
@log_decorator(title='代码生成', business_type=BusinessType.GENCODE)
async def gen_code_path(request: Request, table_name: str,
                        query_db: AsyncSession = Depends(get_db)):
    """
    生成代码（自定义路径方式）。python版从配置读取allowOverwrite，默认不允许
    """
    if not gen_service.GEN_CONFIG.get('allowOverwrite'):
        return ResponseUtil.failure(msg='【系统预设】不允许生成文件覆盖到本地')
    table = await gen_service.get_gen_table_by_name(query_db, table_name)
    if not table:
        return ResponseUtil.failure(msg='生成数据不存在')
    columns = await gen_service.get_gen_columns(query_db, table.table_id)
    rendered = gen_render.render_table(table, columns)
    # 写到gen_path（相对项目根）
    written = 0
    for template, content in rendered.items():
        file_name = gen_render.get_file_name(template, table)
        full_path = table.gen_path.rstrip('/') + '/' + file_name if table.gen_path != '/' else file_name
        import os
        os.makedirs(os.path.dirname(full_path), exist_ok=True)
        with open(full_path, 'w', encoding='utf-8') as f:
            f.write(content)
        written += 1
    return ResponseUtil.success(msg=f'成功生成{written}个文件')


@genController.get('/tool/gen/synchDb/{table_name}',
                   dependencies=[Depends(require_perm('tool:gen:edit'))])
@log_decorator(title='代码生成', business_type=BusinessType.UPDATE)
async def synch_db(request: Request, table_name: str, query_db: AsyncSession = Depends(get_db)):
    """
    同步数据库表结构（新列插入、消失列删除；对齐java synchDb）
    """
    try:
        from module_admin.service.login_service import TokenService
        login_user = await TokenService.get_login_user(request) or {}
        table = await gen_service.get_gen_table_by_name(query_db, table_name)
        if not table:
            return ResponseUtil.failure(msg='生成数据不存在')
        col_metas = await gen_service.get_db_columns(query_db, table_name)
        db_columns = {c['columnName']: c for c in col_metas}
        existing = {c.column_name: c for c in await gen_service.get_gen_columns(query_db, table.table_id)}
        # 新列
        sort = len(existing)
        for name, meta in db_columns.items():
            if name not in existing:
                sort += 1
                meta['_sort'] = sort
                query_db.add(gen_service.init_column_field(meta, table.table_id,
                                                            login_user.get('user_name', '')))
        # 消失列
        for name, col in existing.items():
            if name not in db_columns:
                await query_db.delete(col)
        await query_db.commit()
        return ResponseUtil.success()
    except Exception as e:
        await query_db.rollback()
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


from datetime import datetime  # noqa: E402


@genController.get('/tool/gen/download/{table_name}',
                   dependencies=[Depends(require_perm('tool:gen:code'))])
@log_decorator(title='代码生成', business_type=BusinessType.GENCODE)
async def download(request: Request, table_name: str, query_db: AsyncSession = Depends(get_db)):
    try:
        table = await gen_service.get_gen_table_by_name(query_db, table_name)
        if not table:
            return ResponseUtil.failure(msg='生成数据不存在')
        columns = await gen_service.get_gen_columns(query_db, table.table_id)
        data = gen_render.render_zip([(table, columns)])
        import urllib.parse
        filename = urllib.parse.quote('ruoyi.zip')
        from fastapi.responses import Response
        return Response(
            content=data,
            media_type='application/octet-stream; charset=UTF-8',
            headers={'Content-Disposition': f'attachment; filename="{filename}"',
                     'Access-Control-Expose-Headers': 'Content-Disposition'})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@genController.get('/tool/gen/batchGenCode',
                   dependencies=[Depends(require_perm('tool:gen:code'))])
@log_decorator(title='代码生成', business_type=BusinessType.GENCODE)
async def batch_gen_code(request: Request, tables: str = '',
                         query_db: AsyncSession = Depends(get_db)):
    try:
        table_tuples = []
        for name in [t.strip() for t in tables.split(',') if t.strip()]:
            table = await gen_service.get_gen_table_by_name(query_db, name)
            if not table:
                continue
            columns = await gen_service.get_gen_columns(query_db, table.table_id)
            table_tuples.append((table, columns))
        if not table_tuples:
            return ResponseUtil.failure(msg='未找到有效数据')
        data = gen_render.render_zip(table_tuples)
        import urllib.parse
        filename = urllib.parse.quote('ruoyi.zip')
        from fastapi.responses import Response
        return Response(
            content=data,
            media_type='application/octet-stream; charset=UTF-8',
            headers={'Content-Disposition': f'attachment; filename="{filename}"',
                     'Access-Control-Expose-Headers': 'Content-Disposition'})
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))
