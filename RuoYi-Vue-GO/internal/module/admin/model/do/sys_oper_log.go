// Package do GORM 表模型（对位 Java domain / 数据库表）。
//
// 约定：列名（gorm column tag）与 DDL 逐字一致；json tag 显式驼峰（AST 卡点）；
// DO 只对表负责，响应出参在 vo 层另建模型，不复用 DO。
package do

import "ruoyi-vue-go/pkg/types"

// SysOperLog 操作日志记录（对位 sys_oper_log 表 17 字段 / Java SysOperLog.java）。
type SysOperLog struct {
	OperID        int64          `gorm:"column:oper_id;primaryKey;autoIncrement" json:"operId"`
	Title         string         `gorm:"column:title" json:"title"`
	BusinessType  int            `gorm:"column:business_type" json:"businessType"` // 枚举值 = Java BusinessType.ordinal
	Method        string         `gorm:"column:method" json:"method"`              // 包路径.处理函数()
	RequestMethod string         `gorm:"column:request_method" json:"requestMethod"`
	OperatorType  int            `gorm:"column:operator_type" json:"operatorType"` // 枚举值 = Java OperatorType.ordinal
	OperName      string         `gorm:"column:oper_name" json:"operName"`
	DeptName      string         `gorm:"column:dept_name" json:"deptName"`
	OperURL       string         `gorm:"column:oper_url" json:"operUrl"`
	OperIP        string         `gorm:"column:oper_ip" json:"operIp"`
	OperLocation  string         `gorm:"column:oper_location" json:"operLocation"`
	OperParam     string         `gorm:"column:oper_param" json:"operParam"`   // 截断 2000，敏感字段已剔除
	JsonResult    string         `gorm:"column:json_result" json:"jsonResult"` // 截断 2000
	Status        int            `gorm:"column:status" json:"status"`          // 0 成功 / 1 失败（Java BusinessStatus.ordinal）
	ErrorMsg      string         `gorm:"column:error_msg" json:"errorMsg"`     // 截断 2000
	OperTime      types.DateTime `gorm:"column:oper_time" json:"operTime"`
	CostTime      int64          `gorm:"column:cost_time" json:"costTime"` // 毫秒
}

// TableName 表名。
func (SysOperLog) TableName() string { return "sys_oper_log" }
