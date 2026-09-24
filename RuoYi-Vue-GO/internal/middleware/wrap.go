// Package middleware gin 中间件：统一错误映射（Wrap）与 panic 兜底（Recovery）。
//
// 对位 Java GlobalExceptionHandler / Python exceptions/handle.py：
//   - 业务错误（errors.BusinessError）→ 按其 Code/Msg 输出信封；
//   - 其他 error → code=500 + 错误文本（对位 Java error(e.getMessage())）；
//   - panic → Recovery 兜底转 code=500 信封；
//   - 禁止各 handler 自行 c.JSON 拼错误信封。
package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	bizerrors "ruoyi-vue-go/internal/common/errors"
	"ruoyi-vue-go/pkg/response"
)

// HandlerFunc 统一 handler 签名：handler 返回 error，由 Wrap 映射为响应信封。
type HandlerFunc func(c *gin.Context) error

// Wrap 把 handler 返回的 error 映射为响应信封（HTTP 恒 200）。
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			var be *bizerrors.BusinessError
			if asBusinessError(err, &be) {
				body := gin.H{"code": be.Code, "msg": be.Msg}
				c.JSON(http.StatusOK, body)
				return
			}
			// 非业务错误：对位 Java error(e.getMessage())
			log.Printf("请求地址'%s',发生异常: %v", c.Request.URL.Path, err)
			response.Error(c, err.Error()).JSON()
		}
	}
}

// asBusinessError 判断 err 是否为 *BusinessError（便于测试与后续扩展错误类型）
func asBusinessError(err error, target **bizerrors.BusinessError) bool {
	if be, ok := err.(*bizerrors.BusinessError); ok {
		*target = be
		return true
	}
	return false
}
