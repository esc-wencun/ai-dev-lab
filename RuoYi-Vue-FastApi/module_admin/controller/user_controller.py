"""
用户模块控制器：个人中心 + 注册 + 通用上传下载（spec-02）
（对应java版SysProfileController / CommonController / SysRegisterController）
"""
import os
from fastapi import APIRouter, Request, Depends, UploadFile, File
from sqlalchemy.ext.asyncio import AsyncSession
from config.get_db import get_db
from config.env import UploadConfig
from common.constant import Constants, CacheConstants, SysConfig
from common.enums import BusinessType
from module_admin.entity.vo.login_vo import RegisterModel
from module_admin.service.login_service import LoginService, TokenService
from module_admin.service.user_service import ProfileService, RegisterService
from module_admin.annotation.log_annotation import log_decorator
from utils.upload_util import upload_one, upload_files, FileUploadException, is_allowed_download
from utils.response_util import ResponseUtil
from utils.log_util import logger

userController = APIRouter()
commonController = APIRouter()


# ==================== 个人中心 ====================

@userController.get('/system/user/profile')
async def profile(request: Request, query_db: AsyncSession = Depends(get_db)):
    """
    个人信息（{code, data, roleGroup, postGroup}）
    """
    try:
        login_user = await LoginService.get_current_user(request, query_db)
        result = await ProfileService.get_profile(query_db, login_user.get('user_id'))
        return ResponseUtil.success(msg='操作成功', dict_content=result)
    except Exception as e:
        logger.exception(e)
        return _handle(e)


