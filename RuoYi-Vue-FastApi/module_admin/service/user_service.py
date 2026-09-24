"""
用户管理服务层（spec-02 个人中心部分，对应java版SysProfileController + SysRegisterService）

分层职责：本模块为profile/register业务逻辑；用户CRUD在spec-04扩展
"""
from datetime import datetime
from typing import Optional
from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession
from module_admin.entity.do.entity import SysUser, SysRole, SysPost, SysUserRole, SysUserPost
from module_admin.dao.login_dao import get_user_by_id, get_user_by_user_name
from common.constant import Constants, UserConstants, SysConfig as SysConfigKey
from common.enums import UserStatus
from common.message_util import message as msg
from config.redis_cache import RedisCache
from config.env import JwtConfig
from exceptions.exception import LoginException, ServiceException
from utils.pwd_util import PwdUtil
from utils.log_util import logger


class ProfileService:
    """
    个人信息服务（对应java版SysProfileController）
    """

    @classmethod
    async def get_profile(cls, query_db: AsyncSession, user_id: int) -> dict:
        """
        个人信息（对应java版profile()：user + roleGroup + postGroup）
        """
        from utils.common_util import transform_result
        user = await get_user_by_id(query_db, user_id)
        if not user:
            raise ServiceException(message='查询个人信息失败')

        role_group = await cls._role_group(query_db, user_id)
        post_group = await cls._post_group(query_db, user_id)
        return {
            'data': transform_result(user),
            'roleGroup': role_group,
            'postGroup': post_group,
        }

    @classmethod
    async def _role_group(cls, query_db: AsyncSession, user_id: int) -> str:
        """
        角色名组（逗号分隔，对应java selectUserRoleGroup）
        """
        roles = (await query_db.execute(
            select(SysRole.role_name)
                .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
                .where(SysUserRole.user_id == user_id)
        )).scalars().all()
        return ','.join(roles)

    @classmethod
    async def _post_group(cls, query_db: AsyncSession, user_id: int) -> str:
        """
        岗位名组（对应java selectUserPostGroup）
        """
        posts = (await query_db.execute(
            select(SysPost.post_name)
                .join(SysUserPost, SysPost.post_id == SysUserPost.post_id)
                .where(SysUserPost.user_id == user_id)
        )).scalars().all()
        return ','.join(posts)

    @classmethod
    async def update_profile(cls, query_db: AsyncSession, user_id: int,
                             nick_name: Optional[str], email: Optional[str],
                             phonenumber: Optional[str], sex: Optional[str]) -> str:
        """
        修改个人信息（对应java updateProfile：仅四字段；手机号/邮箱唯一性校验）
        :return: 错误文案；空串=成功
        """
        user = await get_user_by_id(query_db, user_id)
        if not user:
            return '修改个人信息异常，请联系管理员'

        if phonenumber:
            dup = await cls._check_unique(query_db, SysUser.phonenumber == phonenumber,
                                          SysUser.user_id != user_id, SysUser.del_flag == '0')
            if dup:
                return f"修改用户'{user.user_name}'失败，手机号码已存在"
        if email:
            dup = await cls._check_unique(query_db, SysUser.email == email,
                                          SysUser.user_id != user_id, SysUser.del_flag == '0')
            if dup:
                return f"修改用户'{user.user_name}'失败，邮箱账号已存在"

        user.nick_name = nick_name if nick_name is not None else user.nick_name
        user.email = email if email is not None else user.email
        user.phonenumber = phonenumber if phonenumber is not None else user.phonenumber
        user.sex = sex if sex is not None else user.sex
        await query_db.commit()
        return (user, '')

    @classmethod
    async def refresh_user_session(cls, request, query_db: AsyncSession, user_id: int):
        """
        刷新该用户的Redis会话用户信息（对应java tokenService.setLoginUser：
        profile/改密/头像更新后，getInfo等会话消费方立即可见）
        """
        from module_admin.service.login_service import get_redis_cache
        from common.constant import CacheConstants
        cache = get_redis_cache(request)
        for key in await cache.keys_by_prefix(CacheConstants.LOGIN_TOKEN_KEY):
            data = await cache.get_cache_object(key)
            if isinstance(data, dict) and data.get('user_id') == user_id:
                await cache.save_login_user(data.get('token'), data,
                                             JwtConfig.jwt_expire_minutes)

    @staticmethod
    async def _check_unique(query_db: AsyncSession, *conditions) -> bool:
        """
        存在即不唯一
        """
        found = (await query_db.execute(
            select(SysUser.user_id).where(*conditions).limit(1)
        )).first()
        return found is not None

    @classmethod
    async def update_pwd(cls, request, query_db: AsyncSession, user_id: int,
                         old_password: str, new_password: str) -> str:
        """
        修改密码（对应java updatePwd；含chrtype复杂度策略）
        :return: 错误文案；空串=成功
        """
        if not old_password:
            return msg('user.password.old.error')
        if not new_password:
            return msg('user.password.same')
        user = await get_user_by_id(query_db, user_id)
        if not user:
            return msg('user.not.exists')
        if not PwdUtil.verify_password(old_password, user.password):
            return msg('user.password.old.error')
        if PwdUtil.verify_password(new_password, user.password):
            return msg('user.password.same')

        # 密码字符范围策略（java版sys.account.chrtype：0任意/1数字/2字母/3字母数字/4字母数字特殊字符）
        from module_admin.dao.login_dao import get_config_by_key
        from common.constant import SysConfig
        config = await get_config_by_key(query_db, SysConfig.ACCOUNT_CHRTYPE)
        chrtype = config.config_value if config else '0'
        error = _check_password_chrype(new_password, chrtype)
        if error:
            return error

        user.password = PwdUtil.get_password_hash(new_password)
        user.pwd_update_date = datetime.now()
        await query_db.commit()

        # 刷新会话（对齐java tokenService.setLoginUser：后续getInfo取到新pwdUpdateDate）
        await cls.refresh_user_session(request, query_db, user_id)
        return ''


