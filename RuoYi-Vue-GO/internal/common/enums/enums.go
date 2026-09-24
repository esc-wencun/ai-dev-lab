// Package enums 通用枚举（对位 Java com.ruoyi.common.enums）。
//
// 约定：全部用 typed const 显式赋值，不用 iota 制造隐式序数；
// 落库值（如 @Log business_type、操作类别、操作状态）必须显式等于 Java 枚举的 ordinal。
// 未移植：DataSourceType（GO 版单数据源，不复刻动态数据源切换）、
// DesensitizedType（脱敏工具枚举，待有模块用到时再补）。
package enums

// BusinessType 业务操作类型（对位 BusinessType.java，@Log business_type 落库值 = 序数 0-9）
type BusinessType int

const (
	BusinessTypeOther   BusinessType = 0 // 其它
	BusinessTypeInsert  BusinessType = 1 // 新增
	BusinessTypeUpdate  BusinessType = 2 // 修改
	BusinessTypeDelete  BusinessType = 3 // 删除
	BusinessTypeGrant   BusinessType = 4 // 授权
	BusinessTypeExport  BusinessType = 5 // 导出
	BusinessTypeImport  BusinessType = 6 // 导入
	BusinessTypeForce   BusinessType = 7 // 强退
	BusinessTypeGencode BusinessType = 8 // 生成代码
	BusinessTypeClean   BusinessType = 9 // 清空数据
)

// String 返回 Java 枚举名，便于日志与对表
func (t BusinessType) String() string {
	switch t {
	case BusinessTypeOther:
		return "OTHER"
	case BusinessTypeInsert:
		return "INSERT"
	case BusinessTypeUpdate:
		return "UPDATE"
	case BusinessTypeDelete:
		return "DELETE"
	case BusinessTypeGrant:
		return "GRANT"
	case BusinessTypeExport:
		return "EXPORT"
	case BusinessTypeImport:
		return "IMPORT"
	case BusinessTypeForce:
		return "FORCE"
	case BusinessTypeGencode:
		return "GENCODE"
	case BusinessTypeClean:
		return "CLEAN"
	default:
		return "UNKNOWN"
	}
}

// OperatorType 操作人类别（对位 OperatorType.java，操作日志 operate_type 落库值 = 序数 0-2）
type OperatorType int

const (
	OperatorTypeOther  OperatorType = 0 // 其它
	OperatorTypeManage OperatorType = 1 // 后台用户
	OperatorTypeMobile OperatorType = 2 // 手机端用户
)

// String 返回 Java 枚举名
func (t OperatorType) String() string {
	switch t {
	case OperatorTypeOther:
		return "OTHER"
	case OperatorTypeManage:
		return "MANAGE"
	case OperatorTypeMobile:
		return "MOBILE"
	default:
		return "UNKNOWN"
	}
}

// BusinessStatus 操作状态（对位 BusinessStatus.java，操作日志 status 落库值 = 序数 0-1）
type BusinessStatus int

const (
	BusinessStatusSuccess BusinessStatus = 0 // 成功
	BusinessStatusFail    BusinessStatus = 1 // 失败
)

// String 返回 Java 枚举名
func (s BusinessStatus) String() string {
	switch s {
	case BusinessStatusSuccess:
		return "SUCCESS"
	case BusinessStatusFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// LimitType 限流类型（对位 LimitType.java）
type LimitType int

const (
	LimitTypeDefault LimitType = 0 // 默认策略全局限流
	LimitTypeIP      LimitType = 1 // 根据请求者 IP 进行限流
)

// String 返回 Java 枚举名
func (t LimitType) String() string {
	switch t {
	case LimitTypeDefault:
		return "DEFAULT"
	case LimitTypeIP:
		return "IP"
	default:
		return "UNKNOWN"
	}
}
