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
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.ai.domain.AiMessage;
import com.ruoyi.ai.mapper.AiMessageMapper;
import com.ruoyi.ai.service.IAiConversationService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;

/**
 * AI 会话接口。功能门禁统一 ai:chat:chat（菜单 2020 perms，进入对话页即授权）；
 * 数据隔离不靠权限点、靠 service 层归属校验（getOwnedConversation）——双层。
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/conversations")
@RequiredArgsConstructor
public class AiConversationController extends BaseController
{
    private final IAiConversationService conversationService;

    private final AiMessageMapper messageMapper;

    /** 我的会话列表 */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @GetMapping
    public AjaxResult list()
    {
        List<AiConversation> list = conversationService.selectMyConversations();
        return success(list);
    }

    /** 新建会话（title/modelId/agentId 均可空） */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PostMapping
    public AjaxResult add(@RequestBody AiConversation body)
    {
        AiConversation conv = conversationService.createConversation(
                body.getTitle(), body.getModelId(), body.getAgentId());
        return success(conv);
    }

    /** 重命名 */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PutMapping("/{conversationId}/title")
    public AjaxResult rename(@PathVariable Long conversationId, @RequestBody AiConversation body)
    {
        conversationService.renameConversation(conversationId, body.getTitle());
        return success();
    }

    /** 置顶/取消 */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PutMapping("/{conversationId}/pin")
    public AjaxResult pin(@PathVariable Long conversationId, @RequestBody AiConversation body)
    {
        conversationService.pinConversation(conversationId, Boolean.TRUE.equals(body.getPinned()));
        return success();
    }

    /** 切换会话模型 */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PutMapping("/{conversationId}/model")
    public AjaxResult changeModel(@PathVariable Long conversationId, @RequestBody AiConversation body)
    {
        conversationService.changeModel(conversationId, body.getModelId());
        return success();
    }

    /** 会话推理参数（temperature/maxTokens/maxContexts；null 字段不动） */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PutMapping("/{conversationId}/params")
    public AjaxResult updateParams(@PathVariable Long conversationId, @RequestBody AiConversation body)
    {
        conversationService.updateParams(conversationId, body.getTemperature(), body.getMaxTokens(), body.getMaxContexts());
        return success();
    }

    /** 删除（连带消息软删） */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @DeleteMapping("/{conversationId}")
    public AjaxResult remove(@PathVariable Long conversationId)
    {
        conversationService.deleteConversation(conversationId);
        return success();
    }

    /** 会话历史消息（升序） */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @GetMapping("/{conversationId}/messages")
    public AjaxResult messages(@PathVariable Long conversationId)
    {
        conversationService.getOwnedConversation(conversationId); // 归属校验
        return success(messageMapper.selectByConversation(conversationId));
    }
}
