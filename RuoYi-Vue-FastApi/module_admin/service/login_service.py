import uuid as uuid_pkg
from datetime import datetime
from typing import Optional, List
from fastapi import Request
from jose import JWTError, jwt
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.dao.login_dao import *
from module_admin.entity.vo.login_vo import *
from exceptions.exception import LoginException, AuthException
from config.database import AsyncSessionLocal
from config.env import JwtConfig
from config.redis_cache import RedisCache
from common.constant import Constants, CacheConstants, UserConstants
from common.constant import SysConfig as SysConfigKey
from common.enums import UserStatus
from common import message_util
from utils.common_util import transform_result
from utils.pwd_util import PwdUtil
from utils.log_util import logger

# 管理员用户ID（java版SysUser.isAdmin：userId != null && 1L == userId）
ADMIN_USER_ID = 1
# 密码错误锁定配置（java版application.yml user.password.maxRetryCount/lockTime）
PASSWORD_MAX_RETRY_COUNT = 5
PASSWORD_LOCK_TIME_MINUTES = 10


def get_redis_cache(request: Request) -> RedisCache:
    """
    从请求中获取RedisCache实例（app启动时注入）
    """
    return request.app.state.redis_cache


class TokenService:
    """
    token服务（对应java版TokenService：JWT中仅存uuid，会话信息存redis）
    """

    @classmethod
    async def create_token(cls, request: Request, user_id: int, user_name: str, permissions: List[str]) -> str:
        """
        创建token：生成uuid作为会话键，LoginUser信息存redis，JWT中携带uuid
        """
        session_uuid = str(uuid_pkg.uuid4())
        now_ms = int(datetime.now().timestamp() * 1000)
        expire_minutes = JwtConfig.jwt_expire_minutes

        # 会话信息存入redis（key前缀CacheConstants.LOGIN_TOKEN_KEY，与java版一致）
        login_user = {
            'user_id': user_id,
            'user_name': user_name,
            'token': session_uuid,
            'login_time': now_ms,
            'expire_time': now_ms + expire_minutes * 60 * 1000,
            'permissions': list(permissions),
        }
        await get_redis_cache(request).save_login_user(session_uuid, login_user, expire_minutes)

        # JWT载荷（与java版Constants.LOGIN_USER_KEY/JWT_USERNAME保持一致）
        claims = {
            Constants.LOGIN_USER_KEY: session_uuid,
            Constants.JWT_USERNAME: user_name,
        }
        encoded_jwt = jwt.encode(claims, JwtConfig.jwt_secret_key, algorithm=JwtConfig.jwt_algorithm)
        return encoded_jwt

    @classmethod
    async def get_login_user(cls, request: Request) -> Optional[dict]:
        """
        从请求中解析token，获取redis中的会话信息
        """
        token = cls.get_token(request)
        if not token:
            return None
        try:
            payload = jwt.decode(token, JwtConfig.jwt_secret_key, algorithms=[JwtConfig.jwt_algorithm])
            session_uuid = payload.get(Constants.LOGIN_USER_KEY)
            if not session_uuid:
                return None
            return await get_redis_cache(request).get_login_user(session_uuid)
        except JWTError:
            logger.warning("用户token已失效，请重新登录")
            return None
        except Exception as e:
            logger.error(f"获取用户信息异常：{e}")
            return None

    @classmethod
    async def refresh_token(cls, request: Request, login_user: dict):
        """
        刷新令牌有效期（与java版verifyToken逻辑一致：剩余不足20分钟自动刷新）
        """
        expire_time = login_user.get('expire_time')
        now_ms = int(datetime.now().timestamp() * 1000)
        if expire_time and expire_time - now_ms <= 20 * 60 * 1000:
            expire_minutes = JwtConfig.jwt_expire_minutes
            login_user['login_time'] = now_ms
            login_user['expire_time'] = now_ms + expire_minutes * 60 * 1000
            await get_redis_cache(request).save_login_user(
                login_user.get('token'), login_user, expire_minutes)

    @classmethod
    async def del_login_user(cls, request: Request, token: str):
        """
        删除会话缓存
        """
        if token:
            await get_redis_cache(request).delete_login_user(token)

    @classmethod
    def get_token(cls, request: Request) -> Optional[str]:
        """
        从请求头中获取token（与java版一致：Authorization: Bearer xxx）
        """
        token = request.headers.get('Authorization')
        if token and token.startswith(Constants.TOKEN_PREFIX):
            token = token[len(Constants.TOKEN_PREFIX):]
            return token
        return None


