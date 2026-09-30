<?php
declare(strict_types=1);

/**
 * 业务异常：service 层业务失败统一抛出，由全局异常处理渲染为 AjaxResult::error（HTTP 200 + code 500）
 */
class BusinessException extends \RuntimeException
{
}
