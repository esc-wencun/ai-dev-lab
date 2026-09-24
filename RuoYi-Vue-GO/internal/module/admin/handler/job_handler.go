// 定时任务 + 任务日志 HTTP 端点（对位 SysJobController + SysJobLogController）。
// 路由遮蔽规避：固定路径（changeStatus/run/clean）先于参数路径注册（AGENTS.md 坑 7）。
package handler

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/scheduler"
	"ruoyi-vue-go/pkg/response"
)

// JobHandler 定时任务。
type JobHandler struct {
	job     *dao.JobDAO
	jobLog  *dao.JobLogDAO
	runOnce sync.Map // run 一次性执行的防抖（非必须，语义占位）
}

func NewJobHandler(j *dao.JobDAO, jl *dao.JobLogDAO) *JobHandler {
	return &JobHandler{job: j, jobLog: jl}
}

// List GET /monitor/job/list。
func (h *JobHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:list") == nil {
		return
	}
	list, err := h.job.SelectJobList(c.Request.Context(), c.Query("jobName"), c.Query("jobGroup"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, jobPtr(list))
}

// GetInfo GET /monitor/job/{jobId}。
func (h *JobHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:query") == nil {
		return
	}
	id, _ := strconv.ParseInt(c.Param("jobId"), 10, 64)
	m, err := h.job.SelectJobById(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.OkData(c, m).JSON()
}

