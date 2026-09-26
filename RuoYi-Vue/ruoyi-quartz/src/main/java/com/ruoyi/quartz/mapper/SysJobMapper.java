package com.ruoyi.quartz.mapper;

import com.baomidou.mybatisplus.core.metadata.IPage;
import java.util.List;
import org.apache.ibatis.annotations.Param;
import com.ruoyi.quartz.domain.SysJob;

/**
 * 调度任务信息 数据层
 * 
 * @author ruoyi
 */
public interface SysJobMapper
{
    /**
     * 查询调度任务日志集合
     * 
     * @param job 调度信息
     * @return 操作日志集合
     */
    List<SysJob> selectJobList(@Param("job") SysJob job);

    /**
     * 分页查询定时任务（IPage 首参）
     *
     * @param page 分页对象
     * @param job 定时任务信息
     * @return 分页结果
     */
    IPage<SysJob> selectJobPage(IPage<SysJob> page, @Param("job") SysJob job);

    /**
     * 查询所有调度任务
     * 
     * @return 调度任务列表
     */
    List<SysJob> selectJobAll();

    /**
     * 通过调度ID查询调度任务信息
     * 
     * @param jobId 调度ID
     * @return 角色对象信息
     */
    SysJob selectJobById(Long jobId);

    /**
     * 通过调度ID删除调度任务信息
     * 
     * @param jobId 调度ID
     * @return 结果
     */
    int deleteJobById(Long jobId);

    /**
     * 批量删除调度任务信息
     * 
     * @param ids 需要删除的数据ID
     * @return 结果
     */
    int deleteJobByIds(Long[] ids);

    /**
     * 修改调度任务信息
     * 
     * @param job 调度任务信息
     * @return 结果
     */
    int updateJob(SysJob job);

    /**
     * 新增调度任务信息
     * 
     * @param job 调度任务信息
     * @return 结果
     */
    int insertJob(SysJob job);
}
