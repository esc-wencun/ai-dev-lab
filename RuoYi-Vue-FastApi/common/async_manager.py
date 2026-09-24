"""
后台任务管理（对应java版AsyncManager）

语义：不阻塞请求主流程的异步任务统一入口；异常捕获并记录sys-error.log。
与FastAPI BackgroundTasks的取舍：本类用于"发出去就不用管"的日志/通知类任务；
需要与响应生命周期绑定的（响应完成后才执行）用FastAPI原生BackgroundTasks。
"""
import asyncio
from utils.log_util import logger


class AsyncManager:
    """
    异步任务执行器
    """

    @classmethod
    def submit(cls, coro):
        """
        提交后台任务（对应java AsyncManager.me().execute(task)）
        :param coro: 协程对象；内部异常被捕获并记录，不外泄
        """
        task = asyncio.create_task(cls._run(coro))
        return task

    @staticmethod
    async def _run(coro):
        try:
            await coro
        except Exception as e:
            logger.error(f"后台任务执行异常：{e}")
            logger.exception(e)
