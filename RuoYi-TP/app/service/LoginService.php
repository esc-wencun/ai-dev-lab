<?php
declare(strict_types=1);

namespace app\service;

use PasswordService;
use RedisCache;
use TpConstant;
use think\facade\Db;

/**
 * 登录服务（对位经典版 SysLoginController.ajaxLogin + SysPasswordService + SysLoginService）
 *
 * 全链路：验证码校验 → 用户查询 → 锁定检查 → 密码验证 → 失败计数/成功清零 →
 * 匿名会话转正（装载 permissions/roles）→ sys_logininfor 落库。
 * 文案对位经典版 messages.properties（已逐条核对）。
 */
final class LoginService
{
    public const MSG_CAPTCHA_ERROR = '验证码错误';
    public const MSG_NOT_MATCH     = '用户不存在/密码错误';
    public const MSG_LOCKED_FMT    = '密码输入错误%d次，帐户锁定10分钟';
    public const MSG_RETRY_FMT     = '密码输入错误%d次';
    public const MSG_DISABLED      = '账号已停用';
    public const MSG_DELETED       = '对不起，您的账号已删除';
    public const MSG_SUCCESS       = '登录成功';

    /**
     * 登录。成功返回匿名会话 uuid（已转正为完整会话）；失败抛 BusinessException（msg 面向页面展示）。
     *
     * @param string|null $anonUuid 匿名会话 uuid（cookie 携带）
     */
    public static function login(string $username, string $password, string $validateCode, bool $rememberMe, ?string $anonUuid): string
    {
        $ip = request()->ip();

        // 1. 验证码（开关来自 config，对位 yml；校验失败也记登录日志）
        if (config('tp.captcha.enabled')) {
            $anon = $anonUuid !== null ? RedisCache::get(TpConstant::PREFIX_SESSION . $anonUuid) : null;
            $ok = is_array($anon) && CaptchaService::verify($validateCode, $anon);
            if ($ok) {
                SessionService::write((string)$anonUuid, $anon); // 用后即删（写回去掉 captcha 的会话）
            } else {
                self::recordLogininfor($username, $ip, false, self::MSG_CAPTCHA_ERROR);
                throw new \BusinessException(self::MSG_CAPTCHA_ERROR);
            }
        }

        // 2. 用户查询（login_name 定位；经典版按 loginName 查）
        $user = Db::table('sys_user')
            ->where('login_name', $username)
            ->where('del_flag', '0')
            ->find();

        // 3. 锁定检查（在密码验证前——经典版 SysPasswordService.validate 先查 retryCache）
        $retryKey = TpConstant::PREFIX_PWD_RETRY . $username;
        $retryCount = (int)(RedisCache::get($retryKey) ?? 0);
        if ($retryCount >= TpConstant::PWD_MAX_RETRY) {
            self::recordLogininfor($username, $ip, false, sprintf(self::MSG_LOCKED_FMT, TpConstant::PWD_MAX_RETRY));
            throw new \BusinessException(sprintf(self::MSG_LOCKED_FMT, TpConstant::PWD_MAX_RETRY));
        }

        // 4. 用户存在性 + 密码验证（统一文案对位 user.password.not.match）
        if ($user === null) {
            self::failWithCount($username, $ip, $retryKey, $retryCount, self::MSG_NOT_MATCH);
        }
        if ((string)$user['status'] !== '0') {
            self::recordLogininfor($username, $ip, false, self::MSG_DISABLED);
            throw new \BusinessException(self::MSG_DISABLED);
        }
        if (!PasswordService::verify($username, $password, (string)$user['salt'], (string)$user['password'])) {
            self::failWithCount($username, $ip, $retryKey, $retryCount, self::MSG_NOT_MATCH);
        }

        // 5. 成功：清零计数 + 记日志 + 会话转正
        RedisCache::delete($retryKey);
        self::recordLogininfor($username, $ip, true, self::MSG_SUCCESS);

        $ua = (string)request()->header('user-agent', '');
        $sessionData = [
            'user_id'     => (int)$user['user_id'],
            'loginName'   => $user['login_name'],
            'userName'    => $user['user_name'],
            'deptId'      => $user['dept_id'] !== null ? (int)$user['dept_id'] : null,
            'avatar'      => (string)$user['avatar'],
            'isAdmin'     => PermissionService::isAdmin(['user_id' => (int)$user['user_id']]),
            'permissions' => PermissionService::permissionsOf(['user_id' => (int)$user['user_id']]),
            'roles'       => self::rolesOf((int)$user['user_id']),
            'pwdUpdateDate' => $user['pwd_update_date'],
            'rememberMe'  => $rememberMe,
            // 在线用户元数据（9.0.0；旧会话无此键显示「未知」并禁强退——向后兼容）
            'ip'          => $ip,
            'browser'     => self::parseBrowser($ua),
            'os'          => self::parseOs($ua),
            'loginTime'   => date('Y-m-d H:i:s'),
            'lastAccessTime' => date('Y-m-d H:i:s'),
        ];

        $uuid = $anonUuid ?? bin2hex(random_bytes(16));
        // rememberMe：cookie 30 天（会话 TTL 语义不变——deviations #17 降级版）
        cookie(SessionService::COOKIE_NAME, $uuid, $rememberMe && config('tp.rememberMe') ? 30 * 86400 : 0);
        SessionService::write($uuid, $sessionData);
        return $uuid;
    }