class LoginService:
    """
    登录模块服务层（对应java版SysLoginService）
    """

    @classmethod
    async def authenticate_user(cls, request: Request, query_db: AsyncSession, login_model: LoginModel) -> str:
        """
        用户登录校验，成功返回token
        :param request: Request对象
        :param query_db: orm对象
        :param login_model: 登录参数
        :return: token
        """
        username = login_model.username
        password = login_model.password

        # 验证码校验（对应java版validateCaptcha）
        await cls.__validate_captcha(request, query_db, login_model.code, login_model.uuid)

        # 用户名密码长度前置校验（对应java版loginPreCheck）
        if not username or not password:
            raise LoginException(message=message_util.message('user.not.exists'))
        if len(password) < UserConstants.PASSWORD_MIN_LENGTH or len(password) > UserConstants.PASSWORD_MAX_LENGTH:
            raise LoginException(message=message_util.message('user.not.exists'))
        if len(username) < UserConstants.USERNAME_MIN_LENGTH or len(username) > UserConstants.USERNAME_MAX_LENGTH:
            raise LoginException(message=message_util.message('user.not.exists'))

        # IP黑名单校验
        config = await get_config_by_key(query_db, SysConfigKey.LOGIN_BLACK_IP_LIST)
        if config and config.config_value:
            client_ip = cls.get_client_ip(request)
            for black_ip in config.config_value.replace(';', ',').split(','):
                black_ip = black_ip.strip()
                if black_ip and black_ip == client_ip:
                    await cls.__record_logininfor(request, query_db, username, Constants.LOGIN_FAIL,
                                                  message_util.message('login.blocked'))
                    raise LoginException(message=message_util.message('login.blocked'))

        # 查询用户
        user = await get_user_by_user_name(query_db, username)
        if not user:
            await cls.__record_logininfor(request, query_db, username, Constants.LOGIN_FAIL,
                                          message_util.message('user.not.exists'))
            raise LoginException(message=message_util.message('user.not.exists'))
        if user.del_flag == UserStatus.DELETED.code:
            await cls.__record_logininfor(request, query_db, username, Constants.LOGIN_FAIL,
                                          message_util.message('user.password.delete'))
            raise LoginException(message=message_util.message('user.password.delete'))
        if user.status == UserStatus.DISABLE.code:
            await cls.__record_logininfor(request, query_db, username, Constants.LOGIN_FAIL,
                                          message_util.message('user.blocked'))
            raise LoginException(message=message_util.message('user.blocked'))

        # 密码错误次数校验（对应java版SysPasswordService.validate）
        await cls.__validate_password(request, user, password)

        # 登录成功，记录登录日志
        await cls.__record_logininfor(request, query_db, username, Constants.LOGIN_SUCCESS,
                                      message_util.message('user.login.success'))

        # 记录用户登录信息（对应java版recordLoginInfo）
        user.login_ip = cls.get_client_ip(request)
        user.login_date = datetime.now()
        await query_db.commit()

        # 权限查询并生成token
        permissions = await cls.__get_menu_permission(query_db, user)
        token = await TokenService.create_token(request, user.user_id, user.user_name, permissions)
        return token

    @classmethod
    async def __validate_captcha(cls, request: Request, query_db: AsyncSession,
                                  code: Optional[str], uuid: Optional[str]):
        """
        校验验证码（对应java版validateCaptcha）
        """
        captcha_enabled = await cls.__get_captcha_enabled(request, query_db)
        if captcha_enabled:
            cache = get_redis_cache(request)
            verify_key = cache.build_key(CacheConstants.CAPTCHA_CODE_KEY, uuid or '')
            captcha = await cache.get_string(verify_key)
            await cache.delete_object(verify_key)
            if not captcha:
                raise LoginException(message=message_util.message('user.jcaptcha.expire'))
            if not code or code.lower() != str(captcha).lower():
                raise LoginException(message=message_util.message('user.jcaptcha.error'))

    @classmethod
    async def __get_captcha_enabled(cls, request: Request, query_db: AsyncSession) -> bool:
        """
        查询验证码开关（spec-06改造：redis缓存优先，miss回源；与java SysConfigServiceImpl一致）
        """
        value = await cls.__get_config_value(request, query_db, SysConfigKey.CAPTCHA_ENABLED)
        if not value:
            return True
        return str(value).lower() == 'true'

    @classmethod
    async def __validate_password(cls, request: Request, user, password: str):
        """
        密码校验（对应java版SysPasswordService.validate：错误5次锁定10分钟）
        """
        username = user.user_name
        cache = get_redis_cache(request)
        cache_key = cache.build_key(CacheConstants.PWD_ERR_CNT_KEY, username)
        retry_count = await cache.get_counter(cache_key)

        if retry_count >= PASSWORD_MAX_RETRY_COUNT:
            raise LoginException(
                message=message_util.message('user.password.retry.limit.exceed',
                                             PASSWORD_MAX_RETRY_COUNT, PASSWORD_LOCK_TIME_MINUTES))

        if not PwdUtil.verify_password(password, user.password):
            await cache.increment(cache_key, PASSWORD_LOCK_TIME_MINUTES)
            raise LoginException(message=message_util.message('user.not.exists'))
        else:
            # 登录成功时清除密码错误次数缓存（与java版clearLoginRecordCache一致）
            if await cache.has_key(cache_key):
                await cache.delete_object(cache_key)

    @classmethod
    async def __get_menu_permission(cls, query_db: AsyncSession, user) -> List[str]:
        """
        获取菜单权限（对应java版SysPermissionService.getMenuPermission）
        """
        # 管理员拥有所有权限
        if user.user_id == ADMIN_USER_ID:
            return [Constants.ALL_PERMISSION]

        roles = await get_user_roles(query_db, user.user_id)
        perms = set()
        if roles:
            for role in roles:
                if role.status == UserConstants.NORMAL and role.role_id != ADMIN_USER_ID:
                    role_perms = await get_user_perms_by_role_id(query_db, role.role_id)
                    perms.update(p for p in role_perms if p)
        else:
            user_perms = await get_user_perms_by_user_id(query_db, user.user_id)
            perms.update(p for p in user_perms if p)
        return list(perms)

    @classmethod
    async def __record_logininfor(cls, request: Request, query_db: AsyncSession, username: str,
                                   status: str, message: str):
        """
        记录登录日志（对应java版AsyncFactory.recordLogininfor）
        status为Constants.LOGIN_SUCCESS('Success')/LOGIN_FAIL('Error')，落库转换'0'/'1'
        """
        from module_admin.entity.do.entity import SysLogininfor
        from utils.user_agent_util import get_browser, get_os
        from utils.log_util import record_logininfor_log
        from common.async_manager import AsyncManager
        ip = cls.get_client_ip(request)
        address = '内网IP' if cls.__is_inner_ip(ip) else 'XX XX'
        # 应用日志输出（对应java版sys-user logger）
        record_logininfor_log(ip, address, username, status, message)
        db_status = '0' if status == Constants.LOGIN_SUCCESS else '1'

        async def _save():
            async with AsyncSessionLocal() as session:
                session.add(SysLogininfor(
                    user_name=username,
                    ipaddr=ip,
                    login_location=address,
                    browser=get_browser(request.headers.get('User-Agent', '')),
                    os=get_os(request.headers.get('User-Agent', '')),
                    status=db_status,
                    msg=message,
                    login_time=datetime.now(),
                ))
                await session.commit()

        # 后台异步写库（对应java AsyncManager），不阻塞登录响应
        AsyncManager.submit(_save())

    @classmethod
    def __is_inner_ip(cls, ip: str) -> bool:
        """
        判断是否为内网IP
        """
        if not ip:
            return False
        return ip.startswith('127.') or ip.startswith('192.168.') or ip.startswith('10.') or ip.startswith('172.')

    @classmethod
    def get_client_ip(cls, request: Request) -> str:
        """
        获取客户端IP
        """
        forwarded = request.headers.get('X-Forwarded-For')
        if forwarded:
            return forwarded.split(',')[0].strip()
        real_ip = request.headers.get('X-Real-IP')
        if real_ip:
            return real_ip
        return request.client.host if request.client else ''

    @classmethod
    async def get_current_user(cls, request: Request, query_db: AsyncSession = None) -> dict:
        """
        获取当前登录用户信息（对应java版SecurityUtils.getLoginUser）
        :param query_db: 兼容保留（会话信息全在redis，暂不需要db）
        """
        login_user = await TokenService.get_login_user(request)
        if not login_user:
            raise AuthException(message=message_util.message('user.notfound'))
        # 会话续期（与java版verifyToken一致：剩余不足20分钟自动刷新）
        await TokenService.refresh_token(request, login_user)
        return login_user

    @classmethod
    async def get_user_info(cls, request: Request, query_db: AsyncSession) -> dict:
        """
        获取用户信息（对应java版SysLoginController.getInfo）
        用户主体经CamelCaseUtil自动转驼峰（密码列不在此查询，无泄漏风险），组合字段单独附加
        """
        login_user = await cls.get_current_user(request, query_db)
        user = await get_user_by_id(query_db, login_user.get('user_id'))
        if not user:
            raise AuthException(message=message_util.message('user.notfound'))

        roles = await cls.__get_role_permission(query_db, user)
        permissions = await cls.__get_menu_permission(query_db, user)

        # 查询部门信息（前端个人中心使用user.dept.deptName）
        dept_dict = None
        if user.dept_id:
            dept = await get_dept_by_id(query_db, user.dept_id)
            if dept:
                dept_dict = transform_result(dept)

        user_dict = transform_result(user)
        user_dict.update({
            'dept': dept_dict,
            'roles': transform_result(await get_user_roles(query_db, user.user_id)),
            'admin': user.user_id == ADMIN_USER_ID,
        })

        # 密码策略配置（对应java版getSysAccountChrtype等）
        chrtype = await cls.__get_config_value(request, query_db, SysConfigKey.ACCOUNT_CHRTYPE, '0')
        init_password_modify = await cls.__get_config_value(request, query_db, SysConfigKey.INIT_PASSWORD_MODIFY, None)
        password_validate_days = await cls.__get_config_value(request, query_db, SysConfigKey.PASSWORD_VALIDATE_DAYS, None)

        is_default_modify_pwd = init_password_modify == '1' and user.pwd_update_date is None
        is_password_expired = False
        if password_validate_days and password_validate_days.isdigit() and int(password_validate_days) > 0:
            if user.pwd_update_date is None:
                is_password_expired = True
            else:
                diff_days = (datetime.now() - user.pwd_update_date).days
                is_password_expired = diff_days > int(password_validate_days)

        return {
            'user': user_dict,
            'roles': roles,
            'permissions': permissions,
            'pwdChrtype': chrtype,
            'isDefaultModifyPwd': is_default_modify_pwd,
            'isPasswordExpired': is_password_expired,
        }

    @classmethod
    async def __get_config_value(cls, request: Request, query_db: AsyncSession, key: str,
                                  default=None):
        """
        读取参数配置（spec-06：redis缓存优先，miss回源数据库并写缓存）
        """
        cache = get_redis_cache(request)
        cache_key = cache.build_key(CacheConstants.SYS_CONFIG_KEY, key)
        cached = await cache.get_cache_object(cache_key)
        if cached is not None:
            return cached
        config = await get_config_by_key(query_db, key)
        if config:
            await cache.set_cache_object(cache_key, config.config_value)
            return config.config_value
        return default

    @classmethod
    async def __get_role_permission(cls, query_db: AsyncSession, user) -> List[str]:
        """
        获取角色权限（对应java版SysPermissionService.getRolePermission）
        """
        if user.user_id == ADMIN_USER_ID:
            return [Constants.SUPER_ADMIN]
        role_keys = await get_user_role_keys(query_db, user.user_id)
        return list(role_keys)

    @classmethod
    async def logout(cls, request: Request, query_db: AsyncSession):
        """
        退出登录
        """
        token = TokenService.get_token(request)
        if token:
            try:
                payload = jwt.decode(token, JwtConfig.jwt_secret_key, algorithms=[JwtConfig.jwt_algorithm],
                                     options={'verify_exp': False})
                session_uuid = payload.get(Constants.LOGIN_USER_KEY)
                if session_uuid:
                    await TokenService.del_login_user(request, session_uuid)
            except JWTError:
                pass
        return '退出成功'

    @classmethod
    async def get_routers(cls, request: Request, query_db: AsyncSession):
        """
        获取路由信息（对应java版SysLoginController.getRouters）
        """
        login_user = await cls.get_current_user(request, query_db)
        user_id = login_user.get('user_id')

        if user_id == ADMIN_USER_ID:
            menus = await get_menu_tree_all(query_db)
        else:
            menus = await get_menu_tree_by_user_id(query_db, user_id)

        return build_menus(menus)


