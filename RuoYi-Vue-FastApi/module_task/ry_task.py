"""
RyTask 等价任务（对齐java版com.ruoyi.quartz.task.RyTask）

java版sys_job预置的3条任务都指向这里：
- ryTask.ryNoParams()
- ryTask.ryParams('ry')
- ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)
"""
from module_task.registry import register


@register('ryTask.ryNoParams')
async def ry_no_params():
    """
    无参方法（对应java ryNoParams：输出"执行无参方法"）
    """
    from utils.log_util import logger
    logger.info('执行无参方法')
    print('执行无参方法')


@register('ryTask.ryParams')
async def ry_params(params: str):
    """
    单字符串参数（对应java ryParams：输出"执行有参方法：ry"）
    """
    from utils.log_util import logger
    logger.info(f'执行有参方法：{params}')
    print(f'执行有参方法：{params}')


@register('ryTask.ryMultipleParams')
async def ry_multiple_params(s: str, b: bool, l: int, d: float, i: int):
    """
    多参方法（对应java ryMultipleParams：
    输出"执行多参方法： 字符串类型ry，布尔类型True，长整型2000，浮点型316.5，整形100"）
    """
    from utils.log_util import logger
    msg = (f'执行多参方法： 字符串类型{s}，布尔类型{b}，'
           f'长整型{l}，浮点型{d:g}，整形{i}')
    logger.info(msg)
    print(msg)
