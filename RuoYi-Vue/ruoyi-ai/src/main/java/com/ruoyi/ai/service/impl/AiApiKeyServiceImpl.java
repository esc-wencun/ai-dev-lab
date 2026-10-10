package com.ruoyi.ai.service.impl;

import java.util.Arrays;
import java.util.List;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.mapper.AiApiKeyMapper;
import com.ruoyi.ai.mapper.AiModelMapper;
import com.ruoyi.ai.service.IAiApiKeyService;
import com.ruoyi.common.core.service.BaseService;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.SecurityUtils;
import com.ruoyi.common.utils.StringUtils;

/**
 * AI密钥 服务实现
 *
 * @author ruoyi
 */
@Service
public class AiApiKeyServiceImpl extends BaseService<AiApiKeyMapper, AiApiKey> implements IAiApiKeyService
{
    private static final String DEL_FLAG_REMOVED = "2";

    @Autowired
    private AiModelMapper modelMapper;

    @Override
    public Page<AiApiKey> selectKeyPage(AiApiKey key)
    {
        return maskKeys(baseMapper.selectKeyPage(key));
    }

    @Override
    public AiApiKey selectKeyById(Long keyId)
    {
        return maskKey(baseMapper.selectById(keyId));
    }

    @Override
    public List<AiApiKey> selectKeysByProvider(Long providerId)
    {
        List<AiApiKey> keys = baseMapper.selectList(Wrappers.<AiApiKey>lambdaQuery()
                .eq(AiApiKey::getProviderId, providerId)
                .eq(AiApiKey::getStatus, "0")
                .eq(AiApiKey::getDelFlag, "0")
                .orderByDesc(AiApiKey::getKeyId));
        for (AiApiKey k : keys)
        {
            k.setApiKey(mask(k.getApiKey()));
        }
        return keys;
    }

    @Override
    public boolean insertKey(AiApiKey key)
    {
        if (!checkKeyNameUnique(null, key))
        {
            return false;
        }
        if (!checkKeyContentUnique(null, key))
        {
            return false;
        }
        key.setCreateBy(SecurityUtils.getUsername());
        return baseMapper.insert(key) > 0;
    }

    /**
     * 编辑密钥。禁改 providerId（设计文档坑 3）：密钥挂模型后挪厂商会让模型静默打错厂商，
     * 挪厂商一律删了重建——providerId 与库中不一致直接抛业务异常。
     */
    @Override
    public boolean updateKey(AiApiKey key)
    {
        AiApiKey db = baseMapper.selectById(key.getKeyId());
        if (db == null)
        {
            throw new ServiceException("密钥不存在或已删除");
        }
        if (!db.getProviderId().equals(key.getProviderId()))
        {
            throw new ServiceException("密钥不允许修改所属厂商，请删除后重新创建");
        }
        if (!checkKeyNameUnique(key.getKeyId(), key))
        {
            return false;
        }
        if (!checkKeyContentUnique(key.getKeyId(), key))
        {
            return false;
        }
        key.setUpdateBy(SecurityUtils.getUsername());
        return baseMapper.updateById(key) > 0;
    }

    /**
     * 批量软删密钥。级联校验（设计文档坑 5）：任一密钥下存在未删模型即整体拒绝。
     */
    @Override
    @Transactional
    public String deleteKeyByIds(Long[] keyIds)
    {
        List<Long> ids = Arrays.asList(keyIds);
        Long modelCount = modelMapper.selectCount(Wrappers.<AiModel>lambdaQuery()
                .in(AiModel::getKeyId, ids)
                .eq(AiModel::getDelFlag, "0"));
        if (modelCount != null && modelCount > 0)
        {
            return "密钥下存在 " + modelCount + " 个模型，请先删除模型";
        }
        AiApiKey del = new AiApiKey();
        del.setDelFlag(DEL_FLAG_REMOVED);
        del.setUpdateBy(SecurityUtils.getUsername());
        baseMapper.update(del, Wrappers.<AiApiKey>lambdaUpdate().in(AiApiKey::getKeyId, ids));
        return null;
    }

    private boolean checkKeyNameUnique(Long keyId, AiApiKey key)
    {
        return checkUnique(AiApiKey::getKeyId, keyId, AiApiKey::getKeyName, key.getKeyName());
    }

    /** 同厂商内 api_key 内容查重（设计文档坑 7） */
    private boolean checkKeyContentUnique(Long keyId, AiApiKey key)
    {
        Long count = baseMapper.selectCount(Wrappers.<AiApiKey>lambdaQuery()
                .eq(AiApiKey::getProviderId, key.getProviderId())
                .eq(AiApiKey::getApiKey, key.getApiKey())
                .eq(AiApiKey::getDelFlag, "0")
                .ne(keyId != null, AiApiKey::getKeyId, keyId));
        return count == null || count == 0;
    }

    /** 出口脱敏：前4 + *** + 后4；${VAR} 占位符原样返回（管理页要能看到变量名） */
    static String mask(String apiKey)
    {
        if (StringUtils.isEmpty(apiKey) || apiKey.startsWith("${"))
        {
            return apiKey;
        }
        if (apiKey.length() <= 8)
        {
            return "***";
        }
        return apiKey.substring(0, 4) + "***" + apiKey.substring(apiKey.length() - 4);
    }

    private AiApiKey maskKey(AiApiKey key)
    {
        if (key != null)
        {
            key.setApiKey(mask(key.getApiKey()));
        }
        return key;
    }

    private Page<AiApiKey> maskKeys(Page<AiApiKey> page)
    {
        for (AiApiKey key : page.getRecords())
        {
            key.setApiKey(mask(key.getApiKey()));
        }
        return page;
    }
}
