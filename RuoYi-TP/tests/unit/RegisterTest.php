<?php
declare(strict_types=1);

use app\service\RegisterService;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 8.0.0 纯逻辑单测：注册校验链文案（不连库分支——空值/长度四连先行）
 */
final class RegisterTest extends TestCase
{
    public function testEmptyLoginName(): void
    {
        $this->assertSame('用户名不能为空', RegisterService::register('', 'x12345'));
    }

    public function testEmptyPassword(): void
    {
        $this->assertSame('用户密码不能为空', RegisterService::register('abc', ''));
    }

    public function testPasswordTooShort(): void
    {
        $this->assertSame('密码长度必须在5到20个字符之间', RegisterService::register('abc', 'a123'));
    }

    public function testPasswordTooLong(): void
    {
        $this->assertSame('密码长度必须在5到20个字符之间', RegisterService::register('abc', str_repeat('a', 21)));
    }

    public function testLoginNameTooShort(): void
    {
        $this->assertSame('账户长度必须在2到20个字符之间', RegisterService::register('a', 'a12345'));
    }

    public function testLoginNameTooLong(): void
    {
        $this->assertSame('账户长度必须在2到20个字符之间', RegisterService::register(str_repeat('a', 21), 'a12345'));
    }
}
