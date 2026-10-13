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
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.mapper.AiApiKeyMapper;
import com.ruoyi.ai.mapper.AiModelMapper;
import com.ruoyi.ai.service.IAiModelService;
import com.ruoyi.common.core.service.BaseService;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.SecurityUtils;

/**
 * AI模型 服务实现
 *
 * @author ruoyi
 */
@Service
@RequiredArgsConstructor
public class AiModelServiceImpl extends BaseService<AiModelMapper, AiModel> implements IAiModelService
{

    private final AiApiKeyMapper keyMapper;

    @Override
    public Page<AiModel> selectModelPage(AiModel model)
    {
        return baseMapper.selectModelPage(model);
    }

    @Override
    public AiModel selectModelById(Long modelId)
    {
        return baseMapper.selectById(modelId);
    }

    @Override
    public boolean insertModel(AiModel model)
    {
        assertKeyProviderConsistent(model.getKeyId(), model.getProviderId());
        boolean unique = checkUnique(AiModel::getModelId, null, AiModel::getModelCode, model.getModelCode()) ;
        if (!unique)
        {
            return false;
        }
        // checkUnique 只按 modelCode 查，多通道唯一键还含 provider+key——同三元组的重复在这里补查
        Long dup = baseMapper.selectCount(Wrappers.<AiModel>lambdaQuery()
                .eq(AiModel::getProviderId, model.getProviderId())
                .eq(AiModel::getKeyId, model.getKeyId())
                .eq(AiModel::getModelCode, model.getModelCode()));
        if (dup != null && dup > 0)
        {
            return false;
        }
        model.setCreateBy(SecurityUtils.getUsername());
        return baseMapper.insert(model) > 0;
    }

    @Override
    public boolean updateModel(AiModel model)
    {
        AiModel db = baseMapper.selectById(model.getModelId());
        if (db == null)
        {
            throw new ServiceException("模型不存在或已删除");
        }
        if (!db.getProviderId().equals(model.getProviderId()))
        {
            throw new ServiceException("模型不允许修改所属厂商，请删除后重新创建");
        }
        assertKeyProviderConsistent(model.getKeyId(), model.getProviderId());
        Long dup = baseMapper.selectCount(Wrappers.<AiModel>lambdaQuery()
                .eq(AiModel::getProviderId, model.getProviderId())
                .eq(AiModel::getKeyId, model.getKeyId())
                .eq(AiModel::getModelCode, model.getModelCode())
                .ne(AiModel::getModelId, model.getModelId()));
        if (dup != null && dup > 0)
        {
            return false;
        }
        model.setUpdateBy(SecurityUtils.getUsername());
        return baseMapper.updateById(model) > 0;
    }

    @Override
    @Transactional
    public void deleteModelByIds(Long[] modelIds)
    {
        baseMapper.update(null, Wrappers.<AiModel>lambdaUpdate()
                .set(AiModel::getDelFlag, AiConstants.DEL_FLAG_REMOVED)
                .set(AiModel::getUpdateBy, SecurityUtils.getUsername())
                .in(AiModel::getModelId, Arrays.asList(modelIds)));
    }

    /** key 与厂商一致性（设计文档坑 3）：key 不存在 / key 已删 / key 挂别的厂商 → 业务异常 */
    private void assertKeyProviderConsistent(Long keyId, Long providerId)
    {
        AiApiKey key = keyMapper.selectById(keyId);
        if (key == null || AiConstants.DEL_FLAG_REMOVED.equals(key.getDelFlag()))
        {
            throw new ServiceException("绑定的密钥不存在，请先在密钥管理中创建");
        }
        if (!key.getProviderId().equals(providerId))
        {
            throw new ServiceException("密钥与厂商不一致：该密钥属于其他厂商，请换绑同厂商密钥");
        }
    }
}