    /** 失败并计数（对位 SysPasswordService：先 incr 计数，达限出锁定文案，未达限出次数文案） */
    private static function failWithCount(string $username, string $ip, string $retryKey, int $current, string $baseMsg): never
    {
        $newCount = RedisCache::incrWithExpire($retryKey, TpConstant::PWD_LOCK_SECONDS);
        if ($newCount >= TpConstant::PWD_MAX_RETRY) {
            $msg = sprintf(self::MSG_LOCKED_FMT, TpConstant::PWD_MAX_RETRY);
        } else {
            // 对位经典版：非锁定失败记"密码输入错误{n}次"？经典版 baseMsg 优先——保持「用户不存在/密码错误」
            $msg = $baseMsg;
        }
        self::recordLogininfor($username, $ip, false, $msg);
        throw new \BusinessException($msg);
    }

    /** 用户角色数组（供数据权限与会话）：role_id/roleKey/roleName/dataScope */
    private static function rolesOf(int $userId): array
    {
        return Db::table('sys_role r')
            ->join('sys_user_role ur', 'ur.role_id = r.role_id')
            ->where('ur.user_id', $userId)
            ->where('r.status', '0')
            ->where('r.del_flag', '0')
            ->field('r.role_id,r.role_key,r.role_name,r.data_scope')
            ->select()->toArray();
    }

    /** 登录日志（对位 AsyncFactory.recordLogininfor；status 0 成功 1 失败） */
    public static function recordLogininfor(string $username, string $ip, bool $success, string $msg): void
    {
        $ua = (string)(request()->header('user-agent', ''));
        Db::table('sys_logininfor')->insert([
            'login_name'     => mb_substr($username, 0, 50),
            'ipaddr'         => substr($ip, 0, 128),
            'login_location' => '内网',
            'browser'        => mb_substr(self::parseBrowser($ua), 0, 50),
            'os'             => mb_substr(self::parseOs($ua), 0, 50),
            'status'         => $success ? '0' : '1',
            'msg'            => mb_substr($msg, 0, 255),
            'login_time'     => date('Y-m-d H:i:s'),
        ]);
    }

    /** UA 简析浏览器（对位经典版 UserAgent.getBrowser 的常见值） */
    public static function parseBrowser(string $ua): string
    {
        return match (true) {
            str_contains($ua, 'Edg/')   => 'Edge',
            str_contains($ua, 'Chrome') => 'Chrome',
            str_contains($ua, 'Firefox') => 'Firefox',
            str_contains($ua, 'Safari') && str_contains($ua, 'Version/') => 'Safari',
            default => 'Unknown',
        };
    }

    /** UA 简析操作系统 */
    public static function parseOs(string $ua): string
    {
        return match (true) {
            str_contains($ua, 'Windows NT 10') => 'Windows 10/11',
            str_contains($ua, 'Windows')       => 'Windows',
            str_contains($ua, 'Mac OS X')      => 'macOS',
            str_contains($ua, 'Android')       => 'Android',
            str_contains($ua, 'iPhone|iPad')   => 'iOS',
            default => 'Unknown',
        };
    }
}
