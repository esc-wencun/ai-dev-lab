<?php
declare(strict_types=1);

namespace app\task;

/**
 * Java 语法调用目标解析器（对位经典版 JobInvokeUtil：beanName/methodName/getMethodParams）
 *
 * 形态：`ryTask.ryParams('ry')` / `ryTask.ryNoParams`；
 * 切参正则对位 `,(?=(?:[^"']*["'][^"']*["'])*[^"']*$)`（引号内逗号不切分）；
 * 类型推断照抄：'x'/"x" → string；true/false → bool；L 结尾 → int；D 结尾 → float；纯数字 → int。
 */
class TargetParser
{
    /** 解析 invokeTarget：返回 [bean, method, params[]]；解析失败抛 BusinessException */
    public static function parse(string $invokeTarget): array
    {
        $head = $invokeTarget;
        $paramsStr = '';
        $paren = strpos($invokeTarget, '(');
        if ($paren !== false) {
            $head = substr($invokeTarget, 0, $paren);
            $close = strrpos($invokeTarget, ')');
            $paramsStr = $close !== false ? substr($invokeTarget, $paren + 1, $close - $paren - 1) : substr($invokeTarget, $paren + 1);
        }
        $dot = strrpos($head, '.');
        if ($dot === false || $dot === 0 || $dot === strlen($head) - 1) {
            throw new \BusinessException('调用目标字符串格式错误');
        }
        return [substr($head, 0, $dot), substr($head, $dot + 1), self::parseParams($paramsStr)];
    }

    /** 切参 + 类型推断（空串 → []） */
    public static function parseParams(string $paramsStr): array
    {
        $paramsStr = trim($paramsStr);
        if ($paramsStr === '') {
            return [];
        }
        // 对位 Java split(",(?=([^"']*["'][^"']*["'])*[^"']*$)")：引号内逗号不切分
        $chunks = preg_split('/,(?=(?:[^"\']*[\'"][^"\']*[\'"])*[^"\']*$)/', $paramsStr) ?: [];
        return array_map(fn(string $c) => self::cast(trim($c)), $chunks);
    }

    /** 单参类型推断（照抄 JobInvokeUtil 的后缀/字面量判定） */
    public static function cast(string $chunk): int|float|bool|string
    {
        if ($chunk === '') {
            return '';
        }
        $first = $chunk[0];
        if ($first === "'" || $first === '"') {
            return substr($chunk, 1, -1); // 去引号 string
        }
        if (strcasecmp($chunk, 'true') === 0) {
            return true;
        }
        if (strcasecmp($chunk, 'false') === 0) {
            return false;
        }
        if (preg_match('/^\d+L$/i', $chunk)) {
            return (int)substr($chunk, 0, -1); // 2000L → int
        }
        if (preg_match('/^\d+\.\d+D$/i', $chunk)) {
            return (float)substr($chunk, 0, -1); // 316.50D → float
        }
        if (is_numeric($chunk)) {
            return strpos($chunk, '.') !== false ? (float)$chunk : (int)$chunk;
        }
        return $chunk;
    }
}
