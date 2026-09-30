<?php

// 代码生成配置（对位经典版 ruoyi-generator/src/main/resources/generator.yml 五项，原值照存）
return [
    // 作者（写入 gen_table.function_author）
    'author'         => 'ruoyi',
    // 默认生成包路径（字面照存：gen_table.package_name 列即契约，值只是存进 DB 的字符串）
    'packageName'    => 'com.ruoyi.system',
    // 自动去除表前缀，默认 false
    'autoRemovePre'  => false,
    // 表前缀（生成类名不会包含表前缀，多个用逗号分隔）
    'tablePrefix'    => 'sys_',
    // 是否允许生成文件覆盖到本地（仅经典版 genCode 端点消费；本版该端点按范围拍板排除，照存不用）
    'allowOverwrite' => false,
];
