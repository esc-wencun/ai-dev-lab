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
     * 注意：标了 @TableLogic 的实体会被 MP 自动追加 del_flag 过滤，与 RuoYi 全表唯一性语义不一致——
     * 有 del_flag 的表不适用本方法（保留 XML 版 checkXxxUnique），见 0.0.0 spec 3.6。
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
