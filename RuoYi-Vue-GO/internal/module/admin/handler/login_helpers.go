// 登录闭环 handler 的辅助函数（会话组装 / 用户 JSON 视图 / 时间工具）。
package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/service"
)

// context0 保证非 nil context（防御）。
func context0(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// jsonRaw 把会话内嵌 user 原始 JSON 包装为可序列化对象（保持原字段形态）。
func jsonRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("{}")
	}
	return raw
}

// buildLoginUser 组装会话对象（对位 new LoginUser(userId, deptId, user, permissions)）。
// 视图组装在 service 层（service.BuildLoginUser）。
func buildLoginUser(ud *dao.UserDetail, roles, permissions []string) (*model.LoginUser, error) {
	return service.BuildLoginUser(ud, roles, permissions)
}

// userPwdUpdateDate 从会话 user JSON 取 pwd_update_date（nil=从未改密）。
func userPwdUpdateDate(lu *model.LoginUser) *time.Time {
	if len(lu.User) == 0 {
		return nil
	}
	var view struct {
		PwdUpdateDate *struct {
			Time time.Time `json:"Time"`
		} `json:"pwdUpdateDate"`
		PwdUpdateDateStr *string `json:"pwdUpdateDateStr"`
	}
	// types.DateTime 的 JSON 形态是 "yyyy-MM-dd HH:mm:ss" 字符串，直接按字符串解析
	var v2 struct {
		PwdUpdateDate *string `json:"pwdUpdateDate"`
	}
	if err := json.Unmarshal(lu.User, &v2); err != nil || v2.PwdUpdateDate == nil || *v2.PwdUpdateDate == "" {
		return nil
	}
	_ = view
	t, err := time.ParseInLocation("2006-01-02 15:04:05", *v2.PwdUpdateDate, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

// parsePositiveInt 字符串转正整数（对位 Convert.toInt 的本场景语义；非法返回 0）。
func parsePositiveInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// elapsedDays 距今天数（对位 DateUtils.differentDaysByMillisecond(now, pwdUpdateDate)）。
func elapsedDays(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}
