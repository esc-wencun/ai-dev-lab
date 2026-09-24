// 预置任务注册（对位 com.ruoyi.quartz.task.RyTask：Java sys_job 预置 3 条任务的目标）。
package scheduler

import (
	"context"
	"fmt"

	"ruoyi-vue-go/pkg/logger"
)

func init() {
	Register("ryTask.ryNoParams", ryNoParams)
	Register("ryTask.ryParams", ryParams)
	Register("ryTask.ryMultipleParams", ryMultipleParams)
}

// ryNoParams 无参方法（Java 输出"执行无参方法"）。
func ryNoParams(ctx context.Context, params []any) {
	logger.Info("执行无参方法")
}

// ryParams 单字符串参数。
func ryParams(ctx context.Context, params []any) {
	logger.Info(fmt.Sprintf("执行有参方法：%v", params))
}

// ryMultipleParams 多参数（str,bool,long,double,int）。
func ryMultipleParams(ctx context.Context, params []any) {
	logger.Info(fmt.Sprintf("执行多参方法：%v", params))
}
