"""
代码生成器服务层（spec-10，对齐java版GenTableServiceImpl + GenUtils + VelocityUtils）

负责：库表读取、列属性推断、gen_table维护、模板渲染上下文
"""
import json
import re
from datetime import datetime
from typing import List, Optional, Tuple
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.entity.do.gen_do import GenTable, GenTableColumn
from utils.log_util import logger

# ============ 配置（对齐java generator.yml） ============
GEN_CONFIG = {
    'author': 'ruoyi',
    'packageName': 'module_admin',       # python侧生成到module_admin下
    'autoRemovePre': False,
    'tablePrefix': 'sys_',
    'allowOverwrite': False,
}

# ============ 列类型分类（对齐java GenConstants） ============
COLUMNTYPE_STR = {'char', 'varchar', 'nvarchar', 'varchar2'}
COLUMNTYPE_TEXT = {'tinytext', 'text', 'mediumtext', 'longtext'}
COLUMNTYPE_TIME = {'datetime', 'time', 'date', 'timestamp'}
COLUMNTYPE_NUMBER = {'tinyint', 'smallint', 'mediumint', 'int', 'number', 'integer',
                     'bigint', 'float', 'float4', 'float8', 'double', 'decimal', 'numeric'}

COLUMNNAME_NOT_EDIT = {'id', 'create_by', 'create_time', 'del_flag'}
COLUMNNAME_NOT_LIST = {'id', 'create_by', 'create_time', 'del_flag', 'update_by', 'update_time'}
COLUMNNAME_NOT_QUERY = {'id', 'create_by', 'create_time', 'del_flag', 'update_by', 'update_time', 'remark'}

HTML_INPUT = 'input'
HTML_TEXTAREA = 'textarea'
HTML_SELECT = 'select'
HTML_RADIO = 'radio'
HTML_DATETIME = 'datetime'

# python类型推断（对位java TYPE_STRING/INTEGER/LONG/DOUBLE/BIGDECIMAL/DATE）
PY_TYPE_STR = 'str'
PY_TYPE_INT = 'int'
PY_TYPE_FLOAT = 'float'
PY_TYPE_DATE = 'datetime'

TPL_CRUD = 'crud'
TPL_TREE = 'tree'


def to_camel_case(name: str) -> str:
    """列名转驼峰（对齐java StringUtils.toCamelCase）：user_name -> userName"""
    parts = name.split('_')
    return parts[0] + ''.join(p.title() for p in parts[1:] if p)


def get_db_type(column_type: str) -> str:
    """取列基础类型（varchar(50) -> varchar，对齐java getDbType）"""
    return re.split(r'\(', column_type or '')[0].lower()


def get_column_length(column_type: str) -> int:
    """取列长度（varchar(50) -> 50）"""
    m = re.search(r'\((\d+)\)', column_type or '')
    return int(m.group(1)) if m else 0


def convert_class_name(table_name: str) -> str:
    """
    表名转类名（对齐java convertClassName：去表前缀 + 下划线转驼峰大写）
    sys_user -> SysUser（tablePrefix=sys_ + autoRemovePre=False时java保留前缀转驼峰，
    实测java默认生成 SysUser——因为GenUtils在initTable时先removePrefix再转）
    """
    prefix = GEN_CONFIG.get('tablePrefix', '')
    name = table_name
    if prefix and name.startswith(prefix):
        name = name[len(prefix):]
    return ''.join(p.title() for p in name.split('_') if p)


def get_business_name(table_name: str) -> str:
    """业务名（表名去前缀后的最后一段，对齐java getBusinessName）"""
    last = table_name.rsplit('_', 1)[-1] if '_' in table_name else table_name
    if GEN_CONFIG.get('autoRemovePre'):
        prefix = GEN_CONFIG.get('tablePrefix', '')
        if table_name.startswith(prefix):
            last = table_name[len(prefix):]
    return last


def replace_comment(comment: str) -> str:
    """清理表/列注释（去换行，对齐java replaceText）"""
    return re.sub(r'\r|\n|\t', '', comment or '')


