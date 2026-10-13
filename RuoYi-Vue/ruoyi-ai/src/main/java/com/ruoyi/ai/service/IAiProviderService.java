package com.ruoyi.ai.service;

import java.util.List;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.domain.AiProvider;

/**
 * AI厂商 服务接口
 *
 * @author ruoyi
 */
public interface IAiProviderService
{
    /**
     * 分页查询厂商列表
     */
    Page<AiProvider> selectProviderPage(AiProvider provider);

    /**
     * 查询全部正常状态厂商（下拉选择用，不分页）
     */
    List<AiProvider> selectProviderAll();

    /**
     * 按ID查询厂商
     */
    AiProvider selectProviderById(Long providerId);

    /**
     * 新增厂商
     *
     * @return true=成功；false=providerCode 重复
     */
    boolean insertProvider(AiProvider provider);

    /**
     * 修改厂商
     *
     * @return true=成功；false=providerCode 重复
     */
    boolean updateProvider(AiProvider provider);

    /**
     * 批量软删厂商（有未删密钥的厂商拒删）
     *
     * @return 拒删原因；null=成功
     */
    String deleteProviderByIds(Long[] providerIds);
}
