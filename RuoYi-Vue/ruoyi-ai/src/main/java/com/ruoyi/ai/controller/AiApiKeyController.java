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
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.ai.service.IAiApiKeyService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import com.ruoyi.common.core.page.TableDataInfo;

/**
 * AI密钥管理 Controller
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/key")
public class AiApiKeyController extends BaseController
{
    @Autowired
    private IAiApiKeyService keyService;

    /**
     * 分页查询密钥列表（出口已脱敏）
     */
    @PreAuthorize("@ss.hasPermi('ai:key:list')")
    @GetMapping("/list")
    public TableDataInfo list(AiApiKey key)
    {
        return getDataTable(keyService.selectKeyPage(key));
    }

    /**
     * 按厂商查启用密钥（下拉选择用，不分页，出口已脱敏）
     */
    @GetMapping("/listByProvider/{providerId}")
    public AjaxResult listByProvider(@PathVariable Long providerId)
    {
        return success(keyService.selectKeysByProvider(providerId));
    }

    /**
     * 密钥明细（出口已脱敏；编辑时 api_key 原样提交覆盖）
     */
    @PreAuthorize("@ss.hasPermi('ai:key:query')")
    @GetMapping("/{keyId}")
    public AjaxResult getInfo(@PathVariable Long keyId)
    {
        return success(keyService.selectKeyById(keyId));
    }

    /**
     * 新增密钥
     */
    @PreAuthorize("@ss.hasPermi('ai:key:add')")
    @PostMapping
    public AjaxResult add(@RequestBody AiApiKey key)
    {
        return toAjax(keyService.insertKey(key) ? 1 : 0);
    }

    /**
     * 修改密钥（禁改所属厂商）
     */
    @PreAuthorize("@ss.hasPermi('ai:key:edit')")
    @PutMapping
    public AjaxResult edit(@RequestBody AiApiKey key)
    {
        return toAjax(keyService.updateKey(key) ? 1 : 0);
    }

    /**
     * 删除密钥（有未删模型拒删）
     */
    @PreAuthorize("@ss.hasPermi('ai:key:remove')")
    @DeleteMapping("/{keyIds}")
    public AjaxResult remove(@PathVariable Long[] keyIds)
    {
        String rejectReason = keyService.deleteKeyByIds(keyIds);
        return rejectReason != null ? error(rejectReason) : success();
    }
}
