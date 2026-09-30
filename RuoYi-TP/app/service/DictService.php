<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;
use think\facade\Db;

/**
 * 字典服务（6.0.0 收编完整版；对位 SysDictTypeServiceImpl + SysDictDataServiceImpl + DictUtils）
 *
 * 两域：类型域（sys_dict_type 管理）+ 数据域（sys_dict_data 管理）；
 * 缓存门面：`dict:<type>` 键（set/remove/clear/loading/reset），listByType 双键输出三不变（3.0.0 消费方零感知）。
 */
final class DictService
{
    /* ================= 缓存门面（对位 DictUtils） ================= */

    /** 按类型取启用字典数据（含徽章渲染字段；order by dict_sort）；缓存 miss 查库回填 */
    public static function listByType(string $dictType): array
    {
        $cacheKey = TpConstant::PREFIX_DICT . $dictType;
        $cached = RedisCache::get($cacheKey);
        if (is_array($cached)) {
            return $cached;
        }
        return self::setCache($dictType, null);
    }

    /** 写缓存（null 等价删键——对位经典 setDictCache(type, null)）；$rows 为空时按库全量重查 */
    public static function setCache(string $dictType, ?array $rows): array
    {
        if ($rows === null) {
            $rows = self::queryDictDataByType($dictType);
        }
        RedisCache::set(TpConstant::PREFIX_DICT . $dictType, $rows, 0);
        return $rows;
    }

    /** 删单类型缓存 */
    public static function removeCache(string $dictType): void
    {
        RedisCache::delete(TpConstant::PREFIX_DICT . $dictType);
    }

    /** 清全部 dict:* 键 */
    public static function clearCache(): void
    {
        foreach (RedisCache::keysScan(TpConstant::PREFIX_DICT . '*') as $key) {
            RedisCache::delete($key, true);
        }
    }

    /** 全量预热（对位 @PostConstruct loadingDictCache：全表 status='0' group by dict_type） */
    public static function loadingCache(): void
    {
        $types = Db::table('sys_dict_data')
            ->where('status', '0')
            ->group('dict_type')
            ->column('dict_type');
        foreach ($types as $type) {
            self::setCache((string)$type, null);
        }
    }

    /** 清 + 全量预热（对位 resetDictCache） */
    public static function resetCache(): void
    {
        self::clearCache();
        self::loadingCache();
    }

    /** 按类型查启用数据（status='0'，order by dict_sort；双键输出） */
    private static function queryDictDataByType(string $dictType): array
    {
        $rows = Db::table('sys_dict_data')
            ->where('dict_type', $dictType)
            ->where('status', '0')
            ->field('dict_label,dict_value,dict_type,css_class,list_class,is_default')
            ->order('dict_sort')
            ->select()->toArray();
        return array_map(static function (array $r): array {
            return [
                'dictLabel'  => $r['dict_label'],
                'dictValue'  => $r['dict_value'],
                'dictType'   => $r['dict_type'],
                'cssClass'   => $r['css_class'],
                'listClass'  => $r['list_class'],
                'isDefault'  => $r['is_default'],
                'dict_label' => $r['dict_label'],
                'dict_value' => $r['dict_value'],
                'is_default' => $r['is_default'],
            ];
        }, $rows);
    }

    /** 按类型+值反查 label（view 页性别等展示；未命中返回原值） */
    public static function getLabel(string $dictType, string $value): string
    {
        foreach (self::listByType($dictType) as $row) {
            if ((string)($row['dictValue'] ?? $row['dict_value'] ?? '') === $value) {
                return (string)($row['dictLabel'] ?? $row['dict_label'] ?? $value);
            }
        }
        return $value;
    }

    /* ================= 类型域（对位 SysDictTypeServiceImpl） ================= */

