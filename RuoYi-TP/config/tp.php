<?php

// RuoYi-TP 项目级配置（对位经典版 application.yml 的 shiro 段 + RuoYiConfig）
return [
    // 系统版本（对位 RuoYiConfig.version，pom.xml 4.8.3）
    'version'     => '4.8.3',

    // 验证码（对位 shiro.user.captchaEnabled / captchaType——经典版 sys_config 无 captcha 键，开关在配置文件）
    'captcha'     => [
        'enabled' => true,
        'type'    => 'math', // math 数字计算 / char 字符验证
    ],

    // 记住我（对位 shiro.rememberMe.enabled，yml 实际默认 true；TP 版降级语义见 deviations #17）
    'rememberMe'  => true,

    // 登录页（对位 shiro.user.loginUrl）
    'loginUrl'    => '/login',
];
