<?php
// 全局中间件定义文件
return [
    // 登录态（对位经典版 Shiro 主链 user filter；匿名路径在中间件内放行）
    \app\middleware\LoginAuth::class,
];
