package com.ruoyi.ai.controller;

import java.util.List;
import lombok.RequiredArgsConstructor;

import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import com.ruoyi.ai.domain.AiAgent;
import com.ruoyi.ai.mapper.AiAgentMapper;
import com.ruoyi.ai.service.IAiAgentService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import com.ruoyi.common.utils.SecurityUtils;

/**
 * AI 智能体接口：下拉（登录态）+ 管理页 CRUD（权限点）
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/agents")
@RequiredArgsConstructor
public class AiAgentController extends BaseController
{
    private final AiAgentMapper agentMapper;

    private final IAiAgentService agentService;

    /** 智能体下拉：公开 + 本人私有（对话页角色下拉用，登录态即可） */
    @GetMapping("/options")
    public AjaxResult options()
    {
        return success(agentMapper.selectVisibleAgents(SecurityUtils.getUserId()));
    }

    /** 管理页列表 */
    @PreAuthorize("@ss.hasPermi('ai:agent:list')")
    @GetMapping
    public AjaxResult list(AiAgent query)
    {
        return success(agentMapper.selectAgentList(query));
    }

    @PreAuthorize("@ss.hasPermi('ai:agent:query')")
    @GetMapping("/{agentId}")
    public AjaxResult getInfo(@PathVariable Long agentId)
    {
        return success(agentMapper.selectById(agentId));
    }

    @PreAuthorize("@ss.hasPermi('ai:agent:add')")
    @PostMapping
    public AjaxResult add(@RequestBody AiAgent agent)
    {
        agentService.insertAgent(agent);
        return success();
    }

    @PreAuthorize("@ss.hasPermi('ai:agent:edit')")
    @PutMapping
    public AjaxResult edit(@RequestBody AiAgent agent)
    {
        agentService.updateAgent(agent);
        return success();
    }

    @PreAuthorize("@ss.hasPermi('ai:agent:remove')")
    @DeleteMapping("/{agentId}")
    public AjaxResult remove(@PathVariable Long agentId)
    {
        agentService.deleteAgent(agentId);
        return success();
    }
}
