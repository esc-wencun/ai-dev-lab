"""
任务目标注册表（spec-09 Task 1，替代java反射invokeTarget）

规范：所有可调度的任务函数必须在此注册；
invoke_target字符串格式 task.函数名（比java白名单更严格）。
"""
from typing import Callable, Dict
from utils.log_util import logger

# 注册表：invoke_target -> 可调用函数
_REGISTRY: Dict[str, Callable] = {}


def register(name: str):
    """
    任务注册装饰器
    用法：
        @register('task.sample')
        async def sample_task(): ...
    """

    def decorator(func: Callable):
        _REGISTRY[name] = func
        return func
    return decorator


def get_task(target: str) -> Callable:
    """
    取任务函数；未注册抛ValueError（对应java JobInvokeUtil的合法性校验）
    """
    func = _REGISTRY.get(target)
    if func is None:
        raise ValueError(f'任务目标未注册: {target}，可用: {sorted(_REGISTRY.keys())}')
    return func


def is_registered(target: str) -> bool:
    return target in _REGISTRY


def all_targets():
    return sorted(_REGISTRY.keys())


# ============ 内置示例任务 ============

@register('task.no_params')
async def no_params_task():
    """无参示例任务"""
    logger.info('执行无参示例任务')


@register('task.with_params')
async def with_params_task(s: str = 'ry', n: int = 1):
    """带参示例任务（对应java ryTask.ryParams('ry')演示）"""
    logger.info(f'执行带参示例任务: s={s}, n={n}')


@register('task.fail_demo')
async def fail_demo_task():
    """失败示例任务（验证失败日志记录）"""
    raise RuntimeError('演示性任务失败')
