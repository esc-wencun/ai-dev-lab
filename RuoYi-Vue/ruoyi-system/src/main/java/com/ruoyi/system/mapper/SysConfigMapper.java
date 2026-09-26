package com.ruoyi.system.mapper;

import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;
import com.ruoyi.system.domain.SysConfig;

/**
 * 参数配置 数据层
 *
 * @author ruoyi
 */
public interface SysConfigMapper extends BaseMapper<SysConfig>
{
    /**
     * 分页查询参数配置列表（动态条件 Wrapper 组装；beginTime/endTime 时间区间走 betweenIfPresent）
     *
     * @param config 参数配置信息
     * @return 分页结果
     */
    default Page<SysConfig> selectConfigPage(SysConfig config)
    {
        java.util.Map<String, Object> params = config.getParams();
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<SysConfig>()
                .likeIfPresent(SysConfig::getConfigName, config.getConfigName())
                .likeIfPresent(SysConfig::getConfigKey, config.getConfigKey())
                .eqIfPresent(SysConfig::getConfigType, config.getConfigType())
                .betweenIfPresent(SysConfig::getCreateTime,
                        params != null ? params.get("beginTime") : null,
                        params != null ? params.get("endTime") : null));
    }

    /**
     * 查询参数配置信息（configId/configKey 条件，供 selectConfigByKey 缓存回源）
     *
     * @param config 参数配置信息
     * @return 参数配置信息
     */
    default SysConfig selectConfig(SysConfig config)
    {
        return selectOne(new LambdaQueryWrapperX<SysConfig>()
                .eqIfPresent(SysConfig::getConfigId, config.getConfigId())
                .eqIfPresent(SysConfig::getConfigKey, config.getConfigKey())
                .last("limit 1"));
    }

    /**
     * 查询参数配置列表（非分页：缓存加载/导出；beginTime/endTime 时间区间走 betweenIfPresent）
     *
     * @param config 参数配置信息
     * @return 参数配置集合
     */
    default List<SysConfig> selectConfigList(SysConfig config)
    {
        java.util.Map<String, Object> params = config.getParams();
        return selectList(new LambdaQueryWrapperX<SysConfig>()
                .likeIfPresent(SysConfig::getConfigName, config.getConfigName())
                .likeIfPresent(SysConfig::getConfigKey, config.getConfigKey())
                .eqIfPresent(SysConfig::getConfigType, config.getConfigType())
                .betweenIfPresent(SysConfig::getCreateTime,
                        params != null ? params.get("beginTime") : null,
                        params != null ? params.get("endTime") : null)
                .orderByAsc(SysConfig::getConfigId));
    }
}
