<?php
declare(strict_types=1);

use app\task\TargetParser;
use app\task\TargetValidator;
use app\task\TaskRegistry;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 10.0.0 纯逻辑单测：调用目标解析器 / 类型推断 / 黑名单 / 注册表（Task 2）
 */
final class JobTargetTest extends TestCase
{
    /* ---------- 解析四形态 ---------- */

    public function testParseNoParams(): void
    {
        [$bean, $method, $params] = TargetParser::parse('ryTask.ryNoParams');
        $this->assertSame(['ryTask', 'ryNoParams', []], [$bean, $method, $params]);
    }

    public function testParseSingleParam(): void
    {
        [$bean, $method, $params] = TargetParser::parse("ryTask.ryParams('ry')");
        $this->assertSame('ryTask', $bean);
        $this->assertSame('ryParams', $method);
        $this->assertSame(['ry'], $params);
    }

    public function testParseMultipleParamsPreset(): void
    {
        // 预置任务 3 原样：'ry', true, 2000L, 316.50D, 100
        [, , $params] = TargetParser::parse("ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)");
        $this->assertSame(['ry', true, 2000, 316.5, 100], $params);
        $this->assertSame(['string', 'boolean', 'integer', 'double', 'integer'], array_map(fn($p) => gettype($p), $params));
    }

    public function testParseCommaInsideQuotes(): void
    {
        // 引号内逗号不切分（对位 Java 正则 lookahead）
        [, , $params] = TargetParser::parse("ryTask.ryParams('a,b,c')");
        $this->assertSame(['a,b,c'], $params);
    }

    public function testParseMalformedThrows(): void
    {
        $this->expectException(\BusinessException::class);
        TargetParser::parse('nodot');
    }

    /* ---------- 类型推断逐型 ---------- */

    public function testCastMatrix(): void
    {
        $this->assertSame('txt', TargetParser::cast("'txt'"));
        $this->assertSame('txt', TargetParser::cast('"txt"'));
        $this->assertTrue(TargetParser::cast('true'));
        $this->assertFalse(TargetParser::cast('FALSE'));
        $this->assertSame(2000, TargetParser::cast('2000L'));
        $this->assertSame(316.5, TargetParser::cast('316.50D'));
        $this->assertSame(100, TargetParser::cast('100'));
        $this->assertSame(1, TargetParser::cast('1'));
        $this->assertSame(3.14, TargetParser::cast('3.14'));
        $this->assertSame('abc', TargetParser::cast('abc'), '非数字裸串原样');
    }

    /* ---------- 黑名单/违规串 ---------- */

    public function testBlacklistRmi(): void
    {
        $this->assertSame("目标字符串不允许'rmi'调用", TargetValidator::check('rmi://evil/x'));
        $this->assertSame("目标字符串不允许'rmi'调用", TargetValidator::check('RMI://evil/x'), '大小写不敏感');
    }

    public function testBlacklistLdap(): void
    {
        $this->assertSame("目标字符串不允许'ldap(s)'调用", TargetValidator::check('ldap://evil'));
        $this->assertSame("目标字符串不允许'ldap(s)'调用", TargetValidator::check('ldaps://evil'));
    }

    public function testBlacklistHttp(): void
    {
        $this->assertSame("目标字符串不允许'http(s)'调用", TargetValidator::check('http://evil/x'));
        $this->assertSame("目标字符串不允许'http(s)'调用", TargetValidator::check('https://evil/x'));
    }

    public function testJobErrorStrings(): void
    {
        $this->assertSame('目标字符串存在违规', TargetValidator::check('java.net.URL/foo'));
        $this->assertSame('目标字符串存在违规', TargetValidator::check('org.springframework.xx'));
        $this->assertSame('目标字符串存在违规', TargetValidator::check('com.ruoyi.generator.util'));
    }

    public function testPassThrough(): void
    {
        $this->assertNull(TargetValidator::check("ryTask.ryParams('ry')"));
    }

    /* ---------- 注册表 ---------- */

    public function testRegistryResolvesPreset(): void
    {
        $this->assertSame(\app\task\RyTask::class, TaskRegistry::resolve('ryTask'));
        $this->assertSame('ryTask', TaskRegistry::beanOf("ryTask.ryMultipleParams('ry', true)"));
        $this->assertSame('ryTask', TaskRegistry::beanOf('ryTask.ryNoParams'));
    }

    public function testRegistryRejectsUnknown(): void
    {
        $this->expectException(\BusinessException::class);
        $this->expectExceptionMessage('目标字符串不在白名单内');
        TaskRegistry::resolve('evilTask');
    }
}
