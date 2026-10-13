package com.ruoyi.ai.service;

import java.util.List;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.domain.AiApiKey;

/**
 * AI密钥 服务接口
 *
 * @author ruoyi
 */
public interface IAiApiKeyService
{
    /**
     * 分页查询密钥列表（出口 api_key 脱敏）
     */
    Page<AiApiKey> selectKeyPage(AiApiKey key);

    /**
     * 按ID查询密钥（出口 api_key 脱敏）
     */
    AiApiKey selectKeyById(Long keyId);

    /**
     * 按厂商查密钥（下拉选择用，不分页，出口脱敏）
     */
    List<AiApiKey> selectKeysByProvider(Long providerId);

    /**
     * 新增密钥
     *
     * @return true=成功；false=同厂商下 keyName 或 apiKey 重复
     */
    boolean insertKey(AiApiKey key);

    /**
     * 修改密钥（禁改 providerId；出口不涉及）
     *
     * @return true=成功；false=keyName 或 apiKey 重复
     */
    boolean updateKey(AiApiKey key);

    /**
     * 批量软删密钥（有未删模型挂靠的密钥拒删）
     *
     * @return 拒删原因；null=成功
     */
    String deleteKeyByIds(Long[] keyIds);
}
