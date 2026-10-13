package com.ruoyi.common.core.service;

import org.springframework.beans.factory.annotation.Autowired;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.core.toolkit.support.SFunction;

/**
 * Service 层基类：注入泛型 Mapper，收敛公用样板（唯一性校验等）。
 * 复杂 ServiceImpl（多 Mapper 注入场景）可选择性继承，不强扭。
 *
 * @author ruoyi
 */
public class BaseService<M extends BaseMapper<T>, T>
{
    @Autowired
    protected M baseMapper;

    /**
     * 唯一性校验：true=唯一。keyValue 传 null = 新增校验；传当前 id = 编辑时排除自身。
     * 语义（0.0.0 Task 10 起反转）：实体启用全局逻辑删除（delFlag 属性命中 DbConfig.logicDeleteField）时，
     * selectCount 自动带 del_flag='0'——查重范围为存活行，已删名字/编码可复用（AI 域用户拍板 2026-10-13）；
     * 仍需"含已删行全表唯一"语义的表，查重走 XML 手写 SQL。
     */
    public boolean checkUnique(SFunction<T, ?> keyField, Object keyValue, SFunction<T, ?> field, Object value)
    {
        LambdaQueryWrapper<T> wrapper = new LambdaQueryWrapper<T>().eq(field, value);
        if (keyValue != null)
        {
            wrapper.ne(keyField, keyValue);
        }
        return baseMapper.selectCount(wrapper) == 0;
    }
}
