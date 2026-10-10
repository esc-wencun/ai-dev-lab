package com.ruoyi.ai.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.service.IAiModelService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import com.ruoyi.common.core.page.TableDataInfo;

/**
 * AI模型管理 Controller
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/model")
public class AiModelController extends BaseController
{
    @Autowired
    private IAiModelService modelService;

    /**
     * 分页查询模型列表
     */
    @PreAuthorize("@ss.hasPermi('ai:model:list')")
    @GetMapping("/list")
    public TableDataInfo list(AiModel model)
    {
        return getDataTable(modelService.selectModelPage(model));
    }

    /**
     * 模型明细
     */
    @PreAuthorize("@ss.hasPermi('ai:model:query')")
    @GetMapping("/{modelId}")
    public AjaxResult getInfo(@PathVariable Long modelId)
    {
        return success(modelService.selectModelById(modelId));
    }

    /**
     * 新增模型
     */
    @PreAuthorize("@ss.hasPermi('ai:model:add')")
    @PostMapping
    public AjaxResult add(@RequestBody AiModel model)
    {
        return toAjax(modelService.insertModel(model) ? 1 : 0);
    }

    /**
     * 修改模型（禁改所属厂商；校验密钥与厂商一致）
     */
    @PreAuthorize("@ss.hasPermi('ai:model:edit')")
    @PutMapping
    public AjaxResult edit(@RequestBody AiModel model)
    {
        return toAjax(modelService.updateModel(model) ? 1 : 0);
    }

    /**
     * 删除模型
     */
    @PreAuthorize("@ss.hasPermi('ai:model:remove')")
    @DeleteMapping("/{modelIds}")
    public AjaxResult remove(@PathVariable Long[] modelIds)
    {
        modelService.deleteModelByIds(modelIds);
        return success();
    }
}
