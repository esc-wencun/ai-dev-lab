package com.ruoyi.common.utils;

import java.util.ArrayList;
import java.util.List;
import com.baomidou.mybatisplus.core.metadata.OrderItem;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.page.PageDomain;
import com.ruoyi.common.core.page.TableSupport;
import com.ruoyi.common.utils.sql.SqlUtil;

/**
 * 分页工具类（MyBatis-Plus 实现）
 *
 * @author ruoyi
 */
public class PageUtils
{
    /**
     * 构建 MyBatis-Plus 分页对象：pageNum/pageSize/orderBy 从请求参数解析，
     * orderBy 沿用 PageDomain 的驼峰转下划线 + SqlUtil 注入转义
     */
    public static <T> Page<T> buildPage()
    {
        PageDomain pageDomain = TableSupport.buildPageRequest();
        Page<T> page = new Page<>(pageDomain.getPageNum(), pageDomain.getPageSize());
        String orderBy = SqlUtil.escapeOrderBySql(pageDomain.getOrderBy());
        if (StringUtils.isNotEmpty(orderBy))
        {
            // orderBy 形如 "column asc"（PageDomain 已做驼峰转下划线），按末位方向词拆列
            String[] parts = orderBy.trim().split("\\s+");
            String column = parts[0];
            boolean isAsc = parts.length < 2 || !"desc".equalsIgnoreCase(parts[parts.length - 1]);
            OrderItem orderItem = isAsc ? OrderItem.asc(column) : OrderItem.desc(column);
            List<OrderItem> orders = new ArrayList<>();
            orders.add(orderItem);
            page.addOrder(orders);
        }
        return page;
    }
}
