<?php
declare(strict_types=1);

namespace app\middleware;

use app\attribute\Log;
use Closure;
use think\Request;
use think\Response;

/**
 * 操作日志中间件（对位经典版 @Log + LogAspect）
 *
 * 反射 #[Log(title, businessType)]，请求完成后同步落库 sys_oper_log（deviations #13）。
 * 敏感字段排除对位 LogAspect.EXCLUDE_PROPERTIES；参数 JSON 截断 2000。
 */
class OperLog
{
    /** 对位经典版 LogAspect.EXCLUDE_PROPERTIES（键统一小写比较） */
    private const EXCLUDE_FIELDS = ['password', 'oldpassword', 'newpassword', 'confirmpassword'];

    public function handle(Request $request, Closure $next): Response
    {
        $start = microtime(true);
        $response = $next($request);

        try {
            $log = self::resolveLog($request);
            if ($log !== null) {
                self::save($request, $response, $log, (int)((microtime(true) - $start) * 1000));
            }
        } catch (\Throwable $e) {
            // 日志失败不影响主流程（对位经典版 LogAspect 捕获）
        }
        return $response;
    }

    private static function resolveLog(Request $request): ?Log
    {
        $controller = $request->controller();
        $action = $request->action();
        if ($controller === '' || $action === '') {
            return null;
        }
        // controller 形态 = 'system.Dept'（已含 layer 前缀，同 CheckPerm 口径）
        $class = 'app\\controller\\' . str_replace('.', '\\', $controller) . 'Controller';
        $method = $action;
        if (!class_exists($class) || !method_exists($class, $method)) {
            return null;
        }
        $attrs = (new \ReflectionMethod($class, $method))->getAttributes(Log::class);
        return $attrs ? $attrs[0]->newInstance() : null;
    }

    private static function save(Request $request, Response $response, Log $log, int $costMs): void
    {
        $sessionData = $request->middleware('session') ?? [];

        // 敏感字段排除
        $params = $request->param();
        $params = self::stripSensitive(is_array($params) ? $params : []);
        $operParam = mb_substr(json_encode($params, JSON_UNESCAPED_UNICODE) ?: '', 0, 2000);

        // 响应 body 截取（对位 json_result）
        $jsonResult = mb_substr($response->getContent() ?: '', 0, 2000);

        \think\facade\Db::table('sys_oper_log')->insert([
            'title'          => $log->title,
            'business_type'  => $log->businessType,
            'method'         => $request->controller() . '::' . $request->action(),
            'request_method' => $request->method(),
            'operator_type'  => 1,
            'oper_name'      => $sessionData['loginName'] ?? '',
            'dept_name'      => '',
            'oper_url'       => mb_substr('/' . ltrim($request->pathinfo(), '/'), 0, 255),
            'oper_ip'        => $request->ip(),
            'oper_location'  => '内网',
            'oper_param'     => $operParam,
            'json_result'    => $jsonResult,
            'status'         => 0,
            'error_msg'      => '',
            'oper_time'      => date('Y-m-d H:i:s'),
            'cost_time'      => $costMs,
        ]);
    }

    /** 敏感字段排除（PHPUnit 固化） */
    public static function stripSensitive(array $params): array
    {
        array_walk_recursive($params, function (&$v, $k) {
            if (in_array(strtolower((string)$k), self::EXCLUDE_FIELDS, true)) {
                $v = '******';
            }
        });
        return $params;
    }
}
