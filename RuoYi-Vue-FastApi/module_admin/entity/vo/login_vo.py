from pydantic import BaseModel, ConfigDict
from pydantic.alias_generators import to_camel
from typing import Optional, List, Any
from datetime import datetime


class LoginModel(BaseModel):
    """
    登录模型（RuoYi-Vue3前端以JSON提交：username/password/code/uuid）
    """
    model_config = ConfigDict(alias_generator=to_camel)

    username: str
    password: str
    code: Optional[str] = ''
    uuid: Optional[str] = ''


class RegisterModel(BaseModel):
    """
    注册模型
    """
    model_config = ConfigDict(alias_generator=to_camel)

    username: str
    password: str
    confirm_password: Optional[str] = None
    code: Optional[str] = ''
    uuid: Optional[str] = ''


class SysUserModel(BaseModel):
    """
    用户信息模型（对应java版SysUser实体，返回给前端的用户信息）
    """
    model_config = ConfigDict(alias_generator=to_camel)

    user_id: Optional[int] = None
    dept_id: Optional[int] = None
    user_name: Optional[str] = None
    nick_name: Optional[str] = None
    user_type: Optional[str] = None
    email: Optional[str] = None
    phonenumber: Optional[str] = None
    sex: Optional[str] = None
    avatar: Optional[str] = None
    password: Optional[str] = None
    status: Optional[str] = None
    del_flag: Optional[str] = None
    login_ip: Optional[str] = None
    login_date: Optional[datetime] = None
    pwd_update_date: Optional[datetime] = None
    create_by: Optional[str] = None
    create_time: Optional[datetime] = None
    update_by: Optional[str] = None
    update_time: Optional[datetime] = None
    remark: Optional[str] = None
    dept: Optional[Any] = None
    roles: Optional[List[Any]] = []
    role_ids: Optional[List[int]] = None
    post_ids: Optional[List[int]] = None
    role_group: Optional[str] = None
    post_group: Optional[str] = None


class CurrentUserModel(BaseModel):
    """
    当前登录用户信息（/getInfo接口返回）
    """
    model_config = ConfigDict(alias_generator=to_camel)

    user: Optional[Any] = None
    roles: Optional[List[str]] = []
    permissions: Optional[List[str]] = []
    is_default_modify_pwd: Optional[bool] = None
    is_password_expired: Optional[bool] = None
    pwd_chrtype: Optional[str] = None


class MetaModel(BaseModel):
    """
    路由显示信息
    """
    model_config = ConfigDict(alias_generator=to_camel)

    title: Optional[str] = None
    icon: Optional[str] = None
    no_cache: Optional[bool] = False
    link: Optional[str] = None


class RouterModel(BaseModel):
    """
    路由配置信息（/getRouters接口返回）
    """
    model_config = ConfigDict(alias_generator=to_camel)

    name: Optional[str] = None
    path: Optional[str] = None
    hidden: Optional[bool] = False
    redirect: Optional[str] = None
    component: Optional[str] = None
    query: Optional[str] = None
    always_show: Optional[bool] = None
    meta: Optional[MetaModel] = None
    children: Optional[List['RouterModel']] = []


class TokenModel(BaseModel):
    """
    令牌信息
    """
    token: str