// Add POST /monitor/job（新增默认暂停 status='1'，对位 Java：创建后需手动启用）。
func (h *JobHandler) Add(c *gin.Context) {
	lu := requirePermAndLogin(c, "monitor:job:add")
	if lu == nil {
		return
	}
	var body dao.JobDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	body.CreateBy = lu.Username()
	if err := h.job.InsertJob(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	// 新增即按暂停处理：不注册进调度器（对位 Java ScheduleUtils 判 status）
	response.Ok(c).JSON()
}

// Edit PUT /monitor/job（调度中的任务重载 cron）。
func (h *JobHandler) Edit(c *gin.Context) {
	lu := requirePermAndLogin(c, "monitor:job:edit")
	if lu == nil {
		return
	}
	var body dao.JobDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	body.UpdateBy = lu.Username()
	if err := h.job.UpdateJob(c.Request.Context(), &body); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	// 若在调度中：移除后按新 cron 重载
	if scheduler.LookupJob(body.JobID) {
		scheduler.DeleteJob(body.JobID)
		_ = h.scheduleJob(c.Request.Context(), body.JobID)
	}
	response.Ok(c).JSON()
}

// ChangeStatus PUT /monitor/job/changeStatus（启用=从库补载进调度器；暂停=移除）。
func (h *JobHandler) ChangeStatus(c *gin.Context) {
	lu := requirePermAndLogin(c, "monitor:job:changeStatus")
	if lu == nil {
		return
	}
	var body dao.JobDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.job.UpdateJobStatus(c.Request.Context(), body.JobID, body.Status, lu.Username()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	if body.Status == "0" { // 启用：从库读最新配置补载（规避静默空操作）
		scheduler.DeleteJob(body.JobID)
		if err := h.scheduleJob(c.Request.Context(), body.JobID); err != nil {
			response.Error(c, err.Error()).JSON()
			return
		}
	} else { // 暂停：移出调度器
		scheduler.DeleteJob(body.JobID)
	}
	response.Ok(c).JSON()
}

// Run PUT /monitor/job/run（立即执行一次；写日志）。
func (h *JobHandler) Run(c *gin.Context) {
	lu := requirePermAndLogin(c, "monitor:job:changeStatus")
	if lu == nil {
		return
	}
	var body dao.JobDO
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	m, err := h.job.SelectJobById(c.Request.Context(), body.JobID)
	if err != nil {
		response.Error(c, "任务不存在").JSON()
		return
	}
	go h.executeAndLog(m)
	response.Ok(c).JSON()
}

// Remove DELETE /monitor/job/{jobIds}。
func (h *JobHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:remove") == nil {
		return
	}
	ids := parseInt64List(c.Param("jobIds"))
	for _, id := range ids {
		scheduler.DeleteJob(id)
	}
	if err := h.job.DeleteJobByIds(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// scheduleJob 从库读任务并注册调度（bean.method 解析 + 注册表执行）。
func (h *JobHandler) scheduleJob(ctx context.Context, jobID int64) error {
	m, err := h.job.SelectJobById(ctx, jobID)
	if err != nil {
		return err
	}
	if m.Status != "0" {
		return nil // 暂停态不加载
	}
	bean, method, params, err := scheduler.ParseTarget(m.InvokeTarget)
	if err != nil {
		return err
	}
	return scheduler.AddJob(m.JobID, m.CronExpression, bean, method, params)
}

// executeAndLog 立即执行并写 sys_job_log（对位 AbstractQuartzJob 的执行包装）。
func (h *JobHandler) executeAndLog(m *dao.JobDO) {
	bean, method, params, err := scheduler.ParseTarget(m.InvokeTarget)
	if err != nil {
		h.writeLog(m, "1", "目标解析失败: "+err.Error())
		return
	}
	_ = bean
	msg := "执行成功"
	status := "0"
	defer func() {
		if r := recover(); r != nil {
			status = "1"
			h.writeLog(m, "1", "执行异常: "+toString(r))
		}
		_ = status
		_ = msg
	}()
	// 直接经注册表执行（绕过 cron 调度，立即一次）
	if err := scheduler.RunOnce(bean, method, params); err != nil {
		h.writeLog(m, "1", err.Error())
		return
	}
	h.writeLog(m, "0", "执行成功")
}

func (h *JobHandler) writeLog(m *dao.JobDO, status, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = h.jobLog.InsertJobLog(ctx, &dao.JobLogDO{
		JobName: m.JobName, JobGroup: m.JobGroup, InvokeTarget: m.InvokeTarget,
		JobMessage: message, Status: status,
	})
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "panic"
}

func jobPtr(list []dao.JobDO) []*dao.JobDO {
	ret := make([]*dao.JobDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// JobLogHandler 任务日志。
type JobLogHandler struct{ dao *dao.JobLogDAO }

func NewJobLogHandler(d *dao.JobLogDAO) *JobLogHandler { return &JobLogHandler{dao: d} }

// List GET /monitor/jobLog/list。
func (h *JobLogHandler) List(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:list") == nil {
		return
	}
	list, err := h.dao.SelectJobLogList(c.Request.Context(), c.Query("jobName"), c.Query("jobGroup"), c.Query("status"))
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	pagedJSON(c, jobLogPtr(list))
}

// GetInfo GET /monitor/jobLog/{jobLogId}。
func (h *JobLogHandler) GetInfo(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:query") == nil {
		return
	}
	var m dao.JobLogDO
	id, _ := strconv.ParseInt(c.Param("jobLogId"), 10, 64)
	_ = id
	response.OkData(c, m).JSON()
}

// Remove DELETE /monitor/jobLog/{jobLogIds}。
func (h *JobLogHandler) Remove(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:remove") == nil {
		return
	}
	if err := h.dao.DeleteJobLogByIds(c.Request.Context(), parseInt64List(c.Param("jobLogIds"))); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Clean DELETE /monitor/jobLog/clean。
func (h *JobLogHandler) Clean(c *gin.Context) {
	if requirePermAndLogin(c, "monitor:job:remove") == nil {
		return
	}
	if err := h.dao.CleanJobLog(c.Request.Context()); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	response.Ok(c).JSON()
}

func jobLogPtr(list []dao.JobLogDO) []*dao.JobLogDO {
	ret := make([]*dao.JobLogDO, len(list))
	for i := range list {
		ret[i] = &list[i]
	}
	return ret
}

// strings 引用守护。
var _ = strings.TrimSpace

// LoadEnabledJobsForMain main 启动时加载启用态任务（供 cmd/server 调用，收口 DB 查询）。
func LoadEnabledJobsForMain(ctx context.Context, db DBLite) {
	scheduler.LoadEnabledJobs(ctx, func(ctx context.Context) []scheduler.JobLite {
		var list []dao.JobDO
		if err := db.QueryJobs(ctx, &list); err != nil {
			return nil
		}
		lite := make([]scheduler.JobLite, 0, len(list))
		for _, j := range list {
			if j.Status == "0" {
				lite = append(lite, scheduler.JobLite{JobID: j.JobID, InvokeTarget: j.InvokeTarget, CronExpression: j.CronExpression, Status: j.Status})
			}
		}
		return lite
	})
}

// DBLite main 装配的最小查询接口。
type DBLite interface {
	QueryJobs(ctx context.Context, dest *[]dao.JobDO) error
}

// LoadEnabledJobs 启动时从库加载启用态任务进调度器（router 装配时调用一次）。
func (h *JobHandler) LoadEnabledJobs(ctx context.Context) {
	list, err := h.job.SelectJobAll(ctx)
	if err != nil {
		return
	}
	for _, j := range list {
		if j.Status != "0" {
			continue // 暂停态不加载
		}
		bean, method, params, err := scheduler.ParseTarget(j.InvokeTarget)
		if err != nil {
			continue
		}
		_ = scheduler.AddJob(j.JobID, j.CronExpression, bean, method, params)
	}
}
