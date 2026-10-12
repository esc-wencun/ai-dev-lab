package com.ruoyi.ai.service.impl;

import com.ruoyi.ai.common.AiConstants;
import java.util.Arrays;
import java.util.List;
import lombok.RequiredArgsConstructor;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.ai.domain.AiProvider;
import com.ruoyi.ai.mapper.AiApiKeyMapper;
import com.ruoyi.ai.mapper.AiProviderMapper;
import com.ruoyi.ai.service.IAiProviderService;
import com.ruoyi.common.core.service.BaseService;
import com.ruoyi.common.utils.SecurityUtils;

/**
 * AI厂商 服务实现
 *
 * @author ruoyi
 */
@Service
@RequiredArgsConstructor
public class AiProviderServiceImpl extends BaseService<AiProviderMapper, AiProvider> implements IAiProviderService
{

    private final AiApiKeyMapper apiKeyMapper;

    @Override
    public Page<AiProvider> selectProviderPage(AiProvider provider)
    {
        return baseMapper.selectProviderPage(provider);
    }

    @Override
    public List<AiProvider> selectProviderAll()
    {
        return baseMapper.selectList(Wrappers.<AiProvider>lambdaQuery()
                .eq(AiProvider::getStatus, AiConstants.STATUS_NORMAL)
                .orderByAsc(AiProvider::getSortOrder));
    }

    @Override
    public AiProvider selectProviderById(Long providerId)
    {
        return baseMapper.selectById(providerId);
    }

    @Override
    public boolean insertProvider(AiProvider provider)
    {
        boolean unique = checkUnique(AiProvider::getProviderId, null, AiProvider::getProviderCode, provider.getProviderCode());
        if (!unique)
        {
            return false;
        }
        provider.setCreateBy(SecurityUtils.getUsername());
        return baseMapper.insert(provider) > 0;
    }

    @Override
    public boolean updateProvider(AiProvider provider)
    {
        boolean unique = checkUnique(AiProvider::getProviderId, provider.getProviderId(), AiProvider::getProviderCode, provider.getProviderCode());
        if (!unique)
        {
            return false;
        }
        provider.setUpdateBy(SecurityUtils.getUsername());
        return baseMapper.updateById(provider) > 0;
    }

    /**
     * 批量软删厂商。级联校验（设计文档坑 5）：任一厂商下存在未删密钥即整体拒绝，先删子再删父。
     * 注意：实体未标 @TableLogic，BaseMapper 不会自动过滤 del_flag——计数查询须显式加 del_flag='0'。
     */
    @Override
    @Transactional
    public String deleteProviderByIds(Long[] providerIds)
    {
        List<Long> ids = Arrays.asList(providerIds);
        List<AiApiKey> keys = apiKeyMapper.selectList(Wrappers.<AiApiKey>lambdaQuery()
                .in(AiApiKey::getProviderId, ids)
                .eq(AiApiKey::getDelFlag, AiConstants.DEL_FLAG_EXISTS));
        if (!keys.isEmpty())
        {
            return "厂商下存在 " + keys.size() + " 把密钥，请先删除密钥";
        }
        AiProvider del = new AiProvider();
        del.setDelFlag(AiConstants.DEL_FLAG_REMOVED);
        del.setUpdateBy(SecurityUtils.getUsername());
        baseMapper.update(del, Wrappers.<AiProvider>lambdaUpdate().in(AiProvider::getProviderId, ids));
        return null;
    }
}
