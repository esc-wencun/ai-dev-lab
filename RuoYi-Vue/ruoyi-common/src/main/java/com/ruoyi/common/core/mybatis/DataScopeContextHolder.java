package com.ruoyi.common.core.mybatis;

import java.util.ArrayDeque;
import java.util.Deque;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;

/**
 * 数据权限桥接上下文：栈式 ThreadLocal。
 *
 * <p>切面（@DataScope）在方法入口把 {enable, DTO} 入栈、出口（finally）出栈；
 * MP DataPermissionInterceptor 的 handler 在 SQL 执行时读取栈顶——
 * 以此把"注解驱动"的范围语义桥接到无 XML 注入点的纯 BaseMapper/Wrapper 查询。
 *
 * <p>生命周期纪律：栈元素在切面 finally 中出栈，栈空时主动 remove ThreadLocal（防线程复用泄漏）。
 *
 * @author ruoyi
 */
public class DataScopeContextHolder
{
    /** 栈元素：enable=false 表示本次作用域内豁免数据权限 */
    public static class Scope
    {
        private final boolean enable;
        private final RuoYiDeptDataPermissionDTO dto;

        public Scope(boolean enable, RuoYiDeptDataPermissionDTO dto)
        {
            this.enable = enable;
            this.dto = dto;
        }

        public boolean isEnable()
        {
            return enable;
        }

        public RuoYiDeptDataPermissionDTO getDto()
        {
            return dto;
        }
    }

    private static final ThreadLocal<Deque<Scope>> SCOPES = new ThreadLocal<>();

    private DataScopeContextHolder()
    {
    }

    /** 入栈（切面入口 / executeIgnore 开始） */
    public static void add(Scope scope)
    {
        Deque<Scope> deque = SCOPES.get();
        if (deque == null)
        {
            deque = new ArrayDeque<>();
            SCOPES.set(deque);
        }
        deque.push(scope);
    }

    /** 读栈顶（拦截器 handler 调用）；无活跃作用域返回 null */
    public static Scope peek()
    {
        Deque<Scope> deque = SCOPES.get();
        return deque == null ? null : deque.peek();
    }

    /** 出栈（切面 finally / executeIgnore 结束）；栈空时移除 ThreadLocal 防泄漏 */
    public static void remove()
    {
        Deque<Scope> deque = SCOPES.get();
        if (deque != null)
        {
            deque.pop();
            if (deque.isEmpty())
            {
                SCOPES.remove();
            }
        }
    }
}
