package com.ruoyi.framework.mybatis;

import java.util.List;
import net.sf.jsqlparser.expression.Expression;
import net.sf.jsqlparser.schema.Table;
import com.baomidou.mybatisplus.extension.plugins.handler.MultiDataPermissionHandler;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;
import com.ruoyi.common.core.mybatis.DataScopeContextHolder;

/**
 * MP 数据权限拦截器的逐表回调实现。
 *
 * <p>语义：无活跃上下文（未标 @DataScope / executeIgnore 之外）→ null 放行；
 * enable=false（executeIgnore）→ null 放行；DTO.all → null 放行；
 * 表不在任何规则白名单 → null 放行；命中 → 规则生成条件（含 null=null 空集兜底）。
 *
 * @author ruoyi
 */
public class RuoYiDataPermissionHandler implements MultiDataPermissionHandler
{
    private final List<DataPermissionRule> rules;

    public RuoYiDataPermissionHandler(List<DataPermissionRule> rules)
    {
        this.rules = rules;
    }

    @Override
    public Expression getSqlSegment(Table table, Expression where, String mappedStatementId)
    {
        DataScopeContextHolder.Scope scope = DataScopeContextHolder.peek();
        // 无活跃作用域 → 全放行（显式注解才过滤的 RuoYi 语义）
        if (scope == null)
        {
            return null;
        }
        // 豁免作用域
        if (!scope.isEnable())
        {
            return null;
        }
        if (rules == null || rules.isEmpty())
        {
            return null;
        }
        String tableName = stripBackticks(table.getName());
        RuoYiDeptDataPermissionDTO dto = scope.getDto();
        Expression combined = null;
        for (DataPermissionRule rule : rules)
        {
            if (!rule.getTableNames().contains(tableName))
            {
                continue;
            }
            Expression expression = rule.getExpression(tableName, table.getAlias(), dto);
            if (expression == null)
            {
                continue;
            }
            combined = combined == null ? expression : new net.sf.jsqlparser.expression.operators.conditional.AndExpression(combined, expression);
        }
        return combined;
    }

    private String stripBackticks(String name)
    {
        if (name != null && name.startsWith("`") && name.endsWith("`") && name.length() >= 2)
        {
            return name.substring(1, name.length() - 1);
        }
        return name;
    }
}
