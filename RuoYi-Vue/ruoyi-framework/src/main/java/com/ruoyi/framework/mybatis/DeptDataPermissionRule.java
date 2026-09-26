package com.ruoyi.framework.mybatis;

import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import net.sf.jsqlparser.expression.Alias;
import net.sf.jsqlparser.expression.Expression;
import net.sf.jsqlparser.expression.LongValue;
import net.sf.jsqlparser.expression.NullValue;
import net.sf.jsqlparser.expression.operators.conditional.OrExpression;
import net.sf.jsqlparser.expression.operators.relational.EqualsTo;
import net.sf.jsqlparser.expression.operators.relational.ExpressionList;
import net.sf.jsqlparser.expression.operators.relational.InExpression;
import net.sf.jsqlparser.schema.Column;
import net.sf.jsqlparser.schema.Table;
import com.baomidou.mybatisplus.core.metadata.TableInfo;
import com.baomidou.mybatisplus.core.metadata.TableInfoHelper;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;
import com.ruoyi.common.utils.StringUtils;

/**
 * 部门数据权限规则（当前项目唯一规则实现）。
 *
 * <p>白名单机制：deptColumns/userColumns 两个 Map（表名 → 列名）即"表注册表"，
 * 未注册的表 getTableNames() 不含、getExpression 不命中——天然零侵入。
 * 注册入口见 system 模块的数据权限配置类；注册即生效，改动必须走 checklist 数据权限实测。
 *
 * @author ruoyi
 */
public class DeptDataPermissionRule implements DataPermissionRule
{
    /** 表名 → 部门列名 */
    private final Map<String, String> deptColumns = new ConcurrentHashMap<>();

    /** 表名 → 用户列名 */
    private final Map<String, String> userColumns = new ConcurrentHashMap<>();

    /** 注册部门列（实体类推导表名，列名默认 dept_id） */
    public DeptDataPermissionRule addDeptColumn(Class<?> entityClass)
    {
        return addDeptColumn(resolveTableName(entityClass), "dept_id");
    }

    public DeptDataPermissionRule addDeptColumn(String tableName, String columnName)
    {
        deptColumns.put(tableName, columnName);
        return this;
    }

    /** 注册用户列（实体类推导表名，列名默认 user_id） */
    public DeptDataPermissionRule addUserColumn(Class<?> entityClass)
    {
        return addUserColumn(resolveTableName(entityClass), "user_id");
    }

    public DeptDataPermissionRule addUserColumn(String tableName, String columnName)
    {
        userColumns.put(tableName, columnName);
        return this;
    }

    @Override
    public Set<String> getTableNames()
    {
        Set<String> names = new HashSet<>(deptColumns.keySet());
        names.addAll(userColumns.keySet());
        return names;
    }

    @Override
    public Expression getExpression(String tableName, Alias tableAlias, RuoYiDeptDataPermissionDTO dto)
    {
        if (dto == null)
        {
            // 快速失败：上下文激活（入栈）但 DTO 缺失属异常状态
            throw new IllegalStateException(String.format("数据权限范围 DTO 缺失，table=%s", tableName));
        }
        // 情况一：全部数据权限 → 不加条件
        if (dto.isAll())
        {
            return null;
        }
        // 情况二：既无部门集又非仅本人 → 无任何权限，WHERE null = null 保证空结果（不报错不漏数据）
        boolean hasDeptIds = dto.getDeptIds() != null && !dto.getDeptIds().isEmpty();
        if (!hasDeptIds && !dto.isSelf())
        {
            return new EqualsTo(new NullValue(), new NullValue());
        }

        String deptColumn = deptColumns.get(tableName);
        String userColumn = userColumns.get(tableName);
        Expression deptExpression = buildDeptExpression(tableName, tableAlias, deptColumn, dto);
        Expression userExpression = buildUserExpression(tableName, tableAlias, userColumn, dto);

        // 情况三：两维并存 → 带括号 OR
        if (deptExpression != null && userExpression != null)
        {
            return new OrExpression(deptExpression, userExpression);
        }
        if (deptExpression != null) return deptExpression;
        if (userExpression != null) return userExpression;
        // 两列都没注册：表在白名单但列配置缺失属配置错误，快速失败
        throw new IllegalStateException(String.format("表 %s 未注册 dept/user 列，无法生成数据范围条件", tableName));
    }

    private Expression buildDeptExpression(String tableName, Alias tableAlias, String columnName, RuoYiDeptDataPermissionDTO dto)
    {
        if (StringUtils.isEmpty(columnName) || dto.getDeptIds() == null || dto.getDeptIds().isEmpty())
        {
            return null;
        }
        Column column = buildColumn(tableName, tableAlias, columnName);
        // 必须用 ParenthesedExpressionList 显式包括号：单元素时 ExpressionList 会生成 "IN 105"（非法 SQL，实测踩过）
        InExpression in = new InExpression();
        in.setLeftExpression(column);
        in.setRightExpression(new net.sf.jsqlparser.expression.operators.relational.ParenthesedExpressionList<>(
                dto.getDeptIds().stream().map(LongValue::new).collect(java.util.stream.Collectors.toList())));
        return in;
    }

    private Expression buildUserExpression(String tableName, Alias tableAlias, String columnName, RuoYiDeptDataPermissionDTO dto)
    {
        if (StringUtils.isEmpty(columnName) || !dto.isSelf())
        {
            return null;
        }
        return new EqualsTo(buildColumn(tableName, tableAlias, columnName), new LongValue(currentUserId()));
    }

    private Long currentUserId()
    {
        return com.ruoyi.common.utils.SecurityUtils.getUserId();
    }

    private Column buildColumn(String tableName, Alias tableAlias, String columnName)
    {
        if (tableAlias != null && StringUtils.isNotEmpty(tableAlias.getName()))
        {
            return new Column(new Table(tableAlias.getName()), columnName);
        }
        return new Column(tableName + "." + columnName);
    }

    private String resolveTableName(Class<?> entityClass)
    {
        TableInfo tableInfo = TableInfoHelper.getTableInfo(entityClass);
        if (tableInfo == null && !entityClass.isAnnotationPresent(com.baomidou.mybatisplus.annotation.TableName.class))
        {
            throw new IllegalStateException("实体未注册到 MyBatis-Plus 且无 @TableName，无法推导表名：" + entityClass.getName());
        }
        if (tableInfo != null)
        {
            return tableInfo.getTableName();
        }
        // TableInfo 未初始化（如实体所在模块未启动 MP 扫描）时退化为驼峰转下划线的约定推导
        String simple = entityClass.getSimpleName();
        return StringUtils.toUnderScoreCase(simple.startsWith("Sys") ? simple.substring(3) : simple);
    }
}
