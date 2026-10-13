package com.ruoyi.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.ai.common.AiConstants;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;

/**
 * AI模型 数据层
 *
 * @author ruoyi
 */
public interface AiModelMapper extends BaseMapper<AiModel>
{
    /**
     * 分页查询模型列表
     *
     * @param model 模型信息（providerId/modelType/status 精确、modelName/modelCode 模糊，值空自动跳过）
     * @return 分页结果
     */
    default Page<AiModel> selectModelPage(AiModel model)
    {
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<AiModel>()
                .eqIfPresent(AiModel::getProviderId, model.getProviderId())
                .eqIfPresent(AiModel::getModelType, model.getModelType())
                .likeIfPresent(AiModel::getModelName, model.getModelName())
                .likeIfPresent(AiModel::getModelCode, model.getModelCode())
                .eqIfPresent(AiModel::getStatus, model.getStatus())
                .orderByAsc(AiModel::getSortOrder));
    }
}
