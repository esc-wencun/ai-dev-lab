<?php
// +----------------------------------------------------------------------
// | 个人中心配置（对位经典版 RuoYiConfig.profile / getAvatarPath）
// +----------------------------------------------------------------------
// 头像上传根目录 = public/profile（deviations #14 落地形态）；
// URL /profile/** 由 web server 对 public/ 的静态服务直接命中，无需读取端点。
// runtime_path() 是运行时入口注入的助手（返回 runtime/），不用于静态目录——用 root_path()。

return [
    'profile' => root_path() . 'public' . DIRECTORY_SEPARATOR . 'profile',
    // 头像子目录（对位 RuoYiConfig.getAvatarPath() = profile + "/avatar"）
    'avatar'  => 'avatar',
];
