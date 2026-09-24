// 定时任务 + 任务日志数据访问与调度器生命周期管理（对位 SysJobMapper/SysJobLogMapper + SysJobServiceImpl）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// JobDAO 定时任务。
type JobDAO struct{ db *gorm.DB }

func NewJobDAO(db *gorm.DB) *JobDAO { return &JobDAO{db: db} }

// JobDO 定时任务（对位 SysJob）。
type JobDO struct {
	JobID          int64  `gorm:"column:job_id" json:"jobId"`
	JobName        string `gorm:"column:job_name" json:"jobName"`
	JobGroup       string `gorm:"column:job_group" json:"jobGroup"`
	InvokeTarget   string `gorm:"column:invoke_target" json:"invokeTarget"`
	CronExpression string `gorm:"column:cron_expression" json:"cronExpression"`
	MisfirePolicy  string `gorm:"column:misfire_policy" json:"misfirePolicy"`
	Concurrent     string `gorm:"column:concurrent" json:"concurrent"`
	Status         string `gorm:"column:status" json:"status"`
	CreateBy       string `gorm:"column:create_by" json:"createBy"`
	CreateTime     string `gorm:"column:create_time" json:"createTime"`
	UpdateBy       string `gorm:"column:update_by" json:"updateBy"`
	UpdateTime     string `gorm:"column:update_time" json:"updateTime"`
	Remark         string `gorm:"column:remark" json:"remark"`
}

// SelectJobList 条件列表（jobName/jobGroup 相等 + status）。
func (d *JobDAO) SelectJobList(ctx context.Context, jobName, jobGroup, status string) ([]JobDO, error) {
	where := "where 1=1"
	var args []any
	if jobName != "" {
		where += " AND job_name like concat('%', ?, '%')"
		args = append(args, jobName)
	}
	if jobGroup != "" && jobGroup != "0" {
		where += " AND job_group = ?"
		args = append(args, jobGroup)
	}
	if status != "" && status != "-1" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []JobDO
	err := d.db.WithContext(ctx).Raw("select * from sys_job "+where+" order by job_id", args...).Scan(&list).Error
	return list, err
}

// SelectJobAll 全部（调度器启动加载）。
func (d *JobDAO) SelectJobAll(ctx context.Context) ([]JobDO, error) {
	var list []JobDO
	err := d.db.WithContext(ctx).Raw("select * from sys_job").Scan(&list).Error
	return list, err
}

// SelectJobById 按 ID。
func (d *JobDAO) SelectJobById(ctx context.Context, jobID int64) (*JobDO, error) {
	var m JobDO
	err := d.db.WithContext(ctx).Raw("select * from sys_job where job_id = ?", jobID).Scan(&m).Error
	if err != nil || m.JobID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertJob 新增。
func (d *JobDAO) InsertJob(ctx context.Context, m *JobDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_job (job_name, job_group, invoke_target, cron_expression, misfire_policy, concurrent, status, create_by, create_time, remark) values (?, ?, ?, ?, ?, ?, '1', ?, now(), ?)",
		m.JobName, m.JobGroup, m.InvokeTarget, m.CronExpression, m.MisfirePolicy, m.Concurrent, m.CreateBy, m.Remark).Error
}

// UpdateJob 修改。
func (d *JobDAO) UpdateJob(ctx context.Context, m *JobDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_job set job_name = ?, job_group = ?, invoke_target = ?, cron_expression = ?, misfire_policy = ?, concurrent = ?, update_by = ?, update_time = now(), remark = ? where job_id = ?",
		m.JobName, m.JobGroup, m.InvokeTarget, m.CronExpression, m.MisfirePolicy, m.Concurrent, m.UpdateBy, m.Remark, m.JobID).Error
}

// UpdateJobStatus 状态。
func (d *JobDAO) UpdateJobStatus(ctx context.Context, jobID int64, status, updateBy string) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_job set status = ?, update_by = ?, update_time = now() where job_id = ?", status, updateBy, jobID).Error
}

// DeleteJobByIds 批量物理删。
func (d *JobDAO) DeleteJobByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_job where job_id in ("+ph+")", args...).Error
}

// JobLogDAO 任务日志。
type JobLogDAO struct{ db *gorm.DB }

func NewJobLogDAO(db *gorm.DB) *JobLogDAO { return &JobLogDAO{db: db} }

// JobLogDO 任务日志（对位 SysJobLog）。
type JobLogDO struct {
	JobLogID      int64  `gorm:"column:job_log_id" json:"jobLogId"`
	JobName       string `gorm:"column:job_name" json:"jobName"`
	JobGroup      string `gorm:"column:job_group" json:"jobGroup"`
	InvokeTarget  string `gorm:"column:invoke_target" json:"invokeTarget"`
	JobMessage    string `gorm:"column:job_message" json:"jobMessage"`
	Status        string `gorm:"column:status" json:"status"`
	ExceptionInfo string `gorm:"column:exception_info" json:"exceptionInfo"`
	CreateTime    string `gorm:"column:create_time" json:"createTime"`
}

// SelectJobLogList 条件列表。
func (d *JobLogDAO) SelectJobLogList(ctx context.Context, jobName, jobGroup, status string) ([]JobLogDO, error) {
	where := "where 1=1"
	var args []any
	if jobName != "" {
		where += " AND job_name like concat('%', ?, '%')"
		args = append(args, jobName)
	}
	if jobGroup != "" && jobGroup != "0" {
		where += " AND job_group = ?"
		args = append(args, jobGroup)
	}
	if status != "" && status != "-1" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []JobLogDO
	err := d.db.WithContext(ctx).Raw("select * from sys_job_log "+where+" order by job_log_id desc", args...).Scan(&list).Error
	return list, err
}

// InsertJobLog 写日志。
func (d *JobLogDAO) InsertJobLog(ctx context.Context, m *JobLogDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_job_log (job_name, job_group, invoke_target, job_message, status, exception_info, create_time) values (?, ?, ?, ?, ?, ?, now())",
		m.JobName, m.JobGroup, m.InvokeTarget, m.JobMessage, m.Status, m.ExceptionInfo).Error
}

// DeleteJobLogByIds / CleanJobLog。
func (d *JobLogDAO) DeleteJobLogByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_job_log where job_log_id in ("+ph+")", args...).Error
}

func (d *JobLogDAO) CleanJobLog(ctx context.Context) error {
	return d.db.WithContext(ctx).Exec("truncate table sys_job_log").Error
}

// QueryJobs 供 handler.LoadEnabledJobsForMain 的最小查询实现。
func (d *JobDAO) QueryJobs(ctx context.Context, dest *[]JobDO) error {
	return d.db.WithContext(ctx).Raw("select * from sys_job").Scan(dest).Error
}