@userController.put('/system/user/profile')
@log_decorator(title='个人信息', business_type=BusinessType.UPDATE)
async def update_profile(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    """
    修改个人信息（nickName/email/phonenumber/sex四字段）
    """
    try:
        login_user = await LoginService.get_current_user(request, query_db)
        user, error = await ProfileService.update_profile(
            query_db, login_user.get('user_id'),
            body.get('nickName'), body.get('email'),
            body.get('phonenumber'), body.get('sex'))
        if error:
            return ResponseUtil.failure(msg=error)
        # 刷新会话用户信息（对齐java tokenService.setLoginUser）
        await ProfileService.refresh_user_session(request, query_db, user.user_id)
        return ResponseUtil.success()
    except Exception as e:
        logger.exception(e)
        return _handle(e)


@userController.put('/system/user/profile/updatePwd')
@log_decorator(title='个人信息', business_type=BusinessType.UPDATE)
async def update_pwd(request: Request, body: dict, query_db: AsyncSession = Depends(get_db)):
    """
    修改密码（{oldPassword, newPassword}）
    """
    try:
        login_user = await LoginService.get_current_user(request, query_db)
        error = await ProfileService.update_pwd(
            request, query_db, login_user.get('user_id'),
            body.get('oldPassword'), body.get('newPassword'))
        if error:
            return ResponseUtil.failure(msg=error)
        return ResponseUtil.success(msg='修改密码成功')
    except Exception as e:
        logger.exception(e)
        return _handle(e)


@userController.post('/system/user/profile/avatar')
@log_decorator(title='用户头像', business_type=BusinessType.UPDATE)
async def avatar(request: Request, avatarfile: UploadFile = File(...),
                 query_db: AsyncSession = Depends(get_db)):
    """
    头像上传（multipart字段名avatarfile；返回{imgUrl}）
    """
    try:
        login_user = await LoginService.get_current_user(request, query_db)
        if not avatarfile or not avatarfile.filename:
            return ResponseUtil.failure(msg='上传图片异常，请联系管理员')
        from module_admin.entity.do.entity import SysUser
        from module_admin.dao.login_dao import get_user_by_id
        result = await upload_one(avatarfile, UploadConfig.UPLOAD_PATH)
        user = await get_user_by_id(query_db, login_user.get('user_id'))
        if not user:
            return ResponseUtil.failure(msg='上传图片异常，请联系管理员')
        user.avatar = result['fileName']
        await query_db.commit()
        await ProfileService.refresh_user_session(request, query_db, user.user_id)
        return ResponseUtil.success(msg='上传成功', dict_content={'imgUrl': result['fileName']})
    except FileUploadException as e:
        return ResponseUtil.failure(msg=str(e))
    except Exception as e:
        logger.exception(e)
        return _handle(e)


# ==================== 注册 ====================

@userController.post('/register')
async def register(request: Request, register_model: RegisterModel,
                   query_db: AsyncSession = Depends(get_db)):
    """
    用户注册（校验链对齐java SysRegisterService）
    """
    try:
        cache = request.app.state.redis_cache
        register_enabled = await cache.get_cache_object(
            cache.build_key(CacheConstants.SYS_CONFIG_KEY, SysConfig.REGISTER_USER))
        if not register_enabled or str(register_enabled).lower() != 'true':
            return ResponseUtil.failure(msg='当前系统没有开启注册功能！')
        error = await RegisterService.register(
            request, query_db, register_model.username, register_model.password,
            register_model.code, register_model.uuid)
        if error:
            return ResponseUtil.failure(msg=error)
        return ResponseUtil.success(msg='注册成功')
    except Exception as e:
        logger.exception(e)
        return _handle(e)


# ==================== 通用上传下载 ====================

@commonController.post('/common/upload')
async def upload(request: Request, file: UploadFile = File(...)):
    """
    通用上传（{url, fileName, newFileName, originalFilename}）
    """
    try:
        result = await upload_one(file, UploadConfig.UPLOAD_PATH)
        result['url'] = result['fileName']
        return ResponseUtil.success(msg='操作成功', dict_content=result)
    except FileUploadException as e:
        return ResponseUtil.failure(msg=str(e))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@commonController.post('/common/uploads')
async def upload_batch(request: Request, files: list[UploadFile] = File(...)):
    """
    批量上传（{url: [文件url列表]}，对齐java版uploadFiles返回）
    """
    try:
        results = await upload_files(files, UploadConfig.UPLOAD_PATH)
        urls = [r['fileName'] for r in results]
        return ResponseUtil.success(msg='操作成功',
                                    dict_content={'url': urls,
                                                  'files': [r['fileName'] for r in results]})
    except FileUploadException as e:
        return ResponseUtil.failure(msg=str(e))
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@commonController.get('/common/download')
async def download(request: Request, fileName: str, delete: bool = False):
    """
    文件下载（本地文件名；delete=true下载后删除）
    """
    from fastapi.responses import FileResponse
    try:
        if not is_allowed_download(fileName):
            return ResponseUtil.failure(msg=f'文件名称({fileName})非法，不允许下载。')
        file_path = os.path.join(UploadConfig.DOWNLOAD_PATH, fileName)
        if not os.path.exists(file_path):
            return ResponseUtil.failure(msg=f'文件({fileName})不存在。')
        # delete语义：先取内容再删（FileResponse background完成删除）
        response = FileResponse(path=file_path, filename=fileName)
        if delete:
            async def _cleanup():
                try:
                    os.remove(file_path)
                except OSError:
                    pass
            response.background = _cleanup
        return response
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


@commonController.get('/common/download/resource')
async def download_resource(request: Request, resource: str):
    """
    本地资源下载（/profile前缀校验）
    """
    from fastapi.responses import FileResponse
    from urllib.parse import unquote
    try:
        resource = unquote(resource)
        if not resource.startswith(Constants.RESOURCE_PREFIX):
            return ResponseUtil.failure(msg=f'资源文件({resource})非法，不允许下载。')
        local = resource[len(Constants.RESOURCE_PREFIX):].lstrip('/')
        if '..' in local:
            return ResponseUtil.failure(msg=f'资源文件({resource})非法，不允许下载。')
        file_path = os.path.join(UploadConfig.UPLOAD_PATH, local)
        if not os.path.exists(file_path):
            return ResponseUtil.failure(msg=f'资源文件({resource})不存在。')
        return FileResponse(path=file_path)
    except Exception as e:
        logger.exception(e)
        return ResponseUtil.error(msg=str(e))


def _handle(e: Exception):
    """
    profile路由统一异常出口（AuthException优先401）
    """
    from exceptions.exception import AuthException
    if isinstance(e, AuthException):
        return ResponseUtil.unauthorized(msg=e.message)
    return ResponseUtil.error(msg=str(e))
