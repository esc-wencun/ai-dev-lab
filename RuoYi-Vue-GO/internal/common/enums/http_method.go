package enums

// 本文件对位 HttpMethod.java（请求方式，操作日志 request_method 与匿名放行匹配使用）。
type HttpMethod string

const (
	HttpMethodGet     HttpMethod = "GET"
	HttpMethodHead    HttpMethod = "HEAD"
	HttpMethodPost    HttpMethod = "POST"
	HttpMethodPut     HttpMethod = "PUT"
	HttpMethodPatch   HttpMethod = "PATCH"
	HttpMethodDelete  HttpMethod = "DELETE"
	HttpMethodOptions HttpMethod = "OPTIONS"
	HttpMethodTrace   HttpMethod = "TRACE"
)

// Resolve 按字符串解析请求方式（对位 Java HttpMethod.resolve，未匹配返回 false）
func ResolveHttpMethod(method string) (HttpMethod, bool) {
	switch HttpMethod(method) {
	case HttpMethodGet, HttpMethodHead, HttpMethodPost, HttpMethodPut,
		HttpMethodPatch, HttpMethodDelete, HttpMethodOptions, HttpMethodTrace:
		return HttpMethod(method), true
	default:
		return "", false
	}
}
