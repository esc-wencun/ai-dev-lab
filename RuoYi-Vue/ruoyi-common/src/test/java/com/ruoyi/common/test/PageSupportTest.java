package com.ruoyi.common.test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import java.lang.reflect.Proxy;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import com.baomidou.mybatisplus.core.metadata.OrderItem;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.exception.UtilException;
import com.ruoyi.common.utils.PageUtils;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

/**
 * Task 3 单元测试（不连库）：PageUtils.buildPage 参数解析、LambdaQueryWrapperX 条件生成。
 * 用 JDK 动态代理构造最小 HttpServletRequest stub（不引 spring-test，避免 Spring 6/7 版本错位），
 * 经 RequestContextHolder 供 TableSupport 读取。
 */
class PageSupportTest
{
    @AfterEach
    void cleanup()
    {
        RequestContextHolder.resetRequestAttributes();
    }

    @SuppressWarnings("unchecked")
    private void mockRequest(Map<String, String> params)
    {
        jakarta.servlet.http.HttpServletRequest stub =
            (jakarta.servlet.http.HttpServletRequest) Proxy.newProxyInstance(
                getClass().getClassLoader(),
                new Class<?>[] { jakarta.servlet.http.HttpServletRequest.class },
                (proxy, method, args) ->
                {
                    switch (method.getName())
                    {
                        case "getParameter":
                            return params.get((String) args[0]);
                        case "getParameterMap":
                            return params.entrySet().stream()
                                .collect(java.util.stream.Collectors.toMap(Map.Entry::getKey, e -> new String[] { e.getValue() }));
                        case "toString":
                            return "stubRequest" + params;
                        default:
                            return defaultValueFor(method.getReturnType());
                    }
                });
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(stub));
    }

    private static Object defaultValueFor(Class<?> type)
    {
        if (type == boolean.class) { return false; }
        if (type == int.class) { return 0; }
        if (type == long.class) { return 0L; }
        return null;
    }

    // ===== PageUtils.buildPage =====

    @Test
    void buildPageDefaults()
    {
        mockRequest(Map.of());
        Page<Object> page = PageUtils.buildPage();
        assertEquals(1, page.getCurrent());
        assertEquals(10, page.getSize());
        assertTrue(page.orders().isEmpty());
    }

    @Test
    void buildPageWithParams()
    {
        mockRequest(Map.of("pageNum", "3", "pageSize", "20"));
        Page<Object> page = PageUtils.buildPage();
        assertEquals(3, page.getCurrent());
        assertEquals(20, page.getSize());
    }

    @Test
    void buildPageOrderByCamelToUnderline()
    {
        mockRequest(Map.of("pageNum", "1", "pageSize", "10", "orderByColumn", "postSort", "isAsc", "ascending"));
        Page<Object> page = PageUtils.buildPage();
        List<OrderItem> orders = page.orders();
        assertEquals(1, orders.size());
        assertEquals("post_sort", orders.get(0).getColumn());
        assertTrue(orders.get(0).isAsc());
    }

    @Test
    void buildPageOrderByDescending()
    {
        mockRequest(Map.of("pageNum", "1", "pageSize", "10", "orderByColumn", "createTime", "isAsc", "descending"));
        Page<Object> page = PageUtils.buildPage();
        assertEquals("create_time", page.orders().get(0).getColumn());
        assertEquals(false, page.orders().get(0).isAsc());
    }

    @Test
    void buildPageRejectsInjection()
    {
        mockRequest(Map.of("pageNum", "1", "pageSize", "10", "orderByColumn", "postSort;drop"));
        assertThrows(UtilException.class, PageUtils::buildPage);
    }

    // ===== LambdaQueryWrapperX =====

    /** MP lambda 列解析要求真实 getter 方法引用 + TableInfo 缓存，测试专用最小实体 */
    static class Demo
    {
        private String name;
        private Long cnt;

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        public Long getCnt() { return cnt; }
        public void setCnt(Long cnt) { this.cnt = cnt; }
    }

    @BeforeAll
    static void initTableInfo()
    {
        // MP 解析 lambda 列名依赖 TableInfo 缓存（官方测试同款初始化方式）
        org.apache.ibatis.builder.MapperBuilderAssistant assistant =
            new org.apache.ibatis.builder.MapperBuilderAssistant(new com.baomidou.mybatisplus.core.MybatisConfiguration(), "");
        com.baomidou.mybatisplus.core.metadata.TableInfoHelper.initTableInfo(assistant, Demo.class);
    }

    /** 条件计数口径：先强制物化 SQL 片段（MP 条件与参数绑定是惰性的），再数绑定参数 */
    private static int boundParams(LambdaQueryWrapperX<Demo> w)
    {
        w.getSqlSegment();
        return w.getParamNameValuePairs().size();
    }

    @Test
    void eqIfPresentSkipsNullAndBlank()
    {
        LambdaQueryWrapperX<Demo> w = new LambdaQueryWrapperX<>();
        w.eqIfPresent(Demo::getName, null).eqIfPresent(Demo::getName, "  ").eqIfPresent(Demo::getCnt, null);
        assertEquals(0, boundParams(w), "null/空白值不应生成条件");
    }

    @Test
    void eqIfPresentGeneratesCondition()
    {
        LambdaQueryWrapperX<Demo> w = new LambdaQueryWrapperX<>();
        w.eqIfPresent(Demo::getName, "abc");
        assertEquals(1, boundParams(w));
    }

    @Test
    void betweenIfPresentHalfOpenDegeneration()
    {
        LambdaQueryWrapperX<Demo> w = new LambdaQueryWrapperX<>();
        w.betweenIfPresent(Demo::getCnt, 10L, null); // 只传下界 → ge
        assertTrue(w.getSqlSegment().contains(">="), "只传下界应退化为 >=：实际 " + w.getSqlSegment());

        LambdaQueryWrapperX<Demo> w2 = new LambdaQueryWrapperX<>();
        w2.betweenIfPresent(Demo::getCnt, null, 20L); // 只传上界 → le
        assertTrue(w2.getSqlSegment().contains("<="), "只传上界应退化为 <=");

        LambdaQueryWrapperX<Demo> w3 = new LambdaQueryWrapperX<>();
        w3.betweenIfPresent(Demo::getCnt, null, null);
        assertEquals(0, boundParams(w3), "两端空应无条件");
    }

    @Test
    void chainedReturnsXType()
    {
        LambdaQueryWrapperX<Demo> w = new LambdaQueryWrapperX<>();
        LambdaQueryWrapperX<Demo> chained = w.eq(Demo::getName, "a").likeIfPresent(Demo::getName, "b");
        assertEquals(2, boundParams(chained), "eq 后应能继续链 IfPresent 且条件都生效");
    }

    @Test
    void likeIfPresentAndInIfPresent()
    {
        LambdaQueryWrapperX<Demo> w = new LambdaQueryWrapperX<>();
        w.likeIfPresent(Demo::getName, "  "); // 空白跳过
        assertEquals(0, boundParams(w));

        LambdaQueryWrapperX<Demo> w2 = new LambdaQueryWrapperX<>();
        w2.likeIfPresent(Demo::getName, "k");
        java.util.Collection<Long> empty = java.util.List.of();
        w2.inIfPresent(Demo::getCnt, empty); // 空集合跳过（IN () 防护）
        assertEquals(1, boundParams(w2));

        LambdaQueryWrapperX<Demo> w3 = new LambdaQueryWrapperX<>();
        w3.inIfPresent(Demo::getCnt, java.util.List.of(1L, 2L));
        assertEquals(2, boundParams(w3), "IN 各值单独绑定参数");
    }
}
