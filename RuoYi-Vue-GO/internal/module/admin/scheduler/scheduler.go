// Package scheduler 定时任务调度（对位 ruoyi-quartz：Scheduler + JobInvokeUtil + AbstractQuartzJob）。
//
// 设计要点（规避 Python 版踩坑 5）：
//   - robfig/cron v3 以 Quartz 秒级 6 位表达式运行（cron.NewParser 逐字对位 Java cron 语义）；
//   - 暂停态任务不注册进调度器；changeStatus 恢复时从库读 cron 后 AddJob（不存在"静默空操作"）；
//   - invokeTarget 采用注册表方案（同 Python 版）：ryTask.ryNoParams 等预置任务在 tasks.go 注册，
//     Go 无 Java 反射——语法解析兼容（参数类型推断一致），执行走注册表（deviations #4）。
package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"ruoyi-vue-go/pkg/logger"
)

// TaskFunc 任务执行体。
type TaskFunc func(ctx context.Context, params []any)

var (
	mu        sync.RWMutex
	registry  = map[string]TaskFunc{}
	scheduler *cron.Cron
	entries   = map[int64]cron.EntryID{} // job_id → cron entry
)

// Register 注册任务（main 启动时对预置任务调用）。
func Register(name string, fn TaskFunc) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = fn
}

// Registered 全部已注册任务名。
func Registered() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// Start 启动调度器（Quartz 秒级表达式 parser）。
func Start() {
	scheduler = cron.New(cron.WithParser(cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	scheduler.Start()
}

// Stop 优雅停止（等待运行中的任务）。
func Stop() context.Context {
	if scheduler == nil {
		return context.Background()
	}
	return scheduler.Stop()
}

// AddJob 把任务加入调度器（对位 ScheduleUtils.createScheduleJob）。
// beanName+methodName 来自 invokeTarget 解析；misfirePolicy 在 cron v3 无对应物，忽略（补跑语义由
// beginTime 窗口判断的复杂策略不对齐，登记 deviations）。
func AddJob(jobID int64, cronExpr, beanName, methodName string, params []any) error {
	fn, ok := lookup(beanName, methodName)
	if !ok {
		return fmt.Errorf("任务目标未注册: %s.%s，可用: %v", beanName, methodName, Registered())
	}
	id, err := scheduler.AddFunc(cronExpr, func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("定时任务[%d]执行异常: %v", jobID, r))
			}
		}()
		fn(context.Background(), params)
	})
	if err != nil {
		return err
	}
	mu.Lock()
	entries[jobID] = id
	mu.Unlock()
	return nil
}

// DeleteJob 移除任务（对位 deleteScheduleJob）。
func DeleteJob(jobID int64) {
	mu.Lock()
	defer mu.Unlock()
	if id, ok := entries[jobID]; ok {
		scheduler.Remove(id)
		delete(entries, jobID)
	}
}

// LookupJob 任务是否在调度中。
func LookupJob(jobID int64) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := entries[jobID]
	return ok
}

// lookup 按 bean.method 查注册表；兼容注册名 "bean.method" 与裸 "method"。
func lookup(beanName, methodName string) (TaskFunc, bool) {
	mu.RLock()
	defer mu.RUnlock()
	fn, ok := registry[beanName+"."+methodName]
	if !ok {
		fn, ok = registry[methodName]
	}
	return fn, ok
}

// ParseTarget 解析 invokeTarget（对位 JobInvokeUtil：bean.method(参数列表)）。
// 参数类型推断：'x'→string；true/false→bool；123L→int64；1.5D→float64；其他→int。
func ParseTarget(target string) (beanName, methodName string, params []any, err error) {
	open := strings.Index(target, "(")
	if open < 0 {
		parts := strings.SplitN(target, ".", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], nil, nil
		}
		return "", "", nil, fmt.Errorf("目标字符串非法: %s", target)
	}
	if !strings.HasSuffix(target, ")") {
		return "", "", nil, fmt.Errorf("目标字符串非法: %s", target)
	}
	full := target[:open]
	parts := strings.SplitN(full, ".", 2)
	if len(parts) != 2 {
		return "", "", nil, fmt.Errorf("目标字符串非法: %s", target)
	}
	beanName, methodName = parts[0], parts[1]
	paramStr := strings.TrimSuffix(strings.TrimPrefix(target[open:], "("), ")")
	for _, p := range smartSplit(paramStr) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		params = append(params, inferParam(p))
	}
	return beanName, methodName, params, nil
}

// smartSplit 引号感知逗号切分（对位 Python/Java 正则语义）。
func smartSplit(s string) []string {
	var parts []string
	var buf strings.Builder
	var quote rune
	for _, ch := range s {
		switch {
		case quote != 0:
			buf.WriteRune(ch)
			if ch == quote {
				quote = 0
			}
		case ch == '\'' || ch == '"':
			quote = ch
			buf.WriteRune(ch)
		case ch == ',':
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteRune(ch)
		}
	}
	parts = append(parts, buf.String())
	return parts
}

// inferParam 参数类型推断（对位 getMethodParams）。
func inferParam(p string) any {
	if (strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'")) ||
		(strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"")) {
		return strings.Trim(p, "'\"")
	}
	switch strings.ToLower(p) {
	case "true":
		return true
	case "false":
		return false
	}
	if strings.HasSuffix(p, "L") || strings.HasSuffix(p, "l") {
		if v, err := strconv.ParseInt(strings.TrimSuffix(p[:len(p)-1], ""), 10, 64); err == nil {
			return v
		}
	}
	if strings.HasSuffix(p, "D") || strings.HasSuffix(p, "d") {
		if v, err := strconv.ParseFloat(strings.TrimSuffix(p[:len(p)-1], ""), 64); err == nil {
			return v
		}
	}
	if v, err := strconv.ParseInt(p, 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(p, 64); err == nil {
		return v
	}
	return p
}

// guard。
var _ = time.Now

// RunOnce 立即执行一次（对位 run 一键执行一次；不进调度器）。
func RunOnce(beanName, methodName string, params []any) error {
	fn, ok := lookup(beanName, methodName)
	if !ok {
		return fmt.Errorf("任务目标未注册: %s.%s，可用: %v", beanName, methodName, Registered())
	}
	fn(context.Background(), params)
	return nil
}

// LoadEnabledJobs 从库加载全部启用态任务（main 启动调用；对位 init 的 scheduleJob 循环）。
// db 句柄由调用方注入以避免 scheduler 包依赖 dao（保持调度器独立可测）。
func LoadEnabledJobs(ctx context.Context, query func(ctx context.Context) []JobLite) {
	for _, j := range query(ctx) {
		bean, method, params, err := ParseTarget(j.InvokeTarget)
		if err != nil {
			logger.Error(fmt.Sprintf("任务[%d]目标解析失败: %v", j.JobID, err))
			continue
		}
		if err := AddJob(j.JobID, j.CronExpression, bean, method, params); err != nil {
			logger.Error(fmt.Sprintf("任务[%d]加载失败: %v", j.JobID, err))
		}
	}
}

// JobLite 加载所需的任务最小信息。
type JobLite struct {
	JobID          int64
	InvokeTarget   string
	CronExpression string
	Status         string
}
