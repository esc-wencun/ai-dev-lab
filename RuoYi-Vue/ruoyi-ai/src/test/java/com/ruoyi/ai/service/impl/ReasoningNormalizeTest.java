package com.ruoyi.ai.service.impl;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

/**
 * D10 reasoning 归一化单测：GLM 网关发"累积全量"帧、qwen 发"真增量"帧，
 * normalizeReasoningDelta 统一输出真增量（前端 prev+chunk 拼接恒正确）。
 *
 * @author ruoyi
 */
public class ReasoningNormalizeTest
{
    @Test
    void cumulativeFramesYieldSuffixDeltas()
    {
        StringBuilder buf = new StringBuilder();
        assertEquals("The", ReasoningNormalizeTest.invoke(buf, "The"));
        assertEquals(" user", ReasoningNormalizeTest.invoke(buf, "The user"));
        assertEquals(" is", ReasoningNormalizeTest.invoke(buf, "The user is"));
        assertEquals("The user is", buf.toString());
    }

    @Test
    void incrementalFramesPassThrough()
    {
        StringBuilder buf = new StringBuilder();
        assertEquals("今", ReasoningNormalizeTest.invoke(buf, "今"));
        assertEquals("天", ReasoningNormalizeTest.invoke(buf, "天"));
        assertEquals("好", ReasoningNormalizeTest.invoke(buf, "好"));
        assertEquals("今天好", buf.toString());
    }

    @Test
    void duplicateFullFrameAfterIncrementalProducesEmptyDelta()
    {
        // 网关混发：增量流后跟一帧累积全量（罕见但防御）——全量已与 buffer 相等，delta 为空
        StringBuilder buf = new StringBuilder();
        ReasoningNormalizeTest.invoke(buf, "abc");
        ReasoningNormalizeTest.invoke(buf, "d");
        assertEquals("", ReasoningNormalizeTest.invoke(buf, "abcd"));
        assertEquals("abcd", buf.toString());
    }

    @Test
    void divergentFrameAppendsAsIncrement()
    {
        // 与前文无前缀关系（网关换流/异常帧）：按增量透传，不丢内容
        StringBuilder buf = new StringBuilder();
        ReasoningNormalizeTest.invoke(buf, "abc");
        assertEquals("xyz", ReasoningNormalizeTest.invoke(buf, "xyz"));
        assertEquals("abcxyz", buf.toString());
    }

    @Test
    void emptyBufferTreatsAnyFrameAsIncrement()
    {
        StringBuilder buf = new StringBuilder();
        assertEquals("abc", ReasoningNormalizeTest.invoke(buf, "abc"));
        assertEquals("abc", buf.toString());
    }

    private static String invoke(StringBuilder buf, String incoming)
    {
        String delta = AiChatStreamServiceImpl.normalizeReasoningDelta(buf, incoming);
        assertTrue(!incoming.isEmpty() || delta.isEmpty());
        return delta;
    }
}