async def get_db_tables(db: AsyncSession, table_name: str = '', table_comment: str = '',
                        include_imported: bool = False) -> List[dict]:
    """
    读information_schema获取库表列表（对齐java selectDbTableList）
    默认排除qrtz_/gen_前缀与已导入表
    """
    sql = ("select table_name, table_comment, create_time, update_time "
           "from information_schema.tables where table_schema = database() "
           "AND table_name NOT LIKE 'qrtz\\_%' AND table_name NOT LIKE 'gen\\_%'")
    if not include_imported:
        sql += " AND table_name NOT IN (select table_name from gen_table)"
    params = {}
    if table_name:
        sql += " AND lower(table_name) like :tn"
        params['tn'] = f'%{table_name.lower()}%'
    if table_comment:
        sql += " AND lower(table_comment) like :tc"
        params['tc'] = f'%{table_comment.lower()}%'
    sql += " order by create_time desc"
    rows = (await db.execute(text(sql), params)).all()
    return [
        {
            'tableName': r[0], 'tableComment': r[1] or '',
            'createTime': r[2], 'updateTime': r[3],
        } for r in rows
    ]


async def get_db_columns(db: AsyncSession, table_name: str) -> List[dict]:
    """
    读information_schema.columns获取表的列信息
    """
    rows = (await db.execute(text(
        "select column_name, data_type, column_comment, column_key, extra, is_nullable, column_type "
        "from information_schema.columns "
        "where table_schema = database() and table_name = :tn "
        "order by ordinal_position"), {'tn': table_name})).all()
    return [
        {
            'columnName': r[0], 'dataType': r[1], 'columnComment': r[2] or '',
            'columnKey': r[3], 'extra': r[4] or '', 'nullable': r[5], 'columnType': r[6],
        } for r in rows
    ]


def init_column_field(col_meta: dict, table_id: int, create_by: str) -> GenTableColumn:
    """
    初始化列属性（对齐java GenUtils.initColumnField：类型推断/默认勾选规则）
    """
    column_name = col_meta['columnName']
    data_type = get_db_type(col_meta.get('columnType') or col_meta['dataType'])
    is_pk = '1' if col_meta.get('columnKey') == 'PRI' else '0'
    is_increment = '1' if 'auto_increment' in (col_meta.get('extra') or '') else '0'

    col = GenTableColumn(
        table_id=table_id,
        column_name=column_name,
        column_comment=replace_comment(col_meta.get('columnComment')),
        column_type=col_meta.get('columnType') or col_meta['dataType'],
        java_field=to_camel_case(column_name),
        java_type=PY_TYPE_STR,
        is_pk=is_pk,
        is_increment=is_increment,
        is_required='0',
        query_type='EQ',
        sort=col_meta.get('_sort', 0),
        create_by=create_by,
    )

    # python类型与html控件推断
    if data_type in COLUMNTYPE_STR or data_type in COLUMNTYPE_TEXT:
        length = get_column_length(col_meta.get('columnType') or '')
        col.html_type = HTML_TEXTAREA if (length >= 500 or data_type in COLUMNTYPE_TEXT) else HTML_INPUT
        col.java_type = PY_TYPE_STR
    elif data_type in COLUMNTYPE_TIME:
        col.java_type = PY_TYPE_DATE
        col.html_type = HTML_DATETIME
    elif data_type in COLUMNTYPE_NUMBER:
        col.html_type = HTML_INPUT
        # 带小数 -> float；整型 -> int
        m = re.search(r'\((\d+)(?:,(\d+))?\)', col_meta.get('columnType') or '')
        if m and m.group(2) and int(m.group(2)) > 0:
            col.java_type = PY_TYPE_FLOAT
        else:
            col.java_type = PY_TYPE_INT
        # dict下拉的数字字段通常是radio（java在options里处理，这里保持input默认）

    # 默认勾选（对齐java：全部可插入；非主键且不在排除名单的可编辑/列表/查询）
    col.is_insert = '1'
    col.is_edit = '1' if column_name.lower() not in COLUMNNAME_NOT_EDIT and is_pk != '1' else '0'
    col.is_list = '1' if column_name.lower() not in COLUMNNAME_NOT_LIST and is_pk != '1' else '0'
    col.is_query = '1' if column_name.lower() not in COLUMNNAME_NOT_QUERY and is_pk != '1' else '0'
    # 状态/类型等字段默认radio（java GenUtils同逻辑：以name结尾的类型字段）
    if (column_name.endswith('status') or column_name.endswith('type')) and col.java_type == PY_TYPE_STR:
        col.html_type = HTML_RADIO
    # 名称类字段模糊查询（java同逻辑）
    if column_name.endswith('name') and col.html_type == HTML_INPUT:
        col.query_type = 'LIKE'

    col.is_required = '1' if col_meta.get('nullable') == 'NO' and is_pk != '1' and \
        not str(column_name).endswith('_id') else '0'
    return col


