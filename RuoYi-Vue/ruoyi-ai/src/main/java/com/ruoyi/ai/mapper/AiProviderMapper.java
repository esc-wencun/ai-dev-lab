package com.ruoyi.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.common.AiConstants;
import com.ruoyi.ai.domain.AiProvider;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;

/**
 * AI厂商 数据层
 *
 * @author ruoyi
 */
public interface AiProviderMapper extends BaseMapper<AiProvider>
{
    /**
     * 分页查询厂商列表
     *
     * @param provider 厂商信息（providerCode/providerName 模糊、status 精确，值空自动跳过）
     * @return 分页结果
     */
    default Page<AiProvider> selectProviderPage(AiProvider provider)
    {
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<AiProvider>()
                .likeIfPresent(AiProvider::getProviderCode, provider.getProviderCode())
                .likeIfPresent(AiProvider::getProviderName, provider.getProviderName())
                .eqIfPresent(AiProvider::getStatus, provider.getStatus())
                .orderByAsc(AiProvider::getSortOrder));
    }
}
