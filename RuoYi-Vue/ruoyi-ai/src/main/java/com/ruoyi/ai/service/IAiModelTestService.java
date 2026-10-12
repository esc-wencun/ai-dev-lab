package com.ruoyi.ai.service;

import java.util.Map;

/**
 * AI 模型连通性测试服务接口（specs/2.0.3 用户需求）
 *
 * @author ruoyi
 */
public interface IAiModelTestService
{
    /**
     * 按模型解析链路发起一次最小对话请求，验证 厂商地址/密钥/模型标识 全链路连通。
     *
     * @param modelId 模型ID
     * @return {success, durationMs, message, reply?}——失败时 message 带明确原因
     */
    Map<String, Object> testConnectivity(Long modelId);
}
