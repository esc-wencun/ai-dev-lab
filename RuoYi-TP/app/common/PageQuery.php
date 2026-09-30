<?php
declare(strict_types=1);

/**
 * 列表分页查询参数（对位经典版 ry-ui.js queryParams + BaseController.startPage）
 *
 * 解析：pageNum / pageSize / searchValue / orderByColumn / isAsc / params[beginTime] / params[endTime]
 * 排序字段白名单校验（驼峰 → 下划线），防注入。
 */
final class PageQuery
{
    public int $pageNum = 1;
    public int $pageSize = 10;
    public string $searchValue = '';
    /** 已校验的下划线排序字段，null 表示不排序 */
    public ?string $orderBy = null;
    public string $isAsc = 'asc';
    /** 时间范围 params[beginTime] / params[endTime] */
    public ?string $beginTime = null;
    public ?string $endTime = null;

    /** @param array $input 通常传 $request->param() */
    public static function from(array $input, array $sortableFields = []): self
    {
        $q = new self();
        $q->pageNum = max(1, (int)($input['pageNum'] ?? 1));
        $q->pageSize = min(100, max(1, (int)($input['pageSize'] ?? 10)));
        $q->searchValue = trim((string)($input['searchValue'] ?? ''));

        $orderByColumn = trim((string)($input['orderByColumn'] ?? ''));
        if ($orderByColumn !== '') {
            $underscore = self::camelToSnake($orderByColumn);
            if (in_array($underscore, $sortableFields, true)) {
                $q->orderBy = $underscore;
                $q->isAsc = strtolower((string)($input['isAsc'] ?? 'asc')) === 'desc' ? 'desc' : 'asc';
            }
            // 白名单外字段静默忽略（不排序），不报错——对齐经典版容错行为且杜绝注入
        }

        $begin = (string)($input['params']['beginTime'] ?? '');
        $end = (string)($input['params']['endTime'] ?? '');
        $q->beginTime = $begin !== '' ? $begin : null;
        $q->endTime = $end !== '' ? $end : null;
        return $q;
    }

    /** 驼峰转下划线（PHPUnit 固化）：createTime → create_time；已下划线原样 */
    public static function camelToSnake(string $field): string
    {
        return strtolower(preg_replace('/([a-z0-9])([A-Z])/', '$1_$2', $field));
    }
}
