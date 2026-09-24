"""
接口权限校验（对应java版PreAuthorize @ss.hasPermi / @ss.hasRole）

用法：
    @userController.get('/list', dependencies=[Depends(require_perm('system:user:list'))])
规范：业务接口一律声明权限标识，未声明的默认仅需登录。
"""
from fastapi import Request
from exceptions.exception import PermissionException
from module_admin.service.login_service import LoginService, get_redis_cache, ADMIN_USER_ID


def _match(perms: set, pattern: str) -> bool:
    """
    权限匹配（对齐java版PermissionService.hasPermi：
    ALL_PERMISSION通配全部；支持段通配 system:user:*；*:*:* 优先）
    """
    if not pattern:
        return False
    for perm in perms:
        if perm == pattern or perm == '*:*:*':
            return True
        # 分段通配：权限字符串按:切分逐段匹配，*匹配任意段
        p1 = perm.split(':')
        p2 = pattern.split(':')
        if len(p1) == len(p2) and all(a == '*' or a == b for a, b in zip(p1, p2)):
            return True
    return False


async def validate_permission(request: Request, permission: str):
    """
    校验当前用户是否持有指定权限，无权限抛PermissionException（403）
    """
    login_user = await LoginService.get_current_user(request)
    if not login_user:
        raise PermissionException(message='登录状态已过期，请重新登录')
    perms = set(login_user.get('permissions') or [])
    if not _match(perms, permission):
        raise PermissionException(message=f'没有权限，请联系管理员授权 [{permission}]')


async def validate_role(request: Request, role: str):
    """
    校验当前用户是否持有指定角色（admin角色用户恒通过，对应java @ss.hasRole）
    """
    login_user = await LoginService.get_current_user(request)
    if not login_user:
        raise PermissionException(message='登录状态已过期，请重新登录')
    if login_user.get('user_id') == ADMIN_USER_ID:
        return
    perms = set(login_user.get('permissions') or [])
    # 角色信息存于会话外，此处简化：会话permissions中admin角色直接放行（与java行为一致：admin用户全通过）
    from module_admin.dao.login_dao import get_user_role_keys
    # 角色key不入会话（会话只存permissions），走数据库查询（后续spec-05优化为会话内缓存）
    from config.get_db import AsyncSessionLocal
    async with AsyncSessionLocal() as db:
        role_keys = await get_user_role_keys(db, login_user.get('user_id'))
    if role not in role_keys:
        raise PermissionException(message=f'没有权限，请联系管理员授权 [{role}]')


def require_perm(permission: str):
    """
    FastAPI依赖工厂：校验接口权限
    """
    from fastapi import Depends

    async def _dependency(request: Request):
        await validate_permission(request, permission)
    return _dependency


def require_role(role: str):
    """
    FastAPI依赖工厂：校验接口角色
    """
    async def _dependency(request: Request):
        await validate_role(request, role)
    return _dependency
