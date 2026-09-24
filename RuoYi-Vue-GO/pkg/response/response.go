// Package response 响应信封（对位 Java AjaxResult / Python utils/response_util.py）。
//
// 契约（与 Java/Python 版一致，前端按 body code 判断，禁止改成 HTTP 状态码语义）：
//   - 全部响应 HTTP 200；
//   - code：200 成功 / 500 失败 / 601 警告 / 403 无权限 / 401 未登录；
//   - 自由字段（token/uuid/img/rows/total ...）通过 Put 置于顶层，对位 AjaxResult.put。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务码（对位 Java HttpStatus 常量）
const (
	CodeSuccess      = 200
	CodeError        = 500
	CodeWarn         = 601
	CodeForbidden    = 403
	CodeUnauthorized = 401
)

type field struct {
	key string
	val any
}

// R 链式信封构建器
type R struct {
	c       *gin.Context
	code    int
	msg     string
	data    any
	hasData bool
	extra   []field
}

// Ok 成功响应（默认文案对位 AjaxResult.success）
func Ok(c *gin.Context) *R {
	return &R{c: c, code: CodeSuccess, msg: "操作成功"}
}

// OkData 成功响应并携带 data
func OkData(c *gin.Context, data any) *R {
	return Ok(c).Data(data)
}

// Error 失败响应（code=500，默认文案对位 AjaxResult.error）
func Error(c *gin.Context, msg string) *R {
	if msg == "" {
		msg = "操作失败"
	}
	return &R{c: c, code: CodeError, msg: msg}
}

// Warn 警告响应（code=601，对位 AjaxResult.warn）
func Warn(c *gin.Context, msg string) *R {
	if msg == "" {
		msg = "操作警告"
	}
	return &R{c: c, code: CodeWarn, msg: msg}
}

// Forbidden 无权限响应（body code=403，文案对位 GlobalExceptionHandler）
func Forbidden(c *gin.Context) *R {
	return &R{c: c, code: CodeForbidden, msg: "没有权限，请联系管理员授权"}
}

// Unauthorized 未登录响应（HTTP 200 + body code=401，文案对位 AuthenticationEntryPointImpl，
// 前端 axios 拦截器按 code=401 弹重新登录）
func Unauthorized(c *gin.Context) *R {
	return &R{c: c, code: CodeUnauthorized, msg: "请求访问：认证失败，无法访问系统资源"}
}

// Msg 覆盖默认文案
func (r *R) Msg(msg string) *R {
	r.msg = msg
	return r
}

// Data 设置 data 字段（列表接口的 rows/total 优先用 Put）
func (r *R) Data(data any) *R {
	r.data = data
	r.hasData = true
	return r
}

// Put 追加顶层自由字段（对位 AjaxResult.put，重复 key 后写覆盖先写）
func (r *R) Put(key string, val any) *R {
	r.extra = append(r.extra, field{key: key, val: val})
	return r
}

// JSON 输出信封（HTTP 恒为 200）
func (r *R) JSON() {
	body := make(gin.H, 2+len(r.extra)+1)
	body["code"] = r.code
	body["msg"] = r.msg
	if r.hasData {
		body["data"] = r.data
	}
	for _, f := range r.extra {
		body[f.key] = f.val
	}
	r.c.JSON(http.StatusOK, body)
}
