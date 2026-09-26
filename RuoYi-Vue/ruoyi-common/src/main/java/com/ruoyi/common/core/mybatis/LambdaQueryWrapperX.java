package com.ruoyi.common.core.mybatis;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.toolkit.support.SFunction;
import com.ruoyi.common.utils.StringUtils;

/**
 * LambdaQueryWrapper 增强：条件值为空/null 自动跳过（IfPresent 方法族），
 * 替代 XML &lt;if test="xxx != null and xxx != ''"&gt; 动态条件的 Java 侧写法。
 * 重写父类链式方法返回本类型，避免链式调用中途"断链"回父类。
 *
 * @author ruoyi
 */
public class LambdaQueryWrapperX<T> extends LambdaQueryWrapper<T>
{
    public LambdaQueryWrapperX<T> eqIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.eq(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> neIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.ne(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> gtIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.gt(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> geIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.ge(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> ltIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.lt(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> leIfPresent(SFunction<T, ?> column, Object val)
    {
        if (val != null && StringUtils.isNotBlank(String.valueOf(val)))
        {
            return (LambdaQueryWrapperX<T>) super.le(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> likeIfPresent(SFunction<T, ?> column, String val)
    {
        if (StringUtils.isNotBlank(val))
        {
            return (LambdaQueryWrapperX<T>) super.like(column, val);
        }
        return this;
    }

    public LambdaQueryWrapperX<T> inIfPresent(SFunction<T, ?> column, java.util.Collection<?> values)
    {
        if (values != null && !values.isEmpty())
        {
            return (LambdaQueryWrapperX<T>) super.in(column, values);
        }
        return this;
    }

    /**
     * between：两端都非空 → between；只传一端 → 半开区间退化为 ge/le；两端空 → 跳过
     */
    public LambdaQueryWrapperX<T> betweenIfPresent(SFunction<T, ?> column, Object val1, Object val2)
    {
        if (val1 != null && val2 != null)
        {
            return (LambdaQueryWrapperX<T>) super.between(column, val1, val2);
        }
        if (val1 != null)
        {
            return (LambdaQueryWrapperX<T>) super.ge(column, val1);
        }
        if (val2 != null)
        {
            return (LambdaQueryWrapperX<T>) super.le(column, val2);
        }
        return this;
    }

    // ===== 链式返回类型重写（防止 eq()/like() 后返回父类类型导致 IfPresent 接不上） =====

    @Override
    public LambdaQueryWrapperX<T> eq(SFunction<T, ?> column, Object val)
    {
        return (LambdaQueryWrapperX<T>) super.eq(column, val);
    }

    @Override
    public LambdaQueryWrapperX<T> like(SFunction<T, ?> column, Object val)
    {
        return (LambdaQueryWrapperX<T>) super.like(column, val);
    }

    @Override
    public LambdaQueryWrapperX<T> in(SFunction<T, ?> column, java.util.Collection<?> coll)
    {
        return (LambdaQueryWrapperX<T>) super.in(column, coll);
    }

    @Override
    public LambdaQueryWrapperX<T> orderByAsc(SFunction<T, ?> column)
    {
        return (LambdaQueryWrapperX<T>) super.orderByAsc(column);
    }

    @Override
    public LambdaQueryWrapperX<T> orderByDesc(SFunction<T, ?> column)
    {
        return (LambdaQueryWrapperX<T>) super.orderByDesc(column);
    }
}
