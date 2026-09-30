<?php
declare(strict_types=1);

/**
 * TableDataInfo 分页信封（对位经典版 com.ruoyi.common.core.page.TableDataInfo）
 *
 * 输出 {code:0, msg:"查询成功", rows:[...], total:n}
 * 经典版 BaseController.getDataTable 实锤 code=0。
 *
 * 工厂模式返回 think\Response（json），控制器直接 return 即可。
 */
final class TableDataInfo
{
    public static function of(array $rows, int $total, string $msg = '查询成功'): \think\response\Json
    {
        return \think\Response::create([
            'code'  => TpConstant::CODE_SUCCESS,
            'msg'   => $msg,
            'rows'  => $rows,
            'total' => $total,
        ], 'json', 200);
    }
}
