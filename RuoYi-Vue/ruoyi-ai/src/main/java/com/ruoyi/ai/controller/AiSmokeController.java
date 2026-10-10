package com.ruoyi.ai.controller;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.prompt.ChatOptions;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import com.ruoyi.ai.resolver.AiModelResolver;
import com.ruoyi.ai.resolver.ChatModelRouter;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.StringUtils;

/**
 * AI 冒烟验证（specs/2.0.0 引入，2.0.1 改造为按 modelId 走 DB 配置调用）——
 * 临时端点，2.1.0 做正式会话接口时重写或删除。
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai")
public class AiSmokeController extends BaseController
{
    private final ChatClient.Builder chatClientBuilder;

    @Autowired
    private AiModelResolver modelResolver;

    @Autowired
    private ChatModelRouter chatModelRouter;

    public AiSmokeController(ChatClient.Builder chatClientBuilder)
    {
        this.chatClientBuilder = chatClientBuilder;
    }

    /**
     * 登录态对话冒烟：入参 {"modelId": 100, "prompt": "..."}。
     * 路径 A（modelId 有值）：AiModelResolver 三级取数校验 → ChatModelRouter 取或建实例 → 调用；
     * 路径 B（modelId 为空）：回落 application.yml 的 spring.ai 全局配置（2.0.0 行为保留）。
     * 本期不加 @PreAuthorize 权限点（specs/2.0.0：避免建权限表数据），仅登录态校验。
     */
    @PostMapping("/chat")
    public AjaxResult chat(@RequestBody AiChatBody body)
    {
        if (body == null || StringUtils.isEmpty(body.getPrompt()))
        {
            return error("prompt 不能为空");
        }
        String answer;
        if (body.getModelId() != null)
        {
            AiModelResolver.ResolvedAiModel resolved = modelResolver.resolve(body.getModelId());
            answer = ChatClient.builder(chatModelRouter.getChatModel(resolved)).build()
                    .prompt()
                    .user(body.getPrompt())
                    .call()
                    .content();
        }
        else
        {
            answer = chatClientBuilder.build()
                    .prompt()
                    .options(ChatOptions.builder().model(getDefaultModel()))
                    .user(body.getPrompt())
                    .call()
                    .content();
        }
        // 答案放 data（规范信封）。注意不能写 success(answer)：String 实参会绑定到
        // success(String message) 重载把答案塞进 msg——必须用双参版本显式放 data
        return AjaxResult.success("操作成功", answer);
    }

    private String getDefaultModel()
    {
        String model = System.getenv("glm_model");
        return (model == null || model.isBlank()) ? "glm-4.7" : model;
    }

    /** 请求体 */
    public static class AiChatBody
    {
        /** 模型 ID（DB ai_model 主键）；null = 走 yml 全局配置兜底 */
        private Long modelId;

        private String prompt;

        public Long getModelId()
        {
            return modelId;
        }

        public void setModelId(Long modelId)
        {
            this.modelId = modelId;
        }

        public String getPrompt()
        {
            return prompt;
        }

        public void setPrompt(String prompt)
        {
            this.prompt = prompt;
        }
    }
}