def get_child_perms(menus: List, parent_id: int = 0) -> List:
    """
    根据父节点构建菜单树（对应java版getChildPerms）
    """
    tree = []
    for menu in menus:
        if menu.parent_id == parent_id:
            menu.children = get_child_perms(menus, menu.menu_id)
            tree.append(menu)
    return tree


def build_menus(menus: List) -> List[dict]:
    """
    构建前端路由所需菜单（完整复刻java版SysMenuServiceImpl.buildMenus）
    """
    TYPE_DIR = 'M'
    TYPE_MENU = 'C'
    NO_FRAME = '1'
    LAYOUT = 'Layout'
    PARENT_VIEW = 'ParentView'
    INNER_LINK = 'InnerLink'
    MENU_ROOT_ID = 0

    def is_http(link: str) -> bool:
        return link.startswith('http://') or link.startswith('https://')

    def is_menu_frame(menu) -> bool:
        return menu.parent_id == MENU_ROOT_ID and menu.menu_type == TYPE_MENU and menu.is_frame == int(NO_FRAME)

    def is_inner_link(menu) -> bool:
        return menu.is_frame == int(NO_FRAME) and menu.path and is_http(menu.path)

    def is_parent_view(menu) -> bool:
        return menu.parent_id != MENU_ROOT_ID and menu.menu_type == TYPE_DIR

    def get_route_name(menu) -> str:
        if is_menu_frame(menu):
            return ''
        router_name = menu.route_name if menu.route_name else menu.path
        return router_name[:1].upper() + router_name[1:] if router_name else ''

    def get_route_name2(name: str, path: str) -> str:
        router_name = name if name else path
        return router_name[:1].upper() + router_name[1:] if router_name else ''

    def get_router_path(menu) -> str:
        router_path = menu.path
        if menu.parent_id != MENU_ROOT_ID and is_inner_link(menu):
            router_path = router_path.replace('http://', '').replace('https://', '').replace('www.', '').replace('/', '').replace('.', '')
        if menu.parent_id == MENU_ROOT_ID and menu.menu_type == TYPE_DIR and menu.is_frame == int(NO_FRAME):
            router_path = '/' + menu.path
        elif is_menu_frame(menu):
            router_path = '/'
        return router_path

    def get_component(menu) -> str:
        component = LAYOUT
        if menu.component and not is_menu_frame(menu):
            component = menu.component
        elif not menu.component and menu.parent_id != MENU_ROOT_ID and is_inner_link(menu):
            component = INNER_LINK
        elif not menu.component and is_parent_view(menu):
            component = PARENT_VIEW
        return component

    def build(menu_list: List) -> List[dict]:
        routers = []
        for menu in menu_list:
            router = {}
            router['hidden'] = menu.visible == '1'
            router['name'] = get_route_name(menu)
            router['path'] = get_router_path(menu)
            router['component'] = get_component(menu)
            router['query'] = menu.query
            meta = {
                'title': menu.menu_name,
                'icon': menu.icon,
                'noCache': menu.is_cache == 1,
            }
            # 与java版MetaVo(title, icon, noCache, link)四参构造一致：path为http(s)开头即设置link
            if menu.path and is_http(menu.path):
                meta['link'] = menu.path
            router['meta'] = meta

            children = getattr(menu, 'children', None) or []
            if children and menu.menu_type == TYPE_DIR:
                router['alwaysShow'] = True
                router['redirect'] = 'noRedirect'
                router['children'] = build(children)
            elif is_menu_frame(menu):
                router['meta'] = None
                child = {
                    'path': menu.path,
                    'component': menu.component,
                    'name': get_route_name2(menu.route_name, menu.path),
                    'meta': {
                        'title': menu.menu_name,
                        'icon': menu.icon,
                        'noCache': menu.is_cache == 1,
                    },
                    'query': menu.query,
                }
                router['children'] = [child]
            elif menu.parent_id == MENU_ROOT_ID and is_inner_link(menu):
                router['meta'] = {
                    'title': menu.menu_name,
                    'icon': menu.icon,
                }
                router['path'] = '/'
                router_path = menu.path.replace('http://', '').replace('https://', '').replace('www.', '').replace('/', '').replace('.', '')
                child = {
                    'path': router_path,
                    'component': INNER_LINK,
                    'name': get_route_name2(menu.route_name, router_path),
                    'meta': {
                        'title': menu.menu_name,
                        'icon': menu.icon,
                        'link': menu.path,
                    },
                }
                router['children'] = [child]
            routers.append(router)
        return routers

    menu_tree = get_child_perms(menus, MENU_ROOT_ID)
    return build(menu_tree)
