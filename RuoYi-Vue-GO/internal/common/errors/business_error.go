// Package errors 统一业务错误类型（对位 Java ServiceException + GlobalExceptionHandler 的映射约定）。
//
// 全局唯一模式：service 层遇业务失败构造并 return 本错误；
// handler 层不做错误信封拼装，由 internal/middleware.Wrap 统一映射为响应信封。
package errors

// BusinessError 业务错误：Code 为响应体信封 code（500/601/403/401 ...），Msg 为面向用户文案。
type BusinessError struct {
	Code int
	Msg  string
}

// Error 实现 error 接口（返回面向用户的文案）
func (e *BusinessError) Error() string {
	return e.Msg
}

// New 业务失败，code 默认 500（对位 AjaxResult.error(msg)）
func New(msg string) *BusinessError {
	return &BusinessError{Code: 500, Msg: msg}
}

// NewWithCode 指定信封 code 的业务失败（如 601 警告 / 403 无权限）
func NewWithCode(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Msg: msg}
}
