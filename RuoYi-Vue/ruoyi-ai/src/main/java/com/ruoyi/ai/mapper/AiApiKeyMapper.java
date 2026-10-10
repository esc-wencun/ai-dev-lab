package com.ruoyi.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;

/**
 * AI密钥 数据层
 *
 * @author ruoyi
 */
public interface AiApiKeyMapper extends BaseMapper<AiApiKey>
{
    /**
     * 分页查询密钥列表
     *
     * @param key 密钥信息（providerId 精确、keyName 模糊、status 精确，值空自动跳过）
     * @return 分页结果
     */
    default Page<AiApiKey> selectKeyPage(AiApiKey key)
    {
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<AiApiKey>()
                .eqIfPresent(AiApiKey::getProviderId, key.getProviderId())
                .likeIfPresent(AiApiKey::getKeyName, key.getKeyName())
                .eqIfPresent(AiApiKey::getStatus, key.getStatus())
                .orderByDesc(AiApiKey::getKeyId));
    }
}
