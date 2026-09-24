package com.ruoyi.web.controller.system;

import java.util.HashMap;
import java.util.Map;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import com.ruoyi.common.config.RuoYiConfig;
import com.ruoyi.common.core.domain.AjaxResult;

/**
 * 平台标识
 *
 * 提供语言、框架版本与功能开关，前端据此对 Java 特有功能（Druid 数据监控等）
 * 做降级提示；多语言复刻版（Go/Python）按同一契约返回各自的能力声明。
 *
 * @author ruoyi
 */
@RestController
public class SysPlatformController
{
    /** 系统基础配置 */
    @Autowired
    private RuoYiConfig ruoyiConfig;

    /**
     * 获取平台信息（登录即可访问，无 @PreAuthorize）
     */
    @GetMapping("/getPlatformInfo")
    public AjaxResult getPlatformInfo()
    {
        AjaxResult ajax = AjaxResult.success();
        ajax.put("framework", ruoyiConfig.getName());
        ajax.put("version", ruoyiConfig.getVersion());
        ajax.put("language", "java");
        ajax.put("languageVersion", System.getProperty("java.version"));
        Map<String, Boolean> features = new HashMap<>();
        features.put("druidMonitor", true);
        features.put("serverMonitor", true);
        features.put("swaggerDocs", true);
        ajax.put("features", features);
        return ajax;
    }
}
