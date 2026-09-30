<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;

use app\service\UserService;
use AjaxResult;
use PasswordService;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 个人中心控制器（对位经典版 SysProfileController，无 @RequiresPermissions 实锤——全部仅登录态）
 *
 * 经典版 GET /edit 死端点（模板缺失）不复刻。
 * profile 页用户数据查 DB（TP 会话无 phone/email/sex，spec 特殊行为 2——行为差异无害）。
 */
class ProfileController extends \app\BaseController
{
    /** GET /system/user/profile：双 tab 页（数据查 DB） */
    public function index(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)($session['user_id'] ?? 0);
        $user = UserService::selectUserById($userId);
        if ($user === null) {
            throw new \BusinessException('用户不存在');
        }
        return Response::create('user/profile/profile', 'view')->assign([
            'user'      => $user,
            'deptName'  => (string)($user['dept']['dept_name'] ?? ''),
            'postGroup' => UserService::selectUserPostGroup($userId),
            'chrtype'   => \app\service\ConfigService::get('sys.account.chrtype', '0'),
        ]);
    }

    /** GET /system/user/profile/checkPassword：裸 boolean（DB verify） */
    public function checkPassword(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $user = UserService::selectUserById((int)($session['user_id'] ?? 0));
        $ok = false;
        if ($user !== null) {
            $ok = PasswordService::verify(
                (string)$user['login_name'],
                (string)$request->get('password', ''),
                (string)($user['salt'] ?? ''),
                (string)$user['password']
            );
        }
        return \think\Response::create($ok ? 'true' : 'false');
    }

    /** GET /system/user/profile/resetPwd：修改密码弹窗（DB 最新数据） */
    public function resetPwd(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        return Response::create('user/profile/resetPwd', 'view')->assign([
            'user'    => UserService::selectUserById((int)($session['user_id'] ?? 0)),
            'chrtype' => \app\service\ConfigService::get('sys.account.chrtype', '0'),
        ]);
    }

    /** POST /system/user/profile/resetPwd：换盐重加密 + 会话销毁强制重登录（spec 定稿，deviations 登记） */
    #[Log('重置密码', Log::UPDATE)]
    public function resetPwdSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $uuid = (string)($request->middleware('session_uuid') ?? '');
        $user = UserService::selectUserById((int)($session['user_id'] ?? 0));
        if ($user === null) {
            return AjaxResult::error('修改密码异常，请联系管理员');
        }
        $oldPassword = (string)$request->post('oldPassword', '');
        $newPassword = (string)$request->post('newPassword', '');
        if (!PasswordService::verify((string)$user['login_name'], $oldPassword, (string)($user['salt'] ?? ''), (string)$user['password'])) {
            return AjaxResult::error('修改密码失败，旧密码错误');
        }
        if (PasswordService::verify((string)$user['login_name'], $newPassword, (string)($user['salt'] ?? ''), (string)$user['password'])) {
            return AjaxResult::error('新密码不能与旧密码相同');
        }
        $salt = PasswordService::randomSalt();
        $hash = PasswordService::encrypt((string)$user['login_name'], $newPassword, $salt);
        if (UserService::resetUserPwd((int)$user['user_id'], $hash, $salt) > 0) {
            // 会话重建：销毁 Redis 会话 + 清 cookie，前端提示后跳 /logout（spec 特殊行为 1）
            \app\service\SessionService::destroy($uuid);
            return AjaxResult::success()->cookie(\app\service\SessionService::COOKIE_NAME, '', 0);
        }
        return AjaxResult::error('修改密码异常，请联系管理员');
    }

    /** POST /system/user/profile/update：仅四字段（userId 取会话防横向越权）；成功回写会话 userName */
    #[Log('个人信息', Log::UPDATE)]
    public function update(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $uuid = (string)($request->middleware('session_uuid') ?? '');
        $userId = (int)($session['user_id'] ?? 0);
        $loginName = (string)($session['loginName'] ?? '');

        $userName = trim((string)$request->post('userName', ''));
        if ($userName === '') {
            throw new \BusinessException('用户名称不能为空');
        }
        $email = trim((string)$request->post('email', ''));
        $phonenumber = trim((string)$request->post('phonenumber', ''));
        $sex = (string)$request->post('sex', '0');

        if ($phonenumber !== '' && !UserService::checkPhoneUnique($phonenumber, $userId)) {
            return AjaxResult::error("修改用户'{$loginName}'失败，手机号码已存在");
        }
        if ($email !== '' && !UserService::checkEmailUnique($email, $userId)) {
            return AjaxResult::error("修改用户'{$loginName}'失败，邮箱账号已存在");
        }
        if (UserService::updateUserInfo([
            'user_name'   => $userName,
            'email'       => $email,
            'phonenumber' => $phonenumber,
            'sex'         => $sex,
        ], $userId) > 0) {
            // 会话回写 userName（对位 setSysUser 刷新——不重查 DB）
            $session['userName'] = $userName;
            \app\service\SessionService::write($uuid, $session);
            return AjaxResult::success();
        }
        return AjaxResult::error();
    }

    /** GET /system/user/profile/avatar：cropper 弹窗页 */
    public function avatar(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        return Response::create('user/profile/avatar', 'view')->assign([
            'user' => UserService::selectUserById((int)($session['user_id'] ?? 0)),
        ]);
    }

    /** POST /system/user/profile/updateAvatar：multipart avatarfile → public/profile/avatar/{Y/m/d}/{uuid}.{ext} */
    #[Log('个人信息', Log::UPDATE)]
    public function updateAvatar(Request $request): Response
    {
        try {
            $file = $request->file('avatarfile');
            if ($file === null) {
                return AjaxResult::error();
            }
            $originalName = (string)$file->getOriginalName();
            $ext = strtolower(pathinfo($originalName, PATHINFO_EXTENSION));
            if (!in_array($ext, ['bmp', 'gif', 'jpg', 'jpeg', 'png'], true)) {
                return AjaxResult::error('上传图片格式不允许：' . $ext);
            }

            $session = $request->middleware('session') ?? [];
            $uuid = (string)($request->middleware('session_uuid') ?? '');
            $userId = (int)($session['user_id'] ?? 0);

            $baseDir = (string)config('profile.profile') . DIRECTORY_SEPARATOR . (string)config('profile.avatar');
            $subDir = date('Y') . '/' . date('m') . '/' . date('d');
            $dir = $baseDir . DIRECTORY_SEPARATOR . $subDir;
            if (!is_dir($dir)) {
                mkdir($dir, 0755, true);
            }
            $newName = bin2hex(random_bytes(16)) . '.' . $ext;
            $file->move($dir, $newName);
            $avatar = '/profile/avatar/' . $subDir . '/' . $newName;

            if (UserService::updateUserAvatar($avatar, $userId) > 0) {
                // 删旧头像文件（avatar 非空时按旧 URL 映射磁盘路径删——stripPrefix /profile 对位）
                $oldAvatar = (string)($session['avatar'] ?? '');
                if ($oldAvatar !== '') {
                    $oldPath = (string)config('profile.profile') . substr($oldAvatar, strlen('/profile'));
                    $oldPath = str_replace(['/', '\\'], DIRECTORY_SEPARATOR, $oldPath);
                    if (is_file($oldPath)) {
                        @unlink($oldPath);
                    }
                }
                // 会话 avatar 键回写（主框架头像立即生效）
                $session['avatar'] = $avatar;
                \app\service\SessionService::write($uuid, $session);
                return AjaxResult::success();
            }
            return AjaxResult::error();
        } catch (\Throwable $e) {
            return AjaxResult::error($e->getMessage());
        }
    }
}
