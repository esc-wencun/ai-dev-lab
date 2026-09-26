package com.ruoyi.system.mapper;

import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.domain.entity.SysDictType;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;

/**
 * 字典表 数据层
 *
 * @author ruoyi
 */
public interface SysDictTypeMapper extends BaseMapper<SysDictType>
{
    /**
     * 分页查询字典类型（动态条件 Wrapper 组装；beginTime/endTime 时间区间走 betweenIfPresent）
     *
     * @param dictType 字典类型信息
     * @return 分页结果
     */
    default Page<SysDictType> selectDictTypePage(SysDictType dictType)
    {
        java.util.Map<String, Object> params = dictType.getParams();
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<SysDictType>()
                .likeIfPresent(SysDictType::getDictName, dictType.getDictName())
                .likeIfPresent(SysDictType::getDictType, dictType.getDictType())
                .eqIfPresent(SysDictType::getStatus, dictType.getStatus())
                .betweenIfPresent(SysDictType::getCreateTime,
                        params != null ? params.get("beginTime") : null,
                        params != null ? params.get("endTime") : null));
    }

    /**
     * 根据条件查询字典类型列表（非分页：导出等）
     *
     * @param dictType 字典类型信息
     * @return 字典类型集合信息
     */
    default List<SysDictType> selectDictTypeList(SysDictType dictType)
    {
        java.util.Map<String, Object> params = dictType.getParams();
        return selectList(new LambdaQueryWrapperX<SysDictType>()
                .likeIfPresent(SysDictType::getDictName, dictType.getDictName())
                .likeIfPresent(SysDictType::getDictType, dictType.getDictType())
                .eqIfPresent(SysDictType::getStatus, dictType.getStatus())
                .betweenIfPresent(SysDictType::getCreateTime,
                        params != null ? params.get("beginTime") : null,
                        params != null ? params.get("endTime") : null)
                .orderByAsc(SysDictType::getDictId));
    }

    /**
     * 根据所有字典类型
     *
     * @return 字典类型集合信息
     */
    default List<SysDictType> selectDictTypeAll()
    {
        return selectList(new LambdaQueryWrapperX<SysDictType>().orderByAsc(SysDictType::getDictId));
    }

    /**
     * 根据字典类型查询信息
     *
     * @param dictType 字典类型
     * @return 字典类型
     */
    default SysDictType selectDictTypeByType(String dictType)
    {
        return selectOne(new LambdaQueryWrapperX<SysDictType>().eq(SysDictType::getDictType, dictType).last("limit 1"));
    }
}
