<?php
declare(strict_types=1);

/**
 * 密码方案（对位经典版 SysPasswordService + ShiroUtils.randomSalt）
 *
 * 经典版实测复现：md5('admin'.'admin123'.'111111') === '29c67a30398638269fe600f73a054934'
 * salt = 6 位 hex（对位 SecureRandomNumberGenerator.nextBytes(3).toHex()）
 */
final class PasswordService
{
    public static function encrypt(string $loginName, string $password, string $salt): string
    {
        return md5($loginName . $password . $salt);
    }

    public static function randomSalt(): string
    {
        return bin2hex(random_bytes(3));
    }

    public static function verify(string $loginName, string $password, string $salt, string $storedHash): bool
    {
        return hash_equals($storedHash, self::encrypt($loginName, $password, $salt));
    }
}
