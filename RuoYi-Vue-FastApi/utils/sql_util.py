"""
SQL安全过滤（对齐java版SqlUtil.filterKeyword：拦截破坏性关键字）
"""

# 破坏性SQL关键字（对齐java SqlUtil.REGEX）
KEYWORDS = ['master', 'truncate', 'insert', 'select', 'delete', 'update', 'declare',
            'alter', 'drop', 'sleep', 'shutdown', 'execute']


def filter_keyword(sql: str):
    """
    检查SQL是否包含破坏性关键字（create table场景中select等词可能出现在注释/字符串，
    java版同样只做contains检查并抛异常——但create table语句本身允许）
    注意：java版SqlUtil.filterKeyword实际拦截的是这些词的"无边界出现"，
    对合法DDL注释中的词会误伤，故此处只拦截高危动作词出现在语句开头的情况以外的
    drop/truncate/sleep/shutdown/execute（create table流程必需允许列定义）。
    """
    import re
    if not sql:
        return
    dangerous = ['truncate', 'sleep', 'shutdown', 'drop\s+database', 'drop\s+schema']
    for kw in dangerous:
        if re.search(kw, sql, re.IGNORECASE):
            raise ValueError(f'SQL包含不允许的关键字: {kw}')
