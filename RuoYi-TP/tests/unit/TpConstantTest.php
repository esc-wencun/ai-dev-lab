<?php
declare(strict_types=1);

use PHPUnit\Framework\TestCase;

/**
 * Redis 键常量与拼装约定单测（不连真 Redis）
 * 约定：完整键 = PREFIX_* 常量 + 业务后缀，由 RedisCache 收口。
 */
final class TpConstantTest extends TestCase
{
    public function testCodeSystemMatchesClassicRuoYi(): void
    {
        $this->assertSame(0, TpConstant::CODE_SUCCESS);
        $this->assertSame(301, TpConstant::CODE_WARN);
        $this->assertSame(500, TpConstant::CODE_ERROR);
    }

    public function testPasswordRetryPolicyMatchesClassic(): void
    {
        // 对位 yml: maxRetryCount: 5 + ehcache loginRecordTemplate tti 10min
        $this->assertSame(5, TpConstant::PWD_MAX_RETRY);
        $this->assertSame(600, TpConstant::PWD_LOCK_SECONDS);
    }

    public function testSessionIdleTimeoutMatchesClassic(): void
    {
        // 对位 yml: shiro.session.expireTime: 30（分钟，空闲超时）
        $this->assertSame(1800, TpConstant::SESSION_IDLE_SECONDS);
        $this->assertSame(1200, TpConstant::SESSION_RENEW_THRESHOLD);
    }

    public function testKeyPrefixesAreDistinctAndColonSuffixed(): void
    {
        $prefixes = [
            TpConstant::PREFIX_SESSION,
            TpConstant::PREFIX_PWD_RETRY,
            TpConstant::PREFIX_REPEAT_SUBMIT,
            TpConstant::PREFIX_RATE_LIMIT,
            TpConstant::PREFIX_CONFIG,
            TpConstant::PREFIX_DICT,
        ];
        // 互不相同
        $this->assertSame(count($prefixes), count(array_unique($prefixes)));
        // 全部以冒号结尾（键拼装约定）
        foreach ($prefixes as $p) {
            $this->assertStringEndsWith(':', $p, "prefix $p must end with ':'");
        }
    }
}
