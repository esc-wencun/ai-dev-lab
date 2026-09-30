<?php
declare(strict_types=1);

use app\middleware\OperLog;
use PHPUnit\Framework\TestCase;

/**
 * OperLog 敏感字段排除纯逻辑单测（对位经典版 LogAspect.EXCLUDE_PROPERTIES）
 */
final class OperLogStripTest extends TestCase
{
    public function testStripsTopLevelSensitiveFields(): void
    {
        $params = [
            'username'  => 'admin',
            'password'  => 'secret123',
            'oldPassword' => 'a',
            'newPassword' => 'b',
            'confirmPassword' => 'c',
        ];
        $stripped = OperLog::stripSensitive($params);
        $this->assertSame('admin', $stripped['username']);
        $this->assertSame('******', $stripped['password']);
        $this->assertSame('******', $stripped['oldPassword']);
        $this->assertSame('******', $stripped['newPassword']);
        $this->assertSame('******', $stripped['confirmPassword']);
    }

    public function testStripsNestedFields(): void
    {
        $params = [
            'user' => ['loginName' => 'admin', 'password' => 'x'],
            'list' => [['password' => 'y'], ['name' => 'z']],
        ];
        $stripped = OperLog::stripSensitive($params);
        $this->assertSame('******', $stripped['user']['password']);
        $this->assertSame('******', $stripped['list'][0]['password']);
        $this->assertSame('z', $stripped['list'][1]['name']);
    }

    public function testKeepsNonSensitiveFields(): void
    {
        $params = ['passwd_hint' => 'len6', 'passwordStrength' => 'high'];
        $stripped = OperLog::stripSensitive($params);
        $this->assertSame('len6', $stripped['passwd_hint']);
        $this->assertSame('high', $stripped['passwordStrength']);
    }
}
