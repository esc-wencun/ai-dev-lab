import uvicorn
from server import app, AppConfig
from config.env import AppConfig


if __name__ == '__main__':
    uvicorn.run(
        app='server:app',
        host=AppConfig.app_host,
        port=AppConfig.app_port,
        root_path=AppConfig.app_root_path,
        reload=AppConfig.app_reload
    )
