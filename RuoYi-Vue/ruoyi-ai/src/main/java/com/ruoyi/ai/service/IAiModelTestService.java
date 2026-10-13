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

    /**
     * 草稿模式测试（specs/2.0.3 用户需求）：模型尚未入库时（新增弹窗），按前端提交的
     * 厂商/密钥/模型标识走同一条解析链发起测试——解析、调用与判定与正式链路完全一致。
     *
     * @param providerId 厂商ID
     * @param keyId      密钥ID
     * @param modelCode  模型标识
     * @param modelName  模型名称（仅报错定位用，可空）
     * @return {success, durationMs, message, reply?}——失败时 message 带明确原因
     */
    Map<String, Object> testDraft(Long providerId, Long keyId, String modelCode, String modelName);
}
