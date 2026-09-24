import uuid as uuid_pkg
from fastapi import APIRouter, Request, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from config.redis_cache import RedisCache
from common.constant import CacheConstants, Constants, SysConfig
from module_admin.entity.vo.login_vo import LoginModel, RegisterModel
from module_admin.service.login_service import LoginService
from module_admin.service.captcha_service import CaptchaService
from exceptions.exception import LoginException, AuthException
from utils.response_util import ResponseUtil
from utils.log_util import logger

loginController = APIRouter()


@loginController.post('/login')
async def login(request: Request, login_model: LoginModel, query_db: AsyncSession = Depends(get_db)):
    """
    用户登录（RuoYi-Vue3前端以JSON提交，返回{code, msg, token}）
    """
    try:
        token = await LoginService.authenticate_user(request, query_db, login_model)
        return ResponseUtil.success(msg='登录成功', dict_content={'token': token})
    except LoginException as e:
        logger.warning(f'登录失败：{e.message}')
        return ResponseUtil.failure(msg=e.message)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@loginController.get('/captchaImage')
async def get_captcha_image(request: Request):
    """
    获取验证码（返回{code, msg, captchaEnabled, uuid, img}，与java版CaptchaController一致）
    """
    try:
        cache: RedisCache = request.app.state.redis_cache
        # 验证码开关（从redis缓存的系统配置读取）
        captcha_enabled_value = await cache.get_cache_object(
            cache.build_key(CacheConstants.SYS_CONFIG_KEY, SysConfig.CAPTCHA_ENABLED))
        captcha_enabled = True if captcha_enabled_value is None else str(captcha_enabled_value).lower() == 'true'

        result = {
            'captchaEnabled': captcha_enabled
        }
        if not captcha_enabled:
            return ResponseUtil.success(msg='操作成功', dict_content=result)

        session_uuid = str(uuid_pkg.uuid4())
        captcha_result = await CaptchaService.create_captcha_image_service()
        image = captcha_result[0]
        computed_result = captcha_result[1]
        await cache.set_string(
            cache.build_key(CacheConstants.CAPTCHA_CODE_KEY, session_uuid),
            str(computed_result),
            expire_minutes=Constants.CAPTCHA_EXPIRATION
        )
        result['uuid'] = session_uuid
        result['img'] = image
        logger.info(f'编号为{session_uuid}的会话获取图片验证码成功')
        return ResponseUtil.success(msg='操作成功', dict_content=result)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@loginController.get('/getInfo')
async def get_info(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    获取用户信息（返回{code, msg, user, roles, permissions, ...}，与java版getInfo一致）
    """
    try:
        user_info = await LoginService.get_user_info(request, query_db)
        return ResponseUtil.success(msg='操作成功', dict_content=user_info)
    except Exception as e:
        logger.exception(e)
        from exceptions.exception import AuthException
        if isinstance(e, AuthException):
            return ResponseUtil.unauthorized(msg=e.message)
        return ResponseUtil.error(msg=str(e))


@loginController.get('/getRouters')
async def get_routers(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    获取路由信息（返回{code, msg, data:[路由]}，与java版getRouters一致）
    """
    try:
        routers = await LoginService.get_routers(request, query_db)
        return ResponseUtil.success(data=routers)
    except Exception as e:
        logger.exception(e)
        from exceptions.exception import AuthException
        if isinstance(e, AuthException):
            return ResponseUtil.unauthorized(msg=e.message)
        return ResponseUtil.error(msg=str(e))


@loginController.post('/logout')
async def logout(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    退出登录
    """
    try:
        message = await LoginService.logout(request, query_db)
        logger.info(message)
        return ResponseUtil.success(msg=message)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@loginController.post('/unlockscreen')
async def unlockscreen(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    """
    解锁屏幕（与java版SysIndexController.unlockScreen一致）
    """
    try:
        from exceptions.exception import AuthException
        from module_admin.dao.login_dao import get_user_by_id
        from utils.pwd_util import PwdUtil
        password = body.get('password')
        if not password:
            return ResponseUtil.failure(msg='密码不能为空')
        login_user = await LoginService.get_current_user(request, query_db)
        if not login_user:
            return ResponseUtil.failure(msg='服务器超时，请重新登录')
        user = await get_user_by_id(query_db, login_user.get('user_id'))
        if not user:
            return ResponseUtil.failure(msg='服务器超时，请重新登录')
        if not PwdUtil.verify_password(password, user.password):
            return ResponseUtil.failure(msg='密码错误，请重新输入')
        return ResponseUtil.success(msg='解锁成功')
    except Exception as e:
        logger.exception(e)
        from exceptions.exception import AuthException
        if isinstance(e, AuthException):
            return ResponseUtil.unauthorized(msg=e.message)
        return ResponseUtil.error(msg=str(e))


@loginController.get('/')
async def index():
    """
    首页提示语（与java版SysIndexController.index一致）
    """
    from config.env import AppConfig
    return ResponseUtil.success(msg=f'欢迎使用{AppConfig.app_name}后台管理框架，当前版本：v{AppConfig.app_version}，请通过前端地址访问。')
