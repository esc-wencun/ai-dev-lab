package com.ruoyi.ai.config;

import lombok.extern.slf4j.Slf4j;

import lombok.RequiredArgsConstructor;

import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;
import com.ruoyi.ai.service.impl.AiChatStreamServiceImpl;

/**
 * AI 启动任务：孤儿占位消息清理（msg_status=0 且超 1 小时 → 3，服务重启残留）
 *
 * @author ruoyi
 */
@Component
@Slf4j
@RequiredArgsConstructor
public class AiStartupRunner implements ApplicationRunner
{
    
    private final AiChatStreamServiceImpl chatStreamService;

    @Override
    public void run(ApplicationArguments args)
    {
        try
        {
            chatStreamService.cleanupStaleGenerating();
        }
        catch (Exception e)
        {
            log.warn("AI 孤儿消息清理失败（不阻断启动）", e);
        }
    }
}