def init_table_meta(db_table: dict, oper_name: str) -> GenTable:
    """
    初始化表元数据（对齐java GenUtils.initTable）
    """
    return GenTable(
        table_name=db_table['tableName'],
        table_comment=replace_comment(db_table.get('tableComment')),
        class_name=convert_class_name(db_table['tableName']),
        package_name=GEN_CONFIG['packageName'],
        module_name=get_module_name(),
        business_name=get_business_name(db_table['tableName']),
        function_name=replace_comment(db_table.get('tableComment')) or db_table['tableName'],
        function_author=GEN_CONFIG['author'],
        tpl_category=TPL_CRUD,
        tpl_web_type='element-plus',
        gen_type='0',
        gen_path='/',
        create_by=oper_name,
        create_time=datetime.now(),
    )


def get_module_name() -> str:
    """模块名（java取packageName第一段后的第二段；python取固定业务域）"""
    return GEN_CONFIG['packageName'].split('.')[-1] if '.' in GEN_CONFIG['packageName'] \
        else GEN_CONFIG['packageName']


async def get_gen_table(db: AsyncSession, table_id: int) -> Optional[GenTable]:
    return (await db.execute(
        select(GenTable).where(GenTable.table_id == table_id))).scalars().first()


async def get_gen_table_by_name(db: AsyncSession, table_name: str) -> Optional[GenTable]:
    return (await db.execute(
        select(GenTable).where(GenTable.table_name == table_name))).scalars().first()


async def get_gen_columns(db: AsyncSession, table_id: int) -> List[GenTableColumn]:
    return list((await db.execute(
        select(GenTableColumn).where(GenTableColumn.table_id == table_id)
        .order_by(GenTableColumn.sort))).scalars().all())


async def get_all_gen_tables(db: AsyncSession) -> List[GenTable]:
    return list((await db.execute(
        select(GenTable).order_by(GenTable.table_id))).scalars().all())


def get_pk_column(columns: List[GenTableColumn]) -> Optional[GenTableColumn]:
    for c in columns:
        if c.is_pk == '1':
            return c
    return columns[0] if columns else None


def parse_options(table: GenTable) -> dict:
    """
    解析options JSON（treeCode/treeParentCode/treeName等生成选项）
    """
    try:
        return json.loads(table.options) if table.options else {}
    except (ValueError, TypeError):
        return {}


def get_permission_prefix(table: GenTable) -> str:
    """权限前缀（对齐java getPermissionPrefix：moduleName:businessName）"""
    return f"{table.module_name}:{table.business_name}"


def build_template_context(table: GenTable, columns: List[GenTableColumn]) -> dict:
    """
    组装模板渲染上下文（对齐java VelocityUtils.prepareContext的变量集）
    """
    pk = get_pk_column(columns)
    options = parse_options(table)
    return {
        'tplCategory': table.tpl_category,
        'tableName': table.table_name,
        'functionName': table.function_name or '【请填写功能名称】',
        'ClassName': table.class_name,
        'className': (table.class_name[:1].lower() + table.class_name[1:]) if table.class_name else '',
        'moduleName': table.module_name,
        'BusinessName': (table.business_name[:1].upper() + table.business_name[1:]) if table.business_name else '',
        'businessName': table.business_name,
        'packageName': table.package_name,
        'author': table.function_author,
        'datetime': datetime.now().strftime('%Y-%m-%d'),
        'pkColumn': pk,
        'columns': columns,
        'table': table,
        'permissionPrefix': get_permission_prefix(table),
        'dicts': [c.dict_type for c in columns if c.dict_type],
        'genView': False,
        'treeCode': options.get('treeCode', ''),
        'treeParentCode': options.get('treeParentCode', ''),
        'treeName': options.get('treeName', ''),
        'formColNum': table.form_col_num or 1,
    }
