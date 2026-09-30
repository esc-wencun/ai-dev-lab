<?php
declare(strict_types=1);

namespace app\service;

use PasswordService;
use think\facade\Db;

/**
 * 注册服务（对位经典版 SysRegisterService 校验链；错误以 msg 字符串返回，空串=成功——经典版原样形态）
 */
class RegisterService
{
    /**
     * 注册：验证码由控制器先行校验（CaptchaValidateFilter 对位），此处从空值/长度四连开始。
     * 成功记 sys_logininfor（status '0' 成功组、msg「注册成功」——仅成功记，经典版实锤）。
     */
    public static function register(string $loginName, string $password): string
    {
        $msg = '';
        if ($loginName === '') {
            $msg = '用户名不能为空';
        } elseif ($password === '') {
            $msg = '用户密码不能为空';
        } elseif (mb_strlen($password) < 5 || mb_strlen($password) > 20) {
            $msg = '密码长度必须在5到20个字符之间';
        } elseif (mb_strlen($loginName) < 2 || mb_strlen($loginName) > 20) {
            $msg = '账户长度必须在2到20个字符之间';
        } elseif (!UserService::checkLoginNameUnique($loginName)) {
            $msg = "保存用户'{$loginName}'失败，注册账号已存在";
        } else {
            $salt = PasswordService::randomSalt();
            $hash = PasswordService::encrypt($loginName, $password, $salt);
            $ok = (bool)Db::table('sys_user')->insert([
                'login_name'      => $loginName,
                'user_name'       => $loginName,
                'salt'            => $salt,
                'password'        => $hash,
                'pwd_update_date' => date('Y-m-d H:i:s'),
            ]);
            $msg = $ok ? '' : '注册失败,请联系系统管理人员';
            if ($ok) {
                // 对位 AsyncFactory.recordLogininfor：REGISTER 归 status '0' 成功组，msg「注册成功」
                Db::table('sys_logininfor')->insert([
                    'login_name'     => $loginName,
                    'ipaddr'         => request()->ip(),
                    'login_location' => '内网',
                    'browser'        => 'Unknown',
                    'os'             => 'Unknown',
                    'status'         => '0',
                    'msg'            => '注册成功',
                    'login_time'     => date('Y-m-d H:i:s'),
                ]);
            }
        }
        return $msg;
    }
}
