<?php
declare(strict_types=1);

namespace app\task;

/**
 * 调用目标黑名单校验（对位经典版 SysJobController.checkCronExpressionInside… 校验链 ②③④⑤）
 *
 * 黑名单/违规串字面量逐字照抄 Constants.java；命中返回错误文案片段（供控制器拼前缀），放行返回 null。
 */
class TargetValidator
{
    private const LOOKUP_RMI = 'rmi:';
    private const LOOKUP_LDAP = 'ldap:';
    private const LOOKUP_LDAPS = 'ldaps:';
    private const HTTP = 'http://';
    private const HTTPS = 'https://';

    /** 违规串（经典版 JOB_ERROR_STR 逐字照抄） */
    private const JOB_ERROR_STR = [
        'java.net.URL', 'javax.naming.InitialContext', 'org.yaml.snakeyaml',
        'org.springframework', 'org.apache', 'com.ruoyi.common.utils.file',
        'com.ruoyi.common.config', 'com.ruoyi.generator',
    ];

    /**
     * 校验链 ②③④⑤：rmi → ldap(s) → http(s) → 违规串。
     * 返回错误文案（含「目标字符串…」段）或 null（放行）。
     */
    public static function check(string $invokeTarget): ?string
    {
        if (stripos($invokeTarget, self::LOOKUP_RMI) !== false) {
            return "目标字符串不允许'rmi'调用";
        }
        if (stripos($invokeTarget, self::LOOKUP_LDAP) !== false || stripos($invokeTarget, self::LOOKUP_LDAPS) !== false) {
            return "目标字符串不允许'ldap(s)'调用";
        }
        if (stripos($invokeTarget, self::HTTP) !== false || stripos($invokeTarget, self::HTTPS) !== false) {
            return "目标字符串不允许'http(s)'调用";
        }
        foreach (self::JOB_ERROR_STR as $str) {
            if (stripos($invokeTarget, $str) !== false) {
                return '目标字符串存在违规';
            }
        }
        return null;
    }
}
