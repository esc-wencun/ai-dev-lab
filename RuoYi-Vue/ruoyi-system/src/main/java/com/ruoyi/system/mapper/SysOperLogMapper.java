package com.ruoyi.system.mapper;

import java.util.List;
import org.apache.ibatis.annotations.Param;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.ruoyi.system.domain.SysOperLog;

/**
 * 操作日志 数据层
 *
 * <p>sys_oper_log 表无审计列（create_by/create_time 等），实体继承的 BaseEntity 字段无法通过
 * @TableField(exist=false) 按表排除，故本模块不迁 BaseMapper，采用 IPage 首参 + XML 分页模式
 * （0.0.0 spec 3.4 join/DataScope 同款；坑 6 实查记录）。
 *
 * @author ruoyi
 */
public interface SysOperLogMapper
{
    /**
     * 新增操作日志（XML sysdate() 兜底 oper_time）
     *
     * @param operLog 操作日志对象
     */
    void insertOperlog(SysOperLog operLog);

    /**
     * 分页查询系统操作日志（XML 语句 + IPage 首参，分页插件自动 count+limit）
     *
     * @param page 分页对象（buildPage 组装）
     * @param operLog 操作日志对象
     * @return 分页结果
     */
    IPage<SysOperLog> selectOperLogPage(IPage<SysOperLog> page, @Param("operLog") SysOperLog operLog);

    /**
     * 查询系统操作日志集合（非分页：导出）
     *
     * @param operLog 操作日志对象
     * @return 操作日志集合
     */
    List<SysOperLog> selectOperLogList(@Param("operLog") SysOperLog operLog);

    /**
     * 批量删除系统操作日志
     *
     * @param operIds 需要删除的操作日志ID
     * @return 结果
     */
    int deleteOperLogByIds(Long[] operIds);

    /**
     * 查询操作日志详细
     *
     * @param operId 操作ID
     * @return 操作日志对象
     */
    SysOperLog selectOperLogById(Long operId);

    /**
     * 清空操作日志
     */
    void cleanOperLog();
}
