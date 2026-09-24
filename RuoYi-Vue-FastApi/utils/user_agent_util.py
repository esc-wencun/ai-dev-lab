from user_agents import parse


def get_browser(user_agent: str) -> str:
    """
    获取浏览器类型
    """
    ua = parse(user_agent or '')
    return ua.browser.family or ''


def get_os(user_agent: str) -> str:
    """
    获取操作系统类型
    """
    ua = parse(user_agent or '')
    os_family = ua.os.family or ''
    version = ua.os.version_string if hasattr(ua.os, 'version_string') else ''
    if version:
        return f"{os_family} {version}".strip()
    return os_family
