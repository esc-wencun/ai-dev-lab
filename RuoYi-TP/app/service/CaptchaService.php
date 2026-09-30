<?php
declare(strict_types=1);

namespace app\service;

/**
 * 验证码服务（对位经典版 SysCaptchaController + kaptcha，math/char 双型）
 *
 * 答案存「匿名会话」（登录前预生成的 session uuid，对位经典版 Session attribute 语义）。
 * GET /login 时创建匿名会话并写 cookie；POST /login 校验后转正为完整会话（cookie 不变）。
 */
final class CaptchaService
{
    /** 生成验证码：返回 ['code' => 答案, 'image' => jpeg 二进制]；答案写入匿名会话 */
    public static function make(array &$anonSession): array
    {
        $type = config('tp.captcha.type', 'math');

        if ($type === 'math') {
            // 对位 kaptcha math：两个 10 以内数加减（结果非负）
            $a = random_int(1, 9);
            $b = random_int(1, 9);
            $ops = ['+', '-'];
            $op = $ops[random_int(0, 1)];
            if ($op === '-' && $b > $a) {
                [$a, $b] = [$b, $a];
            }
            $answer = $op === '+' ? $a + $b : $a - $b;
            $text = "{$a}{$op}{$b}=?";
            $code = (string)$answer;
        } else {
            // char 型：4 位大写字母+数字（去易混淆 0O1I）
            $chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
            $code = '';
            for ($i = 0; $i < 4; $i++) {
                $code .= $chars[random_int(0, strlen($chars) - 1)];
            }
            $text = $code;
        }

        $anonSession['captcha'] = $type === 'math' ? $code : strtoupper($code);
        return ['code' => $code, 'image' => self::render($text)];
    }

    /** 校验：忽略大小写，用后即删 */
    public static function verify(string $input, array &$anonSession): bool
    {
        $expected = $anonSession['captcha'] ?? null;
        unset($anonSession['captcha']);
        if ($expected === null) {
            return false;
        }
        return strtolower(trim($input)) === strtolower($expected);
    }

    /**
     * 当前请求的匿名会话 uuid（cookie；无则生成——由调用方保证响应时 Set-Cookie 同一 uuid，
     * 写会话统一走 SessionService::write($uuid, $data)，本方法只读不写。
     */
    public static function anonUuid(): string
    {
        $uuid = cookie(SessionService::COOKIE_NAME);
        if (!is_string($uuid) || strlen($uuid) !== 32) {
            $uuid = bin2hex(random_bytes(16));
        }
        return $uuid;
    }

    /** GD 渲染 jpeg：130x40 干扰弧线 + 居中文本（对位 kaptcha 视觉可读即可，deviations 同理） */
    private static function render(string $text): string
    {
        $w = 130;
        $h = 40;
        $img = imagecreate($w, $h);
        imagecolorallocate($img, 243, 246, 249); // 底色（经典版灰白）
        $border = imagecolorallocate($img, 204, 204, 204);
        imagerectangle($img, 0, 0, $w - 1, $h - 1, $border);

        // 干扰弧线
        $noise = imagecolorallocate($img, 170, 178, 189);
        for ($i = 0; $i < 5; $i++) {
            imagearc($img, random_int(0, $w), random_int(0, $h), random_int(20, 60), random_int(15, 45), random_int(0, 180), random_int(181, 360), $noise);
        }

        // 逐字符绘制（随机色 + 轻微 y 抖动）
        $colors = [
            imagecolorallocate($img, 52, 122, 225),
            imagecolorallocate($img, 199, 84, 80),
            imagecolorallocate($img, 73, 160, 96),
            imagecolorallocate($img, 128, 96, 190),
        ];
        $len = strlen($text);
        $charW = (int)(($w - 24) / max($len, 1));
        for ($i = 0; $i < $len; $i++) {
            $color = $colors[random_int(0, count($colors) - 1)];
            $x = 12 + $i * $charW + random_int(0, 2);
            $y = random_int(10, 16);
            imagestring($img, 5, $x, $y, $text[$i], $color);
        }

        ob_start();
        imagejpeg($img, null, 80);
        $jpeg = (string)ob_get_clean();
        imagedestroy($img);
        return $jpeg;
    }
}
