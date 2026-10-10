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
import com.ruoyi.ai.domain.AiProvider;
import com.ruoyi.ai.service.IAiProviderService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import com.ruoyi.common.core.page.TableDataInfo;

/**
 * AI厂商管理 Controller
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/provider")
public class AiProviderController extends BaseController
{
    @Autowired
    private IAiProviderService providerService;

    /**
     * 分页查询厂商列表
     */
    @PreAuthorize("@ss.hasPermi('ai:provider:list')")
    @GetMapping("/list")
    public TableDataInfo list(AiProvider provider)
    {
        return getDataTable(providerService.selectProviderPage(provider));
    }

    /**
     * 全部正常厂商（下拉选择用，不分页）
     */
    @GetMapping("/selectAll")
    public AjaxResult selectAll()
    {
        return success(providerService.selectProviderAll());
    }

    /**
     * 厂商明细
     */
    @PreAuthorize("@ss.hasPermi('ai:provider:query')")
    @GetMapping("/{providerId}")
    public AjaxResult getInfo(@PathVariable Long providerId)
    {
        return success(providerService.selectProviderById(providerId));
    }

    /**
     * 新增厂商
     */
    @PreAuthorize("@ss.hasPermi('ai:provider:add')")
    @PostMapping
    public AjaxResult add(@RequestBody AiProvider provider)
    {
        return toAjax(providerService.insertProvider(provider) ? 1 : 0);
    }

    /**
     * 修改厂商
     */
    @PreAuthorize("@ss.hasPermi('ai:provider:edit')")
    @PutMapping
    public AjaxResult edit(@RequestBody AiProvider provider)
    {
        return toAjax(providerService.updateProvider(provider) ? 1 : 0);
    }

    /**
     * 删除厂商（有未删密钥拒删）
     */
    @PreAuthorize("@ss.hasPermi('ai:provider:remove')")
    @DeleteMapping("/{providerIds}")
    public AjaxResult remove(@PathVariable Long[] providerIds)
    {
        String rejectReason = providerService.deleteProviderByIds(providerIds);
        return rejectReason != null ? error(rejectReason) : success();
    }
}
