import os
import sys
import argparse
from pydantic_settings import BaseSettings
from functools import lru_cache
from dotenv import load_dotenv


class AppSettings(BaseSettings):
    """
    应用配置
    """
    app_env: str = 'dev'
    app_name: str = 'RuoYi-Vue-FastApi'
    app_root_path: str = ''
    app_host: str = '0.0.0.0'
    app_port: int = 8080
    app_version: str = '1.0.0'
    app_reload: bool = True
    app_ip_location_query: bool = False
    # 同一账号是否允许同时登录（与java版行为一致：允许，每个会话独立token）
    app_same_time_login: bool = True


class JwtSettings(BaseSettings):
    """
    Jwt配置
    与java版token配置保持一致：header为Authorization，有效期30分钟
    """
    jwt_secret_key: str = 'abcdefghijklmnopqrstuvwxyz'
    jwt_algorithm: str = 'HS512'
    jwt_expire_minutes: int = 30
    jwt_redis_expire_minutes: int = 30


class DataBaseSettings(BaseSettings):
    """
    数据库配置
    """
    db_host: str = '127.0.0.1'
    db_port: int = 3306
    db_username: str = 'root'
    db_password: str = 'password'
    db_database: str = 'ry-vue'
    db_echo: bool = False
    db_max_overflow: int = 10
    db_pool_size: int = 50
    db_pool_recycle: int = 3600
    db_pool_timeout: int = 30


class RedisSettings(BaseSettings):
    """
    Redis配置
    """
    redis_host: str = '127.0.0.1'
    redis_port: int = 6379
    redis_username: str = ''
    redis_password: str = ''
    redis_database: int = 0


class UploadSettings:
    """
    上传配置
    """
    UPLOAD_PREFIX = '/profile'
    UPLOAD_PATH = 'upload_path'
    DOWNLOAD_PATH = 'download_path'

    def __init__(self):
        if not os.path.exists(self.UPLOAD_PATH):
            os.makedirs(self.UPLOAD_PATH)
        if not os.path.exists(self.DOWNLOAD_PATH):
            os.makedirs(self.DOWNLOAD_PATH)


class GetConfig:
    """
    获取配置
    """

    def __init__(self):
        self.parse_cli_args()

    @lru_cache()
    def get_app_config(self):
        """
        获取应用配置
        """
        return AppSettings()

    @lru_cache()
    def get_jwt_config(self):
        """
        获取Jwt配置
        """
        return JwtSettings()

    @lru_cache()
    def get_database_config(self):
        """
        获取数据库配置
        """
        return DataBaseSettings()

    @lru_cache()
    def get_redis_config(self):
        """
        获取Redis配置
        """
        return RedisSettings()

    @lru_cache()
    def get_upload_config(self):
        """
        获取上传配置
        """
        return UploadSettings()

    @staticmethod
    def parse_cli_args():
        """
        解析命令行参数
        """
        if 'uvicorn' in sys.argv[0]:
            # 使用uvicorn启动时，命令行参数需要按照uvicorn的文档进行配置，无法自定义参数
            pass
        else:
            parser = argparse.ArgumentParser(description='命令行参数')
            parser.add_argument('--env', type=str, default='', help='运行环境')
            args, unknown = parser.parse_known_args()
            os.environ['APP_ENV'] = args.env if args.env else 'dev'
        run_env = os.environ.get('APP_ENV', '')
        env_file = '.env.dev'
        if run_env != '':
            env_file = f'.env.{run_env}'
        load_dotenv(env_file, override=True)


# 实例化获取配置类
GetConfig()
# 应用配置
AppConfig = GetConfig().get_app_config()
# Jwt配置
JwtConfig = GetConfig().get_jwt_config()
# 数据库配置
DataBaseConfig = GetConfig().get_database_config()
# Redis配置
RedisConfig = GetConfig().get_redis_config()
# 上传配置
UploadConfig = GetConfig().get_upload_config()
