package com.ruoyi.system.mapper;

import java.util.List;
import org.apache.ibatis.annotations.Param;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.ruoyi.system.domain.SysLogininfor;

/**
 * 系统访问日志情况信息 数据层
 *
 * <p>sys_logininfor 表无审计列，同 SysOperLogMapper 采用 IPage 首参 + XML 分页模式（坑 6）。
 *
 * @author ruoyi
 */
public interface SysLogininforMapper
{
    /**
     * 新增系统登录日志（XML sysdate() 兜底 login_time）
     *
     * @param logininfor 访问日志对象
     */
    void insertLogininfor(SysLogininfor logininfor);

    /**
     * 分页查询系统登录日志（XML 语句 + IPage 首参）
     *
     * @param page 分页对象
     * @param logininfor 访问日志对象
     * @return 分页结果
     */
    IPage<SysLogininfor> selectLogininforPage(IPage<SysLogininfor> page, @Param("logininfor") SysLogininfor logininfor);

    /**
     * 查询系统登录日志集合（非分页：导出）
     *
     * @param logininfor 访问日志对象
     * @return 登录记录集合
     */
    List<SysLogininfor> selectLogininforList(@Param("logininfor") SysLogininfor logininfor);

    /**
     * 批量删除系统登录日志
     *
     * @param infoIds 需要删除的登录日志ID
     * @return 结果
     */
    int deleteLogininforByIds(Long[] infoIds);

    /**
     * 清空系统登录日志
     *
     * @return 结果
     */
    int cleanLogininfor();
}
