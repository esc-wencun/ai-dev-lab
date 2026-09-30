<?php
declare(strict_types=1);

/**
 * AjaxResult 响应信封（对位经典版 com.ruoyi.common.core.domain.AjaxResult）
 *
 * code 体系 = 经典版 AjaxResult.Type：SUCCESS(0) / WARN(301) / ERROR(500)
 * 注意不是 RuoYi-Vue 的 200/500/601。
 *
 * 工厂模式返回 think\Response（json），控制器直接 return 即可。
 */
final class AjaxResult
{
    /**
     * @param int    $code 0 成功 / 301 警告 / 500 错误
     * @param string $msg
     * @param mixed  $data 附加数据（数组展开到顶层，对位经典版 put(data)）
     */
    public static function of(int $code, string $msg = '操作成功', mixed $data = null): \think\response\Json
    {
        $body = ['code' => $code, 'msg' => $msg];
        if ($data !== null) {
            if (is_array($data)) {
                $body = array_merge($body, $data);
            } else {
                $body['data'] = $data;
            }
        }
        return \think\Response::create($body, 'json', 200);
    }

    public static function success(string $msg = '操作成功', mixed $data = null): \think\response\Json
    {
        return self::of(TpConstant::CODE_SUCCESS, $msg, $data);
    }

    public static function warn(string $msg = '', mixed $data = null): \think\response\Json
    {
        return self::of(TpConstant::CODE_WARN, $msg !== '' ? $msg : '警告', $data);
    }

    public static function error(string $msg = '操作失败', mixed $data = null): \think\response\Json
    {
        return self::of(TpConstant::CODE_ERROR, $msg, $data);
    }
}
