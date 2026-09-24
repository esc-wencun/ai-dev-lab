"""
代码生成渲染引擎（spec-10，对齐java版VelocityUtils.getTemplateList/getFileName）

生成物为本项目Python版代码结构 + RuoYi-Vue3前端 + 菜单SQL
"""
import os
import io
import re
import zipfile
from typing import List
from jinja2 import Environment, FileSystemLoader
from module_admin.service.gen_service import (
    build_template_context, get_permission_prefix, get_pk_column, to_camel_case,
)
from module_admin.entity.do.gen_do import GenTable, GenTableColumn

TEMPLATE_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'gen_templates')
# 生成物输出根路径（zip内的顶层目录名）
PROJECT_PATH = 'ruoyi-vue-fastapi'


def _jinja_env():
    env = Environment(
        loader=FileSystemLoader(TEMPLATE_DIR),
        keep_trailing_newline=True,
        trim_blocks=False,
        lstrip_blocks=False,
    )
    # 自定义过滤器：列类型 -> SQLAlchemy Column类型
    def py_column_type(column_type: str) -> str:
        base = re.split(r'\(', (column_type or ''))[0].lower()
        if base in ('char', 'varchar', 'nvarchar', 'varchar2', 'enum'):
            return 'String(255)'
        if base in ('text', 'tinytext', 'mediumtext', 'longtext'):
            return 'Text'
        if base in ('datetime', 'timestamp'):
            return 'DateTime'
        if base == 'date':
            return 'Date'
        if base == 'time':
            return 'Time'
        if base in ('bigint',):
            return 'BigInteger'
        if base in ('int', 'integer', 'smallint', 'mediumint', 'tinyint'):
            return 'Integer'
        if base in ('float', 'double', 'decimal', 'numeric'):
            return 'Float'
        return 'String(255)'
    env.filters['py_column_type'] = py_column_type

    def to_camel(value: str) -> str:
        return to_camel_case(value or '')
    env.filters['to_camel'] = to_camel

    def snake(value: str) -> str:
        """驼峰转下划线"""
        s1 = re.sub(r'(.)([A-Z][a-z]+)', r'\1_\2', value or '')
        return re.sub(r'([a-z0-9])([A-Z])', r'\1_\2', s1).lower()
    env.filters['snake'] = snake
    return env


def get_template_list(table: GenTable) -> List[str]:
    """
    模板清单（对齐java getTemplateList：固定后端文件+按tplCategory的vue文件）
    """
    templates = [
        'entity_do.py.j2',
        'entity_vo.py.j2',
        'dao.py.j2',
        'service.py.j2',
        'controller.py.j2',
        'menu.sql.j2',
        'api.js.j2',
    ]
    templates.append('index.vue.j2')
    return templates


def get_file_name(template: str, table: GenTable) -> str:
    """
    生成物文件路径（对齐java getFileName的目录结构语义，改为python项目结构）
    """
    class_name = table.class_name
    module_name = table.module_name
    business_name = table.business_name

    py_path = f'{PROJECT_PATH}/module_admin'
    vue_path = 'ruoyi-vue3/src'

    if template == 'entity_do.py.j2':
        return f'{py_path}/entity/do/{module_name}_{business_name}_do.py'
    if template == 'entity_vo.py.j2':
        return f'{py_path}/entity/vo/{module_name}_{business_name}_vo.py'
    if template == 'dao.py.j2':
        return f'{py_path}/dao/{module_name}_{business_name}_dao.py'
    if template == 'service.py.j2':
        return f'{py_path}/service/{module_name}_{business_name}_service.py'
    if template == 'controller.py.j2':
        return f'{py_path}/controller/{module_name}_{business_name}_controller.py'
    if template == 'menu.sql.j2':
        return f'sql/{business_name}_menu.sql'
    if template == 'api.js.j2':
        return f'{vue_path}/api/{module_name}/{business_name}.js'
    if template == 'index.vue.j2':
        return f'{vue_path}/views/{module_name}/{business_name}/index.vue'
    return template


def render_table(table: GenTable, columns: List[GenTableColumn]) -> dict:
    """
    渲染一张表的所有模板
    :return: {文件路径: 文件内容}（key与java previewCode一致为模板名 -> 内容时改为路径）
    """
    env = _jinja_env()
    context = build_template_context(table, columns)
    # 模板需要的派生变量
    pk = get_pk_column(columns)
    context['pk'] = pk
    context['snake_class_name'] = re.sub(r'(?<!^)(?=[A-Z])', '_', table.class_name).lower()
    # 菜单SQL的父菜单ID变量（java从options/parentMenuId取）
    import json as _json
    try:
        options = _json.loads(table.options) if table.options else {}
    except (ValueError, TypeError):
        options = {}
    context['parentMenuId'] = options.get('parentMenuId', 0)

    result = {}
    for template in get_template_list(table):
        tpl = env.get_template(template)
        content = tpl.render(**context)
        result[template] = content
    return result


def render_zip(tables: List[tuple]) -> bytes:
    """
    多表渲染并打zip包
    :param tables: [(GenTable, columns), ...]
    """
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, 'w', zipfile.ZIP_DEFLATED) as zf:
        for table, columns in tables:
            rendered = render_table(table, columns)
            for template, content in rendered.items():
                file_name = get_file_name(template, table)
                zf.writestr(file_name, content)
    return buf.getvalue()
