// Package aspect 业务切面（对位 Java com.ruoyi.framework.aspectj）。
//
// 本文件对位 LogAspect + AsyncManager + AsyncFactory.recordOper：
//   - OperLog 包装 handler（对位 @Log 注解），成功/异常/panic 都记录（status 0/1）；
//   - 记录动作异步执行（goroutine + recover，对位 AsyncManager 消费 TimerTask），
//     失败只打本地日志，绝不影响主请求；错误原样上抛由 middleware.Wrap 映射响应信封；
//   - 敏感参数剔除名单与 Java 逐字一致（password/oldPassword/newPassword/confirmPassword）；
//   - 落库截断对位 AsyncFactory.recordOper：operLocation/operUrl 255、operIp 128、
//     operName 50、operParam/jsonResult/errorMsg 2000。
package aspect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ruoyi-vue-go/internal/common/enums"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/model/do"
	"ruoyi-vue-go/pkg/types"
)

// EXCLUDE_PROPERTIES 敏感参数过滤名单（对位 LogAspect.EXCLUDE_PROPERTIES，逐字一致，禁止改动）
var EXCLUDE_PROPERTIES = []string{"password", "oldPassword", "newPassword", "confirmPassword"}

// paramMaxLength 参数/结果落库最大长度（对位 PARAM_MAX_LENGTH）
const paramMaxLength = 2000

// OperLogOptions 对位 Java @Log 注解属性（Go 无注解，用显式选项传递）。
type OperLogOptions struct {
	Title              string             // 模块标题
	BusinessType       enums.BusinessType // 业务类型（0-9）
	OperatorType       enums.OperatorType // 操作人类别（0-2）
	IsSaveRequestData  bool               // 是否保存请求参数
	IsSaveResponseData bool               // 是否保存响应结果
	ExcludeParamNames  []string           // 需要额外剔除的参数字段名（对位 @Log excludeParamNames）
}

// OperLogOpt 生成默认选项（对位 @Log 默认值：MANAGE / 保存请求与响应数据）。
func OperLogOpt(title string, businessType enums.BusinessType) OperLogOptions {
	return OperLogOptions{
		Title:              title,
		BusinessType:       businessType,
		OperatorType:       enums.OperatorTypeManage,
		IsSaveRequestData:  true,
		IsSaveResponseData: true,
	}
}

// OperLogger 操作日志切面（db 由 main 装配注入；测试用 sqlite 内存库）。
type OperLogger struct {
	db *gorm.DB
}

// NewOperLogger 构造切面。
func NewOperLogger(db *gorm.DB) *OperLogger {
	return &OperLogger{db: db}
}

// OperLog 包装 handler：执行并记录操作日志，返回值（含 error）原样透传。
// 用法：router.POST("/user", middleware.Wrap(logger.OperLog(OperLogOpt("用户管理", enums.BusinessTypeInsert), handler.AddUser)))
func (l *OperLogger) OperLog(opt OperLogOptions, h middleware.HandlerFunc) middleware.HandlerFunc {
	name := handlerName(h) // 处理函数全名，闭包内不变，只算一次
	return func(c *gin.Context) error {
		start := time.Now()
		reqBody := snapshotBody(c.Request)

		// 捕获响应体（对位 @AfterReturning 拿到 jsonResult）
		bw := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = bw

		err := runRecorded(c, h, func(err error) {
			l.record(c, opt, name, reqBody, bw.buf.Bytes(), err, start)
		})
		return err
	}
}

// runRecorded 执行 handler 并在返回/panic 两个出口都触发记录；
// panic 记录（status=1）后继续上抛，交由 gin Recovery 兜底（对位 @AfterThrowing）。
func runRecorded(c *gin.Context, h middleware.HandlerFunc, record func(error)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			record(fmt.Errorf("%v", r))
			panic(r)
		}
	}()
	err = h(c)
	record(err)
	return err
}

// record 组装日志并异步落库（对位 AsyncManager.me().execute）。
// 用脱离请求的独立 ctx：请求返回后 goroutine 仍可写库。
func (l *OperLogger) record(c *gin.Context, opt OperLogOptions, name string, reqBody, respBody []byte, err error, start time.Time) {
	m := buildOperLog(c, opt, name, reqBody, respBody, err, start)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("操作日志落库异常: %v", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if e := l.db.WithContext(ctx).Create(m).Error; e != nil {
			log.Printf("操作日志落库失败: %v", e)
		}
	}()
}

