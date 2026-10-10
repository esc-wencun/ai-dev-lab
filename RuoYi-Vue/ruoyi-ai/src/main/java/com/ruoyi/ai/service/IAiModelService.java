package com.ruoyi.ai.service;

import java.util.List;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.domain.AiModel;

/**
 * AI模型 服务接口
 *
 * @author ruoyi
 */
public interface IAiModelService
{
    /**
     * 分页查询模型列表
     */
    Page<AiModel> selectModelPage(AiModel model);

    /**
     * 按ID查询模型
     */
    AiModel selectModelById(Long modelId);

    /**
     * 新增模型（校验 key 与 provider 一致）
     *
     * @return true=成功；false=同厂商+同密钥下 modelCode 重复
     */
    boolean insertModel(AiModel model);

    /**
     * 修改模型（校验 key 与 provider 一致；禁改 providerId）
     *
     * @return true=成功；false=modelCode 重复
     */
    boolean updateModel(AiModel model);

    /**
     * 批量软删模型
     */
    void deleteModelByIds(Long[] modelIds);
}
