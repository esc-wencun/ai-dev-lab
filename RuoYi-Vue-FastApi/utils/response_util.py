from fastapi import status
from fastapi.responses import JSONResponse, Response, StreamingResponse
from fastapi.encoders import jsonable_encoder
from typing import Any, Dict, Optional
from datetime import datetime


class ResponseUtil:
    """
    响应工具类（返回格式与java版AjaxResult保持一致：{code, msg, ...}）
    """

    @classmethod
    def success(cls, msg: str = '操作成功', data: Optional[Any] = None, rows: Optional[Any] = None,
                dict_content: Optional[Dict] = None) -> Response:
        """
        成功响应方法
        """
        result = {
            'code': 200,
            'msg': msg
        }

        if data is not None:
            result['data'] = data
        if rows is not None:
            result['rows'] = rows
        if dict_content is not None:
            result.update(dict_content)

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def failure(cls, msg: str = '操作失败', code: int = 500, dict_content: Optional[Dict] = None) -> Response:
        """
        失败响应方法（与java版AjaxResult.error一致，code默认500）
        """
        result = {
            'code': code,
            'msg': msg
        }

        if dict_content is not None:
            result.update(dict_content)

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def warn(cls, msg: str = '操作警告', dict_content: Optional[Dict] = None) -> Response:
        """
        警告响应方法（与java版AjaxResult.warn一致，code=601）
        """
        result = {
            'code': 601,
            'msg': msg
        }

        if dict_content is not None:
            result.update(dict_content)

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def unauthorized(cls, msg: str = '请求访问：认证失败，无法访问系统资源') -> Response:
        """
        未认证响应方法（与java版AuthenticationEntryPointImpl一致：HTTP状态200，响应体code=401，
        前端axios响应拦截器根据body中的code=401弹出重新登录提示）
        """
        result = {
            'code': 401,
            'msg': msg
        }

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def forbidden(cls, msg: str = '没有权限，请联系管理员授权') -> Response:
        """
        无权限响应方法（与java版GlobalExceptionHandler一致：HTTP状态200，响应体code=403）
        """
        result = {
            'code': 403,
            'msg': msg
        }

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def error(cls, msg: str = '接口异常') -> Response:
        """
        错误响应方法
        """
        result = {
            'code': 500,
            'msg': msg
        }

        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content=jsonable_encoder(result)
        )

    @classmethod
    def streaming(cls, *, data: Any = None):
        """
        流式响应方法
        """
        return StreamingResponse(
            status_code=status.HTTP_200_OK,
            content=data
        )
