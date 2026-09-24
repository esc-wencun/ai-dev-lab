from sqlalchemy.exc import SQLAlchemyError
from config.database import AsyncSessionLocal
from utils.log_util import logger


async def get_db():
    """
    每一个请求处理完毕后会关闭当前连接，不同的请求使用不同的连接。
    事务语义（对应java @Transactional）：
    - 正常路径：service层自行 commit（Phase 0以来的行为，保持兼容）
    - 异常路径：此处统一 rollback，防止脏数据半提交（多表写入的安全网）
    - 多表写入的service必须使用显式事务（见规范文档），本兜底只做异常回滚
    """
    async with AsyncSessionLocal() as current_db:
        try:
            yield current_db
        except SQLAlchemyError as e:
            await current_db.rollback()
            logger.error(f"数据库事务异常已回滚：{e}")
            raise
        except Exception:
            await current_db.rollback()
            raise
