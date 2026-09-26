package com.ruoyi.framework.test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import java.util.Set;
import org.junit.jupiter.api.Test;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;
import com.ruoyi.common.core.mybatis.DataPermissionUtils;
import com.ruoyi.common.core.mybatis.DataScopeContextHolder;
import com.ruoyi.framework.mybatis.DeptDataPermissionRule;

/**
 * 1.0.0 单元测试（不连库）：ContextHolder 栈语义与防泄漏、executeIgnore、
 * DeptDataPermissionRule 的 DTO→Expression 生成（逐 scope 断言生成 SQL）。
 */
class DataPermissionTest
{
    // ===== ContextHolder 栈语义 =====

    @Test
    void contextStackNestingAndCleanup()
    {
        DataScopeContextHolder.Scope outer = new DataScopeContextHolder.Scope(true, RuoYiDeptDataPermissionDTO.all());
        DataScopeContextHolder.Scope inner = new DataScopeContextHolder.Scope(false, null);
        DataScopeContextHolder.add(outer);
        DataScopeContextHolder.add(inner);
        assertEquals(false, DataScopeContextHolder.peek().isEnable(), "栈顶应为内层");
        DataScopeContextHolder.remove();
        assertEquals(true, DataScopeContextHolder.peek().isEnable(), "出栈后应回到外层");
        DataScopeContextHolder.remove();
        assertNull(DataScopeContextHolder.peek(), "全出栈后应为空");
    }

    @Test
    void contextThreadLocalReleasedWhenEmpty()
    {
        DataScopeContextHolder.add(new DataScopeContextHolder.Scope(true, RuoYiDeptDataPermissionDTO.none()));
        DataScopeContextHolder.remove();
        // 再次 peek 应为 null（ThreadLocal 已 remove，不会残留旧栈）
        assertNull(DataScopeContextHolder.peek());
    }

    @Test
    void executeIgnoreSetsDisableScopeAndRestores()
    {
        DataScopeContextHolder.add(new DataScopeContextHolder.Scope(true, RuoYiDeptDataPermissionDTO.all()));
        String result = DataPermissionUtils.executeIgnore(() -> {
            assertFalse(DataScopeContextHolder.peek().isEnable(), "豁免作用域内 enable 应为 false");
            return "done";
        });
        assertEquals("done", result);
        assertEquals(true, DataScopeContextHolder.peek().isEnable(), "豁免结束后应恢复外层");
        DataScopeContextHolder.remove();
        assertNull(DataScopeContextHolder.peek());
    }

    @Test
    void executeIgnoreRestoresOnException()
    {
        try
        {
            DataPermissionUtils.executeIgnore(() -> { throw new IllegalStateException("boom"); });
        }
        catch (RuntimeException expected)
        {
            // ignore
        }
        assertNull(DataScopeContextHolder.peek(), "异常路径也必须清栈（防跨请求越权）");
    }

    // ===== DeptDataPermissionRule 表达式生成 =====

    private DeptDataPermissionRule newRule()
    {
        DeptDataPermissionRule rule = new DeptDataPermissionRule();
        rule.addDeptColumn("biz_order", "dept_id");
        rule.addUserColumn("biz_order", "create_by_id");
        return rule;
    }

    private RuoYiDeptDataPermissionDTO dto(boolean all, boolean self, Set<Long> deptIds)
    {
        return new RuoYiDeptDataPermissionDTO(all, self, deptIds);
    }

    @Test
    void allScopeProducesNoCondition()
    {
        assertNull(newRule().getExpression("biz_order", null, dto(true, false, null)));
    }

    @Test
    void noPermissionProducesNullEqualsNull()
    {
        String sql = newRule().getExpression("biz_order", null, dto(false, false, null)).toString();
        assertTrue(sql.replace("NULL", "null").equalsIgnoreCase("null = null"), "应为恒假条件（空集）：实际 " + sql);
    }

    @Test
    void deptOnlyProducesIn()
    {
        String sql = newRule().getExpression("biz_order", null, dto(false, false, Set.of(100L, 101L))).toString();
        assertTrue(sql.startsWith("biz_order.dept_id IN"), "应为 dept_id IN：实际 " + sql);
        assertTrue(sql.contains("100") && sql.contains("101"), "IN 集合应含两个 ID：实际 " + sql);
    }

    @Test
    void selfOnlyProducesEquals()
    {
        loginAs(7L);
        try
        {
            String sql = newRule().getExpression("biz_order", null, dto(false, true, null)).toString();
            assertEquals("biz_order.create_by_id = 7", sql);
        }
        finally
        {
            clearLogin();
        }
    }

    @Test
    void deptAndSelfProducesParenthesizedOr()
    {
        loginAs(7L);
        try
        {
            String sql = newRule().getExpression("biz_order", null, dto(false, true, Set.of(100L))).toString();
            assertTrue(sql.contains("biz_order.dept_id IN") && sql.contains("biz_order.create_by_id = 7")
                    && sql.contains(" OR "), "应为 dept IN OR user = 组合：实际 " + sql);
        }
        finally
        {
            clearLogin();
        }
    }

    @Test
    void unregisteredTableNotInWhitelist()
    {
        assertFalse(newRule().getTableNames().contains("other_table"));
        assertTrue(newRule().getTableNames().contains("biz_order"));
    }

    // ===== 测试辅助：构造 Spring Security 上下文（UserIdentity + SysUser 最小集） =====

    private void loginAs(Long userId)
    {
        com.ruoyi.common.core.domain.entity.SysUser user = new com.ruoyi.common.core.domain.entity.SysUser();
        user.setUserId(userId);
        com.ruoyi.common.core.domain.model.LoginUser loginUser =
            new com.ruoyi.common.core.domain.model.LoginUser(userId, null, user, java.util.Collections.emptySet());
        org.springframework.security.authentication.UsernamePasswordAuthenticationToken auth =
            new org.springframework.security.authentication.UsernamePasswordAuthenticationToken(
                loginUser, null, java.util.Collections.emptyList());
        org.springframework.security.core.context.SecurityContextHolder.getContext().setAuthentication(auth);
    }

    private void clearLogin()
    {
        org.springframework.security.core.context.SecurityContextHolder.clearContext();
    }
}
