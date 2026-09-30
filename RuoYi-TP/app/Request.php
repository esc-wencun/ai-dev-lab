<?php
namespace app;

// 应用请求对象类
class Request extends \think\Request
{
    /**
     * 覆盖路径参数名：TP 默认 's'（兼容模式 URL 参数），与经典若依验证码 URL 的
     * 防缓存参数 ?s=<rand> 撞名（s=0.123 会被当成 pathinfo 导致路由 302）。
     * 改用 'r'，业务代码与第三方参数均不含 r=。
     */
    protected $varPathinfo = 'r';
}
