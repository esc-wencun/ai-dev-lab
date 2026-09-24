"""
XSS输入过滤（对应java版XssFilter）

规范：默认对所有POST/PUT的JSON body中字符串字段清洗HTML标签；
排除名单（如/system/notice富文本）内的端点跳过清洗。
"""
import json
import re
from fastapi import Request

# 排除清洗的路径前缀（对齐java版application.yml xss.excludes）
EXCLUDED_PATHS = {'/system/notice'}

# HTML标签模式（对齐java版HTMLFilter的核心语义：剥离标签）
_TAG_RE = re.compile(r'<[^>]*>')


def _clean_str(value: str) -> str:
    return _TAG_RE.sub('', value) if isinstance(value, str) else value


def _clean_obj(obj, depth=0):
    """
    递归清洗dict/list中的字符串
    """
    if depth > 6:
        return obj
    if isinstance(obj, dict):
        return {k: _clean_obj(v, depth + 1) for k, v in obj.items()}
    if isinstance(obj, list):
        return [_clean_obj(i, depth + 1) for i in obj]
    return _clean_str(obj)


async def xss_filter(request: Request):
    """
    FastAPI依赖：清洗JSON body中的HTML标签；body缓存到request.state供handler复用
    """
    if request.method not in ('POST', 'PUT'):
        return
    path = request.url.path
    if any(path.startswith(p) for p in EXCLUDED_PATHS):
        # 排除端点也预读body保持一致性
        body = await request.body()
        request.state.cached_body = body.decode('utf-8', errors='replace')
        return
    body = await request.body()
    text = body.decode('utf-8', errors='replace')
    request.state.cached_body = text
    try:
        parsed = json.loads(text)
        cleaned = _clean_obj(parsed)
        import json as _json
        request.state.cached_body = _json.dumps(cleaned, ensure_ascii=False)
    except (ValueError, TypeError):
        pass  # 非JSON body不处理
