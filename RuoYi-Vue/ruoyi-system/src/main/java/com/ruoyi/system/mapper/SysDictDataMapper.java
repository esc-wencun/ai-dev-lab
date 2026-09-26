package com.ruoyi.system.mapper;

import java.util.List;
import org.apache.ibatis.annotations.Param;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.domain.entity.SysDictData;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;

/**
 * 字典表 数据层
 *
 * @author ruoyi
 */
public interface SysDictDataMapper extends BaseMapper<SysDictData>
{
    /**
     * 分页查询字典数据（动态条件 Wrapper 组装，固定 dict_sort 升序）
     *
     * @param dictData 字典数据信息
     * @return 分页结果
     */
    default Page<SysDictData> selectDictDataPage(SysDictData dictData)
    {
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<SysDictData>()
                .eqIfPresent(SysDictData::getDictType, dictData.getDictType())
                .likeIfPresent(SysDictData::getDictLabel, dictData.getDictLabel())
                .eqIfPresent(SysDictData::getStatus, dictData.getStatus())
                .orderByAsc(SysDictData::getDictSort));
    }

    /**
     * 根据条件查询字典数据列表（非分页：缓存加载/导出）
     *
     * @param dictData 字典数据信息
     * @return 字典数据集合信息
     */
    default List<SysDictData> selectDictDataList(SysDictData dictData)
    {
        return selectList(new LambdaQueryWrapperX<SysDictData>()
                .eqIfPresent(SysDictData::getDictType, dictData.getDictType())
                .likeIfPresent(SysDictData::getDictLabel, dictData.getDictLabel())
                .eqIfPresent(SysDictData::getStatus, dictData.getStatus())
                .orderByAsc(SysDictData::getDictSort));
    }

    /**
     * 根据字典类型查询字典数据（status=0，固定 dict_sort 升序）
     *
     * @param dictType 字典类型
     * @return 字典数据集合信息
     */
    default List<SysDictData> selectDictDataByType(String dictType)
    {
        return selectList(new LambdaQueryWrapperX<SysDictData>()
                .eq(SysDictData::getStatus, "0")
                .eq(SysDictData::getDictType, dictType)
                .orderByAsc(SysDictData::getDictSort));
    }

    /**
     * 根据字典类型和字典键值查询字典标签
     *
     * @param dictType 字典类型
     * @param dictValue 字典键值
     * @return 字典标签
     */
    default String selectDictLabel(@Param("dictType") String dictType, @Param("dictValue") String dictValue)
    {
        SysDictData data = selectOne(new LambdaQueryWrapperX<SysDictData>()
                .eq(SysDictData::getDictType, dictType)
                .eq(SysDictData::getDictValue, dictValue)
                .last("limit 1"));
        return data != null ? data.getDictLabel() : null;
    }

    /**
     * 查询字典数据个数
     *
     * @param dictType 字典类型
     * @return 个数
     */
    default int countDictDataByType(String dictType)
    {
        return Math.toIntExact(selectCount(new LambdaQueryWrapperX<SysDictData>().eq(SysDictData::getDictType, dictType)));
    }

    /**
     * 同步修改字典类型
     *
     * @param oldDictType 旧字典类型
     * @param newDictType 新旧字典类型
     * @return 结果
     */
    default int updateDictDataType(@Param("oldDictType") String oldDictType, @Param("newDictType") String newDictType)
    {
        SysDictData data = new SysDictData();
        data.setDictType(newDictType);
        return update(data, new LambdaQueryWrapperX<SysDictData>().eq(SysDictData::getDictType, oldDictType));
    }
}
