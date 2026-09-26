package com.ruoyi.common.core.domain.dto;

import java.util.Set;

/**
 * 部门数据权限范围（结构化中间表示，0.0.0/1.0.0 spec）。
 * 把"范围语义"与 SQL 方言/表别名解耦：切面、MP 拦截器两条路径都消费它。
 *
 * @author ruoyi
 */
public class RuoYiDeptDataPermissionDTO
{
    /** 全部数据权限 */
    private boolean all;

    /** 仅本人 */
    private boolean self;

    /** 可见的部门 ID 集合（自定义/本部门/本部门及以下 的并集累加） */
    private Set<Long> deptIds;

    public RuoYiDeptDataPermissionDTO()
    {
    }

    public RuoYiDeptDataPermissionDTO(boolean all, boolean self, Set<Long> deptIds)
    {
        this.all = all;
        this.self = self;
        this.deptIds = deptIds;
    }

    public static RuoYiDeptDataPermissionDTO all()
    {
        return new RuoYiDeptDataPermissionDTO(true, false, null);
    }

    public static RuoYiDeptDataPermissionDTO none()
    {
        return new RuoYiDeptDataPermissionDTO(false, false, null);
    }

    public boolean isAll()
    {
        return all;
    }

    public void setAll(boolean all)
    {
        this.all = all;
    }

    public boolean isSelf()
    {
        return self;
    }

    public void setSelf(boolean self)
    {
        this.self = self;
    }

    public Set<Long> getDeptIds()
    {
        return deptIds;
    }

    public void setDeptIds(Set<Long> deptIds)
    {
        this.deptIds = deptIds;
    }
}
