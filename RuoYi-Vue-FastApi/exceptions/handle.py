from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.exceptions import HTTPException
from exceptions.exception import AuthException, PermissionException, LoginException, ServiceException
from utils.response_util import ResponseUtil, JSONResponse, jsonable_encoder
from utils.log_util import logger


def handle_exception(app: FastAPI):
    """
    全局异常处理（返回码约定与java版GlobalExceptionHandler保持一致）
    """

    # 自定义登录异常（业务校验失败，code=500，HTTP状态200，前端走ElMessage提示）
    @app.exception_handler(LoginException)
    async def login_exception_handler(request: Request, exc: LoginException):
        return ResponseUtil.failure(msg=exc.message)

    # 自定义token检验异常
    @app.exception_handler(AuthException)
    async def auth_exception_handler(request: Request, exc: AuthException):
        return ResponseUtil.unauthorized(msg=exc.message)

    # 自定义权限检验异常
    @app.exception_handler(PermissionException)
    async def permission_exception_handler(request: Request, exc: PermissionException):
        return ResponseUtil.forbidden(msg=exc.message)

    # 业务异常
    @app.exception_handler(ServiceException)
    async def service_exception_handler(request: Request, exc: ServiceException):
        return ResponseUtil.failure(msg=exc.message)

    # 请求参数校验异常
    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(request: Request, exc: RequestValidationError):
        return ResponseUtil.failure(msg=f"请求参数校验异常：{exc.errors()}")

    # 处理其他http请求异常
    @app.exception_handler(HTTPException)
    async def http_exception_handler(request: Request, exc: HTTPException):
        return JSONResponse(
            content=jsonable_encoder({"code": exc.status_code, "msg": exc.detail}),
            status_code=exc.status_code
        )

    # 兜底异常
    @app.exception_handler(Exception)
    async def exception_handler(request: Request, exc: Exception):
        logger.exception(exc)
        return ResponseUtil.error(msg=str(exc))
