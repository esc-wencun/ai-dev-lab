<?php
declare(strict_types=1);

use PHPUnit\Framework\TestCase;

/**
 * 密码方案单测（对位经典版 SysPasswordService.encryptPassword）
 * 官方 SQL 预置值实测：admin/admin123/salt=111111
 */
final class PasswordServiceTest extends TestCase
{
    public function testEncryptMatchesOfficialSeedHash(): void
    {
        $this->assertSame(
            '29c67a30398638269fe600f73a054934',
            PasswordService::encrypt('admin', 'admin123', '111111')
        );
    }

    public function testEncryptMatchesRySeedHash(): void
    {
        $this->assertSame(
            '8e6d98b90472783cc73c17047ddccf36',
            PasswordService::encrypt('ry', 'admin123', '222222')
        );
    }

    public function testVerifyAcceptsCorrectPassword(): void
    {
        $this->assertTrue(
            PasswordService::verify('admin', 'admin123', '111111', '29c67a30398638269fe600f73a054934')
        );
    }

    public function testVerifyRejectsWrongPassword(): void
    {
        $this->assertFalse(
            PasswordService::verify('admin', 'wrong', '111111', '29c67a30398638269fe600f73a054934')
        );
    }

    public function testRandomSaltIsSixHexChars(): void
    {
        for ($i = 0; $i < 20; $i++) {
            $salt = PasswordService::randomSalt();
            $this->assertMatchesRegularExpression('/^[0-9a-f]{6}$/', $salt);
        }
    }

    public function testRandomSaltIsRandom(): void
    {
        $salts = [];
        for ($i = 0; $i < 10; $i++) {
            $salts[] = PasswordService::randomSalt();
        }
        $this->assertGreaterThan(1, count(array_unique($salts)));
    }
}
