package com.ruoyi.framework.mybatis;

import java.util.Set;

import net.sf.jsqlparser.expression.Expression;
import net.sf.jsqlparser.expression.Alias;

/**
 * 数据权限规则抽象：声明自己管辖的表集合，并按表生成过滤条件（JSqlParser AST）。
 *
 * @author ruoyi
 */
public interface DataPermissionRule
{
    /**
     * 本规则管辖的表名集合（白名单；未注册的表不会被本规则触碰）
     */
    Set<String> getTableNames();

    /**
     * 为指定表生成数据范围过滤条件
     *
     * @param tableName 表名（已剥反引号）
     * @param tableAlias SQL 中该表的别名（BaseMapper 生成的查询通常无别名）
     * @param dto 当前请求的数据权限范围（结构化，见 RuoYiDeptDataPermissionDTO）
     * @return 过滤条件；null 表示该表在本规则下不加条件（如 all / 缺列配置）
     */
    Expression getExpression(String tableName, Alias tableAlias, com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO dto);
}
