package com.ruoyi.web.config;

import java.util.Collections;
import java.util.List;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import com.ruoyi.framework.mybatis.DataPermissionRule;
import com.ruoyi.framework.mybatis.DeptDataPermissionRule;

/**
 * 数据权限规则装配（1.0.0）。放 ruoyi-admin（组装层）：规则类在 framework，
 * 而 system 模块不能依赖 framework（依赖方向 framework → system）。
 *
 * <p>白名单当前为空表起步：现有走 BaseMapper 的表（post/config/dict_type/dict_data）
 * 均无部门/用户范围语义列。首个带范围语义的纯 MP 业务查询出现时，在此注册：
 *
 * <pre>
 *     rule.addDeptColumn(业务实体类.class);            // 实体推导表名，列默认 dept_id
 *     rule.addDeptColumn("业务表名", "dept_id");        // 或直接表名字符串
 *     rule.addUserColumn("业务表名", "user_id");
 * </pre>
 *
 * <p>纪律：注册即生效，任何白名单改动必须走 checklist 的数据权限受限角色实测。
 */
@Configuration
public class DataPermissionConfiguration
{
    @Bean
    public List<DataPermissionRule> dataPermissionRules()
    {
        DeptDataPermissionRule rule = new DeptDataPermissionRule();
        // 【1.0.0 端到端专测临时注册，测完删除此行——checklist 清理项】
        rule.addDeptColumn("biz_order", "dept_id");
        rule.addUserColumn("biz_order", "create_by_id");
        return java.util.Collections.singletonList(rule);
    }
}