// buildOperLog 组装 SysOperLog（对位 LogAspect.handleLog + getControllerMethodDescription + recordOper 截断）。
func buildOperLog(c *gin.Context, opt OperLogOptions, name string, reqBody, respBody []byte, err error, start time.Time) *do.SysOperLog {
	m := &do.SysOperLog{
		Title:         opt.Title,
		BusinessType:  int(opt.BusinessType),
		Method:        name,
		RequestMethod: c.Request.Method,
		OperatorType:  int(opt.OperatorType),
		OperURL:       truncateRunes(c.Request.URL.RequestURI(), 255),
		OperIP:        truncateRunes(c.ClientIP(), 128),
		OperParam:     buildOperParam(c, opt, reqBody),
		Status:        int(enums.BusinessStatusSuccess),
		OperTime:      types.Now(),
	}
	if lu := middleware.GetLoginUser(c); lu != nil {
		m.OperName = truncateRunes(lu.Username(), 50)
		m.DeptName = truncateRunes(lu.DeptName(), 50)
	}
	if err != nil {
		m.Status = int(enums.BusinessStatusFail)
		m.ErrorMsg = truncateRunes(err.Error(), paramMaxLength)
	}
	if opt.IsSaveResponseData && err == nil && len(respBody) > 0 {
		m.JsonResult = truncateRunes(string(respBody), paramMaxLength)
	}
	m.CostTime = time.Since(start).Milliseconds()
	return m
}

// buildOperParam 请求参数落库值（对位 setRequestValue）：
//   - parameter map 非空（query + form-urlencoded）优先，取首个值序列化为 JSON；
//   - 否则 POST/PUT/DELETE 用 JSON body（剔除敏感字段）；非 JSON body（如 multipart）
//     记空串，对位 Java isFilterObject 跳过文件/流对象。
func buildOperParam(c *gin.Context, opt OperLogOptions, reqBody []byte) string {
	if !opt.IsSaveRequestData {
		return ""
	}
	if pm := paramMap(c, reqBody); len(pm) > 0 {
		removeExcludes(pm, opt.ExcludeParamNames)
		data, e := json.Marshal(pm)
		if e != nil {
			return ""
		}
		return truncateRunes(string(data), paramMaxLength)
	}
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		return truncateRunes(sanitizeJSON(reqBody, opt.ExcludeParamNames), paramMaxLength)
	}
	return ""
}

// handlerName 处理函数全名（对位 Java className+"."+methodName+"()"）。
func handlerName(h middleware.HandlerFunc) string {
	if h == nil {
		return ""
	}
	name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	return truncateRunes(name+"()", 200)
}

// snapshotBody 读取请求体并回放（handler 内仍可正常读取 body）。
func snapshotBody(r *http.Request) []byte {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
		return nil
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

// paramMap 请求参数表（对位 ServletUtils.getParamMap：query + form-urlencoded 表单）。
func paramMap(c *gin.Context, reqBody []byte) map[string]string {
	m := map[string]string{}
	for k, vs := range c.Request.URL.Query() {
		if len(vs) > 0 {
			m[k] = vs[0]
		}
	}
	if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if vs, err := url.ParseQuery(string(reqBody)); err == nil {
			for k, v := range vs {
				if len(v) > 0 {
					m[k] = v[0]
				}
			}
		}
	}
	return m
}

// sanitizeJSON 解析 JSON body、递归剔除敏感字段后重新序列化；
// 空 body / 非 JSON body 返回空串。
func sanitizeJSON(body []byte, exclude []string) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	removeExcludes(m, exclude)
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}

// removeExcludes 递归删除敏感字段（对位 PropertyPreExcludeFilter 的全深度排除语义）。
// 泛型兼容 map[string]any（JSON body）与 map[string]string（parameter map）。
func removeExcludes[M ~map[string]V, V any](m M, exclude []string) {
	ex := map[string]struct{}{}
	for _, k := range EXCLUDE_PROPERTIES {
		ex[k] = struct{}{}
	}
	for _, k := range exclude {
		ex[k] = struct{}{}
	}
	for k, v := range m {
		if _, ok := ex[k]; ok {
			delete(m, k)
			continue
		}
		if sub, ok := any(v).(map[string]any); ok {
			removeExcludes(sub, exclude)
		}
	}
}

// truncateRunes 按字符（rune）截断，对位 Java StringUtils.substring（char 语义）。
func truncateRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}

// bodyWriter 捕获响应体（其余方法全部代理给原 writer）。
type bodyWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyWriter) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
