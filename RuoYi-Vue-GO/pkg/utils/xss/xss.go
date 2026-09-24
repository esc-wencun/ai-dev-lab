// Package xss XSS 输入过滤（对位 Java XssFilter + XssHttpServletRequestWrapper，语义对齐 Python 版 xss_util）。
//
// 规则（对位 Java XssFilter.handleExcludeURL + application.yml xss 配置）：
//   - 只过滤 POST/PUT（GET/DELETE 直接放行）；
//   - 排除名单内的路径跳过清洗（Java 默认排除 /system/notice 富文本）；
//   - JSON body 中的字符串值剥离 HTML 标签（对位 EscapeUtil.clean → HTMLFilter 剥标签语义，
//     对齐 Python 版 `<[^>]*>` 正则实现）；非 JSON body 原样放行。
//
// 用法（router 挂载，对位 urlPatterns /system/*,/monitor/*,/tool/*）：
//
//	r.Use(xss.Middleware(xss.Options{Excludes: []string{"/system/notice"}}))
package xss

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// tagRe HTML 标签模式（对齐 Python 版 _TAG_RE：剥离成对/自闭合标签）。
var tagRe = regexp.MustCompile(`<[^>]*>`)

// DefaultExcludes 默认排除名单（对位 Java application.yml xss.excludes）。
var DefaultExcludes = []string{"/system/notice"}

// Options 中间件选项。
type Options struct {
	Excludes []string // 排除清洗的路径前缀（空则用 DefaultExcludes）
}

// Clean 剥离字符串中的 HTML 标签（导出供单字段清洗复用）。
func Clean(s string) string {
	return tagRe.ReplaceAllString(s, "")
}

// cleanValue 递归清洗 JSON 值中的字符串（dict/list/字符串；深度防循环引用上限 16）。
func cleanValue(v any, depth int) any {
	if depth > 16 {
		return v
	}
	switch t := v.(type) {
	case string:
		return Clean(t)
	case map[string]any:
		for k, sub := range t {
			t[k] = cleanValue(sub, depth+1)
		}
		return t
	case []any:
		for i, sub := range t {
			t[i] = cleanValue(sub, depth+1)
		}
		return t
	default:
		return v
	}
}

// excluded 前缀匹配排除名单（对位 Python 版 startswith；Java 侧为通配匹配，前缀语义等效）。
func excluded(path string, patterns []string) bool {
	for _, p := range patterns {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// Middleware gin 中间件：POST/PUT 的 JSON body 字符串值清洗后回放（handler 正常 bind）。
func Middleware(opt Options) gin.HandlerFunc {
	if len(opt.Excludes) == 0 {
		opt.Excludes = DefaultExcludes
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" { // 对位 GET/DELETE 不过滤
			c.Next()
			return
		}
		if excluded(c.Request.URL.Path, opt.Excludes) {
			c.Next()
			return
		}
		if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			c.Next()
			return
		}
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		var v any
		if json.Unmarshal(body, &v) != nil {
			// 非 JSON body：原样回放放行（对位 isJsonRequest 分支下的空/非 JSON 行为）
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}
		cleaned, err := json.Marshal(cleanValue(v, 0))
		if err != nil {
			cleaned = body
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(cleaned))
		c.Next()
	}
}
