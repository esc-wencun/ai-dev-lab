<?php

// +----------------------------------------------------------------------
// | 缓存设置
// +----------------------------------------------------------------------

return [
    // 默认缓存驱动
    'default' => 'file',

    // 缓存连接方式配置
    'stores'  => [
        'file' => [
            // 驱动方式
            'type'       => 'File',
            // 缓存保存目录
            'path'       => '',
            // 缓存前缀
            'prefix'     => '',
            // 缓存有效期 0表示永久缓存
            'expire'     => 0,
            // 缓存标签前缀
            'tag_prefix' => 'tag:',
            // 序列化机制 例如 ['serialize', 'unserialize']
            'serialize'  => [],
        ],
        // 更多的缓存连接
    ],

    // Redis 连接（RuoYi-TP 自用：RedisCache 门面专用，业务代码禁止直接使用本客户端）
    'redis'   => [
        'host'     => env('RD_HOST', '127.0.0.1'),
        'port'     => (int)env('RD_PORT', 6379),
        'password' => env('RD_PASS', ''),
        'database' => (int)env('RD_DB', 1),
        // 空闲超时（对位经典版会话 30 分钟空闲过期语义的默认 TTL）
        'session_expire' => 1800,
    ],
];