    /** 类型列表构造器（返回 Query；分页/全量由调用方决定） */
    public static function selectDictTypeList(array $filter): \think\db\Query
    {
        $query = Db::table('sys_dict_type')
            ->field('dict_id,dict_name,dict_type,status,create_by,create_time,remark');
        if (($filter['dictName'] ?? '') !== '') {
            $query->whereLike('dict_name', '%' . $filter['dictName'] . '%');
        }
        if (($filter['dictType'] ?? '') !== '') {
            $query->whereLike('dict_type', '%' . $filter['dictType'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', $filter['status']);
        }
        $begin = (string)($filter['beginTime'] ?? '');
        $end = (string)($filter['endTime'] ?? '');
        if ($begin !== '') {
            $query->whereTime('create_time', '>=', date('Y-m-d 00:00:00', strtotime($begin)));
        }
        if ($end !== '') {
            $query->whereTime('create_time', '<=', date('Y-m-d 23:59:59', strtotime($end)));
        }
        return $query;
    }

    /** 类型行 → 驼峰（selectVo 7 列） */
    public static function toTypeResponseRow(array $r): array
    {
        return [
            'dictId'     => (int)$r['dict_id'],
            'dictName'   => $r['dict_name'],
            'dictType'   => $r['dict_type'],
            'status'     => $r['status'],
            'createBy'   => $r['create_by'] ?? '',
            'createTime' => $r['create_time'] ?? null,
            'remark'     => $r['remark'] ?? '',
        ];
    }

    public static function selectDictTypeById(int $dictId): ?array
    {
        return Db::table('sys_dict_type')->where('dict_id', $dictId)->find();
    }

    public static function selectDictTypeByType(string $dictType): ?array
    {
        return Db::table('sys_dict_type')->where('dict_type', $dictType)->find();
    }

    /** 全类型（数据页类型下拉；输出驼峰） */
    public static function selectDictTypeAll(): array
    {
        $rows = Db::table('sys_dict_type')
            ->field('dict_id,dict_name,dict_type,status')
            ->order('dict_id')
            ->select()->toArray();
        return array_map(static fn(array $r): array => [
            'dictId'   => (int)$r['dict_id'],
            'dictName' => $r['dict_name'],
            'dictType' => $r['dict_type'],
            'status'   => $r['status'],
        ], $rows);
    }

    /** dict_type 全局唯一 → bool 唯一 */
    public static function checkDictTypeUnique(string $dictType, int $dictId = 0): bool
    {
        $row = Db::table('sys_dict_type')->where('dict_type', $dictType)->find();
        return $row === null || (int)$row['dict_id'] === $dictId;
    }

    public static function insertDictType(array $dict, string $loginName): void
    {
        Db::table('sys_dict_type')->insert(array_merge($dict, [
            'create_by'   => $loginName,
            'create_time' => date('Y-m-d H:i:s'),
        ]));
        // 缓存联动：removeCache（对位 setDictCache(type, null)——删键下次读回填）
        self::removeCache((string)$dict['dict_type']);
    }

    /** 修改（事务）：sys_dict_data.dict_type 旧→新级联 + 类型行更新 + 新类型缓存重刷；旧类型键不删（经典原样） */
    public static function updateDictType(array $dict, string $loginName): void
    {
        Db::startTrans();
        try {
            $dictId = (int)$dict['dict_id'];
            $old = Db::table('sys_dict_type')->where('dict_id', $dictId)->field('dict_type')->find();
            $oldType = (string)($old['dict_type'] ?? '');
            $newType = (string)$dict['dict_type'];

            if ($oldType !== $newType) {
                Db::table('sys_dict_data')->where('dict_type', $oldType)->update(['dict_type' => $newType]);
            }
            Db::table('sys_dict_type')->where('dict_id', $dictId)->update(array_merge($dict, [
                'update_by'   => $loginName,
                'update_time' => date('Y-m-d H:i:s'),
            ]));
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
        // 按新类型全量重刷缓存（事务外，对位经典 setDictCache(newType, 查库)）
        self::setCache($newType, null);
    }

    /** 删除前占用校验（对位 countDictDataByType） */
    public static function countDictDataByType(string $dictType): int
    {
        return Db::table('sys_dict_data')->where('dict_type', $dictType)->count();
    }

    /** 删除类型（循环删 + removeCache 由控制器驱动；经典无事务原样） */
    public static function deleteDictTypeById(int $dictId): int
    {
        return Db::table('sys_dict_type')->where('dict_id', $dictId)->delete();
    }

    /** Ztree 平铺数组（对位 selectDictTreeData：无 pId；name 含 HTML 空格实体；仅 status='0'） */
    public static function selectDictTree(): array
    {
        $rows = Db::table('sys_dict_type')
            ->where('status', '0')
            ->field('dict_id,dict_name,dict_type')
            ->order('dict_id')
            ->select()->toArray();
        $tree = [];
        foreach ($rows as $r) {
            $tree[] = [
                'id'    => (int)$r['dict_id'],
                'name'  => '(' . $r['dict_name'] . ')&nbsp;&nbsp;&nbsp;' . $r['dict_type'],
                'title' => $r['dict_type'],
            ];
        }
        return $tree;
    }

    /* ================= 数据域（对位 SysDictDataServiceImpl） ================= */

    /** 数据列表构造器 */
    public static function selectDictDataList(array $filter): \think\db\Query
    {
        $query = Db::table('sys_dict_data')
            ->field('dict_code,dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,create_by,create_time,remark');
        if (($filter['dictType'] ?? '') !== '') {
            $query->where('dict_type', $filter['dictType']); // 精确 =
        }
        if (($filter['dictLabel'] ?? '') !== '') {
            $query->whereLike('dict_label', '%' . $filter['dictLabel'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', $filter['status']);
        }
        return $query;
    }

    /** 数据行 → 驼峰（selectVo 12 列） */
    public static function toDataResponseRow(array $r): array
    {
        return [
            'dictCode'   => (int)$r['dict_code'],
            'dictSort'   => (int)$r['dict_sort'],
            'dictLabel'  => $r['dict_label'],
            'dictValue'  => $r['dict_value'],
            'dictType'   => $r['dict_type'],
            'cssClass'   => $r['css_class'],
            'listClass'  => $r['list_class'],
            'isDefault'  => $r['is_default'],
            'status'     => $r['status'],
            'createBy'   => $r['create_by'] ?? '',
            'createTime' => $r['create_time'] ?? null,
            'remark'     => $r['remark'] ?? '',
        ];
    }

    public static function selectDictDataById(int $dictCode): ?array
    {
        return Db::table('sys_dict_data')->where('dict_code', $dictCode)->find();
    }

    public static function insertDictData(array $data, string $loginName): void
    {
        Db::table('sys_dict_data')->insert(array_merge($data, [
            'create_by'   => $loginName,
            'create_time' => date('Y-m-d H:i:s'),
        ]));
        self::setCache((string)$data['dict_type'], null);
    }

    public static function updateDictData(array $data, string $loginName): void
    {
        Db::table('sys_dict_data')->where('dict_code', (int)$data['dict_code'])->update(array_merge($data, [
            'update_by'   => $loginName,
            'update_time' => date('Y-m-d H:i:s'),
        ]));
        self::setCache((string)$data['dict_type'], null);
    }

    /** 批量删（逐个查行 → 物理删 → 按该行 dictType 重刷缓存） */
    public static function deleteDictDataByIds(array $ids): void
    {
        foreach ($ids as $id) {
            $row = Db::table('sys_dict_data')->where('dict_code', (int)$id)->field('dict_type')->find();
            Db::table('sys_dict_data')->where('dict_code', (int)$id)->delete();
            if ($row) {
                self::setCache((string)$row['dict_type'], null);
            }
        }
    }
}