def _check_password_chrype(password: str, chrtype: str) -> str:
    """
    密码字符范围校验（java版sys.account.chrtype语义）
    :return: 错误提示；空=通过
    """
    if not chrtype or chrtype == '0':
        return ''
    if chrtype == '1':  # 纯数字
        if not password.isdigit():
            return '密码只能为0-9数字'
        return ''
    if chrtype == '2':  # 纯字母
        if not password.isalpha():
            return '密码只能为a-z和A-Z字母'
        return ''
    if chrtype == '3':  # 字母和数字
        if not (password.isalnum() and any(c.isalpha() for c in password)
                and any(c.isdigit() for c in password)):
            return '密码必须包含字母和数字'
        return ''
    if chrtype == '4':  # 字母数字特殊字符
        special = set('~!@#$%^&*()-=_+')
        has_alpha = any(c.isalpha() for c in password)
        has_digit = any(c.isdigit() for c in password)
        has_special = any(c in special for c in password)
        valid_chars = all(c.isalnum() or c in special for c in password)
        if not (valid_chars and has_alpha and has_digit and has_special):
            return f'密码必须包含字母、数字和特殊字符（{"~!@#$%^&*()-=_+"}）'
    return ''


class RegisterService:
    """
    用户注册（对应java版SysRegisterService.register，校验顺序与文案一致）
    """

    @classmethod
    async def register(cls, request, query_db: AsyncSession,
                       username: str, password: str,
                       code: Optional[str], uuid: Optional[str]) -> str:
        """
        :return: 错误文案；空串=成功
        """
        from module_admin.service.login_service import LoginService

        # 验证码（开关开启时）
        captcha_enabled = await LoginService._LoginService__get_captcha_enabled(query_db)
        if captcha_enabled:
            cache: RedisCache = request.app.state.redis_cache
            from common.constant import CacheConstants
            verify_key = cache.build_key(CacheConstants.CAPTCHA_CODE_KEY, uuid or '')
            captcha = await cache.get_string(verify_key)
            await cache.delete_object(verify_key)
            if not captcha:
                return msg('user.jcaptcha.expire')
            if not code or code.lower() != str(captcha).lower():
                return msg('user.jcaptcha.error')

        if not username:
            return msg('register.username.empty')
        if not password:
            return msg('register.password.empty')
        if len(username) < UserConstants.USERNAME_MIN_LENGTH or len(username) > UserConstants.USERNAME_MAX_LENGTH:
            return msg('register.username.length')
        if len(password) < UserConstants.PASSWORD_MIN_LENGTH or len(password) > UserConstants.PASSWORD_MAX_LENGTH:
            return msg('register.password.length')
        if await get_user_by_user_name(query_db, username):
            return msg('register.username.exists', username)

        new_user = SysUser(
            user_name=username,
            nick_name=username,
            password=PwdUtil.get_password_hash(password),
            pwd_update_date=datetime.now(),
            create_by=username,
            create_time=datetime.now(),
            status=UserStatus.OK.code,
            del_flag='0',
        )
        query_db.add(new_user)
        await query_db.commit()
        # 注册日志（对齐java AsyncFactory.recordLogininfor REGISTER）
        from module_admin.service.login_service import LoginService as LS
        from common.constant import Constants as C
        await LS._LoginService__record_logininfor(request, query_db, username, C.REGISTER,
                                                  msg('user.register.success'))
        return ''
